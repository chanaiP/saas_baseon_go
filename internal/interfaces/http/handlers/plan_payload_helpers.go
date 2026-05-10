package handlers

import (
	"strings"
)

type tenantPackagePayload struct {
	PlanID             uint64  `json:"plan_id"`
	SubscriptionStatus string  `json:"subscription_status"`
	StartTime          string  `json:"start_time"`
	EndTime            *string `json:"end_time"`
	TrialEndTime       *string `json:"trial_end_time"`
	AutoRenew          bool    `json:"auto_renew"`
	FrozenReason       *string `json:"frozen_reason"`
	Quotas             []struct {
		QuotaID    uint64 `json:"quota_id"`
		QuotaValue int    `json:"quota_value"`
	} `json:"quotas"`
}

type planUpdatePayload struct {
	PlanCode     *string  `json:"plan_code"`
	PlanName     *string  `json:"plan_name"`
	PlanType     *string  `json:"plan_type"`
	BillingCycle *string  `json:"billing_cycle"`
	Price        *float64 `json:"price"`
	Status       *int     `json:"status"`
	IsDefault    *bool    `json:"is_default"`
	SortOrder    *int     `json:"sort_order"`
	Description  *string  `json:"description"`
}

func (p planUpdatePayload) Updates() map[string]interface{} {
	updates := map[string]interface{}{}
	if p.PlanCode != nil {
		updates["plan_code"] = strings.TrimSpace(*p.PlanCode)
	}
	if p.PlanName != nil {
		updates["plan_name"] = strings.TrimSpace(*p.PlanName)
	}
	if p.PlanType != nil {
		updates["plan_type"] = strings.TrimSpace(*p.PlanType)
	}
	if p.BillingCycle != nil {
		updates["billing_cycle"] = strings.TrimSpace(*p.BillingCycle)
	}
	if p.Price != nil {
		updates["price"] = *p.Price
	}
	if p.Status != nil {
		updates["status"] = *p.Status
	}
	if p.IsDefault != nil {
		updates["is_default"] = *p.IsDefault
	}
	if p.SortOrder != nil {
		updates["sort_order"] = *p.SortOrder
	}
	if p.Description != nil {
		updates["description"] = nullableTrimmed(p.Description)
	}
	return updates
}

type featureUpdatePayload struct {
	FeatureCode  *string `json:"feature_code"`
	FeatureName  *string `json:"feature_name"`
	FeatureType  *string `json:"feature_type"`
	ParentID     *uint64 `json:"parent_id"`
	MenuID       *uint64 `json:"menu_id"`
	APIMethod    *string `json:"api_method"`
	APIPath      *string `json:"api_path"`
	ServiceKey   *string `json:"service_key"`
	SortOrder    *int    `json:"sort_order"`
	Status       *int    `json:"status"`
	Description  *string `json:"description"`
	PlatformOnly *bool   `json:"platform_only"`
}

func (p featureUpdatePayload) Updates() map[string]interface{} {
	updates := map[string]interface{}{}
	if p.FeatureCode != nil {
		updates["feature_code"] = strings.TrimSpace(*p.FeatureCode)
	}
	if p.FeatureName != nil {
		updates["feature_name"] = strings.TrimSpace(*p.FeatureName)
	}
	if p.FeatureType != nil {
		updates["feature_type"] = strings.TrimSpace(*p.FeatureType)
	}
	if p.ParentID != nil {
		updates["parent_id"] = *p.ParentID
	}
	if p.MenuID != nil {
		updates["menu_id"] = *p.MenuID
	}
	if p.APIMethod != nil {
		updates["api_method"] = nullableTrimmed(p.APIMethod)
	}
	if p.APIPath != nil {
		updates["api_path"] = nullableTrimmed(p.APIPath)
	}
	if p.ServiceKey != nil {
		updates["service_key"] = nullableTrimmed(p.ServiceKey)
	}
	if p.SortOrder != nil {
		updates["sort_order"] = *p.SortOrder
	}
	if p.Status != nil {
		updates["status"] = *p.Status
	}
	if p.Description != nil {
		updates["description"] = nullableTrimmed(p.Description)
	}
	if p.PlatformOnly != nil {
		updates["platform_only"] = *p.PlatformOnly
	}
	return updates
}

type quotaUpdatePayload struct {
	QuotaCode    *string `json:"quota_code"`
	QuotaName    *string `json:"quota_name"`
	QuotaUnit    *string `json:"quota_unit"`
	DefaultValue *int    `json:"default_value"`
	SortOrder    *int    `json:"sort_order"`
	Status       *int    `json:"status"`
	Description  *string `json:"description"`
}

func (p quotaUpdatePayload) Updates() map[string]interface{} {
	updates := map[string]interface{}{}
	if p.QuotaCode != nil {
		updates["quota_code"] = strings.TrimSpace(*p.QuotaCode)
	}
	if p.QuotaName != nil {
		updates["quota_name"] = strings.TrimSpace(*p.QuotaName)
	}
	if p.QuotaUnit != nil {
		updates["quota_unit"] = strings.TrimSpace(*p.QuotaUnit)
	}
	if p.DefaultValue != nil {
		updates["default_value"] = *p.DefaultValue
	}
	if p.SortOrder != nil {
		updates["sort_order"] = *p.SortOrder
	}
	if p.Status != nil {
		updates["status"] = *p.Status
	}
	if p.Description != nil {
		updates["description"] = nullableTrimmed(p.Description)
	}
	return updates
}

type planQuotaInput struct {
	QuotaID    uint64 `json:"quota_id"`
	QuotaValue int    `json:"quota_value"`
}
