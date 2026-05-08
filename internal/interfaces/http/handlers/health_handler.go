package handlers

import (
	"context"
	"net/http"
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
	postgresOK, redisOK := h.dependencyStatus(c.Request.Context())
	response.OK(c, gin.H{
		"postgres": postgresOK,
		"redis":    redisOK,
	})
}

func (h *HealthHandler) DependencyStatus() (bool, bool) {
	return h.dependencyStatus(context.Background())
}

func (h *HealthHandler) Live(c *gin.Context) {
	response.OK(c, gin.H{"live": true})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	postgresOK, redisOK := h.dependencyStatus(c.Request.Context())
	status := http.StatusOK
	if !postgresOK || !redisOK {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, response.Body{
		Code:    0,
		Message: "ok",
		Data: gin.H{
			"postgres": postgresOK,
			"redis":    redisOK,
			"ready":    postgresOK && redisOK,
		},
	})
}

func (h *HealthHandler) dependencyStatus(parent context.Context) (bool, bool) {
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()

	postgresOK := false
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err == nil && sqlDB.PingContext(ctx) == nil {
			postgresOK = true
		}
	}

	redisOK := false
	if h.redis != nil && h.redis.Ping(ctx).Err() == nil {
		redisOK = true
	}

	return postgresOK, redisOK
}
