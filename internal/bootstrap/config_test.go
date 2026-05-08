package bootstrap

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvBool(t *testing.T) {
	t.Setenv("BOOL_TRUE", "true")
	t.Setenv("BOOL_FALSE", "0")
	t.Setenv("BOOL_BAD", "maybe")

	require.True(t, envBool("BOOL_TRUE", false))
	require.False(t, envBool("BOOL_FALSE", true))
	require.True(t, envBool("BOOL_MISSING", true))
	require.False(t, envBool("BOOL_BAD", false))
}

func TestLoadConfigReadsAutoMigrate(t *testing.T) {
	t.Setenv("DB_AUTO_MIGRATE", "false")

	cfg := LoadConfig()

	require.False(t, cfg.AutoMigrate)
}

func TestValidateForRuntimeAllowsDevelopmentDefaults(t *testing.T) {
	cfg := Config{AppEnv: "development", AutoMigrate: true, AuthSecret: defaultAuthSecret, CORSOrigins: "*", JWTFallback: true}

	require.NoError(t, cfg.ValidateForRuntime())
}

func TestValidateForRuntimeRejectsUnsafeProductionConfig(t *testing.T) {
	cfg := Config{
		AppEnv:      "production",
		DatabaseDSN: "host=postgres user=saas password=saas dbname=saas_baseon",
		RedisAddr:   "127.0.0.1:6380",
		AuthSecret:  defaultAuthSecret,
		AutoMigrate: true,
		CORSOrigins: "*",
		JWTFallback: true,
		UploadDir:   "",
	}

	err := cfg.ValidateForRuntime()

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrUnsafeProductionConfig))
	require.Contains(t, err.Error(), "AUTH_SECRET")
	require.Contains(t, err.Error(), "DB_AUTO_MIGRATE")
	require.Contains(t, err.Error(), "AUTH_JWT_FALLBACK")
	require.Contains(t, err.Error(), "DATABASE_DSN")
	require.Contains(t, err.Error(), "REDIS_ADDR")
	require.Contains(t, err.Error(), "CORS_ORIGINS")
	require.Contains(t, err.Error(), "UPLOAD_DIR")
}

func TestValidateForRuntimeRejectsProductionJWTFallback(t *testing.T) {
	cfg := Config{
		AppEnv:      "production",
		DatabaseDSN: "host=postgres user=app_prod password=strong-secret dbname=saas_baseon",
		RedisAddr:   "redis:6379",
		AuthSecret:  "prod-secret-with-enough-entropy",
		AutoMigrate: false,
		CORSOrigins: "https://admin.example.com",
		JWTFallback: true,
		UploadDir:   "/srv/saas/uploads",
	}

	err := cfg.ValidateForRuntime()

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrUnsafeProductionConfig))
	require.Contains(t, err.Error(), "AUTH_JWT_FALLBACK")
}

func TestValidateForRuntimeRejectsProductionMissingUploadDir(t *testing.T) {
	cfg := Config{
		AppEnv:      "production",
		DatabaseDSN: "host=postgres user=app_prod password=strong-secret dbname=saas_baseon",
		RedisAddr:   "redis:6379",
		AuthSecret:  "prod-secret-with-enough-entropy",
		AutoMigrate: false,
		CORSOrigins: "https://admin.example.com",
		JWTFallback: false,
	}

	err := cfg.ValidateForRuntime()

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrUnsafeProductionConfig))
	require.Contains(t, err.Error(), "UPLOAD_DIR")
}

func TestValidateForRuntimeAllowsHardenedProductionConfig(t *testing.T) {
	cfg := Config{
		AppEnv:      "production",
		DatabaseDSN: "host=postgres user=app_prod password=strong-secret dbname=saas_baseon",
		RedisAddr:   "redis:6379",
		AuthSecret:  "prod-secret-with-enough-entropy",
		AutoMigrate: false,
		CORSOrigins: "https://admin.example.com",
		JWTFallback: false,
		UploadDir:   "/srv/saas/uploads",
	}

	require.NoError(t, cfg.ValidateForRuntime())
}

func TestSafeSummaryDoesNotExposeSecrets(t *testing.T) {
	cfg := Config{
		AppEnv:        "production",
		HTTPAddr:      ":8081",
		DatabaseDSN:   "host=postgres user=app password=secret",
		RedisAddr:     "redis:6379",
		RedisPassword: "redis-secret",
		AuthSecret:    "auth-secret",
		CORSOrigins:   "https://admin.example.com",
		JWTFallback:   false,
	}

	summary := cfg.SafeSummary()

	require.True(t, summary["database_configured"].(bool))
	require.True(t, summary["redis_configured"].(bool))
	require.True(t, summary["auth_secret_set"].(bool))
	require.NotContains(t, summary, "database_dsn")
	require.NotContains(t, summary, "redis_password")
	require.NotContains(t, summary, "auth_secret")
}
