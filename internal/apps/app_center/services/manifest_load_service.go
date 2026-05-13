package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"saas_baseon_go/internal/apps/app_center/dto"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

var ErrManifestLoadBlocked = errors.New("Manifest 存在阻断项，不能装载")

const legacyManifestBaselineHash = "LEGACY_BASELINE"

type manifestEnvelope struct {
	manifest AppManifest
	parse    dto.ManifestParseResponse
	content  []byte
}

func (s *AppService) DiffManifest(ctx context.Context, viewerID uint64, fileName string, filePath string, content []byte) (dto.ManifestDiffResponse, error) {
	return s.DiffManifestRequest(ctx, viewerID, dto.ManifestLoadRequest{
		FileName: fileName,
		FilePath: filePath,
		Content:  string(content),
	})
}

func (s *AppService) DiffManifestRequest(ctx context.Context, viewerID uint64, req dto.ManifestLoadRequest) (dto.ManifestDiffResponse, error) {
	if err := s.requirePlatformViewer(ctx, viewerID); err != nil {
		return dto.ManifestDiffResponse{}, err
	}
	env, err := s.manifestEnvelopeFromRequest(ctx, req)
	if err != nil {
		return dto.ManifestDiffResponse{}, err
	}
	return s.diffManifest(ctx, env)
}

func (s *AppService) LoadManifest(ctx context.Context, viewerID uint64, req dto.ManifestLoadRequest) (dto.ManifestLoadResponse, error) {
	if err := s.requirePlatformViewer(ctx, viewerID); err != nil {
		return dto.ManifestLoadResponse{}, err
	}
	env, err := s.manifestEnvelopeFromRequest(ctx, req)
	if err != nil {
		return dto.ManifestLoadResponse{}, err
	}
	diff, err := s.diffManifest(ctx, env)
	if err != nil {
		return dto.ManifestLoadResponse{}, err
	}
	if !diff.Loadable {
		return dto.ManifestLoadResponse{}, ErrManifestLoadBlocked
	}

	db := s.repo.DB()
	sourceType := strings.ToUpper(strings.TrimSpace(req.SourceType))
	if sourceType == "" {
		sourceType = "UPLOAD"
	}
	var loadID uint64
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		diffRaw, _ := json.Marshal(diff)
		load := models.SysAppManifestLoad{
			AppCode:         diff.Parse.AppCode,
			Action:          diff.Mode,
			SourceType:      sourceType,
			SourceName:      cleanOptionalString(&req.FileName),
			ManifestVersion: diff.Parse.ManifestVersion,
			ManifestHash:    diff.Parse.ManifestHash,
			FragmentRole:    diff.Parse.FragmentRole,
			Status:          "RUNNING",
			DiffSummary:     serviceStringPtr(string(diffRaw)),
			OperatorUserID:  viewerID,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := tx.Create(&load).Error; err != nil {
			return err
		}
		loadID = load.ID

		if err := s.persistManifestFile(tx, load.ID, env, req.FilePath, now); err != nil {
			return markLoadFailed(tx, load.ID, err)
		}
		if err := s.upsertManifestApp(tx, env, now); err != nil {
			return markLoadFailed(tx, load.ID, err)
		}
		app, err := s.repoAppByCode(tx, env.parse.AppCode)
		if err != nil {
			return markLoadFailed(tx, load.ID, err)
		}
		if err := s.syncManifestClients(tx, app.ID, env.manifest.Clients, now); err != nil {
			return markLoadFailed(tx, load.ID, err)
		}
		if err := s.syncManifestAssets(tx, env, now); err != nil {
			return markLoadFailed(tx, load.ID, err)
		}
		if err := s.syncManifestPermissions(tx, env, now); err != nil {
			return markLoadFailed(tx, load.ID, err)
		}
		if err := s.syncManifestPackageCenter(tx, env, now); err != nil {
			return markLoadFailed(tx, load.ID, err)
		}
		userID := viewerID
		detail := string(diffRaw)
		if err := tx.Create(&models.AuditLog{
			UserID:    &userID,
			AppCode:   &env.parse.AppCode,
			Module:    "app_center",
			Action:    "manifest_load",
			Summary:   fmt.Sprintf("装载应用 Manifest：%s", env.parse.AppCode),
			Detail:    &detail,
			Result:    "success",
			CreatedAt: now,
		}).Error; err != nil {
			return markLoadFailed(tx, load.ID, err)
		}
		summary := fmt.Sprintf("新增 %d，更新 %d，不变 %d，停用 %d", diff.Summary.Create, diff.Summary.Update, diff.Summary.NoChange, diff.Summary.Disable)
		return tx.Model(&models.SysAppManifestLoad{}).
			Where("id = ?", load.ID).
			Updates(map[string]interface{}{"status": "SUCCESS", "summary": summary, "updated_at": now}).Error
	})
	if err != nil {
		return dto.ManifestLoadResponse{}, err
	}
	return dto.ManifestLoadResponse{
		LoadID:  loadID,
		AppCode: diff.Parse.AppCode,
		Status:  "SUCCESS",
		Summary: diff.Summary,
		Diff:    diff,
	}, nil
}

