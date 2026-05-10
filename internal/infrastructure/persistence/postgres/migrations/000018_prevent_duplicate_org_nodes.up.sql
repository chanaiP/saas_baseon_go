WITH duplicate_code_nodes AS (
    SELECT id,
           row_number() OVER (
               PARTITION BY tenant_id, COALESCE(parent_id, 0), node_type, lower(btrim(code))
               ORDER BY id
           ) AS rn
    FROM org_node
    WHERE deleted_at IS NULL
      AND code IS NOT NULL
      AND btrim(code) <> ''
),
duplicate_name_nodes AS (
    SELECT id,
           row_number() OVER (
               PARTITION BY tenant_id, COALESCE(parent_id, 0), node_type, lower(btrim(name))
               ORDER BY id
           ) AS rn
    FROM org_node
    WHERE deleted_at IS NULL
      AND btrim(name) <> ''
)
UPDATE org_node
SET deleted_at = now(),
    status = 0,
    updated_at = now()
WHERE id IN (
    SELECT id FROM duplicate_code_nodes WHERE rn > 1
    UNION
    SELECT id FROM duplicate_name_nodes WHERE rn > 1
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_org_node_unique_active_code
    ON org_node (tenant_id, COALESCE(parent_id, 0), node_type, lower(btrim(code)))
    WHERE deleted_at IS NULL
      AND code IS NOT NULL
      AND btrim(code) <> '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_org_node_unique_active_name
    ON org_node (tenant_id, COALESCE(parent_id, 0), node_type, lower(btrim(name)))
    WHERE deleted_at IS NULL
      AND btrim(name) <> '';
