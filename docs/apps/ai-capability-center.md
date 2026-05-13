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
- 导航口径：上述 7 项全部声明为平台左侧菜单，页面内部不再提供页签式二级导航。

## 权限与应用中心

Manifest 声明平台菜单、配置管理操作和 AI Gateway 调用权限。总览、供应商、模型目录、AI 场景、基础路由、策略中心、系统设置均为独立平台菜单；前端仅根据当前菜单路由渲染对应工作台，不在页面内模拟 tab 切换。菜单仅平台可见，`package_features` 和 `quotas` 保持为空，避免进入租户套餐售卖或租户自助开通。

配置写操作统一写入 SaaS 底座操作日志，`app_code=ai-capability-center`，不创建独立审计表。

## 总览统计口径

总览页的统计数据来自 `ai_usage_records`，独立用量统计页已删除，用量明细仅作为总览、审计和 Gateway 调用链路的数据源保留。

- 今日调用量：当天 `called_at >= 今日 00:00` 的 `calls` 汇总。
- 今日成本：当天 `cost_amount` 汇总，展示为成本价。
- 成功率：当天记录数口径，`status=success` 记录数 / 当天总记录数；无记录时默认 `100.00%`。
- P95 延迟：当天 `latency_ms` 升序后的 95 分位；无记录时为 `0ms`。
- 7 天趋势：从今日往前 6 天到今天，按天汇总 `calls/cost_amount/billing_amount`，无数据日期补 0。
- 模型类型成本结构：近 7 天用量左连接模型目录，按 `model_type` 汇总成本，未匹配模型归类为 `unknown`。
- 租户排行：近 7 天按 `tenant_name` 汇总调用量、成本、收入、毛利、成功率和场景数，按收入倒序取前 8。
- 租户指标：近 7 天服务租户数；今日收入和今日毛利。

## 供应商整体导入

供应商页支持通过 `POST /api/ai-capability-center/providers/import` 一次性导入供应商、接入账号和 API 配置。导入在单个事务内执行，按供应商 `code`、账号 `provider_code + account_name`、API `provider_code + account_name + api_name` 幂等 upsert；任一 provider/account/api 引用不成立时整体回滚。

账号密钥字段仅用于写入，不在前端列表中回显明文或密文；页面仅展示 `key_alias`。

```json
{
  "providers": [
    {
      "name": "OpenAI",
      "code": "openai",
      "type": "public_cloud",
      "base_url": "https://api.openai.com",
      "auth_type": "api_key",
      "priority": 80,
      "region": "US",
      "qps_limit": 100,
      "monthly_budget": 10000,
      "owner": "AI 平台组",
      "accounts": [
        {
          "account_name": "prod",
          "endpoint": "https://api.openai.com",
          "key_alias": "OPENAI_API_KEY",
          "encrypted_api_key": "ciphertext",
          "apis": [
            {
              "api_name": "chat.completions",
              "api_path": "/v1/chat/completions",
              "api_type": "chat",
              "capabilities": ["chat_completion"],
              "auth_type": "api_key",
              "qps_limit": 100,
              "timeout_ms": 30000
            }
          ]
        }
      ]
    }
  ],
  "accounts": [
    {
      "provider_code": "openai",
      "account_name": "sandbox",
      "endpoint": "https://api.openai.com",
      "key_alias": "OPENAI_SANDBOX_KEY",
      "encrypted_api_key": "ciphertext"
    }
  ],
  "apis": [
    {
      "provider_code": "openai",
      "account_name": "sandbox",
      "api_name": "embeddings",
      "api_path": "/v1/embeddings",
      "api_type": "embedding",
      "capabilities": ["embedding"],
      "timeout_ms": 30000
    }
  ]
}
```

删除供应商时会逻辑删除其接入账号和 API；如果供应商已被模型或用量明细引用则阻断。删除接入账号时会逻辑删除其 API；如果账号已被用量明细引用则阻断。

## 模型和价格整体导入

模型目录支持通过 `POST /api/ai-capability-center/models/import` 一次性导入模型、价格策略和分档价格。导入在单个事务内执行，按供应商 `provider_code` 定位供应商，按 `provider_code + model_code` 幂等 upsert 模型，按 `provider_code + model_code + feature_key` 幂等 upsert 价格策略，按 `provider_code + model_code + feature_key + tier_name` 幂等 upsert 分档价格。

任一供应商、模型或能力字典引用不成立时整体回滚，避免模型和价格配置半导入。

