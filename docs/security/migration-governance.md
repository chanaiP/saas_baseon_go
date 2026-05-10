# 数据库迁移治理

## 环境规则

- 开发环境可以使用 `DB_AUTO_MIGRATE=true` 做本地快速迭代，但版本化 SQL migration 仍是数据库结构的权威来源。
- 测试、预发和生产环境必须使用 `cmd/migrate`，并设置 `DB_AUTO_MIGRATE=false`。
- 生产环境建议设置 `MIGRATION_STRICT_CHECKSUM=true`；已经执行过的 migration 如果 checksum 与本地文件不一致，必须立即失败并停止部署。

## Checksum 失败处理

当 checksum 校验失败时：

1. 停止部署，保持上一版本应用继续运行。
2. 不要修改已经在目标环境执行过的 migration 文件来“对齐生产”。
3. 追加一个新的正向 migration 修复结构或数据。
4. 如果 migration 被误执行到错误环境，优先从备份恢复，或执行经过评审的补偿 migration。
5. 重新执行 `go run ./cmd/migrate` 和 `go run ./cmd/verify-bootstrap`。

## 回滚与补偿

Down migration 主要用于本地和预发环境恢复。生产环境回滚优先选择备份恢复或追加补偿型正向 migration，因为生产数据安全优先级高于回滚速度。
