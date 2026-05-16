#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

conninfo="${DATABASE_DSN:-host=127.0.0.1 user=saas password=saas dbname=saas_baseon port=5432 sslmode=disable TimeZone=Asia/Shanghai}"
apply="${APPLY:-0}"

where_clause="
  COALESCE(is_demo, FALSE) = TRUE
  OR COALESCE(data_source, '') IN ('demo', 'demo_seed', 'seed')
  OR request_id LIKE 'seed%'
  OR COALESCE(request_params::text, '') LIKE '%demo-seed%'
  OR COALESCE(request_params::text, '') LIKE '%demo-history-seed%'
"

echo "AI demo usage cleanup"
echo "Database: ${conninfo}"

psql "$conninfo" -v ON_ERROR_STOP=1 <<SQL
SELECT
  COUNT(*) AS demo_records,
  MIN(called_at) AS first_called_at,
  MAX(called_at) AS last_called_at
FROM ai_usage_records
WHERE ${where_clause};
SQL

if [[ "$apply" != "1" ]]; then
  echo "Dry run only. Re-run with APPLY=1 to delete these demo usage records."
  exit 0
fi

psql "$conninfo" -v ON_ERROR_STOP=1 <<SQL
WITH deleted AS (
  DELETE FROM ai_usage_records
  WHERE ${where_clause}
  RETURNING id
)
SELECT COUNT(*) AS deleted_records FROM deleted;
SQL

echo "Done."
