UPDATE app_user u
SET
  company_id = (
    SELECT id
    FROM org_node
    WHERE tenant_id = u.tenant_id
      AND node_type = 'company'
      AND parent_id IS NULL
      AND deleted_at IS NULL
    ORDER BY id ASC
    LIMIT 1
  ),
  updated_at = now()
WHERE u.deleted_at IS NULL
  AND u.company_id IS NULL
  AND EXISTS (
    SELECT 1
    FROM org_node
    WHERE tenant_id = u.tenant_id
      AND node_type = 'company'
      AND parent_id IS NULL
      AND deleted_at IS NULL
  );
