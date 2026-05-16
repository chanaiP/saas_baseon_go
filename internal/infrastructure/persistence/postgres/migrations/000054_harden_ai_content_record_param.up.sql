UPDATE sys_param
SET
  remark = 'AI Gateway 内容记录级别：0 不记录，1 记录概要，2 记录脱敏内容，3 记录完整内容',
  value_type = 'number',
  tenant_editable = FALSE,
  is_platform_only = TRUE,
  updated_at = CURRENT_TIMESTAMP
WHERE param_key = 'ai.gateway.content_record_level';
