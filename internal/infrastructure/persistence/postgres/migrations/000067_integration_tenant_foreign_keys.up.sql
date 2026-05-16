ALTER TABLE integration_tenant_connections
  ADD CONSTRAINT fk_integration_tenant_connections_tenant
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) NOT VALID;

ALTER TABLE integration_oauth_states
  ADD CONSTRAINT fk_integration_oauth_states_tenant
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) NOT VALID;

ALTER TABLE integration_tenant_capabilities
  ADD CONSTRAINT fk_integration_tenant_capabilities_tenant
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) NOT VALID;

ALTER TABLE integration_sync_jobs
  ADD CONSTRAINT fk_integration_sync_jobs_tenant
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) NOT VALID;

ALTER TABLE integration_quota_bindings
  ADD CONSTRAINT fk_integration_quota_bindings_tenant
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) NOT VALID;

ALTER TABLE integration_quota_usages
  ADD CONSTRAINT fk_integration_quota_usages_tenant
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) NOT VALID;

ALTER TABLE integration_alerts
  ADD CONSTRAINT fk_integration_alerts_tenant
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) NOT VALID;

ALTER TABLE integration_api_call_logs
  ADD CONSTRAINT fk_integration_api_call_logs_tenant
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) NOT VALID;
