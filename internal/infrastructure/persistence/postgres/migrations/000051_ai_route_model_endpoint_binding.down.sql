DROP INDEX IF EXISTS idx_ai_base_route_models_provider_api_id;
DROP INDEX IF EXISTS idx_ai_base_route_models_provider_account_id;

ALTER TABLE ai_base_route_models
  DROP COLUMN IF EXISTS provider_api_id,
  DROP COLUMN IF EXISTS provider_account_id;
