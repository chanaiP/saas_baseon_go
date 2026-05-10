package handlers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestPasswordHashAndVerify(t *testing.T) {
	hashed, err := hashPassword("112233")

	require.NoError(t, err)
	require.True(t, strings.HasPrefix(hashed, "pbkdf2_sha256$"))
	require.True(t, verifyPassword("112233", hashed))
	require.False(t, verifyPassword("bad-password", hashed))
}

func TestVerifyPasswordRejectsLegacyAndPlaintextPasswords(t *testing.T) {
	require.False(t, verifyPassword("112233", "dev:112233"))
	require.False(t, verifyPassword("112233", "dev-password-placeholder"))
	require.False(t, verifyPassword("112233", "bcrypt:$2a$10$dummydummydummydummydummydummy"))
	require.False(t, verifyPassword("112233", "112233"))
}

func TestGenerateRandomPasswordMatchesOriginalPolicy(t *testing.T) {
	password := generateRandomPassword(14)

	require.Len(t, password, 14)
	require.Regexp(t, `[A-Za-z]`, password)
	require.Regexp(t, `\d`, password)
}

func TestGenerateRandomPasswordEnforcesMinimumLength(t *testing.T) {
	password := generateRandomPassword(4)

	require.GreaterOrEqual(t, len(password), 8)
	require.Regexp(t, `[A-Za-z]`, password)
	require.Regexp(t, `\d`, password)
}

func TestIssueAndParseToken(t *testing.T) {
	token, err := issueToken(7, 11, 3, "secret", time.Hour)

	require.NoError(t, err)
	claims, err := parseToken(token, "secret")
	require.NoError(t, err)
	require.Equal(t, uint64(7), claims.UserID)
	require.Equal(t, uint64(11), claims.TenantID)
	require.Equal(t, 3, claims.SessionVersion)
	require.Greater(t, claims.IssuedAt, int64(0))
	require.Greater(t, claims.Exp, time.Now().Unix())
}

func TestParseTokenRejectsTampering(t *testing.T) {
	token, err := issueToken(7, 11, 1, "secret", time.Hour)
	require.NoError(t, err)

	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)
	tampered := parts[0] + "." + parts[1] + ".bad-signature"

	_, err = parseToken(tampered, "secret")
	require.Error(t, err)
}

func TestParseTokenRejectsExpiredToken(t *testing.T) {
	token, err := issueToken(7, 11, 1, "secret", -time.Hour)
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

func TestResolveAuthIdentityUsesRedisSession(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	handler := &IdentityHandler{redis: client, authSecret: "secret", jwtFallback: false}
	raw := `{"user_id":7,"tenant_id":11,"session_version":3,"issued_at":123}`
	require.NoError(t, client.Set(context.Background(), authSessionKey("opaque-token"), raw, time.Hour).Err())

	identity, ok := handler.resolveAuthIdentity("opaque-token")

	require.True(t, ok)
	require.Equal(t, uint64(7), identity.UserID)
	require.Equal(t, uint64(11), identity.TenantID)
	require.Equal(t, 3, identity.SessionVersion)
	require.Equal(t, int64(123), identity.IssuedAt)
}

func TestResolveAuthIdentityRejectsMissingRedisSessionEvenWithValidJWT(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	token, err := issueToken(7, 11, 3, "secret", time.Hour)
	require.NoError(t, err)
	handler := &IdentityHandler{redis: client, authSecret: "secret", jwtFallback: true}

	_, ok := handler.resolveAuthIdentity(token)

	require.False(t, ok)
}

func TestResolveAuthIdentityRejectsJWTWhenFallbackDisabledAndRedisUnavailable(t *testing.T) {
	token, err := issueToken(7, 11, 3, "secret", time.Hour)
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  10 * time.Millisecond,
		ReadTimeout:  10 * time.Millisecond,
		WriteTimeout: 10 * time.Millisecond,
	})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	handler := &IdentityHandler{redis: client, authSecret: "secret", jwtFallback: false}

	_, ok := handler.resolveAuthIdentity(token)

	require.False(t, ok)
}

func TestResolveAuthIdentityAllowsJWTFallbackWhenRedisUnavailableAndEnabled(t *testing.T) {
	token, err := issueToken(7, 11, 3, "secret", time.Hour)
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  10 * time.Millisecond,
		ReadTimeout:  10 * time.Millisecond,
		WriteTimeout: 10 * time.Millisecond,
	})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	handler := &IdentityHandler{redis: client, authSecret: "secret", jwtFallback: true}

	identity, ok := handler.resolveAuthIdentity(token)

	require.True(t, ok)
	require.Equal(t, uint64(7), identity.UserID)
	require.Equal(t, uint64(11), identity.TenantID)
	require.Equal(t, 3, identity.SessionVersion)
}

func TestResolveAuthIdentityRejectsJWTWhenNoRedisAndFallbackDisabled(t *testing.T) {
	token, err := issueToken(7, 11, 3, "secret", time.Hour)
	require.NoError(t, err)
	handler := &IdentityHandler{authSecret: "secret", jwtFallback: false}

	_, ok := handler.resolveAuthIdentity(token)

	require.False(t, ok)
}
