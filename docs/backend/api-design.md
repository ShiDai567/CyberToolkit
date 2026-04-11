# CyberToolkit API 文档

本文档描述当前 Go 后端已经实现的 API，而不是未来规划接口。

## 基础约定

- Base URL：`/api/v1`
- 数据格式：`application/json`
- 成功响应：

```json
{
  "data": {}
}
```

- 带分页时：

```json
{
  "data": [],
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 4,
    "totalPages": 1
  }
}
```

- 错误响应：

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "type and payload are required",
    "details": null
  }
}
```

## 认证方式

需要登录的接口使用：

```http
Authorization: Bearer <accessToken>
```

当前角色：

- `viewer`
- `editor`
- `admin`

当前实现里：

- 普通认证接口校验“是否已登录”
- 管理接口校验“是否为 admin”

## CORS

后端已处理浏览器预检请求：

- 支持 `OPTIONS`
- 可通过 `CORS_ALLOWED_ORIGINS` 配置允许来源

## 公共接口

### GET /api/v1/health

健康检查。

响应示例：

```json
{
  "data": {
    "status": "ok"
  }
}
```

### GET /api/v1/home

首页聚合数据。

返回：

- `stats`
- `featuredTools`
- `categories`

### GET /api/v1/categories

分类列表。

查询参数：

- `includeCounts=true` 时返回每个分类的 `toolCount`

### GET /api/v1/tags

标签列表。

### GET /api/v1/tools

工具列表。

查询参数：

- `q`
- `category`
- `difficulty`
- `tag`
- `featured`
- `page`
- `pageSize`
- `sort`

`sort` 当前支持：

- `name`
- `popular`
- 其他值默认按创建时间倒序

响应示例：

```json
{
  "data": [
    {
      "id": "nmap",
      "name": "Nmap",
      "description": "Industry-standard network discovery and audit tool.",
      "category": {
        "id": "network-scanning",
        "name": "Network Scanning"
      },
      "difficulty": "intermediate",
      "icon": "Radar",
      "featured": true,
      "tags": ["Host Discovery", "Port Scan"],
      "links": {
        "website": "https://nmap.org",
        "github": "https://github.com/nmap/nmap"
      }
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 1,
    "totalPages": 1
  }
}
```

### GET /api/v1/tools/{slug}

工具详情。

响应字段：

- `id`
- `name`
- `description`
- `longDescription`
- `category`
- `difficulty`
- `icon`
- `featured`
- `tags`
- `links`
- `relatedTools`

### POST /api/v1/submissions

提交工具建议。

请求体示例：

```json
{
  "type": "new_tool",
  "submitterEmail": "alice@example.com",
  "payload": {
    "name": "Example Tool",
    "category": "osint",
    "description": "Example description",
    "website": "https://example.com"
  }
}
```

成功返回 `201 Created`。

## 认证接口

### POST /api/v1/auth/register

注册普通用户。

请求体：

```json
{
  "email": "alice@example.com",
  "password": "secret123",
  "displayName": "Alice"
}
```

成功返回：

```json
{
  "data": {
    "accessToken": "random-access-token",
    "refreshToken": "random-refresh-token",
    "user": {
      "id": "user_4",
      "email": "alice@example.com",
      "displayName": "Alice",
      "role": "viewer",
      "createdAt": "2026-04-12T08:00:00Z"
    }
  }
}
```

### POST /api/v1/auth/login

通用登录接口，适用于所有角色。

请求体：

```json
{
  "email": "admin@cybertoolkit.local",
  "password": "admin123456"
}
```

返回结构与注册接口一致。

### GET /api/v1/auth/me

获取当前登录用户。

需要 `Authorization` 头。

### POST /api/v1/auth/logout

注销当前会话。

成功响应：

```json
{
  "data": {
    "loggedOut": true
  }
}
```

## 管理接口

以下接口当前都要求 `admin` 角色。

### GET /api/v1/admin/me

获取当前管理员资料。

### GET /api/v1/admin/categories

获取后台分类列表，包含隐藏分类。

### POST /api/v1/admin/categories

创建分类。

请求体：

```json
{
  "slug": "cloud-security",
  "name": "Cloud Security",
  "description": "Cloud tooling and security operations.",
  "icon": "Cloud",
  "sortOrder": 9,
  "isVisible": true
}
```

### PATCH /api/v1/admin/categories/{id}

更新分类。

### DELETE /api/v1/admin/categories/{id}

当前实现不是物理删除，而是把 `isVisible` 置为 `false`。

### GET /api/v1/admin/tools

后台工具列表。

查询参数：

- `q`
- `status`
- `category`
- `difficulty`
- `page`
- `pageSize`
- `sort`

### POST /api/v1/admin/tools

创建工具。

请求体：

```json
{
  "slug": "amass",
  "name": "Amass",
  "shortDescription": "Attack surface mapping and asset discovery.",
  "longDescription": "Long description here.",
  "categoryId": "cat_4",
  "difficulty": "intermediate",
  "icon": "Globe",
  "featured": false,
  "status": "published",
  "websiteUrl": "https://owasp.org/www-project-amass/",
  "githubUrl": "https://github.com/owasp-amass/amass",
  "tags": ["Asset Discovery", "OSINT"]
}
```

### GET /api/v1/admin/tools/{id}

按内部 ID 获取工具。

### PATCH /api/v1/admin/tools/{id}

更新工具。`tags` 传入后会整体替换。

### DELETE /api/v1/admin/tools/{id}

当前实现为归档，返回：

```json
{
  "data": {
    "archived": true
  }
}
```

## 当前限制

- 用户、工具、分类、会话都保存在内存中
- `refreshToken` 目前只是返回给前端，未实现刷新接口
- 管理接口当前只开放给 `admin`，还没有细分到 `editor`
- 密码哈希当前使用简化实现，后续应替换为 `bcrypt`
