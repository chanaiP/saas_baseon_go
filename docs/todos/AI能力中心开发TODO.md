# AI 能力中心开发 TODO

> 分支：`codex/ai-capability-center`
> 原型：`/Users/chen.ai/Project/ai-model-center-demo-v10-tenant-strategy-dashboard.zip`
> 需求：`/Users/chen.ai/Project/ai_model_center_requirement_design.md`

## 方案与接入

- [x] 阅读 AI 能力中心需求设计文档，确认总览、供应商、模型目录、AI 场景、基础路由、用量统计、策略中心、系统设置边界。
- [x] 解包并审阅前端原型，确认 React/Vite 原型需转译为当前 Vue 3 + Go 合并部署应用。
- [x] 在项目 git 根目录 `saas_baseon_go` 切换到开发分支 `codex/ai-capability-center`。
- [x] 确认应用形态为 `deployment_mode=MERGED`，不是独立部署应用。
- [x] 确认应用编码为 `ai-capability-center`，后端目录为 `internal/apps/ai_capability_center`，前端目录为 `frontend/src/apps/ai-capability-center`。

## 后端与数据库

- [x] 新增 AI 能力中心后端应用包和 `AppCode` 常量。
- [x] 新增 AI 能力中心 GORM 模型定义。
- [x] 新增生产迁移 SQL 和回滚 SQL。
- [x] 同步更新 `current_schema.sql`。
- [x] 新增 AI 能力中心 service，提供概览、分页列表、配置写入、逻辑删除和 Gateway invoke 骨架。
- [x] 新增 AI 能力中心 handler，统一输出标准响应。
- [x] 将 AI 能力中心模型加入本地 AutoMigrate 兜底。
- [x] 将 AI 能力中心 API 注册进后端路由。
- [x] 将 AI 能力中心 API 纳入权限路由分类，避免启动门禁失败。
- [x] 补充引用校验、删除限制和策略优先级校验。
- [x] 补充后端单测或路由分类测试。

## Manifest 与应用装载

- [x] 新增 `internal/apps/ai_capability_center/app.manifest.yaml`。
- [x] 声明 8 个页面菜单：总览、供应商、模型目录、AI 场景、基础路由、用量统计、策略中心、系统设置。
- [x] 声明配置管理和 AI Gateway 调用权限。
- [x] 声明 API 权限矩阵、套餐功能点和配额。
- [x] 新增应用说明文档 `docs/apps/ai-capability-center.md`。
- [x] 验证 Manifest 可被运行目录扫描识别。
- [x] 验证 Manifest 装载 diff 无阻断项。

## 前端

- [x] 新增前端应用 Manifest 声明。
- [x] 新增前端 API 封装和类型定义。
- [x] 新增 AI 能力中心 Vue 页面，按原型转译信息架构和交互。
- [x] 挂载 8 个前端路由。
- [x] 页面接入真实 API 并处理 loading、empty、error。
- [ ] 复核移动端和桌面端布局不溢出、不重叠。

## 验证

- [x] 运行 `gofmt`。
- [x] 运行相关 Go 测试或至少编译验证。
- [x] 运行前端构建。
- [x] 运行 `git diff --check`。
- [x] 汇总已完成项、未完成风险和验证结果。
