# 项目工程执行规则

本文件是本项目的工程执行规则。Claude Code 或其他 AI coding agent 在生成、修改、重构、排查代码时，必须优先遵守本文件。

## 0. 先读入口

开始任何任务前，按任务类型读取：

- 运行、端口、Docker、演示账号：`README.md`
- AI 项目导航：`AGENTS.md`
- 总体架构与目录结构：`docs/tech_design/总体技术方案.md`
- 新增应用、Manifest、套餐功能点来源：`docs/tech_design/应用中心接入强制规则.md`、`docs/tech_design/业务开发标准.md`
- 独立部署应用：`docs/tech_design/独立部署应用接入约束.md`
- 数据库初始化与生产 SQL：`README.md`、`cmd/migrate`、`internal/infrastructure/persistence/postgres/schema/current_schema.sql`、`internal/infrastructure/persistence/postgres/migrations/`
- 系统管理需求/技术/数据库文档：`docs/req_design/`、`docs/tech_design/`、`docs/sql_design/`

如果这些文档与当前代码冲突，以当前代码为准，并同步修正文档。

## 1. 安全红线

- 不得提交或暴露密钥、token、数据库密码、生产连接串、私有证书、本机绝对路径、内部堆栈日志、真实敏感数据或第三方凭据。
- 生产配置必须通过环境变量管理，禁止硬编码到源码、SQL 或文档示例中。
- 生产环境必须设置 `APP_ENV=production`、强 `JWT_SECRET`、显式 `CORS_ORIGINS` 和非默认数据库账号。
- 启动配置必须拒绝开发默认 `JWT_SECRET`、`CORS_ORIGINS=*`、默认 `saas/saas` 数据库连接等不安全生产配置。
- 错误信息不得泄露 SQL 原文、内部堆栈、数据库连接信息、密钥、token 或其他敏感实现细节。
- 必须写 raw SQL 时，所有外部输入必须参数化，禁止字符串拼接 SQL。

## 2. 数据库与生产初始化

- 生产或准生产数据库初始化以 `cmd/migrate` 执行的 `internal/infrastructure/persistence/postgres/schema/current_schema.sql` baseline 和 `internal/infrastructure/persistence/postgres/migrations/*.up.sql` 版本化 SQL 为准。
- 开发期 GORM AutoMigrate 和启动补列脚本只作为本地兜底，不得作为生产建库交付方案。
- 数据库结构变更必须同步维护生产 SQL、回滚说明和验证 SQL。
- 迁移脚本禁止清空业务表；历史迁移遇到非空目标表应拒绝执行或显式保护。
- 全站业务数据禁止物理删除。删除默认实现为逻辑删除、停用或归档，并保留历史数据、审计链路、外键可追溯性。
- 列表查询默认排除已逻辑删除数据；详情、审计、历史报表如需追溯已删除数据，必须显式说明。
- 允许物理删除的例外仅限无业务历史意义的数据，例如会话/token、缓存、限流 key、上传失败临时文件、导入中间表、租户覆盖值恢复默认等，并必须在 service/crud 注释或文档中说明原因。
- 逻辑删除记录如占用唯一键，必须释放唯一键。优先使用统一 `tombstone_unique_value` 处理主体编码、用户工号/手机号、角色编码、套餐编码等。

## 3. 多租户、权限与套餐

- 所有业务域必须严格执行 tenant 隔离，禁止跨 tenant 读写。
- 租户内业务数据查询、创建、更新、删除必须带当前 `tenant_id` 约束。
- 创建业务数据时必须写入当前 tenant、company、department 上下文；更新和删除必须限制在当前 tenant 内。
- 数据访问控制禁止仅依赖前端过滤。数据权限必须在 SQL 或 ORM 查询层生效，禁止先查出数据后再过滤权限。
- 业务表默认应包含 `tenant_id`、`company_id`、`department_id`。系统级全局表可以例外，但必须有明确理由。
- 系统层级为：`tenant -> company -> department -> user -> role -> permission`。
- 用户与组织关系必须通过多对多关系表实现。角色与权限关系必须通过 `user_role`、`role_permission` 等关系表实现，禁止将角色、权限列表硬编码到 user 主表。
- `data_scope` 仅允许 `ALL`、`ORG`、`ORG_SUB`、`SELF`、`CUSTOM`。涉及数据范围的查询必须转换为明确 SQL/ORM 过滤条件。

