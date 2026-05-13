# AI 能力中心

AI 能力中心是合并部署应用，`app_code=ai-capability-center`。

本应用统一治理 AI 供应商、接入账号、API 配置、AI 能力字典、模型目录、模型价格策略、AI 场景、基础路由、租户策略和网关设置。用量趋势、成本结构和租户排行统一收敛在总览页展示。

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
- 菜单：总览、供应商、模型目录、AI 场景、基础路由、策略中心、系统设置

## 权限与套餐

Manifest 声明菜单、配置管理操作、AI Gateway 调用权限、套餐功能点和月度配额。

配置写操作统一写入 SaaS 底座操作日志，`app_code=ai-capability-center`，不创建独立审计表。