func (s *AppService) parseManifestForLoad(ctx context.Context, fileName string, filePath string, content []byte) (manifestEnvelope, error) {
	if len(strings.TrimSpace(string(content))) == 0 {
		return manifestEnvelope{}, ErrManifestContentRequired
	}
	manifest, err := decodeManifest(fileName, content)
	if err != nil {
		return manifestEnvelope{}, err
	}
	sum := sha256.Sum256(content)
	parse := dto.ManifestParseResponse{
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
	parse.Blockers, parse.Warnings = validateManifest(manifest)
	if parse.AppCode != "" {
		_, err := s.repo.GetByCode(ctx, parse.AppCode)
		if err == nil {
			parse.Exists = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return manifestEnvelope{}, err
		}
	}
	parse.Valid = len(parse.Blockers) == 0
	parse.Importable = parse.Valid && !parse.Exists
	return manifestEnvelope{manifest: manifest, parse: parse, content: content}, nil
}

func (s *AppService) diffManifest(ctx context.Context, env manifestEnvelope) (dto.ManifestDiffResponse, error) {
	mode := "CREATE"
	if env.parse.Exists {
		mode = "SYNC"
	}
	result := dto.ManifestDiffResponse{
		Parse:    env.parse,
		Mode:     mode,
		Blockers: append([]string{}, env.parse.Blockers...),
		Warnings: append([]string{}, env.parse.Warnings...),
	}
	addChange := func(change dto.ManifestDiffChange) {
		result.Changes = append(result.Changes, change)
		switch change.Action {
		case "CREATE":
			result.Summary.Create++
		case "UPDATE":
			result.Summary.Update++
		case "NO_CHANGE":
			result.Summary.NoChange++
		case "DISABLE":
			result.Summary.Disable++
		case "CONFLICT":
			result.Summary.Conflict++
			result.Blockers = append(result.Blockers, change.Message)
		}
	}

	db := s.repo.DB().WithContext(ctx)
	appAction := "CREATE"
	if env.parse.Exists {
		var row models.SysApp
		if err := db.Where("app_code = ? AND deleted_at IS NULL", env.parse.AppCode).First(&row).Error; err != nil {
			return dto.ManifestDiffResponse{}, err
		}
		if strings.EqualFold(row.AppName, env.parse.AppName) &&
			strings.EqualFold(defaultString(row.AppType, "BUSINESS_APP"), defaultString(env.parse.AppType, "BUSINESS_APP")) &&
			strings.EqualFold(defaultString(row.DeploymentMode, "MERGED"), defaultString(env.parse.DeploymentMode, "MERGED")) &&
			stringValue(row.ManifestHash) == env.parse.ManifestHash {
			appAction = "NO_CHANGE"
		} else {
			appAction = "UPDATE"
		}
	}
	addChange(dto.ManifestDiffChange{ResourceType: "APP", ResourceCode: env.parse.AppCode, Name: env.parse.AppName, Action: appAction, Severity: "INFO", Message: "应用主档"})

	for _, client := range env.parse.ClientCodes {
		action := "CREATE"
		if env.parse.Exists {
			var count int64
			if err := db.Model(&models.SysAppClient{}).
				Joins("JOIN sys_app ON sys_app.id = sys_app_client.app_id").
				Where("sys_app.app_code = ? AND sys_app_client.client_code = ? AND sys_app_client.deleted_at IS NULL", env.parse.AppCode, client).
				Count(&count).Error; err != nil {
				return dto.ManifestDiffResponse{}, err
			}
			if count > 0 {
				action = "NO_CHANGE"
			}
		}
		addChange(dto.ManifestDiffChange{ResourceType: "CLIENT", ResourceCode: client, Name: manifestClientName(client), Action: action, Severity: "INFO", Message: "客户端形态"})
	}

	declaredMenus := map[string]struct{}{}
	for _, menu := range env.manifest.Menus {
		declaredMenus[strings.TrimSpace(menu.Code)] = struct{}{}
		if menu.IncludeInPackage && menu.PlatformOnly {
			addChange(conflictChange("MENU", menu.Code, menu.Name, "平台专属菜单不能纳入套餐功能点"))
			continue
		}
		if conflict, err := s.permissionPathOwnedByOtherApp(db, env.parse.AppCode, menu.Path); err != nil {
			return dto.ManifestDiffResponse{}, err
		} else if conflict {
			addChange(conflictChange("MENU", menu.Code, menu.Name, "菜单 path 已被其他应用占用"))
			continue
		}
		action, err := s.diffByManifestAsset(db, &models.SysAppEntry{}, env.parse.AppCode, "resource_code", menu.Code, env.parse.ManifestHash)
		if err != nil {
			return dto.ManifestDiffResponse{}, err
		}
		addChange(dto.ManifestDiffChange{ResourceType: "MENU", ResourceCode: menu.Code, Name: menu.Name, Action: action, Severity: severityForManifestAction(action), Message: messageForManifestAction(action, "菜单入口")})
	}
	if env.parse.Exists {
		changes, err := s.diffMissingManifestEntries(db, env.parse.AppCode, declaredMenus)
		if err != nil {
			return dto.ManifestDiffResponse{}, err
		}
		for _, change := range changes {
			addChange(change)
		}
	}
	declaredPermissions := map[string]struct{}{}
	for _, op := range env.manifest.Operations {
		declaredPermissions[strings.TrimSpace(defaultString(op.PermissionCode, op.Code))] = struct{}{}
		if op.IncludeInPackage && op.PlatformOnly {
			addChange(conflictChange("OPERATION", op.Code, op.Name, "平台专属操作不能纳入套餐功能点"))
			continue
		}
		if conflict, err := s.permissionPathOwnedByOtherApp(db, env.parse.AppCode, defaultString(op.PermissionCode, op.Code)); err != nil {
			return dto.ManifestDiffResponse{}, err
		} else if conflict {
			addChange(conflictChange("OPERATION", op.Code, op.Name, "权限码已被其他应用占用"))
			continue
		}
		action, err := s.diffByManifestAsset(db, &models.SysAppPermission{}, env.parse.AppCode, "permission_code", op.PermissionCode, env.parse.ManifestHash)
		if err != nil {
			return dto.ManifestDiffResponse{}, err
		}
		addChange(dto.ManifestDiffChange{ResourceType: "OPERATION", ResourceCode: op.PermissionCode, Name: op.Name, Action: action, Severity: severityForManifestAction(action), Message: messageForManifestAction(action, "操作权限")})
	}
	for _, perm := range env.manifest.Permissions {
		declaredPermissions[strings.TrimSpace(perm.Code)] = struct{}{}
		if perm.IncludeInPackage && perm.PlatformOnly {
			addChange(conflictChange("PERMISSION", perm.Code, perm.Name, "平台专属权限不能纳入套餐功能点"))
			continue
		}
		if conflict, err := s.permissionPathOwnedByOtherApp(db, env.parse.AppCode, perm.Code); err != nil {
			return dto.ManifestDiffResponse{}, err
		} else if conflict {
			addChange(conflictChange("PERMISSION", perm.Code, perm.Name, "权限码已被其他应用占用"))
			continue
		}
		action, err := s.diffByManifestAsset(db, &models.SysAppPermission{}, env.parse.AppCode, "permission_code", perm.Code, env.parse.ManifestHash)
		if err != nil {
			return dto.ManifestDiffResponse{}, err
		}
		addChange(dto.ManifestDiffChange{ResourceType: "PERMISSION", ResourceCode: perm.Code, Name: perm.Name, Action: action, Severity: severityForManifestAction(action), Message: messageForManifestAction(action, "权限点")})
	}
	if env.parse.Exists {
		changes, err := s.diffMissingManifestPermissions(db, env.parse.AppCode, declaredPermissions)
		if err != nil {
			return dto.ManifestDiffResponse{}, err
		}
		for _, change := range changes {
			addChange(change)
		}
	}
	declaredAPIs := map[string]struct{}{}
	for _, api := range env.manifest.APIs {
		method := strings.ToUpper(strings.TrimSpace(api.Method))
		path := strings.TrimSpace(api.Path)
		declaredAPIs[method+" "+path] = struct{}{}
		var other models.SysAppAPI
		err := db.Where("method = ? AND path = ? AND app_code <> ? AND deleted_at IS NULL", method, path, env.parse.AppCode).First(&other).Error
		if err == nil {
			addChange(conflictChange("API", method+" "+path, path, "API 路径已被其他应用声明"))
			continue
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ManifestDiffResponse{}, err
		}
		action, err := s.diffByManifestAPI(db, env.parse.AppCode, method, path, env.parse.ManifestHash)
		if err != nil {
			return dto.ManifestDiffResponse{}, err
		}
		addChange(dto.ManifestDiffChange{ResourceType: "API", ResourceCode: method + " " + path, Name: path, Action: action, Severity: severityForManifestAction(action), Message: messageForManifestAction(action, "API 权限矩阵")})
	}
	if env.parse.Exists {
		changes, err := s.diffMissingManifestAPIs(db, env.parse.AppCode, declaredAPIs)
		if err != nil {
			return dto.ManifestDiffResponse{}, err
		}
		for _, change := range changes {
			addChange(change)
		}
	}
	declaredFeatures := map[string]struct{}{}
	for _, feature := range env.manifest.PackageFeatures {
		declaredFeatures[strings.TrimSpace(feature.FeatureCode)] = struct{}{}
		var existing models.SaasFeature
		err := db.Where("feature_code = ?", strings.TrimSpace(feature.FeatureCode)).First(&existing).Error
		if err == nil && existing.AppCode != "" && existing.AppCode != env.parse.AppCode {
			addChange(conflictChange("PACKAGE_FEATURE", feature.FeatureCode, feature.FeatureName, "套餐功能点编码已被其他应用占用"))
			continue
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ManifestDiffResponse{}, err
		}
		action, err := s.diffByManifestAsset(db, &models.SysAppPackageFeature{}, env.parse.AppCode, "feature_code", feature.FeatureCode, env.parse.ManifestHash)
		if err != nil {
			return dto.ManifestDiffResponse{}, err
		}
		addChange(dto.ManifestDiffChange{ResourceType: "PACKAGE_FEATURE", ResourceCode: feature.FeatureCode, Name: feature.FeatureName, Action: action, Severity: severityForManifestAction(action), Message: messageForManifestAction(action, "套餐功能点")})
	}
	if env.parse.Exists {
		changes, err := s.diffMissingManifestPackageFeatures(db, env.parse.AppCode, declaredFeatures)
		if err != nil {
			return dto.ManifestDiffResponse{}, err
		}
		for _, change := range changes {
			addChange(change)
		}
	}
	declaredQuotas := map[string]struct{}{}
	for _, quota := range env.manifest.Quotas {
		declaredQuotas[strings.TrimSpace(quota.QuotaCode)] = struct{}{}
		var existing models.SaasQuota
		err := db.Where("quota_code = ?", strings.TrimSpace(quota.QuotaCode)).First(&existing).Error
		if err == nil && (!strings.EqualFold(existing.QuotaType, defaultString(quota.QuotaType, "STATIC")) || stringValue(existing.Unit) != strings.TrimSpace(quota.Unit)) {
			addChange(conflictChange("QUOTA", quota.QuotaCode, quota.QuotaName, "配额编码已存在且定义不一致"))
			continue
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ManifestDiffResponse{}, err
		}
		action, err := s.diffByManifestAsset(db, &models.SysAppQuota{}, env.parse.AppCode, "quota_code", quota.QuotaCode, env.parse.ManifestHash)
		if err != nil {
			return dto.ManifestDiffResponse{}, err
		}
		addChange(dto.ManifestDiffChange{ResourceType: "QUOTA", ResourceCode: quota.QuotaCode, Name: quota.QuotaName, Action: action, Severity: severityForManifestAction(action), Message: messageForManifestAction(action, "配额定义")})
	}
	if env.parse.Exists {
		changes, err := s.diffMissingManifestQuotas(db, env.parse.AppCode, declaredQuotas)
		if err != nil {
			return dto.ManifestDiffResponse{}, err
		}
		for _, change := range changes {
			addChange(change)
		}
	}

	result.Loadable = len(result.Blockers) == 0
	return result, nil
}

func (s *AppService) manifestEnvelopeFromRequest(ctx context.Context, req dto.ManifestLoadRequest) (manifestEnvelope, error) {
	if len(req.FilePaths) > 0 {
		envs := make([]manifestEnvelope, 0, len(req.FilePaths))
		for _, filePath := range req.FilePaths {
			env, err := s.manifestEnvelopeFromFilePath(ctx, filePath)
			if err != nil {
				return manifestEnvelope{}, err
			}
			envs = append(envs, env)
		}
		merged, blockers, _ := mergeManifestEnvelopesForPreview(envs)
		if len(blockers) > 0 {
			return manifestEnvelope{}, ErrManifestLoadBlocked
		}
		raw, err := json.MarshalIndent(merged.manifest, "", "  ")
		if err != nil {
			return manifestEnvelope{}, err
		}
		merged.content = raw
		merged.parse.FileName = defaultString(req.FileName, merged.parse.AppCode+".manifest.merged.json")
		merged.parse.FilePath = strings.Join(req.FilePaths, ",")
		return merged, nil
	}
	if strings.TrimSpace(req.Content) != "" {
		return s.parseManifestForLoad(ctx, req.FileName, req.FilePath, []byte(req.Content))
	}
	if strings.TrimSpace(req.FilePath) != "" {
		return s.manifestEnvelopeFromFilePath(ctx, req.FilePath)
	}
	return s.parseManifestForLoad(ctx, req.FileName, req.FilePath, nil)
}

func (s *AppService) manifestEnvelopeFromFilePath(ctx context.Context, filePath string) (manifestEnvelope, error) {
	cleanPath := strings.TrimSpace(filePath)
	if cleanPath == "" {
		return manifestEnvelope{}, ErrManifestContentRequired
	}
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return manifestEnvelope{}, err
	}
	raw, err := os.ReadFile(absPath)
	if err != nil {
		return manifestEnvelope{}, err
	}
	return s.parseManifestForLoad(ctx, filepath.Base(absPath), absPath, raw)
}

