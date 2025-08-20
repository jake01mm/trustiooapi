package user_auth

import (
	"trusioo_api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	auth := router.Group("/auth")
	{
		// 公开路由 - 不需要认证
		auth.POST("/register", handler.Register)              // 注册用户（未激活）
		auth.POST("/login", handler.Login)                    // 登录第一步：验证email+password
		auth.POST("/login/verify", handler.LoginVerify)       // 登录第二步：验证登录验证码
		auth.POST("/refresh", handler.RefreshToken)

		// 需要认证的路由
		authRoutes := auth.Group("")
		authRoutes.Use(middleware.AuthMiddleware())
		{
			authRoutes.GET("/profile", handler.GetProfile)
		}
	}
}
