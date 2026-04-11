# CyberToolkit API 文档

## 目标

API 分两层：

- 公共接口：前端站点直接调用
- 管理接口：后台录入、编辑、审核使用

基础约定：

- Base URL：`/api/v1`
- 数据格式：`application/json`
- 时间字段：ISO 8601 UTC
- 分页：`page` + `pageSize`

## 认证与权限

### 公共接口

- 无需登录

### 管理接口

- 推荐使用 `JWT + Refresh Token`
- 请求头：

```http
Authorization: Bearer <token>
```

角色权限建议：

- `admin`：全部权限
- `editor`：分类、工具、标签的增删改查
- `viewer`：只读后台数据

## 通用响应格式

### 成功

```json
{
  "data": {},
  "meta": {
    "requestId": "req_123"
  }
}
```

### 分页

```json
{
  "data": [],
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 135,
    "totalPages": 7
  }
}
```

### 错误

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid query parameter",
    "details": {
      "field": "pageSize"
    }
  }
}
```

## 公共接口

### 1. 获取首页数据

`GET /api/v1/home`

用途：

- 首页精选工具
- 分类卡片
- 统计信息

响应示例：

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
        "description": "网络发现与安全审计的行业标准工具",
        "category": "network-scanning",
        "difficulty": "intermediate",
        "icon": "Radar",
        "featured": true,
        "tags": ["端口扫描", "主机发现"],
        "links": {
          "website": "https://nmap.org",
          "github": "https://github.com/nmap/nmap"
        }
      }
    ],
    "categories": [
      {
        "id": "network-scanning",
        "name": "网络扫描",
        "description": "发现主机、端口、服务与基础网络暴露面",
        "icon": "Radar",
        "toolCount": 3
      }
    ]
  }
}
```

### 2. 分类列表

`GET /api/v1/categories`

查询参数：

- `includeCounts`：`true | false`

响应字段：

- `id`
- `name`
- `description`
- `icon`
- `toolCount`

### 3. 工具列表

`GET /api/v1/tools`

查询参数：

- `q`：关键词
- `category`：分类 slug
- `difficulty`：难度
- `tag`：标签 slug
- `featured`：是否精选
- `page`
- `pageSize`
- `sort`

`sort` 可选值：

- `latest`
- `name`
- `popular`

示例：

```http
GET /api/v1/tools?q=scan&category=network-scanning&difficulty=beginner&page=1&pageSize=12
```

响应示例：

```json
{
  "data": [
    {
      "id": "rustscan",
      "name": "RustScan",
      "description": "结合高速扫描与 Nmap 深度探测的现代工具",
      "category": {
        "id": "network-scanning",
        "name": "网络扫描"
      },
      "difficulty": "beginner",
      "icon": "ScanLine",
      "featured": false,
      "tags": ["Rust", "快速扫描", "Nmap 集成"],
      "links": {
        "website": "https://rustscan.github.io/RustScan/",
        "github": "https://github.com/RustScan/RustScan"
      }
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 12,
    "total": 1,
    "totalPages": 1
  }
}
```

### 4. 工具详情

`GET /api/v1/tools/{slug}`

路径参数：

- `slug`：工具唯一标识

响应示例：

```json
{
  "data": {
    "id": "nmap",
    "name": "Nmap",
    "description": "网络发现与安全审计的行业标准工具",
    "longDescription": "Nmap 是经典的网络扫描工具，可用于主机发现、端口扫描、服务识别、操作系统探测和脚本化安全审计。",
    "category": {
      "id": "network-scanning",
      "name": "网络扫描",
      "description": "发现主机、端口、服务与基础网络暴露面"
    },
    "difficulty": "intermediate",
    "icon": "Radar",
    "featured": true,
    "tags": ["端口扫描", "主机发现", "服务识别", "操作系统探测"],
    "links": [
      {
        "type": "website",
        "label": "官方网站",
        "url": "https://nmap.org"
      },
      {
        "type": "github",
        "label": "源码仓库",
        "url": "https://github.com/nmap/nmap"
      }
    ],
    "relatedTools": [
      {
        "id": "masscan",
        "name": "Masscan"
      }
    ]
  }
}
```

### 5. 标签列表

`GET /api/v1/tags`

用途：

- 前端筛选条件
- 后台表单标签候选项

### 6. 提交工具建议

`POST /api/v1/submissions`

用途：

- 游客或登录用户提交新工具
- 提交纠错建议

请求体：

```json
{
  "type": "new_tool",
  "submitterEmail": "alice@example.com",
  "payload": {
    "name": "Example Tool",
    "category": "osint",
    "description": "示例描述",
    "website": "https://example.com",
    "github": "https://github.com/example/repo",
    "tags": ["情报收集", "自动化"]
  }
}
```

响应：

- `201 Created`

## 管理接口

以下接口默认需要 `editor` 及以上权限。

