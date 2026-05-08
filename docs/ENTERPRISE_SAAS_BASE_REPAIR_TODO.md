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

## P0-1 Redis Session / JWT Fallback

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

## P0-4 交付 Gate

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

## P1-2 优先 Service 化

本轮只优先迁移 user、plan/quota、file，避免大面积回归。

- [x] 建立真实 user application service，不接受空 service。
- [ ] 用户创建、更新、删除、导入的事务、租户校验、配额校验进入 user service。
- [x] 建立真实 plan/quota application service，不接受空 service。
- [ ] 套餐 feature/quota 更新、配额扣减、缓存失效进入 plan/quota service。
- [x] 建立真实 file application service，不接受空 service。
- [ ] 文件上传安全校验、存储、元数据、审计进入 file service。
- [ ] handler 只保留参数绑定、上下文提取、调用 service、返回 response。
- [ ] repository 默认带 tenant scope。
- [x] 补 user service 单元测试。
- [x] 补 plan/quota service 单元测试。
- [x] 补 file service 单元测试。

## 本轮验收命令

- [x] `go test ./...`
- [x] `npm run build`
- [x] `npm run security:check`
- [x] `make check-release`

## 交付报告要求

- [ ] 输出修改文件清单。
- [ ] 输出每个 P0/P1 项的修复说明。
- [ ] 输出新增测试清单。
- [ ] 输出验收命令结果。
- [ ] 输出仍未解决风险。
