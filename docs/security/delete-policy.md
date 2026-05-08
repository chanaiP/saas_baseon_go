# 删除治理说明

本项目的业务对象默认不得物理删除。用户、角色、权限、租户、组织、业务单元、字典、参数、文件等可追溯业务数据必须使用 `deleted_at`、`status`、停用或归档表达删除结果，并由 `ReferenceGuard` 在删除前完成引用检查。

允许物理删除的范围仅限关系重建表和租户覆盖表，例如：

- `user_role`、`role_permission`、`app_user_position`、`app_user_department`
- `saas_plan_feature`、`saas_plan_quota`
- `tenant_feature_override`、`tenant_quota_override`
- `permission_custom_department`、`permission_custom_user`
- `tenant_dict_item_override`、`tenant_param_value`

这些表表达“当前关系快照”或“当前覆盖配置”，重建时必须满足：

- 在数据库事务内先删除旧关系，再写入新关系，避免部分成功。
- 上层业务对象本身不得物理删除，必须保留审计记录。
- 重建完成后触发对应授权、菜单、套餐或租户缓存失效。
- 新增物理删除点必须先确认属于关系重建或覆盖值，不得用于业务主表。

当前自动化覆盖：

- `go test ./...` 覆盖角色逻辑删除、关系重建事务、权限/菜单矩阵和引用检查。
- `go run ./cmd/verify-bootstrap` 覆盖关键关系孤儿数据检查。
