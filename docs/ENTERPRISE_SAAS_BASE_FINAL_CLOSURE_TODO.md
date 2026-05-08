# 企业级 SaaS Base V0.9 最后一轮收口 TODO

## 目标

当前项目已完成第一轮 V0.9 返修，但仍不能标记为 V1.0。本轮只做底座阻断项收口，不新增商品、订单、库存、广告、客服、AI Agent 等电商业务功能。

完成本轮后，目标状态为：**企业级 SaaS Base V0.9，可进入电商 AI 运营 OS 业务模块开发阶段**。

## 执行原则

- 只修阻断业务开发的底座问题。
- 不引入新的业务菜单、页面或电商领域模型。
- 继续按完成一项勾一项执行。
- `response.SafeMessage` 只能作为兜底，不能替代关键接口的显式错误治理。
- 本轮修改必须保持 `go test ./...`、前端干净构建、安全检查、交付 gate 全部通过。

## P0-1 交付包收口

- [x] 新增 `make check-source-clean`。
- [x] `check-source-clean` 检查当前源码目录中不得存在 `.git/` 作为交付内容。
- [x] `check-source-clean` 检查不得存在 `frontend/node_modules/`。
- [x] `check-source-clean` 检查不得存在 `frontend/dist/`。
- [x] `check-source-clean` 检查不得存在 `dist/` 交付污染目录，允许脚本生成后再检查 artifact。
- [x] `check-source-clean` 检查不得存在 `__MACOSX/`。
- [x] `check-source-clean` 检查不得存在 `.DS_Store`。
- [x] `check-source-clean` 检查不得存在 `*.log`。
- [x] `check-source-clean` 检查不得存在 `.cache/`。
- [x] `make check-release` 串联执行 self-test、source clean check、package、artifact check。
- [x] `make package` 输出最终 artifact 路径。
- [x] `make check-release` 输出 artifact 内容列表，作为验收证据。
- [x] 交付报告明确：最终只交付 `make package` 生成的 artifact，禁止 Finder/手工压缩整个源码目录。

## P0-2 前端 Build 可复现

- [x] 删除并禁止依赖压缩包中的 `frontend/node_modules`。
- [x] 使用干净依赖流程验证：`cd frontend && rm -rf node_modules dist && npm ci`。
- [x] 使用干净依赖流程验证：`npm run build`。
- [x] 使用干净依赖流程验证：`npm run security:check`。
- [x] 记录 Node 版本。
- [x] 记录 npm 版本。
- [x] 如 Vite/Rolldown optional dependency 存在平台不稳定风险，锁定兼容版本或在文档声明 Node/npm 版本矩阵。
- [x] 交付报告包含前端干净构建真实输出摘要。

## P0-3 关键接口错误响应治理

优先清理以下文件中的 `response.Error(..., err.Error())`：

- [x] `tenant_handler.go`
- [x] `tenant_subscription_handler.go`
- [x] `role_handler.go`
- [x] `plan_handler.go`
- [x] `plan_feature_handler.go`
- [x] `quota_handler.go`
- [x] `file_handler.go`
- [x] `business_unit_handler.go`
- [x] `dict_param_handler.go`
- [x] `csv_helpers.go`

治理要求：

- [x] 新增或完善统一错误响应 helper，例如 `respondBusinessError` / `respondInternalError`。
- [x] 业务错误返回明确中文业务提示。
- [x] 内部错误进入 `gin.Context.Errors` 或日志，不进入 response body。
- [x] SQL/GORM/Redis/文件路径类错误不得出现在 response body。
- [x] `grep -R "response.Error(.*err.Error" -n internal/interfaces/http/handlers` 在关键模块清零，或仅保留经过明确业务错误包装的例外。
- [x] 补测试：关键 handler 返回体不包含 SQL/GORM 错误文本。
- [x] 补测试：关键 handler 返回体不包含 Redis 错误文本。
- [x] 补测试：关键 handler 返回体不包含文件系统路径错误文本。

## P0-4 文件删除一致性

- [x] `DeleteFile()` 不得忽略 `SoftDelete()` 错误。
- [x] 文件移动失败时返回错误，不更新 DB。
- [x] DB soft delete 失败时不能返回 OK。
- [x] 文件已移动但 DB 更新失败时，必须尝试从 trash 补偿移回原路径。
- [x] 删除审计失败时行为必须明确：要么整体失败并补偿，要么记录失败审计并返回明确错误。
- [x] 删除编排迁移到 file service：物理移动、DB 状态、审计由 service 统一编排。
- [x] 补测试：DB soft delete 失败时不返回 OK。
- [x] 补测试：文件移动成功但 DB 失败时会补偿移回。
- [x] 补测试：删除成功时 DB、文件位置、audit 三者一致。

