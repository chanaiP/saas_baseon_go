# SaaS Baseon 文档索引

> 当前仓库是 Go + Gin + GORM + PostgreSQL 技术栈重建版。产品需求、菜单、权限、测试用例保持原语义；技术实现以 [Go DDD/TDD 技术架构方案](tech_design/Go-DDD-TDD技术架构方案.md) 为主依据。

本文档目录只保留当前有效交付文档，历史会话、阶段报告和临时设计稿不作为当前实现依据。

## 文档入口

- [项目总体介绍](项目总体介绍.md)：项目定位、能力范围和面向业务方的总体说明。
- [需求设计](req_design)：系统管理各菜单的产品需求、业务规则和验收关注点。
- [技术设计](tech_design)：总体技术方案、业务开发标准和系统管理各菜单技术实现说明。
- [Go DDD/TDD 技术架构方案](tech_design/Go-DDD-TDD技术架构方案.md)：Go 架构、DDD 分层、TDD 策略和迁移顺序。
- [数据库设计](sql_design)：系统管理模块 SQL / 数据库设计文档。
- [测试用例](test_cases)：系统管理模块功能测试用例和业务功能接入回归清单。

## 相关入口

- 根目录 `README.md`：开发启动、Docker、端口、演示账号和故障排查。
- 根目录 `AGENTS.md`：AI coding agent 项目导航。
- 根目录 `CLAUDE.md`：工程规范、安全规则、租户隔离、权限、逻辑删除和发布验证规则。
- `internal/infrastructure/persistence/postgres/migrations/`：版本化 SQL migration。
- `internal/infrastructure/persistence/postgres/schema/current_schema.sql`：当前 PostgreSQL schema 基线。
