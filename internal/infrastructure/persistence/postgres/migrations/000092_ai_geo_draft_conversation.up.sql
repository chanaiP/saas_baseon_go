ALTER TABLE ai_geo_drafts
  ADD COLUMN IF NOT EXISTS conversation JSONB NOT NULL DEFAULT '[]'::jsonb;
