UPDATE ai_model_price_policies p
SET
  billing_unit = 'tokens',
  platform_unit = '1K tokens',
  base_cost_price = 0.000800,
  base_sale_price = 0.001600,
  base_platform_amount = 0.000200,
  updated_at = NOW()
FROM ai_models m
WHERE p.model_id = m.id
  AND m.model_code = 'qwen3.6-plus'
  AND p.capability_code = 'chat_completion'
  AND p.deleted_at IS NULL;

UPDATE ai_model_price_tiers t
SET
  tier_name = '默认文本价格',
  cost_price = 0.000800,
  sale_price = 0.001600,
  platform_amount = 0.000200,
  updated_at = NOW()
FROM ai_model_price_policies p
JOIN ai_models m ON m.id = p.model_id
WHERE t.price_policy_id = p.id
  AND m.model_code = 'qwen3.6-plus'
  AND p.capability_code = 'chat_completion'
  AND p.deleted_at IS NULL
  AND t.deleted_at IS NULL;

UPDATE ai_model_price_policies p
SET
  billing_unit = 'tokens',
  platform_unit = '1K tokens',
  base_platform_amount = 0.000200,
  updated_at = NOW()
FROM ai_models m
WHERE p.model_id = m.id
  AND m.model_code = 'kimi-k2.6'
  AND p.capability_code = 'chat_completion'
  AND p.deleted_at IS NULL;

UPDATE ai_model_price_tiers t
SET
  platform_amount = 0.000200,
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
    ROUND((r.usage_amount * COALESCE(t.cost_price, p.base_cost_price))::numeric, 6) AS cost_amount,
    ROUND((r.usage_amount * COALESCE(t.sale_price, p.base_sale_price))::numeric, 6) AS billing_amount,
    p.platform_unit,
    ROUND((r.usage_amount * COALESCE(NULLIF(t.platform_amount, 0), p.base_platform_amount))::numeric, 6) AS platform_amount
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
