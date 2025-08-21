package router

import (
	"time"

	"trusioo_api/config"
	admin_auth "trusioo_api/internal/auth/admin_auth"
	user_auth "trusioo_api/internal/auth/user_auth"
	"trusioo_api/internal/carddetection"
	"trusioo_api/internal/health"
	"trusioo_api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	// 设置运行模式
	if config.AppConfig.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建不使用默认中间件的路由引擎
	r := gin.New()

	// 核心中间件
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggingMiddleware())

	// 请求 ID 追踪
	if config.AppConfig.Request.EnableRequestID {
		r.Use(middleware.RequestIDMiddleware())
		r.Use(middleware.CorrelationMiddleware())
	}

	// 安全中间件
	if config.AppConfig.Security.EnableSecureHeaders {
		r.Use(middleware.SecurityHeadersMiddleware())
	}
	
	// HTTPS 重定向（生产环境）
	if config.AppConfig.Security.EnableHTTPS {
		r.Use(middleware.HTTPSRedirectMiddleware())
	}

	// CORS 配置
	r.Use(middleware.CORSMiddleware())

	// 请求大小限制
	r.Use(middleware.RequestSizeLimitMiddleware(config.AppConfig.Request.MaxSize))

	// 内容类型验证
	r.Use(middleware.ContentTypeValidationMiddleware())

	// 全局速率限制
	if config.AppConfig.RateLimit.Enabled {
		rateLimiter := middleware.NewRateLimiter()
		r.Use(middleware.RateLimitMiddleware(rateLimiter))
	}

	// 请求超时
	timeoutDuration := time.Duration(config.AppConfig.Request.Timeout) * time.Second
	r.Use(middleware.TimeoutMiddleware(timeoutDuration))

	// 健康检查端点（无需认证）
	r.GET("/health", health.HealthCheck)
	r.GET("/health/ready", health.ReadinessCheck)
	r.GET("/health/live", health.LivenessCheck)
	r.GET("/metrics", health.MetricsCheck)

	// API 路由组
	api := r.Group("/api/v1")

	// API 版本的健康检查端点
	healthGroup := api.Group("/health")
	{
		healthGroup.GET("", health.HealthCheck)           // /api/v1/health
		healthGroup.GET("/ready", health.ReadinessCheck)  // /api/v1/health/ready
		healthGroup.GET("/live", health.LivenessCheck)    // /api/v1/health/live
		healthGroup.GET("/metrics", health.MetricsCheck)  // /api/v1/health/metrics
		healthGroup.GET("/detailed", health.DetailedHealthCheck) // /api/v1/health/detailed
		healthGroup.GET("/database", health.DatabaseHealthCheck) // /api/v1/health/database
		healthGroup.GET("/redis", health.RedisHealthCheck)       // /api/v1/health/redis
	}

	// 认证相关路由（特殊速率限制）
	authGroup := api.Group("/auth")
	if config.AppConfig.RateLimit.Enabled {
		authGroup.Use(middleware.AuthRateLimitMiddleware())
	}

	// 初始化服务
	userRepo := user_auth.NewRepository()
	authService := user_auth.NewService(userRepo)
	adminService := admin_auth.NewService()

	// 初始化处理器
	authHandler := user_auth.NewHandler(authService)
	adminHandler := admin_auth.NewHandler(adminService)
	cardDetectionHandler := carddetection.NewHandler()

	// 注册路由
	user_auth.RegisterRoutes(authGroup, authHandler)
	admin_auth.RegisterRoutes(api, adminHandler)
	carddetection.RegisterRoutes(api, cardDetectionHandler)

	return r
}