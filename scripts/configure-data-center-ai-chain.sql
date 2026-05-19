-- Idempotent validation setup for Ai经营决策中心 real anomaly analysis.
-- This script does not contain provider credentials. Provider accounts and API keys
-- must already be configured in AI 能力中心.

DO $$
DECLARE
  v_policy_id uuid;
  v_route_id uuid;
  v_scenario_name varchar(150);
BEGIN
  SELECT default_base_route_id, ai_scenario_name
    INTO v_route_id, v_scenario_name
  FROM ai_scenarios
  WHERE app_code = 'data-center'
    AND ai_scenario_code = 'anomaly_analysis'
    AND status = 'active'
    AND deleted_at IS NULL
  LIMIT 1;

  IF v_route_id IS NULL THEN
    RAISE EXCEPTION 'Missing active data-center/anomaly_analysis scenario or default route';
  END IF;

  SELECT id
    INTO v_policy_id
  FROM ai_tenant_strategy_policies
  WHERE app_code = 'data-center'
    AND ai_scenario_code = 'anomaly_analysis'
    AND tenant_ids @> '["1"]'::jsonb
    AND deleted_at IS NULL
  ORDER BY updated_at DESC
  LIMIT 1;

  IF v_policy_id IS NULL THEN
    INSERT INTO ai_tenant_strategy_policies (
      policy_name, tenant_scope, tenant_ids, app_code, app_name,
      ai_scenario_code, ai_scenario_name, default_base_route_id,
      description, status, created_at, updated_at
    ) VALUES (
      '租户1 Ai经营异常诊断策略', 'include', '["1"]'::jsonb,
      'data-center', 'Ai经营决策中心',
      'anomaly_analysis', v_scenario_name, v_route_id,
      'Ai经营决策中心真实模型链路验收策略', 'active', now(), now()
    )
    RETURNING id INTO v_policy_id;
  ELSE
    UPDATE ai_tenant_strategy_policies
    SET default_base_route_id = v_route_id,
        status = 'active',
        updated_at = now()
    WHERE id = v_policy_id;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM ai_strategy_quota_rules
    WHERE policy_id = v_policy_id
      AND dimension = 'scenario'
      AND subject_code = 'anomaly_analysis'
      AND deleted_at IS NULL
  ) THEN
    INSERT INTO ai_strategy_quota_rules (
      policy_id, dimension, subject_code, usage_unit, period,
      quota_limit, warning_threshold, over_limit_action,
      status, created_at, updated_at
    ) VALUES (
      v_policy_id, 'scenario', 'anomaly_analysis', 'tokens', 'day',
      500000, 80, 'alert_only', 'active', now(), now()
    );
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM ai_strategy_rate_limit_rules
    WHERE policy_id = v_policy_id
      AND dimension = 'scenario'
      AND subject_code = 'anomaly_analysis'
      AND deleted_at IS NULL
  ) THEN
    INSERT INTO ai_strategy_rate_limit_rules (
      policy_id, dimension, subject_code, qps, concurrency,
      minute_limit, hour_limit, day_limit, over_limit_action,
      status, created_at, updated_at
    ) VALUES (
      v_policy_id, 'scenario', 'anomaly_analysis', 1, 2,
      10, 100, 300, 'queue', 'active', now(), now()
    );
  END IF;
END $$;

INSERT INTO tenant_quota_override (
  tenant_id, quota_id, quota_value, reason,
  start_time, end_time, source, source_ref, created_at, updated_at
)
SELECT
  1, q.id, 50,
  'Ai经营决策中心真实模型链路验收：允许租户1执行每日 AI 分析。',
  now() - interval '1 day',
  now() + interval '30 days',
  'MANUAL',
  'data-center-ai-chain-20260519',
  now(),
  now()
FROM saas_quota q
WHERE q.quota_code = 'data_center_daily_ai_analyses'
ON CONFLICT (tenant_id, quota_id) DO UPDATE
SET quota_value = EXCLUDED.quota_value,
    reason = EXCLUDED.reason,
    start_time = EXCLUDED.start_time,
    end_time = EXCLUDED.end_time,
    source = EXCLUDED.source,
    source_ref = EXCLUDED.source_ref,
    updated_at = now();
