# C 端个人空间与账号体系数据库设计

## 1. 迁移文件

本次数据库变更由以下 migration 承载：

- `internal/infrastructure/persistence/postgres/migrations/000085_consumer_account_space_scope.up.sql`
- `internal/infrastructure/persistence/postgres/migrations/000085_consumer_account_space_scope.down.sql`

生产结构以 migration 和 `internal/infrastructure/persistence/postgres/schema/current_schema.sql` 为准。

## 2. 新增表

### 2.1 `account`

登录账户表，负责账号、手机号、邮箱和密码凭证。

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `bigserial` | 主键 |
| `login_account` | `varchar(128)` | 登录账号 |
| `phone` | `varchar(32)` | 手机号 |
| `email` | `varchar(200)` | 邮箱 |
| `password_hash` | `varchar(200)` | 密码哈希 |
| `status` | `bigint` | 状态，1 为启用 |
| `session_version` | `bigint` | 会话版本，用于强制失效 |
| `password_changed_at` | `timestamptz` | 密码变更时间 |
| `created_at` / `updated_at` / `deleted_at` | `timestamptz` | 审计与逻辑删除 |

关键索引：

- `uq_account_login_account_active`：未删除账号唯一。
- `uq_account_phone_active`：未删除手机号唯一。
- `uq_account_email_active`：未删除邮箱唯一。
- `idx_account_status`：状态筛选。

### 2.2 `user_identity`

自然人身份表，负责用户资料主档。

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `bigserial` | 主键 |
| `account_id` | `bigint` | 关联 `account.id` |
| `display_name` | `varchar(100)` | 显示名称 |
| `avatar_url` | `text` | 头像 |
| `phone` | `varchar(32)` | 手机号 |
| `email` | `varchar(200)` | 邮箱 |
| `status` | `bigint` | 状态 |
| `created_at` / `updated_at` / `deleted_at` | `timestamptz` | 审计与逻辑删除 |

关键索引：

- `idx_user_identity_account`：按账号查自然人身份。
- `idx_user_identity_phone`：按手机号查自然人身份。

## 3. 既有表扩展

### 3.1 `tenant`

新增字段：

| 字段 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `tenant_type` | `varchar(32)` | `enterprise` | 空间类型：`platform`、`enterprise`、`personal` |
| `owner_user_id` | `bigint` | 空 | 空间拥有者 `app_user.id` |

补偿逻辑：

- `is_platform_tenant=true` 的主体设置为 `platform`。
- 其他历史主体设置为 `enterprise`。

### 3.2 `app_user`

新增字段：

| 字段 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `account_id` | `bigint` | 空 | 关联登录账户 |
| `identity_user_id` | `bigint` | 空 | 关联自然人身份 |
| `member_type` | `varchar(32)` | `enterprise_member` | 空间成员类型 |

`member_type` 当前约定：

| 值 | 说明 |
|---|---|
| `platform_admin` | 平台管理员 |
| `enterprise_admin` | 企业管理员 |
| `enterprise_member` | 企业成员 |
| `personal_owner` | 个人空间拥有者 |

### 3.3 `permission`

新增字段：

| 字段 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `tenant_scope` | `varchar(32)` | `enterprise_only` | 菜单/权限适用空间范围 |

补偿逻辑：

- 原 `is_platform_only=true` 的权限标记为 `platform_only`。
- `/home`、`/profile`、`brand:edit` 标记为 `all`。
- `/data-center` 系列和 `/integration-center/my-connections` 标记为 `enterprise_personal`。
- 其他历史权限保持 `enterprise_only`。

### 3.4 `sys_app_entry` 和 `sys_app_permission`

新增字段：

| 表 | 字段 | 默认值 | 说明 |
|---|---|---|---|
| `sys_app_entry` | `tenant_scope` | `enterprise_only` | 应用菜单适用空间范围 |
| `sys_app_permission` | `tenant_scope` | `enterprise_only` | 应用 API/操作权限适用空间范围 |

补偿逻辑与 `permission` 保持一致，`data-center` 应用权限整体标记为 `enterprise_personal`。

## 4. 回滚说明

`down.sql` 会移除本次新增索引、字段和表。回滚前必须确认：

- 是否已有 C 端注册用户。
- 是否已有个人空间业务数据。
- 是否已有应用菜单或权限依赖 `tenant_scope`。

若生产环境已有个人空间数据，不建议直接回滚，应先导出或迁移个人空间数据。

## 5. 后续数据库增强

- 第三方登录需要新增外部身份绑定表，例如 `account_external_identity`。
- Account 绑定/解绑需要新增绑定历史或安全审计表。
- 手机号变更需要新增验证码、旧手机号确认和变更流水。
- 个人升级企业空间需要补充空间类型转换记录和数据迁移状态表。
