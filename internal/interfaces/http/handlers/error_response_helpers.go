package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"

	"saas_baseon_go/internal/interfaces/http/response"
)

func businessErrorMessage(err error) string {
	if err == nil {
		return "请求处理失败"
	}
	message := response.SafeMessage(400, err.Error())
	if strings.TrimSpace(message) == "" {
		return "请求处理失败"
	}
	return message
}

func respondBadRequest(c *gin.Context, err error) {
	if err != nil {
		_ = c.Error(err)
	}
	response.Error(c, 400, response.CodeBadRequest, businessErrorMessage(err))
}

func respondForbidden(c *gin.Context, err error) {
	if err != nil {
		_ = c.Error(err)
	}
	response.Error(c, 403, response.CodeForbidden, response.SafeMessage(403, businessErrorMessage(err)))
}

func respondRateLimited(c *gin.Context, err error) {
	if err != nil {
		_ = c.Error(err)
	}
	response.Error(c, 429, response.CodeBadRequest, businessErrorMessage(err))
}
