WITH chat_route AS (
  SELECT id
  FROM ai_base_routes
  WHERE route_code = 'chat-default'
    AND deleted_at IS NULL
  ORDER BY updated_at DESC
  LIMIT 1
),
scenario_seed(app_code, app_name, ai_scenario_code, ai_scenario_name, scenario_type, capability_code, model_type, description, version, default_base_route_id) AS (
  SELECT *
  FROM (
    VALUES
      ('ai-geo', 'AI GEO', 'ai_geo_draft_generation', 'AI GEO 母稿生成', 'text', 'chat_completion', 'text', '基于品牌、商品、Skill、热点和用户提示生成可审核母稿。', 'v1.0'),
      ('ai-geo', 'AI GEO', 'ai_geo_channel_rewrite', 'AI GEO 渠道内容改写', 'text', 'chat_completion', 'text', '基于母稿和渠道资料改写渠道内容版本。', 'v1.0'),
      ('ai-geo', 'AI GEO', 'ai_geo_audit_suggestion', 'AI GEO 内容审核建议', 'text', 'chat_completion', 'text', '对母稿和渠道内容输出风险等级、通过建议和结构化审核意见。', 'v1.0')
  ) AS v(app_code, app_name, ai_scenario_code, ai_scenario_name, scenario_type, capability_code, model_type, description, version)
  CROSS JOIN chat_route
),
updated AS (
  UPDATE ai_scenarios s
  SET
    app_name = seed.app_name,
    ai_scenario_name = seed.ai_scenario_name,
    scenario_type = seed.scenario_type,
    capability_code = seed.capability_code,
    model_type = seed.model_type,
    owner = COALESCE(NULLIF(s.owner, ''), 'AI GEO 产品组'),
    description = seed.description,
    version = seed.version,
    status = 'active',
    updated_at = NOW(),
    deleted_at = NULL
  FROM scenario_seed seed
  WHERE s.app_code = seed.app_code
    AND s.ai_scenario_code = seed.ai_scenario_code
    AND s.deleted_at IS NULL
  RETURNING s.app_code, s.ai_scenario_code
)
INSERT INTO ai_scenarios (
  app_code,
  app_name,
  ai_scenario_code,
  ai_scenario_name,
  scenario_type,
  capability_code,
  model_type,
  default_base_route_id,
  owner,
  description,
  version,
  status,
  created_at,
  updated_at
)
SELECT
  seed.app_code,
  seed.app_name,
  seed.ai_scenario_code,
  seed.ai_scenario_name,
  seed.scenario_type,
  seed.capability_code,
  seed.model_type,
  seed.default_base_route_id,
  'AI GEO 产品组',
  seed.description,
  seed.version,
  'active',
  NOW(),
  NOW()
FROM scenario_seed seed
WHERE NOT EXISTS (
  SELECT 1
  FROM updated u
  WHERE u.app_code = seed.app_code
    AND u.ai_scenario_code = seed.ai_scenario_code
)
  AND NOT EXISTS (
    SELECT 1
    FROM ai_scenarios s
    WHERE s.app_code = seed.app_code
      AND s.ai_scenario_code = seed.ai_scenario_code
      AND s.deleted_at IS NULL
  );
