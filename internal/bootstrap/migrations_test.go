package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateMigrationChecksumsDetectsMismatch(t *testing.T) {
	files := []migrationFile{{Version: "000001", Checksum: "local"}}
	applied := map[string]string{"000001": "applied"}

	err := validateMigrationChecksums(files, applied)

	require.Error(t, err)
	require.Contains(t, err.Error(), "checksum mismatch")
}

func TestValidateMigrationChecksumsAllowsPendingMigration(t *testing.T) {
	files := []migrationFile{{Version: "000002", Checksum: "local"}}
	applied := map[string]string{}

	require.NoError(t, validateMigrationChecksums(files, applied))
}
