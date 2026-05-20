# AI GEO

## 应用定位

AI GEO 是合并部署的租户应用，`app_code=ai-geo`。它面向多租户、多品牌的 GEO 内容增长场景，负责沉淀品牌、商品、SKU、竞品等资料，并串起工作台生成母稿、渠道内容适配、发布审核与发布计划执行闭环。

当前阶段已按资料原型完成应用中心装载与前端原型实现，并开始补齐生产级后端闭环：

- Manifest：`internal/apps/ai_geo/app.manifest.yaml`
- 后端目录：`internal/apps/ai_geo`
- 前端目录：`frontend/src/apps/ai-geo`
- 前端路由：`/ai-geo`、`/ai-geo/:section` 与 `/ai-geo/:section/:subsection`
- 生产迁移：`internal/infrastructure/persistence/postgres/migrations/000086_ai_geo_core.up.sql`、`000087_ai_geo_import_errors.up.sql`、`000088_ai_geo_audit_suggestions.up.sql`
- GORM 模型：`internal/infrastructure/persistence/postgres/models/ai_geo.go`
- 后端分层：`internal/apps/ai_geo/{dto,repositories,services,handlers}`

## 菜单

- 总览
- 工作台
- 母稿
- 发布计划
  - 发布日历
  - 发布队列
- 渠道管理
  - 渠道管理
  - 渠道账号
- 资料中心
  - 品牌资料卡
  - 商品资料卡

## 权限

读权限由菜单 path 承载，写操作由稳定权限码承载：

- `ai_geo:workbench:generate`
- `ai_geo:draft:manage`
- `ai_geo:channel_content:manage`
- `ai_geo:publish_plan:manage`
- `ai_geo:channel:manage`
- `ai_geo:channel_account:manage`
- `ai_geo:data:import`

后续业务接口必须继续从登录态读取 `tenant_id`，不得信任前端传入租户参数。

## 套餐与配额

套餐功能点由 Manifest 的 `package_features` 声明，配额由 `quotas` 声明。当前配额：

- `ai_geo_brand_count`：品牌资料卡数量
- `ai_geo_product_count`：商品资料卡数量
- `ai_geo_monthly_draft_generations`：月度母稿生成次数
- `ai_geo_monthly_publish_tasks`：月度发布任务数量
- `ai_geo_channel_account_count`：渠道账号数量

## API

当前 Manifest 声明并已接入第一批生产 API：

- `/api/ai-geo/overview`
- `/api/ai-geo/materials/brands`
- `/api/ai-geo/materials/products`
- `/api/ai-geo/materials/skus`
- `/api/ai-geo/materials/competitors`
- `/api/ai-geo/materials/keywords`
- `/api/ai-geo/materials/keywords/:id`
- `/api/ai-geo/materials/assets`
- `/api/ai-geo/materials/assets/:id`
- `/api/ai-geo/materials/hotspots`
- `/api/ai-geo/materials/hotspots/:id`
- `/api/ai-geo/materials/imports`
- `/api/ai-geo/materials/imports/:id/errors`
- `/api/ai-geo/workbench/drafts/generate`
- `/api/ai-geo/drafts`
- `/api/ai-geo/drafts/:id/channel-contents`
- `/api/ai-geo/channel-contents`
- `/api/ai-geo/channel-contents/:id`
- `/api/ai-geo/channel-contents/:id/approve`
- `/api/ai-geo/channel-contents/:id/reject`
- `/api/ai-geo/publish-plans`
- `/api/ai-geo/publish-plans/:id`
- `/api/ai-geo/channels`
- `/api/ai-geo/channel-accounts`

第一批已落地能力：

- 总览统计：品牌、商品、SKU、渠道、账号、今日母稿、待审母稿、渠道内容、发布计划、资料完整度和配额用量。
- 资料中心：品牌、商品、SKU、竞品、关键词、素材和热点资料列表/详情/新增/更新/归档。
- 资料导入：导入批次记录、字段预校验、字段映射落库、品牌/商品/SKU/竞品幂等导入、部分成功状态和错误行明细查询。
- 工作台：可持久化生成母稿，生产启动时通过 AI 能力中心场景 `ai_geo_draft_generation` 调用 Gateway；测试和未注入场景时保留本地 generator 降级实现。生成上下文已包含品牌、商品、SKU、Skill、热点和用户提示。
- 母稿：列表、详情、新增、提交审核、审核通过、驳回、归档。
- 审核建议：前端已接入母稿和渠道内容 AI 审核建议入口，结果区展示风险等级、摘要和结构化建议。
- 渠道内容：从母稿生成渠道版本，支持列表、详情、编辑、人工确认、驳回和审核意见留痕。
- 渠道管理：渠道资料、渠道账号列表/新增。
- 发布计划：列表、新增、调整时间、状态更新、失败重试、发布链接回填、状态机校验，并同步渠道内容发布状态。
- 审计：写操作记录 `audit_log`，`app_code=ai-geo`，`module=ai_geo`。
- 权限：后端路由已纳入 `auth_policy.go`，读接口走菜单权限，写接口走操作权限。
- 前端权限：AI GEO 写操作按钮按 `permissionStore.canUseAction` 做显隐，后端仍保留二次校验。
- 配额：品牌、商品、渠道账号使用静态数量门禁；母稿生成和发布任务使用套餐配额扣减，月度周期按 `yyyyMM` 记录。

