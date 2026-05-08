package handlers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"mime/multipart"
	"net/textproto"
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

func TestValidateUploadFileRejectsForgedContentType(t *testing.T) {
	file := multipartUploadFileHeader(t, "note.txt", "image/jpeg", []byte("plain text content"))

	_, err := validateUploadFile("note.txt", file)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Content-Type")
}

func TestValidateUploadFileRejectsJPGExtensionWithMismatchedContent(t *testing.T) {
	file := multipartUploadFileHeader(t, "avatar.jpg", "image/jpeg", []byte("not really a jpeg"))

	_, err := validateUploadFile("avatar.jpg", file)

	require.Error(t, err)
	require.Contains(t, err.Error(), "扩展名")
}

func TestValidateUploadFileRejectsZipPathTraversal(t *testing.T) {
	var data bytes.Buffer
	zw := zip.NewWriter(&data)
	w, err := zw.Create("../evil.txt")
	require.NoError(t, err)
	_, err = w.Write([]byte("bad"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	file := multipartUploadFileHeader(t, "archive.zip", "application/zip", data.Bytes())

	_, err = validateUploadFile("archive.zip", file)

	require.Error(t, err)
	require.Contains(t, err.Error(), "路径")
}

func TestFileObjectOwnerCanAccessWithoutPermissionLookup(t *testing.T) {
	row := models.FileObject{TenantID: 9, CreatedBy: 7}

	require.True(t, fileObjectAccessAllowed(row, models.AppUser{ID: 7, TenantID: 9}, "file:download", nil))
	require.True(t, fileObjectAccessAllowed(row, models.AppUser{ID: 8, TenantID: 9}, "file:download", map[string]struct{}{"file:download": {}}))
	require.False(t, fileObjectAccessAllowed(row, models.AppUser{ID: 8, TenantID: 9}, "file:download", nil))
	require.False(t, fileObjectAccessAllowed(row, models.AppUser{ID: 7, TenantID: 10}, "file:download", map[string]struct{}{"file:download": {}}))
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
