# AI Agent 项目导航

本文件为 AI coding agents 提供项目导航。详细执行规则见 `CLAUDE.md`，运行方式见 `README.md`。

## 项目定位

本项目是 SaaS 多租户管理后台：Gin + PostgreSQL + Redis，前端 Vue 3 + Element Plus。

## 目录结构

- `README.md`：启动方式、端口、Docker、演示账号、故障排查。
- `CLAUDE.md`：Claude Code 专用执行规则。
- `cmd/api`：Gin API 入口。
- `cmd/migrate`：PostgreSQL SQL 迁移入口。
- `internal/bootstrap`：配置、数据库、Redis、路由和种子数据装配。
- `internal/interfaces/http`：handler、middleware、统一响应与 HTTP DTO。
- `internal/application`：用例服务。
- `internal/domain`：领域对象与 repository contract。
- `internal/infrastructure/persistence/postgres`：GORM model、repository、schema baseline 与迁移 SQL。
- `frontend/src/`：Vue3 前端代码，API 在 `api/`，页面在 `views/`，组件在 `components/`，路由在 `router/`。
- `docs/`：原项目文档已复制并更新为当前 Go 技术架构，最终交接见 `docs/FINAL_PARITY_REPORT.md`。
- `docs/tech_design/应用中心接入强制规则.md`：应用 `app_code`、Manifest、独立目录、菜单/权限/套餐同步和装载门禁。
- `docs/tech_design/应用中心.md`：应用中心当前页面、数据模型、统计口径和菜单权限边界。
- `internal/infrastructure/persistence/postgres/schema/current_schema.sql`：当前 PostgreSQL schema baseline。
- `internal/infrastructure/persistence/postgres/migrations/`：后续增量迁移 SQL。

## 数据库与 Docker 初始化

AI 工具拿到代码后，如需构建可运行环境，优先按以下入口判断数据库与 Docker 环境：

- Docker 编排文件：`docker-compose.yml`
- 数据库服务名：`postgres`
- 数据库镜像：PostgreSQL 16
- 默认开发库名：`saas_baseon`
- Docker 内部连接串示例：`host=postgres user=saas password=saas dbname=saas_baseon port=5432 sslmode=disable TimeZone=Asia/Shanghai`
- 本机开发连接串示例：`host=127.0.0.1 user=saas password=saas dbname=saas_baseon port=5432 sslmode=disable TimeZone=Asia/Shanghai`
- Redis 服务名：`redis`
- SQL baseline：`internal/infrastructure/persistence/postgres/schema/current_schema.sql`
- 增量迁移目录：`internal/infrastructure/persistence/postgres/migrations/`

开发环境可使用 `docker compose up --build` 启动 Docker PostgreSQL、Redis、API 和前端 Web 服务。

生产或准生产环境不得依赖开发默认密码、默认 `JWT_SECRET` 或 `CORS_ORIGINS=*`。生产数据库初始化应优先使用 `go run ./cmd/migrate` 执行当前 schema baseline 与 `internal/infrastructure/persistence/postgres/migrations/` 下的版本化增量 SQL。

## 后端

后端接口、分页、错误码、tenant 约束必须遵守 `CLAUDE.md`。涉及租户内数据访问时，必须带 tenant 约束。

全站业务数据不允许物理删除；删除必须按 `CLAUDE.md` 统一做逻辑删除、停用或归档，并保留历史与审计可追溯性。

后端规范见 `CLAUDE.md` 的“5. 后端规范”。新增功能默认按 model -> repository -> domain/application service -> dto/handler -> router 的方向落位，并同步补 service 或 handler 单测。

## 前端

前端使用 Vue 3 Composition API 与 `<script setup>`。历史全局模块 API 调用集中在 `src/api`，新应用私有 API 放在 `src/apps/{app_code}/api.ts`。列表页面优先使用 `NeuroAgentListPage`。权限与菜单以后端权限数据为准。

## Agent 执行规则

新增业务模块前，先阅读 `docs/tech_design/业务开发标准.md`，确认业务对象、租户归属、权限、套餐、配额、数据权限、审计和测试准入要求。

新增应用或业务模块前，必须先阅读 `docs/tech_design/应用中心接入强制规则.md`。新应用必须声明全局唯一 `app_code`、Manifest、独立后端目录 `internal/apps/{app_slug}` 和前端目录 `frontend/src/apps/{app_code}`；`系统管理`、`系统监控` 也必须作为内置应用登记并拥有自己的 `app_code`。

新增可操作功能时的链路：**应用定义 / Manifest 装载（app_code） → 菜单与权限定义 → 套餐功能映射 → 租户菜单（含覆盖） → 角色权限**。不要只改页面或只绑角色；后端权限码、套餐能力、租户菜单覆盖和角色权限必须保持一致。

应用装载必须按 `app_code` 同步生成或更新底座菜单、角色权限资源、API 权限矩阵、套餐中心功能点/配额和租户菜单入口。Manifest、目录结构、权限、菜单、套餐、API 或 `app_code` 任一项不符合规范时，必须拒绝装载，不允许生成半成品数据。

菜单管理生成的能力进入套餐时，必须始终保持应用 -> 目录 -> 菜单 -> 操作树；seed、服务重启或 Manifest 重装载不得覆盖人工移出套餐、租户覆盖、套餐配置和角色授权。

修改代码前先理解现有结构，优先复用已有模式。

涉及接口、权限、分页、tenant 上下文的改动，必须同步检查前后端一致性。

不得提交密钥、token、生产连接串或真实敏感数据。

如需运行、端口、Docker、演示账号或故障排查信息，先看 `README.md`。

如果当前 Agent 支持读取 `CLAUDE.md`，必须同时遵守其中的详细规则。
