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
- Profile 权限码改为菜单权限与按钮权限一起返回，排除软删除角色/权限。
- 权限码固定补齐 `/home` 与 `home:view`，取消空权限时返回开发权限全集的临时逻辑。
- 非平台主体的权限码和接口门禁都会按当前套餐过滤套餐功能权限。
- 门禁权限码复用同一套当前用户权限码计算，避免 Profile 可见与接口可访问不一致。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

状态：

- F 组菜单权限与角色已完成。

## 组织架构对比

已对齐项：

- 组织树改为按 `parent_id` 返回层级结构，不再返回扁平列表。
- 组织树排除软删除节点，并按节点 ID 保持原项目排序语义。
- 组织树补齐公司节点自身 `company_id`、部门类节点门店归属 `store_id`、公司/门店/集团的 `company_type` 回显。
- 组织树按当前用户 `data:org` 数据权限过滤，支持 `ALL/SELF/ORG/ORG_SUB/CUSTOM`，并保留通向授权子节点的中间节点。
- 组织创建接入 `org_manage` 功能门禁和公司/门店/部门类配额。
- 公司创建校验上级公司/上级部门互斥和父级存在性；部门创建校验所属公司、门店与上级部门归属一致；门店创建校验上级公司/部门/门店。
- 统一 org-node 创建支持任意节点类型，非公司节点按父链解析 `company_id`，公司节点清空 `company_id`。
- 组织更新支持名称、编码、状态、节点类型、公司类型、父级调整，并阻止父级指向自身或子孙节点。
- 父级或节点类型变化后会级联刷新当前子树 `company_id`。
- 组织删除改为软删除并停用节点，删除前校验下级组织、用户主归属、用户任职关联、业务单元组织映射。
- 组织创建、更新、删除均写入审计；Go 版无独立组织树缓存层，因此无需额外缓存失效动作。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

状态：

- G 组组织架构已完成。

## 岗位管理对比

已对齐项：

- 岗位类型列表改为当前主体作用域，排除软删除，支持 `skip/limit/keyword` 与 `position_count`。
- 岗位类型创建、更新、删除均按当前主体隔离，并补齐审计。
- 岗位类型删除改为软删除并 tombstone 编码，删除前校验类型下仍有岗位。
- 岗位列表改为当前主体作用域，排除软删除，支持 `skip/limit/keyword/position_type_id`。
- 岗位创建/更新会校验岗位类型属于当前主体。
- 岗位删除改为软删除并 tombstone 编码，删除前校验用户岗位关联。
- 岗位创建、更新、删除均补齐审计。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

状态：

- H 组岗位管理已完成。

## 业务单元对比

已对齐项：

- 业务单元列表改为当前主体作用域，排除软删除，支持 `skip/limit/keyword/status` 并按 ID 倒序返回。
- 业务单元树/平铺接口排除软删除数据。
- 创建业务单元接入 `business_unit_manage` 功能门禁，启用状态下校验 `max_business_units` 配额。
- 创建/更新业务单元会校验业务单元类型非空，维护 `billing_enabled` 与 `statistic_enabled` 系统标志。
- 创建/更新业务单元支持批量 PRIMARY 组织映射，并校验组织节点存在和同一组织不能重复挂到其他启用业务单元。
- 业务单元组织映射列表只返回启用映射。
- 新增组织映射支持 `scope_type`、`priority`、组织类型自动识别和租户隔离；删除映射改为停用。
- 业务单元删除改为软删除、停用、tombstone 编码，并保留数据权限范围引用保护和审计。
- 业务单元列表与树接入 `data:business_unit` 数据权限范围，支持 `SPECIFIED_BU` 与 `CURRENT_ORG_BU` 过滤。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

状态：

- I 组业务单元已完成。

## 数据字典与系统参数对比

已对齐项：

- 字典类型列表改为当前主体作用域，排除软删除，支持 `skip/limit/keyword/platform_only`，非平台管理员过滤平台专属字典。
- 字典类型创建、更新、删除按当前主体隔离，删除改为软删除并 tombstone 编码，删除前校验字典项引用。
- 字典类型创建、更新、删除补齐审计。
- 字典项列表按当前主体和字典类型分页返回，并合并租户覆盖值。
- 按字典编码读取字典项时会合并租户覆盖、过滤停用项并按排序返回。
- 非平台管理员编辑字典项改为写入租户覆盖；不允许租户覆盖的字典会拒绝编辑。
- 平台管理员编辑字典项更新默认值；删除字典项改为软删除并停用。
- 恢复字典项默认值会删除租户覆盖行并返回默认值。
- 系统参数列表改为当前主体作用域，排除软删除，支持 `skip/limit/keyword`，非平台管理员过滤平台专属参数。
- 系统参数批量读取支持 `keys` 逗号分隔，只返回当前主体可见参数并合并租户覆盖值；缺失 key 返回 `null`。
- 系统参数创建、更新、删除按当前主体隔离；删除改为软删除并 tombstone key。
- 非平台管理员更新系统参数改为写入租户覆盖；平台专属或不可租户编辑参数会拒绝覆盖。
- 平台管理员更新系统参数会更新默认值、备注、类型、可编辑范围和平台专属标识。
- 恢复系统参数默认值会删除租户覆盖行并返回默认值。
- 系统参数创建、更新、删除、恢复默认补齐审计。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

状态：

