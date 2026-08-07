# Chirpy

一个基于 Go 标准库构建的 HTTP JSON API 服务，是 [boot.dev](https://www.boot.dev) **Learn HTTP Servers in Go** 课程的实践项目。

课程地址：<https://www.boot.dev/courses/learn-http-servers-golang>

## 项目简介

Chirpy 是一个类 Twitter 的微型博客后端服务。项目从零开始使用 Go 标准库 `net/http` 构建，涵盖了路由、中间件、JSON 序列化、PostgreSQL 数据持久化、JWT 认证、授权、Webhook 集成等后端核心主题。项目采用 Go 社区惯用的按业务域分包架构，通过显式依赖注入组织各模块间的协作。

## 课程章节与项目功能映射

| 课程章节 | 涉及功能 |
| --- | --- |
| 1. Servers | 基于 `net/http` 构建 HTTP 服务器，静态文件服务 |
| 2. Routing | Go 1.22+ 模式路由（`GET /api/chirps/{chirpID}`） |
| 3. Architecture | 按业务域分包 + 显式依赖注入的分层架构 |
| 4. JSON | 请求/响应 JSON 编解码，`platform.WriteJSON` 工具函数 |
| 5. Storage | PostgreSQL + sqlc 代码生成 + Goose 数据库迁移 |
| 6. Authentication | JWT 签发与校验，Argon2id 密码哈希，刷新令牌 |
| 7. Authorization | 基于 JWT 的资源所有权校验（ chirp 删除权限） |
| 8. Webhooks | Polka Webhook 集成（Chirpy Red 会员激活） |
| 9. Documentation | 项目文档与 API 参考 |

## 技术栈

| 类别 | 技术 |
| --- | --- |
| 语言 | Go 1.26.4 |
| HTTP 服务器 | Go 标准库 `net/http`（Go 1.22+ 模式路由） |
| 数据库 | PostgreSQL |
| 数据库驱动 | `github.com/lib/pq` |
| SQL → Go 代码生成 | [sqlc](https://sqlc.dev) |
| 数据库迁移 | [Goose](https://github.com/pressly/goose) |
| 认证 | `github.com/golang-jwt/jwt/v5`（JWT） |
| 密码哈希 | `github.com/alexedwards/argon2id`（Argon2id） |
| 环境配置 | `github.com/joho/godotenv` |
| UUID | `github.com/google/uuid` |

## 项目结构

```
.
├── cmd/server/main.go          # 应用入口：配置加载、依赖注入、路由注册、启动服务
├── internal/
│   ├── platform/http.go        # 跨业务 HTTP 基础设施（WriteJSON、ErrorResponse）
│   ├── chirps/                 # Chirp 资源：创建、列表、详情、删除
│   ├── users/                  # 用户资源：注册、更新、登录
│   ├── auth/                   # 认证工具：JWT、密码哈希、刷新令牌、Webhook 鉴权
│   ├── webhooks/               # Polka Webhook 处理
│   ├── admin/                  # 管理后台：健康检查、访问统计、用户重置
│   └── database/               # sqlc 生成的数据访问代码（不手动修改）
├── sql/
│   ├── schema/                 # Goose 数据库迁移脚本（001–005）
│   └── queries/                # sqlc 查询定义（chirps / users / refresh_token）
├── assets/                     # 静态资源
├── index.html                  # 前端页面
├── sqlc.yaml                   # sqlc 配置
├── go.mod / go.sum             # Go 模块依赖
└── .env                        # 环境变量（不纳入版本控制）
```

### 架构约定

- **单一入口**：`cmd/server/main.go` 负责配置加载、依赖装配、路由注册与服务启动。
- **显式依赖注入**：各业务包定义自己的 `Handler` 结构体（持有 `*database.Queries` 及所需密钥），通过 `NewHandler(...)` 在 `main.go` 中注入，取代全局单例。
- **Handler 命名**：使用动词或资源动作命名（`Create`、`List`、`Login`、`Refresh`），不使用 `XxxHandler` 后缀。
- **文件组织**：`model.go` 定义业务实体与数据转换逻辑，`handlers.go` 实现 HTTP 处理器。
- **数据访问层**：`internal/database` 由 sqlc 自动生成，业务层仅消费生成类型，不直接拼接 SQL。

## API 端点

### 健康 & 管理

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/healthz` | 健康检查 |
| `GET` | `/admin/metrics` | 获取请求访问计数 |
| `POST` | `/admin/resetHits` | 重置访问计数器 |
| `POST` | `/admin/reset` | 重置用户数据（仅 dev 平台） |

### 用户 & 认证

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/users` | 注册用户 |
| `PUT` | `/api/users` | 更新用户信息（需 JWT 认证） |
| `POST` | `/api/login` | 用户登录，签发 JWT + 刷新令牌 |
| `POST` | `/api/refresh` | 使用刷新令牌换取新 JWT |
| `POST` | `/api/revoke` | 吊销刷新令牌 |

### Chirps

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/chirps` | 创建 chirp（需 JWT 认证） |
| `GET` | `/api/chirps` | 获取 chirp 列表（支持 `author_id` 过滤、`sort` 排序） |
| `GET` | `/api/chirps/{chirpID}` | 获取单个 chirp |
| `DELETE` | `/api/chirps/{chirpID}` | 删除 chirp（需 JWT 认证 + 所有权校验） |

### Webhook

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/polka/webhooks` | Polka Webhook 回调（Chirpy Red 会员激活，需 API Key 鉴权） |

## 数据库设计

| 表 | 说明 |
| --- | --- |
| `users` | 用户信息（id, email, hashed_password, is_chirpy_red） |
| `chirps` | 短消息，外键关联 `users(id)`，级联删除 |
| `refresh_tokens` | 刷新令牌，外键关联 `users(id)`，含过期与吊销时间 |

迁移脚本位于 `sql/schema/`，使用 Goose 按序号递增执行：

| 迁移文件 | 内容 |
| --- | --- |
| `001_user.sql` | 创建 `users` 表 |
| `002_chirp.sql` | 创建 `chirps` 表 |
| `003_users_hashed_password.sql` | `users` 表新增 `hashed_password` 字段 |
| `004_refresh_tokens.sql` | 创建 `refresh_tokens` 表 |
| `005_users_chirpy_red.sql` | `users` 表新增 `is_chirpy_red` 字段 |

## 快速开始

### 前置条件

- [Go 1.22+](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/) 运行实例
- [sqlc](https://docs.sqlc.dev/en/latest/install.html)（仅在修改 SQL 查询后需重新生成代码）
- [Goose](https://github.com/pressly/goose#install)（仅在执行数据库迁移时需要）

### 配置

在项目根目录创建 `.env` 文件：

```env
DB_URL=postgres://user:password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
SECRET_KEY=your-jwt-secret-key
POLKA_KEY=your-polka-api-key
```

### 数据库迁移

```bash
# 使用 Goose 执行迁移
goose -dir sql/schema postgres "postgres://user:password@localhost:5432/chirpy?sslmode=disable" up
```

### 运行

```bash
go run ./cmd/server
```

服务启动后监听 `:8080`，健康检查端点：<http://localhost:8080/api/healthz>

### 重新生成数据访问代码（修改 SQL 查询后）

```bash
sqlc generate
```

## License

本项目为 [boot.dev](https://www.boot.dev) 课程学习产物，仅供学习用途。
