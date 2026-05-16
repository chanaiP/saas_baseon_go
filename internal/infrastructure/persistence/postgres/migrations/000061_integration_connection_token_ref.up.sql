ALTER TABLE integration_tenant_connections
  ADD COLUMN IF NOT EXISTS token_credential_ref VARCHAR(240);
