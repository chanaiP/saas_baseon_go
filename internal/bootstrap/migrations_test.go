package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrationVersionFromFile(t *testing.T) {
	require.Equal(t, "000001_create_system_param", migrationVersionFromFile("/repo/migrations/000001_create_system_param.up.sql"))
	require.Empty(t, migrationVersionFromFile("/repo/migrations/000001_create_system_param.down.sql"))
}
