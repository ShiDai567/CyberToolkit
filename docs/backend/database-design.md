# CyberToolkit 数据库设计

## 目标

本设计面向当前 `CyberToolkit` 的核心场景：

- 工具目录展示
- 分类与标签筛选
- 工具详情页
- 搜索
- 后台管理工具与分类
- 为后续收藏、提交、审核、统计预留扩展能力

推荐数据库：`PostgreSQL 16+`

## 设计原则

- 公开展示数据与后台管理数据分层
- 优先规范化，避免把分类、标签硬编码在前端文件里
- 使用 `slug` 作为稳定 URL 标识
- 使用 `status` 控制发布状态，避免草稿直接暴露
- 通过关联表支持多标签、多分类扩展

## 实体关系

核心实体：

- `users`：后台用户
- `categories`：工具分类
- `tools`：工具主表
- `tags`：标签
- `tool_tags`：工具与标签多对多关系
- `tool_links`：工具外部链接
- `tool_submissions`：用户提交或编辑建议
- `audit_logs`：后台审计日志

关系说明：

- 一个 `category` 可关联多个 `tools`
- 一个 `tool` 可关联多个 `tags`
- 一个 `tool` 可关联多个 `tool_links`
- 一个 `user` 可创建或更新多个 `tools`
- 一个 `user` 可提交多个 `tool_submissions`

## 枚举建议

### tool_status

- `draft`
- `published`
- `archived`

### submission_status

- `pending`
- `approved`
- `rejected`

### user_role

- `admin`
- `editor`
- `viewer`

### tool_difficulty

- `beginner`
- `intermediate`
- `advanced`
- `expert`

## 表结构

### 1. users

用于后台登录和权限控制。

```sql
create table users (
  id uuid primary key default gen_random_uuid(),
  email varchar(255) not null unique,
  password_hash varchar(255) not null,
  display_name varchar(100) not null,
  role varchar(20) not null default 'editor',
  is_active boolean not null default true,
  last_login_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint chk_users_role
    check (role in ('admin', 'editor', 'viewer'))
);
```

### 2. categories

工具分类表。

```sql
create table categories (
  id uuid primary key default gen_random_uuid(),
  slug varchar(100) not null unique,
  name varchar(100) not null unique,
  description text,
  icon varchar(50),
  sort_order integer not null default 0,
  is_visible boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);
```

### 3. tools

工具主表。

```sql
create table tools (
  id uuid primary key default gen_random_uuid(),
  slug varchar(150) not null unique,
  name varchar(150) not null unique,
  short_description varchar(280) not null,
  long_description text not null,
  category_id uuid not null references categories(id),
  difficulty varchar(20) not null,
  icon varchar(50),
  featured boolean not null default false,
  status varchar(20) not null default 'draft',
  source_type varchar(30) not null default 'manual',
  view_count integer not null default 0,
  favorite_count integer not null default 0,
  published_at timestamptz,
  created_by uuid references users(id),
  updated_by uuid references users(id),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint chk_tools_difficulty
    check (difficulty in ('beginner', 'intermediate', 'advanced', 'expert')),
  constraint chk_tools_status
    check (status in ('draft', 'published', 'archived'))
);
```

字段说明：

- `slug`：详情页 URL 标识，替代当前前端的 `id`
- `short_description`：列表卡片摘要
- `long_description`：详情页正文
- `source_type`：保留未来从脚本、爬虫或外部源同步的能力
- `status`：控制是否对外可见

### 4. tags

标签主表。

```sql
create table tags (
  id uuid primary key default gen_random_uuid(),
  slug varchar(100) not null unique,
  name varchar(100) not null unique,
  created_at timestamptz not null default now()
);
```

### 5. tool_tags

工具与标签关联表。

```sql
create table tool_tags (
  tool_id uuid not null references tools(id) on delete cascade,
  tag_id uuid not null references tags(id) on delete cascade,
  primary key (tool_id, tag_id)
);
```

### 6. tool_links

用于管理官网、源码、文档等外部链接，不把链接字段锁死在工具主表。

```sql
create table tool_links (
  id uuid primary key default gen_random_uuid(),
  tool_id uuid not null references tools(id) on delete cascade,
  link_type varchar(30) not null,
  label varchar(50) not null,
  url text not null,
  sort_order integer not null default 0,
  created_at timestamptz not null default now(),
  constraint chk_tool_links_type
    check (link_type in ('website', 'github', 'docs', 'download', 'other'))
);
```

