# AI 能力中心生产级诊断 TODO

诊断时间：2026-05-15  
范围：仅 `ai-capability-center` 应用，包括 `internal/apps/ai_capability_center`、`frontend/src/apps/ai-capability-center`、AI 相关表、Manifest 与应用文档。

## 总结

当前 AI 能力中心还不是完整生产级可用状态。它已经具备供应商、模型、场景、路由、策略、用量日志和平台页壳，并已打通 OpenAI-compatible chat/text_generation、responses、embeddings、images 的真实调用骨架；但 demo 清理脚本、内容记录完整治理、Manifest 重装载保护、文档和前端可信表达仍需继续收口。

最关键的问题不是单个页面样式，而是：真实供应商配置和健康状态尚未恢复，历史 seed/demo 数据需要被长期隔离；`Invoke` 已补上 provider request id、重试、超时记录、平台/租户门禁和 OpenAI-compatible 主要协议族，但完整治理、文档和可观测表达还要继续补齐。

## 已验证事实

- `ai_usage_records` 共 18 条，18 条都是 `request_id like 'seed%'`，且 `request_params.channel` 只有 `demo-history-seed` 和 `demo-seed`。
- 当前库里没有 2026-05-15 的真实调用记录；总览今日指标为 0，近 7 天趋势来自历史 demo seed。
- `ai_provider_apis` 中 26 个启用 API 的 `health_status` 全部为 `error`。
- 多数 API 异常原因是未配置 API Key / Key Alias 环境变量；部分是占位 Endpoint，例如 Azure 的 `{resource}`；OpenAI 相关接口存在 TLS 证书校验失败记录。
- `marketing-center / campaign_image_generate` 和 `product-center / product_copy_generate` 当前绑定的基础路由已逻辑删除：`image-marketing-default`、`chat-cost-first` 的 `deleted_at` 非空。
- `internal/apps/ai_capability_center/services/service.go` 的 `Invoke` 会校验场景、策略、路由、配额、限流和价格，并已通过 provider adapter 支持 OpenAI-compatible chat/text_generation 真实 HTTP 调用。
- `Invoke` 的 chat/text_generation 成功路径会返回标准化 `Data`、按供应商 usage 回填计费、记录真实响应 hash；其他协议仍待补齐。
- 文档 `docs/apps/ai-capability-center.md` 已恢复“调用日志”菜单、真实调用链路、内容记录级别和 demo 数据隔离说明。
- 系统参数 `ai.gateway.content_record_level` 已存在，值为 `1`，但数据库显示 `tenant_editable=true`，与 seed 代码里 `TenantEditable: false` 的意图不一致，需要确认参数覆写逻辑。

## P0 阻断项

- [x] 实现真实 AI Gateway 调用链路
  - [x] 在 `Invoke` 中接入 provider adapter，不再只写日志。
  - [x] 先支持 OpenAI-compatible chat/text_generation 真实调用。
  - [x] 继续补齐 OpenAI-compatible responses、embeddings、images。
  - [x] 真实记录请求开始时间、结束时间、供应商 HTTP 状态、错误码、延迟、重试次数。
  - [x] `Data` 必须返回模型供应商响应的标准化结果，不能继续为空对象。
  - [x] `ResponseHash` 必须基于真实响应或标准化响应计算。

- [x] 修复供应商 API 连通性
  - [x] 明确 Key Alias 到环境变量 / KMS / 密钥服务的解析规则。
  - [x] 对未配置密钥的供应商，Gateway 调用必须明确失败且不能请求供应商或产生成功计费。
  - [x] 对未配置密钥的供应商，不能在 UI 上给用户“可用”的暗示。
  - [x] 占位 Endpoint 必须标记为模板或禁用，不能参与可执行路由。
  - [x] 解决容器内 TLS 证书链问题，至少给出 CA 配置和失败诊断。
  - [x] 连通性检查结果需要区分“未配置”“网络失败”“鉴权失败”“协议不兼容”“供应商返回错误”。

- [x] 清理 demo 用量和生产数据口径
  - [x] seed/demo 用量不能混入生产指标；需要增加 `data_source`、`is_demo` 或独立 demo tenant 标记。
  - [x] 总览、趋势、排行榜、调用日志默认必须排除 demo 数据。
  - [x] 如果保留演示数据，页面必须明确标识“演示数据”，不能伪装成真实生产调用；当前生产口径默认不展示 demo，并提供清理脚本。
  - [x] 提供一键清理 demo AI 用量的脚本或迁移说明。

