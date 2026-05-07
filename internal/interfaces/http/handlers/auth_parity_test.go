package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthSessionKeyUsesOriginalRedisNamespace(t *testing.T) {
	require.Equal(t, "auth:session:opaque-token", authSessionKey("opaque-token"))
}

func TestLooksLikeMobileAccount(t *testing.T) {
	require.True(t, looksLikeMobileAccount("13800138000"))
	require.False(t, looksLikeMobileAccount("23800138000"))
	require.False(t, looksLikeMobileAccount("1380013800"))
	require.False(t, looksLikeMobileAccount("1380013800a"))
}
