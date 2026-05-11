package dto

import "time"

type AppListRequest struct {
	Skip    int
	Limit   int
	Keyword string
	Type    string
	Status  string
	Source  string
}

type AppCreateRequest struct {
	AppCode         string  `json:"app_code"`
	AppName         string  `json:"app_name"`
	Icon            *string `json:"icon"`
	AppType         string  `json:"app_type"`
	Status          string  `json:"status"`
	ChargeMode      string  `json:"charge_mode"`
	VisibilityScope string  `json:"visibility_scope"`
	Owner           *string `json:"owner"`
	Version         *string `json:"version"`
	Description     *string `json:"description"`
	SortOrder       int     `json:"sort_order"`
}

type AppUpdateRequest struct {
	AppName         string  `json:"app_name"`
	Icon            *string `json:"icon"`
	AppType         string  `json:"app_type"`
	ChargeMode      string  `json:"charge_mode"`
	VisibilityScope string  `json:"visibility_scope"`
	Owner           *string `json:"owner"`
	Version         *string `json:"version"`
	Description     *string `json:"description"`
	SortOrder       int     `json:"sort_order"`
}

type AppStatusRequest struct {
	Status string `json:"status"`
}

type AppResponse struct {
	ID              uint64    `json:"id"`
	AppCode         string    `json:"app_code"`
	AppName         string    `json:"app_name"`
	Icon            *string   `json:"icon"`
	AppType         string    `json:"app_type"`
	Source          string    `json:"source"`
	Status          string    `json:"status"`
	ChargeMode      string    `json:"charge_mode"`
	VisibilityScope string    `json:"visibility_scope"`
	Owner           *string   `json:"owner"`
	Version         *string   `json:"version"`
	Description     *string   `json:"description"`
	IsBuiltin       bool      `json:"is_builtin"`
	IsPlatformOnly  bool      `json:"is_platform_only"`
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AppListResponse struct {
	Items []AppResponse `json:"items"`
	Total int64         `json:"total"`
	Skip  int           `json:"skip"`
	Limit int           `json:"limit"`
}

type AppStatsResponse struct {
	Total          int64 `json:"total"`
	Online         int64 `json:"online"`
	Beta           int64 `json:"beta"`
	Developing     int64 `json:"developing"`
	Builtin        int64 `json:"builtin"`
	Disabled       int64 `json:"disabled"`
	Categories     int64 `json:"categories"`
	ClientApps     int64 `json:"client_apps"`
	TenantOpenings int64 `json:"tenant_openings"`
	TrialInvites   int64 `json:"trial_invites"`
	ManifestLoads  int64 `json:"manifest_loads"`
	AuditLogs      int64 `json:"audit_logs"`
}
