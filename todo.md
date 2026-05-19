# 业务单元 v3 重构 TODO

> 完成标准：本文件所有任务全部打勾，且后端测试、前端构建、浏览器验证、`git diff --check` 全部通过。

## 1. 需求定版

- [x] 确认业务单元模块定位为基础业务对象登记中心。
- [x] 确认页面结构为“类型 -> 分组 -> 列表”。
- [x] 确认只使用一个业务单元数据字典。
- [x] 确认业务单元字典一级项为类型，二级项为分组。
- [x] 确认真实业务单元由用户登记，不是字典项。
- [x] 确认新增入口改为“新增业务单元”。
- [x] 确认父子关系用于层级归属和 BI 汇总。
- [x] 确认关联关系用于表达业务单元之间的横向依赖。
- [x] 确认自定义属性使用模板加载，填写后转换为 JSON 保存。
- [x] 确认属性模板匹配优先级为“类型 + 分组”优先，“类型 + 空分组”兜底。
- [x] 确认负责组织和负责人非必填，并复用 `OrgUserPicker`。

## 2. 文档

- [x] 更新 `docs/req_design/业务单元.md`。
- [x] 更新 `docs/tech_design/业务单元.md`。
- [x] 更新 `docs/sql_design/05-业务单元-数据库设计.md`。
- [x] 更新 `docs/test_cases/05-业务单元-测试用例.md`。
- [x] 根据最终实现补充 API 文档和字段说明。
- [x] 根据最终实现补充交接说明。

## 3. 数据字典与数据库

- [x] 清理旧版业务单元字典，仅保留运行态 `business_unit`。
- [x] 将旧 `business_unit_tree` 数据迁移到 `business_unit` 后软归档旧字典。
- [x] 将旧 `business_unit_type`、`base_business.unit_scenario`、`base_business.unit_form` 软归档，避免新旧字典混用。
- [x] 保持数据字典表结构不变，继续使用 `dict_item.parent_id`。
- [x] 将一级字典项作为 `unit_type_code` 可选范围。
- [x] 将二级字典项作为 `unit_group_code` 可选范围。
- [x] 调整或新增业务单元主表字段：`unit_type_code`、`unit_type_name`、`unit_group_code`、`unit_group_name`、`unit_name`、`unit_code`、`parent_id`、`attr_template_id`、`attrs`、`remark`。
- [x] 新增或调整业务单元关系表，支持 `source_unit_id`、`target_unit_id`、`relation_type_code`。
- [x] 新增或调整业务单元责任方表，支持负责组织和负责人。
- [x] 新增属性模板表。
- [x] 新增属性模板字段表。
- [x] 新增属性模板唯一约束：同租户、同模板名称、同适用类型、同适用分组不可重复。
- [x] 更新 GORM model。
- [x] 更新 `current_schema.sql`。
- [x] 新增幂等迁移 SQL。
- [x] 确保迁移不破坏历史业务单元、业务资源、角色授权、套餐和租户覆盖配置。

## 4. 后端接口

- [x] 新增或调整 `GET /api/base/business-units/dictionary/tree`，读取业务单元字典。
- [x] 新增或调整 `GET /api/base/business-units/summary`，按类型和分组聚合业务单元数量。
- [x] 新增或调整 `GET /api/base/business-units`，按类型、分组、关键字、状态分页查询。
- [x] 新增或调整 `POST /api/base/business-units`，新增业务单元。
- [x] 新增或调整 `PUT /api/base/business-units/:id`，编辑业务单元。
- [x] 新增或调整 `DELETE /api/base/business-units/:id`，逻辑归档业务单元。
- [x] 新增或调整 `GET /api/base/business-units/:id/actors`，查询负责组织和负责人。
- [x] 新增或调整 `PUT /api/base/business-units/:id/actors`，保存负责组织和负责人。
- [x] 新增或调整 `GET /api/base/business-units/:id/relations`，查询关联业务单元。
- [x] 新增或调整 `PUT /api/base/business-units/:id/relations`，保存关联业务单元。
- [x] 新增 `GET /api/base/business-unit-attr-templates/match`，按类型和分组匹配模板。
- [x] 新增属性模板 CRUD 接口。
- [x] 后端校验类型和分组必须来自同一个业务单元字典且父子匹配。
- [x] 后端校验父级业务单元不能跨租户、不能自引用、不能形成循环。
- [x] 后端校验关联业务单元不能跨租户、不能关联自身。
- [x] 后端校验负责组织和负责人不能跨租户。
- [x] 后端校验 attrs 必须为合法 JSON 对象，并按模板校验必填字段。
- [x] 后端校验属性模板同一适用范围下名称唯一。