func (s *AppService) diffMissingManifestEntries(db *gorm.DB, appCode string, declared map[string]struct{}) ([]dto.ManifestDiffChange, error) {
	var rows []models.SysAppEntry
	if err := db.Where("app_code = ? AND managed_by_manifest = ? AND status = ? AND manifest_hash <> ? AND deleted_at IS NULL", appCode, true, "ACTIVE", legacyManifestBaselineHash).Find(&rows).Error; err != nil {
		return nil, err
	}
	changes := []dto.ManifestDiffChange{}
	for _, row := range rows {
		if _, ok := declared[row.ResourceCode]; ok {
			continue
		}
		changes = append(changes, disableChange("MENU", row.ResourceCode, row.Name, "Manifest 已不再声明，装载时应停用或归档，不物理删除"))
	}
	return changes, nil
}

func (s *AppService) diffMissingManifestPermissions(db *gorm.DB, appCode string, declared map[string]struct{}) ([]dto.ManifestDiffChange, error) {
	var rows []models.SysAppPermission
	if err := db.Where("app_code = ? AND managed_by_manifest = ? AND status = ? AND manifest_hash <> ? AND deleted_at IS NULL", appCode, true, "ACTIVE", legacyManifestBaselineHash).Find(&rows).Error; err != nil {
		return nil, err
	}
	changes := []dto.ManifestDiffChange{}
	for _, row := range rows {
		if _, ok := declared[row.PermissionCode]; ok {
			continue
		}
		changes = append(changes, disableChange("PERMISSION", row.PermissionCode, row.Name, "Manifest 已不再声明，装载时应停用或归档，不物理删除"))
	}
	return changes, nil
}

