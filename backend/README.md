# CyberToolkit Backend

CyberToolkit 的后端服务，当前使用 Go 标准库实现，适合本地联调和后续迁移到 PostgreSQL。

## 当前能力

- 公共接口：首页、分类、标签、工具列表、工具详情、工具提交
- 认证接口：注册、登录、当前用户、登出
- 管理接口：管理员资料、分类管理、工具管理
- 跨域预检支持：处理浏览器 `OPTIONS` 请求
- 内存数据存储：便于快速迭代
- SQL 草案：见 `backend/sql/schema.sql`

## 启动

```bash
cd backend
go run ./cmd/api
```

默认地址：

```text
http://localhost:8080
```

## 环境变量

- `PORT`：监听端口，默认 `8080`
- `CORS_ALLOWED_ORIGINS`：允许的前端来源，逗号分隔
- `ADMIN_EMAIL`：默认管理员邮箱
- `ADMIN_PASSWORD`：默认管理员密码

示例：

```env
PORT=8081
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000
ADMIN_EMAIL=admin@cybertoolkit.local
ADMIN_PASSWORD=admin123456
```

兼容旧变量名：

- `APP_ADDR`
- `APP_ADMIN_EMAIL`
- `APP_ADMIN_PASSWORD`

## 认证说明

当前认证为内存会话模式：

- `POST /api/v1/auth/login` 登录成功后返回随机 `accessToken` 和 `refreshToken`
- 后续请求通过 `Authorization: Bearer <accessToken>` 鉴权
- 会话保存在内存中，服务重启后会失效

当前是过渡实现，后续建议升级为：

- `bcrypt` 密码哈希
- JWT 或数据库 / Redis session
- refresh token 持久化

## 默认演示账号

- `viewer@cybertoolkit.local / viewer123456`
- `editor@cybertoolkit.local / editor123456`
- `admin@cybertoolkit.local / admin123456`

## 说明

- 当前只使用 Go 标准库，没有引入第三方框架
- 当前数据和会话都在内存里，不适合生产
- 数据库设计和接口设计文档位于 `docs/backend/`
