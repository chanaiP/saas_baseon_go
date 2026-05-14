ALTER TABLE ai_base_route_models
  ADD COLUMN IF NOT EXISTS provider_account_id UUID REFERENCES ai_provider_accounts(id),
  ADD COLUMN IF NOT EXISTS provider_api_id UUID REFERENCES ai_provider_apis(id);

CREATE INDEX IF NOT EXISTS idx_ai_base_route_models_provider_account_id
  ON ai_base_route_models(provider_account_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_ai_base_route_models_provider_api_id
  ON ai_base_route_models(provider_api_id)
  WHERE deleted_at IS NULL;
