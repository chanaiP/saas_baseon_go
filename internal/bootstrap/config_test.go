package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvBool(t *testing.T) {
	t.Setenv("BOOL_TRUE", "true")
	t.Setenv("BOOL_FALSE", "0")
	t.Setenv("BOOL_BAD", "maybe")

	require.True(t, envBool("BOOL_TRUE", false))
	require.False(t, envBool("BOOL_FALSE", true))
	require.True(t, envBool("BOOL_MISSING", true))
	require.False(t, envBool("BOOL_BAD", false))
}

func TestLoadConfigReadsAutoMigrate(t *testing.T) {
	t.Setenv("DB_AUTO_MIGRATE", "false")

	cfg := LoadConfig()

	require.False(t, cfg.AutoMigrate)
}