## P0-5 导入失败不得消耗导入次数配额

- [x] `ImportUsersCSV()` 在 CSV 解析失败时不得扣减 `daily_import_times`。
- [x] 必填字段缺失时不得扣减 `daily_import_times`。
- [x] 业务校验失败时不得扣减 `daily_import_times`。
- [x] DB 写入失败时不得扣减 `daily_import_times`，或必须补偿回滚 usage。
- [x] 导入成功后才扣减一次 `daily_import_times`。
- [x] 导入流程由 user service 统一编排：解析后的用户数据 -> 事务写入 -> 成功后扣减导入配额。
- [x] 补测试：CSV 格式错误不扣配额。
- [x] 补测试：必填字段缺失不扣配额。
- [x] 补测试：DB 写入失败不扣配额。
- [x] 补测试：导入成功扣一次配额。

## P0-6 收敛 File Service 重复实现

- [x] 删除或废弃 handler 侧旧 `validateUploadFile()`。
- [x] 删除或废弃 handler 侧旧 `validateZipUpload()`。
- [x] 删除或废弃 handler 侧旧 `saveUploadedFile()`。
- [x] 删除或废弃 handler 侧旧 `uploadVirusScanner`。
- [x] 上传安全规则只保留 `application/file.Service` 一个主入口。
- [x] 旧 handler 上传安全测试迁移到 `internal/application/file`。
- [x] 保留 handler 侧路径/租户访问辅助函数时，不得重复 MIME/zip/virus scanner 规则。

## P1-1 跨租户审计收口

- [x] `requestTenantID()` 使用 `tenantContextForUser()`。
- [x] 普通租户用户 query `tenant_id` 无法越权。
- [x] 普通租户用户 body/path 伪造 `tenant_id` 继续被拒绝或无效。
- [x] 平台管理员通过 query 切换 tenant 时记录 cross-tenant audit。
- [x] cross-tenant audit 至少记录 actor_user_id、actor_tenant_id、target_tenant_id、method、path、request_id。
- [x] 平台跨租户能力仍必须受 route permission 控制，不允许只靠前端隐藏。
- [x] 补测试：普通用户 query `tenant_id` 无法越权。
- [x] 补测试：平台管理员跨租户访问会记录 audit。
- [x] 补测试：无权限平台用户访问平台接口被拒绝。

## P1-2 ReferenceGuard 统一

- [x] 梳理 `ReferenceGuard` 与 `blockDeleteIfReferenced()` 并存点。
- [x] 保留一个删除保护主入口。
- [x] 删除保护失败消息统一为业务错误。
- [x] 删除保护内部错误不泄露到 response body。
- [x] 补测试：被引用对象删除失败且数据不变。

## P1-3 Tenant-Scoped Repository 扩展准入

- [x] 当前 user/file/quota usage 已使用 tenant-scoped repository 的路径保留并补齐测试。
- [x] 文档明确新增业务模块 repository 默认必须 tenant scoped。
- [x] 新增电商业务模块前必须提供 tenant scope 测试。
- [x] 第一批后续可扩展对象：BusinessUnit、OrgNode、Role，本轮不强制全量迁移。

## 不做范围

- [x] 不新增商品模块。
- [x] 不新增订单模块。
- [x] 不新增库存模块。
- [x] 不新增广告模块。
- [x] 不新增客服模块。
- [x] 不新增 AI Agent 业务模块。
- [x] 不新增电商菜单、页面、流程。

## 验收命令

- [x] `go test ./...`
- [x] `cd frontend && rm -rf node_modules dist && npm ci && npm run build && npm run security:check`
- [x] `make check-source-clean`
- [x] `make check-release`

## 交付报告要求

- [x] 输出修改文件清单。
- [x] 输出每个 P0/P1 项的修复说明。
- [x] 输出新增测试清单。
- [x] 输出验收命令真实结果。
- [x] 输出 Node/npm 版本。
- [x] 输出最终 artifact 路径。
- [x] 输出 artifact 内容检查结果。
- [x] 判断当前是否可作为电商 AI 运营 OS 的业务开发底座。
- [x] 输出仍未解决但不阻断业务开发的风险。

## 本轮交付报告

### 修改文件摘要

