-- Keep usage_amount in the raw business unit for display and quota analysis.
-- Pricing policies may bill by 1K/1M tokens, but usage_unit should still be tokens.
UPDATE ai_usage_records
SET usage_unit = 'tokens'
WHERE lower(replace(replace(usage_unit, ' ', ''), '_', '')) IN ('1ktokens', '1mtokens')
  AND lower(replace(replace(platform_unit, ' ', ''), '_', '')) IN ('1ktokens', '1mtokens');