## AI 能力中心接入

AI GEO 通过 `internal/apps/ai_geo/services` 下的 Gateway adapter 接入 AI 能力中心，启动装配在 `internal/bootstrap/router.go`。

当前已接入场景：

- `ai_geo_draft_generation`：工作台母稿生成。输入包含用户提示、Skill、品牌资料、商品资料、SKU 明细和热点资料；输出解析为 `title`、`summary`、`body`、`keywords` 并落库为母稿。
- `ai_geo_channel_rewrite`：渠道内容改写。输入包含母稿、渠道资料和人工覆盖字段；输出解析为 `title`、`body` 并落库为渠道内容。
- `ai_geo_audit_suggestion`：母稿和渠道内容审核建议。输入包含待审内容、渠道上下文；输出解析为风险等级、通过建议、摘要和结构化建议，并写入 `ai_geo_audit_suggestions` 保留历史。

`000089_ai_geo_ai_scenarios` 会把三条场景幂等登记到 AI 能力中心，默认绑定 `chat-default` 基础路由；启动 seed 也会修复新环境中的场景缺失。上线前仍需确认 `chat-default` 已绑定可用模型、供应商账号/API、租户策略和限流/配额规则。Gateway 成功调用会写入 `ai_usage_records`。

## 数据模型

第一批生产表：

- `ai_geo_brand_cards`
- `ai_geo_product_cards`
- `ai_geo_skus`
- `ai_geo_competitors`
- `ai_geo_channel_profiles`
- `ai_geo_channel_accounts`
- `ai_geo_drafts`
- `ai_geo_channel_contents`
- `ai_geo_publish_plans`
- `ai_geo_import_batches`
- `ai_geo_import_errors`
- `ai_geo_audit_suggestions`
- `ai_geo_material_assets`
- `ai_geo_hotspots`

表设计要求：

- 租户业务表必须带 `tenant_id`。
- 涉及组织上下文的主表带 `company_id`、`department_id`。
- 主数据和流程数据默认保留 `deleted_at`，业务删除走归档/逻辑删除。
- 租户内编码字段使用唯一约束，例如品牌编码、商品编码、渠道编码、母稿编码、发布计划编码。

## 迁移验证与回滚

上线执行：

```bash
go run ./cmd/migrate
```

执行后至少核对以下 SQL：

```sql
select to_regclass('public.ai_geo_brand_cards') is not null as has_brand_cards;
select to_regclass('public.ai_geo_import_errors') is not null as has_import_errors;
select to_regclass('public.ai_geo_audit_suggestions') is not null as has_audit_suggestions;
select count(*) from app_api_permissions where path like '/api/ai-geo/%';
select count(*) from package_features where feature_code like 'ai_geo_%';
select count(*) from saas_quota where quota_code like 'ai_geo_%';
select ai_scenario_code, status from ai_scenarios where app_code = 'ai-geo' and deleted_at is null order by ai_scenario_code;
```

回滚说明：

- 结构回滚按版本倒序执行 `.down.sql`：`000088`、`000087`、`000086`。
- 回滚前必须确认没有生产租户继续写入 AI GEO 资料、母稿、渠道内容和发布计划。
- Manifest 重装载与回滚必须成对执行，避免菜单、API 权限和套餐功能点仍指向已删除业务表。

## 状态机

母稿：

- `draft`：草稿
- `pending`：待审核
- `approved`：已通过
- `rejected`：已驳回

渠道内容：

- `audit_status=pending/approved/rejected`
- `publish_status=not_planned/planned/publishing/published/failed`

发布计划：

- `scheduled`：待发布
- `publishing`：发布中
- `published`：已发布
- `failed`：发布失败
- `cancelled`：已取消

## 后续业务准入

新增业务对象前，先补齐租户归属、逻辑删除、审计、权限、套餐校验、配额消费和单测。业务表默认包含 `tenant_id`，涉及组织维度时补 `company_id`、`department_id` 或在设计文档中说明例外。

## 生产级剩余项

- 资料中心仍需补素材文件上传解析和素材权限边界。
- 资料导入需继续补文件上传解析、模板下载和大批量异步处理。
- 渠道发布需接第三方集成中心或 Agent 执行，补 OAuth、Webhook、失败重试和发布链接回填。
- 发布计划还需补日历聚合接口和真实发布执行器闭环。
- 当前前端仍有部分原型态数据用于展示，应逐步替换为真实接口，避免生产环境误判。
- 上线前必须跑通普通租户、无套餐租户、无角色权限用户、平台管理员四类访问差异测试。

## 上线检查清单

1. 执行 `go run ./cmd/migrate`，确认 AI GEO 表、Manifest、API 权限、套餐功能点、配额和 AI 场景均已落库。
2. 确认 AI 能力中心 `chat-default` 或租户策略已绑定可用模型供应商，且 `ai_geo_draft_generation`、`ai_geo_channel_rewrite`、`ai_geo_audit_suggestion` 均可调用。
3. 使用演示租户完成资料导入、工作台对话生成、母稿提交审核、AI 审核建议、渠道改写、发布计划创建和发布状态更新。
4. 核对租户套餐、租户菜单覆盖、角色权限变更后，页面入口和 API 写操作即时生效。
5. 生产环境禁止使用开发默认数据库密码、默认 `JWT_SECRET` 和 `CORS_ORIGINS=*`。
