# 企业级 SaaS Base V0.9 返修 TODO

## 目标

当前项目定位为企业级 SaaS Base V0.7/V0.8。本轮只做第一阶段返修，目标是推进到 V0.9，进入 V1.0 验收前置状态。

本轮不新增业务功能，不用文档修改掩盖实现缺口。P0 项必须全部完成。

## 执行顺序

1. Redis Session / JWT fallback 边界
2. 配额扣减并发安全
3. 文件上传安全增强
4. 交付 gate
5. 关键接口错误响应治理
6. 关键写操作事务一致性
7. user、plan/quota、file 优先 service 化

## P0-1 Redis 会话与 JWT 兜底边界

- [x] Redis session 模式下，Redis session 缺失、过期、删除时必须认证失败。
- [x] 只有 `jwtFallback=true` 时才允许解析 JWT fallback。
- [x] `jwtFallback=false` 时，即使请求携带合法签名 JWT，也必须认证失败。
- [x] Redis 不可用且 `jwtFallback=false` 时必须认证失败。
- [x] Redis 不可用且 `jwtFallback=true` 时允许 JWT fallback。
- [x] 生产环境 `jwtFallback=true` 启动配置校验必须失败。
- [x] 补测试：Redis session 存在时认证通过。
- [x] 补测试：Redis session 缺失但 JWT 合法时认证失败。
- [x] 补测试：`jwtFallback=false` 时不允许 fallback。
- [x] 补测试：生产环境 `jwtFallback=true` 启动失败。

## P0-2 配额扣减并发安全

- [x] 禁止 `used_value` 先查再内存加法再 update 的扣减模式。
- [x] 改为数据库事务 + 行锁，或条件 UPDATE 原子扣减。
- [x] 超配额返回稳定业务错误，不暴露 SQL 或内部实现。
- [x] 扣减失败不得增加 usage。
- [x] 配额扣减写入 usage log 或 audit log，便于追踪。
- [x] 补并发测试：并发扣减不会突破配额限制。
- [x] 补测试：超配额失败请求不会增加 usage。

## P0-3 文件上传安全

- [x] 增加 magic bytes / MIME sniffing，不能只信客户端 `Content-Type`。
- [x] 扩展名、服务端 sniff MIME、允许列表必须一致，或有明确映射规则。
- [x] `.jpg` 扩展但真实内容不匹配时必须拒绝。
- [x] 伪造 `Content-Type` 上传必须拒绝。
- [x] zip 上传增加最大文件数限制。
- [x] zip 上传增加最大解压后总大小限制。
- [x] zip 上传禁止路径穿越。
- [x] 生产环境必须显式配置 `UPLOAD_DIR`，不允许默认相对路径。
- [x] 预留病毒扫描 `interface`，本轮可先提供 noop 实现。
- [x] 文件下载继续保持租户、owner、权限校验。
- [x] 补测试：伪造 `Content-Type` 上传失败。
- [x] 补测试：jpg 扩展但真实内容不匹配失败。
- [x] 补测试：zip 路径穿越失败。
- [x] 补测试：生产环境未配置 `UPLOAD_DIR` 失败。

## P0-4 交付门禁

- [x] 增加 `make package`。
- [x] 增加 `make check-release`。
- [x] 交付产物排除 `.git/`。
- [x] 交付产物排除 `frontend/node_modules/`。
- [x] 交付产物排除 `frontend/dist/`。
- [x] 交付产物排除 `__MACOSX/`。
- [x] 交付产物排除 `.DS_Store`。
- [x] 交付产物排除临时日志、缓存、测试产物。
- [x] `make check-release` 检测到脏交付内容时必须失败阻断。
- [x] 补测试或脚本验证：`make check-release` 能阻断脏交付产物。

## P0-5 关键事务一致性

- [x] 用户创建 + 组织/岗位/角色关系写入必须在同一事务内完成。
- [x] 用户更新 + 组织/岗位/角色关系替换必须在同一事务内完成。
- [x] 用户导入批量写入必须具备事务边界，失败不得留下半成功数据。
- [x] 删除引用检查 + 删除动作必须在统一事务或统一保护入口内完成。
- [x] 角色授权必须在事务内完成，失败时原授权不得被破坏。
- [x] 套餐 feature 更新必须在事务内完成。
- [x] 套餐 quota 更新必须在事务内完成。
- [x] 事务内部禁止混用 `tx` 和 `h.db`。
- [x] 补测试：用户创建关系写入失败时主用户回滚。
- [x] 补测试：角色授权失败时原授权不被破坏。
- [x] 补测试：套餐配置失败时配置不半更新。
- [x] 补测试：删除引用对象失败时数据不变。

## P1-1 关键接口错误响应治理

本轮不要求清空全部 `err.Error()`，但以下模块关键接口必须先清理：

- [x] auth 关键接口不直接返回内部 `err.Error()`。
- [x] user 关键写接口不直接返回内部 `err.Error()`。
- [x] tenant 关键写接口不直接返回内部 `err.Error()`。
- [x] role 关键写接口不直接返回内部 `err.Error()`。
- [x] plan/quota 关键写接口不直接返回内部 `err.Error()`。
- [x] file 上传、下载、删除不直接返回内部 `err.Error()`。
- [x] import/export 不直接返回内部 `err.Error()`。
- [x] 引入 typed business error 或统一业务错误封装。
- [x] 内部错误写日志，响应只返回安全消息。
- [x] 补测试：SQL/GORM 错误不会进入响应 body。
- [x] 补测试：Redis 错误不会进入响应 body。
- [x] 补测试：文件系统错误不会进入响应 body。

## P1-2 优先服务化

本轮只优先迁移 user、plan/quota、file，避免大面积回归。

