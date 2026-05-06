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
	identityHandler := handlers.NewMockIdentityHandler(db)

	paramRepo := repositories.NewSystemParamRepository(db)
	paramService := system.NewParamService(paramRepo)
	paramHandler := handlers.NewParamHandler(paramService)

	router.GET("/health", healthHandler.Check)
	router.GET("/openapi.json", func(c *gin.Context) {
		c.JSON(200, gin.H{"openapi": "3.0.0", "info": gin.H{"title": "SaaS Baseon Go API", "version": "0.1.0"}})
	})

	api := router.Group("/api")
	{
		api.GET("/auth/captcha", identityHandler.Captcha)
		api.GET("/auth/phone-login-tenants", identityHandler.PhoneLoginTenants)
		api.POST("/auth/login", identityHandler.Login)
		api.POST("/auth/logout", identityHandler.Logout)
		api.GET("/auth/switchable-tenants", identityHandler.SwitchableTenants)
		api.POST("/auth/switch-tenant", identityHandler.SwitchTenant)
		api.GET("/users/me", identityHandler.Profile)
		api.PUT("/users/me", identityHandler.UpdateProfile)
		api.PUT("/users/me/password", identityHandler.UpdatePassword)
		api.GET("/users/me/preferences", identityHandler.Preferences)
		api.PUT("/users/me/preferences", identityHandler.SavePreferences)
		api.GET("/tenant/branding", identityHandler.TenantBranding)
		api.PUT("/tenant/branding", identityHandler.SaveTenantBranding)
		api.GET("/public/tenant-footer", identityHandler.PublicTenantFooter)
		api.GET("/roles/permission-menu-bundles", identityHandler.MenuBundles)
		api.GET("/permission-menu-bundles", identityHandler.MenuBundles)
		api.GET("/permissions/menu-bundles", identityHandler.MenuBundles)
		api.GET("/permissions/menu-overrides", identityHandler.MenuOverrides)
		api.PUT("/permissions/menu-overrides", identityHandler.SaveMenuOverrides)
		api.GET("/params", paramHandler.List)
		api.POST("/params", paramHandler.Create)
		api.GET("/params/:key", paramHandler.GetByKey)
	}

	return router
}
