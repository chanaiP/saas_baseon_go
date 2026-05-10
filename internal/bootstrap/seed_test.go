package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSeedInitialAdminPasswordAllowsDevelopmentDefault(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "")

	password, err := seedInitialAdminPassword()

	require.NoError(t, err)
	require.Equal(t, "112233", password)
}

func TestSeedInitialAdminPasswordRejectsProductionDefault(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "112233")

	_, err := seedInitialAdminPassword()

	require.Error(t, err)
	require.Contains(t, err.Error(), "BOOTSTRAP_ADMIN_PASSWORD")
}

func TestSeedInitialAdminPasswordAllowsExplicitProductionPassword(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "StrongBootstrap123")

	password, err := seedInitialAdminPassword()

	require.NoError(t, err)
	require.Equal(t, "StrongBootstrap123", password)
}
