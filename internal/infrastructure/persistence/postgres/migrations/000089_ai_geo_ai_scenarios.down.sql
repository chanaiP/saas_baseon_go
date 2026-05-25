UPDATE ai_scenarios s
SET
  deleted_at = NOW(),
  updated_at = NOW(),
  status = 'inactive'
WHERE s.app_code = 'ai-geo'
  AND s.ai_scenario_code IN (
    'ai_geo_draft_generation',
    'ai_geo_channel_rewrite',
    'ai_geo_audit_suggestion'
  )
  AND s.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM ai_usage_records r
    WHERE r.app_code = s.app_code
      AND r.ai_scenario_code = s.ai_scenario_code
  );
