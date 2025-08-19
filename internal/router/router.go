package router

import (
	"net/http"

	"trusioo_api/config"
	admin_auth "trusioo_api/internal/auth/admin_auth"
	user_auth "trusioo_api/internal/auth/user_auth"
	"trusioo_api/internal/auth/verification"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	// 设置运行模式
	if config.AppConfig.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 配置 CORS
	corsConfig := cors.DefaultConfig()
	if config.AppConfig.CORS.AllowAll {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = config.AppConfig.CORS.Origins
	}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	r.Use(cors.New(corsConfig))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"message": "Trusioo API is running",
		})
	})

	// API 路由组
	api := r.Group("/api/v1")

	// 初始化服务
	userRepo := user_auth.NewRepository()
	authService := user_auth.NewService(userRepo)
	adminService := admin_auth.NewService()

	// 初始化处理器
	authHandler := user_auth.NewHandler(authService)
	adminHandler := admin_auth.NewHandler(adminService)
	verificationHandler := verification.NewHandler()

	// 注册路由
	user_auth.RegisterRoutes(api, authHandler)
	admin_auth.RegisterRoutes(api, adminHandler)
	verification.RegisterRoutes(api, verificationHandler)

	return r
}