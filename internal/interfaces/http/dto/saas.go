package dto

import "time"

type PlanResponse struct {
	ID           uint64    `json:"id"`
	PlanCode     string    `json:"plan_code"`
	PlanName     string    `json:"plan_name"`
	PlanType     string    `json:"plan_type"`
	BillingCycle string    `json:"billing_cycle"`
	Price        float64   `json:"price"`
	Status       int       `json:"status"`
	IsDefault    bool      `json:"is_default"`
	SortOrder    int       `json:"sort_order"`
	Description  *string   `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type FeatureResponse struct {
	ID          uint64  `json:"id"`
	FeatureCode string  `json:"feature_code"`
	FeatureName string  `json:"feature_name"`
	FeatureType string  `json:"feature_type"`
	ParentID    uint64  `json:"parent_id"`
	MenuID      *uint64 `json:"menu_id"`
	APIMethod   *string `json:"api_method"`
	APIPath     *string `json:"api_path"`
	ServiceKey  *string `json:"service_key"`
	Status      int     `json:"status"`
	Description *string `json:"description"`
}

type QuotaResponse struct {
	ID          uint64  `json:"id"`
	QuotaCode   string  `json:"quota_code"`
	QuotaName   string  `json:"quota_name"`
	QuotaType   string  `json:"quota_type"`
	PeriodType  *string `json:"period_type"`
	Unit        *string `json:"unit"`
	Status      int     `json:"status"`
	Description *string `json:"description"`
}

type PlanFeatureIDsResponse struct {
	PlanID     uint64   `json:"plan_id"`
	FeatureIDs []uint64 `json:"feature_ids"`
}

type PlanQuotaItemResponse struct {
	QuotaID    uint64  `json:"quota_id"`
	QuotaCode  string  `json:"quota_code"`
	QuotaName  string  `json:"quota_name"`
	QuotaValue int     `json:"quota_value"`
	PeriodType *string `json:"period_type"`
	Unit       *string `json:"unit"`
}

type PlanQuotasResponse struct {
	PlanID uint64                  `json:"plan_id"`
	Quotas []PlanQuotaItemResponse `json:"quotas"`
}

type PlanMatrixQuotaValue struct {
	QuotaID    uint64 `json:"quota_id"`
	QuotaCode  string `json:"quota_code"`
	QuotaName  string `json:"quota_name"`
	QuotaValue int    `json:"quota_value"`
}

type PlanMatrixCell struct {
	PlanID      uint64                 `json:"plan_id"`
	PlanCode    string                 `json:"plan_code"`
	Enabled     bool                   `json:"enabled"`
	State       string                 `json:"state"`
	FeatureIDs  []uint64               `json:"feature_ids"`
	QuotaValues []PlanMatrixQuotaValue `json:"quota_values"`
}

type PlanMatrixNode struct {
	ID          string           `json:"id"`
	Label       string           `json:"label"`
	NodeType    string           `json:"node_type"`
	FeatureID   uint64           `json:"feature_id"`
	FeatureCode string           `json:"feature_code"`
	FeatureType string           `json:"feature_type"`
	Description *string          `json:"description"`
	Children    []PlanMatrixNode `json:"children"`
	Cells       []PlanMatrixCell `json:"cells"`
}

type PlanMatrixResponse struct {
	Plans []PlanResponse   `json:"plans"`
	Nodes []PlanMatrixNode `json:"nodes"`
}

type PlanCapabilityQuotaInputResponse struct {
	QuotaID    uint64 `json:"quota_id"`
	QuotaValue int    `json:"quota_value"`
}

type PlanCapabilitiesResponse struct {
	PlanID     uint64                             `json:"plan_id"`
	FeatureIDs []uint64                           `json:"feature_ids"`
	Quotas     []PlanCapabilityQuotaInputResponse `json:"quotas"`
}
