# Original Parity Report

## 路由对比

对比时间：2026-05-07

对比范围：

- 原项目：`/Users/Shared/aiproject/Baseon/Saas_Baseon`
- Go 版：`/Users/chen.ai/project/saas_baseon_go`

结果：

- 原项目后端路由：136 个。
- Go 版后端路由：146 个。
- 原项目路由在 Go 版中缺失：0 个。
- Go 版额外路由：10 个，均为兼容别名或开发入口。

Go 版额外路由：

- `GET /api/business-units/org-mappings`
- `GET /api/dict-types/by-code`
- `GET /api/params`
- `GET /api/params/{key}`
- `GET /api/permission-menu-bundles`
- `GET /api/public`
- `GET /openapi.json`
- `POST /api/params`
- `POST /api/tenants/with-package`
- `PUT /api/permissions/menu-data-perm-mode`

处理原则：

- 额外兼容别名不作为新增产品功能使用。
- 这些入口保留时必须指向已有产品逻辑，不引入新菜单、新页面或新业务流程。

## 前端文件对比

结果：

- `frontend/src` 页面、组件、API、store、utils 的文件集合与原项目一致。
- Go 版新增 `frontend/src/utils/debug.ts` 仅用于静默替代开发期 `console.log`，不改变产品逻辑。

后续重点：

- 文件集合已一致，下一阶段转向行为级对比：接口参数、返回结构、校验、权限、数据范围和测试覆盖。

## 认证与会话对比

已对齐项：

- 登录账号匹配范围保持原项目逻辑：账号、工号、手机号，并支持 `tenant_id` / `tenant_code` 限定主体。
- 手机号匹配多个可登录主体时返回 `code=2`、`message=请选择主体` 和主体列表，前端可继续使用原选择主体流程。
- 登录失败 3 次后要求验证码；验证码 Redis 命名空间使用 `auth:captcha:`，兼容历史 `captcha:` key。
- IP 登录失败 15 次后按原项目返回 `请求过于频繁，请 {ttl} 秒后重试`。
- 登录失败同时累计账号失败次数和 IP 失败次数；登录成功清理两类计数。
- token 优先使用 Redis opaque session，key 为 `auth:session:{token}`；JWT 仅作为 Redis 不可用时的兼容降级。
- 退出登录删除 Redis session；切换主体删除旧 session，并按同一手机号在目标主体下的账号重新签发 session。
- 禁用账号、禁用主体、过期或冻结订阅均禁止登录，且写入登录日志。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

仍在后续清单中：

- B 组认证与会话已完成。

新增 B4 对齐项：

- 新写入密码哈希切换为原项目 `pbkdf2_sha256$salt$hash` 格式。
- 修改密码要求验证码，验证码错误返回 `验证码错误或已过期，请刷新验证码后重试`。
- 修改密码的新密码长度、字母数字强度、确认一致、新旧密码不同等规则与原项目一致。
- 旧密码错误 5 次后锁定 15 分钟，Redis key 使用 `auth:pwd_fail:{user_id}` / `auth:pwd_block:{user_id}`。
- 管理员重置用户密码、重置主体主管理员密码均生成 14 位字母数字随机串。

## 用户管理对比

已对齐项：

- 用户列表按 `skip` / `limit` 分页，支持 `keyword` / `kw`、`status`、`company_id`、`department_id` 筛选，并排除软删除用户。
- 手机号按原项目规则规范化：允许数字、空格、短横线、括号，保存时转为 10-15 位纯数字。
- 创建用户支持 `department_ids`，去重后与 `department_id` 合并；主部门和公司继承逻辑与原项目一致。
- 创建/更新用户时校验部门、岗位、角色必须属于当前主体且未删除。
- 更新用户时同步角色、部门、岗位关系；部门清空时主部门同步清空。
- 非平台管理员不能设置系统管理员；系统内至少保留一名平台管理员。
- 删除用户改为软删除：设置 `deleted_at`、停用用户，并 tombstone 工号、账号、手机号以释放唯一值。
- 删除当前登录用户会被拒绝；重置密码继续按 B4 的随机密码策略执行。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

- 快捷入口偏好通过 `user_preference` 持久化，`GET/PUT /api/users/me/preferences` 与原项目返回 `shortcut_ids`。
- `/api/users/me` 返回角色、权限码、是否平台主体、快捷入口、订阅、功能和配额上下文。
- 非平台主体、非平台管理员的权限码按订阅启用功能过滤，避免套餐外按钮/接口权限泄漏。

状态：

- C 组用户管理已完成。

## 主体管理与品牌对比

已对齐项：

