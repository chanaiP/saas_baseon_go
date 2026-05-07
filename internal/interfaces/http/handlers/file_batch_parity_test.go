package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileIDValidationRejectsGlobPatterns(t *testing.T) {
	require.True(t, safeFileIDPattern.MatchString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	require.False(t, safeFileIDPattern.MatchString("*"))
	require.False(t, safeFileIDPattern.MatchString("../aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
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