func (s *AppService) diffMissingManifestAPIs(db *gorm.DB, appCode string, declared map[string]struct{}) ([]dto.ManifestDiffChange, error) {
	var rows []models.SysAppAPI
	if err := db.Where("app_code = ? AND managed_by_manifest = ? AND status = ? AND manifest_hash <> ? AND deleted_at IS NULL", appCode, true, "ACTIVE", legacyManifestBaselineHash).Find(&rows).Error; err != nil {
		return nil, err
	}
	changes := []dto.ManifestDiffChange{}
	for _, row := range rows {
		code := strings.ToUpper(strings.TrimSpace(row.Method)) + " " + strings.TrimSpace(row.Path)
		if _, ok := declared[code]; ok {
			continue
		}
		changes = append(changes, disableChange("API", code, row.Path, "Manifest 已不再声明，装载时应停用或归档，不物理删除"))
	}
	return changes, nil
}

func (s *AppService) diffMissingManifestPackageFeatures(db *gorm.DB, appCode string, declared map[string]struct{}) ([]dto.ManifestDiffChange, error) {
	var rows []models.SysAppPackageFeature
	if err := db.Where("app_code = ? AND managed_by_manifest = ? AND status = ? AND manifest_hash <> ? AND deleted_at IS NULL", appCode, true, "ACTIVE", legacyManifestBaselineHash).Find(&rows).Error; err != nil {
		return nil, err
	}
	changes := []dto.ManifestDiffChange{}
	for _, row := range rows {
		if _, ok := declared[row.FeatureCode]; ok {
			continue
		}
		changes = append(changes, disableChange("PACKAGE_FEATURE", row.FeatureCode, row.FeatureName, "Manifest 已不再声明，装载时应移出套餐或归档，不物理删除"))
	}
	return changes, nil
}

func (s *AppService) diffMissingManifestQuotas(db *gorm.DB, appCode string, declared map[string]struct{}) ([]dto.ManifestDiffChange, error) {
	var rows []models.SysAppQuota
	if err := db.Where("app_code = ? AND managed_by_manifest = ? AND status = ? AND manifest_hash <> ? AND deleted_at IS NULL", appCode, true, "ACTIVE", legacyManifestBaselineHash).Find(&rows).Error; err != nil {
		return nil, err
	}
	changes := []dto.ManifestDiffChange{}
	for _, row := range rows {
		if _, ok := declared[row.QuotaCode]; ok {
			continue
		}
		changes = append(changes, disableChange("QUOTA", row.QuotaCode, row.QuotaName, "Manifest 已不再声明，装载时应停用或归档，不物理删除"))
	}
	return changes, nil
}

func (s *AppService) permissionPathOwnedByOtherApp(db *gorm.DB, appCode string, path string) (bool, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return false, nil
	}
	var row models.Permission
	err := db.Where("path = ? AND app_code <> ? AND deleted_at IS NULL", trimmed, appCode).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *AppService) diffByManifestAsset(db *gorm.DB, model interface{}, appCode string, codeColumn string, code string, hash string) (string, error) {
	if strings.TrimSpace(code) == "" {
		return "CONFLICT", nil
	}
	var row struct {
		ManifestHash      string
		ManagedByManifest bool
		ProtectionSource  *string
	}
	err := db.Model(model).
		Select("manifest_hash", "managed_by_manifest", "protection_source").
		Where("app_code = ? AND "+codeColumn+" = ? AND deleted_at IS NULL", appCode, strings.TrimSpace(code)).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "CREATE", nil
	}
	if err != nil {
		return "", err
	}
	if manifestAssetProtected(row.ManagedByManifest, row.ProtectionSource) {
		return "CONFLICT", nil
	}
	if row.ManifestHash == hash {
		return "NO_CHANGE", nil
	}
	return "UPDATE", nil
}

