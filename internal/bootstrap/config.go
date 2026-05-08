package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const defaultAuthSecret = "saas-baseon-go-development-secret"

type Config struct {
	AppEnv        string
	HTTPAddr      string
	DatabaseDSN   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	AuthSecret    string
	TokenTTLHours int
	AutoMigrate   bool
	CORSOrigins   string
	JWTFallback   bool
}

func LoadConfig() Config {
	_ = godotenv.Load()

	return Config{
		AppEnv:        env("APP_ENV", "development"),
		HTTPAddr:      env("HTTP_ADDR", ":8081"),
		DatabaseDSN:   env("DATABASE_DSN", "host=127.0.0.1 user=saas password=saas dbname=saas_baseon port=5432 sslmode=disable TimeZone=Asia/Shanghai"),
		RedisAddr:     env("REDIS_ADDR", "127.0.0.1:6380"),
		RedisPassword: env("REDIS_PASSWORD", ""),
		RedisDB:       envInt("REDIS_DB", 0),
		AuthSecret:    env("AUTH_SECRET", defaultAuthSecret),
		TokenTTLHours: envInt("TOKEN_TTL_HOURS", 24),
		AutoMigrate:   envBool("DB_AUTO_MIGRATE", true),
		CORSOrigins:   env("CORS_ORIGINS", "*"),
		JWTFallback:   envBool("AUTH_JWT_FALLBACK", true),
	}
}

func (c Config) ValidateForRuntime() error {
	if !strings.EqualFold(c.AppEnv, "production") {
		return nil
	}

	var problems []string
	if c.AuthSecret == "" || c.AuthSecret == defaultAuthSecret {
		problems = append(problems, "AUTH_SECRET must be set to a non-default value")
	}
	if c.AutoMigrate {
		problems = append(problems, "DB_AUTO_MIGRATE must be false in production")
	}
	if c.JWTFallback {
		problems = append(problems, "AUTH_JWT_FALLBACK must be false in production")
	}
	if usesDefaultDatabaseCredential(c.DatabaseDSN) {
		problems = append(problems, "DATABASE_DSN must not use the default development database credentials")
	}
	if redisAddressUnsafe(c.RedisAddr) {
		problems = append(problems, "REDIS_ADDR must be explicitly configured for production")
	}
	if corsOriginsUnsafe(c.CORSOrigins) {
		problems = append(problems, "CORS_ORIGINS must be explicitly configured and cannot allow all origins in production")
	}
	if len(problems) > 0 {
		return fmt.Errorf("unsafe production configuration: %w: %s", ErrUnsafeProductionConfig, strings.Join(problems, "; "))
	}
	return nil
}

func (c Config) SafeSummary() map[string]interface{} {
	return map[string]interface{}{
		"app_env":              c.AppEnv,
		"http_addr":            c.HTTPAddr,
		"database_configured":  strings.TrimSpace(c.DatabaseDSN) != "",
		"redis_configured":     strings.TrimSpace(c.RedisAddr) != "",
		"redis_db":             c.RedisDB,
		"auth_secret_set":      c.AuthSecret != "" && c.AuthSecret != defaultAuthSecret,
		"token_ttl_hours":      c.TokenTTLHours,
		"auto_migrate":         c.AutoMigrate,
		"cors_explicit":        !corsOriginsUnsafe(c.CORSOrigins),
		"jwt_fallback_enabled": c.JWTFallback,
	}
}

var ErrUnsafeProductionConfig = errors.New("unsafe production configuration")

func usesDefaultDatabaseCredential(dsn string) bool {
	normalized := strings.ToLower(strings.TrimSpace(dsn))
	if normalized == "" {
		return true
	}
	return strings.Contains(normalized, "user=saas") || strings.Contains(normalized, "password=saas")
}

func corsOriginsUnsafe(origins string) bool {
	parts := strings.Split(origins, ",")
	hasOrigin := false
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		hasOrigin = true
		if origin == "*" {
			return true
		}
	}
	return !hasOrigin
}

func redisAddressUnsafe(addr string) bool {
	normalized := strings.ToLower(strings.TrimSpace(addr))
	return normalized == "" || normalized == "127.0.0.1:6380"
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "no", "NO", "off", "OFF":
		return false
	default:
		return fallback
	}
}
