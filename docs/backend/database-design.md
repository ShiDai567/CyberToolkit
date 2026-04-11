# CyberToolkit 数据库设计

本文档描述当前项目采用的关系型数据设计草案，对应实现参考：

- `backend/sql/schema.sql`
- `backend/internal/domain/models.go`

推荐数据库：`PostgreSQL 16+`

## 设计目标

- 支撑工具目录展示、分类筛选、标签筛选、详情页
- 支撑登录注册和后台管理
- 为后续 PostgreSQL 持久化保留结构基础

## 核心实体

- `users`
- `categories`
- `tags`
- `tools`
- `tool_tags`
- `tool_submissions`
- `audit_logs`

## 枚举约定

### user_role

- `admin`
- `editor`
- `viewer`

### tool_status

- `draft`
- `published`
- `archived`

### tool_difficulty

- `beginner`
- `intermediate`
- `advanced`
- `expert`

### submission_status

- `pending`
- `approved`
- `rejected`

## 表设计

### users

用于登录、角色控制和后台管理。

关键字段：

- `email`
- `password_hash`
- `display_name`
- `role`
- `is_active`
- `last_login_at`

说明：

- 当前 Go 内存版已支持 `admin / editor / viewer`
- 正式落库后建议增加唯一索引、登录审计和密码重置字段

### categories

工具分类表。

关键字段：

- `slug`
- `name`
- `description`
- `icon`
- `sort_order`
- `is_visible`

说明：

- 前台默认只展示 `is_visible = true`
- 后台可以看到全部分类

### tags

标签主表。

关键字段：

- `slug`
- `name`

### tools

工具主表。

关键字段：

- `slug`
- `name`
- `short_description`
- `long_description`
- `category_id`
- `difficulty`
- `icon`
- `featured`
- `status`
- `website_url`
- `github_url`
- `view_count`
- `favorite_count`
- `published_at`

说明：

- 当前版本把链接直接放在 `tools` 表中，保留 `website_url` 和 `github_url`
- 这是当前阶段的简化设计，避免过早拆出 `tool_links`
- 如果后续出现文档、下载、演示、视频等多类型链接，再考虑拆成独立表

### tool_tags

工具和标签的多对多关系表。

作用：

- 一个工具可以挂多个标签
- 一个标签可以被多个工具复用
- 保持结构规范，便于筛选、统计和后台管理

### tool_submissions

收集用户提交的新工具或纠错建议。

关键字段：

- `submitted_by`
- `tool_id`
- `submitter_email`
- `payload`
- `status`
- `reviewer_id`
- `review_note`

### audit_logs

后台审计日志。

关键字段：

- `user_id`
- `action`
- `resource_type`
- `resource_id`
- `before_data`
- `after_data`

## 关系说明

- 一个 `category` 对应多个 `tools`
- 一个 `tool` 对应多个 `tags`
- 一个 `tag` 对应多个 `tools`
- 一个 `user` 可以提交多个 `tool_submissions`
- 一个 `user` 可以产生多条 `audit_logs`

## 索引建议

当前 `schema.sql` 已包含这些重点索引：

- `categories(sort_order)`
- `tools(category_id)`
- `tools(status)`
- `tools(featured)`
- `tools(published_at desc)`
- `tools(category_id, status)`
- `tool_submissions(status)`
- `audit_logs(user_id)`
- `audit_logs(resource_type, resource_id)`

## 为什么保留 tool_tags

`tool_tags` 是标准多对多建模，不建议把标签直接塞进 `tools` 表。

原因：

- 一个工具有多个标签
- 一个标签会被多个工具复用
- 需要支持按标签筛选、统计和后台维护

## 为什么当前没有 tool_links

当前版本只稳定支持两种链接：

- 官网 `website_url`
- 源码 `github_url`

对于这个阶段，把它们直接放在 `tools` 表里更简单。等链接类型明显变多，再拆 `tool_links` 更合适。

## 最小可用表集

如果先做 PostgreSQL 持久化，最小闭环是：

- `users`
- `categories`
- `tags`
- `tools`
- `tool_tags`
- `tool_submissions`

`audit_logs` 可以后补，但建议 schema 先保留。