func (s *AppService) diffByManifestAPI(db *gorm.DB, appCode string, method string, path string, hash string) (string, error) {
	var row models.SysAppAPI
	err := db.Select("manifest_hash", "managed_by_manifest", "protection_source").
		Where("app_code = ? AND method = ? AND path = ? AND deleted_at IS NULL", appCode, method, path).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "CREATE", nil
	}
	if err != nil {
		return "", err
	}
	if manifestAssetProtected(row.ManagedByManifest, row.ProtectionSource) {
		return "CONFLICT", nil
	}
	if row.ManifestHash == hash {
		return "NO_CHANGE", nil
	}
	return "UPDATE", nil
}

func (s *AppService) persistManifestFile(tx *gorm.DB, loadID uint64, env manifestEnvelope, filePath string, now time.Time) error {
	summary := fmt.Sprintf("app=%s menus=%d operations=%d permissions=%d apis=%d features=%d quotas=%d",
		env.parse.AppCode,
		env.parse.Counts.Menus,
		env.parse.Counts.Operations,
		env.parse.Counts.Permissions,
		env.parse.Counts.APIs,
		env.parse.Counts.PackageFeatures,
		env.parse.Counts.Quotas,
	)
	file := models.SysAppManifestFile{
		LoadID:          &loadID,
		AppCode:         env.parse.AppCode,
		FileName:        defaultString(env.parse.FileName, "app.manifest.yaml"),
		FilePath:        cleanOptionalString(&filePath),
		FragmentRole:    env.parse.FragmentRole,
		ManifestVersion: env.parse.ManifestVersion,
		ManifestHash:    env.parse.ManifestHash,
		ContentSummary:  &summary,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	return tx.Create(&file).Error
}

func (s *AppService) upsertManifestApp(tx *gorm.DB, env manifestEnvelope, now time.Time) error {
	app := env.manifest.App
	chargeMode := manifestChargeMode(app.ChargePolicy, app.BillingMode)
	commModes := strings.Join(normalizedStringList(app.CommunicationModes), ",")
	row := models.SysApp{
		AppCode:          env.parse.AppCode,
		AppName:          env.parse.AppName,
		Icon:             cleanOptionalString(&app.Icon),
		AppType:          defaultString(app.AppType, "BUSINESS_APP"),
		Source:           defaultString(app.Source, "MANIFEST"),
		Status:           defaultString(app.Status, "INITIATED"),
		ChargeMode:       chargeMode,
		VisibilityScope:  defaultString(app.VisibilityScope, "TENANT"),
		Version:          cleanOptionalString(&app.Version),
		Description:      cleanOptionalString(&app.Description),
		DeploymentMode:   defaultString(app.DeploymentMode, "MERGED"),
		CommModes:        cleanOptionalString(&commModes),
		HealthCheckURL:   cleanOptionalString(&app.HealthCheckURL),
		APIBaseURL:       cleanOptionalString(&app.APIBaseURL),
		WebhookURL:       cleanOptionalString(&app.WebhookURL),
		IsPlatformOnly:   strings.EqualFold(defaultString(app.VisibilityScope, "TENANT"), "PLATFORM_ONLY"),
		ManifestHash:     &env.parse.ManifestHash,
		ManifestVersion:  &env.parse.ManifestVersion,
		LastManifestSync: &now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if strings.EqualFold(row.Source, "BUILTIN") {
		row.IsBuiltin = true
	}
	assignments := map[string]interface{}{
		"app_name":                row.AppName,
		"icon":                    row.Icon,
		"app_type":                row.AppType,
		"source":                  row.Source,
		"status":                  row.Status,
		"charge_mode":             row.ChargeMode,
		"visibility_scope":        row.VisibilityScope,
		"version":                 row.Version,
		"description":             row.Description,
		"deployment_mode":         row.DeploymentMode,
		"communication_modes":     row.CommModes,
		"health_check_url":        row.HealthCheckURL,
		"api_base_url":            row.APIBaseURL,
		"webhook_url":             row.WebhookURL,
		"is_platform_only":        row.IsPlatformOnly,
		"is_builtin":              row.IsBuiltin,
		"manifest_hash":           row.ManifestHash,
		"manifest_version":        row.ManifestVersion,
		"last_manifest_synced_at": row.LastManifestSync,
		"updated_at":              now,
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "app_code"}},
		DoUpdates: clause.Assignments(assignments),
	}).Create(&row).Error
}

func (s *AppService) repoAppByCode(tx *gorm.DB, appCode string) (models.SysApp, error) {
	var row models.SysApp
	err := tx.Where("app_code = ? AND deleted_at IS NULL", appCode).First(&row).Error
	return row, err
}

func (s *AppService) syncManifestClients(tx *gorm.DB, appID uint64, clients []string, now time.Time) error {
	for i, client := range normalizedManifestClients(clients) {
		row := models.SysAppClient{
			AppID:      appID,
			ClientCode: client,
			ClientName: manifestClientName(client),
			Enabled:    true,
			SortOrder:  i + 1,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := upsertManifestClient(tx, row, now); err != nil {
			return err
		}
	}
	return nil
}

func upsertManifestClient(tx *gorm.DB, row models.SysAppClient, now time.Time) error {
	var existing models.SysAppClient
	err := tx.Where("app_id = ? AND client_code = ?", row.AppID, row.ClientCode).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.Create(&row).Error
	}
	if err != nil {
		return err
	}
	return tx.Model(&existing).Updates(map[string]interface{}{
		"client_name": row.ClientName,
		"enabled":     true,
		"sort_order":  row.SortOrder,
		"deleted_at":  nil,
		"updated_at":  now,
	}).Error
}

