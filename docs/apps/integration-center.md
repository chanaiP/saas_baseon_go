# 第三方集成中心

第三方集成中心是合并部署内置连接器应用，`app_code=integration-center`。它用于平台侧统一治理接入平台、服务商应用、应用能力连接、租户连接实例、同步任务、配额限流、异常监控和调用日志。

## 接入方式

- 后端目录：`internal/apps/integration_center`
- 前端目录：`frontend/src/apps/integration-center`
- Manifest：`internal/apps/integration_center/app.manifest.yaml`
- 应用类型：`CONNECTOR_APP`
- 部署方式：`MERGED`
- 客户端：`PC_WEB`、`API_ONLY`

## 后台菜单

Manifest 声明平台后台管理菜单：

- 总览：`/integration-center`
- 接入平台：`/integration-center/platforms`
- 集成工作台：`/integration-center/workspace`
- 租户连接：`/integration-center/tenant-connections`
- 同步监控：`/integration-center/sync-monitor`
- 配额与限流：`/integration-center/quota`
- 异常监控：`/integration-center/alerts`
- 调用日志：`/integration-center/logs`

上述菜单是平台治理入口，`platform_only=true`，不进入租户套餐中心。

## 套餐与配额

套餐中心功能点由 Manifest 的 `package_features` 声明：

- `integration_tenant_authorization`：第三方集成授权连接
- `integration_data_sync`：第三方数据同步

配额由 Manifest 的 `quotas` 声明：

- `integration_connection_count`：第三方连接实例数
- `integration_api_calls_daily`：第三方接口日调用量
- `integration_sync_records_daily`：第三方同步日记录数

平台专属治理菜单不得作为租户套餐功能点。租户可购买能力只来自 Manifest 的 `package_features`。

## 权限与 API

受保护 API 均绑定 Manifest 权限码：

- `GET /api/integration-center/overview` -> `/integration-center`
- `GET /api/integration-center/platforms` -> `/integration-center/platforms`
- `GET /api/integration-center/workspace` -> `/integration-center/workspace`
- `GET /api/integration-center/tenant-connections` -> `/integration-center/tenant-connections`
- `GET /api/integration-center/sync-monitor` -> `/integration-center/sync-monitor`
- `GET /api/integration-center/quota` -> `/integration-center/quota`
- `GET /api/integration-center/alerts` -> `/integration-center/alerts`
- `GET /api/integration-center/logs` -> `/integration-center/logs`

写操作权限预留：

- `integration_center:platform_manage`
- `integration_center:app_manage`
- `integration_center:connection_manage`
- `integration_center:quota_manage`

## 初始化

生产初始化通过当前 schema baseline 和增量迁移执行。`000052_integration_center_manifest_seed` 将 `integration-center` 从规划状态升级为上线状态，并按 Manifest 基线补齐应用中心资产、菜单权限、套餐功能点和配额模板；`000053_integration_center_core_schema` 提供平台、服务商应用、租户连接、能力、同步、配额、异常和调用日志等业务表。
