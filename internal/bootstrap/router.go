package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"saas_baseon_go/internal/application/system"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/repositories"
	"saas_baseon_go/internal/interfaces/http/handlers"
	"saas_baseon_go/internal/interfaces/http/middleware"
)

func NewRouter(cfg Config, db *gorm.DB, redisClient *redis.Client) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery(), middleware.RequestID())

	healthHandler := handlers.NewHealthHandler(db, redisClient)

	paramRepo := repositories.NewSystemParamRepository(db)
	paramService := system.NewParamService(paramRepo)
	paramHandler := handlers.NewParamHandler(paramService)

	router.GET("/health", healthHandler.Check)

	api := router.Group("/api")
	{
		api.GET("/params", paramHandler.List)
		api.POST("/params", paramHandler.Create)
		api.GET("/params/:key", paramHandler.GetByKey)
	}

	return router
}
