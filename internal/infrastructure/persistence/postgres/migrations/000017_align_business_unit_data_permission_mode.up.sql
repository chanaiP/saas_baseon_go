UPDATE permission
SET data_perm_mode = 'BU',
    updated_at = now()
WHERE tenant_id = 1
  AND path IN ('/business-units', 'data:business_unit')
  AND deleted_at IS NULL;
