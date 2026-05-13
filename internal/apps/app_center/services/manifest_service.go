package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"saas_baseon_go/internal/apps/app_center/dto"
)

var ErrManifestContentRequired = errors.New("Manifest 内容不能为空")
var ErrManifestInvalidFormat = errors.New("Manifest 格式不合法")

type AppManifest struct {
	ManifestVersion string                   `json:"manifest_version" yaml:"manifest_version"`
	FragmentRole    string                   `json:"fragment_role" yaml:"fragment_role"`
	App             ManifestApp              `json:"app" yaml:"app"`
	Clients         []string                 `json:"clients" yaml:"clients"`
	Menus           []ManifestMenu           `json:"menus" yaml:"menus"`
	Operations      []ManifestOperation      `json:"operations" yaml:"operations"`
	Permissions     []ManifestPermission     `json:"permissions" yaml:"permissions"`
	APIs            []ManifestAPI            `json:"apis" yaml:"apis"`
	PackageFeatures []ManifestPackageFeature `json:"package_features" yaml:"package_features"`
	Quotas          []ManifestQuota          `json:"quotas" yaml:"quotas"`
	Documents       []ManifestDocument       `json:"documents" yaml:"documents"`
	Meta            map[string]interface{}   `json:"meta,omitempty" yaml:"meta,omitempty"`
}

type ManifestApp struct {
	AppCode             string   `json:"app_code" yaml:"app_code"`
	AppName             string   `json:"app_name" yaml:"app_name"`
	AppType             string   `json:"app_type" yaml:"app_type"`
	Source              string   `json:"source" yaml:"source"`
	Status              string   `json:"status" yaml:"status"`
	Icon                string   `json:"icon" yaml:"icon"`
	DeploymentMode      string   `json:"deployment_mode" yaml:"deployment_mode"`
	CommunicationModes  []string `json:"communication_modes" yaml:"communication_modes"`
	VisibilityScope     string   `json:"visibility_scope" yaml:"visibility_scope"`
	ChargePolicy        string   `json:"charge_policy" yaml:"charge_policy"`
	BillingMode         string   `json:"billing_mode" yaml:"billing_mode"`
	PackagePolicy       string   `json:"package_policy" yaml:"package_policy"`
	Version             string   `json:"version" yaml:"version"`
	Description         string   `json:"description" yaml:"description"`
	BackendDir          string   `json:"backend_dir" yaml:"backend_dir"`
	FrontendDir         string   `json:"frontend_dir" yaml:"frontend_dir"`
	HealthCheckURL      string   `json:"health_check_url" yaml:"health_check_url"`
	APIBaseURL          string   `json:"api_base_url" yaml:"api_base_url"`
	WebhookURL          string   `json:"webhook_url" yaml:"webhook_url"`
	OpenAPIScopes       []string `json:"open_api_scopes" yaml:"open_api_scopes"`
	TenantContext       string   `json:"tenant_context" yaml:"tenant_context"`
	SignatureStrategy   string   `json:"signature_strategy" yaml:"signature_strategy"`
	IdempotencyStrategy string   `json:"idempotency_strategy" yaml:"idempotency_strategy"`
}

type ManifestMenu struct {
	Code             string `json:"code" yaml:"code"`
	Name             string `json:"name" yaml:"name"`
	Path             string `json:"path" yaml:"path"`
	ParentCode       string `json:"parent_code" yaml:"parent_code"`
	SortOrder        int    `json:"sort_order" yaml:"sort_order"`
	PlatformOnly     bool   `json:"platform_only" yaml:"platform_only"`
	TenantVisible    bool   `json:"tenant_visible" yaml:"tenant_visible"`
	ShowInAdmin      *bool  `json:"show_in_admin" yaml:"show_in_admin"`
	TenantEditable   bool   `json:"tenant_editable" yaml:"tenant_editable"`
	IncludeInPackage bool   `json:"include_in_package" yaml:"include_in_package"`
	FeatureCode      string `json:"feature_code" yaml:"feature_code"`
	DataPermMode     string `json:"data_perm_mode" yaml:"data_perm_mode"`
}

