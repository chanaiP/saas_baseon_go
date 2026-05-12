UPDATE sys_app
SET status = 'DRAFT'
WHERE status = 'INITIATED';

UPDATE dict_item di
SET label = '草稿',
    value = 'DRAFT',
    updated_at = now()
FROM dict_type dt
WHERE di.dict_type_id = dt.id
  AND dt.code = 'app_status'
  AND di.value = 'INITIATED';
