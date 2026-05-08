package file

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestCreateMetadataPersistsTenantScopedFile(t *testing.T) {
	db := newFileServiceTestDB(t)
	service := NewService(db)
	now := time.Now()

	row, err := service.CreateMetadata(context.Background(), CreateMetadataCommand{
		TenantID: 1, FileID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", CreatedBy: 7,
		OriginalName: "report.pdf", StoredName: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.pdf",
		StoragePath: "/tmp/report.pdf", MimeType: "application/pdf", FileSize: 12, Now: now,
	})

	require.NoError(t, err)
	require.NotZero(t, row.ID)
	require.Equal(t, uint64(1), row.TenantID)
	require.Equal(t, "application/pdf", row.MimeType)
	var auditCount int64
	require.NoError(t, db.Model(&models.AuditLog{}).Where("tenant_id = ? AND module = ? AND action = ?", 1, "file", "upload").Count(&auditCount).Error)
	require.Equal(t, int64(1), auditCount)
}

func TestSoftDeleteMarksFileDeleted(t *testing.T) {
	db := newFileServiceTestDB(t)
	service := NewService(db)
	row, err := service.CreateMetadata(context.Background(), CreateMetadataCommand{
		TenantID: 1, FileID: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", CreatedBy: 7,
		OriginalName: "report.pdf", StoredName: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.pdf",
		StoragePath: "/tmp/report.pdf", MimeType: "application/pdf", FileSize: 12,
	})
	require.NoError(t, err)

	require.NoError(t, service.SoftDelete(context.Background(), &row, "/tmp/.trash/report.pdf", 7, time.Now()))

	var stored models.FileObject
	require.NoError(t, db.First(&stored, row.ID).Error)
	require.Equal(t, 0, stored.Status)
	require.NotNil(t, stored.DeletedAt)
	require.Equal(t, "/tmp/.trash/report.pdf", stored.StoragePath)
	var auditCount int64
	require.NoError(t, db.Model(&models.AuditLog{}).Where("tenant_id = ? AND module = ? AND action = ?", 1, "file", "delete").Count(&auditCount).Error)
	require.Equal(t, int64(1), auditCount)
}

func TestValidateUploadRejectsForgedContentType(t *testing.T) {
	service := NewService(nil)
	header := multipartUploadFileHeader(t, "note.txt", "image/jpeg", []byte("plain text"))

	_, err := service.ValidateUpload("note.txt", header)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Content-Type")
}

func TestValidateUploadRejectsZipPathTraversal(t *testing.T) {
	service := NewService(nil)
	var data bytes.Buffer
	zw := zip.NewWriter(&data)
	w, err := zw.Create("../evil.txt")
	require.NoError(t, err)
	_, err = w.Write([]byte("bad"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	header := multipartUploadFileHeader(t, "archive.zip", "application/zip", data.Bytes())

	_, err = service.ValidateUpload("archive.zip", header)

	require.Error(t, err)
	require.Contains(t, err.Error(), "路径")
}

func newFileServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(&models.FileObject{}, &models.AuditLog{}))
	return db
}

func multipartUploadFileHeader(t *testing.T, filename string, contentType string, content []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	reader := multipart.NewReader(&body, writer.Boundary())
	form, err := reader.ReadForm(int64(len(content) + 1024))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, form.RemoveAll()) })
	require.Len(t, form.File["file"], 1)
	return form.File["file"][0]
}
