# 经营数据决策中心生产级开发 TODO

创建时间：2026-05-18
应用定位：合并部署内置业务中台应用
应用编码：`data-center`
后端目录：`internal/apps/data_center`
前端目录：`frontend/src/apps/data-center`
API 前缀：`/api/data-center`
验收口径：本文件所有任务全部打勾后，才算“经营数据决策中心”真正完成。禁止用 mock 数据冒充生产链路，禁止只完成前端静态页面。

## 0. 执行纪律

- [x] 每开始一个开发阶段前，先确认本 TODO 中对应任务未完成项。
- [x] 每完成一个可验证任务后，只在验证通过后打勾。
- [x] 任一任务如果降级实现，必须在本文件追加“降级原因、影响范围、后续补齐方案”，不得直接打勾。
- [x] 所有新增业务数据访问必须带 `tenant_id` 约束，不能信任前端传入的租户参数。
- [x] 所有列表接口必须返回统一分页结构：`items`、`total`、`skip`、`limit`。
- [x] 所有接口必须返回统一响应结构：`code`、`message`、`data`。
- [x] 删除、关闭、忽略、归档类操作必须保留历史，不做业务数据物理删除。

## 1. 应用接入与 Manifest

- [x] 创建后端应用目录 `internal/apps/data_center`，包含 `app.go`、`app.manifest.yaml`、`handlers`、`services`、`repositories`、`dto`、`domain`、`tests`。
- [x] 创建前端应用目录 `frontend/src/apps/data-center`，包含 `manifest.ts`、`routes.ts`、`api.ts`、`views`、`components`、`composables`、`types`。
- [x] 在 `app.manifest.yaml` 声明 `app_code=data-center`、`deployment_mode=MERGED`、`source=BUILTIN`、`app_type=BUSINESS_MIDDLE_PLATFORM`。
- [x] 在 Manifest 中声明 9 个菜单：经营看板、数据总览、原始数据、标准数据、指标中心、异常分析、异常规则、整改任务、整改复盘。
- [x] 在 Manifest 中声明所有操作权限：指标维护、规则维护、异常扫描、AI 分析、生成任务、任务流转、复盘确认、批次重试、数据导出。
- [x] 在 Manifest 中声明 `/api/data-center/**` API 权限矩阵，GET 对应菜单权限，写操作对应操作权限。
- [x] 在 Manifest 中声明套餐功能点 `package_features`，来源必须是应用能力而非前端临时列表。
- [x] 在 Manifest 中声明配额 `quotas`，至少包含数据批次保留量、每日异常扫描次数、每日 AI 分析次数、活跃规则数量。
- [x] 确保应用装载重复执行后不会重复生成菜单、权限、套餐功能点和配额。
- [x] 更新或补充应用说明文档，记录菜单、权限、套餐、配额、API 和初始化方式。

## 2. 数据库结构与迁移

- [x] 新增版本化迁移 SQL，创建数据中心业务表，不依赖 GORM AutoMigrate 作为生产建库方式。
- [x] 同步更新 `internal/infrastructure/persistence/postgres/schema/current_schema.sql`。
- [x] 所有租户业务表使用 `tenant_id bigint not null`，必要时包含 `company_id`、`department_id`、`business_unit_id`。
- [x] 创建原始数据批次表 `data_center_raw_data_batches`，支持批次、来源、状态、错误摘要、逻辑删除和审计字段。
- [x] 创建原始错误明细表 `data_center_raw_data_errors`，用于批次错误追溯。
- [x] 创建标准销售订单表 `data_center_std_sales_orders`。
- [x] 创建标准投流日表 `data_center_std_ad_daily`。
- [x] 创建标准库存日表 `data_center_std_inventory_daily`。
- [x] 创建标准退款、商品、门店销售相关表，满足第一版页面与指标计算需要。
- [x] 创建指标定义表 `data_center_metric_definitions`，支持启停、异常判断标记、逻辑删除。
- [x] 创建指标结果表 `data_center_metric_results`，按租户、指标、对象、周期、日期建立唯一约束或去重索引。
- [x] 创建异常规则表 `data_center_anomaly_rules`，保存条件 JSON、等级 JSON、置信度 JSON、AI 设置、任务设置、复盘规则。
- [x] 创建异常记录表 `data_center_anomaly_records`，保存规则、对象、证据 JSON、置信度、影响金额、AI/任务/复盘状态。
- [x] 创建 AI 分析记录表 `data_center_ai_diagnosis_records`，结构化保存问题摘要、原因、证据引用、建议、任务建议。
- [x] 创建整改任务表 `data_center_rectification_tasks`，关联异常、责任人、协同人、目标、进度、状态和反馈。
- [x] 创建任务过程记录表 `data_center_task_logs`，记录状态流转、反馈、操作人、操作时间。
- [x] 创建整改复盘表 `data_center_rectification_reviews`，保存整改前后指标、改善幅度、AI 总结、人工结论和经验沉淀。
- [x] 创建关键唯一约束：同一租户下指标 code、规则 code、批次 code、异常 code、任务 code、复盘 code 不重复。
- [x] 创建异常去重约束：`tenant_id + rule_code + object_type + object_code + stat_date` 同周期不重复生成。
- [x] 为列表查询建立必要索引：租户、状态、时间、业务域、等级、来源批次、责任人。
- [x] 迁移脚本提供 down 回滚，且不清空已有业务表。

