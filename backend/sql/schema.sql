create extension if not exists pgcrypto;

-- 用户表：存储所有注册用户信息
create table users (
  id uuid primary key default gen_random_uuid(),        -- 用户唯一标识
  username varchar(50) not null unique,                 -- 用户名，用于登录
  email varchar(255) not null unique,                   -- 邮箱地址
  password_hash varchar(255) not null,                  -- 密码哈希（SHA-256）
  display_name varchar(100) not null,                   -- 显示名称/昵称
  role varchar(20) not null default 'editor',           -- 用户角色：admin / editor / viewer
  is_active boolean not null default true,              -- 是否启用
  last_login_at timestamptz,                            -- 最后登录时间
  created_at timestamptz not null default now(),        -- 注册时间
  updated_at timestamptz not null default now(),        -- 信息更新时间
  constraint chk_users_role
    check (role in ('admin', 'editor', 'viewer'))
);

-- 工具分类表
create table categories (
  id uuid primary key default gen_random_uuid(),        -- 分类唯一标识
  slug varchar(100) not null unique,                    -- 分类 URL 标识
  name varchar(100) not null unique,                    -- 分类名称
  description text,                                     -- 分类描述
  icon varchar(50),                                     -- 图标名称
  sort_order integer not null default 0,                -- 排序权重
  is_visible boolean not null default true,             -- 是否可见
  created_at timestamptz not null default now(),        -- 创建时间
  updated_at timestamptz not null default now()         -- 更新时间
);

-- 标签表
create table tags (
  id uuid primary key default gen_random_uuid(),        -- 标签唯一标识
  slug varchar(100) not null unique,                    -- 标签 URL 标识
  name varchar(100) not null unique,                    -- 标签名称
  created_at timestamptz not null default now()         -- 创建时间
);

-- 工具表：收录的网络安全工具
create table tools (
  id uuid primary key default gen_random_uuid(),        -- 工具唯一标识
  slug varchar(150) not null unique,                    -- 工具 URL 标识
  name varchar(150) not null unique,                    -- 工具名称
  short_description varchar(280) not null,              -- 简短描述（280字符内）
  long_description text not null,                       -- 详细描述
  category_id uuid not null references categories(id),  -- 所属分类
  difficulty varchar(20) not null,                      -- 难度等级：beginner / intermediate / advanced / expert
  icon varchar(50),                                     -- 图标名称
  featured boolean not null default false,              -- 是否精选
  status varchar(20) not null default 'draft',          -- 状态：draft / published / archived
  website_url text not null,                            -- 官方网站
  github_url text,                                      -- GitHub 仓库
  view_count integer not null default 0,                -- 浏览次数
  favorite_count integer not null default 0,            -- 收藏次数
  published_at timestamptz,                             -- 发布时间
  created_by uuid references users(id),                 -- 创建者
  updated_by uuid references users(id),                 -- 最后更新者
  created_at timestamptz not null default now(),        -- 创建时间
  updated_at timestamptz not null default now(),        -- 更新时间
  constraint chk_tools_difficulty
    check (difficulty in ('beginner', 'intermediate', 'advanced', 'expert')),
  constraint chk_tools_status
    check (status in ('draft', 'published', 'archived'))
);

-- 工具-标签关联表（多对多）
create table tool_tags (
  tool_id uuid not null references tools(id) on delete cascade,  -- 工具ID
  tag_id uuid not null references tags(id) on delete cascade,    -- 标签ID
  primary key (tool_id, tag_id)
);

-- 工具投稿表：用户提交的工具审核
create table tool_submissions (
  id uuid primary key default gen_random_uuid(),        -- 投稿唯一标识
  submitted_by uuid references users(id),               -- 投稿用户
  tool_id uuid references tools(id),                    -- 关联的已有工具（可选）
  submitter_email varchar(255),                         -- 投稿者邮箱
  payload jsonb not null,                               -- 投稿内容（JSON）
  status varchar(20) not null default 'pending',        -- 审核状态：pending / approved / rejected
  reviewer_id uuid references users(id),                -- 审核人
  review_note text,                                     -- 审核备注
  created_at timestamptz not null default now(),        -- 投稿时间
  reviewed_at timestamptz,                              -- 审核时间
  constraint chk_tool_submissions_status
    check (status in ('pending', 'approved', 'rejected'))
);

