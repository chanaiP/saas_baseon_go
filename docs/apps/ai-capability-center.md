# AI 能力中心

AI 能力中心是合并部署应用，`app_code=ai-capability-center`。它属于业务中台的平台能力，仅平台侧可见和可管理，不作为租户后台菜单，也不进入租户套餐。

本应用统一治理 AI 供应商、接入账号、API 配置、AI 能力字典、模型目录、模型价格策略、AI 场景、基础路由、租户策略和网关设置。用量趋势、成本结构和租户排行统一收敛在总览页展示。

租户业务系统通过 AI Gateway 获得 AI 路由、策略覆盖、配额限流和用量沉淀能力，但租户不在租户后台维护本应用菜单或配置。

业务侧调用 AI Gateway 时只传：

```json
{
  "tenant_id": "tenant-hd-flagship",
  "app_code": "product_center",
  "ai_scenario_code": "product_image_generate",
  "input": {}
}
```

平台侧根据已注册 AI 场景、默认基础路由、租户策略、配额限流和价格策略完成模型路由与用量沉淀。

## 应用接入

- 后端目录：`internal/apps/ai_capability_center`
- 前端目录：`frontend/src/apps/ai-capability-center`
- Manifest：`internal/apps/ai_capability_center/app.manifest.yaml`
- 部署方式：`MERGED`
- 可见范围：`PLATFORM_ONLY`
- 售卖策略：`NON_SELLABLE`
- 套餐策略：`NON_SELLABLE`
- 菜单：总览、供应商、模型目录、AI 场景、基础路由、策略中心、系统设置

## 权限与应用中心

Manifest 声明平台菜单、配置管理操作和 AI Gateway 调用权限。菜单仅平台可见，`package_features` 和 `quotas` 保持为空，避免进入租户套餐售卖或租户自助开通。

配置写操作统一写入 SaaS 底座操作日志，`app_code=ai-capability-center`，不创建独立审计表。
