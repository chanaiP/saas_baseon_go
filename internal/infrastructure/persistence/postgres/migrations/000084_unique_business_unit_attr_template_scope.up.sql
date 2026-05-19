WITH ranked_templates AS (
  SELECT
    id,
    ROW_NUMBER() OVER (
      PARTITION BY COALESCE(tenant_id, 0), template_name, unit_type_code, COALESCE(unit_group_code, '')
      ORDER BY created_at ASC, id ASC
    ) AS rn
  FROM business_unit_attr_template
  WHERE deleted_at IS NULL
)
UPDATE business_unit_attr_template t
SET
  deleted_at = NOW(),
  updated_at = NOW(),
  status = 'archived'
FROM ranked_templates r
WHERE t.id = r.id
  AND r.rn > 1;

CREATE UNIQUE INDEX IF NOT EXISTS uq_business_unit_attr_template_scope
  ON business_unit_attr_template (
    COALESCE(tenant_id, 0),
    template_name,
    unit_type_code,
    COALESCE(unit_group_code, '')
  )
  WHERE deleted_at IS NULL;
