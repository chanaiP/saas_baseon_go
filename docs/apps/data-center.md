# Ai经营决策中心

## 应用定位

Ai经营决策中心是合并部署的内置业务中台应用，`app_code` 为 `data-center`，后端目录为 `internal/apps/data_center`，前端目录为 `frontend/src/apps/data-center`。

应用链路：

```text
原始数据接入 -> 标准数据 -> 指标结果 -> 规则异常 -> AI 分析 -> 整改任务 -> 整改复盘
```

## 菜单

- 经营看板
- 数据总览
- 原始数据
- 标准数据
- 指标中心
- 异常分析
- 异常规则
- 整改任务
- 整改复盘

## 权限

Manifest 文件：`internal/apps/data_center/app.manifest.yaml`。

权限资源由 Manifest 声明，API 前缀为 `/api/data-center`。读接口绑定对应菜单权限，写接口绑定操作权限：

- `data-center:metric:manage`
- `data-center:rule:manage`
- `data-center:anomaly:scan`
- `data-center:ai:analyze`
- `data-center:task:generate`
- `data-center:task:flow`
- `data-center:review:confirm`
- `data-center:raw:retry`
- `data-center:data:export`

## 套餐与配额

套餐功能点由 Manifest 的 `package_features` 声明，配额由 `quotas` 声明。当前第一版包含：

- 原始数据批次保留量
- 每日异常扫描次数
- 每日 AI 分析次数
- 活跃异常规则数量

## 数据库

生产初始化使用：

```bash
go run ./cmd/migrate
```

本应用的版本化迁移为：

- `internal/infrastructure/persistence/postgres/migrations/000074_data_center_schema.up.sql`
- `internal/infrastructure/persistence/postgres/migrations/000074_data_center_schema.down.sql`

Schema baseline 已同步到：

- `internal/infrastructure/persistence/postgres/schema/current_schema.sql`

## API

主要 API 分组：

- `/api/data-center/dashboard/*`
- `/api/data-center/overview/*`
- `/api/data-center/raw/*`
- `/api/data-center/standard/*`
- `/api/data-center/metrics/*`
- `/api/data-center/anomaly-rules/*`
- `/api/data-center/anomalies/*`
- `/api/data-center/tasks/*`
- `/api/data-center/reviews/*`

所有业务查询必须从登录态读取 `tenant_id`，禁止信任前端传入租户参数。列表接口统一返回 `items,total,skip,limit`。

## 数据导入

第一版导入入口：

```bash
POST /api/data-center/raw/batches
```

请求体示例：

```json
{
  "batch_code": "BATCH-SALES-20260518",
  "data_type": "sales",
  "records": [
    {
      "order_code": "ORDER-001",
      "order_time": "2026-05-18T01:00:00+08:00",
      "sales_amount": 1200,
      "paid_amount": 1100,
      "brand_code": "brand-a"
    }
  ]
}
```

支持 `sales`、`ad`、`inventory`、`refund`、`product`、`store-sales`。导入会创建原始批次并同步标准化入库；失败行写入 `data_center_raw_data_errors`，批次状态根据成功/失败数量变为 `success`、`warning` 或 `failed`。

## 当前 MVP 闭环

已实现可由 API 手动触发的异常扫描：

```bash
POST /api/data-center/anomalies/scan
```

扫描会基于标准表计算指标结果，并完成 3 条 MVP 异常识别：

- GMV 下滑
- 投流增加但 ROI 下降
- 库存充足但销量下降

异常记录会写入证据 JSON，并按 `tenant_id + rule_code + object_type + object_code + stat_date` 去重。

## 验证

后端重点测试：

```bash
go test ./internal/apps/data_center/... ./internal/bootstrap ./internal/interfaces/http/handlers
```

全量测试：

```bash
go test ./...
```

前端构建：

```bash
cd frontend && npm run build
```
