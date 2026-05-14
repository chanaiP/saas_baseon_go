package handlers

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
	"saas_baseon_go/internal/interfaces/http/response"
)

func (h *IdentityHandler) LoginLogs(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	skip, limit := paginationParams(c)
	var rows []models.LoginLog
	query := h.db.Model(&models.LoginLog{})
	if !h.viewerHasPlatformScope(user) {
		query = query.Where("login_log.tenant_id = ?", user.TenantID)
	} else if tenantName := logTenantNameQuery(c); tenantName != "" {
		like := "%" + tenantName + "%"
		query = query.Joins("LEFT JOIN tenant t ON t.id = login_log.tenant_id").Where("t.name LIKE ? OR t.code LIKE ?", like, like)
	}
	if raw := strings.TrimSpace(c.Query("success")); raw != "" {
		query = query.Where("login_log.success = ?", raw == "true" || raw == "1")
	}
	if account := strings.TrimSpace(c.Query("account")); account != "" {
		query = query.Where("login_log.account LIKE ?", "%"+account+"%")
	}
	if ip := strings.TrimSpace(c.Query("ip")); ip != "" {
		query = query.Where("login_log.ip LIKE ?", "%"+ip+"%")
	}
	query = applyDateRange(query, "login_log.created_at", c.Query("date_from"), c.Query("date_to"))
	var total int64
	_ = query.Count(&total).Error
	_ = query.Order("login_log.id desc").Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		items = append(items, loginLogToJSON(row, h.tenantName(row.TenantID)))
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func (h *IdentityHandler) AuditLogs(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		response.Error(c, 401, response.CodeUnauthorized, "请先登录")
		return
	}
	skip, limit := paginationParams(c)
	var rows []models.AuditLog
	query := h.db.Model(&models.AuditLog{})
	if !h.viewerHasPlatformScope(user) {
		query = query.Where("audit_log.tenant_id = ?", user.TenantID)
	} else if tenantName := logTenantNameQuery(c); tenantName != "" {
		like := "%" + tenantName + "%"
		query = query.Joins("LEFT JOIN tenant t ON t.id = audit_log.tenant_id").Where("t.name LIKE ? OR t.code LIKE ?", like, like)
	}
	if appCode := strings.TrimSpace(c.Query("app_code")); appCode != "" {
		query = applyAuditAppFilter(query, appCode)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("audit_log.module LIKE ? OR audit_log.action LIKE ? OR audit_log.summary LIKE ? OR audit_log.detail LIKE ?", like, like, like, like)
	} else if module := strings.TrimSpace(c.Query("module")); module != "" {
		query = query.Where("audit_log.module = ?", module)
	}
	if account := strings.TrimSpace(c.Query("account")); account != "" {
		like := "%" + account + "%"
		query = query.Joins("LEFT JOIN app_user au ON au.id = audit_log.user_id").Where("au.employee_no LIKE ? OR au.name LIKE ? OR au.phone LIKE ? OR au.email LIKE ?", like, like, like, like)
	}
	if ip := strings.TrimSpace(c.Query("ip")); ip != "" {
		query = query.Where("audit_log.ip LIKE ?", "%"+ip+"%")
	}
	query = applyDateRange(query, "audit_log.created_at", c.Query("date_from"), c.Query("date_to"))
	var total int64
	_ = query.Count(&total).Error
	_ = query.Order("audit_log.id desc").Offset(skip).Limit(limit).Find(&rows).Error
	items := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		appCode := auditLogAppCode(row)
		userName, userAccount, userEmployeeNo := h.userIdentity(row.UserID)
		items = append(items, auditLogToJSON(row, h.tenantName(row.TenantID), userName, userAccount, userEmployeeNo, appCode, h.appName(appCode)))
	}
	response.OK(c, paginatedWithTotal(items, total, skip, limit))
}

func applyAuditAppFilter(query *gorm.DB, appCode string) *gorm.DB {
	modules := auditModulesForAppCode(appCode)
	if len(modules) == 0 {
		return query.Where("audit_log.app_code = ? OR audit_log.module = ? OR audit_log.module = ?", appCode, appCode, strings.ReplaceAll(appCode, "-", "_"))
	}
	return query.Where("audit_log.app_code = ? OR (audit_log.app_code IS NULL AND audit_log.module IN ?)", appCode, modules)
}

func auditModulesForAppCode(appCode string) []string {
	switch appCode {
	case "app-center":
		return []string{"app", "app_center", "application"}
	case "system-monitor":
		return []string{"monitor", "system_monitor"}
	case "model-manager":
		return []string{"model", "model_manager", "model-manager"}
	case "system-management":
		return []string{
			"business_unit", "business_unit_org_map", "dict_item", "dict_type", "file",
			"menu", "organization", "permission", "plan", "position", "position_type",
			"profile", "quota", "role", "sys_param", "tenant", "tenant_branding",
			"tenant_feature_override", "tenant_quota_override", "tenant_subscription", "user",
		}
	default:
		return nil
	}
}

