-- Remove the retired Alibaba Bailian / DashScope provider catalog from AI capability center.
-- Records are logically deleted to preserve historical usage and audit references.

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
SET deleted_at = NOW(),
    updated_at = NOW(),
    status = 'inactive'
WHERE brm.deleted_at IS NULL
  AND brm.model_id IN (SELECT id FROM model_scope);

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
SET deleted_at = NOW(),
    updated_at = NOW(),
    enabled = false
WHERE t.deleted_at IS NULL
  AND t.price_policy_id IN (SELECT id FROM price_policy_scope);

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
SET deleted_at = NOW(),
    updated_at = NOW(),
    status = 'inactive'
WHERE pp.deleted_at IS NULL
  AND pp.model_id IN (SELECT id FROM model_scope);

WITH provider_scope AS (
  SELECT id
  FROM ai_providers
  WHERE code IN ('dashscope', 'aliyun-bailian')
     OR base_url ILIKE '%dashscope.aliyuncs.com%'
)
UPDATE ai_provider_apis api
SET deleted_at = NOW(),
    updated_at = NOW(),
    status = 'inactive'
WHERE api.deleted_at IS NULL
  AND api.provider_id IN (SELECT id FROM provider_scope);

WITH provider_scope AS (
  SELECT id
  FROM ai_providers
  WHERE code IN ('dashscope', 'aliyun-bailian')
     OR base_url ILIKE '%dashscope.aliyuncs.com%'
)
UPDATE ai_provider_accounts account
SET deleted_at = NOW(),
    updated_at = NOW(),
    status = 'inactive'
WHERE account.deleted_at IS NULL
  AND account.provider_id IN (SELECT id FROM provider_scope);

WITH provider_scope AS (
  SELECT id
  FROM ai_providers
  WHERE code IN ('dashscope', 'aliyun-bailian')
     OR base_url ILIKE '%dashscope.aliyuncs.com%'
)
UPDATE ai_models model
SET deleted_at = NOW(),
    updated_at = NOW(),
    status = 'inactive'
WHERE model.deleted_at IS NULL
  AND model.provider_id IN (SELECT id FROM provider_scope);

UPDATE ai_providers
SET deleted_at = NOW(),
    updated_at = NOW(),
    status = 'inactive'
WHERE deleted_at IS NULL
  AND (
    code IN ('dashscope', 'aliyun-bailian')
    OR base_url ILIKE '%dashscope.aliyuncs.com%'
  );