func (s *AppService) syncManifestAssets(tx *gorm.DB, env manifestEnvelope, now time.Time) error {
	for _, menu := range env.manifest.Menus {
		row := models.SysAppEntry{
			AppCode:           env.parse.AppCode,
			ResourceCode:      strings.TrimSpace(menu.Code),
			Name:              strings.TrimSpace(menu.Name),
			Path:              strings.TrimSpace(menu.Path),
			ParentCode:        cleanOptionalString(&menu.ParentCode),
			SortOrder:         menu.SortOrder,
			PlatformOnly:      menu.PlatformOnly,
			TenantVisible:     menu.TenantVisible,
			TenantEditable:    menu.TenantEditable,
			IncludeInPackage:  menu.IncludeInPackage,
			FeatureCode:       cleanOptionalString(&menu.FeatureCode),
			DataPermMode:      defaultString(menu.DataPermMode, "ORG"),
			ManifestHash:      env.parse.ManifestHash,
			ManagedByManifest: true,
			Status:            "ACTIVE",
			LastSyncedAt:      now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := upsertManifestAsset(tx, row, []clause.Column{{Name: "app_code"}, {Name: "resource_code"}}); err != nil {
			return err
		}
	}
	for _, api := range env.manifest.APIs {
		row := models.SysAppAPI{
			AppCode:           env.parse.AppCode,
			Method:            strings.ToUpper(strings.TrimSpace(api.Method)),
			Path:              strings.TrimSpace(api.Path),
			PermissionCode:    cleanOptionalString(&api.PermissionCode),
			Public:            api.Public,
			Audit:             api.Audit,
			ManifestHash:      env.parse.ManifestHash,
			ManagedByManifest: true,
			Status:            "ACTIVE",
			LastSyncedAt:      now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := upsertManifestAsset(tx, row, []clause.Column{{Name: "app_code"}, {Name: "method"}, {Name: "path"}}); err != nil {
			return err
		}
	}
	for _, op := range env.manifest.Operations {
		row := models.SysAppPermission{
			AppCode:           env.parse.AppCode,
			PermissionCode:    defaultString(op.PermissionCode, op.Code),
			Name:              strings.TrimSpace(op.Name),
			PermissionType:    "OPERATION",
			MenuCode:          cleanOptionalString(&op.MenuCode),
			PlatformOnly:      op.PlatformOnly,
			IncludeInPackage:  op.IncludeInPackage,
			DataPermMode:      "NONE",
			ManifestHash:      env.parse.ManifestHash,
			ManagedByManifest: true,
			Status:            "ACTIVE",
			LastSyncedAt:      now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := upsertManifestAsset(tx, row, []clause.Column{{Name: "app_code"}, {Name: "permission_code"}}); err != nil {
			return err
		}
	}
	for _, perm := range env.manifest.Permissions {
		row := models.SysAppPermission{
			AppCode:           env.parse.AppCode,
			PermissionCode:    strings.TrimSpace(perm.Code),
			Name:              strings.TrimSpace(perm.Name),
			PermissionType:    defaultString(perm.Type, "PERMISSION"),
			MenuCode:          cleanOptionalString(&perm.MenuCode),
			PlatformOnly:      perm.PlatformOnly,
			IncludeInPackage:  perm.IncludeInPackage,
			DataPermMode:      defaultString(perm.DataPermMode, "ORG"),
			ManifestHash:      env.parse.ManifestHash,
			ManagedByManifest: true,
			Status:            "ACTIVE",
			LastSyncedAt:      now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := upsertManifestAsset(tx, row, []clause.Column{{Name: "app_code"}, {Name: "permission_code"}}); err != nil {
			return err
		}
	}
	for _, feature := range env.manifest.PackageFeatures {
		row := models.SysAppPackageFeature{
			AppCode:           env.parse.AppCode,
			FeatureCode:       strings.TrimSpace(feature.FeatureCode),
			FeatureName:       strings.TrimSpace(feature.FeatureName),
			FeatureType:       defaultString(feature.FeatureType, "MENU"),
			ParentCode:        cleanOptionalString(&feature.ParentCode),
			SourceCode:        cleanOptionalString(&feature.SourceCode),
			PackagePolicy:     defaultString(feature.PackagePolicy, "IN_PACKAGE"),
			IncludeInPackage:  feature.IncludeInPackage,
			Description:       cleanOptionalString(&feature.Description),
			ManifestHash:      env.parse.ManifestHash,
			ManagedByManifest: true,
			Status:            "ACTIVE",
			LastSyncedAt:      now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := upsertManifestAsset(tx, row, []clause.Column{{Name: "app_code"}, {Name: "feature_code"}}); err != nil {
			return err
		}
	}
	for _, quota := range env.manifest.Quotas {
		row := models.SysAppQuota{
			AppCode:           env.parse.AppCode,
			QuotaCode:         strings.TrimSpace(quota.QuotaCode),
			QuotaName:         strings.TrimSpace(quota.QuotaName),
			QuotaType:         defaultString(quota.QuotaType, "STATIC"),
			Unit:              cleanOptionalString(&quota.Unit),
			PeriodType:        cleanOptionalString(&quota.PeriodType),
			IncludeInPackage:  quota.IncludeInPackage,
			Description:       cleanOptionalString(&quota.Description),
			ManifestHash:      env.parse.ManifestHash,
			ManagedByManifest: true,
			Status:            "ACTIVE",
			LastSyncedAt:      now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := upsertManifestAsset(tx, row, []clause.Column{{Name: "app_code"}, {Name: "quota_code"}}); err != nil {
			return err
		}
	}
	return nil
}

func upsertManifestAsset[T any](tx *gorm.DB, row T, _ []clause.Column) error {
	switch value := any(row).(type) {
	case models.SysAppEntry:
		var existing models.SysAppEntry
		err := tx.Where("app_code = ? AND resource_code = ? AND deleted_at IS NULL", value.AppCode, value.ResourceCode).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&value).Error
		}
		if err != nil {
			return err
		}
		if manifestAssetProtected(existing.ManagedByManifest, existing.ProtectionSource) {
			return protectedManifestAssetError("菜单入口", value.ResourceCode)
		}
		value.ID = existing.ID
		value.CreatedAt = existing.CreatedAt
		return tx.Save(&value).Error
	case models.SysAppAPI:
		var existing models.SysAppAPI
		err := tx.Where("app_code = ? AND method = ? AND path = ? AND deleted_at IS NULL", value.AppCode, value.Method, value.Path).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&value).Error
		}
		if err != nil {
			return err
		}
		if manifestAssetProtected(existing.ManagedByManifest, existing.ProtectionSource) {
			return protectedManifestAssetError("API", value.Method+" "+value.Path)
		}
		value.ID = existing.ID
		value.CreatedAt = existing.CreatedAt
		return tx.Save(&value).Error
	case models.SysAppPermission:
		var existing models.SysAppPermission
		err := tx.Where("app_code = ? AND permission_code = ? AND deleted_at IS NULL", value.AppCode, value.PermissionCode).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&value).Error
		}
		if err != nil {
			return err
		}
		if manifestAssetProtected(existing.ManagedByManifest, existing.ProtectionSource) {
			return protectedManifestAssetError("权限", value.PermissionCode)
		}
		value.ID = existing.ID
		value.CreatedAt = existing.CreatedAt
		return tx.Save(&value).Error
	case models.SysAppPackageFeature:
		var existing models.SysAppPackageFeature
		err := tx.Where("app_code = ? AND feature_code = ? AND deleted_at IS NULL", value.AppCode, value.FeatureCode).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&value).Error
		}
		if err != nil {
			return err
		}
		if manifestAssetProtected(existing.ManagedByManifest, existing.ProtectionSource) {
			return protectedManifestAssetError("套餐功能点", value.FeatureCode)
		}
		value.ID = existing.ID
		value.CreatedAt = existing.CreatedAt
		return tx.Save(&value).Error
	case models.SysAppQuota:
		var existing models.SysAppQuota
		err := tx.Where("app_code = ? AND quota_code = ? AND deleted_at IS NULL", value.AppCode, value.QuotaCode).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&value).Error
		}
		if err != nil {
			return err
		}
		if manifestAssetProtected(existing.ManagedByManifest, existing.ProtectionSource) {
			return protectedManifestAssetError("配额", value.QuotaCode)
		}
		value.ID = existing.ID
		value.CreatedAt = existing.CreatedAt
		return tx.Save(&value).Error
	default:
		return fmt.Errorf("unsupported Manifest asset type %T", row)
	}
}

