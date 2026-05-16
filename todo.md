# 第三方集成中心生产级整改 TODO

> 诊断日期：2026-05-15
> 应用：第三方集成中心
> `app_code`：`integration-center`
> 当前判断：约 100% 生产就绪。已有 Manifest、数据库表、菜单权限、API 注册、前端页面、租户范围隔离、租户侧授权连接入口、OAuth state/callback/token refresh、Webhook 签名幂等、API 调用配额、同步记录配额消费、真实第三方同步批次处理器、受控第三方 API 网关代理边界、关键前端误操作防护、核心对象详情接口、服务端筛选排序时间范围查询、高频查询索引、调用日志保留归档、调用链路追踪字段、OpenAPI schema、生产检查脚本、E2E/性能验证和主要运维文档；调用日志 helper 已沉淀为应用内 recorder/SDK 形态，后续只剩跨应用复用时抽包。
> 整改进度：136 / 136，约 100%。

## 0. 当前已具备

- [x] 独立后端应用目录：`internal/apps/integration_center`
- [x] 独立前端应用目录：`frontend/src/apps/integration-center`
- [x] 应用 Manifest：`internal/apps/integration_center/app.manifest.yaml`
- [x] 菜单、权限、API 权限矩阵、套餐功能点和配额声明已具备基础形态
- [x] 核心业务表已进入 schema baseline 和迁移
- [x] 前端 8 个菜单已接入后端查询接口
- [x] 平台、服务商应用、应用能力、连接、同步任务、配额策略、异常和日志已有基础读写接口
- [x] 已有 service 层基础单测，覆盖部分查询和状态变更 happy path

## 1. P0：上线阻断项

- [x] 明确产品边界：平台治理后台继续 `platform_only=true`，另补租户侧“授权连接 / 同步能力”的 API 或菜单入口。
- [x] 将所有列表接口改为统一分页契约：`items`、`total`、`skip`、`limit`。
- [x] 所有列表接口支持后端筛选、搜索、排序和时间范围过滤，避免只靠前端内存过滤。
- [x] 为 `tenant-connections`、`sync-monitor`、`quota-usages`、`logs` 等租户数据补 tenant scope 查询。
- [x] 所有按 ID 更新连接、任务、异常、能力的接口必须校验当前用户是否有平台范围或对应租户范围。
- [x] 后端 service 接收当前 viewer/user 上下文，不再只传 `context.Context` 和资源 ID。
- [x] 接入后端二次权限校验，确保按钮隐藏不是安全边界。
- [x] 接入套餐功能校验：租户侧创建或使用第三方授权前必须校验 `integration_tenant_authorization`。
- [x] 接入连接数配额校验：创建连接前必须校验 `integration_connection_count`。
- [x] 接入 API 日调用配额校验：第三方接口调用前必须校验并消费 `integration_api_calls_daily`。
- [x] 已补 `integration_api_calls_daily` 服务层消费边界：成功累计 `used_amount`，超限累计 `limited_count` 并拒绝调用；已接入受控 API 网关代理。
- [x] 接入同步日记录配额校验：同步写入前必须校验并消费 `integration_sync_records_daily`。
- [x] 已补 `integration_sync_records_daily` 服务层消费边界：成功累计同步记录数，超限累计 `limited_count` 并拒绝；已接入同步 worker due job 调度。
- [x] 实现租户授权连接创建接口，不能只依赖 seed 数据或平台侧状态变更。
- [x] 实现 OAuth callback 流程，包括 state 校验、授权码交换、token 加密引用、过期时间和授权主体映射。
- [x] OAuth 授权发起已生成一次性 state，并持久化租户、服务商应用、redirect_uri、scope、过期时间和创建人。
- [x] OAuth callback 已校验 app/state 匹配、pending 状态、过期时间和重复消费；授权码交换与 token 引用仍在上一项未完成范围内。
- [x] 实现 token refresh 流程，包括刷新失败、过期、吊销和告警状态机。
- [x] 实现 Webhook 接收入口，包括签名校验、时间戳校验、防重放和幂等键。
- [x] Webhook 事件已落库到 `integration_webhook_events`，并通过服务商应用 + 幂等键唯一约束避免重复写入。
- [x] Webhook 事件必须落库或进入事件队列，并具备失败重试与死信记录。
- [x] 实现同步任务 worker，支持游标、批次、重试上限、失败补偿和任务审计。
- [x] 同步任务 worker 已具备 due job 扫描、自动启动、同步记录配额消费、超限排队、完成态和重试上限失败态；真实第三方数据拉取/写入处理器仍在上一项未完成范围内。
- [x] 实现真实第三方 API client 或网关代理边界，当前连通性检测不能只做本地运行态统计。
- [x] 网关代理已限制 provider base URL、HTTP method、path、请求体大小和可转发 header，并只返回状态与 digest，避免第三方响应正文泄露到底座调用方。
- [x] 凭证只能保存 `credential_ref`，禁止保存明文 secret/token。
- [x] 增加凭证读取、轮换、脱敏展示和日志防泄露策略。
- [x] 所有写操作写入 `created_by`、`updated_by` 或 `handled_by`。
- [x] 所有 Manifest 标记 `audit: true` 的接口必须写入 `audit_log`。
- [x] 异常处理、恢复、忽略必须记录处理人、处理时间和处理备注。
- [x] 删除或停用类能力必须使用逻辑删除、停用或归档，不允许物理删除业务历史。
- [x] 后端错误必须归一化，不泄露 SQL、堆栈、连接串、token 或第三方敏感响应。

