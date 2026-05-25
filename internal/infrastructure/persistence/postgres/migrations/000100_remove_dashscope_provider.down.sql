-- Best-effort rollback for the retired DashScope catalog removal.

WITH provider_scope AS (
  SELECT id
  FROM ai_providers
  WHERE code IN ('dashscope', 'aliyun-bailian')
     OR base_url ILIKE '%dashscope.aliyuncs.com%'
),
model_scope AS (
  SELECT m.id
  FROM ai_models m
  JOIN provider_scope p ON p.id = m.provider_id
)
UPDATE ai_base_route_models brm
SET deleted_at = NULL,
    updated_at = NOW(),
    status = 'active'
WHERE brm.model_id IN (SELECT id FROM model_scope);

WITH provider_scope AS (
  SELECT id
  FROM ai_providers
  WHERE code IN ('dashscope', 'aliyun-bailian')
     OR base_url ILIKE '%dashscope.aliyuncs.com%'
),
model_scope AS (
  SELECT m.id
  FROM ai_models m
  JOIN provider_scope p ON p.id = m.provider_id
),
price_policy_scope AS (
  SELECT pp.id
  FROM ai_model_price_policies pp
  JOIN model_scope m ON m.id = pp.model_id
)
UPDATE ai_model_price_tiers t
SET deleted_at = NULL,
    updated_at = NOW(),
    enabled = true
WHERE t.price_policy_id IN (SELECT id FROM price_policy_scope);

WITH provider_scope AS (
  SELECT id
  FROM ai_providers
  WHERE code IN ('dashscope', 'aliyun-bailian')
     OR base_url ILIKE '%dashscope.aliyuncs.com%'
),
model_scope AS (
  SELECT m.id
  FROM ai_models m
  JOIN provider_scope p ON p.id = m.provider_id
)
UPDATE ai_model_price_policies pp
SET deleted_at = NULL,
    updated_at = NOW(),
    status = 'active'
WHERE pp.model_id IN (SELECT id FROM model_scope);

WITH provider_scope AS (
  SELECT id
  FROM ai_providers
  WHERE code IN ('dashscope', 'aliyun-bailian')
     OR base_url ILIKE '%dashscope.aliyuncs.com%'
)
UPDATE ai_provider_apis api
SET deleted_at = NULL,
    updated_at = NOW(),
    status = 'active'
WHERE api.provider_id IN (SELECT id FROM provider_scope);

WITH provider_scope AS (
  SELECT id
  FROM ai_providers
  WHERE code IN ('dashscope', 'aliyun-bailian')
     OR base_url ILIKE '%dashscope.aliyuncs.com%'
)
UPDATE ai_provider_accounts account
SET deleted_at = NULL,
    updated_at = NOW(),
    status = 'active'
WHERE account.provider_id IN (SELECT id FROM provider_scope);

WITH provider_scope AS (
  SELECT id
  FROM ai_providers
  WHERE code IN ('dashscope', 'aliyun-bailian')
     OR base_url ILIKE '%dashscope.aliyuncs.com%'
)
UPDATE ai_models model
SET deleted_at = NULL,
    updated_at = NOW(),
    status = 'active'
WHERE model.provider_id IN (SELECT id FROM provider_scope);

UPDATE ai_providers
SET deleted_at = NULL,
    updated_at = NOW(),
    status = 'active'
WHERE code IN ('dashscope', 'aliyun-bailian')
   OR base_url ILIKE '%dashscope.aliyuncs.com%';