type ManifestOperation struct {
	Code             string `json:"code" yaml:"code"`
	Name             string `json:"name" yaml:"name"`
	MenuCode         string `json:"menu_code" yaml:"menu_code"`
	PermissionCode   string `json:"permission_code" yaml:"permission_code"`
	PlatformOnly     bool   `json:"platform_only" yaml:"platform_only"`
	IncludeInPackage bool   `json:"include_in_package" yaml:"include_in_package"`
	FeatureCode      string `json:"feature_code" yaml:"feature_code"`
}

type ManifestPermission struct {
	Code             string `json:"code" yaml:"code"`
	Name             string `json:"name" yaml:"name"`
	Type             string `json:"type" yaml:"type"`
	MenuCode         string `json:"menu_code" yaml:"menu_code"`
	PlatformOnly     bool   `json:"platform_only" yaml:"platform_only"`
	IncludeInPackage bool   `json:"include_in_package" yaml:"include_in_package"`
	DataPermMode     string `json:"data_perm_mode" yaml:"data_perm_mode"`
}

type ManifestAPI struct {
	Method         string `json:"method" yaml:"method"`
	Path           string `json:"path" yaml:"path"`
	PermissionCode string `json:"permission_code" yaml:"permission_code"`
	Public         bool   `json:"public" yaml:"public"`
	Audit          bool   `json:"audit" yaml:"audit"`
}

type ManifestPackageFeature struct {
	FeatureCode      string `json:"feature_code" yaml:"feature_code"`
	FeatureName      string `json:"feature_name" yaml:"feature_name"`
	FeatureType      string `json:"feature_type" yaml:"feature_type"`
	ParentCode       string `json:"parent_code" yaml:"parent_code"`
	SourceCode       string `json:"source_code" yaml:"source_code"`
	PackagePolicy    string `json:"package_policy" yaml:"package_policy"`
	IncludeInPackage bool   `json:"include_in_package" yaml:"include_in_package"`
	Description      string `json:"description" yaml:"description"`
}

type ManifestQuota struct {
	QuotaCode        string `json:"quota_code" yaml:"quota_code"`
	QuotaName        string `json:"quota_name" yaml:"quota_name"`
	QuotaType        string `json:"quota_type" yaml:"quota_type"`
	Unit             string `json:"unit" yaml:"unit"`
	PeriodType       string `json:"period_type" yaml:"period_type"`
	IncludeInPackage bool   `json:"include_in_package" yaml:"include_in_package"`
	Description      string `json:"description" yaml:"description"`
}

type ManifestDocument struct {
	Code string `json:"code" yaml:"code"`
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url" yaml:"url"`
}

