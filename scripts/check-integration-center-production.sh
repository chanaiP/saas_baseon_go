#!/usr/bin/env bash
set -euo pipefail

dsn="${DATABASE_URL:-${DATABASE_DSN:-}}"
if [[ -z "$dsn" ]]; then
  echo "DATABASE_URL or DATABASE_DSN is required" >&2
  exit 2
fi
psql_dsn="$(printf '%s' "$dsn" | sed -E 's/[[:space:]]+TimeZone=[^[:space:]]+//g')"

run_psql() {
  if [[ -n "${PSQL_CMD:-}" ]]; then
    # shellcheck disable=SC2086
    $PSQL_CMD "$@"
    return
  fi
  psql "$@"
}

run_psql "$psql_dsn" -v ON_ERROR_STOP=1 -At <<'SQL'
WITH checks AS (
  SELECT 'sys_app' AS name, COUNT(*)::int AS actual, 1 AS expected
  FROM sys_app
  WHERE app_code = 'integration-center' AND status = 'ONLINE' AND deleted_at IS NULL
  UNION ALL
  SELECT 'sys_app_entry', COUNT(*)::int, 8
  FROM sys_app_entry
  WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
  UNION ALL
  SELECT 'sys_app_permission', COUNT(*)::int, 12
  FROM sys_app_permission
  WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
  UNION ALL
  SELECT 'sys_app_api', COUNT(*)::int, 43
  FROM sys_app_api
  WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
  UNION ALL
  SELECT 'sys_app_package_feature', COUNT(*)::int, 2
  FROM sys_app_package_feature
  WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
  UNION ALL
  SELECT 'sys_app_quota', COUNT(*)::int, 3
  FROM sys_app_quota
  WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
  UNION ALL
  SELECT 'saas_feature', COUNT(*)::int, 2
  FROM saas_feature
  WHERE app_code = 'integration-center' AND feature_code IN ('integration_tenant_authorization', 'integration_data_sync')
  UNION ALL
  SELECT 'saas_quota', COUNT(*)::int, 3
  FROM saas_quota
  WHERE quota_code IN ('integration_connection_count', 'integration_api_calls_daily', 'integration_sync_records_daily')
  UNION ALL
  SELECT 'integration_sync_records_table', COUNT(*)::int, 1
  FROM information_schema.tables
  WHERE table_schema = 'public' AND table_name = 'integration_sync_records'
),
failed AS (
  SELECT * FROM checks WHERE actual < expected
)
SELECT 'check=' || name || ' actual=' || actual || ' expected_min=' || expected ||
  CASE WHEN actual >= expected THEN ' status=PASS' ELSE ' status=FAIL' END
FROM checks
ORDER BY name;

WITH orphan_checks AS (
  SELECT COUNT(*) AS n FROM integration_tenant_connections i LEFT JOIN tenant t ON t.id = i.tenant_id WHERE t.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_oauth_states i LEFT JOIN tenant t ON t.id = i.tenant_id WHERE t.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_tenant_capabilities i LEFT JOIN tenant t ON t.id = i.tenant_id WHERE t.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_sync_jobs i LEFT JOIN tenant t ON t.id = i.tenant_id WHERE t.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_quota_usages i LEFT JOIN tenant t ON t.id = i.tenant_id WHERE t.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_provider_apps i LEFT JOIN integration_platforms p ON p.id = i.platform_id AND p.deleted_at IS NULL WHERE p.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_provider_app_capabilities i LEFT JOIN integration_provider_apps p ON p.id = i.provider_app_id AND p.deleted_at IS NULL WHERE p.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_tenant_connections i LEFT JOIN integration_provider_apps p ON p.id = i.provider_app_id AND p.deleted_at IS NULL WHERE p.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_sync_jobs i LEFT JOIN integration_tenant_connections c ON c.id = i.tenant_connection_id AND c.deleted_at IS NULL WHERE c.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_sync_records i LEFT JOIN integration_sync_jobs j ON j.id = i.sync_job_id AND j.deleted_at IS NULL WHERE j.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_sync_records i LEFT JOIN integration_tenant_connections c ON c.id = i.tenant_connection_id AND c.deleted_at IS NULL WHERE c.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_webhook_events i LEFT JOIN integration_provider_apps p ON p.id = i.provider_app_id AND p.deleted_at IS NULL WHERE p.id IS NULL
),
summary AS (
  SELECT COALESCE(SUM(n), 0) AS orphan_count FROM orphan_checks
)
SELECT 'check=orphan_references actual=' || orphan_count || ' expected=0 status=' ||
  CASE WHEN orphan_count = 0 THEN 'PASS' ELSE 'FAIL' END