- [x] 建立真实 user application service，不接受空 service。
- [x] 用户创建、更新、删除、导入的事务、租户校验、配额校验进入 user service。
- [x] 建立真实 plan/quota application service，不接受空 service。
- [x] 套餐 feature/quota 更新、配额扣减、缓存失效进入 plan/quota service。
- [x] 建立真实 file application service，不接受空 service。
- [x] 文件上传安全校验、存储、元数据、审计进入 file service。
- [x] handler 只保留参数绑定、上下文提取、调用 service、返回 response。
- [x] repository 默认带 tenant scope。
- [x] 补 user service 单元测试。
- [x] 补 plan/quota service 单元测试。
- [x] 补 file service 单元测试。

## 本轮验收命令

- [x] `go test ./...`
- [x] `npm run build`
- [x] `npm run security:check`
- [x] `make check-release`

## 交付报告要求

- [x] 输出修改文件清单。
- [x] 输出每个 P0/P1 项的修复说明。
- [x] 输出新增测试清单。
- [x] 输出验收命令结果。
- [x] 输出仍未解决风险。

## 交付报告

### 修改文件清单

- `.gitignore`
- `Makefile`
- `go.mod`
- `go.sum`
- `scripts/check-delivery-clean.sh`
- `scripts/check-release.sh`
- `scripts/package-release.sh`
- `internal/bootstrap/config.go`
- `internal/bootstrap/config_test.go`
- `internal/application/file/service.go`
- `internal/application/file/service_test.go`
- `internal/application/quota/service.go`
- `internal/application/quota/service_test.go`
- `internal/application/user/service.go`
- `internal/application/user/service_test.go`
- `internal/infrastructure/persistence/postgres/repositories/tenant_scoped_repository.go`
- `internal/infrastructure/persistence/postgres/repositories/tenant_scoped_repository_test.go`
- `internal/interfaces/http/handlers/auth_context_helpers.go`
- `internal/interfaces/http/handlers/auth_security_test.go`
- `internal/interfaces/http/handlers/deletion_and_value_helpers.go`
- `internal/interfaces/http/handlers/file_batch_parity_test.go`
- `internal/interfaces/http/handlers/file_handler.go`
- `internal/interfaces/http/handlers/file_storage_helpers.go`
- `internal/interfaces/http/handlers/identity_handler.go`
- `internal/interfaces/http/handlers/parse_helpers.go`
- `internal/interfaces/http/handlers/plan_feature_handler.go`
- `internal/interfaces/http/handlers/plan_handler.go`
- `internal/interfaces/http/handlers/quota_handler.go`
- `internal/interfaces/http/handlers/quota_service.go`
- `internal/interfaces/http/handlers/quota_service_test.go`
- `internal/interfaces/http/handlers/transaction_integrity_test.go`
- `internal/interfaces/http/handlers/user_handler.go`
- `internal/interfaces/http/handlers/user_relation_helpers.go`
- `internal/interfaces/http/handlers/user_security_handler.go`
- `internal/interfaces/http/response/response.go`
- `internal/interfaces/http/response/response_test.go`
- `docs/ENTERPRISE_SAAS_BASE_REPAIR_TODO.md`

### P0/P1 修复说明

- P0-1：Redis session 模式下 session 缺失不再 fallback 到 JWT；只有 Redis 不可用且 `jwtFallback=true` 时允许 fallback；生产环境禁止 `AUTH_JWT_FALLBACK=true`。
- P0-2：配额扣减改为 service 内事务 + 原子 upsert，失败不增加 usage，并记录 quota audit。
- P0-3：文件上传增加服务端 MIME sniff、扩展名/MIME 映射、zip 数量/解压大小/路径穿越防护、生产 `UPLOAD_DIR` 校验和病毒扫描接口。
- P0-4：新增 `make package` 和 `make check-release`，交付包排除 `.git`、`frontend/node_modules`、`frontend/dist`、`.DS_Store`、`__MACOSX`、日志缓存等内容。
- P0-5：用户创建/更新/导入、删除引用检查、角色授权、套餐 feature/quota 更新进入事务边界，并补回滚测试。
- P1-1：统一响应层对 SQL/GORM/Redis/文件系统路径类错误脱敏，内部错误挂到 gin context，响应只返回安全消息。
- P1-2：完成 user、plan/quota、file 三个优先模块 service 化；新增 tenant-scoped repository，service 默认通过租户范围入口处理用户唯一性、文件、配额 usage 等租户数据。

### 新增测试清单

- Redis session 存在/缺失、JWT fallback 开关、生产 fallback 禁止测试。
- 配额并发扣减不超限、超配额不增加 usage 测试。
- 文件伪造 Content-Type、扩展名内容不匹配、zip 路径穿越、生产缺失 `UPLOAD_DIR` 测试。
- 用户创建关系写入失败回滚、用户导入失败回滚、角色授权失败回滚、套餐配置失败回滚、删除引用对象不变测试。
- user/quota/file service 单元测试。
- tenant-scoped repository 租户过滤测试。
- 统一错误响应脱敏测试。
- check-release 自测脏交付产物阻断。

### 验收命令结果

- `go test ./...`：通过。
- `npm run build`：通过。
- `npm run security:check`：通过。
- `make check-release`：通过。

### 仍未解决风险

- 本轮 service 化按返修要求只覆盖 user、plan/quota、file 三个优先模块，其他模块仍有继续 service 化和 repository 收敛空间。
- `identity_handler.go` 仍是历史大文件，已降低优先模块风险，但全量拆分仍建议作为下一阶段工程治理任务。
- 当前病毒扫描接口为 noop，实现真实 AV/对象存储扫描仍需接入外部安全组件。
