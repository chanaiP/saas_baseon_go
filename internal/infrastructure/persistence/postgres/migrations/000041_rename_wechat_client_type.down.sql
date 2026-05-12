UPDATE dict_item di
SET label = '微信',
    updated_at = now()
FROM dict_type dt
WHERE di.dict_type_id = dt.id
  AND dt.code = 'app_client_type'
  AND di.value = 'WECHAT'
  AND di.deleted_at IS NULL;
