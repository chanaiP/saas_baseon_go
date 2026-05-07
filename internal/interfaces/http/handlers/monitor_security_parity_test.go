package handlers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedisInfoMapExtractsCacheMetrics(t *testing.T) {
	info := redisInfoMap("# Memory\r\nused_memory_human:1M\r\n# Clients\r\nconnected_clients:2\r\n")

	require.Equal(t, "1M", info["used_memory_human"])
	require.Equal(t, "2", info["connected_clients"])
}

func TestMonitorCacheKeyQueryNormalizesLimitAndPattern(t *testing.T) {
	limit, pattern, err := monitorCacheKeyQuery("2", " auth:* ")

	require.NoError(t, err)
	require.Equal(t, int64(2), limit)
	require.Equal(t, "auth:*", pattern)

	limit, pattern, err = monitorCacheKeyQuery("999", "")
	require.NoError(t, err)
	require.Equal(t, int64(200), limit)
	require.Equal(t, "*", pattern)
}

func TestMonitorCacheKeyQueryRejectsInvalidPattern(t *testing.T) {
	_, _, err := monitorCacheKeyQuery("50", "bad\nkey")
	require.EqualError(t, err, "pattern 非法")

	_, _, err = monitorCacheKeyQuery("50", strings.Repeat("a", 129))
	require.EqualError(t, err, "pattern 过长")
}

func TestLoginRateLimitKeysMatchOriginalPolicy(t *testing.T) {
	require.Equal(t, "login_ip:10.0.0.1", loginIPKey("10.0.0.1"))
	require.Equal(t, "auth:session:token", authSessionKey("token"))
}
