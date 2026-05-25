DELETE FROM saas_plan_quota pq
USING saas_plan p, saas_quota q
WHERE pq.plan_id = p.id
  AND pq.quota_id = q.id
  AND p.plan_code IN ('TRIAL', 'BASIC', 'PRO', 'ENTERPRISE', 'PERSONAL_FREE', 'PERSONAL_BASIC')
  AND q.quota_code IN (
      'ai_geo_brand_count',
      'ai_geo_product_count',
      'ai_geo_monthly_draft_generations',
      'ai_geo_monthly_publish_tasks',
      'ai_geo_channel_account_count'
  );
