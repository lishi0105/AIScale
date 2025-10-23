```csharp
.
├── cmd/
│   └── foodapp/
│       └── main.go           # 程序入口（从这里启动）
├── internal/                 # 仅本项目私有的代码（Go 工具会限制外部 import）
│   ├── server/               # HTTP 层（Gin）
│   │   ├── middleware/       # 日志、恢复、鉴权、限流等
│   │   ├── handler/          # 按业务拆包：account、food（只做入参出参、调用 service）
│   │   └── server.go         # 组装 Gin 引擎、注册路由
│   ├── domain/               # 领域模型（POJO/实体），无框架依赖
│   │   ├── account/
│   │   │   └── model.go
│   │   └── food/
│   │       └── model.go
│   ├── repository/           # 数据访问接口 & GORM 实现（可替换为 mock/其他 DB）
│   │   ├── account/
│   │   │   ├── repo.go       # interface（可选） + gorm 实现
│   │   │   └── repo_gorm.go
│   │   └── food/
│   │       ├── repo.go
│   │       └── repo_gorm.go
│   ├── service/              # 业务服务（应用层编排、事务、跨仓储组合）
│   │   ├── account/
│   │   │   └── service.go
│   │   └── food/
│   │       └── service.go
│   ├── storage/              # 技术设施：DB、Cache、MQ 等
│   │   └── db/
│   │       ├── gorm.go       # GORM 初始化、连接池
│   │       └── migrate.go    # 迁移（goose/sqlc/自研都行）
│   └── app/
│       └── wiring.go         # 依赖装配（把 logger/db/service/router 组装起来）
├── pkg/                      # 可复用的通用包（允许外部项目 import）
│   └── logger/
│       └── logger.go         # zap + lumberjack 封装
├── configs/                  # 配置（yaml/json/env 示例）
├── scripts/                  # devops 脚本（migrate、lint、gen 等）
├── go.mod
└── .gitignore
```

## docker构建
```bash
# 构建
docker build -t foodapp-run:v1.0 .
```