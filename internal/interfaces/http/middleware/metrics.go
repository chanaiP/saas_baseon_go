package middleware

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type MetricsCollector struct {
	mu       sync.Mutex
	requests map[string]int64
	latency  map[string]int64
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{requests: map[string]int64{}, latency: map[string]int64{}}
}

func (m *MetricsCollector) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		key := metricKey(c.Request.Method, route, c.Writer.Status())
		m.mu.Lock()
		m.requests[key]++
		m.latency[key] += time.Since(start).Milliseconds()
		m.mu.Unlock()
	}
}

func (m *MetricsCollector) Handler(dependencyStatus func() (bool, bool)) gin.HandlerFunc {
	return func(c *gin.Context) {
		postgresOK, redisOK := dependencyStatus()
		c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		c.String(http.StatusOK, m.Snapshot(postgresOK, redisOK))
	}
}

func (m *MetricsCollector) Snapshot(postgresOK bool, redisOK bool) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var b strings.Builder
	b.WriteString("# HELP saas_http_requests_total Total HTTP requests.\n")
	b.WriteString("# TYPE saas_http_requests_total counter\n")
	keys := make([]string, 0, len(m.requests))
	for key := range m.requests {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		method, route, status := splitMetricKey(key)
		b.WriteString(fmt.Sprintf("saas_http_requests_total{method=%q,route=%q,status=%q} %d\n", method, route, status, m.requests[key]))
	}
	b.WriteString("# HELP saas_http_request_duration_ms_sum Total HTTP request duration in milliseconds.\n")
	b.WriteString("# TYPE saas_http_request_duration_ms_sum counter\n")
	for _, key := range keys {
		method, route, status := splitMetricKey(key)
		b.WriteString(fmt.Sprintf("saas_http_request_duration_ms_sum{method=%q,route=%q,status=%q} %d\n", method, route, status, m.latency[key]))
	}
	b.WriteString("# HELP saas_dependency_up Dependency health status.\n")
	b.WriteString("# TYPE saas_dependency_up gauge\n")
	b.WriteString(fmt.Sprintf("saas_dependency_up{name=%q} %d\n", "postgres", boolMetric(postgresOK)))
	b.WriteString(fmt.Sprintf("saas_dependency_up{name=%q} %d\n", "redis", boolMetric(redisOK)))
	return b.String()
}

func metricKey(method, route string, status int) string {
	return fmt.Sprintf("%s\x00%s\x00%d", method, route, status)
}

func splitMetricKey(key string) (string, string, string) {
	parts := strings.Split(key, "\x00")
	if len(parts) != 3 {
		return "unknown", "unknown", "0"
	}
	return parts[0], parts[1], parts[2]
}

func boolMetric(value bool) int {
	if value {
		return 1
	}
	return 0
}