新增可操作业务功能必须按以下顺序接入：

```text
应用定义 / Manifest 装载（app_code）
  -> 菜单与权限定义
  -> 套餐功能映射
  -> 租户菜单（含覆盖）
  -> 角色权限
  -> 前端页面与按钮
  -> 后端接口与 service 校验
  -> 数据权限
  -> 审计日志
```

规则：

- 应用中心是新增应用的上游入口。新增应用必须声明全局唯一 `app_code`、Manifest、独立后端目录 `internal/apps/{app_slug}` 和前端目录 `frontend/src/apps/{app_code}`。
- `系统管理`、`系统监控` 均必须作为内置应用登记，应用来源为 `BUILTIN`，并分别拥有自己的 `app_code`。
- 内置应用也必须输出标准 `app.manifest.yaml`，包括 `app-center`、`system-management`、`system-monitor`。
- 应用装载必须按 `app_code` 同步生成或更新底座菜单、角色权限资源、API 权限矩阵、套餐中心功能点/配额和租户菜单入口。
- Manifest、目录结构、权限、菜单、套餐、API 或 `app_code` 任一项不符合规范时，必须拒绝装载；不得生成半成品数据、不得静默跳过、不得要求后续手工补录。
- 菜单管理生成的目录、菜单、操作必须保持三层结构，并能追溯到所属 `app_code`。
- 平台专属能力不得进入租户套餐功能池；仅主体能力不得作为租户可购买套餐能力。
- 菜单管理是目录、菜单、操作三层结构来源；应用 Manifest 是应用归属和装载来源；套餐中心只承载租户可购买能力与配额。不得在前端或套餐页维护第二套功能树。
- 套餐中心功能点必须由 Manifest 的 `package_features` 声明，配额必须由 Manifest 的 `quotas` 声明。菜单自动生成套餐功能点只允许作为历史兼容兜底；新增应用、内置应用和后续业务模块不得依赖菜单自动生成作为标准来源。
- `include_in_package=false` 的能力不得进入套餐中心；如果后续需要售卖，应走套餐外订阅、应用单独订阅或明确的非售卖治理。
- 如果应用声明 `deployment_mode=STANDALONE`，必须读取并遵守 `docs/tech_design/独立部署应用接入约束.md`。独立部署应用不得直接读写底座数据库、Redis、文件目录或内部表结构，不得复用底座内部 JWT secret、数据库账号或生产连接串。
- 独立部署应用必须通过 Manifest、开放 API、应用凭证、scope、Webhook 签名、数据同步幂等键或网关代理接入底座；没有通讯模式、应用凭证、租户上下文、幂等键和审计点时，不得实现外部调用、回调或同步任务。
- 服务重启、seed 重跑、Manifest 重装载不得覆盖人工移出套餐中心、租户覆盖、套餐列配置或角色授权。
- 统一使用 `permission` 作为权限来源，覆盖 menu、operation、API、data scope。权限标识必须稳定，并与前端一致。
- 租户菜单运行时按：平台专属过滤 -> 套餐功能过滤 -> 角色权限过滤 -> 租户覆盖 计算最终结果。
- 套餐控制租户购买能力，角色权限控制用户能做什么，数据权限控制用户能看什么数据。三者不得混淆。
- 前端只负责体验隐藏，后端必须通过 `_require_menu_path`、`_require_operation`、`require_feature_access`、配额校验等入口二次校验。
- 主体创建与套餐配置必须分步：主体基础资料、默认公司、首个超级管理员和管理员角色由主体创建流程保存；套餐订阅、订阅起止时间、用户/公司等席位配额覆盖由套餐配置流程保存。
- 运行时配额来源优先级为 `tenant_quota_override > saas_plan_quota`，不得再把 `tenant.max_users`、`tenant.max_companies`、`tenant.expire_date` 作为业务判断来源。

## 4. API 契约

