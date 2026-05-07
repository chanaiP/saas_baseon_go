package bootstrap

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	HTTPAddr      string
	DatabaseDSN   string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	AuthSecret    string
	TokenTTLHours int
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
		AuthSecret:    env("AUTH_SECRET", "saas-baseon-go-development-secret"),
		TokenTTLHours: envInt("TOKEN_TTL_HOURS", 24),
	}
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
