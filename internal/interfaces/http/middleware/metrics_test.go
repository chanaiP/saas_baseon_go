package middleware

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetricsSnapshotIncludesRequestsAndDependencies(t *testing.T) {
	collector := NewMetricsCollector()
	key := metricKey("GET", "/health", 200)
	collector.requests[key] = 2
	collector.latency[key] = 15

	snapshot := collector.Snapshot(true, false)

	require.Contains(t, snapshot, `saas_http_requests_total{method="GET",route="/health",status="200"} 2`)
	require.Contains(t, snapshot, `saas_dependency_up{name="postgres"} 1`)
	require.Contains(t, snapshot, `saas_dependency_up{name="redis"} 0`)
	require.False(t, strings.Contains(snapshot, "Authorization"))
}