- 所有接口必须返回统一结构，禁止返回裸数据。
- 统一响应字段必须包含 `code`、`message`、`data`。
- 成功响应：`code=0`、`message="ok"`、`data` 为业务数据；无业务数据时 `data` 可以为 `null`。
- 失败响应：`code` 必须是非 0 业务错误码，`message` 必须可读，`data` 通常为 `null`。
- 业务错误码必须定义在统一枚举、常量或错误码模块中，禁止在 router、handler、service 中散落魔法数字。
- HTTP 状态码与业务错误码必须同时明确。`4xx` 用于参数、认证、权限、资源不存在；`5xx` 用于服务端错误或依赖服务异常。
- 列表接口必须返回分页结构，禁止直接返回数组。
- 请求分页参数统一使用 `skip` 和 `limit`。
- 分页响应的 `data` 至少包含 `items` 和 `total`。已有页面需要时保留 `skip`、`limit`。
- 创建成功返回创建后的完整对象；更新成功返回更新后的最新对象；删除成功返回可预测结果。
- 接口字段命名必须稳定。新增字段必须向后兼容。删除、改名、改类型、改变字段语义必须同步迁移说明、前端 API、页面调用和文档。

## 5. 后端规范

后端分层优先采用：

```text
router -> dto -> application service -> domain/repository contract -> postgres repository -> GORM model
```

### 5.1 路由层

`cmd/api` 和 `internal/interfaces/http/handlers/*.go` 只做 HTTP 边界：

- 在 `internal/bootstrap/router.go` 注册路径、方法和 handler。
- 读取当前用户、权限依赖、请求对象、Header、Query、Path、File。
- 提取 IP、User-Agent、Authorization 等 HTTP 上下文。
- 调用 `internal/application` 服务或当前 handler 内已存在的兼容流程。
- 用 `internal/interfaces/http/response` 统一包装返回结果。
- 不得在 router 中写数据库查询、业务判断、审计日志、CSV/文件解析、Redis 读写、密码校验、token 签发、数据范围过滤、复杂响应组装。

### 5.2 服务层

`internal/application/**/*_service.go` 承载业务流程；历史尚未下沉的接口可在 handler 中保持兼容，但新增稳定业务规则应继续下沉：

- 服务函数命名优先使用领域语义，例如 `LoginForRequest`、`CreateUserForViewer`、`TenantFilePath`。
- 涉及当前用户权限或租户上下文的参数统一使用 `viewer` 或 `user`，并从该对象读取 `tenant_id`。
- 负责业务规则、跨 repository 编排、事务边界、异常转换、审计日志、外部依赖降级、文件/CSV/Redis 等非 HTTP 逻辑。
- 可以返回明确的 HTTP DTO 或简单领域结果，但不得让 handler 直接透出裸 GORM model。
- 返回错误必须安全、可读，并由 HTTP 层转换为统一响应；不得泄露 SQL、堆栈、连接串或内部实现。
- 新增可测试分支必须同步新增 `internal/application`、`internal/domain` 或 handler 的 Go 单测。

### 5.3 数据访问层

`internal/domain/*/*_repository.go` 定义 repository contract，`internal/infrastructure/persistence/postgres/repositories/*.go` 实现数据访问与可复用持久化操作：

- 所有租户内查询、更新、删除必须带 `tenant_id` 过滤。
- 删除业务数据不得调用 GORM 物理删除或直接 `DELETE`，应写入 `deleted_at`、`is_deleted`、`status=deleted`，或按领域语义改为停用/归档。
- 新增支持删除的业务表默认预留逻辑删除字段，列表查询默认过滤逻辑删除数据。
- 逻辑删除后必须处理唯一键释放。
- repository 函数不理解 HTTP 请求，不读取 Header、Request、Authorization。
- repository 可以执行单表或明确的持久化操作，但多步骤事务、审计日志、权限判断、异常消息编排应由 service 组织。
- 平台管理员跨租户逻辑必须在 service 中显式表达，不得为了方便在 repository 中绕过 tenant。

### 5.4 DTO 与模型

- `internal/interfaces/http/dto/*.go` 定义请求、响应、分页契约。
- `internal/infrastructure/persistence/postgres/models/*.go` 只负责 GORM 映射，不承载业务流程。
- 请求与响应校验必须放在 DTO、handler 绑定或 service 边界中，避免直接返回 GORM model、裸数组或临时拼接结构。
- 数据库会话必须通过 Gin 依赖注入获取，禁止使用全局可变 session。
- 数据库完整性异常必须捕获，并转换为 API 层可理解错误。
- 公开函数、service 接口、repository 接口必须使用明确类型。
- 非必要不写 raw SQL，优先使用 ORM 查询。
- 必须提供健康检查端点，例如 `/health`。

