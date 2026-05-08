package file

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type Service struct {
	db *gorm.DB
}

type VirusScanner interface {
	ScanUpload(originalName string, content []byte) error
}

type noopVirusScanner struct{}

func (noopVirusScanner) ScanUpload(string, []byte) error { return nil }

var defaultVirusScanner VirusScanner = noopVirusScanner{}

type UploadValidation struct {
	Ext      string
	MimeType string
}

var allowedExtensions = map[string]struct{}{
	".csv": {}, ".doc": {}, ".docx": {}, ".gif": {}, ".jpeg": {}, ".jpg": {}, ".pdf": {}, ".png": {},
	".ppt": {}, ".pptx": {}, ".txt": {}, ".webp": {}, ".xls": {}, ".xlsx": {}, ".zip": {},
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ValidateUpload(originalName string, file *multipart.FileHeader) (UploadValidation, error) {
	ext := safeExt(originalName)
	if _, ok := allowedExtensions[ext]; !ok {
		return UploadValidation{}, errors.New("不支持的文件类型")
	}
	clientType := normalizeContentType(file.Header.Get("Content-Type"))
	if highRiskContentType(clientType) {
		return UploadValidation{}, errors.New("不允许上传高风险文件类型")
	}
	content, err := readUploadContent(file, 50*1024*1024)
	if err != nil {
		return UploadValidation{}, errors.New("读取文件失败")
	}
	detectedType := sniffContentType(content)
	if highRiskContentType(detectedType) {
		return UploadValidation{}, errors.New("不允许上传高风险文件类型")
	}
	if !extensionMatchesContent(ext, detectedType, content) {
		return UploadValidation{}, errors.New("文件扩展名与实际内容不一致")
	}
	if !clientTypeAllowed(ext, clientType) {
		return UploadValidation{}, errors.New("文件 Content-Type 与扩展名不一致")
	}
	if zipBacked(ext) {
		if err := validateZip(content); err != nil {
			return UploadValidation{}, err
		}
	}
	if err := defaultVirusScanner.ScanUpload(originalName, content); err != nil {
		return UploadValidation{}, errors.New("文件安全扫描未通过")
	}
	return UploadValidation{Ext: ext, MimeType: detectedType}, nil
}

func (s *Service) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
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
	AuditUserID  uint64
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
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		return recordAudit(tx, row.TenantID, cmd.AuditUserID, "upload", "上传文件 "+row.OriginalName, map[string]interface{}{"file_id": row.FileID, "file_name": row.OriginalName, "size": row.FileSize}, now)
	})
	return row, err
}

func (s *Service) SoftDelete(ctx context.Context, row *models.FileObject, trashPath string, actorID uint64, now time.Time) error {
	if now.IsZero() {
		now = time.Now()
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(row).Updates(map[string]interface{}{"status": 0, "deleted_at": now, "updated_at": now, "storage_path": trashPath}).Error; err != nil {
			return err
		}
		return recordAudit(tx, row.TenantID, actorID, "delete", "删除文件 "+row.FileID, map[string]interface{}{"file_id": row.FileID, "trash_path": trashPath}, now)
	})
}

func (s *Service) Delete(ctx context.Context, row *models.FileObject, trashPath string, actorID uint64, now time.Time) error {
	originalPath := row.StoragePath
	if err := os.MkdirAll(filepath.Dir(trashPath), 0o750); err != nil {
		return err
	}
	if err := os.Rename(originalPath, trashPath); err != nil {
		return err
	}
	if err := s.SoftDelete(ctx, row, trashPath, actorID, now); err != nil {
		_ = os.Rename(trashPath, originalPath)
		return err
	}
	return nil
}

func recordAudit(db *gorm.DB, tenantID uint64, userID uint64, action string, summary string, detail map[string]interface{}, now time.Time) error {
	raw, _ := json.Marshal(detail)
	text := string(raw)
	return db.Create(&models.AuditLog{
		TenantID:  &tenantID,
		UserID:    &userID,
		Module:    "file",
		Action:    action,
		Summary:   summary,
		Detail:    &text,
		Result:    "success",
		CreatedAt: now,
	}).Error
}

