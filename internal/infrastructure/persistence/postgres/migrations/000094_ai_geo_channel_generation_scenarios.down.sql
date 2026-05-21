UPDATE ai_scenarios s
SET deleted_at = NOW(),
    updated_at = NOW()
WHERE app_code = 'ai-geo'
  AND ai_scenario_code IN (
    'channel_content_standard_generate',
    'ai_geo_channel_content_editor'
  )
  AND deleted_at IS NULL;
