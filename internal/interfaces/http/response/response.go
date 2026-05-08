package response

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CodeOK           = 0
	CodeBadRequest   = 40000
	CodeUnauthorized = 40100
	CodeForbidden    = 40300
	CodeConflict     = 40900
	CodeNotFound     = 40400
	CodeInternal     = 50000
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Body{Code: CodeOK, Message: "ok", Data: data})
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	if code == CodeBadRequest && IsUniqueConstraintMessage(message) {
		httpStatus = 409
		code = CodeConflict
		message = "数据已存在，请检查唯一字段"
	}
	c.JSON(httpStatus, Body{Code: code, Message: SafeMessage(httpStatus, message)})
}

func IsUniqueConstraintMessage(message string) bool {
	lower := strings.ToLower(strings.TrimSpace(message))
	return strings.Contains(lower, "duplicate key") ||
		strings.Contains(lower, "unique constraint") ||
		strings.Contains(lower, "sqlstate 23505")
}

func SafeMessage(httpStatus int, message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return defaultMessage(httpStatus)
	}
	lower := strings.ToLower(trimmed)
	sensitivePatterns := []string{
		"authorization", "bearer ", "database_dsn", "password=", "token", "secret",
		"sqlstate", "duplicate key", "violates", "pq:", "pgconn", "gorm", "driver:",
		"connection refused", "no such host", "syntax error at or near",
	}
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lower, pattern) {
			return defaultMessage(httpStatus)
		}
	}
	if httpStatus >= 500 {
		return defaultMessage(httpStatus)
	}
	return trimmed
}

func defaultMessage(httpStatus int) string {
	switch {
	case httpStatus == 401:
		return "请先登录"
	case httpStatus == 403:
		return "无操作权限"
	case httpStatus == 404:
		return "资源不存在"
	case httpStatus >= 500:
		return "服务暂时不可用"
	default:
		return "请求处理失败"
	}
}