```json
{
  "models": [
    {
      "provider_code": "openai",
      "model_code": "gpt-4.1",
      "model_name": "GPT 4.1",
      "model_type": "text",
      "capabilities": ["chat_completion"],
      "context_window": 128000,
      "unit": "tokens",
      "success_rate": 99.5,
      "price_policies": [
        {
          "feature_key": "chat_tokens",
          "feature_name": "对话 Token",
          "model_type": "text",
          "capability_code": "chat_completion",
          "billing_mode": "tiered",
          "billing_unit": "tokens",
          "platform_unit": "tokens",
          "base_cost_price": 0.01,
          "base_sale_price": 0.02,
          "currency": "CNY",
          "tiers": [
            {
              "tier_name": "standard",
              "mode": "sync",
              "cost_price": 0.01,
              "sale_price": 0.02,
              "platform_amount": 0.01,
              "enabled": true,
              "sort_order": 10
            }
          ]
        }
      ]
    }
  ],
  "price_policies": [
    {
      "provider_code": "openai",
      "model_code": "gpt-4.1",
      "feature_key": "chat_tokens",
      "feature_name": "对话 Token",
      "model_type": "text",
      "capability_code": "chat_completion",
      "billing_mode": "tiered",
      "billing_unit": "tokens",
      "platform_unit": "tokens",
      "base_cost_price": 0.01,
      "base_sale_price": 0.02
    }
  ],
  "price_tiers": [
    {
      "provider_code": "openai",
      "model_code": "gpt-4.1",
      "feature_key": "chat_tokens",
      "tier_name": "standard",
      "mode": "sync",
      "cost_price": 0.01,
      "sale_price": 0.02,
      "platform_amount": 0.01,
      "enabled": true
    }
  ]
}
```

## AI 场景批量导入

AI 场景通过 `POST /api/ai-capability-center/scenarios/import` 批量注册或更新。场景唯一键为 `app_code + ai_scenario_code`；导入时会校验能力字典 `capability_code` 存在、默认基础路由 `default_base_route_id` 存在。任一场景引用无效时整批回滚。

```json
{
  "scenarios": [
    {
      "app_code": "product_center",
      "app_name": "商品中心",
      "ai_scenario_code": "product_copy_generate",
      "ai_scenario_name": "商品文案生成",
      "scenario_type": "text",
      "capability_code": "chat_completion",
      "model_type": "text",
      "default_base_route_id": "route-uuid",
      "owner": "商品平台组",
      "description": "用于商品标题、卖点和详情文案生成",
      "version": "v1.0",
      "status": "active"
    }
  ]
}
```

## 基础路由批量导入

基础路由通过 `POST /api/ai-capability-center/routes/import` 一次性导入基础路由和模型池。导入在单个事务内执行，按 `route_code` 幂等 upsert 基础路由，按 `base_route + model + role` 幂等 upsert 模型池节点。基础路由会校验能力字典存在，策略只允许 `fixed/fallback/priority/load_balance/cost_first/quality_first/latency_first/quota_aware/tenant_custom/capability_match`；模型池会校验模型存在，且 `role` 只允许 `primary/fallback/candidate`，`priority/weight/timeout_ms` 必须为正数，`max_retry` 不能为负数。任一引用或数值不合法时整批回滚。

```json
{
  "base_routes": [
    {
      "route_code": "chat-default",
      "route_name": "对话默认路由",
      "capability_code": "chat_completion",
      "model_type": "text",
      "strategy": "fallback",
      "timeout_ms": 30000,
      "max_retry": 2,
      "description": "平台默认对话生成模型池",
      "status": "active",
      "route_models": [
        {
          "provider_code": "openai",
          "model_code": "gpt-4.1",
          "role": "primary",
          "priority": 1,
          "weight": 100,
          "max_retry": 1,
          "timeout_ms": 25000,
          "status": "active"
        }
      ]
    }
  ],
  "route_models": [
    {
      "base_route_code": "chat-default",
      "provider_code": "openai",
      "model_code": "gpt-4.1-mini",
      "role": "fallback",
      "priority": 2,
      "weight": 60,
      "max_retry": 1,
      "timeout_ms": 20000,
      "status": "active"
    }
  ]
}
```

## 租户策略批量导入

策略中心通过 `POST /api/ai-capability-center/tenant-strategies/import` 一次性导入租户策略、配额规则和限流规则。导入在单个事务内执行，按 `policy_name + tenant_scope + app_code + ai_scenario_code` 幂等 upsert 策略，按 `policy + dimension + subject_code + usage_unit + period` 幂等 upsert 配额规则，按 `policy + dimension + subject_code` 幂等 upsert 限流规则。导入会校验 AI 场景已注册、默认基础路由存在、覆盖路由存在；规则维度只允许 `tenant/app/scenario/model/feature_sku/provider_account/user/amount/api`，超限动作只允许 `alert_only/degrade_route/queue/reject/approval`。

```json
{
  "policies": [
    {
      "policy_name": "重点租户商品文案策略",
      "tenant_scope": "include",
      "tenant_ids": ["tenant-a"],
      "app_code": "product_center",
      "app_name": "商品中心",
      "ai_scenario_code": "product_copy_generate",
      "ai_scenario_name": "商品文案生成",
      "default_base_route_id": "route-uuid",
      "override_base_route_id": "fast-route-uuid",
      "status": "active",
      "quota_rules": [
        {
          "dimension": "scenario",
          "subject_code": "product_copy_generate",
          "usage_unit": "tokens",
          "period": "day",
          "quota_limit": 100000,
          "warning_threshold": 80,
          "over_limit_action": "alert_only",
          "status": "active"
        }
      ],
      "rate_limit_rules": [
        {
          "dimension": "user",
          "subject_code": "user-a",
          "qps": 20,
          "concurrency": 5,
          "over_limit_action": "queue",
          "status": "active"
        }
      ]
    }
  ]
}
```
