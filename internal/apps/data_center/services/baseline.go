package services

import (
	"context"
	"time"

	"saas_baseon_go/internal/apps/data_center/dto"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type baselineMetric struct {
	Code     string
	Name     string
	Category string
	Formula  string
	Period   string
	Source   string
	Anomaly  bool
}

type baselineRule struct {
	Code       string
	Name       string
	Domain     string
	ObjectType string
	Conditions string
	OwnerRole  string
	Review     string
}

var baselineMetrics = []baselineMetric{
	{"gmv", "GMV", "销售指标", "已支付订单销售金额之和", "日 / 周 / 月", "销售标准数据", true},
	{"net_sales", "净销售额", "销售指标", "实付金额 - 退款金额", "日 / 周 / 月", "销售标准数据", true},
	{"order_count", "订单数", "销售指标", "订单数量", "日 / 周 / 月", "销售标准数据", true},
	{"avg_order_value", "客单价", "销售指标", "净销售额 / 订单数", "日 / 周 / 月", "销售标准数据", true},
	{"refund_rate", "退款率", "退款指标", "退款金额 / 销售金额", "日 / 周 / 月", "销售标准数据", true},
	{"ad_cost", "广告消耗", "投流指标", "广告消耗金额", "日 / 周", "投流标准数据", true},
	{"ad_roi", "ROI", "投流指标", "广告成交金额 / 广告消耗", "日 / 周", "投流标准数据", true},
	{"click_rate", "点击率", "投流指标", "点击数 / 曝光数", "日 / 周", "投流标准数据", true},
	{"conversion_rate", "转化率", "投流指标", "订单数 / 点击数", "日 / 周", "投流标准数据", true},
	{"available_days", "可售天数", "库存指标", "可售库存 / 近 7 日日均销量", "日", "库存标准数据", true},
}

var baselineRules = []baselineRule{
	{
		Code: "gmv_drop", Name: "GMV 下滑", Domain: "销售异常", ObjectType: "brand",
		Conditions: `{"relation":"AND","conditions":[{"metric_code":"gmv","compare_type":"mom_rate","operator":"<","value":-15,"unit":"%"},{"metric_code":"order_count","compare_type":"mom_rate","operator":"<","value":-10,"unit":"%"}]}`,
		OwnerRole:  "品牌负责人", Review: `["gmv","order_count","avg_order_value"]`,
	},
	{
		Code: "ad_spend_up_roi_down", Name: "投流增加但 ROI 下降", Domain: "投流异常", ObjectType: "campaign",
		Conditions: `{"relation":"AND","conditions":[{"metric_code":"ad_cost","compare_type":"mom_rate","operator":">","value":20,"unit":"%"},{"metric_code":"ad_roi","compare_type":"mom_rate","operator":"<","value":-15,"unit":"%"}]}`,
		OwnerRole:  "投流负责人", Review: `["ad_roi","ad_cost","gmv","conversion_rate"]`,
	},
	{
		Code: "inventory_enough_sales_down", Name: "库存充足但销量下降", Domain: "库存异常", ObjectType: "sku",
		Conditions: `{"relation":"AND","conditions":[{"metric_code":"available_days","compare_type":"value","operator":">","value":30,"unit":"天"}]}`,
		OwnerRole:  "商品运营负责人", Review: `["available_days","gmv"]`,
	},
}

func (s *Service) EnsureTenantBaseline(ctx context.Context, viewer dto.Viewer) error {
	now := time.Now()
	for _, item := range baselineMetrics {
		row := models.DataCenterMetricDefinition{
			TenantID: viewer.TenantID, MetricCode: item.Code, MetricName: item.Name, MetricCategory: item.Category,
			Formula: &item.Formula, StatisticPeriod: &item.Period, Dimensions: "[]", DataSource: &item.Source,
			Enabled: true, AnomalyEnabled: item.Anomaly, CreatedBy: &viewer.UserID, UpdatedBy: &viewer.UserID, CreatedAt: now, UpdatedAt: now,
		}
		var count int64
		if err := s.repo.DB().WithContext(ctx).Model(&models.DataCenterMetricDefinition{}).Where("tenant_id = ? AND metric_code = ? AND deleted_at IS NULL", viewer.TenantID, item.Code).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := s.repo.DB().WithContext(ctx).Create(&row).Error; err != nil {
			return err
		}
	}
	for _, item := range baselineRules {
		row := models.DataCenterAnomalyRule{
			TenantID: viewer.TenantID, RuleCode: item.Code, RuleName: item.Name, BusinessDomain: item.Domain, TargetObjectType: item.ObjectType,
			ScopeJSON: "{}", MetricConditionsJSON: item.Conditions, LevelConfigJSON: "{}", ConfidenceConfigJSON: "{}",
			AIEnabled: true, TaskEnabled: true, AutoTaskConfidenceThreshold: 80, DefaultOwnerRole: &item.OwnerRole,
			DefaultDeadlineDays: 3, ReviewMetricCodes: item.Review, ReviewAfterDays: 3, Priority: 100, Enabled: true,
			CreatedBy: &viewer.UserID, UpdatedBy: &viewer.UserID, CreatedAt: now, UpdatedAt: now,
		}
		var count int64
		if err := s.repo.DB().WithContext(ctx).Model(&models.DataCenterAnomalyRule{}).Where("tenant_id = ? AND rule_code = ? AND deleted_at IS NULL", viewer.TenantID, item.Code).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := s.repo.DB().WithContext(ctx).Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}
