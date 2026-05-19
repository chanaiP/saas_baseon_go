-- DeepSeek official CNY token prices are billed per 1M tokens.
-- deepseek-chat: cache hit input 0.5, cache miss input 2, output 8 CNY / 1M tokens.
-- deepseek-reasoner: cache hit input 1, cache miss input 4, output 16 CNY / 1M tokens.
--
-- The price-policy table has one base price, so store the cache-miss input price
-- as the default reference and keep the existing 2x tenant sale markup. Runtime
-- billing uses the provider-returned token breakdown for exact DeepSeek costs.

UPDATE ai_model_price_policies p
SET
  billing_unit = '1M tokens',
  platform_unit = '1M tokens',
  base_cost_price = 2.000000,
  base_sale_price = 4.000000,
  base_platform_amount = 1.000000,
  updated_at = NOW()
FROM ai_models m
WHERE p.model_id = m.id
  AND m.model_code = 'deepseek-chat'
  AND p.capability_code = 'chat_completion'
  AND p.deleted_at IS NULL;

UPDATE ai_model_price_policies p
SET
  billing_unit = '1M tokens',
  platform_unit = '1M tokens',
  base_cost_price = 4.000000,
  base_sale_price = 8.000000,
  base_platform_amount = 1.000000,
  updated_at = NOW()
FROM ai_models m
WHERE p.model_id = m.id
  AND m.model_code = 'deepseek-reasoner'
  AND p.capability_code = 'reasoning'
  AND p.deleted_at IS NULL;

UPDATE ai_model_price_tiers t
SET
  platform_amount = 1.000000,
  updated_at = NOW()
FROM ai_model_price_policies p
JOIN ai_models m ON m.id = p.model_id
WHERE t.price_policy_id = p.id
  AND m.model_code IN ('deepseek-chat', 'deepseek-reasoner')
  AND p.deleted_at IS NULL
  AND t.deleted_at IS NULL;

WITH recalculated AS (
  SELECT
    r.id,
    p.platform_unit,
    ROUND((r.usage_amount / 1000000)::numeric, 6) AS platform_amount,
    ROUND(((r.usage_amount / 1000000) * p.base_cost_price)::numeric, 6) AS cost_amount,
    ROUND(((r.usage_amount / 1000000) * p.base_sale_price)::numeric, 6) AS billing_amount
  FROM ai_usage_records r
  JOIN ai_model_price_policies p ON p.id = r.price_policy_id
  JOIN ai_models m ON m.id = p.model_id
  WHERE m.model_code IN ('deepseek-chat', 'deepseek-reasoner')
)
UPDATE ai_usage_records r
SET
  usage_unit = 'tokens',
  platform_unit = recalculated.platform_unit,
  platform_amount = recalculated.platform_amount,
  cost_amount = recalculated.cost_amount,
  billing_amount = recalculated.billing_amount
FROM recalculated
WHERE r.id = recalculated.id;
