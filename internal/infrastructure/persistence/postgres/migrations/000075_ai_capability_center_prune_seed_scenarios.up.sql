-- Remove cross-application AI scenario seeds from the AI capability center.
-- The scenario registry should show real scenarios added for actual provider/API use,
-- not bootstrap examples for unrelated built-in apps.

UPDATE ai_scenarios s
SET deleted_at = NOW(), updated_at = NOW()
WHERE s.deleted_at IS NULL
  AND (s.app_code, s.ai_scenario_code) IN (
    ('app-center', 'app_description_generate'),
    ('integration-center', 'connector_mapping_reasoning'),
    ('data-center', 'knowledge_embedding'),
    ('workbench', 'poster_image_generate'),
    ('system-management', 'policy_doc_review')
  )
  AND NOT EXISTS (
    SELECT 1
    FROM ai_usage_records r
    WHERE r.app_code = s.app_code
      AND r.ai_scenario_code = s.ai_scenario_code
  );