- `Makefile`、`scripts/check-source-clean.sh`、`scripts/check-release.sh`、`scripts/package-release.sh`：补齐源码清洁检查、release gate 串联、artifact 内容输出。
- `internal/application/file/service.go`、`internal/application/file/service_test.go`：文件删除由 service 统一编排，覆盖文件移动、DB soft delete、审计及失败补偿。
- `internal/application/user/service.go`、`internal/application/user/service_test.go`、`internal/interfaces/http/handlers/file_handler.go`：用户导入成功后才扣减 `daily_import_times`，失败路径不消耗导入次数。
- `internal/interfaces/http/handlers/error_response_helpers.go`、核心 handler 文件：关键接口不再直接返回 `err.Error()`，内部错误进入 `gin.Context.Errors`。
- `internal/interfaces/http/handlers/file_storage_helpers.go`、`internal/interfaces/http/handlers/file_batch_parity_test.go`：移除 handler 侧旧 MIME/zip/virus scanner/upload save 重复实现，上传安全规则集中到 `application/file.Service`。
- `internal/interfaces/http/handlers/parse_helpers.go`、`internal/interfaces/http/handlers/final_closure_test.go`：`requestTenantID()` 接入 `TenantContext`，平台管理员跨租户访问写入 audit。
- `internal/interfaces/http/handlers/deletion_and_value_helpers.go`、`internal/interfaces/http/handlers/reference_guard.go`：删除保护统一为 `blockDeleteIfReferenced()` / `checkDeletionReferences()` 主入口，删除闲置 `ReferenceGuard` 类型。
- `docs/tech_design/业务开发标准.md`、`docs/security/delete-policy.md`：补充 tenant-scoped repository 准入和删除保护主入口说明。

### P0/P1 修复说明

- P0-1：`make check-source-clean` 已加入，检查源码目录中的交付污染；`make check-release` 已串联 self-test、source clean、package、artifact check，并输出 artifact 内容。
- P0-2：已按干净依赖流程执行 `rm -rf node_modules dist && npm ci && npm run build && npm run security:check`；Node `v25.9.0`，npm `11.12.1`。
- P0-3：关键 handler 中 `response.Error(... err.Error())` 已清零，新增安全错误响应 helper；SQL/GORM/Redis/文件路径类错误不进入 response body。
- P0-4：`DeleteFile()` 不再忽略 DB/审计错误；文件已移动但 DB 失败时会尝试补偿移回原路径。
- P0-5：`ImportUsersCSV()` 删除入口处预扣导入次数；解析后的导入由 user service 事务写入，创建成功后才扣一次 `daily_import_times`。
- P0-6：handler 侧旧上传校验、zip 校验、保存函数、scanner 变量已移除，上传安全规则由 `application/file.Service` 单一入口承担。
- P1-1：`requestTenantID()` 统一使用 `tenantContextForUser()`，平台管理员 query 切租记录 `cross_tenant_access` audit。
- P1-2：删除保护主入口收敛为 `blockDeleteIfReferenced()` / `checkDeletionReferences()`，删除保护失败消息统一为业务错误。
- P1-3：现有 user/file/quota tenant-scoped repository 测试保留；业务开发标准明确新增业务 repository 默认 tenant scoped，新增电商模块前必须补 tenant scope 测试。

### 新增/调整测试

- `TestDeleteMovesFileAndMarksMetadataDeleted`
- `TestDeleteRestoresFileWhenMetadataUpdateFails`
- `TestImportUsersRollsBackBatchWhenOneRowFails`
- `TestImportUsersConsumesDailyQuotaOnlyAfterSuccessfulCreate`
- `TestCrossTenantAccessWritesSingleAuditLog`
- `TestInternalErrorMessageDoesNotLeakToResponseBody`
- `TestResponseSafeMessageMasksFilesystemError`
- `TestDeleteReferenceCheckKeepsReferencedPlan`

### 验收命令结果

- `go test ./...`：通过。
- `cd frontend && rm -rf node_modules dist && npm ci && npm run build && npm run security:check`：通过；`npm ci` 报告 4 个 npm audit 漏洞，项目自定义 `security:check` 通过。
- `make check-source-clean`：通过。
- `make check-release`：通过。

### Artifact

- 最终 artifact：`dist/release/saas_baseon_go.tar.gz`
- artifact 检查结果：未发现 `.git`、`frontend/node_modules`、`frontend/dist`、`__MACOSX`、`.DS_Store`、日志或缓存污染；`check-release` 已输出内容列表作为证据。

### 当前判断

当前项目可作为“电商 AI 运营 OS”业务模块开发底座进入下一阶段，但仍应定位为 **企业级 SaaS Base V0.9**，不是最终 V1.0 可验收态。

### 剩余但不阻断业务开发的风险

- 仍建议继续逐步 service 化 BusinessUnit、OrgNode、Role 等模块，减少 handler 承载。
- npm audit 仍报告间接依赖漏洞，当前自定义安全检查通过，但后续应结合前端依赖升级窗口治理。
- OpenAPI 契约和更细粒度的跨租户平台权限模型仍可继续增强。
