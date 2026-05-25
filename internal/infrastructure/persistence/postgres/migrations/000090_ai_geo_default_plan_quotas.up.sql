WITH ai_geo_defaults(plan_code, quota_code, quota_value) AS (
    VALUES
        ('TRIAL', 'ai_geo_brand_count', 2),
        ('TRIAL', 'ai_geo_product_count', 10),
        ('TRIAL', 'ai_geo_monthly_draft_generations', 20),
        ('TRIAL', 'ai_geo_monthly_publish_tasks', 10),
        ('TRIAL', 'ai_geo_channel_account_count', 2),
        ('BASIC', 'ai_geo_brand_count', 5),
        ('BASIC', 'ai_geo_product_count', 50),
        ('BASIC', 'ai_geo_monthly_draft_generations', 100),
        ('BASIC', 'ai_geo_monthly_publish_tasks', 50),
        ('BASIC', 'ai_geo_channel_account_count', 5),
        ('PRO', 'ai_geo_brand_count', 20),
        ('PRO', 'ai_geo_product_count', 300),
        ('PRO', 'ai_geo_monthly_draft_generations', 800),
        ('PRO', 'ai_geo_monthly_publish_tasks', 300),
        ('PRO', 'ai_geo_channel_account_count', 20),
        ('ENTERPRISE', 'ai_geo_brand_count', -1),
        ('ENTERPRISE', 'ai_geo_product_count', -1),
        ('ENTERPRISE', 'ai_geo_monthly_draft_generations', -1),
        ('ENTERPRISE', 'ai_geo_monthly_publish_tasks', -1),
        ('ENTERPRISE', 'ai_geo_channel_account_count', -1),
        ('PERSONAL_FREE', 'ai_geo_brand_count', 1),
        ('PERSONAL_FREE', 'ai_geo_product_count', 5),
        ('PERSONAL_FREE', 'ai_geo_monthly_draft_generations', 10),
        ('PERSONAL_FREE', 'ai_geo_monthly_publish_tasks', 5),
        ('PERSONAL_FREE', 'ai_geo_channel_account_count', 1),
        ('PERSONAL_BASIC', 'ai_geo_brand_count', 2),
        ('PERSONAL_BASIC', 'ai_geo_product_count', 20),
        ('PERSONAL_BASIC', 'ai_geo_monthly_draft_generations', 50),
        ('PERSONAL_BASIC', 'ai_geo_monthly_publish_tasks', 20),
        ('PERSONAL_BASIC', 'ai_geo_channel_account_count', 3)
)
INSERT INTO saas_plan_quota (plan_id, quota_id, quota_value, created_at, updated_at)
SELECT p.id, q.id, d.quota_value, now(), now()
FROM ai_geo_defaults d
JOIN saas_plan p ON p.plan_code = d.plan_code
JOIN saas_quota q ON q.quota_code = d.quota_code
ON CONFLICT (plan_id, quota_id) DO NOTHING;