## 2. P1：生产可运营项

- [x] 补平台能力新增、编辑、停用接口，目前页面有入口但后端 CRUD 不完整。
- [x] 补服务商应用能力连接的完整配置模型，不只更新 `enabled`。
- [x] 明确 `open_to_tenant`、`default_enabled`、`tenant_configurable` 的真实存储字段与后端更新语义。
- [x] 补连接详情接口，返回授权范围、最终能力、同步任务、配额用量和异常摘要。
- [x] 补应用详情接口，返回凭证引用、回调配置、能力连接、授权租户和运行指标。
- [x] 补平台详情接口，返回平台能力、应用、连接、调用、异常和配置状态。
- [x] 补同步任务详情接口，返回执行日志、批次、游标、错误和重试历史。
- [x] 补调用日志详情接口，只返回摘要、digest、链路 ID 和安全错误信息。
- [x] 接入 `integration_quota_bindings`，支持租户、平台、应用、连接级策略绑定。
- [x] 配额策略增加优先级、适用范围和冲突解析规则。
- [x] 调用日志写入统一 SDK 或中间件，记录 request/response digest，不记录请求响应正文。
- [x] API 网关代理已写入 request/response digest、request_id、tenant_id、platform_id、provider_app_id、connection_id、状态码、耗时和安全错误码。
- [x] 补 API 调用链路中的 request_id、trace_id、tenant_id、platform_id、provider_app_id、connection_id。
- [x] 平台、应用、能力、连接、任务、异常状态收口为明确枚举。
- [x] 非法状态流转必须拒绝，例如已恢复异常不能重新处理，已暂停任务不能重复暂停。
- [x] 数据库补充必要外键到 `tenant`，并检查孤儿数据。
- [x] 数据库补充高频查询索引，覆盖租户、平台、应用、状态、时间范围组合查询。
- [x] 制定日志和调用记录保留策略，避免 `integration_api_call_logs` 无限增长。
- [x] 前端移除固定趋势文案，如“+1 本月”“+12.6%”，改为后端统计。
- [x] 前端移除固定默认选中值，如 `wecom`、`wecom-suite-main`、`policy-tenant-standard`。
- [x] 前端表格统一使用后端分页，当前页无数据时自动回退。
- [x] 前端所有危险操作补二次确认，包括暂停连接、恢复任务、忽略异常、停用策略。
- [x] 前端错误展示映射为可读提示，表单错误定位到字段。
- [x] 前端 loading、empty、error 三态在 8 个菜单中保持一致。
- [x] 操作按钮显示以后端权限和套餐能力为准，不硬编码角色判断。
- [x] 补导出任务真实实现或接入统一任务中心，不能只返回“queued”提示。
- [x] 补全 OpenAPI schema，当前 integration-center 多数路径只有概览式注册。

## 3. P2：交付与长期维护项

- [x] 补应用说明文档：业务边界、平台侧能力、租户侧能力和非目标范围。
- [x] 补 Manifest 字段说明，解释菜单、权限、API、套餐功能点和配额来源。
- [x] 补 API 文档：请求、响应、错误码、权限码、审计点、配额消耗。
- [x] 补 Webhook 文档：事件类型、签名算法、时间戳、防重放、幂等键和重试策略。
- [x] 补同步文档：同步方向、游标、批次、冲突解决、补偿和死信处理。
- [x] 补凭证运维文档：密钥存储、轮换、吊销、脱敏和访问审计。
- [x] 补部署与回滚文档：迁移顺序、回滚影响、初始化验证和数据保护。
- [x] 补告警与 SLO 文档：成功率、延迟、失败率、同步积压、token 过期和超限告警。
- [x] 补验收脚本：Manifest 装载幂等、菜单权限数量、API 权限矩阵、套餐功能点、配额项。
- [x] 补孤儿数据检查：平台、应用、能力、连接、同步、配额、异常、日志引用完整性。
- [x] 补迁移幂等检查，确保空库初始化和重复执行数据数量稳定。
- [x] 补性能压测：调用日志列表、同步任务列表、租户连接列表和总览聚合。
- [x] 补安全测试：签名失败、重放攻击、跨租户访问、无权限写操作、敏感字段泄露。
- [x] 补浏览器 E2E：8 个菜单、分页、筛选、弹窗、抽屉、错误态和移动端布局。

