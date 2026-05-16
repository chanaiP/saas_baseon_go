UPDATE sys_param
SET
  tenant_editable = TRUE,
  is_platform_only = FALSE,
  updated_at = CURRENT_TIMESTAMP
WHERE param_key = 'ai.gateway.content_record_level';