- [x] 修复场景到路由的失效绑定
  - [x] 所有 `ai_scenarios.default_base_route_id` 必须指向 `deleted_at IS NULL` 的 active 路由。
  - [x] 删除或归档基础路由时，必须阻断被 active 场景引用的路由，或同步迁移场景绑定。
  - [x] 需要增加启动/装载校验：active 场景不能绑定已删除路由、空模型池路由或全量不可用模型。

- [x] 把用量、计费、成功率建立在真实执行结果上
  - [x] `calls` 只能表示一次网关请求或一次标准化模型调用，不能由客户端任意传入或 seed 随意构造。
  - [x] chat/text_generation token 用量优先来自供应商响应 usage 字段。
  - [x] image、embedding 等用量应优先来自供应商响应 usage 字段；拿不到时图片按返回数量或请求数量回填。
  - [x] chat/text_generation 成本、销售额、平台费用按真实 usage 和命中价格策略计算。
  - [x] 成功率必须基于真实调用状态，区分 success、provider_error、gateway_error、rejected、timeout。

## P1 高优先级

- [x] 补齐权限与平台门禁
  - [x] AI 能力中心是 `PLATFORM_ONLY`，前端路由需要显式 `requiresPlatformAdmin` 或统一应用门禁。
  - [x] 后端 API 当前在登录后可访问，需要确认是否有菜单/权限二次校验，而不是只靠前端隐藏。
  - [x] `GET /api/ai-capability-center/{resource}` 当前统一使用 `/ai-capability-center` 权限，生产级应按资源或菜单拆分读权限。
  - [x] `POST /api/ai-gateway/v1/invoke` 应明确调用方身份、scope、租户上下文和调用来源，不能只信任 body 里的 `tenant_id`。

- [x] 收紧内容记录策略
  - [x] `ai.gateway.content_record_level` 的 0/1/2/3 方案可用，但必须补齐安全边界。
  - [x] 级别 2 的脱敏规则需要覆盖手机号、邮箱、身份证、银行卡、token、apikey、authorization、cookie、地址等。
  - [x] 级别 3 记录完整内容必须有强告警、权限限制、保留周期和审计。
  - [x] 调用日志抽屉需要明确“内容记录级别”和“实际记录范围”，避免误解为完整 AI 结果。

- [x] 建立路由执行计划的生产校验
  - [x] 基础路由必须至少有一个 active route model。
  - [x] route model 对应模型、供应商账号、供应商 API 必须全部可用。
  - [x] 如果健康状态为 error，执行计划应拒绝或降级到可用 fallback，并记录原因。
  - [x] 概览健康检查要展示“影响哪些场景”，而不仅是总数。

- [x] 修复 seed 与运行时数据互相污染
  - [x] `seedAIProviderCatalog` 会逻辑删除 legacy route，但场景表仍可能保留旧引用，需要幂等修复。
  - [x] seed 不应在生产环境写入演示调用记录。
  - [x] seed 不得覆盖人工维护的供应商状态、账号密钥 alias、API/模型/路由/场景状态；场景绑定仅在失效时修复。
  - [x] Manifest 重装载、导入接口不得覆盖人工维护的供应商状态、密钥 alias、健康状态和场景绑定。

- [x] 补齐测试
  - [x] 增加 `Invoke` 的真实 provider adapter 单测和 mock HTTP 集成测试。
  - [x] 覆盖供应商不可用、路由为空、场景绑定删除路由、配额拒绝、限流拒绝、供应商超时、供应商返回 4xx/5xx。
  - [x] 覆盖供应商返回 5xx 时记录 `provider_error`，且不累计成功配额。
  - [x] 覆盖供应商 5xx 可重试成功并记录 `retry_count`。
  - [x] 覆盖供应商超时时记录 `timeout` 且不产生成功计费。
  - [x] 覆盖 `content_record_level` 0/1/2/3 的日志存储行为和脱敏规则。
  - [x] 覆盖平台权限、租户伪造 `tenant_id`、非平台用户访问 AI 能力中心。

