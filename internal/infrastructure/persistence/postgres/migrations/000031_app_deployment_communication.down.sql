UPDATE dict_item di
SET enabled = TRUE, deleted_at = NULL, updated_at = now()
FROM dict_type dt
WHERE di.dict_type_id = dt.id
  AND dt.code = 'app_type'
  AND di.value = 'CLIENT_APP';

UPDATE dict_item di
SET label = CASE di.value
    WHEN 'SYSTEM_APP' THEN '系统内置型'
    WHEN 'BUSINESS_APP' THEN '独立业务型'
    WHEN 'ABILITY_APP' THEN '业务中台型'
    WHEN 'API_APP' THEN 'API 能力型'
    WHEN 'CONNECTOR_APP' THEN '连接器型'
    WHEN 'AI_APP' THEN 'AI / Agent 型'
    WHEN 'SUITE_APP' THEN '组合套件型'
    ELSE di.label
  END,
  updated_at = now()
FROM dict_type dt
WHERE di.dict_type_id = dt.id
  AND dt.code = 'app_type'
  AND di.value IN ('SYSTEM_APP', 'BUSINESS_APP', 'ABILITY_APP', 'API_APP', 'CONNECTOR_APP', 'AI_APP', 'SUITE_APP');

DROP INDEX IF EXISTS idx_sys_app_deployment_mode_deleted;

ALTER TABLE sys_app
  DROP COLUMN IF EXISTS communication_modes,
  DROP COLUMN IF EXISTS deployment_mode;