func (s *AppService) StandardManifestTemplate() string {
	return strings.TrimSpace(`manifest_version: "1.0"
fragment_role: main
app:
  app_code: demo-app
  app_name: 示例应用
  app_type: BUSINESS_APP
  source: MANIFEST
  status: INITIATED
  icon: Boxes
  deployment_mode: MERGED
  communication_modes: []
  visibility_scope: TENANT
  charge_policy: PAID
  billing_mode: SUBSCRIPTION
  package_policy: IN_PACKAGE
  version: 0.1.0
  description: 示例应用说明
  backend_dir: internal/apps/demo_app
  frontend_dir: frontend/src/apps/demo-app
  health_check_url: ""
  api_base_url: ""
  webhook_url: ""
  open_api_scopes: []
  tenant_context: ""
  signature_strategy: ""
  idempotency_strategy: ""
clients:
  - PC_WEB
menus:
  - code: demo_list
    name: 示例列表
    path: /demo
    sort_order: 1
    tenant_visible: true
    show_in_admin: true
    tenant_editable: true
    include_in_package: true
    feature_code: demo_manage
    data_perm_mode: ORG
operations:
  - code: demo_create
    name: 示例-新增
    menu_code: demo_list
    permission_code: demo:create
    include_in_package: true
    feature_code: button_demo_create
permissions:
  - code: demo:create
    name: 示例-新增
    type: OPERATION
    menu_code: demo_list
    include_in_package: true
apis:
  - method: POST
    path: /api/demo
    permission_code: demo:create
    audit: true
package_features:
  - feature_code: demo_manage
    feature_name: 示例列表
    feature_type: MENU
    source_code: demo_list
    package_policy: IN_PACKAGE
    include_in_package: true
  - feature_code: button_demo_create
    feature_name: 示例-新增
    feature_type: OPERATION
    parent_code: demo_manage
    source_code: demo:create
    package_policy: IN_PACKAGE
    include_in_package: true
quotas:
  - quota_code: max_demo_items
    quota_name: 示例数据上限
    quota_type: STATIC
    unit: COUNT
    period_type: NONE
    include_in_package: true
documents:
  - code: access-guide
    name: 接入说明
    url: docs/tech_design/应用中心接入强制规则.md
`) + "\n"
}

func (s *AppService) ParseManifestContent(ctx context.Context, viewerID uint64, fileName string, content []byte) (dto.ManifestParseResponse, error) {
	if err := s.requirePlatformViewer(ctx, viewerID); err != nil {
		return dto.ManifestParseResponse{}, err
	}
	return s.parseManifestContent(ctx, fileName, "", content)
}

func (s *AppService) ScanManifests(ctx context.Context, viewerID uint64, root string) (dto.ManifestScanResponse, error) {
	if err := s.requirePlatformViewer(ctx, viewerID); err != nil {
		return dto.ManifestScanResponse{}, err
	}
	startedAt := time.Now()
	scanRoot, err := resolveManifestScanRoot(root)
	if err != nil {
		return dto.ManifestScanResponse{}, err
	}
	items := []dto.ManifestParseResponse{}
	envelopes := []manifestEnvelope{}
	err = filepath.WalkDir(scanRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if shouldSkipManifestScanDir(name) {
				return filepath.SkipDir
			}
			return nil
		}
		if !isManifestFileName(d.Name()) {
			return nil
		}
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		env, parseErr := s.parseManifestForLoad(ctx, d.Name(), path, raw)
		if parseErr != nil {
			items = append(items, dto.ManifestParseResponse{
				FileName: d.Name(),
				FilePath: path,
				Valid:    false,
				Blockers: []string{parseErr.Error()},
			})
			return nil
		}
		envelopes = append(envelopes, env)
		items = append(items, env.parse)
		return nil
	})
	if err != nil {
		return dto.ManifestScanResponse{}, err
	}
	groups := buildManifestScanGroups(envelopes)
	response := dto.ManifestScanResponse{
		Items:     items,
		Groups:    groups,
		Total:     len(items),
		ScanRoot:  scanRoot,
		ElapsedMs: time.Since(startedAt).Milliseconds(),
	}
	for _, item := range items {
		if item.Importable {
			response.ImportableCount++
		}
		if len(item.Blockers) > 0 {
			response.BlockedCount++
		}
	}
	for _, group := range groups {
		if len(group.Blockers) > 0 {
			response.BlockedCount++
		}
	}
	return response, nil
}

