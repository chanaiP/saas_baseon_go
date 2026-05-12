ALTER TABLE sys_app
  DROP COLUMN IF EXISTS app_position;

UPDATE dict_item di
SET label = '业务系统', updated_at = now()
FROM dict_type dt
WHERE di.dict_type_id = dt.id
  AND dt.code = 'app_type'
  AND di.value = 'BUSINESS_APP';
