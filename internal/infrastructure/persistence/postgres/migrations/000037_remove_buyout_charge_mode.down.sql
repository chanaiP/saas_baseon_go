UPDATE dict_item di
SET enabled = TRUE,
    label = '买断制',
    updated_at = NOW()
FROM dict_type dt
WHERE di.dict_type_id = dt.id
  AND dt.code = 'app_charge_mode'
  AND di.value = 'BUYOUT';