func buildManifestScanGroups(envelopes []manifestEnvelope) []dto.ManifestScanGroup {
	grouped := map[string][]manifestEnvelope{}
	order := []string{}
	for _, env := range envelopes {
		key := strings.TrimSpace(env.parse.AppCode)
		if key == "" {
			key = "__missing_app_code__:" + env.parse.FilePath + ":" + env.parse.FileName
		}
		if _, ok := grouped[key]; !ok {
			order = append(order, key)
		}
		grouped[key] = append(grouped[key], env)
	}

	groups := make([]dto.ManifestScanGroup, 0, len(grouped))
	for _, key := range order {
		envs := grouped[key]
		group := dto.ManifestScanGroup{
			AppCode:       envs[0].parse.AppCode,
			AppName:       envs[0].parse.AppName,
			Mode:          "CREATE",
			FragmentCount: len(envs),
			Files:         make([]dto.ManifestParseResponse, 0, len(envs)),
			Blockers:      []string{},
			Warnings:      []string{},
		}
		merged, blockers, warnings := mergeManifestEnvelopesForPreview(envs)
		if merged.parse.Exists {
			group.Mode = "SYNC"
		}
		group.AppCode = merged.parse.AppCode
		group.AppName = merged.parse.AppName
		group.Merged = merged.parse
		group.Blockers = append(group.Blockers, blockers...)
		group.Warnings = append(group.Warnings, warnings...)
		for _, env := range envs {
			group.Files = append(group.Files, env.parse)
			if strings.EqualFold(defaultString(env.parse.FragmentRole, "main"), "main") {
				group.MainCount++
			}
			group.Blockers = append(group.Blockers, env.parse.Blockers...)
			group.Warnings = append(group.Warnings, env.parse.Warnings...)
		}
		group.Loadable = strings.TrimSpace(group.AppCode) != "" && group.MainCount == 1 && len(group.Blockers) == 0
		groups = append(groups, group)
	}
	return groups
}

