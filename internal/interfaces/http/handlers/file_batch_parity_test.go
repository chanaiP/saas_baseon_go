package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

func TestFileIDValidationRejectsGlobPatterns(t *testing.T) {
	require.True(t, safeFileIDPattern.MatchString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	require.False(t, safeFileIDPattern.MatchString("*"))
	require.False(t, safeFileIDPattern.MatchString("../aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
}

func TestUploadFileTypePolicyRejectsHighRiskTypes(t *testing.T) {
	_, pdfAllowed := allowedUploadExtensions[safeFileExt("report.PDF")]
	_, svgAllowed := allowedUploadExtensions[safeFileExt("logo.svg")]
	_, shellAllowed := allowedUploadExtensions[safeFileExt("deploy.sh")]

	require.True(t, pdfAllowed)
	require.False(t, svgAllowed)
	require.False(t, shellAllowed)
	require.True(t, isHighRiskUploadContentType("image/svg+xml"))
	require.True(t, isHighRiskUploadContentType("text/html"))
	require.False(t, isHighRiskUploadContentType("application/pdf"))
	require.Equal(t, "text/html", normalizeContentType("Text/HTML; charset=utf-8"))
}

func TestFileObjectOwnerCanAccessWithoutPermissionLookup(t *testing.T) {
	row := models.FileObject{TenantID: 9, CreatedBy: 7}

	require.True(t, fileObjectAccessAllowed(row, models.AppUser{ID: 7, TenantID: 9}, "file:download", nil))
	require.True(t, fileObjectAccessAllowed(row, models.AppUser{ID: 8, TenantID: 9}, "file:download", map[string]struct{}{"file:download": {}}))
	require.False(t, fileObjectAccessAllowed(row, models.AppUser{ID: 8, TenantID: 9}, "file:download", nil))
	require.False(t, fileObjectAccessAllowed(row, models.AppUser{ID: 7, TenantID: 10}, "file:download", map[string]struct{}{"file:download": {}}))
}

func TestFileObjectPathMustStayInsideTenantRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv("UPLOAD_DIR", root)
	inside := filepath.Join(root, "9", "2026", "01", "01", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.txt")
	outside := filepath.Join(root, "10", "2026", "01", "01", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.txt")

	require.True(t, filePathInsideTenantRoot(9, inside))
	require.False(t, filePathInsideTenantRoot(9, outside))
}

func TestDecodeCSVContentRemovesUTF8BOM(t *testing.T) {
	require.Equal(t, []byte("employee_no\n"), decodeCSVContent([]byte{0xEF, 0xBB, 0xBF, 'e', 'm', 'p', 'l', 'o', 'y', 'e', 'e', '_', 'n', 'o', '\n'}))
}

func TestTenantStorageBytesSkipsTrash(t *testing.T) {
	root := t.TempDir()
	t.Setenv("UPLOAD_DIR", root)
	require.NoError(t, os.MkdirAll(filepath.Join(root, "1", "2026", "01", "01"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "1", ".trash", "2026", "01", "01"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "1", "2026", "01", "01", "live.txt"), []byte("live"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "1", ".trash", "2026", "01", "01", "old.txt"), []byte("deleted"), 0o644))

	total, err := tenantStorageBytes(1)

	require.NoError(t, err)
	require.Equal(t, int64(4), total)
}
