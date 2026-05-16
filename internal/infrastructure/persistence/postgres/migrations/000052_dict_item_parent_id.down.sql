DROP INDEX IF EXISTS idx_dict_item_parent_id;

ALTER TABLE dict_item
  DROP COLUMN IF EXISTS parent_id;
