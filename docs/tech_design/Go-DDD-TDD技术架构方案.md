# Go DDD/TDD 技术架构方案

## 1. 目标

本项目在不改变产品需求、业务规则、菜单结构、权限语义和前端交互的前提下，将后端技术栈调整为：

- Go 1.22+ / Gin
- GORM
- PostgreSQL 16
- Redis 7
- DDD 分层
- TDD 优先

前端继续沿用 Vue 3 + TypeScript + Element Plus + Pinia + Vue Router + Vite。

## 2. 工程结构

```text
cmd/api                         Gin API 入口
internal/bootstrap              配置、数据库、Redis、路由装配
internal/domain                 领域对象、领域服务、仓储接口、领域错误
internal/application            用例编排、事务边界、权限/套餐/审计协调
internal/interfaces/http        Gin handler、middleware、DTO、统一响应
internal/infrastructure         GORM/PostgreSQL、Redis、外部依赖实现
frontend                        管理端前端
docs                            需求、技术、数据库、测试文档
```

## 3. 分层规则

| 层 | 责任 | 禁止事项 |
|---|---|---|
| Handler | HTTP 入参、鉴权中间件、响应包装 | 写业务规则、直接拼复杂 SQL |
| Application | 用例编排、事务、审计、权限、套餐校验 | 依赖 Gin Context |
| Domain | 业务规则、实体不变量、领域错误 | 依赖数据库、Redis、HTTP |
| Repository | GORM 持久化、查询封装 | 决定业务是否允许执行 |
| Infrastructure | PostgreSQL、Redis、JWT、外部服务 | 反向依赖 application |

## 4. 领域划分

| 领域 | 包 | 覆盖业务 |
|---|---|---|
| 租户域 | `tenant` | 主体、品牌、订阅、套餐覆盖 |
| 身份域 | `identity` | 登录、用户、密码、会话、验证码 |
| 权限域 | `permission` | 菜单、按钮、角色、数据权限 |
| 组织域 | `organization` | 组织架构、岗位、业务单元 |
| 系统域 | `system` | 字典、参数、操作日志、登录日志 |
| 监控域 | `monitor` | 健康检查、服务状态、缓存状态 |

## 5. API 兼容

后端响应保持当前前端契约：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

分页响应保持：

```json
{
  "items": [],
  "total": 0
}
```

接口路径优先保持原路径，不因技术栈调整改变产品行为。

## 6. 数据库策略

数据库目标为 PostgreSQL 16。结构演进使用版本化 migration，GORM AutoMigrate 仅作为开发兜底，不作为生产变更依据。

类型映射原则：

| 原语义 | PostgreSQL |
|---|---|
| 自增主键 | `BIGSERIAL` 或 identity |
| 时间 | `TIMESTAMPTZ` |
| 布尔 | `BOOLEAN` |
| JSON | `JSONB` |
| 软删除 | `deleted_at TIMESTAMPTZ` |

## 7. TDD 流程

新增或迁移模块按以下顺序：

```text
先写 domain test
-> 写 application service test
-> 写 repository integration test
-> 写 handler contract test
-> 实现代码
-> 接入前端或 E2E
```

必须覆盖：

- 租户隔离
- 菜单/按钮权限拒绝
- 套餐功能拒绝
- 配额拒绝
- 软删除过滤
- 审计日志写入
- 关键字段唯一性

## 8. 迁移顺序

建议迁移顺序保持低风险到高风险：

1. 健康检查、监控、参数管理、数据字典
2. 操作日志、登录日志
3. 岗位管理、组织架构、业务单元
4. 主体管理、套餐中心
5. 用户管理、角色权限、菜单管理
6. 登录认证、Redis session、数据权限核心

## 9. 当前状态

当前新项目已完成：

- Gin API 启动入口、配置加载、Request ID 中间件、统一响应
- PostgreSQL + Redis + API + Web 的 Docker Compose 本地运行环境
- GORM AutoMigrate 开发期建表，当前覆盖主体、套餐、组织、岗位、业务单元、用户、角色、权限、字典、参数、日志、订阅、配额等核心表
- 初始化数据：平台主体、演示账号、管理员角色、菜单/按钮权限、套餐、功能、配额、字典、参数、登录日志、操作日志
- 登录认证：账号/工号/手机号登录、密码校验、Bearer token 签发、个人密码修改
- 前端可见系统管理页面的后端 API 覆盖：主体管理、套餐中心、组织架构、岗位管理、业务单元、用户管理、角色权限、菜单管理、数据字典、参数管理、操作日志、登录日志
- 监控页面 API 覆盖：健康检查、服务器信息、定时任务、服务监控、Redis 缓存统计、Redis key 扫描
- OpenAPI 基础文档输出，支持开发者中心接口卡片展示
- 前端项目复制与 Web 容器接入
- 原项目 `docs` 文档复制并更新为 Go/PostgreSQL 技术栈方向
- 验证结果：`go test ./...` 通过，`npm run build` 通过，Playwright E2E `41 passed`

当前仍建议在下一阶段继续增强：

- 将 handler 中仍偏通用 CRUD 的实现继续下沉到 application service 和 repository，收敛 DDD 边界
- 为登录、权限、租户隔离、套餐限制、审计写入补充更完整的 handler/repository/integration 测试
- 将 OpenAPI 从手写摘要升级为完整 schema
- 将开发期 `AutoMigrate` 固化为生产可审计 migration
