package config

import (
	"os"
	"strconv"
)

type Config struct {
	Environment string // dev, prod, local
	LogLevel    string // info, debug, error
	ApiEndpoint string // 外部APIのエンドポイント
	ApiTimeout  int    // HTTPタイムアウト(秒)
}

// NewConfig creates a new Config from environment variables with default values
func NewConfig() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "local"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		ApiEndpoint: getEnv("API_ENDPOINT", "https://api.example.com"),
		ApiTimeout:  getEnvAsInt("API_TIMEOUT", 30),
	}
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as int with a fallback default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
