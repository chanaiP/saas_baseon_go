ALTER TABLE dict_item
  ADD COLUMN IF NOT EXISTS parent_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_dict_item_parent_id
  ON dict_item (parent_id);