FROM summary;

WITH failures AS (
  SELECT 1
  FROM (
    SELECT COUNT(*)::int AS actual, 1 AS expected FROM sys_app WHERE app_code = 'integration-center' AND status = 'ONLINE' AND deleted_at IS NULL
    UNION ALL SELECT COUNT(*)::int, 8 FROM sys_app_entry WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
    UNION ALL SELECT COUNT(*)::int, 12 FROM sys_app_permission WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
    UNION ALL SELECT COUNT(*)::int, 43 FROM sys_app_api WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
    UNION ALL SELECT COUNT(*)::int, 2 FROM sys_app_package_feature WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
    UNION ALL SELECT COUNT(*)::int, 3 FROM sys_app_quota WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
    UNION ALL SELECT COUNT(*)::int, 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'integration_sync_records'
  ) c
  WHERE actual < expected
)
SELECT CASE WHEN EXISTS (SELECT 1 FROM failures) THEN 1 ELSE 0 END AS failed;
SQL

failed="$(run_psql "$psql_dsn" -v ON_ERROR_STOP=1 -At <<'SQL'
WITH checks AS (
  SELECT COUNT(*)::int AS actual, 1 AS expected FROM sys_app WHERE app_code = 'integration-center' AND status = 'ONLINE' AND deleted_at IS NULL
  UNION ALL SELECT COUNT(*)::int, 8 FROM sys_app_entry WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
  UNION ALL SELECT COUNT(*)::int, 12 FROM sys_app_permission WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
  UNION ALL SELECT COUNT(*)::int, 43 FROM sys_app_api WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
  UNION ALL SELECT COUNT(*)::int, 2 FROM sys_app_package_feature WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
  UNION ALL SELECT COUNT(*)::int, 3 FROM sys_app_quota WHERE app_code = 'integration-center' AND status = 'ACTIVE' AND deleted_at IS NULL
),
orphan_checks AS (
  SELECT COUNT(*) AS n FROM integration_tenant_connections i LEFT JOIN tenant t ON t.id = i.tenant_id WHERE t.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_sync_jobs i LEFT JOIN integration_tenant_connections c ON c.id = i.tenant_connection_id AND c.deleted_at IS NULL WHERE c.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_sync_records i LEFT JOIN integration_sync_jobs j ON j.id = i.sync_job_id AND j.deleted_at IS NULL WHERE j.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_sync_records i LEFT JOIN integration_tenant_connections c ON c.id = i.tenant_connection_id AND c.deleted_at IS NULL WHERE c.id IS NULL
  UNION ALL SELECT COUNT(*) FROM integration_webhook_events i LEFT JOIN integration_provider_apps p ON p.id = i.provider_app_id AND p.deleted_at IS NULL WHERE p.id IS NULL
)
SELECT CASE WHEN EXISTS (SELECT 1 FROM checks WHERE actual < expected) OR (SELECT COALESCE(SUM(n), 0) FROM orphan_checks) > 0 THEN 1 ELSE 0 END;
SQL
)"

if [[ "$failed" == "1" ]]; then
  echo "integration_center_production_check=FAIL" >&2
  exit 1
fi

echo "integration_center_production_check=PASS"
