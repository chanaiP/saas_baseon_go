package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/apps/ai_geo/dto"
	"saas_baseon_go/internal/apps/ai_geo/repositories"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestBrandsAreTenantScoped(t *testing.T) {
	db := newAiGeoTestDB(t)
	now := time.Now()
	require.NoError(t, db.Create(&models.AiGeoBrandCard{TenantID: 1, BrandCode: "B1", BrandName: "租户一品牌", Keywords: "[]", Status: "active", CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, db.Create(&models.AiGeoBrandCard{TenantID: 2, BrandCode: "B2", BrandName: "租户二品牌", Keywords: "[]", Status: "active", CreatedAt: now, UpdatedAt: now}).Error)

	result, err := NewService(repositories.NewRepository(db)).Brands(context.Background(), dto.Viewer{TenantID: 1, UserID: 10}, dto.PageRequest{Limit: 20})

	require.NoError(t, err)
	require.Equal(t, int64(1), result.Total)
	require.Len(t, result.Items, 1)
	require.Equal(t, "B1", result.Items[0].BrandCode)
}

func TestDraftReviewStateMachine(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}

	draft, err := service.CreateDraft(context.Background(), viewer, dto.DraftPayload{Title: "新品种草", Body: "正文"})
	require.NoError(t, err)
	require.Equal(t, "draft", draft.AuditStatus)

	submitted, err := service.SubmitDraft(context.Background(), viewer, draft.ID)
	require.NoError(t, err)
	require.Equal(t, "pending", submitted.AuditStatus)

	approved, err := service.ReviewDraft(context.Background(), viewer, draft.ID, true)
	require.NoError(t, err)
	require.Equal(t, "approved", approved.AuditStatus)

	_, err = service.ReviewDraft(context.Background(), viewer, draft.ID, false)
	require.ErrorIs(t, err, ErrInvalidStatus)
}

func TestDraftCannotCrossTenant(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	draft, err := service.CreateDraft(context.Background(), dto.Viewer{TenantID: 2, UserID: 20}, dto.DraftPayload{Title: "其他租户", Body: "正文"})
	require.NoError(t, err)

	_, err = service.SubmitDraft(context.Background(), dto.Viewer{TenantID: 1, UserID: 10}, draft.ID)

	require.ErrorIs(t, err, ErrNotFound)
}

func TestRecordAuditStoresAiGeoContext(t *testing.T) {
	db := newAiGeoTestDB(t)
	service := NewService(repositories.NewRepository(db))
	viewer := dto.Viewer{TenantID: 1, UserID: 10}

	err := service.RecordAudit(context.Background(), viewer, dto.RequestMeta{IP: "127.0.0.1", UserAgent: "ai-geo-test", RequestID: "rid-ai-geo"}, "draft_create", "DRAFT-1", "创建母稿", map[string]interface{}{"title": "新品"})

	require.NoError(t, err)
	var log models.AuditLog
	require.NoError(t, db.Where("tenant_id = ? AND user_id = ? AND module = ? AND action = ?", viewer.TenantID, viewer.UserID, "ai_geo", "draft_create").First(&log).Error)
	require.NotNil(t, log.AppCode)
	require.Equal(t, "ai-geo", *log.AppCode)
	require.NotNil(t, log.Detail)
	require.Contains(t, *log.Detail, `"title":"新品"`)
}

func newAiGeoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(
		&models.AiGeoBrandCard{},
		&models.AiGeoProductCard{},
		&models.AiGeoSKU{},
		&models.AiGeoCompetitor{},
		&models.AiGeoChannelProfile{},
		&models.AiGeoChannelAccount{},
		&models.AiGeoDraft{},
		&models.AiGeoChannelContent{},
		&models.AiGeoPublishPlan{},
		&models.AiGeoImportBatch{},
		&models.AiGeoMaterialAsset{},
		&models.AiGeoHotspot{},
		&models.AuditLog{},
	))
	return db
}
