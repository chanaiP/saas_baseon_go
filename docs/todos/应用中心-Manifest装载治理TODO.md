# 应用中心 Manifest 装载治理 TODO

本文档用于跟踪应用中心从“应用主档管理”升级为“Manifest 装载治理中心”的实现进度。后续每完成一项，必须同步勾选。

## 1. 已完成基础

- [x] 应用中心创建/编辑弹窗改为连续滚动表单，左侧锚点弱化，右侧表单强化。
- [x] 应用列表保持现有内容，不重做列表布局。
- [x] 应用详情弹窗补齐主档、收费策略、试用、客户端、资产和项目同步信息。
- [x] 收费策略改为第一层 `免费 / 收费 / 非售卖`，收费后再选订阅制、按量收费或组合收费。
- [x] 试用策略改为平铺卡片，支持自定义试用时间。
- [x] 客户端配置从 `app_client_type` 数据字典读取，企业微信、钉钉、飞书拆开。
- [x] 客户端配置新建时不默认选中。
- [x] 手工新建应用主档增加二次确认。
- [x] 新建应用支持 Manifest 文件导入并回填表单。
- [x] Manifest 模板下载改为后端输出标准 YAML。
- [x] Manifest 上传改为后端解析校验，不再只在浏览器本地解析。
- [x] 后端支持扫描运行环境内的 `app.manifest.yaml|yml|json`。
- [x] 已存在 `app_code` 的 Manifest 新建导入会被阻断，只能进入同步升级。
- [x] Docker 镜像已拷贝 `internal/apps`，容器内可扫描内置应用 Manifest。
- [x] 内置应用 `app-center` 输出标准 Manifest。
- [x] 内置应用 `system-management` 输出标准 Manifest。
- [x] 内置应用 `system-monitor` 输出标准 Manifest。
- [x] 内置应用按应用目录补基础工程骨架。
- [x] 新增 `app:load` 权限及 migration。
- [x] 业务开发标准补充 Manifest 持续声明规则。
- [x] 应用中心接入强制规则补充标准 Manifest 输出约束。
- [x] 套餐中心文档明确功能点来源改为 Manifest `package_features`，配额来源改为 Manifest `quotas`。
- [x] 菜单管理文档明确菜单自动生成功能点只保留为历史兼容。
- [x] `AGENTS.md` 与 `CLAUDE.md` 补充 AI agent 执行门禁。
- [x] 新增独立部署应用接入约束文档。

## 2. P0 Manifest 装载闭环

- [x] 设计并创建 `sys_app_manifest_load` 表，记录装载批次、来源、hash、版本、操作者、结果、错误摘要。
- [x] 设计并创建 `sys_app_manifest_file` 表，记录上传或扫描到的 Manifest 文件、片段角色、hash 和摘要。
- [x] 设计并创建应用资产表：`sys_app_entry`、`sys_app_api`、`sys_app_permission`、`sys_app_package_feature`、`sys_app_quota`。
- [x] 补 GORM model、repository 和 schema/migration 文档。
- [x] 实现 Manifest diff service，输出新增、更新、停用、冲突、阻断项。
- [ ] Manifest diff 支持多文件按 `app_code` 分组。
- [ ] Manifest diff 支持同一 `app_code` 多片段按 `fragment_role` 合并预览。
- [ ] Manifest diff 阻断同一 `app_code` 多个 `main`。
- [x] Manifest diff 阻断菜单 path、权限码、API method + path、套餐功能码、配额码冲突。
- [x] 实现 Manifest load service，确认后事务写入主档、客户端、菜单、权限、API、套餐功能和配额。
- [x] Manifest load 写入装载记录和审计日志。
- [x] Manifest load 支持失败回滚，禁止半成品数据。
- [ ] Manifest 删除资源时只生成停用、归档或移出套餐 diff，不物理删除。

## 3. P0 套餐与权限同步

- [x] `package_features` 同步到 `saas_feature`。
- [x] `quotas` 同步到 `saas_quota`。
- [x] `menus` 和 `operations` 同步到 `permission`。
- [x] `apis` 同步到 API 权限矩阵或应用 API 资产表。
- [x] `include_in_package=false` 的能力不得进入套餐中心。
- [x] 平台专属能力进入套餐中心时阻断装载。
- [x] 菜单自动生成功能点逻辑降级为历史兼容，新增能力不再依赖该逻辑。

## 4. P0 防覆盖规则

- [x] Manifest 重装载不得覆盖租户菜单覆盖。
- [x] Manifest 重装载不得覆盖套餐列开关。
- [x] Manifest 重装载不得覆盖角色授权。
- [x] Manifest 重装载不得覆盖租户订阅。
- [x] Manifest 重装载不得自动加回人工移出套餐的功能点。
- [ ] 人工覆盖、人工移出套餐、租户覆盖和角色授权必须有保护标记或可追溯来源。

## 5. P1 前端装载流程

- [x] Manifest 上传后展示 diff 预览。
- [ ] Manifest 扫描后展示候选应用列表。
- [x] 区分“新建导入”和“同步升级”。
- [ ] 多 Manifest 按 `app_code` 分组展示。
- [x] 已存在 `app_code` 只能选择同步升级，不能新建导入。
- [x] 冲突项高亮展示，并给出阻断原因。
- [x] 确认装载后调用 load 接口。
- [x] 装载成功后刷新应用列表、详情和装载记录。

## 6. P1 应用详情增强

- [x] 详情页展示 Manifest 版本、hash、来源和最近同步时间。
- [ ] 详情页增加菜单/权限/API/套餐功能/配额页签。
- [ ] 详情页增加装载记录页签。
- [ ] 详情页展示同步状态和错误摘要。
- [x] 详情页展示独立部署健康检查、API 地址、Webhook 地址和通讯模式。

## 7. P1 独立部署预留

- [x] 应用主档补健康检查地址。
- [x] 应用主档补 API 基础地址。
- [x] 应用主档补 Webhook 地址。
- [x] 应用主档补通讯模式详情。
- [ ] Manifest 校验独立部署应用必须声明通讯模式。
- [ ] Manifest 校验独立部署应用不得缺少租户上下文、scope、签名或幂等策略说明。
- [ ] 暂不发放真实凭证，只保留字段、校验口径和文档约束。

## 8. P2 租户侧应用开通

- [ ] 租户侧可见应用列表。
- [ ] 应用订阅或套餐外订阅。
- [ ] 应用安装到租户菜单。
- [ ] 应用角色授权。
- [ ] 试用转付费。
- [ ] 试用结束终止权益。

## 9. P2 运行治理

- [ ] API 调用统计。
- [ ] 应用使用记录。
- [ ] 套餐拦截记录。
- [ ] 权限拦截记录。
- [ ] 健康检查记录。
- [ ] Webhook 投递日志。
- [ ] 数据同步任务日志。

## 10. 每轮实现后必须执行

- [x] 更新本 TODO 勾选状态。
- [x] 相关 Go 测试通过。
- [x] 前端构建通过，或说明未影响前端。
- [x] `git diff --check` 通过。
- [x] 必要文档同步更新。
- [x] 提交 Git commit。