## 6. 前端规范

- 新页面和新组件默认使用 Vue 3 Composition API 与 `<script setup>`。除维护历史代码或明确兼容原因外，不新写 Options API。
- API 调用必须集中在 `frontend/src/api`，禁止在 view、page、业务组件中直接写 fetch、axios 原始请求。
- 鉴权与权限判断必须统一放在路由守卫层或统一权限工具中。
- 菜单必须根据后端权限数据动态生成，最终可见菜单必须受后端权限数据约束。
- 禁止在组件中硬编码角色可见性、菜单可见性或权限判断。
- 权限标识必须稳定，并与后端 permission 定义一致，不得在前端随意发明权限标识。
- 列表页面优先使用项目内统一列表组件 `NeuroAgentListPage`。如组件无法满足，应优先扩展统一组件能力，避免重复裸写 `el-table`。
- 页面状态与可复用逻辑必须分离，可复用逻辑应提取为 composable。
- 数据页必须统一处理 loading、empty、error 三态，不得只在 console 中打印错误。
- 表单提交前必须完成校验，后端校验错误应映射为用户可理解的表单反馈或页面反馈。
- UI 状态与请求中必须保持 tenant、company 上下文一致。切换 tenant、company、department 后，应同步刷新菜单、权限、路由、页面数据、缓存状态和选择器状态。
- 危险操作必须二次确认，并给出明确成功或失败反馈。
- 全局样式、按钮、表格、卡片、布局间距、主题变量应集中管理。禁止在多个组件 `<style>` 中重复编写全局覆盖规则。
- 敏感 token 存储必须遵循项目鉴权策略，不得在 URL、console 日志、错误提示中暴露敏感信息。

## 7. 测试与验证

- 每个新增 service 或安全分支必须有对应单测。
- 优先覆盖 tenant 隔离、权限拒绝、错误分支、边界输入、外部依赖异常、文件/CSV/Redis 安全行为。
- 套餐主链路改动必须覆盖主体创建、套餐绑定、能力上下文、菜单/按钮裁剪、后端套餐校验、配额校验、套餐变更后即时生效。
- 数据权限改动必须覆盖越权访问、普通用户范围、平台管理员范围。
- 测试 fixture 使用替身 repository 或测试数据库时，必须避免误用生产 PostgreSQL 连接。

提交前验证：

- 代码改动：至少运行相关后端测试或前端构建；无法运行必须说明原因。
- 文档改动：至少运行 `git diff --check`。
- 底座级发布前至少执行：

```bash
docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'
cd frontend && npm run build
git diff --check
```

## 8. Git 与提交

- 每完成一个独立、用户可见、可运行或可联调的功能后，必须在同一轮对话结束前执行一次 Git 提交。
- 提交前必须在项目根目录执行 `git status`，只 `git add` 本次任务相关文件。
- 禁止提交无关文件、临时文件、本机路径、密钥、token、数据库密码、生产连接串或其他敏感信息。
- 提交信息必须是完整句子，说明做了什么以及为什么。
- 如果项目尚未初始化 Git，应先初始化仓库，添加合适的 `.gitignore`，再首次提交。
- 如果环境无法提交，必须说明原因，并建议用户本地手动提交或使用其他方式备份。

## 9. 全局一致性与优先级

前端调用、后端接口、权限标识、分页字段、错误结构必须保持一致。

接口契约变化时，必须同步更新：

- 后端 DTO
- 前端 api 层
- 页面调用逻辑
- 权限配置
- 套餐功能映射
- 必要文档

新增功能时，优先复用既有分层结构、DTO、application service、repository、model、composable、permission 模型、API 响应工具、分页工具、错误处理工具、统一列表组件和全局样式体系。

不要为了快速实现而绕过统一架构。如果现有结构不足以支撑新功能，应优先抽象公共能力，而不是复制粘贴业务代码。

规则冲突时，按以下优先级处理：

1. 安全与敏感信息保护
2. 多租户隔离与数据权限
3. API 契约稳定性
4. 后端分层与数据库规范
5. 前端权限、交互与样式规范
6. Git 备份提交规范

任何情况下都不得为了快速实现功能而破坏 tenant 隔离、权限控制或泄露敏感信息。