## 3. 后端分层与领域模型

- [x] 定义 `internal/apps/data_center/domain` 下的领域类型和状态枚举，避免魔法字符串散落。
- [x] 定义 DTO：筛选请求、分页响应、创建/更新指标、规则、任务、复盘、批次重试等请求响应结构。
- [x] 实现 repository 层，所有 list/detail/update/delete 均强制接收租户上下文。
- [x] repository 不读取 HTTP、Header、Authorization，不做权限判断。
- [x] 实现 service 层，负责业务规则、事务、权限/套餐/配额校验、异常转换、审计点。
- [x] 实现 handler 层，只做参数绑定、当前用户读取、调用 service、统一响应。
- [x] 在 `internal/bootstrap/router.go` 初始化 data-center service、repository、handler。
- [x] 在 `internal/bootstrap/api_routes.go` 注册 `/api/data-center` 路由。
- [x] 更新 `internal/interfaces/http/handlers/auth_policy.go`，补齐所有 data-center 路由权限映射。
- [x] 更新 router policy 测试，确保所有新增 API 都有权限分类或明确 optional。
- [x] 更新 OpenAPI 标签和路径，包含 data-center API。

## 4. 真实业务 API

- [x] 实现经营看板 API：summary、trends、rankings、anomalies、tasks，全部从数据库聚合。
- [x] 实现数据总览 API：pipeline、jobs、errors，展示真实批次和处理状态。
- [x] 实现原始数据 API：批次列表、详情、错误明细、重新清洗、重新同步占位门禁。
- [x] 实现标准数据 API：sales、ad、inventory、refund、store-sales、详情。
- [x] 实现指标中心 API：列表、新增、编辑、启用、禁用、指标结果查询。
- [x] 实现异常规则 API：列表、详情、新增、编辑、启用、禁用、规则测试。
- [x] 实现异常分析 API：列表、详情、触发 AI 分析、重新分析、生成任务、确认、忽略、关闭。
- [x] 实现整改任务 API：列表、详情、新建、编辑、开始处理、反馈、更新进度、完成、关闭。
- [x] 实现整改复盘 API：列表、详情、生成复盘、确认复盘、编辑复盘。
- [x] 所有列表支持 `skip`、`limit`、时间、品牌、渠道、平台、门店、商品、状态等筛选。
- [x] 所有写操作记录审计信息，至少包含操作人、租户、操作类型、对象 code、IP、User-Agent。

## 5. 数据导入、标准化与种子数据

- [x] 提供生产可用的第一版数据录入或导入入口，不能只依赖前端 mock。
- [x] 提供原始批次写入服务，支持订单、退款、广告、商品、库存、门店销售。
- [x] 提供标准化服务，将原始批次转入标准表并记录成功/失败数量。
- [x] 失败数据必须写入错误明细，页面可追溯。
- [x] 初始 seed 只允许写入系统内置指标定义和内置异常规则，不写伪装成真实经营的业务流水。
- [x] 内置指标 seed 幂等，不覆盖用户修改的启停、名称、公式说明和异常判断开关。
- [x] 内置异常规则 seed 幂等，不覆盖用户修改的阈值、范围、AI 设置、任务生成和复盘规则。
- [x] 如需要演示样例，必须标记为 demo tenant 或 demo 数据源，默认生产 API 不展示。

## 6. 指标计算与规则引擎

- [x] 实现 GMV、净销售额、订单数、客单价、退款率指标计算。
- [x] 实现广告消耗、广告 GMV、ROI、点击率、转化率、成交成本指标计算。
- [x] 实现库存、销量、动销率、可售天数相关指标计算。
- [x] 指标计算结果写入 `data_center_metric_results`，并保留 compare value、compare rate、target value。
- [x] 实现规则条件解析，支持固定阈值、环比、目标偏差、多条件 AND/OR。
- [x] 实现异常等级计算，支持低、中、高、严重。
- [x] 实现系统置信度计算，AI 不参与异常是否命中判断。
- [x] 异常命中必须生成证据 JSON，包含 metric_code、label、value、desc。
- [x] 同一租户、规则、对象、周期内不得重复生成异常。
- [x] 提供异常扫描 service，可由 API 手动触发，后续可接定时任务。
- [x] 为 MVP 完成 3 条闭环：GMV 下滑、投流增加但 ROI 下降、库存销售异常。

