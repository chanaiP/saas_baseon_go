WITH active_tenants AS (
    SELECT id AS tenant_id
    FROM tenant
    WHERE status = 1 AND deleted_at IS NULL
),
channel_seed AS (
    SELECT *
    FROM (VALUES
        ('website', '独立站', '自有站点', NULL::text, '["SEO文章","商品详情页","落地页"]'::jsonb, '["渠道 API","人工发布"]'::jsonb, 'api_auto'),
        ('xiaohongshu', '小红书', '社交种草平台', 'https://www.xiaohongshu.com'::text, '["图文笔记","视频笔记"]'::jsonb, '["渠道 API","Agent 执行","人工发布"]'::jsonb, 'api_draft_manual_confirm'),
        ('douyin', '抖音', '短视频平台', 'https://www.douyin.com'::text, '["短视频","图文","直播预告"]'::jsonb, '["渠道 API","Agent 执行","人工发布"]'::jsonb, 'agent_manual_confirm'),
        ('wechat_official_account', '微信公众号', '内容平台', 'https://mp.weixin.qq.com'::text, '["长图文","多图文","草稿"]'::jsonb, '["渠道 API","人工发布"]'::jsonb, 'api_draft_manual_confirm'),
        ('weibo', '微博', '社交媒体平台', 'https://weibo.com'::text, '["短图文","话题","长文"]'::jsonb, '["渠道 API","人工发布"]'::jsonb, 'api_draft_manual_confirm'),
        ('baijiahao', '百家号', '内容平台', 'https://baijiahao.baidu.com'::text, '["图文文章","动态","视频"]'::jsonb, '["渠道 API","人工发布"]'::jsonb, 'api_draft_manual_confirm'),
        ('zhihu', '知乎', '问答平台', 'https://www.zhihu.com'::text, '["回答","文章","想法"]'::jsonb, '["渠道 API","Agent 执行","人工发布"]'::jsonb, 'agent_manual_confirm')
    ) AS seed(channel_code, channel_name, channel_type, entry_url, content_forms, support_modes, default_publish_mode)
)
INSERT INTO ai_geo_channel_profiles (
    tenant_id,
    channel_code,
    channel_name,
    channel_type,
    entry_url,
    content_forms,
    support_modes,
    default_publish_mode,
    status,
    created_at,
    updated_at
)
SELECT
    t.tenant_id,
    s.channel_code,
    s.channel_name,
    s.channel_type,
    s.entry_url,
    s.content_forms,
    s.support_modes,
    s.default_publish_mode,
    'active',
    now(),
    now()
FROM active_tenants t
CROSS JOIN channel_seed s
ON CONFLICT (tenant_id, channel_code) WHERE deleted_at IS NULL DO NOTHING;
