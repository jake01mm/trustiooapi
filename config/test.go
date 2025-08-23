package config

import (
	"os"
	"strconv"
)

// TestConfig 测试环境配置
type TestConfig struct {
	DB TestDBConfig `json:"db"`
	JWT TestJWTConfig `json:"jwt"`
	Redis TestRedisConfig `json:"redis"`
}

type TestDBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SSLMode  string `json:"ssl_mode"`
}

type TestJWTConfig struct {
	Secret        string `json:"secret"`
	RefreshSecret string `json:"refresh_secret"`
	AccessExpire  int    `json:"access_expire"`
	RefreshExpire int    `json:"refresh_expire"`
}

type TestRedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// LoadTestConfig 加载测试配置
func LoadTestConfig() *TestConfig {
	return &TestConfig{
		DB: TestDBConfig{
			Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
			Port:     getEnvAsIntOrDefault("TEST_DB_PORT", 5432),
			User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
			Password: getEnvOrDefault("TEST_DB_PASSWORD", "password"),
			Name:     getEnvOrDefault("TEST_DB_NAME", "trusioo_test"),
			SSLMode:  getEnvOrDefault("TEST_DB_SSL_MODE", "disable"),
		},
		JWT: TestJWTConfig{
			Secret:        getEnvOrDefault("TEST_JWT_SECRET", "test_jwt_secret_key_for_testing"),
			RefreshSecret: getEnvOrDefault("TEST_JWT_REFRESH_SECRET", "test_jwt_refresh_secret_key_for_testing"),
			AccessExpire:  getEnvAsIntOrDefault("TEST_JWT_ACCESS_EXPIRE", 3600),
			RefreshExpire: getEnvAsIntOrDefault("TEST_JWT_REFRESH_EXPIRE", 86400),
		},
		Redis: TestRedisConfig{
			Host:     getEnvOrDefault("TEST_REDIS_HOST", "localhost"),
			Port:     getEnvAsIntOrDefault("TEST_REDIS_PORT", 6379),
			Password: getEnvOrDefault("TEST_REDIS_PASSWORD", ""),
			DB:       getEnvAsIntOrDefault("TEST_REDIS_DB", 1),
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}