## 7. AI 分析接入

- [x] AI 分析必须读取异常记录、触发规则和 evidence_json 作为输入。
- [x] AI 不得决定异常是否成立，只能输出原因、影响、建议、任务文案、复盘总结。
- [x] 接入现有 AI 能力中心或网关，不在 data-center 内硬编码供应商密钥。
- [x] AI 输出必须解析为结构化 JSON，并保存到 `data_center_ai_diagnosis_records`。
- [x] AI 调用失败时更新 `ai_status=failed`，保留安全错误信息，不泄露密钥、SQL、内部堆栈。
- [x] 重新分析必须保留历史记录或可追溯版本，不直接覆盖无痕结果。
- [x] 达到规则任务生成阈值时，允许从 AI 建议生成任务，但必须防重复。

## 8. 整改任务与复盘闭环

- [x] 异常生成任务时必须关联 anomaly_code、rule_code、evidence、AI 建议。
- [x] 已生成任务的异常不能重复生成相同任务。
- [x] 任务状态流转必须受控：待处理、处理中、已完成、已逾期、已关闭。
- [x] 任务开始、反馈、进度更新、完成、关闭均写入过程记录。
- [x] 任务完成后进入待复盘状态。
- [x] 复盘生成必须读取整改前后指标并计算改善幅度。
- [x] 复盘结论支持：整改有效、效果不明显、整改无效、需继续跟进。
- [x] 人工确认或修改复盘结论时必须保存操作人和时间。
- [x] 复盘完成后回写异常和任务复盘状态。

## 9. 前端生产化实现

- [x] 从原型迁移 UI 体验，但不得在组件中写死业务数据。
- [x] 前端所有请求集中在 `frontend/src/apps/data-center/api.ts`。
- [x] 前端路由集中在 `frontend/src/apps/data-center/routes.ts`，再由全局 router 引入。
- [x] 前端类型集中在 `frontend/src/apps/data-center/types.ts`。
- [x] 实现全局筛选 composable，时间、品牌、渠道、平台、门店、商品筛选联动所有页面 API。
- [x] 经营看板从真实 API 获取指标、趋势、排行、重点异常、待处理任务。
- [x] 数据总览从真实 API 获取链路状态和批次列表。
- [x] 原始数据页支持批次筛选、详情抽屉、错误明细、重新清洗操作。
- [x] 标准数据页支持销售、投流、商品、库存、门店、退款数据域切换。
- [x] 指标中心支持指标列表、新增、编辑、启用、禁用。
- [x] 异常规则页支持规则列表、详情抽屉、新增/编辑表单、JSON 条件预览和规则测试。
- [x] 异常分析页支持统计、筛选、详情抽屉、触发/重新 AI 分析、生成任务、确认、忽略、关闭。
- [x] 整改任务页支持列表/看板、详情抽屉、开始处理、反馈、进度、完成、关闭。
- [x] 整改复盘页支持列表、详情、前后指标对比、确认结论、经验沉淀。
- [x] 所有页面处理 loading、empty、error 三态。
- [x] 所有危险操作有二次确认和成功/失败反馈。
- [x] 按后端权限和套餐结果显示菜单与按钮，不硬编码角色判断。
- [x] 页面不得出现“演示数据”“mockData”作为生产数据来源。

## 10. 权限、套餐、配额与租户隔离

- [x] 后端所有 data-center API 通过登录态获取当前用户和租户。
- [x] 普通租户不能通过 query/body/path 伪造 `tenant_id` 访问其他租户数据。
- [x] 平台管理员跨租户查询必须显式表达，并有权限保护。
- [x] 菜单可见性来源于 Manifest 装载后的权限数据。
- [x] 写操作必须校验操作权限。
- [x] 进入可售套餐的能力来自 Manifest `package_features`。
- [x] AI 分析、异常扫描、活跃规则数量等能力必须校验套餐和配额。
- [x] 配额消耗失败不得产生业务成功状态。
- [x] 租户菜单运行时应受平台专属、套餐、角色权限、租户覆盖共同约束。

## 11. 测试

- [x] 后端 repository/service 单测覆盖 tenant 隔离。
- [x] 后端测试覆盖普通租户伪造 `tenant_id` 被拒绝。
- [x] 后端测试覆盖指标计算核心口径。
- [x] 后端测试覆盖规则引擎命中、未命中、AND/OR、置信度、去重。
- [x] 后端测试覆盖异常生成任务防重复。
- [x] 后端测试覆盖任务状态流转和过程记录。
- [x] 后端测试覆盖复盘前后指标计算。
- [x] 后端测试覆盖 AI 分析成功、失败、结构化解析错误。
- [x] 后端测试覆盖权限映射和 unclassified API 失败关闭策略。
- [x] 前端构建通过。
- [x] 关键前端交互至少通过浏览器验证：看板、异常详情、规则编辑、任务流转、复盘详情。

