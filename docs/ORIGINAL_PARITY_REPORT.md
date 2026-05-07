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

