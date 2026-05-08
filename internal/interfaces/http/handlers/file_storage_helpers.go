package handlers

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

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

func validateUploadFile(originalName string, file *multipart.FileHeader) (string, error) {
	ext := safeFileExt(originalName)
	if _, ok := allowedUploadExtensions[ext]; !ok {
		return "", errors.New("不支持的文件类型")
	}
	if isHighRiskUploadContentType(file.Header.Get("Content-Type")) {
		return "", errors.New("不允许上传高风险文件类型")
	}
	return ext, nil
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
