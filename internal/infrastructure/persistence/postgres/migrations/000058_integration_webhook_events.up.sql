CREATE TABLE IF NOT EXISTS integration_webhook_events (
  id BIGSERIAL PRIMARY KEY,
  provider_app_id BIGINT NOT NULL REFERENCES integration_provider_apps(id),
  platform_id BIGINT NOT NULL REFERENCES integration_platforms(id),
  event_type VARCHAR(120) NOT NULL,
  idempotency_key VARCHAR(160) NOT NULL,
  signature VARCHAR(160) NOT NULL,
  payload_digest VARCHAR(128) NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  status VARCHAR(32) NOT NULL DEFAULT 'received',
  received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  processed_at TIMESTAMPTZ,
  retry_count INT NOT NULL DEFAULT 0,
  next_retry_at TIMESTAMPTZ,
  error_message TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_integration_webhook_events_idempotency
  ON integration_webhook_events(provider_app_id, idempotency_key)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_webhook_events_status
  ON integration_webhook_events(provider_app_id, status, received_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_webhook_events_retry
  ON integration_webhook_events(status, next_retry_at, received_at)
  WHERE deleted_at IS NULL;