## 4. 测试补齐清单

- [x] Service 单测：tenant scope 正常隔离。
- [x] Service 单测：普通租户不能通过 ID 操作其他租户连接。
- [x] Service 单测：平台管理员可跨租户查询和治理。
- [x] Service 单测：无权限操作返回明确错误。
- [x] Service 单测：套餐未开通时拒绝创建连接。
- [x] Service 单测：连接数配额超限时拒绝创建连接。
- [x] Service 单测：API 调用配额超限时拒绝调用并记录超限次数。
- [x] Service 单测：同步记录配额超限时进入排队或拒绝策略。
- [x] Service 单测：网关代理调用前消费 API 配额，成功后写入 digest 调用日志。
- [x] Service 单测：网关代理日志写入 trace_id、tenant_id、platform_id、provider_app_id 和 connection_id。
- [x] Service 单测：网关代理 API 配额超限时拒绝转发并记录 `limited` 日志。
- [x] Service 单测：Webhook 签名错误拒绝处理。
- [x] Service 单测：Webhook 重放请求拒绝处理。
- [x] Service 单测：Webhook 幂等重复请求不会重复写入业务事件。
- [x] Service 单测：OAuth state 错误拒绝 callback。
- [x] Service 单测：token refresh 失败进入异常状态并写告警。
- [x] Service 单测：同步任务重试上限生效。
- [x] Service 单测：异常处理写入处理人和审计日志。
- [x] Service 单测：连接、同步任务、异常的非法状态流转会被拒绝。
- [x] Service 单测：平台能力新增、编辑、停用会写审计并保持逻辑停用。
- [x] Service 单测：应用能力连接配置会持久化 `open_to_tenant`、`default_enabled`、`tenant_configurable` 和自定义配置。
- [x] Service 单测：服务商应用凭证引用轮换只接受安全引用、返回脱敏值并写入脱敏审计。
- [x] Service 单测：调用日志详情会清理 query secret、credential_ref、token 和 client_secret。
- [x] Service 单测：连接级配额绑定按优先级覆盖套餐限额并阻止超额 API 消费。
- [x] Service 单测：调用日志 CSV 导出会生成真实内容并脱敏 endpoint、token 和 credential_ref。
- [x] Service 单测：平台、服务商应用和应用能力拒绝未知状态枚举。
- [x] Service 单测：Webhook 调用通过统一日志 helper 写入 request digest 且不保存请求正文。
- [x] Service 单测：普通租户执行平台治理写操作返回 `ErrForbidden`。
- [x] Service 单测：列表时间范围过滤和白名单排序生效。
- [x] Service 单测：调用日志保留策略会归档过期日志且列表默认隐藏归档记录。
- [x] Repository 单测：列表分页 total 正确。
- [x] Repository 单测：软删除数据默认不可见。
- [x] Handler 单测：统一响应结构稳定。
- [x] Handler 单测：请求参数错误返回 `CodeBadRequest`。
- [x] Router/Auth 单测：Manifest、运行时鉴权策略、前端权限码一致。
- [x] Migration 测试：空库执行 baseline + migrations 成功。
- [x] Migration 测试：重复执行或重复装载不产生重复数据。
- [x] Frontend 测试：API 失败时不展示过期旧数据。
- [x] Frontend 测试：分页、筛选、空态和错误态可用。
- [x] E2E 测试：平台管理员完成平台、应用、能力、连接、任务、异常、日志主流程。
- [x] E2E 测试：普通租户无法进入平台治理后台。

## 5. 建议推进顺序

- [x] 第一批：API 分页契约、tenant scope、viewer 上下文、审计写入。
- [x] 第二批：租户授权连接、套餐校验、配额校验、凭证引用。
- [x] 第三批：OAuth、Webhook、同步 worker、调用日志 SDK。
- [x] 第三批核心运行时：OAuth、Webhook、同步 worker、API 网关 digest 日志已完成；剩余为统一 SDK/中间件抽象。
- [x] 第四批：前端分页筛选、危险确认、详情接口、配额绑定。
- [x] 第五批：测试矩阵、验收脚本、文档和性能安全验证。

## 6. 验证命令

- [x] `go test ./internal/apps/integration_center/...`
- [x] `go test ./internal/interfaces/http/handlers/...`
- [x] `go test ./internal/application/...`
- [x] `go test ./...`
- [x] `npm --prefix frontend run build`
- [x] `go run ./cmd/migrate`
- [x] `go run ./cmd/verify-bootstrap`
- [x] `API_BASE_URL=http://127.0.0.1:8083 node scripts/check-integration-center-performance.mjs`
- [x] `scripts/check-integration-center-production.sh`
- [x] `scripts/check-migration-idempotency.sh`
- [x] `npm --prefix frontend run test:e2e -- integration-center.spec.ts`
- [x] 浏览器验证 `/integration-center` 及 7 个子菜单。
