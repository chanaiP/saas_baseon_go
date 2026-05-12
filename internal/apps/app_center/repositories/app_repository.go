package repositories

import (
	"context"
	"strings"
	"time"

	"saas_baseon_go/internal/apps/app_center/dto"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"

	"gorm.io/gorm"
)

type AppRepository struct {
	db *gorm.DB
}

func NewAppRepository(db *gorm.DB) *AppRepository {
	return &AppRepository{db: db}
}

func (r *AppRepository) List(ctx context.Context, req dto.AppListRequest) ([]models.SysApp, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.SysApp{}).Where("deleted_at IS NULL")
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("app_code ILIKE ? OR app_name ILIKE ? OR COALESCE(description, '') ILIKE ?", like, like, like)
	}
	if value := strings.TrimSpace(req.Type); value != "" {
		query = query.Where("app_type = ?", value)
	}
	if value := strings.TrimSpace(req.Status); value != "" {
		statuses := splitStatusFilter(value)
		if len(statuses) == 1 {
			query = query.Where("status = ?", statuses[0])
		} else if len(statuses) > 1 {
			query = query.Where("status IN ?", statuses)
		}
	}
	if value := strings.TrimSpace(req.Source); value != "" {
		query = query.Where("source = ?", value)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []models.SysApp
	err := query.
		Order("sort_order ASC, id ASC").
		Offset(req.Skip).
		Limit(req.Limit).
		Find(&rows).Error
	return rows, total, err
}

func (r *AppRepository) GetByID(ctx context.Context, id uint64) (models.SysApp, error) {
	var row models.SysApp
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&row).Error
	return row, err
}

func (r *AppRepository) GetByCode(ctx context.Context, appCode string) (models.SysApp, error) {
	var row models.SysApp
	err := r.db.WithContext(ctx).Where("app_code = ? AND deleted_at IS NULL", appCode).First(&row).Error
	return row, err
}

func (r *AppRepository) Create(ctx context.Context, row *models.SysApp) error {
	return r.CreateWithClients(ctx, row, nil)
}

func (r *AppRepository) CreateWithClients(ctx context.Context, row *models.SysApp, clients []models.SysAppClient) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		desiredPlatformOnly := row.IsPlatformOnly
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		if !desiredPlatformOnly {
			row.IsPlatformOnly = false
			if err := tx.Model(&models.SysApp{}).
				Where("id = ?", row.ID).
				Update("is_platform_only", false).Error; err != nil {
				return err
			}
		}
		for i := range clients {
			clients[i].AppID = row.ID
			if err := tx.Create(&clients[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *AppRepository) Update(ctx context.Context, row *models.SysApp, updates map[string]interface{}) error {
	if err := r.db.WithContext(ctx).
		Model(row).
		Where("id = ? AND deleted_at IS NULL", row.ID).
		Updates(updates).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", row.ID).First(row).Error
}

func (r *AppRepository) UpdateWithClients(ctx context.Context, row *models.SysApp, updates map[string]interface{}, clients []models.SysAppClient) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(row).
			Where("id = ? AND deleted_at IS NULL", row.ID).
			Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.SysAppClient{}).
			Where("app_id = ? AND deleted_at IS NULL", row.ID).
			Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}
		for i := range clients {
			clients[i].AppID = row.ID
			if err := tx.Create(&clients[i]).Error; err != nil {
				return err
			}
		}
		return tx.Where("id = ? AND deleted_at IS NULL", row.ID).First(row).Error
	})
}

