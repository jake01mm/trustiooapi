package admin

import (
	"trusioo_api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	admin := router.Group("/admin")
	{
		// 管理员认证相关路由 - 不需要认证
		auth := admin.Group("/auth")
		{
			auth.POST("/login", handler.Login)
			auth.POST("/login/verify", handler.LoginVerify)
			auth.POST("/forgot-password", handler.ForgotPassword) // 管理员忘记密码：发送重置验证码
			auth.POST("/reset-password", handler.ResetPassword)   // 管理员重置密码：验证码+新密码
			auth.POST("/refresh", handler.RefreshToken)
		}

		// 需要管理员认证的路由
		adminRoutes := admin.Group("")
		adminRoutes.Use(middleware.AdminAuthMiddleware())
		{
			// 管理员个人信息
			adminRoutes.GET("/profile", handler.GetProfile)

			// 用户管理
			users := adminRoutes.Group("/users")
			{
				users.GET("/stats", handler.GetUserStats)
				users.GET("", handler.GetUserList)
				users.GET("/:id", handler.GetUserDetail)
			}
		}
	}
}