## 5. 菜单、权限与套餐

- [x] 保持业务单元页面挂在系统管理下的 `/business-units` 菜单。
- [x] 将新增按钮权限改为 `business_unit:create`。
- [x] 将编辑权限改为 `business_unit:edit`。
- [x] 将归档权限改为 `business_unit:delete`。
- [x] 新增或调整 `business_unit:relation_manage`。
- [x] 新增或调整 `business_unit:actor_manage`。
- [x] 新增或调整 `business_unit:attr_template_manage`。
- [x] 更新后端接口权限矩阵。
- [x] 更新前端按钮权限判断。
- [x] 更新系统管理 Manifest。
- [x] 更新套餐功能点，保持应用 -> 目录 -> 菜单 -> 操作树。
- [x] 验证 Manifest 装载不覆盖人工套餐配置、租户菜单覆盖和角色授权。

## 6. 前端页面

- [x] 将页面主对象从“业务资源”调整为“业务单元”。
- [x] 左侧展示有业务单元数据的类型。
- [x] 右上展示当前类型下有业务单元数据的分组统计。
- [x] 右下展示当前类型 + 分组下的业务单元列表。
- [x] 新增按钮文案改为“新增业务单元”。
- [x] 新增表单包含类型、分组、名称、编码、父级、状态、备注。
- [x] 新增表单按类型和分组自动匹配属性模板。
- [x] 支持手动选择属性模板。
- [x] 根据模板字段渲染 text、number、date、select、textarea、switch 控件。
- [x] 保存时将模板字段值转换成 `attrs` JSON。
- [x] 编辑时根据 `attrs` 回填模板字段。
- [x] 负责组织和负责人区域复用 `OrgUserPicker`。
- [x] 将 `OrgUserPicker` 输出的 `c_`、`d_`、`u_` key 转换为 actor 记录。
- [x] 负责组织和负责人允许为空。
- [x] 增加关联业务单元维护区域。
- [x] 增加属性模板页内维护入口和抽屉，不新增左侧菜单。
- [x] 增加业务单元详情侧边抽屉，展示基础信息、责任方、关联关系和自定义属性。
- [x] 右上分组卡片固定尺寸，文字适配，不撑满屏幕。
- [x] 页面风格保持 NeuroAgent 深色风格。
- [x] 页面短内容不出现多余滚动条。

## 7. 后端测试

- [x] 字典父子校验：类型和分组必须匹配。
- [x] Summary 聚合：按类型和分组统计业务单元数量。
- [x] 新增业务单元：名称、编码、类型、分组保存正确。
- [x] 编码唯一：同租户有效业务单元编码唯一。
- [x] 父子关系：禁止自引用、禁止循环、禁止跨租户。
- [x] 关联关系：禁止自关联、禁止跨租户。
- [x] 属性模板匹配：类型+分组优先，类型+空分组兜底。
- [x] 属性模板唯一性：重复名称 + 类型 + 分组创建被拒绝。
- [x] attrs 校验：非法 JSON 或必填缺失被拒绝。
- [x] 责任方：组织和人员保存、回显、跨租户保护。
- [x] 权限矩阵：无权限接口返回 403。
- [x] Manifest、菜单、套餐功能点存在且不覆盖人工配置。

## 8. 前端验证

- [x] 运行 `cd frontend && npm run build`。
- [x] 浏览器验证 `http://127.0.0.1:5177/business-units` 页面能访问。
- [x] 验证左侧类型、右上分组、右下列表联动。
- [x] 验证抖音、京东等分组能显示业务单元数量。
- [x] 验证新增业务单元可保存。
- [x] 验证属性模板加载、填写、保存、编辑回显。
- [x] 验证属性模板维护抽屉入口可见并可打开。
- [x] 验证业务单元详情抽屉可打开并回显当前业务单元。
- [x] 验证负责组织和负责人可为空。
- [x] 验证 `OrgUserPicker` 可选择组织和人员并保存回显。
- [x] 验证关联业务单元可保存和回显。
- [x] 验证无权限时按钮隐藏、接口拒绝。

## 9. 全量验证

- [x] 运行 `go test ./internal/interfaces/http/handlers`。
- [x] 运行 `go test ./...`。
- [x] 运行 `cd frontend && npm run build`。
- [x] 运行 `git diff --check`。
- [x] 浏览器完成主流程验收。
