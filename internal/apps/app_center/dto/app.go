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
	HealthCheckURL  *string            `json:"health_check_url"`
	APIBaseURL      *string            `json:"api_base_url"`
	WebhookURL      *string            `json:"webhook_url"`
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
	HealthCheckURL  *string            `json:"health_check_url"`
	APIBaseURL      *string            `json:"api_base_url"`
	WebhookURL      *string            `json:"webhook_url"`
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
	ID               uint64              `json:"id"`
	AppCode          string              `json:"app_code"`
	AppName          string              `json:"app_name"`
	Icon             *string             `json:"icon"`
	AppType          string              `json:"app_type"`
	Source           string              `json:"source"`
	Status           string              `json:"status"`
	ChargeMode       string              `json:"charge_mode"`
	VisibilityScope  string              `json:"visibility_scope"`
	Owner            *string             `json:"owner"`
	OwnerUserIDs     *string             `json:"owner_user_ids"`
	Version          *string             `json:"version"`
	Description      *string             `json:"description"`
	DetailDesc       *string             `json:"detail_description"`
	DeploymentMode   string              `json:"deployment_mode"`
	CommModes        *string             `json:"communication_modes"`
	VisibilityMode   *string             `json:"visibility_mode"`
	VisibleTenants   *string             `json:"visible_tenants"`
	OpenMethod       *string             `json:"open_method"`
	TrialPolicy      *string             `json:"trial_policy"`
	TrialStartRule   *string             `json:"trial_start_rule"`
	AssetConfig      *string             `json:"asset_config"`
	DocConfig        *string             `json:"doc_config"`
	ReleaseChannel   *string             `json:"release_channel"`
	ReleaseNote      *string             `json:"release_note"`
	HealthCheckURL   *string             `json:"health_check_url"`
	APIBaseURL       *string             `json:"api_base_url"`
	WebhookURL       *string             `json:"webhook_url"`
	ManifestHash     *string             `json:"manifest_hash"`
	ManifestVersion  *string             `json:"manifest_version"`
	LastManifestSync *time.Time          `json:"last_manifest_synced_at"`
	IsBuiltin        bool                `json:"is_builtin"`
	IsPlatformOnly   bool                `json:"is_platform_only"`
	SortOrder        int                 `json:"sort_order"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	Clients          []AppClientResponse `json:"clients"`
	Assets           *AppAssetsResponse  `json:"assets,omitempty"`
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

type AppAssetsResponse struct {
	Entries         []AppEntryResponse          `json:"entries"`
	APIs            []AppAPIResponse            `json:"apis"`
	Permissions     []AppPermissionResponse     `json:"permissions"`
	PackageFeatures []AppPackageFeatureResponse `json:"package_features"`
	Quotas          []AppQuotaResponse          `json:"quotas"`
	ManifestLoads   []AppManifestLoadRecord     `json:"manifest_loads"`
}

type AppEntryResponse struct {
	ResourceCode     string     `json:"resource_code"`
	Name             string     `json:"name"`
	Path             string     `json:"path"`
	ParentCode       *string    `json:"parent_code"`
	SortOrder        int        `json:"sort_order"`
	PlatformOnly     bool       `json:"platform_only"`
	TenantVisible    bool       `json:"tenant_visible"`
	TenantEditable   bool       `json:"tenant_editable"`
	IncludeInPackage bool       `json:"include_in_package"`
	FeatureCode      *string    `json:"feature_code"`
	DataPermMode     string     `json:"data_perm_mode"`
	Status           string     `json:"status"`
	LastSyncedAt     *time.Time `json:"last_synced_at,omitempty"`
	ProtectionSource *string    `json:"protection_source,omitempty"`
	ProtectionReason *string    `json:"protection_reason,omitempty"`
	ProtectedAt      *time.Time `json:"protected_at,omitempty"`
}

type AppAPIResponse struct {
	Method           string     `json:"method"`
	Path             string     `json:"path"`
	PermissionCode   *string    `json:"permission_code"`
	Public           bool       `json:"public"`
	Audit            bool       `json:"audit"`
	Status           string     `json:"status"`
	LastSyncedAt     *time.Time `json:"last_synced_at,omitempty"`
	ProtectionSource *string    `json:"protection_source,omitempty"`
	ProtectionReason *string    `json:"protection_reason,omitempty"`
	ProtectedAt      *time.Time `json:"protected_at,omitempty"`
}

type AppPermissionResponse struct {
	PermissionCode   string     `json:"permission_code"`
	Name             string     `json:"name"`
	PermissionType   string     `json:"permission_type"`
	MenuCode         *string    `json:"menu_code"`
	PlatformOnly     bool       `json:"platform_only"`
	IncludeInPackage bool       `json:"include_in_package"`
	DataPermMode     string     `json:"data_perm_mode"`
	Status           string     `json:"status"`
	LastSyncedAt     *time.Time `json:"last_synced_at,omitempty"`
	ProtectionSource *string    `json:"protection_source,omitempty"`
	ProtectionReason *string    `json:"protection_reason,omitempty"`
	ProtectedAt      *time.Time `json:"protected_at,omitempty"`
}

type AppPackageFeatureResponse struct {
	FeatureCode      string     `json:"feature_code"`
	FeatureName      string     `json:"feature_name"`
	FeatureType      string     `json:"feature_type"`
	ParentCode       *string    `json:"parent_code"`
	SourceCode       *string    `json:"source_code"`
	PackagePolicy    string     `json:"package_policy"`
	IncludeInPackage bool       `json:"include_in_package"`
	Status           string     `json:"status"`
	LastSyncedAt     *time.Time `json:"last_synced_at,omitempty"`
	ProtectionSource *string    `json:"protection_source,omitempty"`
	ProtectionReason *string    `json:"protection_reason,omitempty"`
	ProtectedAt      *time.Time `json:"protected_at,omitempty"`
}

type AppQuotaResponse struct {
	QuotaCode        string     `json:"quota_code"`
	QuotaName        string     `json:"quota_name"`
	QuotaType        string     `json:"quota_type"`
	Unit             *string    `json:"unit"`
	PeriodType       *string    `json:"period_type"`
	IncludeInPackage bool       `json:"include_in_package"`
	Status           string     `json:"status"`
	LastSyncedAt     *time.Time `json:"last_synced_at,omitempty"`
	ProtectionSource *string    `json:"protection_source,omitempty"`
	ProtectionReason *string    `json:"protection_reason,omitempty"`
	ProtectedAt      *time.Time `json:"protected_at,omitempty"`
}

type AppManifestLoadRecord struct {
	ID              uint64    `json:"id"`
	Action          string    `json:"action"`
	SourceType      string    `json:"source_type"`
	SourceName      *string   `json:"source_name"`
	ManifestVersion string    `json:"manifest_version"`
	ManifestHash    string    `json:"manifest_hash"`
	FragmentRole    string    `json:"fragment_role"`
	Status          string    `json:"status"`
	Summary         *string   `json:"summary"`
	ErrorSummary    *string   `json:"error_summary"`
	OperatorUserID  uint64    `json:"operator_user_id"`
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

type ManifestParseRequest struct {
	FileName string `json:"file_name"`
	Content  string `json:"content"`
}

type ManifestLoadRequest struct {
	FileName   string   `json:"file_name"`
	FilePath   string   `json:"file_path"`
	FilePaths  []string `json:"file_paths"`
	Content    string   `json:"content"`
	SourceType string   `json:"source_type"`
}

type ManifestScanRequest struct {
	Root string `json:"root"`
}

type ManifestScanResponse struct {
	Items           []ManifestParseResponse `json:"items"`
	Groups          []ManifestScanGroup     `json:"groups"`
	Total           int                     `json:"total"`
	ImportableCount int                     `json:"importable_count"`
	BlockedCount    int                     `json:"blocked_count"`
	ScanRoot        string                  `json:"scan_root"`
	ElapsedMs       int64                   `json:"elapsed_ms"`
}

type ManifestScanGroup struct {
	AppCode       string                  `json:"app_code"`
	AppName       string                  `json:"app_name"`
	Mode          string                  `json:"mode"`
	Loadable      bool                    `json:"loadable"`
	FragmentCount int                     `json:"fragment_count"`
	MainCount     int                     `json:"main_count"`
	Files         []ManifestParseResponse `json:"files"`
	Merged        ManifestParseResponse   `json:"merged"`
	Blockers      []string                `json:"blockers"`
	Warnings      []string                `json:"warnings"`
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
	CommModes       []string            `json:"communication_modes"`
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

type ManifestDiffResponse struct {
	Parse    ManifestParseResponse `json:"parse"`
	Mode     string                `json:"mode"`
	Loadable bool                  `json:"loadable"`
	Summary  ManifestDiffSummary   `json:"summary"`
	Changes  []ManifestDiffChange  `json:"changes"`
	Blockers []string              `json:"blockers"`
	Warnings []string              `json:"warnings"`
}

type ManifestDiffSummary struct {
	Create   int `json:"create"`
	Update   int `json:"update"`
	NoChange int `json:"no_change"`
	Disable  int `json:"disable"`
	Conflict int `json:"conflict"`
}

type ManifestDiffChange struct {
	ResourceType string `json:"resource_type"`
	ResourceCode string `json:"resource_code"`
	Name         string `json:"name"`
	Action       string `json:"action"`
	Severity     string `json:"severity"`
	Message      string `json:"message"`
}

type ManifestLoadResponse struct {
	LoadID  uint64               `json:"load_id"`
	AppCode string               `json:"app_code"`
	Status  string               `json:"status"`
	Summary ManifestDiffSummary  `json:"summary"`
	Diff    ManifestDiffResponse `json:"diff"`
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