func safeExt(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if len(ext) == 0 || len(ext) > 16 {
		return ".bin"
	}
	for _, r := range ext[1:] {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') {
			return ".bin"
		}
	}
	return ext
}

func readUploadContent(file *multipart.FileHeader, maxBytes int64) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	return io.ReadAll(io.LimitReader(src, maxBytes+1))
}

func sniffContentType(content []byte) string {
	if len(content) == 0 {
		return "application/octet-stream"
	}
	limit := len(content)
	if limit > 512 {
		limit = 512
	}
	return normalizeContentType(http.DetectContentType(content[:limit]))
}

func normalizeContentType(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "application/octet-stream"
	}
	if i := strings.Index(value, ";"); i >= 0 {
		value = strings.TrimSpace(value[:i])
	}
	return value
}

func highRiskContentType(contentType string) bool {
	switch normalizeContentType(contentType) {
	case "text/html", "image/svg+xml", "application/javascript", "text/javascript", "application/x-sh", "application/x-msdownload":
		return true
	default:
		return false
	}
}

func clientTypeAllowed(ext string, clientType string) bool {
	if clientType == "" || clientType == "application/octet-stream" {
		return true
	}
	for _, allowed := range allowedMIMETypes(ext) {
		if clientType == allowed {
			return true
		}
	}
	return false
}

func extensionMatchesContent(ext string, detectedType string, content []byte) bool {
	if zipBacked(ext) {
		return bytes.HasPrefix(content, []byte("PK\x03\x04")) || bytes.HasPrefix(content, []byte("PK\x05\x06")) || bytes.HasPrefix(content, []byte("PK\x07\x08"))
	}
	if oleDocument(ext) {
		return bytes.HasPrefix(content, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})
	}
	for _, allowed := range allowedMIMETypes(ext) {
		if detectedType == allowed {
			return true
		}
	}
	return false
}

func allowedMIMETypes(ext string) []string {
	switch ext {
	case ".csv":
		return []string{"text/plain", "text/csv", "application/csv", "application/vnd.ms-excel"}
	case ".txt":
		return []string{"text/plain"}
	case ".jpg", ".jpeg":
		return []string{"image/jpeg"}
	case ".png":
		return []string{"image/png"}
	case ".gif":
		return []string{"image/gif"}
	case ".webp":
		return []string{"image/webp"}
	case ".pdf":
		return []string{"application/pdf"}
	case ".zip":
		return []string{"application/zip", "application/x-zip-compressed", "application/octet-stream"}
	case ".docx":
		return []string{"application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/zip", "application/octet-stream"}
	case ".xlsx":
		return []string{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "application/zip", "application/octet-stream"}
	case ".pptx":
		return []string{"application/vnd.openxmlformats-officedocument.presentationml.presentation", "application/zip", "application/octet-stream"}
	case ".doc":
		return []string{"application/msword", "application/octet-stream"}
	case ".xls":
		return []string{"application/vnd.ms-excel", "application/octet-stream"}
	case ".ppt":
		return []string{"application/vnd.ms-powerpoint", "application/octet-stream"}
	default:
		return nil
	}
}

func zipBacked(ext string) bool {
	return ext == ".zip" || ext == ".docx" || ext == ".xlsx" || ext == ".pptx"
}

func oleDocument(ext string) bool {
	return ext == ".doc" || ext == ".xls" || ext == ".ppt"
}

func validateZip(content []byte) error {
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return errors.New("ZIP 文件格式错误")
	}
	var total uint64
	for i, item := range reader.File {
		if i >= 1000 {
			return errors.New("ZIP 文件数量超限")
		}
		if zipPathUnsafe(item.Name) {
			return errors.New("ZIP 文件路径非法")
		}
		total += item.UncompressedSize64
		if total > uint64(200*1024*1024) {
			return errors.New("ZIP 解压后大小超限")
		}
	}
	return nil
}

func zipPathUnsafe(name string) bool {
	normalized := strings.ReplaceAll(name, "\\", "/")
	if strings.HasPrefix(normalized, "/") || strings.Contains(normalized, "\x00") {
		return true
	}
	cleaned := filepath.Clean(normalized)
	return cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../")
}
