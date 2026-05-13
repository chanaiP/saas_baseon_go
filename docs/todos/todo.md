# AI 能力中心开发 TODO

> 分支：`codex/ai-capability-center`
> 原型：`/Users/chen.ai/Project/ai-model-center-demo-v10-tenant-strategy-dashboard.zip`
> 需求：`/Users/chen.ai/Project/ai_model_center_requirement_design.md`
> 交付口径：AI 能力中心属于业务中台的平台能力，仅平台可见和管理；非售卖、非订阅、非套餐能力。独立“用量统计”页面已按产品决策删除；总览页承接用量趋势、成本结构、租户排行，用量明细保留为后端数据源。

## 0. 方案与范围

- [x] 阅读需求设计文档和前端原型，确认 AI 能力中心合并部署形态。
- [x] 创建并切换分支 `codex/ai-capability-center`。
- [x] 确认菜单收敛为 7 个：总览、供应商、模型目录、AI 场景、基础路由、策略中心、系统设置。
- [x] 删除独立用量统计菜单、路由、套餐功能点和前端入口。
- [x] 保留 `usage-records` 后端资源，作为总览和 Gateway 调用审计的数据源。
- [x] 明确应用定位为业务中台平台能力：`PLATFORM_ONLY + NON_SELLABLE + billing_mode=NONE + package_policy=NON_SELLABLE`。

## 1. 应用装载与权限

- [x] 新增 `internal/apps/ai_capability_center/app.manifest.yaml`。
- [x] 声明 7 个菜单和 AI Gateway 调用权限。
- [x] 声明配置管理权限 `ai_capability_center:manage`。
- [x] 声明 API 权限矩阵，不声明套餐功能点和配额。
- [x] 验证 Manifest 可被应用中心解析。
- [x] 复核 manifest 菜单、权限、非售卖、非套餐与删除用量统计页后的最终口径一致。
- [x] 在本地应用中心数据库执行 manifest 装载闭环，确认后台应用记录、菜单、权限、API 已写入，套餐功能点和配额不再纳入。
- [x] 修复 Manifest 装载中 `false` 布尔值被数据库默认值覆盖的问题，确保平台菜单、权限和套餐开关按声明落库。

## 2. 基础后端与数据库

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

## 3. 基础前端信息架构

- [x] 新增前端应用 manifest、API 封装和类型定义。
- [x] 挂载 7 个前端路由。
- [x] 页面接入真实 API 并处理 loading、empty、error。
- [x] 总览页展示指标、7 天趋势、成本结构、租户排行、健康检查、核心路由。
- [x] 新增、编辑、删除按钮按 `ai_capability_center:manage` 做权限门禁。

## 4. 菜单一：总览

- [x] 后端：总览接口聚合今日调用、今日成本、成功率、P95 延迟。
- [x] 后端：总览接口聚合 7 天趋势、模型类型成本结构、租户排行榜。
- [x] 前端：展示指标卡、7 天趋势、成本结构、租户排行榜、健康检查、核心基础路由。
- [x] 文档：补充总览统计口径和用量明细数据源说明。
- [x] 测试：补充总览聚合 service 单测。
- [x] Review：检查总览 SQL 聚合、空数据兜底、数值格式和性能风险。

## 5. 菜单二：供应商

- [x] 后端：供应商整体导入接口支持 providers/accounts/apis 一次性导入。
- [x] 后端：供应商整体导入做 provider/account/api 引用校验和事务回滚。
- [x] 后端：删除供应商时处理账号/API 级联逻辑，并在引用模型/用量时阻断。
- [x] 后端：删除账号时逻辑删除账号下 API。
- [x] 供应商页改为三级管理：供应商列表、接入账号、API 配置。
- [x] 供应商页提供供应商/账号/API 三类表单抽屉，不依赖通用 JSON 编辑。
- [x] 供应商页提供整体导入入口和导入结果反馈。
- [x] 文档：补充供应商整体导入 payload 示例。
- [x] 测试：补充供应商导入、级联删除和引用阻断 service 单测。
- [x] Review：检查供应商账号密钥字段不回显、事务边界、软删除一致性。

## 6. 菜单三：模型目录

- [x] 后端：模型导入接口支持模型、价格策略、分档价格一次性导入。
- [x] 后端：模型导入做 provider/model/policy/tier 引用校验和事务回滚。
- [x] 后端：删除模型时阻断已被路由模型池或用量明细引用的模型。
- [x] 模型目录页改为供应商侧栏 + 模型列表。
- [x] 模型目录页提供模型详情抽屉。
- [x] 模型目录页提供价格策略抽屉和分档价格编辑。
- [x] 模型目录页提供模型/价格整体导入入口。
- [x] 文档：补充模型、价格策略、分档价格导入 payload 示例。
- [x] 测试：补充模型导入和价格策略分档单测。
- [x] Review：检查价格字段精度、分档匹配字段、供应商筛选和空状态。

## 7. 菜单四：AI 场景

- [x] 后端：AI 场景导入接口支持批量 upsert。
- [x] 后端：AI 场景导入校验 `app_code + ai_scenario_code` 唯一、能力字典存在、默认基础路由存在。
- [x] 后端：删除场景时阻断被租户策略或用量明细引用。
- [x] AI 场景页按应用分组展示。
- [x] AI 场景页能力编码从能力字典选择。
- [x] AI 场景页默认基础路由从路由列表选择。
- [x] AI 场景页展示租户覆盖策略数。
- [x] 文档：补充 AI 场景注册和导入 payload 示例。
- [x] 测试：补充 AI 场景批量 upsert 和引用阻断单测。
- [x] Review：检查场景注册强约束、能力字典选择、默认路由选择一致性。

