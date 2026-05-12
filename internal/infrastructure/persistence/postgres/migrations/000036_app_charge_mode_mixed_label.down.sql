UPDATE dict_item di
SET label = '组合计费',
    updated_at = NOW()
FROM dict_type dt
WHERE di.dict_type_id = dt.id
  AND dt.code = 'app_charge_mode'
  AND di.value = 'MIXED';