func mergeManifestEnvelopesForPreview(envs []manifestEnvelope) (manifestEnvelope, []string, []string) {
	if len(envs) == 0 {
		return manifestEnvelope{}, []string{"Manifest 分组为空"}, nil
	}
	mainIndex := -1
	mainCount := 0
	for i, env := range envs {
		if strings.EqualFold(defaultString(env.parse.FragmentRole, "main"), "main") {
			mainCount++
			if mainIndex < 0 {
				mainIndex = i
			}
		}
	}
	blockers := []string{}
	warnings := []string{}
	if mainCount == 0 {
		blockers = append(blockers, "同一 app_code 必须包含一个 fragment_role=main 的 Manifest")
		mainIndex = 0
	}
	if mainCount > 1 {
		blockers = append(blockers, "同一 app_code 只能包含一个 fragment_role=main 的 Manifest")
	}
	merged := envs[mainIndex]
	merged.manifest.Clients = nil
	merged.manifest.Menus = nil
	merged.manifest.Operations = nil
	merged.manifest.Permissions = nil
	merged.manifest.APIs = nil
	merged.manifest.PackageFeatures = nil
	merged.manifest.Quotas = nil
	merged.manifest.Documents = nil

	hashParts := make([]string, 0, len(envs))
	seenClients := map[string]struct{}{}
	seenMenus := map[string]string{}
	seenOps := map[string]string{}
	seenPerms := map[string]string{}
	seenAPIs := map[string]string{}
	seenFeatures := map[string]string{}
	seenQuotas := map[string]string{}
	seenDocs := map[string]string{}

	for _, env := range envs {
		hashParts = append(hashParts, env.parse.ManifestHash)
		if merged.parse.AppCode != "" && env.parse.AppCode != "" && env.parse.AppCode != merged.parse.AppCode {
			blockers = append(blockers, "Manifest 分组内 app_code 不一致")
		}
		for _, client := range normalizedManifestClients(env.manifest.Clients) {
			if _, ok := seenClients[client]; ok {
				continue
			}
			seenClients[client] = struct{}{}
			merged.manifest.Clients = append(merged.manifest.Clients, client)
		}
		for _, item := range env.manifest.Menus {
			code := strings.TrimSpace(item.Code)
			if addManifestDuplicateBlocker(&blockers, seenMenus, "菜单", code, env.parse.FileName) {
				continue
			}
			merged.manifest.Menus = append(merged.manifest.Menus, item)
		}
		for _, item := range env.manifest.Operations {
			code := strings.TrimSpace(defaultString(item.PermissionCode, item.Code))
			if addManifestDuplicateBlocker(&blockers, seenOps, "操作权限", code, env.parse.FileName) {
				continue
			}
			merged.manifest.Operations = append(merged.manifest.Operations, item)
		}
		for _, item := range env.manifest.Permissions {
			code := strings.TrimSpace(item.Code)
			if addManifestDuplicateBlocker(&blockers, seenPerms, "权限点", code, env.parse.FileName) {
				continue
			}
			merged.manifest.Permissions = append(merged.manifest.Permissions, item)
		}
		for _, item := range env.manifest.APIs {
			code := strings.ToUpper(strings.TrimSpace(item.Method)) + " " + strings.TrimSpace(item.Path)
			if addManifestDuplicateBlocker(&blockers, seenAPIs, "API", code, env.parse.FileName) {
				continue
			}
			merged.manifest.APIs = append(merged.manifest.APIs, item)
		}
		for _, item := range env.manifest.PackageFeatures {
			code := strings.TrimSpace(item.FeatureCode)
			if addManifestDuplicateBlocker(&blockers, seenFeatures, "套餐功能点", code, env.parse.FileName) {
				continue
			}
			merged.manifest.PackageFeatures = append(merged.manifest.PackageFeatures, item)
		}
		for _, item := range env.manifest.Quotas {
			code := strings.TrimSpace(item.QuotaCode)
			if addManifestDuplicateBlocker(&blockers, seenQuotas, "配额", code, env.parse.FileName) {
				continue
			}
			merged.manifest.Quotas = append(merged.manifest.Quotas, item)
		}
		for _, item := range env.manifest.Documents {
			code := strings.TrimSpace(item.Code)
			if addManifestDuplicateBlocker(&blockers, seenDocs, "文档", code, env.parse.FileName) {
				continue
			}
			merged.manifest.Documents = append(merged.manifest.Documents, item)
		}
	}

	sum := sha256.Sum256([]byte(strings.Join(hashParts, "|")))
	merged.parse.ManifestHash = hex.EncodeToString(sum[:])
	merged.parse.ClientCodes = normalizedManifestClients(merged.manifest.Clients)
	merged.parse.Counts = dto.ManifestAssetCounts{
		Clients:         len(merged.manifest.Clients),
		Menus:           len(merged.manifest.Menus),
		Operations:      len(merged.manifest.Operations),
		Permissions:     len(merged.manifest.Permissions),
		APIs:            len(merged.manifest.APIs),
		PackageFeatures: len(merged.manifest.PackageFeatures),
		Quotas:          len(merged.manifest.Quotas),
		Documents:       len(merged.manifest.Documents),
	}
	validationBlockers, validationWarnings := validateManifest(merged.manifest)
	blockers = append(blockers, validationBlockers...)
	warnings = append(warnings, validationWarnings...)
	merged.parse.Blockers = blockers
	merged.parse.Warnings = warnings
	merged.parse.Valid = len(blockers) == 0
	merged.parse.Importable = merged.parse.Valid && !merged.parse.Exists
	return merged, blockers, warnings
}

func addManifestDuplicateBlocker(blockers *[]string, seen map[string]string, label string, code string, fileName string) bool {
	if strings.TrimSpace(code) == "" {
		*blockers = append(*blockers, label+"编码不能为空")
		return true
	}
	if previous, ok := seen[code]; ok {
		*blockers = append(*blockers, label+"编码重复："+code+"（"+previous+" / "+fileName+"）")
		return true
	}
	seen[code] = fileName
	return false
}

