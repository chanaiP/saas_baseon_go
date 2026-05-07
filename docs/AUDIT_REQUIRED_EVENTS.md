# 审计必记事件清单

## 当前落库规则

- 写操作成功后记录 `audit_log`，失败请求不写业务审计。
- 审计主体使用当前登录用户，记录模块、动作、摘要、详情、IP、用户与租户。
- 详情字段只记录业务标识和变更摘要，不记录明文密码、token、文件内容等敏感数据。

## 已覆盖事件

| 模块 | 动作 |
| --- | --- |
| tenant | create |
| tenant | create_with_package |
| tenant | update |
| tenant | status_update |
| tenant | delete |
| user | create |
| user | update |
| user | password_reset |
| user | delete |
| role | create |
| role | update |
| role | delete |
| organization | create |
| organization | update |
| organization | delete |
| business_unit | create |
| business_unit | update |
| business_unit | delete |
| business_unit | org_mapping_create |
| business_unit | org_mapping_delete |

## 已有历史覆盖

- profile/update：个人资料更新。
- tenant_branding/update：主体品牌更新。
- file/upload、file/delete：文件上传与删除。
- user/batch_import：用户批量导入。
- bootstrap/seed：初始化基础数据。

## 后续扩展

- 套餐、功能、配额、字典、系统参数、岗位等写操作继续按本清单模式补齐。
- 业务级 E2E 需要增加“写操作后能在操作日志查询到记录”的回归断言。

