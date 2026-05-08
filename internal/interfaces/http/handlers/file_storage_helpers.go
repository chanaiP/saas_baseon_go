package handlers

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

const (
	maxZipFileCount               = 1000
	maxZipUncompressedBytes int64 = 200 * 1024 * 1024
)

type uploadValidationResult struct {
	Ext      string
	MimeType string
}

type UploadVirusScanner interface {
	ScanUpload(originalName string, content []byte) error
}

type noopUploadVirusScanner struct{}

func (noopUploadVirusScanner) ScanUpload(string, []byte) error { return nil }

var uploadVirusScanner UploadVirusScanner = noopUploadVirusScanner{}

func uploadRoot() string {
	root := os.Getenv("UPLOAD_DIR")
	if root == "" {
		root = "uploads"
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return root
	}
	return abs
}

func uploadDir(tenantID uint64, now time.Time) string {
	return filepath.Join(uploadRoot(), strconv.FormatUint(tenantID, 10), now.Format("2006/01/02"))
}

func safeOriginalName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "" || base == "." || base == string(filepath.Separator) {
		base = "unknown"
	}
	if len(base) > 255 {
		base = base[:255]
	}
	return base
}

func safeFileExt(name string) string {
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

func validateUploadFile(originalName string, file *multipart.FileHeader) (uploadValidationResult, error) {
	ext := safeFileExt(originalName)
	if _, ok := allowedUploadExtensions[ext]; !ok {
		return uploadValidationResult{}, errors.New("不支持的文件类型")
	}
	clientType := normalizeContentType(file.Header.Get("Content-Type"))
	if isHighRiskUploadContentType(clientType) {
		return uploadValidationResult{}, errors.New("不允许上传高风险文件类型")
	}
	content, err := readUploadSample(file, 50*1024*1024)
	if err != nil {
		return uploadValidationResult{}, errors.New("读取文件失败")
	}
	detectedType := sniffUploadContentType(content)
	if isHighRiskUploadContentType(detectedType) {
		return uploadValidationResult{}, errors.New("不允许上传高风险文件类型")
	}
	if !uploadExtensionMatchesContent(ext, detectedType, content) {
		return uploadValidationResult{}, errors.New("文件扩展名与实际内容不一致")
	}
	if !uploadClientTypeAllowed(ext, clientType) {
		return uploadValidationResult{}, errors.New("文件 Content-Type 与扩展名不一致")
	}
	if uploadIsZipBacked(ext) {
		if err := validateZipUpload(content); err != nil {
			return uploadValidationResult{}, err
		}
	}
	if err := uploadVirusScanner.ScanUpload(originalName, content); err != nil {
		return uploadValidationResult{}, errors.New("文件安全扫描未通过")
	}
	return uploadValidationResult{Ext: ext, MimeType: detectedType}, nil
}

func saveUploadedFile(file *multipart.FileHeader, dst string) error {
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

func readUploadSample(file *multipart.FileHeader, maxBytes int64) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	return io.ReadAll(io.LimitReader(src, maxBytes+1))
}

func sniffUploadContentType(content []byte) string {
	if len(content) == 0 {
		return "application/octet-stream"
	}
	limit := len(content)
	if limit > 512 {
		limit = 512
	}
	return normalizeContentType(http.DetectContentType(content[:limit]))
}

func uploadClientTypeAllowed(ext string, clientType string) bool {
	if clientType == "" || clientType == "application/octet-stream" {
		return true
	}
	for _, allowed := range allowedUploadMIMETypes(ext) {
		if clientType == allowed {
			return true
		}
	}
	return false
}

func uploadExtensionMatchesContent(ext string, detectedType string, content []byte) bool {
	if uploadIsZipBacked(ext) {
		return bytes.HasPrefix(content, []byte("PK\x03\x04")) || bytes.HasPrefix(content, []byte("PK\x05\x06")) || bytes.HasPrefix(content, []byte("PK\x07\x08"))
	}
	if uploadIsOLEDocument(ext) {
		return bytes.HasPrefix(content, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})
	}
	for _, allowed := range allowedUploadMIMETypes(ext) {
		if detectedType == allowed {
			return true
		}
	}
	return false
}

func allowedUploadMIMETypes(ext string) []string {
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

func uploadIsZipBacked(ext string) bool {
	return ext == ".zip" || ext == ".docx" || ext == ".xlsx" || ext == ".pptx"
}

func uploadIsOLEDocument(ext string) bool {
	return ext == ".doc" || ext == ".xls" || ext == ".ppt"
}

func validateZipUpload(content []byte) error {
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return errors.New("ZIP 文件格式错误")
	}
	var total uint64
	for i, item := range reader.File {
		if i >= maxZipFileCount {
			return errors.New("ZIP 文件数量超限")
		}
		if zipPathUnsafe(item.Name) {
			return errors.New("ZIP 文件路径非法")
		}
		total += item.UncompressedSize64
		if total > uint64(maxZipUncompressedBytes) {
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

func isHighRiskUploadContentType(contentType string) bool {
	switch normalizeContentType(contentType) {
	case "text/html", "image/svg+xml", "application/javascript", "text/javascript", "application/x-sh", "application/x-msdownload":
		return true
	default:
		return false
	}
}

func (h *IdentityHandler) findTenantFileObject(user models.AppUser, fileID string, requiredPermission string) (models.FileObject, error) {
	if !safeFileIDPattern.MatchString(fileID) {
		return models.FileObject{}, errInvalidFileID
	}
	var row models.FileObject
	if err := h.db.Where("tenant_id = ? AND file_id = ? AND status = ? AND deleted_at IS NULL", user.TenantID, fileID, 1).First(&row).Error; err != nil {
		return models.FileObject{}, err
	}
	if !filePathInsideTenantRoot(user.TenantID, row.StoragePath) {
		return models.FileObject{}, errors.New("文件路径非法")
	}
	codes := map[string]struct{}{}
	for _, code := range h.permissionCodesForUser(user, true) {
		codes[code] = struct{}{}
	}
	if !fileObjectAccessAllowed(row, user, requiredPermission, codes) {
		return models.FileObject{}, gorm.ErrRecordNotFound
	}
	return row, nil
}

func filePathInsideTenantRoot(tenantID uint64, path string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	root := filepath.Join(uploadRoot(), strconv.FormatUint(tenantID, 10))
	rel, err := filepath.Rel(root, abs)
	return err == nil && rel != "." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."
}

func fileObjectAccessAllowed(row models.FileObject, user models.AppUser, requiredPermission string, codes map[string]struct{}) bool {
	if row.TenantID != user.TenantID {
		return false
	}
	if user.IsPlatformAdmin || row.CreatedBy == user.ID {
		return true
	}
	_, ok := codes[requiredPermission]
	return ok
}

func (h *IdentityHandler) findTenantFile(tenantID uint64, fileID string) (string, error) {
	if !safeFileIDPattern.MatchString(fileID) {
		return "", errInvalidFileID
	}
	root := filepath.Join(uploadRoot(), strconv.FormatUint(tenantID, 10))
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if d.IsDir() && d.Name() == ".trash" {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), fileID+".") {
			found = path
			return io.EOF
		}
		return nil
	})
	if found != "" {
		return found, nil
	}
	if err != nil && err != io.EOF {
		return "", err
	}
	return "", fmt.Errorf("not found")
}

func tenantStorageBytes(tenantID uint64) (int64, error) {
	root := filepath.Join(uploadRoot(), strconv.FormatUint(tenantID, 10))
	var total int64
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			if d != nil && d.IsDir() && d.Name() == ".trash" {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	return total, err
}