func (s *AppService) syncManifestPermissions(tx *gorm.DB, env manifestEnvelope, now time.Time) error {
	platformTenantID, err := platformTenantID(tx)
	if err != nil {
		return err
	}
	menuIDs := map[string]uint64{}
	for _, menu := range env.manifest.Menus {
		path := strings.TrimSpace(menu.Path)
		if path == "" {
			continue
		}
		row := models.Permission{
			TenantID:         platformTenantID,
			Name:             strings.TrimSpace(menu.Name),
			Path:             path,
			PermType:         3,
			SortOrder:        menu.SortOrder,
			Enabled:          true,
			Visible:          menu.TenantVisible,
			IsPlatformOnly:   menu.PlatformOnly,
			IsPackageFeature: menu.IncludeInPackage,
			TenantEditable:   menu.TenantEditable,
			AppCode:          env.parse.AppCode,
			FeatureCode:      cleanOptionalString(&menu.FeatureCode),
			FeatureType:      serviceStringPtr("MENU"),
			DataPermMode:     defaultString(menu.DataPermMode, "ORG"),
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := upsertPermissionByPath(tx, &row); err != nil {
			return err
		}
		menuIDs[menu.Code] = row.ID
	}
	for _, menu := range env.manifest.Menus {
		if strings.TrimSpace(menu.ParentCode) == "" {
			continue
		}
		id := menuIDs[menu.Code]
		parentID := menuIDs[menu.ParentCode]
		if id > 0 && parentID > 0 {
			if err := tx.Model(&models.Permission{}).Where("id = ?", id).Update("parent_id", parentID).Error; err != nil {
				return err
			}
		}
	}
	for _, op := range env.manifest.Operations {
		path := strings.TrimSpace(defaultString(op.PermissionCode, op.Code))
		if path == "" {
			continue
		}
		parentID := menuIDs[op.MenuCode]
		row := models.Permission{
			TenantID:         platformTenantID,
			ParentID:         optionalUint64(parentID),
			Name:             strings.TrimSpace(op.Name),
			Path:             path,
			PermType:         2,
			Enabled:          true,
			Visible:          false,
			IsPlatformOnly:   op.PlatformOnly,
			IsPackageFeature: op.IncludeInPackage,
			AppCode:          env.parse.AppCode,
			FeatureCode:      cleanOptionalString(&op.FeatureCode),
			FeatureType:      serviceStringPtr("OPERATION"),
			DataPermMode:     "NONE",
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := upsertPermissionByPath(tx, &row); err != nil {
			return err
		}
	}
	for _, perm := range env.manifest.Permissions {
		path := strings.TrimSpace(perm.Code)
		if path == "" {
			continue
		}
		parentID := menuIDs[perm.MenuCode]
		row := models.Permission{
			TenantID:         platformTenantID,
			ParentID:         optionalUint64(parentID),
			Name:             strings.TrimSpace(perm.Name),
			Path:             path,
			PermType:         manifestPermType(perm.Type),
			Enabled:          true,
			Visible:          false,
			IsPlatformOnly:   perm.PlatformOnly,
			IsPackageFeature: perm.IncludeInPackage,
			AppCode:          env.parse.AppCode,
			FeatureType:      cleanOptionalString(&perm.Type),
			DataPermMode:     defaultString(perm.DataPermMode, "ORG"),
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := upsertPermissionByPath(tx, &row); err != nil {
			return err
		}
	}
	return nil
}

func upsertPermissionByPath(tx *gorm.DB, row *models.Permission) error {
	var existing models.Permission
	err := tx.Where("tenant_id = ? AND path = ? AND deleted_at IS NULL", row.TenantID, row.Path).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.Create(row).Error
	}
	if err != nil {
		return err
	}
	row.ID = existing.ID
	return tx.Model(&existing).Updates(map[string]interface{}{
		"parent_id":          row.ParentID,
		"name":               row.Name,
		"perm_type":          row.PermType,
		"sort_order":         row.SortOrder,
		"enabled":            row.Enabled,
		"visible":            row.Visible,
		"is_platform_only":   row.IsPlatformOnly,
		"is_package_feature": row.IsPackageFeature,
		"tenant_editable":    row.TenantEditable,
		"app_code":           row.AppCode,
		"feature_code":       row.FeatureCode,
		"feature_type":       row.FeatureType,
		"data_perm_mode":     row.DataPermMode,
		"updated_at":         row.UpdatedAt,
	}).Error
}

func (s *AppService) syncManifestPackageCenter(tx *gorm.DB, env manifestEnvelope, now time.Time) error {
	featureIDByCode := map[string]uint64{}
	for _, feature := range env.manifest.PackageFeatures {
		if !feature.IncludeInPackage {
			continue
		}
		row := models.SaasFeature{
			FeatureCode: strings.TrimSpace(feature.FeatureCode),
			FeatureName: strings.TrimSpace(feature.FeatureName),
			FeatureType: defaultString(feature.FeatureType, "MENU"),
			AppCode:     env.parse.AppCode,
			Status:      1,
			Description: cleanOptionalString(&feature.Description),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if strings.EqualFold(row.FeatureType, "API") {
			row.APIPath = cleanOptionalString(&feature.SourceCode)
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "feature_code"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"feature_name": row.FeatureName,
				"feature_type": row.FeatureType,
				"app_code":     row.AppCode,
				"status":       1,
				"description":  row.Description,
				"updated_at":   now,
			}),
		}).Create(&row).Error; err != nil {
			return err
		}
		var saved models.SaasFeature
		if err := tx.Where("feature_code = ?", row.FeatureCode).First(&saved).Error; err != nil {
			return err
		}
		featureIDByCode[row.FeatureCode] = saved.ID
	}
	for _, feature := range env.manifest.PackageFeatures {
		if !feature.IncludeInPackage || strings.TrimSpace(feature.ParentCode) == "" {
			continue
		}
		id := featureIDByCode[feature.FeatureCode]
		parentID := featureIDByCode[feature.ParentCode]
		if id > 0 && parentID > 0 {
			if err := tx.Model(&models.SaasFeature{}).Where("id = ?", id).Update("parent_id", parentID).Error; err != nil {
				return err
			}
		}
	}
	for _, quota := range env.manifest.Quotas {
		if !quota.IncludeInPackage {
			continue
		}
		row := models.SaasQuota{
			QuotaCode:   strings.TrimSpace(quota.QuotaCode),
			QuotaName:   strings.TrimSpace(quota.QuotaName),
			QuotaType:   defaultString(quota.QuotaType, "STATIC"),
			PeriodType:  cleanOptionalString(&quota.PeriodType),
			Unit:        cleanOptionalString(&quota.Unit),
			Status:      1,
			Description: cleanOptionalString(&quota.Description),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "quota_code"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"quota_name":  row.QuotaName,
				"quota_type":  row.QuotaType,
				"period_type": row.PeriodType,
				"unit":        row.Unit,
				"status":      1,
				"description": row.Description,
				"updated_at":  now,
			}),
		}).Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func markLoadFailed(tx *gorm.DB, loadID uint64, err error) error {
	now := time.Now()
	msg := err.Error()
	_ = tx.Model(&models.SysAppManifestLoad{}).
		Where("id = ?", loadID).
		Updates(map[string]interface{}{"status": "FAILED", "error_summary": msg, "updated_at": now}).Error
	return err
}

