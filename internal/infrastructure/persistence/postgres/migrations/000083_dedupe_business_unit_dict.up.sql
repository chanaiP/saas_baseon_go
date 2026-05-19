WITH active_items AS (
  SELECT
    di.id,
    first_value(di.id) OVER (
      PARTITION BY di.dict_type_id, COALESCE(di.parent_id, 0), di.value
      ORDER BY di.sort_order ASC, di.id ASC
    ) AS keep_id
  FROM dict_item di
  JOIN dict_type dt ON dt.id = di.dict_type_id
  WHERE dt.code = 'business_unit'
    AND dt.deleted_at IS NULL
    AND di.deleted_at IS NULL
),
duplicates AS (
  SELECT id, keep_id
  FROM active_items
  WHERE id <> keep_id
)
UPDATE dict_item child
SET parent_id = duplicates.keep_id,
    updated_at = now()
FROM duplicates
WHERE child.parent_id = duplicates.id
  AND child.deleted_at IS NULL;

WITH active_items AS (
  SELECT
    di.id,
    row_number() OVER (
      PARTITION BY di.dict_type_id, COALESCE(di.parent_id, 0), di.value
      ORDER BY di.sort_order ASC, di.id ASC
    ) AS rn
  FROM dict_item di
  JOIN dict_type dt ON dt.id = di.dict_type_id
  WHERE dt.code = 'business_unit'
    AND dt.deleted_at IS NULL
    AND di.deleted_at IS NULL
)
UPDATE dict_item di
SET deleted_at = now(),
    updated_at = now(),
    enabled = false
FROM active_items ai
WHERE di.id = ai.id
  AND ai.rn > 1;
