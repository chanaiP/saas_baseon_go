# 最终等价对比报告

## 范围

- 原项目：`/Users/Shared/aiproject/Baseon/Saas_Baseon`
- Go 项目：`/Users/chen.ai/project/saas_baseon_go`
- 目标：不改变产品需求和功能逻辑，只把后端技术栈调整为 Gin + GORM + PostgreSQL，并保留原前端体验和文档资产。

## P1 路由对比

- 原项目 FastAPI 业务路由：134 条。
- Go 版 Gin 业务路由：143 条。
- 原项目路由缺失：0 条。
- Go 版额外路由：9 条，均为兼容别名或技术支撑端点，不新增产品功能。

Go 版额外路由：

- `GET /api/business-units/org-mappings`
- `GET /api/dict-types/by-code`
- `GET /api/params`
- `GET /api/params/:key`
- `GET /api/permission-menu-bundles`
- `GET /api/public`
- `POST /api/params`
- `POST /api/tenants/with-package`
- `PUT /api/permissions/menu-data-perm-mode`

结论：原项目已暴露的业务 API 在 Go 版全部覆盖，额外路由只服务于兼容旧前端路径、公共品牌兜底、开发标准参数接口或现有“带套餐创建主体”流程。

## P2 前端页面与 API 封装对比

- `frontend/src/api` 文件集合一致：auth、tenant、plan、organization、position、businessUnit、user、role、permission、dict、param、logs、monitor、tenantBranding、http、types。
- `frontend/src/views` 主页面集合一致：登录、首页、主体、套餐、组织、岗位、业务单元、用户、角色、权限配置、菜单、字典、参数、操作日志、登录日志、监控健康/服务/服务器/任务/缓存、个人资料、开发中心。
- 前端静态 endpoint 引用一致，仍使用原路径体系。

结论：前端未引入新的产品页面或业务入口，API 封装维持原功能面。

## P3 后端行为与测试覆盖对比

已完成 A-O 组等价补齐，覆盖：

- 认证、主体/订阅、品牌、套餐功能和配额。
- 组织、岗位、业务单元、用户、角色、权限、菜单与数据权限。
- 字典、参数、登录日志、操作日志。
- 文件上传/下载/删除、批量导入导出。
- 监控、Redis cache key、登录限流、安全守卫。

测试补齐重点：

- 原 Python 测试中的服务层边界在 Go 版转化为 focused Go 单元测试。
- 删除引用保护、CSV 安全、文件 ID、BOM、数据权限 payload、套餐能力矩阵、订阅状态、随机密码策略等关键行为均有测试。
- 每个 O 组完成后均执行 `gofmt` 与 `go test ./...`。

结论：后端产品行为已按原测试意图完成等价覆盖，并补充了 Go 版实现特有的迁移、OpenAPI、schema baseline 和安全 helper 测试。

## P4 数据库模型、迁移与种子数据对比

- 数据库从原 MySQL/SQLAlchemy 迁移到 PostgreSQL/GORM。
- 核心实体完整覆盖：Tenant、Subscription、Plan、Feature、Quota、OrgNode、Position、BusinessUnit、User、Role、Permission、Dict、Param、LoginLog、AuditLog。
- 复合唯一约束、软删除 tombstone、租户覆盖表、业务单元映射、数据权限业务单元范围均已补齐。
- `current_schema.sql` 与 GORM 模型保持一致。
- 种子数据补齐原项目默认菜单、权限、功能、配额、套餐能力矩阵、字典和系统参数。
- 迁移器已修正 `schema_migrations` 写入 schema 问题，并用 PostgreSQL 16 空库验证 baseline 可应用。

结论：数据库结构和初始化数据满足原产品逻辑，技术栈调整为 PostgreSQL 后保持功能等价。

## P6 产品范围确认

本轮新增内容只属于：

- 技术栈替换所需基础设施。
- 原功能等价补齐。
- 兼容路径。
- 测试与文档增强。

未新增新的产品需求、业务流程或额外功能入口。

## P5 验证结果

已通过：

- Go 单元测试：`docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'`
- 前端构建：`npm run build`
- 业务 API 回归：`BASICP_API_PORT=8081 npx playwright test e2e/api-regression.spec.ts`，2 passed。
- 页面 E2E：`npx playwright test e2e/framework.spec.ts e2e/system-pages.spec.ts`，30 passed。

当前可访问服务：

- Go API：`http://127.0.0.1:8081`
- Go Web：`http://127.0.0.1:8082`
- 本地 Playwright/Vite 验证：`http://localhost:5174`
