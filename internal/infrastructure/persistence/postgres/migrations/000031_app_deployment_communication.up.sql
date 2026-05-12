ALTER TABLE sys_app
  ADD COLUMN IF NOT EXISTS deployment_mode VARCHAR(32) NOT NULL DEFAULT 'MERGED',
  ADD COLUMN IF NOT EXISTS communication_modes TEXT;

UPDATE sys_app
SET deployment_mode = 'MERGED'
WHERE deployment_mode IS NULL OR deployment_mode = '';

CREATE INDEX IF NOT EXISTS idx_sys_app_deployment_mode_deleted
  ON sys_app (deployment_mode, deleted_at);

UPDATE dict_item di
SET label = CASE di.value
    WHEN 'SYSTEM_APP' THEN '系统底座'
    WHEN 'BUSINESS_APP' THEN '业务系统'
    WHEN 'ABILITY_APP' THEN '业务中台'
    WHEN 'API_APP' THEN 'API 应用'
    WHEN 'CONNECTOR_APP' THEN '连接器'
    WHEN 'AI_APP' THEN 'AI / Agent'
    WHEN 'SUITE_APP' THEN '组合套件'
    ELSE di.label
  END,
  sort_order = CASE di.value
    WHEN 'SYSTEM_APP' THEN 1
    WHEN 'BUSINESS_APP' THEN 2
    WHEN 'ABILITY_APP' THEN 3
    WHEN 'API_APP' THEN 4
    WHEN 'CONNECTOR_APP' THEN 5
    WHEN 'AI_APP' THEN 6
    WHEN 'SUITE_APP' THEN 7
    ELSE di.sort_order
  END,
  enabled = TRUE,
  deleted_at = NULL,
  updated_at = now()
FROM dict_type dt
WHERE di.dict_type_id = dt.id
  AND dt.code = 'app_type'
  AND di.value IN ('SYSTEM_APP', 'BUSINESS_APP', 'ABILITY_APP', 'API_APP', 'CONNECTOR_APP', 'AI_APP', 'SUITE_APP');

UPDATE dict_item di
SET enabled = FALSE, deleted_at = now(), updated_at = now()
FROM dict_type dt
WHERE di.dict_type_id = dt.id
  AND dt.code = 'app_type'
  AND di.value = 'CLIENT_APP';
