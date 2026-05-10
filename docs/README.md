# SaaS Baseon 文档索引

当前仓库是 SaaS Baseon 的 Go 技术栈重建版。产品需求、菜单、权限、测试用例保持原语义；当前实现以 Go + Gin + GORM + PostgreSQL + Redis 为后端基础，以 Vue 3 + Element Plus 为前端管理台基础。

本文档目录只保留当前项目可继续参考的文档。历史阶段报告用于追溯，不作为新增实现依据；如果文档与代码冲突，以当前代码为准，并同步修正文档。

## 推荐阅读顺序

1. [项目总体介绍](项目总体介绍.md)：了解项目定位、业务范围和系统能力。
2. [当前实现总览](当前实现总览.md)：了解当前代码实现、运行拓扑、核心链路和验证入口。
3. [总体技术方案](tech_design/总体技术方案.md)：了解整体架构、前后端分层、多租户、权限、套餐、数据权限和安全设计。
4. [业务开发标准](tech_design/业务开发标准.md)：新增业务模块前必须阅读，确认租户、权限、套餐、配额、审计和测试准入要求。
5. [模块接入准入清单](tech_design/模块接入准入清单.md)：新增或修改业务能力时逐项检查。
6. [最终交接报告](FINAL_PARITY_REPORT.md)：查看当前 Go 重建版与原项目的最终对齐状态。

## 当前实现边界

当前系统定位为企业级 SaaS Base 的管理后台底座，已覆盖：

- 主体管理、套餐中心、组织架构、岗位管理、业务单元、用户管理、角色权限、菜单管理、数据字典、参数管理、操作日志、登录日志。
- 多租户隔离、平台专属能力、租户菜单覆盖、角色权限、按钮权限、数据权限。
- 套餐功能点、配额项、订阅、租户功能覆盖、租户配额覆盖。
- 登录、验证码、Redis session、JWT 显式降级、登录失败保护、生产配置自检。
- 文件上传下载删除、用户导入导出、审计日志、发布包清洁检查。

当前不把以下内容作为已完成商业化产品能力：

- 订单、支付、发票、自动续费、计费账单。
- 多应用市场、插件中心、第三方应用授权。
- 完整告警中心、全链路追踪平台和多区域部署编排。

## 文档目录

### 需求设计

目录：[req_design](req_design)

记录各系统管理菜单的产品需求、业务规则、页面行为和验收关注点。

### 技术设计

目录：[tech_design](tech_design)

记录总体技术架构、业务开发标准、模块接入准入要求，以及各系统管理菜单的前端、后端、数据模型、权限、审计和交互实现说明。

### 数据库设计

目录：[sql_design](sql_design)

记录系统管理模块的数据表、字段、索引、约束、逻辑删除和风险说明。生产结构以 `internal/infrastructure/persistence/postgres/schema/current_schema.sql` 和版本化 migration 为准。

### 测试用例

目录：[test_cases](test_cases)

记录各模块功能、表单、权限、数据权限、接口、数据库约束、审计和边界场景测试用例。

### 安全与治理

目录：[security](security)

记录 token 存储、逻辑删除、迁移治理等安全和工程治理要求。

## 根目录文档

- [README](../README.md)：开发启动、端口、Docker、演示账号、初始化、验证命令和整体目录。
- [AGENTS](../AGENTS.md)：AI coding agent 项目导航。
- [CLAUDE](../CLAUDE.md)：工程执行规则、安全红线、多租户、权限、逻辑删除、API 契约和验证规则。
- [当前实现总览](当前实现总览.md)：当前代码实现、运行拓扑、治理链路与验证入口。

## 历史追溯文档

- [原项目等价补齐 TODO](ORIGINAL_PARITY_TODO.md)
- [原项目等价对比报告](ORIGINAL_PARITY_REPORT.md)
- [最终交接报告](FINAL_PARITY_REPORT.md)
- [企业级 SaaS 底座优化 TODO](ENTERPRISE_SAAS_BASE_TODO.md)
- [企业级 SaaS Base V0.9 返修 TODO](ENTERPRISE_SAAS_BASE_REPAIR_TODO.md)
- [企业级 SaaS Base V0.9 最后一轮收口 TODO](ENTERPRISE_SAAS_BASE_FINAL_CLOSURE_TODO.md)

这些文档用于理解整改过程和验收背景；后续开发应以 `README.md`、`CLAUDE.md`、`tech_design/业务开发标准.md`、`tech_design/模块接入准入清单.md` 和当前代码为准。

## 文档维护规则

- 改接口时同步更新接口说明、前端 API 调用说明、权限码和测试用例。
- 改菜单、操作或套餐功能点时，同步更新菜单管理、角色权限、套餐中心和模块接入文档。
- 改数据库结构时，同步更新 SQL 设计、schema baseline、migration 和回滚/补偿说明。
- 改安全策略时，同步更新 `CLAUDE.md` 与 `docs/security`。
- 文档修改后至少执行 `git diff --check`。
