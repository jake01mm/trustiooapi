package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	Server   ServerConfig
	CORS     CORSConfig
	Frontend FrontendConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type JWTConfig struct {
	Secret        string
	RefreshSecret string
	AccessExpire  int
	RefreshExpire int
}

type ServerConfig struct {
	Port string
	Env  string
}

type CORSConfig struct {
	Origins  []string
	AllowAll bool
}

type FrontendConfig struct {
	AppURL   string
	AdminURL string
}

var AppConfig *Config

func LoadConfig() error {
	// 尝试从多个可能的路径加载.env文件
	_ = godotenv.Load() // 当前目录
	_ = godotenv.Load(".env") // 当前目录的.env
	_ = godotenv.Load("../.env") // 上级目录的.env
	_ = godotenv.Load("../../.env") // 上两级目录的.env

	AppConfig = &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "trusioo_db"),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your_jwt_secret"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "your_refresh_secret"),
			AccessExpire:  getEnvAsInt("JWT_ACCESS_EXPIRE", 7200),
			RefreshExpire: getEnvAsInt("JWT_REFRESH_EXPIRE", 604800),
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		CORS: CORSConfig{
			Origins:  strings.Split(getEnv("CORS_ORIGINS", "http://localhost:3000"), ","),
			AllowAll: getEnvAsBool("CORS_ALLOW_ALL", false),
		},
		Frontend: FrontendConfig{
			AppURL:   getEnv("FRONTEND_APP_URL", "http://localhost:3000"),
			AdminURL: getEnv("FRONTEND_ADMIN_URL", "http://localhost:3001"),
		},
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}