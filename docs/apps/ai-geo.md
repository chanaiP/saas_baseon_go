# AI GEO

## 应用定位

AI GEO 是合并部署的租户应用，`app_code=ai-geo`。它面向多租户、多品牌的 GEO 内容增长场景，负责沉淀品牌、商品、SKU、竞品等资料，并串起工作台生成母稿、渠道内容适配、发布审核与发布计划执行闭环。

当前阶段已按资料原型完成应用中心装载与前端原型实现：

- Manifest：`internal/apps/ai_geo/app.manifest.yaml`
- 后端目录：`internal/apps/ai_geo`
- 前端目录：`frontend/src/apps/ai-geo`
- 前端路由：`/ai-geo`、`/ai-geo/:section` 与 `/ai-geo/:section/:subsection`

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

当前 Manifest 声明以下 API 前缀，具体 handler、service、repository 和数据库表将在业务阶段补齐：

- `/api/ai-geo/overview`
- `/api/ai-geo/materials/brands`
- `/api/ai-geo/materials/products`
- `/api/ai-geo/materials/imports`
- `/api/ai-geo/workbench/drafts/generate`
- `/api/ai-geo/drafts`
- `/api/ai-geo/drafts/:id/channel-contents`
- `/api/ai-geo/publish-plans`
- `/api/ai-geo/channels`
- `/api/ai-geo/channel-accounts`

## 后续业务准入

新增业务对象前，先补齐租户归属、逻辑删除、审计、权限、套餐校验、配额消费和单测。业务表默认包含 `tenant_id`，涉及组织维度时补 `company_id`、`department_id` 或在设计文档中说明例外。
