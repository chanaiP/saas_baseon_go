WITH ai_geo_defaults(plan_code, quota_code) AS (
    VALUES
        ('TRIAL', 'ai_geo_brand_count'),
        ('TRIAL', 'ai_geo_product_count'),
        ('TRIAL', 'ai_geo_monthly_draft_generations'),
        ('TRIAL', 'ai_geo_monthly_publish_tasks'),
        ('TRIAL', 'ai_geo_channel_account_count'),
        ('BASIC', 'ai_geo_brand_count'),
        ('BASIC', 'ai_geo_product_count'),
        ('BASIC', 'ai_geo_monthly_draft_generations'),
        ('BASIC', 'ai_geo_monthly_publish_tasks'),
        ('BASIC', 'ai_geo_channel_account_count'),
        ('PRO', 'ai_geo_brand_count'),
        ('PRO', 'ai_geo_product_count'),
        ('PRO', 'ai_geo_monthly_draft_generations'),
        ('PRO', 'ai_geo_monthly_publish_tasks'),
        ('PRO', 'ai_geo_channel_account_count'),
        ('ENTERPRISE', 'ai_geo_brand_count'),
        ('ENTERPRISE', 'ai_geo_product_count'),
        ('ENTERPRISE', 'ai_geo_monthly_draft_generations'),
        ('ENTERPRISE', 'ai_geo_monthly_publish_tasks'),
        ('ENTERPRISE', 'ai_geo_channel_account_count')
)
DELETE FROM saas_plan_quota pq
USING ai_geo_defaults d, saas_plan p, saas_quota q
WHERE pq.plan_id = p.id
  AND pq.quota_id = q.id
  AND p.plan_code = d.plan_code
  AND q.quota_code = d.quota_code;
