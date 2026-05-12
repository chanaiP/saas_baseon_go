UPDATE sys_app
SET charge_mode = 'SUBSCRIPTION',
    updated_at = NOW()
WHERE charge_mode = 'BUYOUT';

UPDATE dict_item di
SET enabled = FALSE,
    label = '订阅制',
    updated_at = NOW()
FROM dict_type dt
WHERE di.dict_type_id = dt.id
  AND dt.code = 'app_charge_mode'
  AND di.value = 'BUYOUT';
