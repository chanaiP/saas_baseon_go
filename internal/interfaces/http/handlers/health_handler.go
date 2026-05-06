package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"saas_baseon_go/internal/interfaces/http/response"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redisClient}
}

func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	postgresOK := false
	sqlDB, err := h.db.DB()
	if err == nil && sqlDB.PingContext(ctx) == nil {
		postgresOK = true
	}

	redisOK := false
	if h.redis != nil && h.redis.Ping(ctx).Err() == nil {
		redisOK = true
	}

	response.OK(c, gin.H{
		"postgres": postgresOK,
		"redis":    redisOK,
	})
}