func (r *AppRepository) ClientsByAppID(ctx context.Context, appID uint64) ([]models.SysAppClient, error) {
	var rows []models.SysAppClient
	err := r.db.WithContext(ctx).
		Where("app_id = ? AND deleted_at IS NULL", appID).
		Order("sort_order ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *AppRepository) Stats(ctx context.Context) (dto.AppStatsResponse, error) {
	countApps := func(where string, args ...interface{}) (int64, error) {
		query := r.db.WithContext(ctx).Model(&models.SysApp{}).Where("deleted_at IS NULL")
		if where != "" {
			query = query.Where(where, args...)
		}
		var n int64
		err := query.Count(&n).Error
		return n, err
	}

	total, err := countApps("")
	if err != nil {
		return dto.AppStatsResponse{}, err
	}
	online, err := countApps("status = ?", "ONLINE")
	if err != nil {
		return dto.AppStatsResponse{}, err
	}
	beta, err := countApps("status = ?", "BETA")
	if err != nil {
		return dto.AppStatsResponse{}, err
	}
	developing, err := countApps("status IN ?", []string{"PLANNED", "DEVELOPING"})
	if err != nil {
		return dto.AppStatsResponse{}, err
	}
	builtin, err := countApps("is_builtin = ?", true)
	if err != nil {
		return dto.AppStatsResponse{}, err
	}
	disabled, err := countApps("status = ?", "DISABLED")
	if err != nil {
		return dto.AppStatsResponse{}, err
	}
	var clientApps int64
	if err := r.db.WithContext(ctx).
		Model(&models.SysAppClient{}).
		Where("deleted_at IS NULL AND enabled = ?", true).
		Distinct("app_id").
		Count(&clientApps).Error; err != nil {
		return dto.AppStatsResponse{}, err
	}
	manifestLoads, err := countApps("source = ?", "MANIFEST")
	if err != nil {
		return dto.AppStatsResponse{}, err
	}

	var categories int64
	if err := r.db.WithContext(ctx).
		Model(&models.DictItem{}).
		Joins("JOIN dict_type ON dict_type.id = dict_item.dict_type_id").
		Where("dict_type.code = ? AND dict_type.deleted_at IS NULL AND dict_item.deleted_at IS NULL AND dict_item.enabled = ?", "app_type", true).
		Count(&categories).Error; err != nil {
		return dto.AppStatsResponse{}, err
	}
	if categories == 0 {
		if err := r.db.WithContext(ctx).
			Model(&models.SysApp{}).
			Where("deleted_at IS NULL AND app_type <> ''").
			Distinct("app_type").
			Count(&categories).Error; err != nil {
			return dto.AppStatsResponse{}, err
		}
	}

	var tenantOpenings int64
	if err := r.db.WithContext(ctx).Model(&models.TenantSubscription{}).Count(&tenantOpenings).Error; err != nil {
		return dto.AppStatsResponse{}, err
	}

	var trialInvites int64
	if err := r.db.WithContext(ctx).
		Model(&models.TenantSubscription{}).
		Where("UPPER(subscription_status) LIKE ? OR trial_end_time IS NOT NULL", "%TRIAL%").
		Count(&trialInvites).Error; err != nil {
		return dto.AppStatsResponse{}, err
	}

	var auditLogs int64
	if err := r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("module IN ?", []string{"app", "app_center", "application"}).
		Count(&auditLogs).Error; err != nil {
		return dto.AppStatsResponse{}, err
	}

	return dto.AppStatsResponse{
		Total:          total,
		Online:         online,
		Beta:           beta,
		Developing:     developing,
		Builtin:        builtin,
		Disabled:       disabled,
		Categories:     categories,
		ClientApps:     clientApps,
		TenantOpenings: tenantOpenings,
		TrialInvites:   trialInvites,
		ManifestLoads:  manifestLoads,
		AuditLogs:      auditLogs,
	}, nil
}

func splitStatusFilter(value string) []string {
	parts := strings.Split(value, ",")
	statuses := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		status := strings.ToUpper(strings.TrimSpace(part))
		if status == "" {
			continue
		}
		if _, ok := seen[status]; ok {
			continue
		}
		seen[status] = struct{}{}
		statuses = append(statuses, status)
	}
	return statuses
}

func (r *AppRepository) UserByID(ctx context.Context, userID uint64) (models.AppUser, error) {
	var user models.AppUser
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	return user, err
}
