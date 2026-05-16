# 第三方集成中心

第三方集成中心是合并部署内置连接器应用，`app_code=integration-center`。它用于平台侧统一治理接入平台、服务商应用、应用能力连接、租户连接实例、同步任务、配额限流、异常监控和调用日志。

新增企业微信、钉钉、飞书、CRM、ERP、电商平台等第三方系统时，必须先阅读并遵守 `docs/tech_design/第三方系统接入标准.md`。本文档说明集成中心自身能力边界；第三方平台接入的命名、凭证、授权、Webhook、同步、网关、租户绑定和验收门禁以该标准为准。

## 接入方式

- 后端目录：`internal/apps/integration_center`
- 前端目录：`frontend/src/apps/integration-center`
- Manifest：`internal/apps/integration_center/app.manifest.yaml`
- 应用类型：`CONNECTOR_APP`
- 部署方式：`MERGED`
- 客户端：`PC_WEB`、`API_ONLY`

## 后台菜单

Manifest 声明平台后台管理菜单：

- 总览：`/integration-center`
- 接入平台：`/integration-center/platforms`
- 集成工作台：`/integration-center/workspace`
- 租户连接：`/integration-center/tenant-connections`
- 同步监控：`/integration-center/sync-monitor`
- 配额与限流：`/integration-center/quota`
- 异常监控：`/integration-center/alerts`
- 调用日志：`/integration-center/logs`

上述菜单是平台治理入口，`platform_only=true`，不进入租户套餐中心。

## 套餐与配额

套餐中心功能点由 Manifest 的 `package_features` 声明：

- `integration_tenant_authorization`：第三方集成授权连接
- `integration_data_sync`：第三方数据同步

配额由 Manifest 的 `quotas` 声明：

- `integration_connection_count`：第三方连接实例数
- `integration_api_calls_daily`：第三方接口日调用量
- `integration_sync_records_daily`：第三方同步日记录数

平台专属治理菜单不得作为租户套餐功能点。租户可购买能力只来自 Manifest 的 `package_features`。

## 权限与 API

受保护 API 均绑定 Manifest 权限码：

- `GET /api/integration-center/overview` -> `/integration-center`
- `GET /api/integration-center/platforms` -> `/integration-center/platforms`
- `POST /api/integration-center/platforms` -> `integration_center:platform_manage`
- `PUT /api/integration-center/platforms/{code}` -> `integration_center:platform_manage`
- `GET /api/integration-center/workspace` -> `/integration-center/workspace`
- `GET /api/integration-center/tenant-connections` -> `/integration-center/tenant-connections`
- `GET /api/integration-center/sync-monitor` -> `/integration-center/sync-monitor`
- `GET /api/integration-center/quota` -> `/integration-center/quota`
- `GET /api/integration-center/alerts` -> `/integration-center/alerts`
- `GET /api/integration-center/logs` -> `/integration-center/logs`

写操作权限预留：

- `integration_center:platform_manage`
- `integration_center:app_manage`
- `integration_center:connection_manage`
- `integration_center:quota_manage`

## 初始化

生产初始化通过当前 schema baseline 和增量迁移执行。`000052_integration_center_manifest_seed` 将 `integration-center` 从规划状态升级为上线状态，并按 Manifest 基线补齐应用中心资产、菜单权限、套餐功能点和配额模板；`000053_integration_center_core_schema` 提供平台、服务商应用、租户连接、能力、同步、配额、异常和调用日志等业务表；`000054_integration_center_seed_data` 提供安全、可重复执行的内置平台、能力、服务商应用、配额策略和演示运行数据。

## 运行时数据读取

后端 `internal/apps/integration_center` 按 handler -> service -> repository 分层读取数据库：

- 总览统计来自 `integration_platforms`、`integration_provider_apps`、`integration_tenant_connections`、`integration_alerts`、`integration_api_call_logs` 和 `integration_sync_jobs`。
- 接入平台、集成工作台、租户连接、同步监控、配额与限流、异常监控和调用日志均有独立查询入口。
- 初始化数据中的 `credential_ref` 只保存外部密钥引用，例如 `vault://integration-center/wecom-suite-standard`，不保存真实第三方凭据。

## 业务边界

平台侧能力：接入平台、服务商应用、平台能力、应用能力、配额策略、异常、调用日志和 Webhook/OAuth 基础设施由平台管理员治理。平台侧菜单保持 `platform_only=true`，不得直接作为租户套餐权益。

租户侧能力：租户可使用已进入套餐的 `integration_tenant_authorization` 和 `integration_data_sync` 能力，创建授权连接、刷新/暂停/恢复连接、消费 API 日调用量与同步记录日配额，并查看自身租户范围内的连接、任务、异常和日志。

非目标范围：本应用不承载第三方业务对象主数据，不直接绕过开放 API 读写独立部署应用数据库，不在底座数据库保存真实 access token、client secret 或 webhook secret。

## Manifest 字段来源

Manifest 是装载标准来源：

- `app_code`：全局唯一应用编码，当前固定为 `integration-center`。
- `menus`：生成平台后台菜单与菜单权限资源；平台治理菜单保持 `platform_only=true`。
- `permissions`：生成角色可授权的后端权限码，写操作必须绑定操作权限。
- `apis`：生成 API 权限矩阵，`audit: true` 的接口必须写审计日志。
- `package_features`：生成套餐中心可售能力，新增租户能力必须从这里声明。
- `quotas`：生成套餐中心配额项，运行时配额消费必须引用这些 quota code。

