# AI GEO

## 应用定位

AI GEO 是合并部署的租户应用，`app_code=ai-geo`。它用于管理品牌在生成式搜索、AI 问答和模型引用场景中的可见度监测、引用分析、竞品洞察和内容优化。

当前阶段先完成应用中心装载骨架：

- Manifest：`internal/apps/ai_geo/app.manifest.yaml`
- 后端目录：`internal/apps/ai_geo`
- 前端目录：`frontend/src/apps/ai-geo`
- 前端路由：`/ai-geo` 与 `/ai-geo/:section`

## 菜单

- GEO 总览
- 项目管理
- 提示词库
- 可见度监测
- 引用分析
- 竞品洞察
- 内容优化
- 报告中心

## 权限

读权限由菜单 path 承载，写操作由稳定权限码承载：

- `ai_geo:project:manage`
- `ai_geo:query:manage`
- `ai_geo:monitoring:run`
- `ai_geo:content:optimize`
- `ai_geo:report:export`

后续业务接口必须继续从登录态读取 `tenant_id`，不得信任前端传入租户参数。

## 套餐与配额

套餐功能点由 Manifest 的 `package_features` 声明，配额由 `quotas` 声明。当前配额：

- `ai_geo_project_count`：GEO 项目数量
- `ai_geo_monthly_monitor_runs`：月度监测次数
- `ai_geo_monthly_content_optimizations`：月度内容优化次数

## API

第一版 Manifest 声明以下 API 前缀，具体 handler、service、repository 和数据库表将在业务阶段补齐：

- `/api/ai-geo/overview`
- `/api/ai-geo/projects`
- `/api/ai-geo/queries`
- `/api/ai-geo/monitoring/runs`
- `/api/ai-geo/citations`
- `/api/ai-geo/competitors`
- `/api/ai-geo/content/optimize`
- `/api/ai-geo/reports`
- `/api/ai-geo/reports/export`

## 后续业务准入

新增业务对象前，先补齐租户归属、逻辑删除、审计、权限、套餐校验、配额消费和单测。业务表默认包含 `tenant_id`，涉及组织维度时补 `company_id`、`department_id` 或在设计文档中说明例外。
