# SaaS Baseon Go 多租户底座

本项目是 SaaS 多租户管理后台的 Go 重建版，后端使用 Gin、GORM、PostgreSQL、Redis，前端使用 Vue 3 与 Element Plus。当前目标不是新增业务产品能力，而是在保留原系统产品语义的前提下，形成可继续演进的企业级 SaaS Base。

## 技术栈

- Go 1.22+，Docker 构建环境使用 Go 1.23。
- Gin 作为 HTTP 框架。
- GORM 作为数据访问层，PostgreSQL 作为主数据库。
- Redis 用于登录会话、验证码、限流和缓存。
- Vue 3、Vite、Element Plus 构建管理后台。
- 后端按 `router -> dto -> application service -> domain/repository contract -> postgres repository -> GORM model` 分层推进。
- 数据库结构以 SQL baseline 与版本化 migration 为生产依据，开发期 AutoMigrate 仅作为本地兜底。

## 快速启动

复制开发环境变量：

```bash
cp .env.example .env
```

启动完整开发环境：

```bash
docker compose up --build
```

健康检查：

```bash
curl http://127.0.0.1:8081/health
```

开发访问地址：

- 后端 API：`http://127.0.0.1:8081`
- 前端开发服务：`http://127.0.0.1:5173`
- Docker Web 服务：`http://127.0.0.1:8082`
- OpenAPI：`http://127.0.0.1:8081/openapi.json`

## 数据库初始化

生产或准生产环境必须通过迁移命令初始化数据库：

```bash
go run ./cmd/migrate
```

初始化或部署后校验基础数据：

```bash
go run ./cmd/verify-bootstrap
```

当前 PostgreSQL schema 基线：

```text
internal/infrastructure/persistence/postgres/schema/current_schema.sql
```

后续结构和数据修复只允许追加版本化迁移：

```text
internal/infrastructure/persistence/postgres/migrations/*.up.sql
internal/infrastructure/persistence/postgres/migrations/*.down.sql
```

如需模拟生产方式启动，先执行迁移，再关闭开发 AutoMigrate：

```bash
DB_AUTO_MIGRATE=false docker compose up --build
```

## 演示账号

开发环境默认演示账号：

```text
E10001 / 112233
E10100 / 112233
```

这些密码只允许用于本地开发。生产环境必须设置非默认 `BOOTSTRAP_ADMIN_PASSWORD`，并配置强 `JWT_SECRET`、显式 `CORS_ORIGINS` 和非默认数据库账号密码。

## 目录结构

```text
cmd/api                         Gin API 入口
cmd/migrate                     PostgreSQL 迁移入口
cmd/verify-bootstrap            初始化数据校验入口
internal/bootstrap              配置、数据库、Redis、路由和种子数据装配
internal/interfaces/http        handler、middleware、DTO、统一响应
internal/application            用例服务和业务编排
internal/domain                 领域对象、策略和 repository contract
internal/infrastructure         PostgreSQL、Redis、存储等基础设施实现
frontend                        Vue 3 管理后台
scripts                         发布、清洁检查和交付脚本
tests                           集成与契约测试
docs                            需求、技术、数据库、测试、安全和交接文档
```

## 核心能力

- 多主体、多租户隔离。
- 用户、角色、权限、菜单、按钮和数据权限。
- 套餐、功能点、配额、订阅和租户覆盖。
- 组织架构、岗位、业务单元。
- 数据字典、系统参数、品牌配置。
- 登录日志、操作日志、系统监控。
- 文件上传、下载、删除和用户导入导出。
- Redis session 优先的登录态，JWT 仅作为显式开发降级策略。
- 权限 fail-closed、生产配置自检、安全响应头、文件安全校验和发布包清洁检查。

## API 返回契约

所有接口统一返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

列表接口必须返回分页结构，至少包含 `items` 与 `total`。请求分页参数统一使用 `skip` 和 `limit`。

## 菜单、权限与套餐链路

新增或调整可操作能力必须保持以下链路一致：

```text
应用定义 / Manifest 装载（app_code）
  -> 菜单与权限定义
  -> 套餐功能映射
  -> 租户菜单覆盖
  -> 角色权限
  -> 前端页面与按钮
  -> 后端接口与 service 校验
  -> 数据权限
  -> 审计日志
```

新增应用必须先声明 `app_code` 和 Manifest，并按 `docs/tech_design/应用中心接入强制规则.md` 同步底座菜单、角色权限资源、API 权限矩阵、套餐中心功能点/配额和租户菜单入口。不符合规范的应用必须拒绝装载。

平台专属能力不得进入租户套餐中心；租户菜单和租户操作是否进入套餐中心，以 `internal/domain/permissioncatalog` 中的统一目录策略为准。

## 常用验证命令

后端测试：

```bash
go test ./...
```

前端构建：

```bash
cd frontend
npm run build
```

前端安全扫描：

```bash
cd frontend
npm run security:check
```

发布包检查：

```bash
make check-release
```

文档与 Markdown 修改后至少执行：

```bash
git diff --check
```

## 文档入口

- [文档索引](docs/README.md)
- [项目总体介绍](docs/项目总体介绍.md)
- [当前实现总览](docs/当前实现总览.md)
- [总体技术方案](docs/tech_design/总体技术方案.md)
- [应用中心接入强制规则](docs/tech_design/应用中心接入强制规则.md)
- [应用中心](docs/tech_design/应用中心.md)
- [业务开发标准](docs/tech_design/业务开发标准.md)
- [模块接入准入清单](docs/tech_design/模块接入准入清单.md)
- [最终交接报告](docs/FINAL_PARITY_REPORT.md)
- [安全策略](docs/security/token-storage.md)

## 开发约束

- 业务数据默认逻辑删除、停用或归档，不允许随意物理删除。
- 租户内业务数据必须强制携带 `tenant_id` 约束。
- 前端隐藏按钮不等于授权，后端必须按菜单、操作、套餐和数据权限兜底校验。
- seed 只能做缺失初始化和必要纠偏，不得覆盖用户已经修改的业务数据。
- 生产环境不得使用开发默认密钥、默认数据库密码、`CORS_ORIGINS=*` 或 `DB_AUTO_MIGRATE=true`。