## P2 中优先级

- [x] 更新文档
  - [x] `docs/apps/ai-capability-center.md` 需要恢复“调用日志”菜单说明。
  - [x] 同步说明总览、调用日志、健康检查、内容记录级别、真实调用链路和 demo 数据策略。
  - [x] 文档中的调用示例要和当前字段一致，例如 `product-center` vs `product_center` 的命名口径。

- [x] 优化前端可信表达
  - [x] 总览页需要显示数据时间范围和数据来源，例如“今日真实调用”“近 7 天真实调用”。
  - [x] 如果没有真实数据，应展示空态或接入引导，不应该用 seed 趋势撑场面。
  - [x] 调用日志默认今日是合理的，但需要给“暂无今日调用”的解释和快捷切换近 7 天。
  - [x] 抽屉里 UUID 类字段应尽量展示中文业务名称，技术 ID 放在可复制区域。

- [x] 统一应用内路由来源
  - [x] 当前前端路由仍写在全局 `frontend/src/router/index.ts`，AI 应用目录只有 `manifest.ts`，没有独立 `routes.ts`。
  - [x] 后续应让应用中心/Manifest 成为路由、菜单和权限的准入来源，避免新增菜单后再次漏配。

- [x] 观测与运维
  - [x] 增加网关调用 trace id、provider request id、重试日志、超时分布、错误分布。
  - [x] 调用日志记录并展示 provider request id、HTTP 状态、开始/结束时间、重试次数。
  - [x] Gateway 按路由/模型节点配置执行超时和 5xx/429/网络错误重试，并记录重试次数。
  - [x] 健康检查后台任务现在每 10 分钟跑一次，需要明确是否会对外部供应商产生真实调用成本。
  - [x] 连通性检查应支持 dry-run / HEAD / lightweight model list 等低成本探测方式。

## 建议实施顺序

- [x] 先冻结 demo 数据进入生产指标，给总览和调用日志加数据来源边界。
- [x] 修复 active 场景绑定已删除路由的问题，保证目录配置自洽。
- [x] 建立 provider adapter 接口和 OpenAI-compatible 第一个真实调用实现。
- [x] 把调用结果、usage、latency、error、response hash 写入 `ai_usage_records`。
- [x] 重做健康检查和执行计划校验，让“可用”只代表真实可调用。
- [x] 补齐平台权限、租户上下文和内容记录安全策略的第一阶段门禁。
- [x] 最后再继续打磨前端布局、趋势切换和调用日志抽屉。

## 验收标准

- [x] 任意 active AI 场景都能通过校验：场景 -> 基础路由 -> route model -> 模型 -> 供应商账号 -> API 全链路可用。
- [x] 未配置密钥时，网关调用明确失败并记录安全错误，不请求供应商且不产生成功计费。
- [x] 健康检查失败时，网关调用明确失败，不请求失败 API 且不产生成功用量；存在健康 fallback 时自动降级。
- [x] 一次真实调用能返回模型结果，写入真实延迟、真实状态、真实 usage、真实成本和真实日志。
- [x] 总览、趋势、排行榜、调用日志均默认只统计真实调用；demo 数据不会影响生产指标。
- [x] 非平台管理员不能进入 AI 能力中心管理页；业务调用方不能伪造其他租户的 `tenant_id`。
- [x] `content_record_level` 0/1/2/3 均有测试证明存储行为符合预期。
- [x] Manifest、文档、前端路由、后端 API 权限声明一致。

## 复查命令

```bash
docker exec saas-go-postgres psql -U saas -d saas_baseon -P pager=off -c "select count(*) total, count(*) filter (where request_id like 'seed%') seed_records, count(*) filter (where request_params::text like '%demo%') demo_records from ai_usage_records;"
docker exec saas-go-postgres psql -U saas -d saas_baseon -P pager=off -c "select health_status, count(*) from ai_provider_apis where deleted_at is null group by health_status;"
docker exec saas-go-postgres psql -U saas -d saas_baseon -P pager=off -c "select s.app_code, s.ai_scenario_code, r.route_code, r.deleted_at from ai_scenarios s left join ai_base_routes r on r.id=s.default_base_route_id where s.deleted_at is null order by s.app_code, s.ai_scenario_code;"
```
