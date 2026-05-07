package handlers

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPasswordHashAndVerify(t *testing.T) {
	hashed, err := hashPassword("112233")

	require.NoError(t, err)
	require.True(t, strings.HasPrefix(hashed, "bcrypt:"))
	require.True(t, verifyPassword("112233", hashed))
	require.False(t, verifyPassword("bad-password", hashed))
}

func TestVerifyPasswordRejectsLegacyAndPlaintextPasswords(t *testing.T) {
	require.False(t, verifyPassword("112233", "dev:112233"))
	require.False(t, verifyPassword("112233", "dev-password-placeholder"))
	require.False(t, verifyPassword("112233", "112233"))
}

func TestIssueAndParseToken(t *testing.T) {
	token, err := issueToken(7, 11, "secret", time.Hour)

	require.NoError(t, err)
	claims, err := parseToken(token, "secret")
	require.NoError(t, err)
	require.Equal(t, uint64(7), claims.UserID)
	require.Equal(t, uint64(11), claims.TenantID)
	require.Greater(t, claims.Exp, time.Now().Unix())
}

func TestParseTokenRejectsTampering(t *testing.T) {
	token, err := issueToken(7, 11, "secret", time.Hour)
	require.NoError(t, err)

	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)
	tampered := parts[0] + "." + parts[1] + ".bad-signature"

	_, err = parseToken(tampered, "secret")
	require.Error(t, err)
}

func TestParseTokenRejectsExpiredToken(t *testing.T) {
	token, err := issueToken(7, 11, "secret", -time.Hour)
	require.NoError(t, err)

	_, err = parseToken(token, "secret")
	require.Error(t, err)
}

func TestBearerToken(t *testing.T) {
	require.Equal(t, "abc", bearerToken("Bearer abc"))
	require.Equal(t, "abc", bearerToken("bearer abc"))
	require.Equal(t, "raw", bearerToken("raw"))
	require.Empty(t, bearerToken(""))
}
