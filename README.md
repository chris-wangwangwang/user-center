# user-center (OIDC 用户中心服务)

基于 Gin + GORM 的 OIDC 用户中心，为分布式服务提供统一的注册、登录、鉴权接口。

## 技术栈
- Gin (HTTP)
- GORM + PostgreSQL（默认 `172.20.90.73:5432`）
- JWT (HS256) Access / Refresh Token
- bcrypt 密码哈希
- Swagger 注解 (swag)

## 目录结构

```text
user-center/
├── main.go                            入口
├── config.toml                        配置（可被 .gitignore 忽略）
├── docs/                              swag 生成
└── internal/
    ├── config/                        配置（单例，TOML 加载）
    ├── database/                      GORM 初始化 + AutoMigrate
    ├── exception/                     统一错误码 + 响应
    ├── util/password/                 bcrypt
    ├── util/jwt/                      JWT Manager（单例）
    ├── model/                         GORM 模型
    ├── repository/                    Repo（单例 + gorm.DB 入参）
    ├── service/                       业务（单例，interface + impl）
    ├── handler/                       gin handler + swagger 注解
    ├── middleware/                    AuthRequired / CORS
    └── router/                        路由装配
```

业务代码采用「interface + 单例 + sync.Once」风格，与你提供的 PulseService 模式一致：

```go
type UserService interface { /* ... */ }

type userServiceImpl struct {
    db        *gorm.DB
    jwtMgr    jwt.Manager
    userRepo  repository.UserRepository
    tokenRepo repository.RefreshTokenRepository
}

var (
    userSvcInstance UserService
    userSvcOnce     sync.Once
)

func GetUserService() UserService {
    userSvcOnce.Do(func() {
        userSvcInstance = &userServiceImpl{ /* ... */ }
    })
    return userSvcInstance
}

func (s *userServiceImpl) Login(req *LoginRequest) (*TokenInfo, *exception.Exception) {
    u, ex := s.userRepo.FindByUsername(s.db, req.Username)
    // ...
}
```

## 启动

```bash
go mod tidy
go run .
```

服务监听 `:8080`，配置读取当前目录的 `config.toml`（找不到时回退到内置默认值）：

```toml
http_port = 8080

db_host = "172.20.90.73"
db_port = 5432
db_user = "postgres"
db_password = "123456"
db_name = "usercenter"
db_sslmode = "disable"

jwt_secret = "change-me-in-production-please"
jwt_issuer = "usercenter"
access_token_ttl_minutes = 60
refresh_token_ttl_hours = 168
```

> 部署时建议把 `config.toml` 加入 `.gitignore`，通过环境注入或独立挂载文件提供。

## 文档

```bash
go install github.com/swag/swag/cmd/swag@v1.16.3
swag init -g main.go -o docs
```

打开：<http://localhost:8080/swagger/index.html>

## 主要接口

| Method | Path                              | 鉴权 | 说明                          |
| ------ | --------------------------------- | ---- | ----------------------------- |
| POST   | /api/v1/auth/register             | 否   | 注册                          |
| POST   | /api/v1/auth/login                | 否   | 登录，返回 access/refresh     |
| POST   | /api/v1/auth/refresh              | 否   | 刷新 access token             |
| POST   | /api/v1/auth/logout               | 是   | 注销（撤销 refresh）          |
| GET    | /api/v1/users/me                  | 是   | 当前用户资料                  |
| PUT    | /api/v1/users/me                  | 是   | 更新资料                      |
| POST   | /api/v1/users/me/password         | 是   | 修改密码                      |
| GET    | /api/v1/users                     | 是   | 用户列表（分页）              |
| POST   | /api/v1/users/{id}/disable        | 是   | 启用/禁用用户                 |
| GET    | /oidc/authorize                   | 是   | 授权码端点（简化）            |
| POST   | /oidc/token                       | 否   | Token 端点（auth_code/refresh_token） |
| GET    | /oidc/userinfo                    | 是   | UserInfo                      |
| GET    | /oidc/jwks                        | 否   | JWKS（HS256 占位）            |
| GET    | /.well-known/openid-configuration | 否   | Discovery                     |

## 响应格式

```json
{
  "code": 0,
  "message": "ok",
  "data": { /* ... */ }
}
```

错误码：`4000` bad request、`4001` unauthorized、`4003` forbidden、`4004` not found、
`4009` conflict、`4010` invalid token、`4011` expired token、`4020` invalid password、
`5000` internal、`5001` database。

## 备注

- 数据库 `uc_users` / `uc_refresh_tokens` / `uc_oauth_clients` / `uc_authorization_codes` 在首次启动时自动迁移。
- 鉴权头：`Authorization: Bearer <access_token>`。
- OIDC 当前实现为简化版（HS256），适合内部服务调用；如需对接公网请改为 RS256 + JWKS 公钥。
