UPDATE ai_base_route_models brm
SET deleted_at = NOW(),
    updated_at = NOW(),
    status = 'inactive'
FROM ai_base_routes br, ai_models m, ai_providers p
WHERE brm.base_route_id = br.id
  AND brm.model_id = m.id
  AND m.provider_id = p.id
  AND br.route_code = 'chat-default'
  AND p.code = 'dashscope'
  AND m.model_code IN ('qwen-turbo', 'qwen-plus')
  AND brm.deleted_at IS NULL;