### 1. 登录

`POST /api/v1/admin/auth/login`

请求体：

```json
{
  "email": "admin@example.com",
  "password": "secret"
}
```

响应：

```json
{
  "data": {
    "accessToken": "jwt-access-token",
    "refreshToken": "jwt-refresh-token",
    "user": {
      "id": "f0cc6c7d-6c4c-4d7d-8f7f-0d8360f7f20d",
      "email": "admin@example.com",
      "displayName": "Admin",
      "role": "admin"
    }
  }
}
```

### 2. 刷新令牌

`POST /api/v1/admin/auth/refresh`

### 3. 获取后台当前用户

`GET /api/v1/admin/me`

### 4. 后台分类列表

`GET /api/v1/admin/categories`

支持：

- 查看隐藏分类
- 查看排序字段

### 5. 创建分类

`POST /api/v1/admin/categories`

请求体：

```json
{
  "slug": "cloud-security",
  "name": "云安全",
  "description": "云环境审计与防护相关工具",
  "icon": "Cloud",
  "sortOrder": 9,
  "isVisible": true
}
```

### 6. 更新分类

`PATCH /api/v1/admin/categories/{id}`

### 7. 删除分类

`DELETE /api/v1/admin/categories/{id}`

删除策略建议：

- 若分类下仍有工具，禁止物理删除
- 可改为 `isVisible=false`

### 8. 后台工具列表

`GET /api/v1/admin/tools`

查询参数：

- `status`
- `categoryId`
- `featured`
- `q`
- `page`
- `pageSize`

### 9. 创建工具

`POST /api/v1/admin/tools`

请求体：

```json
{
  "slug": "nmap",
  "name": "Nmap",
  "shortDescription": "网络发现与安全审计的行业标准工具",
  "longDescription": "详细描述",
  "categoryId": "2d3fe08e-5796-4f67-a4c1-a7df761f7637",
  "difficulty": "intermediate",
  "icon": "Radar",
  "featured": true,
  "status": "published",
  "links": [
    {
      "type": "website",
      "label": "官方网站",
      "url": "https://nmap.org"
    },
    {
      "type": "github",
      "label": "源码仓库",
      "url": "https://github.com/nmap/nmap"
    }
  ],
  "tags": ["端口扫描", "主机发现", "服务识别"]
}
```

### 10. 获取后台工具详情

`GET /api/v1/admin/tools/{id}`

### 11. 更新工具

`PATCH /api/v1/admin/tools/{id}`

建议：

- 支持部分字段更新
- 标签与链接按整组覆盖或显式 patch，二选一，不要混用

### 12. 删除工具

`DELETE /api/v1/admin/tools/{id}`

删除策略建议：

- 默认软删除，改 `status=archived`
- 真正物理删除只允许 `admin`

### 13. 标签管理

接口：

- `GET /api/v1/admin/tags`
- `POST /api/v1/admin/tags`
- `PATCH /api/v1/admin/tags/{id}`
- `DELETE /api/v1/admin/tags/{id}`

### 14. 提交审核列表

`GET /api/v1/admin/submissions`

查询参数：

- `status`
- `page`
- `pageSize`

### 15. 审核提交

`POST /api/v1/admin/submissions/{id}/review`

请求体：

```json
{
  "action": "approve",
  "reviewNote": "信息完整，已通过"
}
```

`action` 可选值：

- `approve`
- `reject`

### 16. 审计日志

`GET /api/v1/admin/audit-logs`

## 状态码约定

- `200 OK`
- `201 Created`
- `204 No Content`
- `400 Bad Request`
- `401 Unauthorized`
- `403 Forbidden`
- `404 Not Found`
- `409 Conflict`
- `422 Unprocessable Entity`
- `500 Internal Server Error`

## 校验规则建议

### 分类

- `slug`：小写字母、数字、中划线
- `name`：1 到 100 字符

### 工具

- `slug`：唯一
- `name`：唯一
- `shortDescription`：最多 280 字符
- `difficulty`：必须在枚举内
- `links[].url`：必须是合法 URL
- `tags`：去重后保存

## 前端对接建议

当前前端页面最少需要这几个接口：

- 首页：`GET /api/v1/home`
- 工具列表：`GET /api/v1/tools`
- 工具详情：`GET /api/v1/tools/{slug}`
- 分类列表：`GET /api/v1/categories`

这样可以直接替换掉当前 `frontend/src/data/tools.ts` 的本地静态数据。

## 版本迭代建议

### V1

- 分类
- 工具
- 标签
- 首页
- 搜索

### V2

- 后台登录
- 工具提交与审核
- 审计日志

### V3

- 收藏
- 评论
- 评分
- 搜索增强

如果你下一步要继续，我可以直接把这份文档继续落成：

- `Prisma schema`
- `OpenAPI 3.1 yaml`
- `NestJS` 或 `FastAPI` 的接口骨架