func (s *AppService) parseManifestContent(ctx context.Context, fileName string, filePath string, content []byte) (dto.ManifestParseResponse, error) {
	if len(strings.TrimSpace(string(content))) == 0 {
		return dto.ManifestParseResponse{}, ErrManifestContentRequired
	}
	manifest, err := decodeManifest(fileName, content)
	if err != nil {
		return dto.ManifestParseResponse{}, err
	}
	sum := sha256.Sum256(content)
	result := dto.ManifestParseResponse{
		FileName:        fileName,
		FilePath:        filePath,
		ManifestHash:    hex.EncodeToString(sum[:]),
		ManifestVersion: defaultString(manifest.ManifestVersion, "1.0"),
		FragmentRole:    defaultString(manifest.FragmentRole, "main"),
		AppCode:         strings.TrimSpace(manifest.App.AppCode),
		AppName:         strings.TrimSpace(manifest.App.AppName),
		AppType:         strings.TrimSpace(manifest.App.AppType),
		Source:          strings.TrimSpace(manifest.App.Source),
		Status:          strings.TrimSpace(manifest.App.Status),
		DeploymentMode:  strings.TrimSpace(manifest.App.DeploymentMode),
		CommModes:       normalizedStringList(manifest.App.CommunicationModes),
		VisibilityScope: strings.TrimSpace(manifest.App.VisibilityScope),
		ChargePolicy:    strings.TrimSpace(manifest.App.ChargePolicy),
		BillingMode:     strings.TrimSpace(manifest.App.BillingMode),
		PackagePolicy:   strings.TrimSpace(manifest.App.PackagePolicy),
		ClientCodes:     normalizedManifestClients(manifest.Clients),
		Counts: dto.ManifestAssetCounts{
			Clients:         len(manifest.Clients),
			Menus:           len(manifest.Menus),
			Operations:      len(manifest.Operations),
			Permissions:     len(manifest.Permissions),
			APIs:            len(manifest.APIs),
			PackageFeatures: len(manifest.PackageFeatures),
			Quotas:          len(manifest.Quotas),
			Documents:       len(manifest.Documents),
		},
	}
	result.Blockers, result.Warnings = validateManifest(manifest)
	if result.AppCode != "" {
		_, err := s.repo.GetByCode(ctx, result.AppCode)
		if err == nil {
			result.Exists = true
			result.Blockers = append(result.Blockers, "app_code 已存在，新建导入被阻断，只能进入 Manifest 同步 / 升级")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ManifestParseResponse{}, err
		}
	}
	result.Valid = len(result.Blockers) == 0
	result.Importable = result.Valid && !result.Exists
	return result, nil
}

func normalizedManifestClients(values []string) []string {
	return normalizedStringList(values)
}

func normalizedStringList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		code := strings.ToUpper(strings.TrimSpace(value))
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}
	return out
}

func hasString(values []string, target string) bool {
	normalizedTarget := strings.ToUpper(strings.TrimSpace(target))
	for _, value := range values {
		if strings.ToUpper(strings.TrimSpace(value)) == normalizedTarget {
			return true
		}
	}
	return false
}

func decodeManifest(fileName string, content []byte) (AppManifest, error) {
	var manifest AppManifest
	trimmed := strings.TrimSpace(string(content))
	if strings.HasSuffix(strings.ToLower(fileName), ".json") || strings.HasPrefix(trimmed, "{") {
		if err := json.Unmarshal(content, &manifest); err != nil {
			return AppManifest{}, ErrManifestInvalidFormat
		}
		return manifest, nil
	}
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return AppManifest{}, ErrManifestInvalidFormat
	}
	return manifest, nil
}

