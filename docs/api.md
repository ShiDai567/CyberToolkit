# CyberToolkit API 接口文档

## 概述

**Base URL**: `http://localhost:8080`

**API 前缀**: `/api/v1`

### 认证方式

需要认证的接口需在请求头中携带 Bearer Token：

```
Authorization: Bearer <accessToken>
```

### 响应格式

所有接口返回统一 JSON 格式：

**成功响应**:
```json
{
  "data": { ... },
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 100,
    "totalPages": 5
  }
}
```

`meta` 字段仅在分页接口中返回。

**错误响应**:
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "错误描述",
    "details": null
  }
}
```

### 通用错误码

| HTTP Status | Code | 说明 |
|-------------|------|------|
| 400 | `INVALID_JSON` | 请求体 JSON 格式错误 |
| 401 | `UNAUTHORIZED` | 未认证或 Token 无效 |
| 403 | `FORBIDDEN` | 权限不足（需要 admin 角色） |
| 404 | `NOT_FOUND` | 资源不存在 |
| 405 | `METHOD_NOT_ALLOWED` | HTTP 方法不允许 |
| 422 | `VALIDATION_ERROR` | 请求参数验证失败 |

---

## 1. 公开接口

### 1.1 健康检查

```
GET /api/v1/health
```

**响应** `200`：
```json
{
  "data": {
    "status": "ok"
  }
}
```

---

### 1.2 首页数据

```
GET /api/v1/home
```

返回首页所需的统计数据、精选工具和分类列表。

**响应** `200`：
```json
{
  "data": {
    "stats": {
      "toolCount": 20,
      "categoryCount": 8,
      "featuredCount": 6
    },
    "featuredTools": [
      {
        "id": "nmap",
        "name": "Nmap",
        "description": "Industry-standard network discovery and audit tool.",
        "category": { "id": "network-scanning", "name": "Network Scanning" },
        "difficulty": "intermediate",
        "icon": "Radar",
        "featured": true,
        "tags": ["Port Scan", "Host Discovery"],
        "links": { "website": "https://nmap.org", "github": "https://github.com/nmap/nmap" }
      }
    ],
    "categories": [
      {
        "id": "network-scanning",
        "name": "Network Scanning",
        "description": "Discover hosts, ports, services and exposed assets.",
        "icon": "Radar",
        "toolCount": 3
      }
    ]
  }
}
```

---

### 1.3 分类列表

```
GET /api/v1/categories
```

**查询参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| `includeCounts` | string | `"true"` 时返回每个分类的工具数量 |

**响应** `200`：
```json
{
  "data": [
    {
      "id": "network-scanning",
      "name": "Network Scanning",
      "description": "Discover hosts, ports, services and exposed assets.",
      "icon": "Radar",
      "toolCount": 3
    }
  ]
}
```

---

### 1.4 标签列表

```
GET /api/v1/tags
```

**响应** `200`：
```json
{
  "data": [
    { "id": "port-scan", "name": "Port Scan" },
    { "id": "automation", "name": "Automation" }
  ]
}
```

---

### 1.5 工具列表

```
GET /api/v1/tools
```

**查询参数**:

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `q` | string | - | 关键词搜索（名称、描述） |
| `category` | string | - | 按分类 slug 筛选 |
| `difficulty` | string | - | 按难度筛选：`beginner` / `intermediate` / `advanced` / `expert` |
| `tag` | string | - | 按标签 slug 筛选 |
| `featured` | string | - | `"true"` 仅返回精选工具 |
| `page` | int | 1 | 页码 |
| `pageSize` | int | 20 | 每页数量（最大 100） |
| `sort` | string | - | 排序：`name` / `popular`，默认按创建时间倒序 |

**响应** `200`：
```json
{
  "data": [
    {
      "id": "nmap",
      "name": "Nmap",
      "description": "Industry-standard network discovery and audit tool.",
      "category": { "id": "network-scanning", "name": "Network Scanning" },
      "difficulty": "intermediate",
      "icon": "Radar",
      "featured": true,
      "tags": ["Port Scan", "Host Discovery"],
      "links": { "website": "https://nmap.org", "github": "https://github.com/nmap/nmap" }
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 20,
    "totalPages": 1
  }
}
```

---

### 1.6 工具详情

```
GET /api/v1/tools/{slug}
```

**路径参数**:

| 参数 | 说明 |
|------|------|
| `slug` | 工具的 URL 标识符 |

**响应** `200`：
```json
{
  "data": {
    "id": "nmap",
    "name": "Nmap",
    "description": "Industry-standard network discovery and audit tool.",
    "longDescription": "Nmap supports host discovery, port scanning...",
    "category": {
      "id": "network-scanning",
      "name": "Network Scanning",
      "description": "Discover hosts, ports, services and exposed assets."
    },
    "difficulty": "intermediate",
    "icon": "Radar",
    "featured": true,
    "tags": ["Port Scan", "Host Discovery"],
    "links": [
      { "type": "website", "label": "Official Website", "url": "https://nmap.org" },
      { "type": "github", "label": "Source Repository", "url": "https://github.com/nmap/nmap" }
    ],
    "relatedTools": [
      { "id": "masscan", "name": "Masscan" }
    ]
  }
}
```

**错误** `404`：工具不存在或未发布。

---

### 1.7 提交工具

```
POST /api/v1/submissions
```

**请求体**:
```json
{
  "type": "new_tool",
  "submitterEmail": "user@example.com",
  "payload": {
    "name": "New Tool",
    "description": "A useful security tool",
    "website": "https://example.com"
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `type` | string | 是 | 提交类型，如 `new_tool`、`update`、`report` |
| `submitterEmail` | string | 否 | 提交者邮箱 |
| `payload` | object | 是 | 提交数据（自由格式 JSON） |

**响应** `201`：
```json
{
  "data": {
    "id": "uuid",
    "type": "new_tool",
    "submitterEmail": "user@example.com",
    "payload": { ... },
    "status": "pending",
    "createdAt": "2026-05-19T10:00:00Z"
  }
}
```

---

## 2. 认证接口

### 2.1 用户登录

```
POST /api/v1/auth/login
```

**请求体**:
```json
{
  "email": "admin@cybertoolkit.local",
  "password": "admin123456"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `email` | string | 是 | 用户邮箱 |
| `password` | string | 是 | 用户密码 |

**响应** `200`：
```json
{
  "data": {
    "accessToken": "hex64...",
    "refreshToken": "hex64...",
    "user": {
      "id": "uuid",
      "email": "admin@cybertoolkit.local",
      "displayName": "Admin",
      "role": "admin",
      "isActive": true,
      "createdAt": "2026-05-19T10:00:00Z"
    }
  }
}
```

**错误**:
- `401 INVALID_CREDENTIALS` - 邮箱或密码错误

---

### 2.2 用户注册

```
POST /api/v1/auth/register
```

**请求体**:
```json
{
  "email": "user@example.com",
  "password": "password123",
  "displayName": "Test User"
}
```

| 字段 | 类型 | 必填 | 验证规则 |
|------|------|------|----------|
| `email` | string | 是 | 有效邮箱格式 |
| `password` | string | 是 | 至少 6 个字符 |
| `displayName` | string | 是 | 2-32 个字符 |

**响应** `201`：与登录响应格式相同。新用户默认角色为 `viewer`。

**错误**:
- `409 REGISTER_ERROR` - 邮箱已注册
- `422 VALIDATION_ERROR` - 参数验证失败

---

### 2.3 获取当前用户

```
GET /api/v1/auth/me
```

**需要认证**: 是

**响应** `200`：
```json
{
  "data": {
    "id": "uuid",
    "email": "admin@cybertoolkit.local",
    "displayName": "Admin",
    "role": "admin",
    "isActive": true,
    "createdAt": "2026-05-19T10:00:00Z"
  }
}
```

---

### 2.4 更新用户资料

```
PATCH /api/v1/auth/me
```

**需要认证**: 是

**请求体**:
```json
{
  "displayName": "New Name"
}
```

| 字段 | 类型 | 验证规则 |
|------|------|----------|
| `displayName` | string | 2-32 个字符 |

**响应** `200`：返回更新后的用户信息。

---

### 2.5 用户登出

```
POST /api/v1/auth/logout
```

**需要认证**: 是

**响应** `200`：
```json
{
  "data": { "loggedOut": true }
}
```

---

### 2.6 刷新 Token

```
POST /api/v1/auth/refresh
```

**请求体**:
```json
{
  "refreshToken": "hex64..."
}
```

**响应** `200`：返回新的 accessToken 和 refreshToken。

**错误**:
- `401 INVALID_REFRESH_TOKEN` - refreshToken 无效或已过期

---

### 2.7 修改密码

```
POST /api/v1/auth/password
```

**需要认证**: 是

**请求体**:
```json
{
  "currentPassword": "oldpass",
  "newPassword": "newpass123"
}
```

| 字段 | 类型 | 验证规则 |
|------|------|----------|
| `currentPassword` | string | 必填 |
| `newPassword` | string | 至少 6 个字符，不能与当前密码相同 |

**响应** `200`：
```json
{
  "data": {
    "updated": true,
    "revokedSessions": 2
  }
}
```

修改成功后会自动注销其他设备会话。

---

### 2.8 撤销会话

```
POST /api/v1/auth/sessions/revoke
```

**需要认证**: 是

**请求体**:
```json
{
  "keepCurrent": true
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `keepCurrent` | bool | `true` 保留当前会话，`false` 注销所有会话 |

**响应** `200`：
```json
{
  "data": {
    "revokedSessions": 3,
    "keepCurrent": true
  }
}
```

---

## 3. 管理接口

所有管理接口需要 `admin` 角色的 Bearer Token。非 admin 用户调用将返回 `403 FORBIDDEN`。

### 3.1 管理员登录

```
POST /api/v1/admin/auth/login
```

与 [2.1 用户登录](#21-用户登录) 相同。

---

### 3.2 管理员信息

```
GET /api/v1/admin/me
```

**需要认证**: admin

**响应** `200`：返回当前管理员用户信息，格式同 [2.3](#23-获取当前用户)。

---

### 3.3 仪表盘统计

```
GET /api/v1/admin/stats
```

**需要认证**: admin

**响应** `200`：
```json
{
  "data": {
    "toolCount": 20,
    "publishedToolCount": 15,
    "draftToolCount": 3,
    "categoryCount": 8,
    "userCount": 10,
    "activeUserCount": 8,
    "pendingSubmissionCount": 2,
    "submissionCount": 12
  }
}
```

---

### 3.4 分类管理

#### 3.4.1 分类列表（含隐藏）

```
GET /api/v1/admin/categories
```

**需要认证**: admin

返回所有分类，包括 `isVisible=false` 的隐藏分类。

**响应** `200`：
```json
{
  "data": [
    {
      "id": "uuid",
      "slug": "network-scanning",
      "name": "Network Scanning",
      "description": "Discover hosts, ports, services and exposed assets.",
      "icon": "Radar",
      "sortOrder": 1,
      "isVisible": true,
      "createdAt": "2026-05-19T10:00:00Z",
      "updatedAt": "2026-05-19T10:00:00Z"
    }
  ]
}
```

#### 3.4.2 创建分类

```
POST /api/v1/admin/categories
```

**需要认证**: admin

**请求体**:
```json
{
  "slug": "wireless-security",
  "name": "Wireless Security",
  "description": "Tools for wireless network security testing.",
  "icon": "Wifi",
  "sortOrder": 5,
  "isVisible": true
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `slug` | string | 是 | URL 标识符（唯一） |
| `name` | string | 是 | 分类名称（唯一） |
| `description` | string | 否 | 分类描述 |
| `icon` | string | 否 | Lucide 图标名称 |
| `sortOrder` | int | 否 | 排序权重 |
| `isVisible` | bool | 否 | 是否在前台可见 |

**响应** `201`：返回创建的分类对象。

#### 3.4.3 更新分类

```
PATCH /api/v1/admin/categories/{id}
```

**需要认证**: admin

**路径参数**: `id` - 分类 UUID

**请求体**：与创建相同，所有字段可选（只更新传入的字段）。

**响应** `200`：返回更新后的分类对象。

#### 3.4.4 删除分类（软删除）

```
DELETE /api/v1/admin/categories/{id}
```

**需要认证**: admin

将分类设为不可见（`isVisible=false`），不会真正删除数据。

**响应** `200`：返回更新后的分类对象。

---

### 3.5 工具管理

#### 3.5.1 工具列表（所有状态）

```
GET /api/v1/admin/tools
```

**需要认证**: admin

返回所有状态的工具（包括 draft 和 archived）。

**查询参数**:

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `q` | string | - | 关键词搜索 |
| `status` | string | - | 按状态筛选：`draft` / `published` / `archived` |
| `category` | string | - | 按分类 slug 筛选 |
| `difficulty` | string | - | 按难度筛选 |
| `page` | int | 1 | 页码 |
| `pageSize` | int | 20 | 每页数量（最大 100） |
| `sort` | string | - | 排序方式 |

**响应** `200`：
```json
{
  "data": [
    {
      "id": "uuid",
      "slug": "nmap",
      "name": "Nmap",
      "shortDescription": "Industry-standard network discovery...",
      "longDescription": "Full description...",
      "categoryId": "uuid",
      "difficulty": "intermediate",
      "icon": "Radar",
      "featured": true,
      "status": "published",
      "websiteUrl": "https://nmap.org",
      "githubUrl": "https://github.com/nmap/nmap",
      "viewCount": 1200,
      "favoriteCount": 230,
      "publishedAt": "2026-05-18T10:00:00Z",
      "createdAt": "2026-05-19T10:00:00Z",
      "updatedAt": "2026-05-19T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 20,
    "totalPages": 1
  }
}
```

#### 3.5.2 创建工具

```
POST /api/v1/admin/tools
```

**需要认证**: admin

**请求体**:
```json
{
  "slug": "new-tool",
  "name": "New Tool",
  "shortDescription": "A brief description",
  "longDescription": "A detailed description of the tool...",
  "categoryId": "uuid",
  "difficulty": "beginner",
  "icon": "Shield",
  "featured": false,
  "status": "draft",
  "websiteUrl": "https://example.com",
  "githubUrl": "https://github.com/example",
  "tags": ["Tag1", "Tag2"]
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `slug` | string | 是 | URL 标识符（唯一） |
| `name` | string | 是 | 工具名称（唯一） |
| `shortDescription` | string | 否 | 简短描述（最长 280 字符） |
| `longDescription` | string | 否 | 详细描述 |
| `categoryId` | string | 是 | 所属分类 UUID |
| `difficulty` | string | 否 | 难度：`beginner` / `intermediate` / `advanced` / `expert` |
| `icon` | string | 否 | Lucide 图标名称 |
| `featured` | bool | 否 | 是否精选 |
| `status` | string | 否 | 状态：`draft` / `published` / `archived` |
| `websiteUrl` | string | 是 | 官方网站 URL |
| `githubUrl` | string | 否 | GitHub 仓库 URL |
| `tags` | string[] | 否 | 标签名称列表（自动创建不存在的标签） |

**响应** `201`：返回创建的工具对象。

#### 3.5.3 获取工具详情

```
GET /api/v1/admin/tools/{id}
```

**需要认证**: admin

**路径参数**: `id` - 工具 UUID

**响应** `200`：返回工具完整信息。

#### 3.5.4 更新工具

```
PATCH /api/v1/admin/tools/{id}
```

**需要认证**: admin

**请求体**：与创建相同，所有字段可选。

**响应** `200`：返回更新后的工具对象。

#### 3.5.5 归档工具

```
DELETE /api/v1/admin/tools/{id}
```

**需要认证**: admin

将工具状态设为 `archived`，不会真正删除数据。

**响应** `200`：
```json
{
  "data": { "archived": true }
}
```

---

### 3.6 用户管理

#### 3.6.1 用户列表

```
GET /api/v1/admin/users
```

**需要认证**: admin

**查询参数**:

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码 |
| `pageSize` | int | 20 | 每页数量（最大 100） |

**响应** `200`：
```json
{
  "data": [
    {
      "id": "uuid",
      "email": "user@example.com",
      "displayName": "Test User",
      "role": "viewer",
      "isActive": true,
      "lastLoginAt": "2026-05-19T10:00:00Z",
      "createdAt": "2026-05-19T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 10,
    "totalPages": 1
  }
}
```

#### 3.6.2 修改用户角色

```
PATCH /api/v1/admin/users/{id}
```

**需要认证**: admin

**路径参数**: `id` - 用户 UUID

**请求体**:
```json
{
  "role": "editor"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `role` | string | 是 | `admin` / `editor` / `viewer` |

**响应** `200`：返回更新后的用户信息。

#### 3.6.3 禁用/启用用户

```
DELETE /api/v1/admin/users/{id}
```

**需要认证**: admin

**请求体**:
```json
{
  "active": false
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `active` | bool | `false` 禁用用户（同时注销所有会话），`true` 启用用户 |

**响应** `200`：
```json
{
  "data": { "active": false }
}
```

---

### 3.7 投稿管理

#### 3.7.1 投稿列表

```
GET /api/v1/admin/submissions
```

**需要认证**: admin

**查询参数**:

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `status` | string | - | 按状态筛选：`pending` / `approved` / `rejected` |
| `page` | int | 1 | 页码 |
| `pageSize` | int | 20 | 每页数量（最大 100） |

**响应** `200`：
```json
{
  "data": [
    {
      "id": "uuid",
      "type": "new_tool",
      "submittedBy": "",
      "toolId": "",
      "submitterEmail": "user@example.com",
      "payload": {
        "name": "New Tool",
        "description": "Description"
      },
      "status": "pending",
      "reviewerId": "",
      "reviewNote": "",
      "createdAt": "2026-05-19T10:00:00Z",
      "reviewedAt": null
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 5,
    "totalPages": 1
  }
}
```

#### 3.7.2 审核投稿

```
PATCH /api/v1/admin/submissions/{id}
```

**需要认证**: admin

**路径参数**: `id` - 投稿 UUID

**请求体**:
```json
{
  "status": "approved",
  "note": "Looks good, approved."
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `status` | string | 是 | `approved` 或 `rejected` |
| `note` | string | 否 | 审核备注 |

**响应** `200`：返回更新后的投稿对象（含 reviewerId、reviewNote、reviewedAt）。

---

### 3.8 审计日志

#### 3.8.1 审计日志列表

```
GET /api/v1/admin/audit-logs
```

**需要认证**: admin

**查询参数**:

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `page` | int | 1 | 页码 |
| `pageSize` | int | 20 | 每页数量（最大 100） |

**响应** `200`：
```json
{
  "data": [
    {
      "id": "uuid",
      "userId": "uuid",
      "action": "update",
      "resourceType": "tool",
      "resourceId": "uuid",
      "beforeData": { "name": "Old Name" },
      "afterData": { "name": "New Name" },
      "createdAt": "2026-05-19T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 50,
    "totalPages": 3
  }
}
```

---

## 4. 数据模型

### 4.1 用户 (User)

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 唯一标识 |
| `email` | string | 邮箱（唯一） |
| `displayName` | string | 显示名称 |
| `role` | string | 角色：`admin` / `editor` / `viewer` |
| `isActive` | bool | 是否激活 |
| `lastLoginAt` | datetime | 最后登录时间（可为 null） |
| `createdAt` | datetime | 创建时间 |

### 4.2 分类 (Category)

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 唯一标识 |
| `slug` | string | URL 标识符（唯一） |
| `name` | string | 名称（唯一） |
| `description` | string | 描述 |
| `icon` | string | Lucide 图标名称 |
| `sortOrder` | int | 排序权重 |
| `isVisible` | bool | 是否前台可见 |
| `createdAt` | datetime | 创建时间 |
| `updatedAt` | datetime | 更新时间 |

### 4.3 工具 (Tool)

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 唯一标识 |
| `slug` | string | URL 标识符（唯一） |
| `name` | string | 名称（唯一） |
| `shortDescription` | string | 简短描述（最长 280 字符） |
| `longDescription` | string | 详细描述 |
| `categoryId` | UUID | 所属分类 ID |
| `difficulty` | string | 难度：`beginner` / `intermediate` / `advanced` / `expert` |
| `icon` | string | Lucide 图标名称 |
| `featured` | bool | 是否精选 |
| `status` | string | 状态：`draft` / `published` / `archived` |
| `websiteUrl` | string | 官方网站 |
| `githubUrl` | string | GitHub 仓库 |
| `viewCount` | int | 浏览次数 |
| `favoriteCount` | int | 收藏次数 |
| `publishedAt` | datetime | 发布时间（可为 null） |
| `createdAt` | datetime | 创建时间 |
| `updatedAt` | datetime | 更新时间 |

### 4.4 标签 (Tag)

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 唯一标识 |
| `slug` | string | URL 标识符（唯一） |
| `name` | string | 名称（唯一） |
| `createdAt` | datetime | 创建时间 |

### 4.5 投稿 (Submission)

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 唯一标识 |
| `type` | string | 提交类型 |
| `submittedBy` | UUID | 提交用户 ID（可为空） |
| `toolId` | UUID | 关联工具 ID（可为空） |
| `submitterEmail` | string | 提交者邮箱 |
| `payload` | JSON | 提交数据 |
| `status` | string | 状态：`pending` / `approved` / `rejected` |
| `reviewerId` | UUID | 审核人 ID（可为空） |
| `reviewNote` | string | 审核备注 |
| `createdAt` | datetime | 创建时间 |
| `reviewedAt` | datetime | 审核时间（可为 null） |

### 4.6 审计日志 (AuditLog)

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | UUID | 唯一标识 |
| `userId` | UUID | 操作用户 ID（可为空） |
| `action` | string | 操作类型（如 create、update、delete） |
| `resourceType` | string | 资源类型（如 tool、category、user） |
| `resourceId` | UUID | 资源 ID（可为空） |
| `beforeData` | JSON | 操作前数据 |
| `afterData` | JSON | 操作后数据 |
| `createdAt` | datetime | 操作时间 |

### 4.7 会话 (Session)

| 字段 | 类型 | 说明 |
|------|------|------|
| `accessToken` | string(64) | 访问令牌（主键） |
| `userId` | UUID | 关联用户 ID |
| `refreshToken` | string(64) | 刷新令牌（唯一） |
| `expiresAt` | datetime | 过期时间（24 小时） |
| `createdAt` | datetime | 创建时间 |