func platformTenantID(tx *gorm.DB) (uint64, error) {
	var tenant models.Tenant
	err := tx.Where("is_platform_tenant = ? AND deleted_at IS NULL", true).First(&tenant).Error
	if err == nil {
		return tenant.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	err = tx.Where("code = ? AND deleted_at IS NULL", "platform").First(&tenant).Error
	if err != nil {
		return 0, err
	}
	return tenant.ID, nil
}

func manifestChargeMode(chargePolicy string, billingMode string) string {
	policy := strings.ToUpper(strings.TrimSpace(chargePolicy))
	if policy == "FREE" || policy == "NON_SELLABLE" {
		return policy
	}
	mode := strings.ToUpper(strings.TrimSpace(billingMode))
	if mode == "" {
		return "SUBSCRIPTION"
	}
	return mode
}

func manifestClientName(code string) string {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "PC_WEB":
		return "PC Web"
	case "API_ONLY":
		return "API Only"
	case "H5":
		return "H5"
	case "IOS":
		return "iOS"
	case "ANDROID":
		return "Android"
	case "WEWORK":
		return "企业微信"
	case "DINGTALK":
		return "钉钉"
	case "FEISHU":
		return "飞书"
	case "HARMONYOS":
		return "鸿蒙"
	case "WINDOWS":
		return "Windows"
	case "MACOS":
		return "macOS"
	case "MINI_PROGRAM":
		return "小程序"
	default:
		return strings.ToUpper(strings.TrimSpace(code))
	}
}

func manifestPermType(value string) int {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "MENU":
		return 3
	case "DATA":
		return 4
	default:
		return 2
	}
}

func conflictChange(resourceType string, code string, name string, message string) dto.ManifestDiffChange {
	return dto.ManifestDiffChange{
		ResourceType: resourceType,
		ResourceCode: code,
		Name:         name,
		Action:       "CONFLICT",
		Severity:     "ERROR",
		Message:      message,
	}
}

func disableChange(resourceType string, code string, name string, message string) dto.ManifestDiffChange {
	return dto.ManifestDiffChange{
		ResourceType: resourceType,
		ResourceCode: code,
		Name:         name,
		Action:       "DISABLE",
		Severity:     "WARN",
		Message:      message,
	}
}

func manifestAssetProtected(managedByManifest bool, protectionSource *string) bool {
	source := strings.ToUpper(strings.TrimSpace(stringValue(protectionSource)))
	return !managedByManifest || (source != "" && source != "MANIFEST")
}

func protectedManifestAssetError(resourceType string, code string) error {
	return fmt.Errorf("%s %s 有人工保护标记，Manifest 不允许覆盖", resourceType, code)
}

func severityForManifestAction(action string) string {
	if action == "CONFLICT" {
		return "ERROR"
	}
	return "INFO"
}

func messageForManifestAction(action string, fallback string) string {
	if action == "CONFLICT" {
		return "资源有人工保护标记，Manifest 不允许覆盖"
	}
	return fallback
}

func serviceStringPtr(value string) *string {
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func optionalUint64(value uint64) *uint64 {
	if value == 0 {
		return nil
	}
	return &value
}