## 12. 验证与交付

- [x] `go test ./...` 通过。
- [x] `cd frontend && npm run build` 通过。
- [x] `git diff --check` 通过。
- [x] 本地数据库执行迁移成功。
- [x] Manifest 扫描/装载成功，重复装载资源数量稳定。
- [x] 使用真实数据库数据打开 9 个页面，无控制台错误。
- [x] 从原始批次到标准数据、指标、异常、AI 分析、任务、复盘至少跑通 3 条 MVP 闭环。
- [x] 更新 README 或应用文档，说明启动、迁移、初始化、验收和故障排查。
- [x] 提交前执行 `git status`，只暂存本次 data-center 相关文件。
- [x] 完成一次 Git 提交，提交信息说明做了什么以及为什么。

## 13. 最终验收定义

- [x] 原始数据可以入库、追溯错误并重新清洗。
- [x] 标准数据来自数据库，可被指标计算消费。
- [x] 指标定义和指标结果来自数据库，可维护、可启停。
- [x] 异常由规则引擎基于指标生成，且每条异常都有证据。
- [x] AI 分析只基于异常证据输出结构化原因和建议。
- [x] 整改任务必须关联异常、规则、证据和 AI 建议。
- [x] 整改复盘必须展示整改前后指标和改善幅度。
- [x] 前端 9 个页面全部打通真实 API 和数据库。
- [x] Manifest、菜单、权限、套餐、配额、前端路由、后端 API 保持一致。
- [x] 所有任务打勾，本 TODO 才允许标记为完成。

## 14. 现场问题修复

- [x] 修复框架菜单“总览”点击进入 `/data-center` 后被重定向导致页签与页面不同步的问题。
- [x] 浏览器复测“总览”可进入真实页面，且无控制台错误。
- [x] 修复“总览”和“经营看板”都显示经营看板的问题，`/data-center` 显示总览链路，`/data-center/dashboard` 显示经营看板。
- [x] 修复总览接口启用指标统计误用 `status` 字段导致的请求失败。
- [x] 向本地数据库写入生产级服装行业经营数据，覆盖原始批次、标准数据、指标、异常、AI 诊断、整改任务和复盘。
- [x] 修复 data-center 前端 API 对后端 PascalCase 模型响应的字段兼容，确保真实数据库记录在列表和详情中完整显示。
- [x] 修复 data-center 页面浅色/深色主题适配，卡片、输入框、表格和空状态跟随框架主题。
- [x] 修复经营趋势柱状图按绝对 GMV 放大导致穿透页面的问题，改为按当前数据比例缩放。
- [x] 按原型页面结构恢复经营看板、数据总览、原始数据、标准数据、指标、规则、异常、任务、复盘布局，仅保留框架菜单。
- [x] 将 `DataCenterView` 登记到后台框架页面缓存映射，保证菜单进入和直达刷新都能正常渲染。
- [x] 浏览器复测 9 个子菜单、浅色/深色模式和旧布局节点清理结果。
- [x] 修复异常分析页被后台框架 `.content-main .page` 样式覆盖导致右侧详情掉到底部的问题。
- [x] 浏览器复测异常分析页详情栏固定在右侧且与列表同一行。
- [x] 基于服装行业重建本地经营数据，覆盖 Lee、Mardi Mercredi、安德玛、Happy Socks、Carhartt WIP、New Balance Apparel 等品牌。
- [x] 执行服装行业数据脚本并核对订单、投放、库存、门店、退款、指标、异常、任务和复盘数量。
- [x] 移除经营看板 KPI 卡片中的工程说明文案，改为业务口径展示。
- [x] 浏览器复测看板和异常分析页展示服装行业品牌、经营指标与异常闭环。
- [x] 将数据总览降级合并到原始数据页，保留链路状态但不再作为独立页面呈现。
- [x] `/data-center` 与 `/data-center/overview` 统一进入原始数据内容，避免总览页和原始数据页重复。
- [x] 修复链路状态卡片直接显示后端状态值 `normal` 的问题，改为业务状态文案。
- [x] 修复点击一级目录 Ai经营决策中心后，二级菜单无法按当前页面正确高亮和切换的问题。
- [x] 修复点击一级目录 Ai经营决策中心时左侧仍停留在上一个目录菜单的问题，并过滤已合并的总览入口。
- [x] 修复在第三方集成中心页面点击 Ai经营决策中心时，左侧二级菜单被当前路由强制覆盖成第三方集成中心菜单的问题。
- [x] 将经营看板 GMV 趋势和 ROI 趋势从柱状图改为折线图，并在每个节点显示对应数值。
