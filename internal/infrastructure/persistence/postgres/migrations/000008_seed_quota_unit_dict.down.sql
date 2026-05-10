DELETE FROM dict_item
WHERE dict_type_id IN (
  SELECT id FROM dict_type WHERE code = 'quota_unit'
)
AND value IN ('COUNT', 'MB', 'GB', 'TIMES', 'ITEM');

DELETE FROM dict_type
WHERE code = 'quota_unit'
  AND NOT EXISTS (
    SELECT 1 FROM dict_item WHERE dict_item.dict_type_id = dict_type.id
  );