- 主体创建改为事务执行：主体、根公司、超级管理员角色、主管理员、用户角色关系任一步失败都会回滚。
- 主管理员手机号按原项目规则规范化，根公司使用主体编码大写并设置 `GROUP` 公司类型。
- 合并创建主体并配置套餐改为同一事务提交，套餐阶段失败不会留下半成品主体。
- 主体列表支持分页，排除软删除主体，并返回套餐名称/编码、联系人、启用公司/门店/业务单元/用户用量。
- 主体详情中的联系人优先读主体联系人，缺失时回退到主管理员信息。
- 主体停用会清理该主体 Redis 登录 session；主体删除改为软删除并 tombstone 编码。
- 删除主体前阻止仍存在平台管理员账号的主体，错误信息与原项目一致。
- 品牌配置读取当前登录主体，不再固定 platform 主体；品牌 Logo、版权长度校验与原项目一致。
- 普通租户只有具备 `brand_config` 功能和 `brand:edit` 权限时可编辑品牌；版权信息仅平台范围账号可编辑。
- 公开 footer 读取第一个未删除主体的版权文本，与原项目一致。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

状态：

- D 组主体管理与品牌已完成。

## 套餐、功能与配额对比

已对齐项：

- 套餐列表支持分页并排除软删除套餐。
- 套餐复制会同时复制套餐功能和套餐配额，不再只复制套餐主表。
- 套餐删除改为软删除并 tombstone `plan_code`，仍保留被主体订阅引用时不可删除的保护。
- 保存套餐功能时校验功能必须存在且启用，并在事务内替换。
- 租户功能访问检查先校验订阅有效性；无有效订阅时拒绝套餐功能。
- 配额限制读取租户覆盖优先，其次套餐配额，并在无有效订阅时返回不可用。
- 配额使用统计补齐 `max_stores`、`max_departments`、`max_roles`，并排除软删除/停用记录。
- 配额检查支持 `increment`，返回 `allowed`、`remaining_value` 和原因，与原项目“本次增量是否会超额”语义一致。
- 配额列表支持分页；创建配额时默认启用并规范化编码/名称。
- 套餐配额保存改为事务替换，去重并校验配额存在且启用。
- 套餐能力保存会校验功能启用状态，套餐矩阵单元返回与功能相关的配额值。
- 创建用户、角色、组织节点、业务单元时接入套餐配额限制。
- 组织配额按节点类型映射：公司、门店、部门类节点分别使用 `max_companies`、`max_stores`、`max_departments`。
- 套餐能力矩阵会从权限菜单自动同步套餐功能点，排除平台专属入口。
- 能力矩阵按功能父子关系返回树形结构，按钮功能挂在对应菜单功能下。
- 按钮功能没有显式套餐行时继承父菜单功能，兼容旧套餐数据。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

状态：

- E 组套餐、功能与配额已完成。

## 菜单权限与角色对比

已对齐项：

- 菜单包接口改为按当前登录主体读取菜单权限，不再返回固定全局菜单。
- 非平台视角会过滤平台专属菜单，并按当前主体订阅套餐过滤套餐功能菜单。
- 菜单包补齐按钮操作列表、数据权限 ID、套餐功能标识、租户可见/可编辑等元数据。
- 租户菜单覆盖接口改为读取当前主体的真实覆盖配置。
- 保存租户菜单覆盖支持事务内新增/更新，校验菜单存在、租户隔离、平台专属、租户可编辑和重复配置，并写入审计。
- 权限列表改为当前登录主体作用域，支持原项目一致的 `skip/limit/total/items` 分页结构。
- 权限树改为当前主体内按 `sort_order/id` 生成父子树，不再返回扁平列表。
- 权限详情、创建、更新、删除均按当前主体隔离并排除软删除记录。
- 权限输出补齐 `custom_department_ids`、`custom_user_ids` 和 `tenant_visible`。
- 权限创建/更新补齐 `data_scope`、`CUSTOM` 自定义范围、`data_perm_mode` 与菜单类型约束校验。
- 权限删除改为软删除并停用对应套餐功能，创建/更新后同步套餐功能。
- 权限创建、更新、删除补齐审计日志。
- 角色列表改为当前主体作用域、排除软删除、支持 `kw/skip/limit`，非平台主体列表按套餐过滤可见权限 ID。
- 角色详情返回完整权限 ID，避免配置页保存时丢失当前套餐不可见但已绑定的权限。
- 角色创建补齐套餐功能与角色配额校验，并保持创建后权限集合为空的原项目语义。
- 角色更新在事务内替换权限集合与数据权限覆盖，校验权限归属、订阅套餐可分配性和数据权限模式。
- 角色删除改为软删除并 tombstone 角色编码，保留历史关系与审计线索。
- 角色数据权限覆盖补齐组织、部门、用户、业务单元 ID 与业务单元访问模式的回显和保存。
- 数据权限覆盖按菜单 `data_perm_mode` 校验 `NONE/ORG/BU/ORG_BU` 约束。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

状态：

- F1、F2、F3、F4 已完成。