func validateManifest(manifest AppManifest) ([]string, []string) {
	blockers := []string{}
	warnings := []string{}
	if strings.TrimSpace(manifest.App.AppCode) == "" {
		blockers = append(blockers, "app.app_code 不能为空")
	} else if !appCodePattern.MatchString(strings.TrimSpace(manifest.App.AppCode)) {
		blockers = append(blockers, ErrInvalidAppCode.Error())
	}
	if strings.TrimSpace(manifest.App.AppName) == "" {
		blockers = append(blockers, "app.app_name 不能为空")
	}
	if strings.TrimSpace(manifest.App.PackagePolicy) == "" {
		warnings = append(warnings, "app.package_policy 未声明；未纳入套餐的能力后续应进入套餐外订阅或非售卖治理")
	}
	if len(manifest.Clients) == 0 {
		warnings = append(warnings, "clients 为空，应用不会生成客户端配置")
	}
	if len(manifest.PackageFeatures) == 0 {
		warnings = append(warnings, "package_features 为空，套餐中心不会收录该应用功能点")
	}
	if strings.EqualFold(strings.TrimSpace(manifest.App.DeploymentMode), "STANDALONE") {
		blockers = append(blockers, validateStandaloneManifest(manifest.App)...)
	}
	return blockers, warnings
}

func validateStandaloneManifest(app ManifestApp) []string {
	blockers := []string{}
	commModes := normalizedStringList(app.CommunicationModes)
	if len(commModes) == 0 {
		blockers = append(blockers, "独立部署应用必须声明 app.communication_modes")
	}
	for _, mode := range commModes {
		if !allowedStandaloneCommunicationModes[mode] {
			blockers = append(blockers, "独立部署应用通讯方式仅允许 PLATFORM_API、WEBHOOK、DATA_SYNC、GATEWAY_PROXY")
			break
		}
	}
	if len(normalizedStringList(app.OpenAPIScopes)) == 0 {
		blockers = append(blockers, "独立部署应用必须声明 app.open_api_scopes")
	}
	if strings.TrimSpace(app.TenantContext) == "" {
		blockers = append(blockers, "独立部署应用必须声明 app.tenant_context")
	}
	if strings.TrimSpace(app.SignatureStrategy) == "" {
		blockers = append(blockers, "独立部署应用必须声明 app.signature_strategy")
	}
	if strings.TrimSpace(app.IdempotencyStrategy) == "" {
		blockers = append(blockers, "独立部署应用必须声明 app.idempotency_strategy")
	}
	if hasString(commModes, "PLATFORM_API") && strings.TrimSpace(app.APIBaseURL) == "" {
		blockers = append(blockers, "通讯方式包含 PLATFORM_API 时必须声明 app.api_base_url")
	}
	if hasString(commModes, "WEBHOOK") && strings.TrimSpace(app.WebhookURL) == "" {
		blockers = append(blockers, "通讯方式包含 WEBHOOK 时必须声明 app.webhook_url")
	}
	if hasString(commModes, "GATEWAY_PROXY") && strings.TrimSpace(app.APIBaseURL) == "" {
		blockers = append(blockers, "通讯方式包含 GATEWAY_PROXY 时必须声明 app.api_base_url 或网关路由前缀")
	}
	return blockers
}

var allowedStandaloneCommunicationModes = map[string]bool{
	"PLATFORM_API":  true,
	"WEBHOOK":       true,
	"DATA_SYNC":     true,
	"GATEWAY_PROXY": true,
}

func resolveManifestScanRoot(root string) (string, error) {
	if strings.TrimSpace(root) != "" {
		return filepath.Abs(root)
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, statErr := os.Stat(filepath.Join(wd, "go.mod")); statErr == nil {
			appsRoot := filepath.Join(wd, "internal", "apps")
			if _, appsErr := os.Stat(appsRoot); appsErr == nil {
				return appsRoot, nil
			}
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return os.Getwd()
		}
		wd = parent
	}
}

func shouldSkipManifestScanDir(name string) bool {
	switch name {
	case ".git", ".idea", ".vscode", ".cache", ".next", ".nuxt", ".turbo", "node_modules", "dist", "build", "coverage", "tmp", "vendor":
		return true
	default:
		return false
	}
}

func isManifestFileName(name string) bool {
	lower := strings.ToLower(name)
	return lower == "app.manifest.yaml" || lower == "app.manifest.yml" || lower == "app.manifest.json"
}
