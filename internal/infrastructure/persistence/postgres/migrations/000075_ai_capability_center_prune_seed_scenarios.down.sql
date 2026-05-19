UPDATE ai_scenarios s
SET deleted_at = NULL, updated_at = NOW()
WHERE (s.app_code, s.ai_scenario_code) IN (
  ('app-center', 'app_description_generate'),
  ('integration-center', 'connector_mapping_reasoning'),
  ('data-center', 'knowledge_embedding'),
  ('workbench', 'poster_image_generate'),
  ('system-management', 'policy_doc_review')
);