- J 组数据字典与系统参数已完成。

## 日志对比

已对齐项：

- 登录日志列表改为分页返回，支持 `success/account/ip/date_from/date_to` 筛选。
- 登录日志按当前用户范围隔离：普通主体只能看本主体，平台范围可跨主体并支持 `tenant_name` 模糊筛选。
- 登录日志输出补齐主体名称、用户 ID、账号、成功状态、失败/提示信息、IP 和创建时间。
- 操作日志列表改为分页返回，支持 `module/keyword/account/ip/date_from/date_to` 筛选。
- 操作日志按当前用户范围隔离：普通主体只能看本主体，平台范围可跨主体并支持 `tenant_name` 模糊筛选。
- 操作日志输出补齐主体名称、用户 ID、模块、动作、摘要、明细、IP 和创建时间。
- 日志查询接口保持只读，不产生新的审计日志。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

状态：

- K 组日志已完成。

## 文件与批量导入导出对比

已对齐项：

- 文件上传接入 `file_manage` 套餐功能开关。
- 文件上传保留 50 MB 全局上限，并叠加 `max_file_size_mb` 套餐配额，0 表示当前套餐不支持上传。
- 文件上传接入 `max_storage_gb` 租户容量配额，按租户上传目录统计已有文件大小，0 表示当前套餐不支持文件存储。
- 原始文件名改为安全 basename、空名兜底 `unknown`、最长 255 字节；扩展名只允许字母数字且最长 16 字符，否则保存为 `.bin`。
- 文件仍按 `uploads/{tenant_id}/YYYY/MM/DD` 保存，下载和删除只在当前主体目录查找，并跳过 `.trash`。
- 非法 `file_id` 返回 400，跨主体或不存在文件返回 404。
- 文件删除改为移动到 `uploads/{tenant_id}/.trash/YYYY/MM/DD/HHMMSS_filename`，审计中记录回收路径。
- 用户 CSV 导出接入 `export_data` 套餐功能开关，并消耗 `daily_export_times` 日配额。
- 用户 CSV 导出按 `/users` 数据权限过滤，支持 `ALL/SELF/ORG/ORG_SUB/CUSTOM` 对公司、部门、用户范围的限制。
- 用户 CSV 导出只导出未软删除用户和未软删除组织名称，字段保持 `employee_no,name,phone,email,company_name,department_name,status`，并保留 CSV 公式注入防护。
- 用户 CSV 导入接入 `import_data` 套餐功能开关，并消耗 `daily_import_times` 日配额。
- 用户 CSV 导入支持 UTF-8 BOM 识别，保持 5 MB 文件大小限制，校验工号/姓名必填和手机号格式。
- 用户 CSV 导入按公司名称、部门名称解析组织归属，重复工号只统计未软删除用户，活跃用户创建前校验 `max_users` 配额。
- 用户 CSV 导入审计只记录前 10 条错误，响应返回前 20 条错误，与原项目错误汇总边界一致。
- 公司/部门 CSV 导出接入 `export_data` 套餐功能开关和 `daily_export_times` 日配额，只导出未软删除组织节点。
- 公司 CSV 字段保持 `name,code,company_type,parent_name,status`；部门 CSV 字段保持 `name,code,company_name,parent_name,status`，部门父级名称只取部门父级。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

状态：

- L 组文件与批量导入导出已完成。

## 监控对比

已对齐项：

- 健康检查返回 `mysql/redis` 字段，数据库状态改为真实连接 ping。
- 服务器信息返回 `python_version/pid/cpu_percent/memory_mb/note` 字段，`python_version` 在 Go 版中承载 Go runtime 版本以保持前端契约。
- 服务概览返回 `mysql/redis/python_version/pid/cpu_percent/memory_mb/note` 字段，数据库和 Redis 状态都使用实时检查结果。
- Redis key scan 参数规范化为 `limit` 1-200、默认 pattern 为 `*`，拒绝过长或包含换行/空字符的 pattern。
- Redis key scan 会循环扫描直到达到 limit 或游标归零，返回字段保持 `key/ttl`，Redis 错误返回 503。
- 定时任务展示恢复为原项目 3 条 Redis 相关内置行为说明：会话校验、验证码存储、登录失败计数。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

状态：

- M 组监控已完成。

## 数据库与迁移对比

已对齐项：

- AppUser 模型补齐 `tenant_id + employee_no`、`tenant_id + phone` 唯一约束，并把密码、姓名、邮箱长度调整为原模型口径。
- Role、UserRole、RolePermission、AppUserDepartment、AppUserPosition、UserPreference 模型补齐原模型唯一约束。
- BusinessUnit 统计开关默认值调整为启用，BusinessUnitOrgMap 和 BusinessUnitScope 补齐复合唯一约束。
- TenantMenuOverride、TenantDictItemOverride、TenantParamValue 模型补齐租户维度覆盖唯一约束。
- Permission 字段长度按原模型调整：名称 200、路径 500、功能编码 100、租户编辑范围 100。
- DictType 租户可编辑默认值调整为启用，与原模型一致。

验证：

- `docker run --rm -e GOPROXY=https://goproxy.cn,direct -v "$PWD":/src -w /src golang:1.23-alpine sh -c 'gofmt -w ./cmd ./internal && go test ./...'` 通过。

状态：

- N1 Go 模型字段、索引、唯一约束、软删除字段已完成。
