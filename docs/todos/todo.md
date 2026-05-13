# AI 能力中心开发 TODO

> 分支：`codex/ai-capability-center`
> 原型：`/Users/chen.ai/Project/ai-model-center-demo-v10-tenant-strategy-dashboard.zip`
> 需求：`/Users/chen.ai/Project/ai_model_center_requirement_design.md`
> 交付口径：独立“用量统计”页面已按产品决策删除；总览页承接用量趋势、成本结构、租户排行，用量明细保留为后端数据源。

## 0. 方案与范围

- [x] 阅读需求设计文档和前端原型，确认 AI 能力中心合并部署形态。
- [x] 创建并切换分支 `codex/ai-capability-center`。
- [x] 确认菜单收敛为 7 个：总览、供应商、模型目录、AI 场景、基础路由、策略中心、系统设置。
- [x] 删除独立用量统计菜单、路由、套餐功能点和前端入口。
- [x] 保留 `usage-records` 后端资源，作为总览和 Gateway 调用审计的数据源。

## 1. 应用装载与权限

- [x] 新增 `internal/apps/ai_capability_center/app.manifest.yaml`。
- [x] 声明 7 个菜单和 AI Gateway 调用权限。
- [x] 声明配置管理权限 `ai_capability_center:manage`。
- [x] 声明 API 权限矩阵、套餐功能点和配额。
- [x] 验证 Manifest 可被应用中心解析。
- [x] 复核 manifest 菜单、权限、套餐功能点与删除用量统计页后的最终口径一致。
- [x] 在本地应用中心数据库执行 manifest 装载闭环，确认后台应用记录、菜单、权限、API、套餐功能点和配额已写入。

## 2. 后端与数据库

- [x] 新增 AI 能力中心后端应用包和 `AppCode` 常量。
- [x] 新增 GORM 模型定义。
- [x] 新增 PostgreSQL 迁移和回滚 SQL。
- [x] 同步 `current_schema.sql`。
- [x] 将模型加入本地 AutoMigrate 兜底。
- [x] 注册后端 HTTP 路由和权限路由分类。
- [x] 实现概览、分页列表、配置写入、逻辑删除和 Gateway invoke 骨架。
- [x] 修复用量记录按不存在的 `updated_at` 排序风险。
- [x] 补充引用校验、删除限制和策略优先级基础校验。
- [x] 总览接口返回 7 天趋势、成本结构、租户排行榜真实聚合。
- [ ] 供应商整体导入接口支持 providers/accounts/apis 一次性导入，并做引用校验。
- [ ] 模型导入接口支持模型、价格策略、分档价格一次性导入。
- [ ] 场景、基础路由、租户策略导入接口支持批量 upsert。
- [ ] Gateway invoke 支持租户策略覆盖默认基础路由。
- [ ] Gateway invoke 支持多条配额规则判定，输出判定结果。
- [ ] Gateway invoke 支持多条限流规则判定，输出判定结果。
- [ ] Gateway invoke 支持按 route strategy 选择模型池模型。
- [ ] Gateway invoke 支持价格策略和分档价格匹配。
- [ ] Gateway invoke 写入 cost_amount、billing_amount、platform_unit、platform_amount、price_policy_id、price_tier_id、tenant_strategy_id。
- [ ] 写操作审计补 before/after 差异，不只记录 patch。
- [ ] 删除供应商时处理账号/API 级联逻辑，并在引用模型/用量时阻断。
- [ ] 删除账号时逻辑删除账号下 API。
- [ ] 删除基础路由、场景、策略时补齐引用和子资源处理。

## 3. 前端信息架构

- [x] 新增前端应用 manifest、API 封装和类型定义。
- [x] 挂载 7 个前端路由。
- [x] 页面接入真实 API 并处理 loading、empty、error。
- [x] 总览页展示指标、7 天趋势、成本结构、租户排行、健康检查、核心路由。
- [x] 新增、编辑、删除按钮按 `ai_capability_center:manage` 做权限门禁。
- [ ] 供应商页改为三级管理：供应商列表、接入账号、API 配置。
- [ ] 供应商页提供供应商/账号/API 三类表单抽屉，不依赖通用 JSON 编辑。
- [ ] 供应商页提供整体导入入口和导入结果反馈。
- [ ] 模型目录页改为供应商侧栏 + 模型列表。
- [ ] 模型目录页提供模型详情抽屉。
- [ ] 模型目录页提供价格策略抽屉和分档价格编辑。
- [ ] 模型目录页提供模型/价格整体导入入口。
- [ ] AI 场景页按应用分组展示。
- [ ] AI 场景页能力编码从能力字典选择。
- [ ] AI 场景页默认基础路由从路由列表选择。
- [ ] AI 场景页展示租户覆盖策略数。
- [ ] 基础路由页提供模型池编辑器，支持 role、priority、weight、retry、timeout。
- [ ] 基础路由页策略枚举覆盖 fixed/fallback/priority/load_balance/cost_first/quality_first/latency_first/quota_aware/tenant_custom/capability_match。
- [ ] 策略中心页提供策略抽屉。
- [ ] 策略中心页支持多条配额规则编辑。
- [ ] 策略中心页支持多条限流规则编辑。
- [ ] 策略中心页展示策略优先级说明。
- [ ] 系统设置页拆分 AI 能力字典、网关参数、安全归属三个区域。
- [ ] 系统设置页能力字典可维护 capability_code/name/type/unit/tier_pricing/status。
- [ ] 系统设置页网关参数可维护默认超时、重试、告警通道、异步用量日志。
- [ ] 系统设置页安全归属展示 API Key 加密、Prompt 明文存储关闭、操作日志归属。
- [ ] 复核移动端和桌面端布局不溢出、不重叠。

## 4. 文档

- [x] 新增应用说明文档 `docs/apps/ai-capability-center.md`。
- [x] 文档同步删除独立用量统计页。
- [ ] 文档补充 7 个菜单的最终页面能力说明。
- [ ] 文档补充导入接口 payload 示例。
- [ ] 文档补充 Gateway invoke 策略覆盖、配额限流、价格匹配说明。
- [ ] 文档补充部署与应用装载验证步骤。

## 5. 测试与验证

- [x] 运行 `gofmt`。
- [x] 运行相关 Go 测试或编译验证。
- [x] 运行前端生产构建。
- [x] 运行 `git diff --check`。
- [ ] 补充供应商导入 service 单测。
- [ ] 补充模型导入 service 单测。
- [ ] 补充场景/基础路由/租户策略批量 upsert 单测。
- [ ] 补充 Gateway invoke 策略覆盖、价格匹配、用量写入单测。
- [ ] 补充 manifest 菜单数量和删除用量统计页断言。
- [ ] 执行目标 Go 测试集。
- [ ] 执行前端 `npm run build`。
- [ ] 执行 `git diff --check`。

## 6. 代码 Review

- [ ] Review 后端资源边界：是否存在未校验引用、错误软删除、事务缺失。
- [ ] Review Gateway invoke：策略优先级、价格匹配、用量写入是否与需求一致。
- [ ] Review 前端交互：是否还有通用 JSON 编辑替代生产表单。
- [ ] Review 权限：页面按钮、API、manifest 是否一致。
- [ ] Review 文档：菜单、接口、部署说明是否与代码一致。
- [ ] 修复 review 发现的问题并回归测试。

## 7. 提交

- [ ] 每完成一组可验证功能后更新本 TODO 勾选。
- [ ] 每完成一组可验证功能后提交独立 commit。
- [ ] 最终汇总完成项、测试结果、review 结论和残余风险。
