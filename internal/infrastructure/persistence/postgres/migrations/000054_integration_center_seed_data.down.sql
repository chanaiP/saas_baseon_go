DELETE FROM integration_api_call_logs
WHERE request_id = 'seed-wecom-contact-list';

DELETE FROM integration_alerts
WHERE alert_type = 'oauth_review'
  AND title = '京东服务商应用仍在联调审核中';

DELETE FROM integration_sync_jobs
WHERE capability_code = 'org_sync'
  AND job_type = 'incremental'
  AND trigger_mode = 'schedule';

DELETE FROM integration_tenant_connections
WHERE auth_subject_type = 'corp'
  AND auth_subject_id = 'demo-corp';

DELETE FROM integration_quota_policies
WHERE policy_code IN (
  'integration_default_connections',
  'integration_default_api_daily',
  'integration_default_sync_daily'
);

DELETE FROM integration_provider_app_capabilities
WHERE provider_app_id IN (
  SELECT id FROM integration_provider_apps
  WHERE app_code IN ('wecom-suite-standard', 'douyin-shop-isv-prod', 'jd-shop-isv-beta', 'generic-hmac-template')
);

DELETE FROM integration_provider_apps
WHERE app_code IN ('wecom-suite-standard', 'douyin-shop-isv-prod', 'jd-shop-isv-beta', 'generic-hmac-template');

DELETE FROM integration_platform_capabilities
WHERE capability_code IN ('org_sync', 'approval_event', 'order_sync', 'product_sync', 'api_proxy')
  AND platform_id IN (
    SELECT id FROM integration_platforms
    WHERE platform_code IN ('wecom', 'douyin-shop', 'jd-shop', 'generic-openapi')
  );

DELETE FROM integration_platforms
WHERE platform_code IN ('wecom', 'douyin-shop', 'jd-shop', 'generic-openapi');
