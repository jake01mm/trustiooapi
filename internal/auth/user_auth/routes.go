package user_auth

import (
	"trusioo_api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	auth := router.Group("/auth")
	{
		// 公开路由 - 不需要认证
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.RefreshToken)

		// 需要认证的路由
		authRoutes := auth.Group("")
		authRoutes.Use(middleware.AuthMiddleware())
		{
			authRoutes.GET("/profile", handler.GetProfile)
		}
	}
}