-- 审计日志表：记录用户操作审计
create table audit_logs (
  id uuid primary key default gen_random_uuid(),        -- 日志唯一标识
  user_id uuid references users(id),                    -- 操作用户
  action varchar(50) not null,                          -- 操作类型
  resource_type varchar(50) not null,                   -- 资源类型
  resource_id uuid,                                     -- 资源标识
  before_data jsonb,                                    -- 变更前数据
  after_data jsonb,                                     -- 变更后数据
  created_at timestamptz not null default now()         -- 操作时间
);

create index idx_categories_sort_order on categories(sort_order);
create index idx_tools_category_id on tools(category_id);
create index idx_tools_status on tools(status);
create index idx_tools_featured on tools(featured);
create index idx_tools_published_at on tools(published_at desc);
create index idx_tools_category_status on tools(category_id, status);
create index idx_tool_submissions_status on tool_submissions(status);
create index idx_audit_logs_user_id on audit_logs(user_id);
create index idx_audit_logs_resource on audit_logs(resource_type, resource_id);

-- 数据库注释（COMMENT ON）：使注释在数据库中也可查询
comment on table users is '用户表：存储所有注册用户信息';
comment on table categories is '工具分类表';
comment on table tags is '标签表';
comment on table tools is '工具表：收录的网络安全工具';
comment on table tool_tags is '工具-标签关联表（多对多）';
comment on table tool_submissions is '工具投稿表：用户提交的工具审核';
comment on table audit_logs is '审计日志表：记录用户操作审计';

comment on column users.id is '用户唯一标识';
comment on column users.username is '用户名，用于登录';
comment on column users.email is '邮箱地址';
comment on column users.password_hash is '密码哈希（SHA-256）';
comment on column users.display_name is '显示名称/昵称';
comment on column users.role is '用户角色：admin / editor / viewer';
comment on column users.is_active is '是否启用';
comment on column users.last_login_at is '最后登录时间';
comment on column users.created_at is '注册时间';
comment on column users.updated_at is '信息更新时间';

comment on column categories.id is '分类唯一标识';
comment on column categories.slug is '分类 URL 标识';
comment on column categories.name is '分类名称';
comment on column categories.description is '分类描述';
comment on column categories.icon is '图标名称';
comment on column categories.sort_order is '排序权重';
comment on column categories.is_visible is '是否可见';
comment on column categories.created_at is '创建时间';
comment on column categories.updated_at is '更新时间';

comment on column tags.id is '标签唯一标识';
comment on column tags.slug is '标签 URL 标识';
comment on column tags.name is '标签名称';
comment on column tags.created_at is '创建时间';

comment on column tools.id is '工具唯一标识';
comment on column tools.slug is '工具 URL 标识';
comment on column tools.name is '工具名称';
comment on column tools.short_description is '简短描述（280字符内）';
comment on column tools.long_description is '详细描述';
comment on column tools.category_id is '所属分类';
comment on column tools.difficulty is '难度等级：beginner / intermediate / advanced / expert';
comment on column tools.icon is '图标名称';
comment on column tools.featured is '是否精选';
comment on column tools.status is '状态：draft / published / archived';
comment on column tools.website_url is '官方网站';
comment on column tools.github_url is 'GitHub 仓库';
comment on column tools.view_count is '浏览次数';
comment on column tools.favorite_count is '收藏次数';
comment on column tools.published_at is '发布时间';
comment on column tools.created_by is '创建者';
comment on column tools.updated_by is '最后更新者';
comment on column tools.created_at is '创建时间';
comment on column tools.updated_at is '更新时间';

comment on column tool_submissions.id is '投稿唯一标识';
comment on column tool_submissions.submitted_by is '投稿用户';
comment on column tool_submissions.tool_id is '关联的已有工具（可选）';
comment on column tool_submissions.submitter_email is '投稿者邮箱';
comment on column tool_submissions.payload is '投稿内容（JSON）';
comment on column tool_submissions.status is '审核状态：pending / approved / rejected';
comment on column tool_submissions.reviewer_id is '审核人';
comment on column tool_submissions.review_note is '审核备注';
comment on column tool_submissions.created_at is '投稿时间';
comment on column tool_submissions.reviewed_at is '审核时间';

comment on column audit_logs.id is '日志唯一标识';
comment on column audit_logs.user_id is '操作用户';
comment on column audit_logs.action is '操作类型';
comment on column audit_logs.resource_type is '资源类型';
comment on column audit_logs.resource_id is '资源标识';
comment on column audit_logs.before_data is '变更前数据';
comment on column audit_logs.after_data is '变更后数据';
comment on column audit_logs.created_at is '操作时间';
