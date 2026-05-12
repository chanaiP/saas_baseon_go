UPDATE sys_app
SET status = 'INITIATED'
WHERE status = 'DRAFT';

UPDATE dict_item di
SET label = '立项',
    value = 'INITIATED',
    updated_at = now()
FROM dict_type dt
WHERE di.dict_type_id = dt.id
  AND dt.code = 'app_status'
  AND di.value = 'DRAFT';

INSERT INTO dict_item (tenant_id, dict_type_id, label, value, sort_order, enabled, created_at, updated_at, deleted_at)
SELECT dt.tenant_id, dt.id, '立项', 'INITIATED', 0, true, now(), now(), NULL
FROM dict_type dt
WHERE dt.code = 'app_status'
  AND NOT EXISTS (
    SELECT 1
    FROM dict_item di
    WHERE di.dict_type_id = dt.id
      AND di.value = 'INITIATED'
      AND di.deleted_at IS NULL
  );
