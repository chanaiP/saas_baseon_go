# Ai经营决策中心今日仿真闭环验收报告

验收时间：2026-05-19 09:52-10:00（Asia/Shanghai）

## 结论

今日仿真数据已进入数据库，并通过真实后端接口跑完整条链路：标准数据入库、异常扫描、AI 分析、整改任务、任务反馈、任务完成、复盘生成、复盘确认、异常关闭。最终 4 条今日异常全部进入闭环状态。

## 数据批次

本次脚本：`scripts/seed-data-center-today-closure.sql`

脚本只做 upsert，不删除历史数据。写入租户 1 的今日/昨日服装行业经营数据：

- Lee：昨日 GMV 20,200，今日 GMV 5,800，用于触发销售下滑。
- Mardi Mercredi：昨日投放成本 8,200、ROI 4.30；今日投放成本 12,850、ROI 1.92，用于触发投流异常。
- Happy Socks：昨日可售天数 24、近 7 日销量 18；今日可售天数 86、近 7 日销量 2，用于触发库存滞销。

扫描接口返回：

- scanned：5
- metrics_generated：25
- generated：4
- duplicated：0

二次补扫返回：

- scanned：5
- metrics_generated：25
- generated：0
- duplicated：4

## 闭环结果

| 异常 ID | 业务对象 | 规则 | 等级 | AI | 任务 | 复盘 | 状态 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 49 | Lee | gmv_drop | critical | success | generated | reviewed | closed |
| 50 | 安德玛 | gmv_drop | critical | success | generated | reviewed | closed |
| 51 | Mardi Flower 卫衣直播加热 | ad_spend_up_roi_down | critical | success | generated | reviewed | closed |
| 52 | Happy Socks 4 双礼盒装 | inventory_enough_sales_down | high | success | generated | reviewed | closed |

整改任务：

- `TASK-20260519-407000`：异常 49，completed，reviewed，progress 100。
- `TASK-20260519-673000`：异常 50，completed，reviewed，progress 100。
- `TASK-20260519-797000`：异常 51，completed，reviewed，progress 100。
- `TASK-20260519-266000`：异常 52，completed，reviewed，progress 100。

整改复盘：

- `REV-20260519-159000`：异常 49，effective。
- `REV-20260519-696000`：异常 50，effective。
- `REV-20260519-338000`：异常 51，effective。
- `REV-20260519-434000`：异常 52，effective。

## AI 调用证据

近 45 分钟内，`data-center/anomaly_analysis` 共有 7 次真实 AI 调用记录：

- provider：`deepseek`
- model：`deepseek-reasoner`
- provider_http_status：200
- status：success

本次闭环调用没有写入或暴露密钥、token、SQL 原文或内部堆栈。

## 发现的问题

1. 每日异常扫描被套餐配额拦截，需要为租户配置 `data_center_daily_scans` 配额覆盖。本次已写入租户 1 验收配额。
2. 复盘确认后，异常主记录不会自动从 `processing` 变为 `closed`，需要额外调用 `/api/data-center/anomalies/:id/close`。本次批量脚本已补调用关闭接口完成闭环。
3. 复盘改善值当前为 `0.00%`，因为自动复盘取的是现有整改前后指标快照，没有引入“整改后新一日指标”。如果要看真实改善幅度，需要在任务完成后再写入 T+1 指标，或让复盘接口支持指定 after_metrics。

## 上线建议

- 将 `data_center_daily_scans` 和 `data_center_daily_ai_analyses` 纳入正式套餐配额，不要依赖手工覆盖。
- 增加服务端自动闭环规则：复盘结论为 `effective` 时，可自动把异常状态更新为 `closed`，并保留审计日志。
- 自动复盘应优先使用整改后的独立指标窗口，避免同日闭环时改善率为 0。
