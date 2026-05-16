#!/usr/bin/env bash
set -euo pipefail

dsn="${MIGRATION_IDEMPOTENCY_DSN:-${DATABASE_URL:-${DATABASE_DSN:-}}}"
if [[ -z "$dsn" ]]; then
  echo "MIGRATION_IDEMPOTENCY_DSN or DATABASE_URL/DATABASE_DSN is required" >&2
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

if [[ "${RESET_PUBLIC_SCHEMA:-0}" == "1" ]]; then
  run_psql "$psql_dsn" -v ON_ERROR_STOP=1 -q <<'SQL'
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
SQL
fi

snapshot_counts() {
  run_psql "$psql_dsn" -v ON_ERROR_STOP=1 -At <<'SQL'
WITH tracked(table_name) AS (
  VALUES
    ('schema_migrations'),
    ('sys_app'),
    ('sys_menu'),
    ('permission'),
    ('saas_plan_feature'),
    ('saas_quota'),
    ('saas_plan_quota'),
    ('integration_platforms'),
    ('integration_provider_apps'),
    ('integration_platform_capabilities'),
    ('integration_provider_app_capabilities'),
    ('integration_quota_policies')
)
SELECT tracked.table_name || '=' ||
  CASE WHEN to_regclass('public.' || tracked.table_name) IS NULL THEN 'missing'
       ELSE (xpath('/row/c/text()', query_to_xml(format('SELECT count(*) AS c FROM public.%I', tracked.table_name), false, true, '')))[1]::text
  END
FROM tracked
ORDER BY tracked.table_name;
SQL
}

run_migrate() {
  DATABASE_DSN="$dsn" go run ./cmd/migrate >/tmp/integration-center-migrate-idempotency.log
}

run_migrate
first="$(snapshot_counts)"
run_migrate
second="$(snapshot_counts)"

if [[ "$first" != "$second" ]]; then
  echo "migration is not idempotent: tracked row counts changed after second run" >&2
  diff -u <(printf '%s\n' "$first") <(printf '%s\n' "$second") || true
  exit 1
fi

pending="$(run_psql "$psql_dsn" -v ON_ERROR_STOP=1 -At <<'SQL'
SELECT COUNT(*)
FROM public.schema_migrations
WHERE version IS NULL OR checksum IS NULL OR checksum = '';
SQL
)"
if [[ "$pending" != "0" ]]; then
  echo "schema_migrations contains invalid rows" >&2
  exit 1
fi

echo "migration idempotency check passed"
