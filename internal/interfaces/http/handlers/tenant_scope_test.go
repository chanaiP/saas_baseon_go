package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEffectiveTenantID(t *testing.T) {
	t.Run("platform admin can choose query tenant", func(t *testing.T) {
		require.Equal(t, uint64(8), effectiveTenantID("8", 1, true))
	})

	t.Run("platform admin falls back to default tenant", func(t *testing.T) {
		require.Equal(t, uint64(1), effectiveTenantID("", 9, true))
	})

	t.Run("tenant user cannot switch by query tenant", func(t *testing.T) {
		require.Equal(t, uint64(3), effectiveTenantID("8", 3, false))
	})

	t.Run("missing user context keeps legacy query fallback", func(t *testing.T) {
		require.Equal(t, uint64(5), effectiveTenantID("5", 0, false))
		require.Equal(t, uint64(1), effectiveTenantID("", 0, false))
	})
}

func TestNullableTrimmedMatchesBrandingPayloadPolicy(t *testing.T) {
	raw := "  data:image/jpeg;base64,abc  "
	trimmed := nullableTrimmed(&raw)
	require.NotNil(t, trimmed)
	require.Equal(t, "data:image/jpeg;base64,abc", *trimmed)

	empty := "   "
	require.Nil(t, nullableTrimmed(&empty))
	require.Nil(t, nullableTrimmed(nil))
}
