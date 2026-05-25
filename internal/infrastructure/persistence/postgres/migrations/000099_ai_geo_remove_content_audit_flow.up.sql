ALTER TABLE ai_geo_drafts
    ALTER COLUMN audit_status SET DEFAULT 'approved';

ALTER TABLE ai_geo_channel_contents
    ALTER COLUMN audit_status SET DEFAULT 'approved';

UPDATE ai_geo_drafts
SET audit_status = 'approved',
    updated_at = now()
WHERE deleted_at IS NULL
  AND audit_status <> 'approved';

UPDATE ai_geo_channel_contents
SET audit_status = 'approved',
    updated_at = now()
WHERE deleted_at IS NULL
  AND audit_status <> 'approved';
