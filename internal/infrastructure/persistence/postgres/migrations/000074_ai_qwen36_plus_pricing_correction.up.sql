-- Correct AI capability center model pricing units and recalculate generated usage records.
--
-- qwen3.6-plus text pricing is aligned to Alibaba Cloud Bailian's public mainland
-- tier for 0 < input tokens <= 256K:
-- input: 2 CNY / 1M tokens, output/thinking: 12 CNY / 1M tokens.
--
-- Token prices are stored in production-readable billing units (tokens, 1K tokens,
-- 1M tokens). The gateway converts raw token usage into the configured billing unit
-- before multiplying by cost_price/sale_price.

UPDATE ai_model_price_policies p
SET
  billing_unit = '1M tokens',
  platform_unit = '1M tokens',
  base_cost_price = 2.000000,
  base_sale_price = 12.000000,
  base_platform_amount = 1.000000,
  updated_at = NOW()
FROM ai_models m
WHERE p.model_id = m.id
  AND m.model_code = 'qwen3.6-plus'
  AND p.capability_code = 'chat_completion'
  AND p.deleted_at IS NULL;

UPDATE ai_model_price_tiers t
SET
  tier_name = '官方文本价格 0-256K',
  cost_price = 2.000000,
  sale_price = 12.000000,
  platform_amount = 1.000000,
  updated_at = NOW()
FROM ai_model_price_policies p
JOIN ai_models m ON m.id = p.model_id
WHERE t.price_policy_id = p.id
  AND m.model_code = 'qwen3.6-plus'
  AND p.capability_code = 'chat_completion'
  AND p.deleted_at IS NULL
  AND t.deleted_at IS NULL;

-- kimi-k2.6 was imported with token usage records but a 1K-token display unit.
-- Keep the configured production prices and make the billing unit explicit so the
-- gateway converts raw tokens to 1K-token billing quantity.
UPDATE ai_model_price_policies p
SET
  billing_unit = '1K tokens',
  platform_unit = '1K tokens',
  base_platform_amount = 1.000000,
  updated_at = NOW()
FROM ai_models m
WHERE p.model_id = m.id
  AND m.model_code = 'kimi-k2.6'
  AND p.capability_code = 'chat_completion'
  AND p.deleted_at IS NULL;

UPDATE ai_model_price_tiers t
SET
  platform_amount = 1.000000,
  updated_at = NOW()
FROM ai_model_price_policies p
JOIN ai_models m ON m.id = p.model_id
WHERE t.price_policy_id = p.id
  AND m.model_code = 'kimi-k2.6'
  AND p.capability_code = 'chat_completion'
  AND p.deleted_at IS NULL
  AND t.deleted_at IS NULL;

WITH recalculated AS (
  SELECT
    r.id,
    ROUND((
      CASE
        WHEN LOWER(REPLACE(REPLACE(REPLACE(p.billing_unit, ' ', ''), '_', ''), '-', '')) IN ('1ktoken', '1ktokens', 'ktoken', 'ktokens') THEN r.usage_amount / 1000
        WHEN LOWER(REPLACE(REPLACE(REPLACE(p.billing_unit, ' ', ''), '_', ''), '-', '')) IN ('1mtoken', '1mtokens', 'mtoken', 'mtokens', 'milliontoken', 'milliontokens') THEN r.usage_amount / 1000000
        ELSE r.usage_amount
      END * COALESCE(t.cost_price, p.base_cost_price)
    )::numeric, 6) AS cost_amount,
    ROUND((
      CASE
        WHEN LOWER(REPLACE(REPLACE(REPLACE(p.billing_unit, ' ', ''), '_', ''), '-', '')) IN ('1ktoken', '1ktokens', 'ktoken', 'ktokens') THEN r.usage_amount / 1000
        WHEN LOWER(REPLACE(REPLACE(REPLACE(p.billing_unit, ' ', ''), '_', ''), '-', '')) IN ('1mtoken', '1mtokens', 'mtoken', 'mtokens', 'milliontoken', 'milliontokens') THEN r.usage_amount / 1000000
        ELSE r.usage_amount
      END * COALESCE(t.sale_price, p.base_sale_price)
    )::numeric, 6) AS billing_amount,
    p.platform_unit,
    ROUND((
      CASE
        WHEN LOWER(REPLACE(REPLACE(REPLACE(p.platform_unit, ' ', ''), '_', ''), '-', '')) IN ('1ktoken', '1ktokens', 'ktoken', 'ktokens') THEN r.usage_amount / 1000
        WHEN LOWER(REPLACE(REPLACE(REPLACE(p.platform_unit, ' ', ''), '_', ''), '-', '')) IN ('1mtoken', '1mtokens', 'mtoken', 'mtokens', 'milliontoken', 'milliontokens') THEN r.usage_amount / 1000000
        ELSE r.usage_amount * COALESCE(NULLIF(t.platform_amount, 0), p.base_platform_amount)
      END
    )::numeric, 6) AS platform_amount
  FROM ai_usage_records r
  JOIN ai_model_price_policies p ON p.id = r.price_policy_id
  JOIN ai_models m ON m.id = p.model_id
  LEFT JOIN ai_model_price_tiers t ON t.id = r.price_tier_id AND t.deleted_at IS NULL
  WHERE m.model_code IN ('qwen3.6-plus', 'kimi-k2.6', 'happyhorse-1.0-t2v', 'qwen-image-2.0-pro-2026-04-22')
)
UPDATE ai_usage_records r
SET
  cost_amount = recalculated.cost_amount,
  billing_amount = recalculated.billing_amount,
  platform_unit = recalculated.platform_unit,
  platform_amount = recalculated.platform_amount
FROM recalculated
WHERE r.id = recalculated.id;