当前前端可直接映射：

- 官网：`website`
- 源码：`github`

### 7. tool_submissions

用于未来开放用户提交新工具或修正建议。

```sql
create table tool_submissions (
  id uuid primary key default gen_random_uuid(),
  submitted_by uuid references users(id),
  tool_id uuid references tools(id),
  submitter_email varchar(255),
  payload jsonb not null,
  status varchar(20) not null default 'pending',
  reviewer_id uuid references users(id),
  review_note text,
  created_at timestamptz not null default now(),
  reviewed_at timestamptz,
  constraint chk_tool_submissions_status
    check (status in ('pending', 'approved', 'rejected'))
);
```

### 8. audit_logs

后台操作审计。

```sql
create table audit_logs (
  id uuid primary key default gen_random_uuid(),
  user_id uuid references users(id),
  action varchar(50) not null,
  resource_type varchar(50) not null,
  resource_id uuid,
  before_data jsonb,
  after_data jsonb,
  created_at timestamptz not null default now()
);
```

## 索引设计

```sql
create index idx_categories_sort_order on categories(sort_order);

create index idx_tools_category_id on tools(category_id);
create index idx_tools_status on tools(status);
create index idx_tools_featured on tools(featured);
create index idx_tools_published_at on tools(published_at desc);
create index idx_tools_category_status on tools(category_id, status);

create index idx_tool_links_tool_id on tool_links(tool_id);

create index idx_tool_submissions_status on tool_submissions(status);
create index idx_audit_logs_user_id on audit_logs(user_id);
create index idx_audit_logs_resource on audit_logs(resource_type, resource_id);
```

### 搜索索引

如果使用 PostgreSQL 全文搜索，建议增加：

```sql
alter table tools
add column search_vector tsvector;

create index idx_tools_search_vector
  on tools using gin(search_vector);
```

可由触发器维护：

- `name`
- `short_description`
- `long_description`
- 标签名聚合文本

如果后续搜索需求更复杂，再升级为：

- `Meilisearch`
- `Typesense`
- `Elasticsearch`

## 初期最小可用版本

如果你现在只想先把后端跑起来，最小表集是：

- `categories`
- `tools`
- `tags`
- `tool_tags`
- `tool_links`

这五张表已经足够支撑：

- 首页精选
- 工具列表
- 分类筛选
- 标签展示
- 工具详情
- 搜索

## 示例查询

### 查询已发布的精选工具

```sql
select
  t.id,
  t.slug,
  t.name,
  t.short_description,
  t.difficulty,
  t.icon,
  c.name as category_name
from tools t
join categories c on c.id = t.category_id
where t.status = 'published'
  and t.featured = true
order by t.published_at desc nulls last, t.created_at desc;
```

### 查询工具详情及标签

```sql
select
  t.*,
  c.name as category_name,
  c.slug as category_slug
from tools t
join categories c on c.id = t.category_id
where t.slug = $1
  and t.status = 'published';
```

```sql
select tg.name, tg.slug
from tool_tags tt
join tags tg on tg.id = tt.tag_id
where tt.tool_id = $1
order by tg.name;
```

## 与当前前端字段映射

当前前端 `Tool` 类型：

```ts
interface Tool {
  id: string;
  name: string;
  description: string;
  longDescription: string;
  category: string;
  tags: string[];
  website: string;
  github?: string;
  difficulty: Difficulty;
  icon: string;
  featured?: boolean;
}
```

建议后端响应映射：

- `id` <- `tools.slug`
- `description` <- `tools.short_description`
- `longDescription` <- `tools.long_description`
- `category` <- `categories.slug`
- `tags` <- 聚合后的标签名数组
- `website` / `github` <- 从 `tool_links` 按类型取出

## 后续扩展位

数据库层面已经预留：

- 收藏：`user_favorites`
- 评论：`tool_comments`
- 评分：`tool_ratings`
- 多语言：`tool_translations`
- 自动抓取：`sync_jobs` / `sync_runs`
- SEO：`seo_meta`

如果后面你确认需要，我下一步可以直接把这份设计转成：

- `Prisma Schema`
- `PostgreSQL DDL`
- `OpenAPI 3.1 YAML`
