ALTER TABLE ai_geo_drafts
    ALTER COLUMN audit_status SET DEFAULT 'draft';

ALTER TABLE ai_geo_channel_contents
    ALTER COLUMN audit_status SET DEFAULT 'pending';
