CREATE TABLE IF NOT EXISTS integration_oauth_states (
  id BIGSERIAL PRIMARY KEY,
  state VARCHAR(160) NOT NULL UNIQUE,
  tenant_id BIGINT NOT NULL,
  provider_app_id BIGINT NOT NULL REFERENCES integration_provider_apps(id),
  platform_id BIGINT NOT NULL REFERENCES integration_platforms(id),
  redirect_uri VARCHAR(500) NOT NULL,
  scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ,
  created_by BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_integration_oauth_states_status
  ON integration_oauth_states(provider_app_id, status, expires_at)
  WHERE deleted_at IS NULL;
