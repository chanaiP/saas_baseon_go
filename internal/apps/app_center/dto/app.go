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
	AppCode         string             `json:"app_code"`
	AppName         string             `json:"app_name"`
	Icon            *string            `json:"icon"`
	AppType         string             `json:"app_type"`
	Status          string             `json:"status"`
	ChargeMode      string             `json:"charge_mode"`
	VisibilityScope string             `json:"visibility_scope"`
	Owner           *string            `json:"owner"`
	OwnerUserIDs    *string            `json:"owner_user_ids"`
	Version         *string            `json:"version"`
	Description     *string            `json:"description"`
	DetailDesc      *string            `json:"detail_description"`
	DeploymentMode  string             `json:"deployment_mode"`
	CommModes       *string            `json:"communication_modes"`
	VisibilityMode  *string            `json:"visibility_mode"`
	VisibleTenants  *string            `json:"visible_tenants"`
	OpenMethod      *string            `json:"open_method"`
	TrialPolicy     *string            `json:"trial_policy"`
	TrialStartRule  *string            `json:"trial_start_rule"`
	AssetConfig     *string            `json:"asset_config"`
	DocConfig       *string            `json:"doc_config"`
	ReleaseChannel  *string            `json:"release_channel"`
	ReleaseNote     *string            `json:"release_note"`
	SortOrder       int                `json:"sort_order"`
	Clients         []AppClientRequest `json:"clients"`
}

type AppUpdateRequest struct {
	AppName         string             `json:"app_name"`
	Icon            *string            `json:"icon"`
	AppType         string             `json:"app_type"`
	ChargeMode      string             `json:"charge_mode"`
	VisibilityScope string             `json:"visibility_scope"`
	Owner           *string            `json:"owner"`
	OwnerUserIDs    *string            `json:"owner_user_ids"`
	Version         *string            `json:"version"`
	Description     *string            `json:"description"`
	DetailDesc      *string            `json:"detail_description"`
	DeploymentMode  string             `json:"deployment_mode"`
	CommModes       *string            `json:"communication_modes"`
	VisibilityMode  *string            `json:"visibility_mode"`
	VisibleTenants  *string            `json:"visible_tenants"`
	OpenMethod      *string            `json:"open_method"`
	TrialPolicy     *string            `json:"trial_policy"`
	TrialStartRule  *string            `json:"trial_start_rule"`
	AssetConfig     *string            `json:"asset_config"`
	DocConfig       *string            `json:"doc_config"`
	ReleaseChannel  *string            `json:"release_channel"`
	ReleaseNote     *string            `json:"release_note"`
	SortOrder       int                `json:"sort_order"`
	Clients         []AppClientRequest `json:"clients"`
}

type AppStatusRequest struct {
	Status string `json:"status"`
}

type AppClientRequest struct {
	ClientCode string  `json:"client_code"`
	ClientName string  `json:"client_name"`
	Enabled    bool    `json:"enabled"`
	SortOrder  int     `json:"sort_order"`
	ConfigNote *string `json:"config_note"`
}

type AppResponse struct {
	ID              uint64              `json:"id"`
	AppCode         string              `json:"app_code"`
	AppName         string              `json:"app_name"`
	Icon            *string             `json:"icon"`
	AppType         string              `json:"app_type"`
	Source          string              `json:"source"`
	Status          string              `json:"status"`
	ChargeMode      string              `json:"charge_mode"`
	VisibilityScope string              `json:"visibility_scope"`
	Owner           *string             `json:"owner"`
	OwnerUserIDs    *string             `json:"owner_user_ids"`
	Version         *string             `json:"version"`
	Description     *string             `json:"description"`
	DetailDesc      *string             `json:"detail_description"`
	DeploymentMode  string              `json:"deployment_mode"`
	CommModes       *string             `json:"communication_modes"`
	VisibilityMode  *string             `json:"visibility_mode"`
	VisibleTenants  *string             `json:"visible_tenants"`
	OpenMethod      *string             `json:"open_method"`
	TrialPolicy     *string             `json:"trial_policy"`
	TrialStartRule  *string             `json:"trial_start_rule"`
	AssetConfig     *string             `json:"asset_config"`
	DocConfig       *string             `json:"doc_config"`
	ReleaseChannel  *string             `json:"release_channel"`
	ReleaseNote     *string             `json:"release_note"`
	IsBuiltin       bool                `json:"is_builtin"`
	IsPlatformOnly  bool                `json:"is_platform_only"`
	SortOrder       int                 `json:"sort_order"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	Clients         []AppClientResponse `json:"clients"`
}

type AppClientResponse struct {
	ID         uint64    `json:"id"`
	AppID      uint64    `json:"app_id"`
	ClientCode string    `json:"client_code"`
	ClientName string    `json:"client_name"`
	Enabled    bool      `json:"enabled"`
	SortOrder  int       `json:"sort_order"`
	ConfigNote *string   `json:"config_note"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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

type ManifestParseRequest struct {
	FileName string `json:"file_name"`
	Content  string `json:"content"`
}

type ManifestScanRequest struct {
	Root string `json:"root"`
}

type ManifestScanResponse struct {
	Items           []ManifestParseResponse `json:"items"`
	Total           int                     `json:"total"`
	ImportableCount int                     `json:"importable_count"`
	BlockedCount    int                     `json:"blocked_count"`
}

type ManifestParseResponse struct {
	FileName        string              `json:"file_name"`
	FilePath        string              `json:"file_path,omitempty"`
	ManifestHash    string              `json:"manifest_hash"`
	ManifestVersion string              `json:"manifest_version"`
	FragmentRole    string              `json:"fragment_role"`
	AppCode         string              `json:"app_code"`
	AppName         string              `json:"app_name"`
	AppType         string              `json:"app_type"`
	Source          string              `json:"source"`
	Status          string              `json:"status"`
	DeploymentMode  string              `json:"deployment_mode"`
	VisibilityScope string              `json:"visibility_scope"`
	ChargePolicy    string              `json:"charge_policy"`
	BillingMode     string              `json:"billing_mode"`
	PackagePolicy   string              `json:"package_policy"`
	ClientCodes     []string            `json:"client_codes"`
	Exists          bool                `json:"exists"`
	Importable      bool                `json:"importable"`
	Valid           bool                `json:"valid"`
	Blockers        []string            `json:"blockers"`
	Warnings        []string            `json:"warnings"`
	Counts          ManifestAssetCounts `json:"counts"`
}

type ManifestAssetCounts struct {
	Clients         int `json:"clients"`
	Menus           int `json:"menus"`
	Operations      int `json:"operations"`
	Permissions     int `json:"permissions"`
	APIs            int `json:"apis"`
	PackageFeatures int `json:"package_features"`
	Quotas          int `json:"quotas"`
	Documents       int `json:"documents"`
}
