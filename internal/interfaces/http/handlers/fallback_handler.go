package handlers

import "github.com/gin-gonic/gin"

type FallbackHandler struct{}

func NewFallbackHandler() *FallbackHandler {
	return &FallbackHandler{}
}

func (h *FallbackHandler) NoRoute(c *gin.Context) {
	c.JSON(404, gin.H{"code": 40400, "message": "接口不存在"})
}
