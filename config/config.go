package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Server   ServerConfig
	CORS     CORSConfig
	Frontend FrontendConfig
	Security SecurityConfig
	RateLimit RateLimitConfig
	Request  RequestConfig
	ThirdParty ThirdPartyConfig
}

type DatabaseConfig struct {
	Host           string
	Port           string
	User           string
	Password       string
	Name           string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxLifetime int
}

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  int
	ReadTimeout  int
	WriteTimeout int
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

type SecurityConfig struct {
	EnableHTTPS        bool
	TLSCertFile        string
	TLSKeyFile         string
	EnableSecureHeaders bool
	TrustedProxies     []string
}

type RateLimitConfig struct {
	Enabled           bool
	Requests          int
	Window            int
	AuthRequests      int
	AuthWindow        int
}

type RequestConfig struct {
	Timeout        int
	MaxSize        int64
	EnableRequestID bool
}

type ThirdPartyConfig struct {
	CardDetectionEnabled    bool
	CardDetectionHost       string
	CardDetectionAppID      string
	CardDetectionAppSecret  string
	CardDetectionTimeout    int
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
			Host:           getEnv("DB_HOST", "localhost"),
			Port:           getEnv("DB_PORT", "5432"),
			User:           getEnv("DB_USER", "postgres"),
			Password:       getEnv("DB_PASSWORD", ""),
			Name:           getEnv("DB_NAME", "trusioo_db"),
			MaxOpenConns:   getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:   getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 300),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvAsInt("REDIS_DB", 0),
			PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
			MaxRetries:   getEnvAsInt("REDIS_MAX_RETRIES", 3),
			DialTimeout:  getEnvAsInt("REDIS_DIAL_TIMEOUT", 5),
			ReadTimeout:  getEnvAsInt("REDIS_READ_TIMEOUT", 3),
			WriteTimeout: getEnvAsInt("REDIS_WRITE_TIMEOUT", 3),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "change-this-to-a-secure-secret-key"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "change-this-to-a-secure-refresh-key"),
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
		Security: SecurityConfig{
			EnableHTTPS:        getEnvAsBool("FORCE_HTTPS", false),
			TLSCertFile:        getEnv("TLS_CERT_FILE", ""),
			TLSKeyFile:         getEnv("TLS_KEY_FILE", ""),
			EnableSecureHeaders: getEnvAsBool("ENABLE_SECURE_HEADERS", true),
			TrustedProxies:     strings.Split(getEnv("TRUSTED_PROXIES", ""), ","),
		},
		RateLimit: RateLimitConfig{
			Enabled:      getEnvAsBool("RATE_LIMIT_ENABLED", true),
			Requests:     getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			Window:       getEnvAsInt("RATE_LIMIT_WINDOW", 60),
			AuthRequests: getEnvAsInt("AUTH_RATE_LIMIT_REQUESTS", 10),
			AuthWindow:   getEnvAsInt("AUTH_RATE_LIMIT_WINDOW", 60),
		},
		Request: RequestConfig{
			Timeout:        getEnvAsInt("REQUEST_TIMEOUT", 30),
			MaxSize:        getEnvAsInt64("MAX_REQUEST_SIZE", 10485760), // 10MB
			EnableRequestID: getEnvAsBool("ENABLE_REQUEST_ID", true),
		},
		ThirdParty: ThirdPartyConfig{
			CardDetectionEnabled:   getEnvAsBool("CARD_DETECTION_ENABLED", false),
			CardDetectionHost:      getEnv("CARD_DETECTION_HOST", ""),
			CardDetectionAppID:     getEnv("CARD_DETECTION_APP_ID", ""),
			CardDetectionAppSecret: getEnv("CARD_DETECTION_APP_SECRET", ""),
			CardDetectionTimeout:   getEnvAsInt("CARD_DETECTION_TIMEOUT", 30),
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

func getEnvAsInt64(name string, defaultVal int64) int64 {
	valueStr := getEnv(name, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultVal
}