菜单不得反向自动成为新增套餐能力。新增能力链路必须保持“Manifest -> 菜单/权限/API -> 套餐功能/配额 -> 租户菜单/角色权限”一致。

## API 约定

所有 `/api/integration-center/*` 接口使用统一响应包，错误码遵循全站 HTTP response 规范。平台管理写接口要求平台管理员身份，并写入审计日志；租户连接、配额消费、日志和异常读取接口必须按当前用户 tenant 约束过滤。

关键接口：

- 平台治理：`GET/POST/PUT /platforms`、`GET/POST/PUT /platform-capabilities`。
- 服务商应用：`GET/POST/PUT /provider-apps`、`PATCH /provider-apps/{code}/credential`。
- 租户连接：`GET/POST /tenant-connections`、`POST /tenant-connections/{id}/refresh|pause|resume|retry`。
- 同步任务：`GET /sync-monitor`、`GET /sync-jobs/{id}`、`POST /sync-jobs/{id}/retry|pause|resume`。
- 配额：`GET /quota`、`GET /quota-usages`、`POST/PUT/PATCH /quota-policies`。
- 调用链路：`POST /gateway/invoke`、`POST /webhooks/{provider_app_code}`、`POST /logs/export`。

配额消耗点：

- 网关调用和 API 消费使用 `integration_api_calls_daily`。
- 同步任务和同步 worker 使用 `integration_sync_records_daily`。
- 租户连接创建使用 `integration_connection_count`。

## Webhook

Webhook 入口按服务商应用编码接入，要求请求头携带时间戳、签名和幂等键。签名算法为 HMAC-SHA256，签名内容为 `timestamp + "." + raw_body`，secret 来自安全引用或环境变量，不允许明文入库。

防重放规则：

- 时间戳与服务器时间偏差超过允许窗口时拒绝。
- 相同 `provider_app_id + idempotency_key` 只处理一次。
- payload 只保存摘要和必要结构化内容，异常进入 `integration_webhook_events` 重试队列。

重试规则：

- `received` 或 `retrying` 事件由 worker 拉取处理。
- 失败后写入 `retry_count`、`next_retry_at` 和 `error_message`。
- 超过重试上限进入 `dead_letter`，需要平台侧人工处理或重放。

## 同步

同步任务使用 `integration_sync_jobs`。同步方向由能力配置决定，`cursor_value` 保存外部系统游标或分页令牌，`total_count/success_count/failed_count` 记录批次结果。

任务状态必须在明确枚举内流转：`pending -> running -> completed|failed|queued|retrying|paused`。配额不足时进入 `queued`，并设置下一个日配额窗口；可恢复错误进入 `retrying`，超过上限进入 `failed` 并生成异常。

冲突处理默认以第三方平台事件时间和本地更新时间比较，业务对象落库前必须由具体应用定义字段级策略。底座只负责任务、配额、审计和失败补偿，不替业务应用决定主数据覆盖规则。

## 凭证运维

凭证只保存引用：

- 推荐：`vault://integration/{provider}/{name}`。
- 开发或临时：`env://ENV_KEY`，由运行环境注入。

禁止保存真实 `client_secret`、access token、refresh token 或 webhook secret。凭证展示必须脱敏，审计详情和调用日志必须清理 query secret、token、credential_ref 和 client_secret。

轮换流程：

1. 在外部密钥系统写入新 secret。
2. 调用 `PATCH /api/integration-center/provider-apps/{code}/credential` 更新安全引用。
3. 执行连通性检查和最小 API 调用验证。
4. 保留旧 secret 一个回滚窗口后吊销。

## 部署与回滚

上线顺序：

1. 执行 schema baseline 或 `cmd/migrate` 增量迁移。
2. 装载 Manifest，校验应用、菜单、权限、API、套餐功能和配额。
3. 配置 `JWT_SECRET`、`CORS_ORIGINS`、数据库、Redis 和第三方密钥引用。
4. 启动 API、Webhook worker、同步 worker 和日志归档 worker。
5. 执行连通性检查、Manifest 幂等校验、配额消费 smoke test 和 CSV 导出 smoke test。

回滚原则：

- 业务数据不物理删除，回滚只暂停入口、禁用应用或回滚代码。
- 外键增量使用 `NOT VALID`，历史孤儿数据先用 `integration-center-tenant-orphan-check.sql` 预检，再决定是否 `VALIDATE CONSTRAINT`。
- 回滚迁移前确认新版本写入字段不会被旧版本读取为错误状态。

## 告警与 SLO

建议 SLO：

- API 网关成功率：99.5%。
- Webhook 验签失败率：按服务商应用分组告警，异常突增需排查 secret、时间偏移和攻击流量。
- 同步积压：`pending + retrying + queued` 超过阈值告警。
- token 过期：`token_status=expired` 或刷新失败立即告警。
- 配额超限：`limited_count` 增长按租户、连接和 quota code 聚合告警。
- 调用延迟：P95 超过服务商 SLA 或本地超时阈值告警。

异常处理闭环：所有告警应落入 `integration_alerts`，处理状态从 `open` 到 `processing`，最终为 `resolved` 或 `ignored`，并记录处理人和时间。
