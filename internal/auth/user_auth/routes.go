package user_auth

import (
	"trusioo_api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	// 公开路由 - 不需要认证
	router.POST("/register", handler.Register)        // 注册用户（未激活）
	router.POST("/login", handler.Login)              // 登录第一步：验证email+password
	router.POST("/login/verify", handler.LoginVerify) // 登录第二步：验证登录验证码

	router.POST("/forgot-password", handler.ForgotPassword) // 忘记密码：发送重置验证码
	router.POST("/reset-password", handler.ResetPassword)   // 重置密码：验证码+新密码

	router.POST("/refresh", handler.RefreshToken)

	// 需要认证的路由
	authRoutes := router.Group("")
	authRoutes.Use(middleware.AuthMiddleware())
	{
		authRoutes.GET("/profile", handler.GetProfile)
	}
}