func auditAppCode(module string) string {
	for _, appCode := range []string{"app-center", "system-monitor", "model-manager", "system-management"} {
		for _, item := range auditModulesForAppCode(appCode) {
			if item == module {
				return appCode
			}
		}
	}
	return strings.ReplaceAll(module, "_", "-")
}

func auditLogAppCode(row models.AuditLog) string {
	if row.AppCode != nil && strings.TrimSpace(*row.AppCode) != "" {
		return strings.TrimSpace(*row.AppCode)
	}
	return auditAppCode(row.Module)
}

func logTenantNameQuery(c *gin.Context) string {
	if value := strings.TrimSpace(c.Query("tenant_name_hint")); value != "" {
		return value
	}
	return strings.TrimSpace(c.Query("tenant_name"))
}

func (h *IdentityHandler) MonitorHealthDetail(c *gin.Context) {
	redisOK := false
	if h.redis != nil {
		redisOK = h.redis.Ping(context.Background()).Err() == nil
	}
	response.OK(c, gin.H{"mysql": h.databaseOK(), "redis": redisOK})
}

func (h *IdentityHandler) MonitorServerInfo(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	response.OK(c, gin.H{"python_version": runtime.Version(), "pid": os.Getpid(), "cpu_percent": nil, "memory_mb": float64(m.Alloc) / 1024 / 1024, "note": nil})
}

func (h *IdentityHandler) MonitorScheduledJobs(c *gin.Context) {
	response.OK(c, gin.H{
		"items": []gin.H{
			{"id": "redis-session", "name": "Redis 会话与会话校验", "schedule": "随请求读写", "status": "运行中"},
			{"id": "captcha", "name": "验证码存储", "schedule": "Redis TTL 300s", "status": "运行中"},
			{"id": "login-fail", "name": "登录失败计数", "schedule": "Redis TTL 900s", "status": "运行中"},
			{"id": "ai-api-connectivity", "name": "AI API 连通性检测", "schedule": "每 10 分钟", "status": "运行中"},
		},
		"note": "以下为当前与 Redis 相关的内置行为说明；周期性报表、清理等可接入 APScheduler 后在此登记展示。",
	})
}

func (h *IdentityHandler) MonitorServicesOverview(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	redisOK := false
	if h.redis != nil {
		redisOK = h.redis.Ping(context.Background()).Err() == nil
	}
	response.OK(c, gin.H{"mysql": h.databaseOK(), "redis": redisOK, "python_version": runtime.Version(), "pid": os.Getpid(), "cpu_percent": nil, "memory_mb": float64(m.Alloc) / 1024 / 1024, "note": nil})
}

func (h *IdentityHandler) MonitorCacheStats(c *gin.Context) {
	if h.redis == nil {
		response.OK(c, gin.H{"ok": false, "used_memory_human": nil, "keys": 0, "connected_clients": 0, "message": "Redis 未配置"})
		return
	}
	ctx := context.Background()
	if err := h.redis.Ping(ctx).Err(); err != nil {
		response.OK(c, gin.H{"ok": false, "used_memory_human": nil, "keys": 0, "connected_clients": 0, "message": err.Error()})
		return
	}
	info := redisInfoMap(h.redis.Info(ctx, "memory", "clients").Val())
	keys, _ := h.redis.DBSize(ctx).Result()
	clients := 0
	if raw := info["connected_clients"]; raw != "" {
		clients, _ = strconv.Atoi(raw)
	}
	response.OK(c, gin.H{"ok": true, "used_memory_human": info["used_memory_human"], "keys": keys, "connected_clients": clients, "message": "Redis 已连接"})
}

func (h *IdentityHandler) MonitorCacheKeys(c *gin.Context) {
	if h.redis == nil {
		response.Error(c, 503, response.CodeBadRequest, "Redis 未配置")
		return
	}
	ctx := context.Background()
	cursor, _ := strconv.ParseUint(c.Query("cursor"), 10, 64)
	limit, pattern, err := monitorCacheKeyQuery(c.Query("limit"), c.Query("pattern"))
	if err != nil {
		respondBadRequest(c, err)
		return
	}
	items := make([]gin.H, 0, limit)
	nextCursor := cursor
	for int64(len(items)) < limit {
		keys, scannedCursor, err := h.redis.Scan(ctx, nextCursor, pattern, 200).Result()
		if err != nil {
			_ = c.Error(err)
			response.Error(c, 503, response.CodeBadRequest, "Redis 查询失败")
			return
		}
		nextCursor = scannedCursor
		for _, key := range keys {
			if int64(len(items)) >= limit {
				break
			}
			ttl, err := h.redis.TTL(ctx, key).Result()
			if err != nil {
				_ = c.Error(err)
				response.Error(c, 503, response.CodeBadRequest, "Redis 查询失败")
				return
			}
			items = append(items, gin.H{"key": key, "ttl": int(ttl.Seconds())})
		}
		if nextCursor == 0 {
			break
		}
	}
	response.OK(c, gin.H{"items": items, "cursor": nextCursor})
}

func (h *IdentityHandler) databaseOK() bool {
	if h.db == nil {
		return false
	}
	sqlDB, err := h.db.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}
