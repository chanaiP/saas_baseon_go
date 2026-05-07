package handlers

import (
	"testing"
	"time"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"

	"github.com/stretchr/testify/require"
)

func TestLoginLogJSONMapsRowAndTenantName(t *testing.T) {
	tenantID := uint64(7)
	userID := uint64(9)
	message := "ok"
	ip := "127.0.0.1"
	tenantName := "七号楼主体"
	createdAt := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)

	item := loginLogToJSON(models.LoginLog{ID: 1, TenantID: &tenantID, UserID: &userID, Account: "admin", Success: true, Message: &message, IP: &ip, CreatedAt: createdAt}, &tenantName)

	require.Equal(t, uint64(1), item["id"])
	require.Equal(t, &tenantID, item["tenant_id"])
	require.Equal(t, &tenantName, item["tenant_name"])
	require.Equal(t, true, item["success"])
	require.Equal(t, "admin", item["account"])
	require.Equal(t, createdAt, item["created_at"])
}

func TestAuditLogJSONMapsRowAndTenantName(t *testing.T) {
	tenantID := uint64(8)
	userID := uint64(10)
	detail := `{"id":1}`
	ip := "127.0.0.1"
	tenantName := "八号租户"

	item := auditLogToJSON(models.AuditLog{ID: 2, TenantID: &tenantID, UserID: &userID, Module: "user", Action: "create", Summary: "创建用户", Detail: &detail, IP: &ip}, &tenantName)

	require.Equal(t, uint64(2), item["id"])
	require.Equal(t, &tenantID, item["tenant_id"])
	require.Equal(t, &tenantName, item["tenant_name"])
	require.Equal(t, "user", item["module"])
	require.Equal(t, "create", item["action"])
	require.Equal(t, "创建用户", item["summary"])
}

func TestParseDateOnlyForLogFilters(t *testing.T) {
	parsed, ok := parseDateOnly(" 2026-01-15 ")
	require.True(t, ok)
	require.Equal(t, time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), parsed)

	_, ok = parseDateOnly("2026-99-99")
	require.False(t, ok)
}
