package dictionary

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/repositories"
)

var (
	ErrNotFound          = errors.New("不存在")
	ErrDictItemNotFound  = errors.New("字典项不存在")
	ErrTenantOverrideOff = errors.New("该字典不允许租户覆盖")
	ErrDictTypeNotFound  = errors.New("字典类型不存在")
	ErrPlatformOnly      = errors.New("字典类型维护仅允许平台管理员操作")
)

type Viewer struct {
	TenantID            uint64
	IsPlatformAdmin     bool
	IncludePlatformOnly bool
	ScopeTenantIDs      []uint64
}

type Page[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Skip  int   `json:"skip"`
	Limit int   `json:"limit"`
}

type DictTypeResponse struct {
	ID             uint64  `json:"id"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	TypeCode       string  `json:"type_code"`
	TypeName       string  `json:"type_name"`
	Remark         *string `json:"remark"`
	Scope          string  `json:"scope"`
	TenantEditable bool    `json:"tenant_editable"`
	IsPlatformOnly bool    `json:"is_platform_only"`
	Status         int     `json:"status"`
}

type DictItemResponse struct {
	ID               uint64  `json:"id"`
	DictTypeID       uint64  `json:"dict_type_id"`
	ParentID         *uint64 `json:"parent_id"`
	DefaultLabel     string  `json:"default_label"`
	DefaultValue     string  `json:"default_value"`
	DefaultSortOrder int     `json:"default_sort_order"`
	DefaultEnabled   bool    `json:"default_enabled"`
	Label            string  `json:"label"`
	Value            string  `json:"value"`
	ItemLabel        string  `json:"item_label"`
	ItemValue        string  `json:"item_value"`
	SortOrder        int     `json:"sort_order"`
	Enabled          bool    `json:"enabled"`
	Status           int     `json:"status"`
	IsOverride       bool    `json:"is_override"`
	IsCustom         bool    `json:"is_custom"`
}

type ListTypesQuery struct {
	Keyword      string
	PlatformOnly *bool
	Skip         int
	Limit        int
}

type CreateTypeCommand struct {
	Code           string
	Name           string
	Remark         *string
	Scope          string
	TenantEditable bool
	IsPlatformOnly bool
}

type UpdateTypeCommand struct {
	Name           *string
	Remark         *string
	Scope          *string
	TenantEditable *bool
	IsPlatformOnly *bool
}

type CreateItemCommand struct {
	DictTypeID uint64
	ParentID   *uint64
	Label      string
	Value      string
	SortOrder  int
	Enabled    *bool
}

type UpdateItemCommand struct {
	ParentID  *uint64
	HasParent bool
	Label     *string
	Value     *string
	SortOrder *int
	Enabled   *bool
}

type Service struct {
	repo *repositories.DictionaryRepository
}

func NewService(repo *repositories.DictionaryRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListTypes(ctx context.Context, viewer Viewer, query ListTypesQuery) (Page[DictTypeResponse], error) {
	query.Keyword = strings.TrimSpace(query.Keyword)
	normalizePage(&query.Skip, &query.Limit)
	rows, total, err := s.repo.ListTypes(ctx, repositories.DictionaryTypeListQuery{
		TenantID:            viewer.TenantID,
		ScopeTenantIDs:      scopeTenantIDs(viewer),
		IncludePlatformOnly: viewer.IncludePlatformOnly,
		Keyword:             query.Keyword,
		PlatformOnly:        query.PlatformOnly,
		Skip:                query.Skip,
		Limit:               query.Limit,
	})
	if err != nil {
		return Page[DictTypeResponse]{}, err
	}
	items := make([]DictTypeResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, TypeToResponse(row))
	}
	return Page[DictTypeResponse]{Items: items, Total: total, Skip: query.Skip, Limit: query.Limit}, nil
}

func (s *Service) CreateType(ctx context.Context, viewer Viewer, cmd CreateTypeCommand) (models.DictType, error) {
	if !viewer.IsPlatformAdmin {
		return models.DictType{}, ErrPlatformOnly
	}
	scope := strings.TrimSpace(cmd.Scope)
	if scope == "" {
		scope = "platform"
	}
	row := models.DictType{
		TenantID:       viewer.TenantID,
		Code:           strings.TrimSpace(cmd.Code),
		Name:           strings.TrimSpace(cmd.Name),
		Remark:         nullableTrimmed(cmd.Remark),
		Scope:          scope,
		TenantEditable: cmd.TenantEditable,
		IsPlatformOnly: cmd.IsPlatformOnly,
	}
	if err := s.repo.CreateType(ctx, &row); err != nil {
		return models.DictType{}, err
	}
	return row, nil
}

func (s *Service) UpdateType(ctx context.Context, viewer Viewer, id uint64, cmd UpdateTypeCommand) (models.DictType, error) {
	if !viewer.IsPlatformAdmin {
		return models.DictType{}, ErrPlatformOnly
	}
	row, err := s.repo.GetOwnType(ctx, viewer.TenantID, id)
	if err != nil {
		return models.DictType{}, ErrNotFound
	}
	updates := map[string]interface{}{}
	if cmd.Name != nil {
		updates["name"] = strings.TrimSpace(*cmd.Name)
	}
	if cmd.Remark != nil {
		updates["remark"] = nullableTrimmed(cmd.Remark)
	}
	if cmd.Scope != nil {
		updates["scope"] = strings.TrimSpace(*cmd.Scope)
	}
	if cmd.TenantEditable != nil {
		updates["tenant_editable"] = *cmd.TenantEditable
	}
	if cmd.IsPlatformOnly != nil {
		updates["is_platform_only"] = *cmd.IsPlatformOnly
	}
	if err := s.repo.UpdateType(ctx, &row, updates); err != nil {
		return models.DictType{}, err
	}
	return row, nil
}

func (s *Service) DeleteType(ctx context.Context, viewer Viewer, id uint64) (models.DictType, error) {
	if !viewer.IsPlatformAdmin {
		return models.DictType{}, ErrPlatformOnly
	}
	count, err := s.repo.CountItemsForOwnType(ctx, viewer.TenantID, id)
	if err != nil {
		return models.DictType{}, err
	}
	if count > 0 {
		return models.DictType{}, errors.New("字典类型已被字典项引用，不能删除")
	}
	row, err := s.repo.GetOwnType(ctx, viewer.TenantID, id)
	if err != nil {
		return models.DictType{}, ErrNotFound
	}
	now := time.Now()
	updates := map[string]interface{}{
		"deleted_at": now,
		"code":       tombstoneUniqueValue(row.Code, row.ID, 64),
	}
	if err := s.repo.UpdateType(ctx, &row, updates); err != nil {
		return models.DictType{}, err
	}
	return row, nil
}

func (s *Service) ItemsByCode(ctx context.Context, viewer Viewer, code string) (string, []DictItemResponse, error) {
	code = strings.TrimSpace(code)
	dictType, err := s.repo.VisibleTypeByCode(ctx, repositories.VisibleDictionaryTypeQuery{
		TenantID:            viewer.TenantID,
		ScopeTenantIDs:      scopeTenantIDs(viewer),
		IncludePlatformOnly: viewer.IncludePlatformOnly,
		Code:                code,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code, []DictItemResponse{}, nil
		}
		return code, nil, err
	}
	rows, err := s.repo.ListVisibleItemsByType(ctx, viewer.TenantID, dictType.TenantID, dictType.ID, 0, 0)
	if err != nil {
		return code, nil, err
	}
	return code, s.itemsToResponse(ctx, viewer.TenantID, rows, true), nil
}

func (s *Service) ListItems(ctx context.Context, viewer Viewer, dictTypeID uint64, skip int, limit int) (Page[DictItemResponse], error) {
	normalizePage(&skip, &limit)
	dictType, err := s.repo.VisibleTypeByID(ctx, repositories.VisibleDictionaryTypeQuery{
		TenantID:            viewer.TenantID,
		ScopeTenantIDs:      scopeTenantIDs(viewer),
		IncludePlatformOnly: viewer.IncludePlatformOnly,
		ID:                  dictTypeID,
	})
	if err != nil {
		return Page[DictItemResponse]{Items: []DictItemResponse{}, Total: 0, Skip: skip, Limit: limit}, nil
	}
	rows, total, err := s.repo.ListVisibleItemsByTypePage(ctx, viewer.TenantID, dictType.TenantID, dictType.ID, skip, limit)
	if err != nil {
		return Page[DictItemResponse]{}, err
	}
	return Page[DictItemResponse]{Items: s.itemsToResponse(ctx, viewer.TenantID, rows, false), Total: total, Skip: skip, Limit: limit}, nil
}

func (s *Service) CreateItem(ctx context.Context, viewer Viewer, cmd CreateItemCommand) (models.DictItem, error) {
	dictType, err := s.repo.VisibleTypeByID(ctx, repositories.VisibleDictionaryTypeQuery{
		TenantID:            viewer.TenantID,
		ScopeTenantIDs:      scopeTenantIDs(viewer),
		IncludePlatformOnly: viewer.IncludePlatformOnly,
		ID:                  cmd.DictTypeID,
	})
	if err != nil {
		return models.DictItem{}, ErrDictTypeNotFound
	}
	if !viewer.IsPlatformAdmin && dictType.TenantID != viewer.TenantID && !dictType.TenantEditable {
		return models.DictItem{}, ErrTenantOverrideOff
	}
	if cmd.ParentID != nil {
		if err := s.repo.AssertVisibleItemInType(ctx, viewer.TenantID, dictType.TenantID, cmd.DictTypeID, *cmd.ParentID); err != nil {
			return models.DictItem{}, errors.New("上级字典项不存在")
		}
	}
	enabled := true
	if cmd.Enabled != nil {
		enabled = *cmd.Enabled
	}
	row := models.DictItem{
		TenantID:   viewer.TenantID,
		DictTypeID: cmd.DictTypeID,
		ParentID:   cmd.ParentID,
		Label:      strings.TrimSpace(cmd.Label),
		Value:      strings.TrimSpace(cmd.Value),
		SortOrder:  cmd.SortOrder,
		Enabled:    enabled,
	}
	if err := s.repo.CreateItem(ctx, &row); err != nil {
		return models.DictItem{}, err
	}
	return row, nil
}

func (s *Service) UpdateItem(ctx context.Context, viewer Viewer, id uint64, cmd UpdateItemCommand) (models.DictItem, *models.TenantDictItemOverride, error) {
	row, err := s.repo.GetItem(ctx, id)
	if err != nil {
		return models.DictItem{}, nil, ErrNotFound
	}
	dictType, err := s.repo.VisibleTypeByID(ctx, repositories.VisibleDictionaryTypeQuery{
		TenantID:            viewer.TenantID,
		ScopeTenantIDs:      scopeTenantIDs(viewer),
		IncludePlatformOnly: viewer.IncludePlatformOnly,
		ID:                  row.DictTypeID,
	})
	if err != nil || dictType.ID == 0 || dictType.ID != row.DictTypeID {
		return models.DictItem{}, nil, ErrNotFound
	}
	if !viewer.IsPlatformAdmin && row.TenantID != viewer.TenantID {
		if row.TenantID != dictType.TenantID {
			return models.DictItem{}, nil, ErrNotFound
		}
		if !dictType.TenantEditable {
			return models.DictItem{}, nil, ErrTenantOverrideOff
		}
		override, err := s.repo.UpsertOverride(ctx, viewer.TenantID, row.ID, cmd.Label, cmd.Value, cmd.SortOrder, cmd.Enabled)
		return row, override, err
	}
	if !viewer.IsPlatformAdmin && row.TenantID != viewer.TenantID {
		return models.DictItem{}, nil, ErrNotFound
	}
	updates := itemUpdates(cmd)
	if parent, ok := updates["parent_id"]; ok {
		if parentID, ok := parent.(*uint64); ok && parentID != nil {
			if *parentID == row.ID {
				return models.DictItem{}, nil, errors.New("上级字典项不能选择自身")
			}
			if err := s.repo.AssertVisibleItemInType(ctx, viewer.TenantID, dictType.TenantID, row.DictTypeID, *parentID); err != nil {
				return models.DictItem{}, nil, errors.New("上级字典项不存在")
			}
			if err := s.ensureParentDoesNotCreateCycle(ctx, viewer.TenantID, dictType.TenantID, row.DictTypeID, row.ID, *parentID); err != nil {
				return models.DictItem{}, nil, err
			}
		}
	}
	if err := s.repo.UpdateItem(ctx, &row, updates); err != nil {
		return models.DictItem{}, nil, err
	}
	return row, nil, nil
}

func (s *Service) DeleteItem(ctx context.Context, viewer Viewer, id uint64) (models.DictItem, error) {
	row, err := s.repo.GetItem(ctx, id)
	if err != nil {
		return models.DictItem{}, ErrNotFound
	}
	if row.TenantID != viewer.TenantID {
		return models.DictItem{}, ErrNotFound
	}
	if count, err := s.repo.CountChildItems(ctx, row.ID); err != nil {
		return models.DictItem{}, err
	} else if count > 0 {
		return models.DictItem{}, errors.New("该字典项存在下级项，不能删除")
	}
	now := time.Now()
	if err := s.repo.UpdateItem(ctx, &row, map[string]interface{}{"deleted_at": now, "enabled": false}); err != nil {
		return models.DictItem{}, err
	}
	return row, nil
}

func (s *Service) RestoreItem(ctx context.Context, viewer Viewer, id uint64) (models.DictItem, error) {
	row, err := s.repo.GetItem(ctx, id)
	if err != nil {
		return models.DictItem{}, ErrDictItemNotFound
	}
	dictType, err := s.repo.VisibleTypeByID(ctx, repositories.VisibleDictionaryTypeQuery{
		TenantID:            viewer.TenantID,
		ScopeTenantIDs:      scopeTenantIDs(viewer),
		IncludePlatformOnly: viewer.IncludePlatformOnly,
		ID:                  row.DictTypeID,
	})
	if err != nil || row.TenantID != dictType.TenantID {
		return models.DictItem{}, ErrDictItemNotFound
	}
	if err := s.repo.DeleteOverride(ctx, viewer.TenantID, row.ID); err != nil {
		return models.DictItem{}, err
	}
	return row, nil
}

func TypeToResponse(row models.DictType) DictTypeResponse {
	return DictTypeResponse{ID: row.ID, Code: row.Code, Name: row.Name, TypeCode: row.Code, TypeName: row.Name, Remark: row.Remark, Scope: row.Scope, TenantEditable: row.TenantEditable, IsPlatformOnly: row.IsPlatformOnly, Status: 1}
}

func ItemToResponse(row models.DictItem, override *models.TenantDictItemOverride, viewerTenantID uint64) DictItemResponse {
	label := row.Label
	value := row.Value
	sortOrder := row.SortOrder
	enabled := row.Enabled
	isOverride := override != nil
	if override != nil {
		if override.CustomLabel != nil {
			label = *override.CustomLabel
		}
		if override.CustomValue != nil {
			value = *override.CustomValue
		}
		if override.SortOrder != nil {
			sortOrder = *override.SortOrder
		}
		if override.Enabled != nil {
			enabled = *override.Enabled
		}
	}
	return DictItemResponse{ID: row.ID, DictTypeID: row.DictTypeID, ParentID: row.ParentID, DefaultLabel: row.Label, DefaultValue: row.Value, DefaultSortOrder: row.SortOrder, DefaultEnabled: row.Enabled, Label: label, Value: value, ItemLabel: label, ItemValue: value, SortOrder: sortOrder, Enabled: enabled, Status: boolToStatus(enabled), IsOverride: isOverride, IsCustom: row.TenantID == viewerTenantID}
}

func (s *Service) itemsToResponse(ctx context.Context, tenantID uint64, rows []models.DictItem, enabledOnly bool) []DictItemResponse {
	items := make([]DictItemResponse, 0, len(rows))
	for _, row := range rows {
		override, _ := s.repo.GetOverride(ctx, tenantID, row.ID)
		item := ItemToResponse(row, override, tenantID)
		if enabledOnly && !item.Enabled {
			continue
		}
		items = append(items, item)
	}
	return items
}

func scopeTenantIDs(viewer Viewer) []uint64 {
	if len(viewer.ScopeTenantIDs) > 0 {
		return viewer.ScopeTenantIDs
	}
	return []uint64{viewer.TenantID}
}

func normalizePage(skip *int, limit *int) {
	if *skip < 0 {
		*skip = 0
	}
	if *limit <= 0 || *limit > 200 {
		*limit = 50
	}
}

func itemUpdates(cmd UpdateItemCommand) map[string]interface{} {
	updates := map[string]interface{}{}
	if cmd.HasParent {
		updates["parent_id"] = cmd.ParentID
	}
	if cmd.Label != nil {
		updates["label"] = strings.TrimSpace(*cmd.Label)
	}
	if cmd.Value != nil {
		updates["value"] = strings.TrimSpace(*cmd.Value)
	}
	if cmd.SortOrder != nil {
		updates["sort_order"] = *cmd.SortOrder
	}
	if cmd.Enabled != nil {
		updates["enabled"] = *cmd.Enabled
	}
	return updates
}

func (s *Service) ensureParentDoesNotCreateCycle(ctx context.Context, viewerTenantID uint64, baseTenantID uint64, dictTypeID uint64, currentID uint64, parentID uint64) error {
	rows, err := s.repo.ListVisibleItemsByType(ctx, viewerTenantID, baseTenantID, dictTypeID, 0, 0)
	if err != nil {
		return err
	}
	parentMap := make(map[uint64]*uint64, len(rows))
	for _, item := range rows {
		parentMap[item.ID] = item.ParentID
	}
	visited := map[uint64]bool{}
	cursor := parentID
	for {
		if cursor == currentID {
			return errors.New("上级字典项不能选择自己的下级项")
		}
		if visited[cursor] {
			return errors.New("字典项层级存在循环引用")
		}
		visited[cursor] = true
		next := parentMap[cursor]
		if next == nil {
			return nil
		}
		cursor = *next
	}
}

func nullableTrimmed(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func boolToStatus(value bool) int {
	if value {
		return 1
	}
	return 0
}

func tombstoneUniqueValue(value string, id uint64, maxLen int) string {
	suffix := fmt.Sprintf("__deleted_%d", id)
	if maxLen <= 0 {
		return value + suffix
	}
	limit := maxLen - len(suffix)
	if limit < 1 {
		limit = 1
	}
	base := value
	if len(base) > limit {
		base = base[:limit]
	}
	return base + suffix
}
