# Ai经营决策中心真实模型链路验收报告

验收时间：2026-05-19 09:36-09:42（Asia/Shanghai）

## 结论

`Ai经营决策中心` 的经营异常分析链路已打通到 AI 能力中心真实供应商 API。样本异常通过后端接口触发后，AI 能力中心完成真实网关调用，写入 `ai_usage_records`，并在 `data_center_ai_diagnosis_records` 生成结构化诊断；异常主记录 `ai_status` 已回写为 `success`。异常详情接口已返回 `latest_diagnosis`，前端可直接展示最新诊断内容。

## 验收链路

```text
POST /api/data-center/anomalies/35/reanalyze
  -> data-center anomaly_analysis 场景
  -> AI 能力中心租户策略
  -> reasoning-default 基础路由
  -> deepseek / deepseek-reasoner
  -> ai_usage_records
  -> data_center_ai_diagnosis_records
  -> GET /api/data-center/anomalies/35 latest_diagnosis
```

## 样本数据

- 异常 ID：35
- 异常编号：`ANOM-FASH-CARHARTT-STORE-1`
- 品牌：Carhartt WIP
- 场景：北京 SKP 门店外套销售低于目标
- 业务域：销售异常
- 影响金额：21,840 元
- 置信度：82%

## 数据库证据

`ai_usage_records` 最新记录：

- request_id：`dc_anomaly_35_1779154593850476000`
- status：`success`
- provider_http_status：200
- latency_ms：345
- provider：`deepseek`
- model：`deepseek-reasoner`
- route：`reasoning-default`
- usage_amount：1430
- usage_unit：`1K tokens`
- cost_amount：2.86
- billing_amount：5.72

`data_center_ai_diagnosis_records` 最新记录：

- diagnosis_id：45
- status：`success`
- problem_summary：Carhartt WIP 品牌在北京 SKP 门店的外套销售低于目标，导致品牌整体 GMV 环比下滑超过 18%，涉及金额 21,840 元。
- task_suggestion.title：Carhartt WIP 北京 SKP 门店外套销售提升行动
- task_suggestion.owner_role：门店运营经理
- task_suggestion.deadline_days：3

`data_center_anomaly_records` 回写：

- id：35
- ai_status：`success`
- task_status：`none`
- updated_at：2026-05-19 09:36:33

## 接口验收

`GET /api/data-center/anomalies/35` 已返回：

- `anomaly.AIStatus = success`
- `latest_diagnosis.ID = 45`
- `latest_diagnosis.Status = success`
- `latest_diagnosis.ProblemSummary` 为真实模型诊断内容
- `latest_diagnosis.TaskSuggestionJSON` 包含可执行整改任务

## 本次修复

- 修复 AI 用量记录写入时可空 UUID 字段传空字符串导致 PostgreSQL 报错的问题。
- 调整异常分析网关请求为 OpenAI-compatible messages 结构，并约束模型返回业务 JSON。
- 修复诊断解析逻辑，避免把本地兜底诊断误判为模型解析成功。
- 兼容模型返回 `text` 字段中包含 JSON 代码块的情况。
- 异常详情接口新增 `latest_diagnosis`，前端可直接读取最新成功诊断。
- 新增 `scripts/configure-data-center-ai-chain.sql`，用于复现租户策略、AI 配额/限流和每日 AI 分析套餐配额配置。

## 自测结果

- `go test ./internal/apps/data_center/services ./internal/apps/ai_capability_center/services` 通过。
- `go build -o tmp/saas-api ./cmd/api` 通过。
- 真实接口 `POST /api/data-center/anomalies/35/reanalyze` 返回 HTTP 200。
- 详情接口 `GET /api/data-center/anomalies/35` 返回 HTTP 200，并包含 `latest_diagnosis`。

## 上线建议

- 生产环境继续使用 AI 能力中心维护供应商账号与密钥，不在代码或 SQL 脚本中写入密钥。
- 将 `scripts/configure-data-center-ai-chain.sql` 中租户 1 的验收配置改为正式租户策略后再执行到生产。
- 生产环境建议开启正式 `JWT_SECRET`、明确 `CORS_ORIGINS`，并将 `GIN_MODE` 设置为 `release`。
