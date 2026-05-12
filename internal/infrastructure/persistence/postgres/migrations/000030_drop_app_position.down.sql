ALTER TABLE sys_app
  ADD COLUMN IF NOT EXISTS app_position VARCHAR(64);

UPDATE dict_item di
SET label = '独立业务型', updated_at = now()
FROM dict_type dt
WHERE di.dict_type_id = dt.id
  AND dt.code = 'app_type'
  AND di.value = 'BUSINESS_APP';
