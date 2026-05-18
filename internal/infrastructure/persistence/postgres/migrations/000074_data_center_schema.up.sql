CREATE TABLE IF NOT EXISTS data_center_raw_data_batches (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    company_id bigint,
    department_id bigint,
    batch_code varchar(80) NOT NULL,
    data_type varchar(50) NOT NULL,
    platform_code varchar(80),
    app_code varchar(80),
    connection_code varchar(120),
    record_count bigint NOT NULL DEFAULT 0,
    success_count bigint NOT NULL DEFAULT 0,
    failed_count bigint NOT NULL DEFAULT 0,
    sync_time timestamp with time zone,
    status varchar(32) NOT NULL DEFAULT 'pending',
    error_message text,
    source_params jsonb NOT NULL DEFAULT '{}'::jsonb,
    sample_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_by bigint,
    updated_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    CONSTRAINT data_center_raw_data_batches_status_chk CHECK (status IN ('pending','processing','success','warning','failed')),
    CONSTRAINT data_center_raw_data_batches_count_chk CHECK (record_count >= 0 AND success_count >= 0 AND failed_count >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_raw_batch_tenant_code ON data_center_raw_data_batches(tenant_id, batch_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dc_raw_batch_list ON data_center_raw_data_batches(tenant_id, data_type, status, sync_time DESC) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS data_center_raw_data_errors (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    batch_id bigint NOT NULL REFERENCES data_center_raw_data_batches(id),
    batch_code varchar(80) NOT NULL,
    row_number bigint,
    error_code varchar(80) NOT NULL,
    error_reason text NOT NULL,
    raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_dc_raw_error_batch ON data_center_raw_data_errors(tenant_id, batch_id, id);

CREATE TABLE IF NOT EXISTS data_center_std_sales_orders (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    company_id bigint,
    department_id bigint,
    order_code varchar(128) NOT NULL,
    brand_code varchar(80),
    brand_name varchar(120),
    channel_code varchar(80),
    platform_code varchar(80),
    resource_code varchar(100),
    resource_name varchar(160),
    product_code varchar(100),
    product_name varchar(180),
    sku_code varchar(100),
    sales_amount numeric(18,2) NOT NULL DEFAULT 0,
    paid_amount numeric(18,2) NOT NULL DEFAULT 0,
    refund_amount numeric(18,2) NOT NULL DEFAULT 0,
    order_status varchar(32),
    order_time timestamp with time zone NOT NULL,
    source_batch_code varchar(80),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    CONSTRAINT data_center_std_sales_orders_amount_chk CHECK (sales_amount >= 0 AND paid_amount >= 0 AND refund_amount >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_sales_order_tenant_code ON data_center_std_sales_orders(tenant_id, order_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dc_sales_order_list ON data_center_std_sales_orders(tenant_id, order_time DESC, brand_code, channel_code, platform_code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS data_center_std_ad_daily (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    company_id bigint,
    department_id bigint,
    brand_code varchar(80),
    brand_name varchar(120),
    platform_code varchar(80),
    account_code varchar(100) NOT NULL,
    account_name varchar(160),
    campaign_code varchar(100) NOT NULL,
    campaign_name varchar(180),
    creative_code varchar(100),
    product_code varchar(100),
    stat_date date NOT NULL,
    cost_amount numeric(18,2) NOT NULL DEFAULT 0,
    impression_count bigint NOT NULL DEFAULT 0,
    click_count bigint NOT NULL DEFAULT 0,
    order_amount numeric(18,2) NOT NULL DEFAULT 0,
    order_count bigint NOT NULL DEFAULT 0,
    roi numeric(18,4) NOT NULL DEFAULT 0,
    click_rate numeric(18,4) NOT NULL DEFAULT 0,
    conversion_rate numeric(18,4) NOT NULL DEFAULT 0,
    source_batch_code varchar(80),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    CONSTRAINT data_center_std_ad_daily_count_chk CHECK (cost_amount >= 0 AND impression_count >= 0 AND click_count >= 0 AND order_amount >= 0 AND order_count >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_ad_daily_object_date ON data_center_std_ad_daily(tenant_id, account_code, campaign_code, stat_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dc_ad_daily_list ON data_center_std_ad_daily(tenant_id, stat_date DESC, brand_code, platform_code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS data_center_std_inventory_daily (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    company_id bigint,
    department_id bigint,
    brand_code varchar(80),
    product_code varchar(100) NOT NULL,
    product_name varchar(180),
    sku_code varchar(100) NOT NULL,
    warehouse_code varchar(100),
    store_code varchar(100) NOT NULL DEFAULT '',
    stat_date date NOT NULL,
    available_stock bigint NOT NULL DEFAULT 0,
    in_transit_stock bigint NOT NULL DEFAULT 0,
    sales_7d bigint NOT NULL DEFAULT 0,
    available_days numeric(18,2) NOT NULL DEFAULT 0,
    inventory_status varchar(32),
    source_batch_code varchar(80),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    CONSTRAINT data_center_std_inventory_daily_count_chk CHECK (available_stock >= 0 AND in_transit_stock >= 0 AND sales_7d >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_inventory_object_date ON data_center_std_inventory_daily(tenant_id, product_code, sku_code, store_code, stat_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dc_inventory_list ON data_center_std_inventory_daily(tenant_id, stat_date DESC, brand_code, product_code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS data_center_std_refund_orders (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    company_id bigint,
    department_id bigint,
    refund_code varchar(128) NOT NULL,
    order_code varchar(128),
    brand_code varchar(80),
    brand_name varchar(120),
    channel_code varchar(80),
    platform_code varchar(80),
    product_code varchar(100),
    product_name varchar(180),
    sku_code varchar(100),
    refund_amount numeric(18,2) NOT NULL DEFAULT 0,
    refund_reason varchar(255),
    refund_status varchar(32),
    refund_time timestamp with time zone NOT NULL,
    source_batch_code varchar(80),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    CONSTRAINT data_center_std_refund_orders_amount_chk CHECK (refund_amount >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_refund_tenant_code ON data_center_std_refund_orders(tenant_id, refund_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dc_refund_list ON data_center_std_refund_orders(tenant_id, refund_time DESC, brand_code, channel_code, platform_code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS data_center_std_products (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    company_id bigint,
    department_id bigint,
    product_code varchar(100) NOT NULL,
    product_name varchar(180) NOT NULL,
    sku_code varchar(100),
    brand_code varchar(80),
    brand_name varchar(120),
    category_code varchar(100),
    category_name varchar(160),
    list_price numeric(18,2) NOT NULL DEFAULT 0,
    cost_price numeric(18,2) NOT NULL DEFAULT 0,
    product_status varchar(32) NOT NULL DEFAULT 'active',
    source_batch_code varchar(80),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    CONSTRAINT data_center_std_products_price_chk CHECK (list_price >= 0 AND cost_price >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_product_tenant_code ON data_center_std_products(tenant_id, product_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dc_product_list ON data_center_std_products(tenant_id, brand_code, category_code, product_status) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS data_center_std_store_sales_daily (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    company_id bigint,
    department_id bigint,
    brand_code varchar(80),
    brand_name varchar(120),
    channel_code varchar(80),
    platform_code varchar(80),
    store_code varchar(100) NOT NULL,
    store_name varchar(180),
    stat_date date NOT NULL,
    gmv numeric(18,2) NOT NULL DEFAULT 0,
    net_sales numeric(18,2) NOT NULL DEFAULT 0,
    order_count bigint NOT NULL DEFAULT 0,
    customer_count bigint NOT NULL DEFAULT 0,
    refund_amount numeric(18,2) NOT NULL DEFAULT 0,
    target_amount numeric(18,2),
    source_batch_code varchar(80),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    CONSTRAINT data_center_std_store_sales_daily_amount_chk CHECK (gmv >= 0 AND net_sales >= 0 AND order_count >= 0 AND customer_count >= 0 AND refund_amount >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_store_sales_unique ON data_center_std_store_sales_daily(tenant_id, store_code, stat_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dc_store_sales_list ON data_center_std_store_sales_daily(tenant_id, stat_date DESC, brand_code, channel_code, platform_code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS data_center_metric_definitions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    metric_code varchar(80) NOT NULL,
    metric_name varchar(120) NOT NULL,
    metric_category varchar(80) NOT NULL,
    formula text,
    statistic_period varchar(80),
    dimensions jsonb NOT NULL DEFAULT '[]'::jsonb,
    data_source varchar(160),
    enabled boolean NOT NULL DEFAULT true,
    anomaly_enabled boolean NOT NULL DEFAULT false,
    created_by bigint,
    updated_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_metric_tenant_code ON data_center_metric_definitions(tenant_id, metric_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dc_metric_list ON data_center_metric_definitions(tenant_id, metric_category, enabled) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS data_center_metric_results (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    metric_code varchar(80) NOT NULL,
    resource_type varchar(80) NOT NULL,
    resource_code varchar(120) NOT NULL,
    resource_name varchar(180),
    stat_date date NOT NULL,
    period_type varchar(32) NOT NULL,
    metric_value numeric(18,4) NOT NULL DEFAULT 0,
    compare_value numeric(18,4),
    compare_rate numeric(18,4),
    target_value numeric(18,4),
    dimension_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_metric_result_unique ON data_center_metric_results(tenant_id, metric_code, resource_type, resource_code, stat_date, period_type);
CREATE INDEX IF NOT EXISTS idx_dc_metric_result_query ON data_center_metric_results(tenant_id, metric_code, stat_date DESC, resource_type, resource_code);

CREATE TABLE IF NOT EXISTS data_center_anomaly_rules (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    rule_code varchar(80) NOT NULL,
    rule_name varchar(140) NOT NULL,
    business_domain varchar(80) NOT NULL,
    target_object_type varchar(80) NOT NULL,
    scope_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    metric_conditions_json jsonb NOT NULL,
    level_config_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    confidence_config_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    ai_enabled boolean NOT NULL DEFAULT true,
    task_enabled boolean NOT NULL DEFAULT true,
    auto_task_confidence_threshold integer NOT NULL DEFAULT 80,
    default_owner_role varchar(80),
    default_deadline_days integer NOT NULL DEFAULT 3,
    review_metric_codes jsonb NOT NULL DEFAULT '[]'::jsonb,
    review_after_days integer NOT NULL DEFAULT 3,
    priority integer NOT NULL DEFAULT 100,
    enabled boolean NOT NULL DEFAULT true,
    created_by bigint,
    updated_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    CONSTRAINT data_center_anomaly_rules_threshold_chk CHECK (auto_task_confidence_threshold >= 0 AND auto_task_confidence_threshold <= 100)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_rule_tenant_code ON data_center_anomaly_rules(tenant_id, rule_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dc_rule_list ON data_center_anomaly_rules(tenant_id, business_domain, enabled, priority) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS data_center_anomaly_records (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    anomaly_code varchar(80) NOT NULL UNIQUE,
    rule_code varchar(80) NOT NULL,
    title varchar(255) NOT NULL,
    business_domain varchar(80) NOT NULL,
    object_type varchar(80) NOT NULL,
    object_code varchar(120) NOT NULL,
    object_name varchar(180),
    stat_date date NOT NULL,
    anomaly_level varchar(32) NOT NULL,
    confidence_score integer NOT NULL DEFAULT 0,
    impact_amount numeric(18,2) NOT NULL DEFAULT 0,
    evidence_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    occurred_at timestamp with time zone NOT NULL,
    ai_status varchar(32) NOT NULL DEFAULT 'pending',
    task_status varchar(32) NOT NULL DEFAULT 'none',
    review_status varchar(32) NOT NULL DEFAULT 'none',
    status varchar(32) NOT NULL DEFAULT 'pending',
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    CONSTRAINT data_center_anomaly_records_level_chk CHECK (anomaly_level IN ('low','medium','high','critical','低','中','高','严重')),
    CONSTRAINT data_center_anomaly_records_confidence_chk CHECK (confidence_score >= 0 AND confidence_score <= 100)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_anomaly_dedupe ON data_center_anomaly_records(tenant_id, rule_code, object_type, object_code, stat_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dc_anomaly_list ON data_center_anomaly_records(tenant_id, status, anomaly_level, occurred_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS data_center_ai_diagnosis_records (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    anomaly_id bigint NOT NULL REFERENCES data_center_anomaly_records(id),
    anomaly_code varchar(80) NOT NULL,
    problem_summary text,
    impact_summary text,
    reason_analysis_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    evidence_summary_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    suggestion_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    confidence_explanation text,
    task_suggestion_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    model_code varchar(80),
    status varchar(32) NOT NULL DEFAULT 'success',
    error_message text,
    generated_at timestamp with time zone NOT NULL DEFAULT now(),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_dc_ai_diagnosis_anomaly ON data_center_ai_diagnosis_records(tenant_id, anomaly_id, generated_at DESC);

CREATE TABLE IF NOT EXISTS data_center_rectification_tasks (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    task_code varchar(80) NOT NULL,
    anomaly_id bigint REFERENCES data_center_anomaly_records(id),
    anomaly_code varchar(80),
    title varchar(255) NOT NULL,
    task_type varchar(80) NOT NULL DEFAULT 'anomaly_rectification',
    owner_user_id bigint,
    owner_role varchar(80),
    collaborator_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
    priority varchar(32) NOT NULL DEFAULT 'medium',
    deadline date,
    target_desc text,
    ai_suggestion_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    execution_feedback text,
    progress integer NOT NULL DEFAULT 0,
    status varchar(32) NOT NULL DEFAULT 'pending',
    review_status varchar(32) NOT NULL DEFAULT 'none',
    completed_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    CONSTRAINT data_center_rectification_tasks_progress_chk CHECK (progress >= 0 AND progress <= 100),
    CONSTRAINT data_center_rectification_tasks_status_chk CHECK (status IN ('pending','processing','completed','overdue','closed'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_task_tenant_code ON data_center_rectification_tasks(tenant_id, task_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_task_anomaly_once ON data_center_rectification_tasks(tenant_id, anomaly_id) WHERE anomaly_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dc_task_list ON data_center_rectification_tasks(tenant_id, status, priority, deadline) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS data_center_task_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    task_id bigint NOT NULL REFERENCES data_center_rectification_tasks(id),
    task_code varchar(80) NOT NULL,
    action varchar(60) NOT NULL,
    from_status varchar(32),
    to_status varchar(32),
    content text,
    operator_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_dc_task_log_task ON data_center_task_logs(tenant_id, task_id, created_at DESC);

CREATE TABLE IF NOT EXISTS data_center_rectification_reviews (
    id BIGSERIAL PRIMARY KEY,
    tenant_id bigint NOT NULL REFERENCES tenant(id),
    review_code varchar(80) NOT NULL,
    task_id bigint NOT NULL REFERENCES data_center_rectification_tasks(id),
    task_code varchar(80) NOT NULL,
    anomaly_id bigint REFERENCES data_center_anomaly_records(id),
    anomaly_code varchar(80),
    before_metric_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    after_metric_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    improvement_result varchar(128),
    review_conclusion varchar(64) NOT NULL DEFAULT 'weak',
    ai_review_summary text,
    manual_review_summary text,
    experience_summary text,
    reviewed_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    CONSTRAINT data_center_rectification_reviews_conclusion_chk CHECK (review_conclusion IN ('effective','weak','ineffective','follow_up','整改有效','效果不明显','整改无效','需继续跟进'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_review_tenant_code ON data_center_rectification_reviews(tenant_id, review_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_dc_review_task_once ON data_center_rectification_reviews(tenant_id, task_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dc_review_list ON data_center_rectification_reviews(tenant_id, review_conclusion, reviewed_at DESC) WHERE deleted_at IS NULL;
