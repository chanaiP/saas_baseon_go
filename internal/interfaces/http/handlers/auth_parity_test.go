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

func TestValidateNewPasswordMatchesOriginalPolicy(t *testing.T) {
	require.Empty(t, validateNewPassword("old12345", "new12345", "new12345"))
	require.Equal(t, "请求参数错误", validateNewPassword("", "new12345", "new12345"))
	require.Equal(t, "新密码须同时包含英文字母与数字", validateNewPassword("old12345", "abcdefgh", "abcdefgh"))
	require.Equal(t, "两次输入的新密码不一致", validateNewPassword("old12345", "new12345", "new12346"))
	require.Equal(t, "新密码不能与当前密码相同", validateNewPassword("same1234", "same1234", "same1234"))
}
