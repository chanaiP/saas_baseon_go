CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS ai_providers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL,
  code VARCHAR(100) NOT NULL UNIQUE,
  type VARCHAR(32) NOT NULL,
  base_url VARCHAR(500) NOT NULL,
  auth_type VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  priority INT NOT NULL DEFAULT 0,
  region VARCHAR(100),
  qps_limit INT NOT NULL DEFAULT 0,
  monthly_budget NUMERIC(18, 4) NOT NULL DEFAULT 0,
  owner VARCHAR(100),
  remark TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_provider_accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  provider_id UUID NOT NULL REFERENCES ai_providers(id),
  account_name VARCHAR(100) NOT NULL,
  endpoint VARCHAR(500),
  key_alias VARCHAR(200) NOT NULL,
  login_method VARCHAR(32),
  login_account VARCHAR(200),
  maintainer VARCHAR(100),
  maintainer_contact VARCHAR(100),
  encrypted_api_key TEXT,
  encrypted_secret TEXT,
  quota_limit NUMERIC(24, 4) NOT NULL DEFAULT 0,
  used_quota NUMERIC(24, 4) NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_provider_apis (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  provider_id UUID NOT NULL REFERENCES ai_providers(id),
  account_id UUID NOT NULL REFERENCES ai_provider_accounts(id),
  api_name VARCHAR(150) NOT NULL,
  api_path VARCHAR(500) NOT NULL,
  api_type VARCHAR(32) NOT NULL,
  capabilities JSONB NOT NULL DEFAULT '[]'::jsonb,
  auth_type VARCHAR(32) NOT NULL,
  qps_limit INT NOT NULL DEFAULT 0,
  timeout_ms INT NOT NULL DEFAULT 30000,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  last_called_at TIMESTAMPTZ,
  health_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
  health_message VARCHAR(500),
  health_checked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_capabilities (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  capability_code VARCHAR(100) NOT NULL UNIQUE,
  capability_name VARCHAR(100) NOT NULL,
  scenario_type VARCHAR(32) NOT NULL,
  model_type VARCHAR(32) NOT NULL,
  default_billing_unit VARCHAR(32) NOT NULL,
  supports_tier_pricing BOOLEAN NOT NULL DEFAULT false,
  description TEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_models (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  provider_id UUID NOT NULL REFERENCES ai_providers(id),
  model_code VARCHAR(150) NOT NULL,
  model_name VARCHAR(150) NOT NULL,
  model_type VARCHAR(32) NOT NULL,
  capabilities JSONB NOT NULL DEFAULT '[]'::jsonb,
  context_window INT,
  unit VARCHAR(50),
  latency_p95 INT NOT NULL DEFAULT 0,
  success_rate NUMERIC(8, 4) NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  default_for JSONB NOT NULL DEFAULT '[]'::jsonb,
  remark TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_model_price_policies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  model_id UUID NOT NULL REFERENCES ai_models(id),
  feature_key VARCHAR(200) NOT NULL,
  feature_name VARCHAR(150) NOT NULL,
  model_type VARCHAR(32) NOT NULL,
  capability_code VARCHAR(100) NOT NULL REFERENCES ai_capabilities(capability_code),
  billing_mode VARCHAR(32) NOT NULL,
  billing_unit VARCHAR(32) NOT NULL,
  platform_unit VARCHAR(32) NOT NULL,
  base_cost_price NUMERIC(18, 6) NOT NULL DEFAULT 0,
  base_sale_price NUMERIC(18, 6) NOT NULL DEFAULT 0,
  base_platform_amount NUMERIC(18, 6) NOT NULL DEFAULT 0,
  currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_model_price_tiers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  price_policy_id UUID NOT NULL REFERENCES ai_model_price_policies(id),
  tier_name VARCHAR(150) NOT NULL,
  mode VARCHAR(64),
  resolution VARCHAR(64),
  quality VARCHAR(64),
  duration_seconds INT,
  aspect_ratio VARCHAR(64),
  cost_price NUMERIC(18, 6) NOT NULL DEFAULT 0,
  sale_price NUMERIC(18, 6) NOT NULL DEFAULT 0,
  platform_amount NUMERIC(18, 6) NOT NULL DEFAULT 0,
  enabled BOOLEAN NOT NULL DEFAULT true,
  sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_base_routes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  route_code VARCHAR(150) NOT NULL UNIQUE,
  route_name VARCHAR(150) NOT NULL,
  capability_code VARCHAR(100) NOT NULL REFERENCES ai_capabilities(capability_code),
  model_type VARCHAR(32) NOT NULL,
  strategy VARCHAR(32) NOT NULL,
  timeout_ms INT NOT NULL DEFAULT 30000,
  max_retry INT NOT NULL DEFAULT 0,
  description TEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_base_route_models (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  base_route_id UUID NOT NULL REFERENCES ai_base_routes(id),
  model_id UUID NOT NULL REFERENCES ai_models(id),
  role VARCHAR(32) NOT NULL,
  priority INT NOT NULL DEFAULT 1,
  weight INT NOT NULL DEFAULT 100,
  max_retry INT NOT NULL DEFAULT 0,
  timeout_ms INT NOT NULL DEFAULT 30000,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_scenarios (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  app_code VARCHAR(100) NOT NULL,
  app_name VARCHAR(100) NOT NULL,
  ai_scenario_code VARCHAR(150) NOT NULL,
  ai_scenario_name VARCHAR(150) NOT NULL,
  scenario_type VARCHAR(32) NOT NULL,
  capability_code VARCHAR(100) NOT NULL REFERENCES ai_capabilities(capability_code),
  model_type VARCHAR(32) NOT NULL,
  default_base_route_id UUID NOT NULL REFERENCES ai_base_routes(id),
  owner VARCHAR(100),
  description TEXT,
  version VARCHAR(50),
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_tenant_strategy_policies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  policy_name VARCHAR(150) NOT NULL,
  tenant_scope VARCHAR(32) NOT NULL,
  tenant_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
  app_code VARCHAR(100) NOT NULL,
  app_name VARCHAR(100) NOT NULL,
  ai_scenario_code VARCHAR(150) NOT NULL,
  ai_scenario_name VARCHAR(150) NOT NULL,
  default_base_route_id UUID NOT NULL REFERENCES ai_base_routes(id),
  override_base_route_id UUID REFERENCES ai_base_routes(id),
  description TEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_strategy_quota_rules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  policy_id UUID NOT NULL REFERENCES ai_tenant_strategy_policies(id),
  dimension VARCHAR(32) NOT NULL,
  subject_code VARCHAR(150) NOT NULL,
  usage_unit VARCHAR(32) NOT NULL,
  period VARCHAR(32) NOT NULL,
  quota_limit NUMERIC(24, 6) NOT NULL DEFAULT 0,
  used_amount NUMERIC(24, 6) NOT NULL DEFAULT 0,
  warning_threshold NUMERIC(8, 4) NOT NULL DEFAULT 80,
  over_limit_action VARCHAR(32) NOT NULL DEFAULT 'alert_only',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_strategy_rate_limit_rules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  policy_id UUID NOT NULL REFERENCES ai_tenant_strategy_policies(id),
  dimension VARCHAR(32) NOT NULL,
  subject_code VARCHAR(150) NOT NULL,
  qps INT NOT NULL DEFAULT 0,
  concurrency INT NOT NULL DEFAULT 0,
  minute_limit INT NOT NULL DEFAULT 0,
  hour_limit INT NOT NULL DEFAULT 0,
  day_limit INT NOT NULL DEFAULT 0,
  over_limit_action VARCHAR(32) NOT NULL DEFAULT 'queue',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ai_usage_records (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id VARCHAR(100) NOT NULL UNIQUE,
  tenant_id VARCHAR(100) NOT NULL,
  tenant_name VARCHAR(150),
  app_code VARCHAR(100) NOT NULL,
  app_name VARCHAR(100),
  ai_scenario_code VARCHAR(150) NOT NULL,
  ai_scenario_name VARCHAR(150),
  user_id VARCHAR(100),
  user_name VARCHAR(100),
  provider_id UUID REFERENCES ai_providers(id),
  provider_account_id UUID REFERENCES ai_provider_accounts(id),
  provider_api_id UUID REFERENCES ai_provider_apis(id),
  model_id UUID REFERENCES ai_models(id),
  base_route_id UUID REFERENCES ai_base_routes(id),
  tenant_strategy_id UUID REFERENCES ai_tenant_strategy_policies(id),
  price_policy_id UUID REFERENCES ai_model_price_policies(id),
  price_tier_id UUID REFERENCES ai_model_price_tiers(id),
  usage_amount NUMERIC(24, 6) NOT NULL DEFAULT 0,
  usage_unit VARCHAR(32) NOT NULL,
  usage_detail VARCHAR(200),
  calls INT NOT NULL DEFAULT 1,
  cost_amount NUMERIC(18, 6) NOT NULL DEFAULT 0,
  billing_amount NUMERIC(18, 6) NOT NULL DEFAULT 0,
  platform_unit VARCHAR(32),
  platform_amount NUMERIC(18, 6) NOT NULL DEFAULT 0,
  latency_ms INT NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL,
  error_code VARCHAR(100),
  error_message TEXT,
  request_params JSONB NOT NULL DEFAULT '{}'::jsonb,
  prompt_hash VARCHAR(128),
  response_hash VARCHAR(128),
  called_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ai_gateway_settings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  setting_key VARCHAR(100) NOT NULL UNIQUE,
  setting_value JSONB NOT NULL DEFAULT '{}'::jsonb,
  description TEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_ai_provider_account_name ON ai_provider_accounts(provider_id, account_name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_ai_models_provider_code ON ai_models(provider_id, model_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_ai_model_price_feature ON ai_model_price_policies(model_id, feature_key) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_ai_scenarios_app_scenario ON ai_scenarios(app_code, ai_scenario_code) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_ai_providers_status ON ai_providers(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_provider_apis_provider_account ON ai_provider_apis(provider_id, account_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_models_provider_status ON ai_models(provider_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_base_routes_capability_status ON ai_base_routes(capability_code, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_usage_tenant_time ON ai_usage_records(tenant_id, called_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_usage_app_scenario_time ON ai_usage_records(app_code, ai_scenario_code, called_at DESC);