## 8. 菜单五：基础路由

- [x] 后端：基础路由导入接口支持 base-routes 和 route-models 批量 upsert。
- [x] 后端：基础路由导入校验能力字典、模型存在、模型池权重/优先级合法。
- [x] 后端：删除基础路由时阻断被 AI 场景或租户策略引用，并处理模型池子资源。
- [ ] 基础路由页提供模型池编辑器，支持 role、priority、weight、retry、timeout。
- [ ] 基础路由页策略枚举覆盖 fixed/fallback/priority/load_balance/cost_first/quality_first/latency_first/quota_aware/tenant_custom/capability_match。
- [x] 文档：补充基础路由和模型池导入 payload 示例。
- [x] 测试：补充基础路由批量 upsert、模型池校验、引用阻断单测。
- [ ] Review：检查模型池编辑交互、策略枚举、删除边界。

## 9. 菜单六：策略中心

- [ ] 后端：租户策略导入接口支持 policies/quota-rules/rate-limit-rules 批量 upsert。
- [ ] 后端：租户策略导入校验 AI 场景、默认路由、覆盖路由和规则维度。
- [ ] 后端：删除策略时逻辑删除配额规则和限流规则。
- [ ] 策略中心页提供策略抽屉。
- [ ] 策略中心页支持多条配额规则编辑。
- [ ] 策略中心页支持多条限流规则编辑。
- [ ] 策略中心页展示策略优先级说明。
- [ ] 文档：补充租户策略、配额规则、限流规则导入 payload 示例。
- [ ] 测试：补充租户策略批量 upsert、规则编辑和删除级联单测。
- [ ] Review：检查策略优先级、规则维度、超限动作和多规则表单。

## 10. 菜单七：系统设置

- [ ] 后端：能力字典维护接口复核 capability_code/name/type/unit/tier_pricing/status 校验。
- [ ] 后端：网关参数设置校验默认超时、重试、告警通道、异步用量日志。
- [ ] 系统设置页拆分 AI 能力字典、网关参数、安全归属三个区域。
- [ ] 系统设置页能力字典可维护 capability_code/name/type/unit/tier_pricing/status。
- [ ] 系统设置页网关参数可维护默认超时、重试、告警通道、异步用量日志。
- [ ] 系统设置页安全归属展示 API Key 加密、Prompt 明文存储关闭、操作日志归属。
- [ ] 文档：补充系统设置字段说明和安全归属边界。
- [ ] 测试：补充能力字典和网关设置校验单测。
- [ ] Review：检查能力字典引用保护、JSON 设置格式、安全信息展示。

## 11. 横向能力：AI Gateway 调用链路

- [ ] 后端：Gateway invoke 支持租户策略覆盖默认基础路由。
- [ ] 后端：Gateway invoke 支持多条配额规则判定，输出判定结果。
- [ ] 后端：Gateway invoke 支持多条限流规则判定，输出判定结果。
- [ ] 后端：Gateway invoke 支持按 route strategy 选择模型池模型。
- [ ] 后端：Gateway invoke 支持价格策略和分档价格匹配。
- [ ] 后端：Gateway invoke 写入 cost_amount、billing_amount、platform_unit、platform_amount、price_policy_id、price_tier_id、tenant_strategy_id。
- [ ] 文档：补充 Gateway invoke 策略覆盖、配额限流、价格匹配说明。
- [ ] 测试：补充 Gateway invoke 策略覆盖、价格匹配、用量写入单测。
- [ ] Review：检查策略优先级、价格匹配、用量写入是否与需求一致。

## 12. 横向能力：审计、权限、装载

- [ ] 后端：写操作审计补 before/after 差异，不只记录 patch。
- [x] 应用中心：每次菜单/权限/API/套餐变更后执行 manifest scan/load 或等价装载验证。
- [ ] 权限：复核页面按钮、API、manifest 三方一致。
- [ ] 文档：补充部署与应用装载验证步骤。
- [ ] 测试：补充 manifest 菜单数量和删除用量统计页断言。
- [ ] Review：检查应用中心执行闭环、权限资源归属、套餐功能点口径。

## 13. 横向能力：页面验收

- [x] 运行 `gofmt`。
- [x] 运行相关 Go 测试或编译验证。
- [x] 运行前端生产构建。
- [x] 运行 `git diff --check`。
- [ ] 复核移动端和桌面端布局不溢出、不重叠。
- [ ] Review 前端交互：确认没有用通用 JSON 编辑替代生产表单。

## 14. 最终测试与代码 Review

- [ ] 执行目标 Go 测试集。
- [ ] 执行前端 `npm run build`。
- [ ] 执行 `git diff --check`。
- [ ] Review 后端资源边界：是否存在未校验引用、错误软删除、事务缺失。
- [ ] Review 文档：菜单、接口、部署说明是否与代码一致。
- [ ] 修复 review 发现的问题并回归测试。

## 15. 提交

- [ ] 每完成一组可验证功能后更新本 TODO 勾选。
- [ ] 每完成一组可验证功能后提交独立 commit。
- [ ] 最终汇总完成项、测试结果、review 结论和残余风险。
