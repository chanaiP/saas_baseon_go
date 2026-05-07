# Implementation TODO

> 目标：在不改变既有产品需求、菜单、权限语义和前端交互的前提下，把 Go 重建版从“功能可用”推进到“生产级可维护”。

## P0

- [x] 建立持久 TODO 清单。
- [x] 角色管理下沉到 `application/repository`，补单元测试，作为 DDD 拆分样板。
- [x] 建立数据权限边界清单，并补首批后端拒绝/隔离测试。
- [x] 建立审计必记事件清单，并补关键写操作审计落库。
- [x] 补删除/归档引用约束，避免误删被引用基础资料。

## P1

- [x] OpenAPI 从路径级摘要升级为 request/response schema。
- [x] 完善生产 migration 链路，降低对 AutoMigrate 的依赖。
- [x] 补业务级 E2E/接口回归，覆盖 CRUD、权限拒绝、套餐配额、审计落库。

## P2

- [x] 清理文档中的待定标记，让验收口径和当前实现一致。
- [x] 清理开发者展示组件中的 mock/console 噪音，保留明确标注的 showcase 示例。
