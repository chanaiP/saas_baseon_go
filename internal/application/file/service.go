package file

import (
	"context"
	"time"

	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type CreateMetadataCommand struct {
	TenantID     uint64
	FileID       string
	CreatedBy    uint64
	OriginalName string
	StoredName   string
	StoragePath  string
	MimeType     string
	FileSize     int64
	Now          time.Time
}

func (s *Service) CreateMetadata(ctx context.Context, cmd CreateMetadataCommand) (models.FileObject, error) {
	now := cmd.Now
	if now.IsZero() {
		now = time.Now()
	}
	row := models.FileObject{
		TenantID:     cmd.TenantID,
		FileID:       cmd.FileID,
		CreatedBy:    cmd.CreatedBy,
		OriginalName: cmd.OriginalName,
		StoredName:   cmd.StoredName,
		StoragePath:  cmd.StoragePath,
		MimeType:     cmd.MimeType,
		FileSize:     cmd.FileSize,
		Status:       1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	err := s.db.WithContext(ctx).Create(&row).Error
	return row, err
}

func (s *Service) SoftDelete(ctx context.Context, row *models.FileObject, trashPath string, now time.Time) error {
	if now.IsZero() {
		now = time.Now()
	}
	return s.db.WithContext(ctx).Model(row).Updates(map[string]interface{}{"status": 0, "deleted_at": now, "updated_at": now, "storage_path": trashPath}).Error
}
