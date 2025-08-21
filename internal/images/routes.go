package images

import (
	"github.com/gin-gonic/gin"
	"trusioo_api/internal/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, handler *Handler) {
	images := r.Group("/images")
	{
		// Public routes - 完全公开，无需认证
		images.GET("/public/:key", handler.GetImageByKey)
		
		// User routes - 需要用户认证
		userRoutes := images.Group("")
		userRoutes.Use(middleware.AuthMiddleware()) // 必须登录
		{
			userRoutes.POST("/upload", handler.UploadImage)
			userRoutes.GET("/", handler.ListImages)              // 只显示用户自己的图片
			userRoutes.GET("/:id", handler.GetImage)             // 只能查看自己的图片
			userRoutes.PUT("/:id/refresh", handler.RefreshURL)   // 只能刷新自己的图片URL
			userRoutes.DELETE("/:id", handler.DeleteImage)       // 只能删除自己的图片
		}
		
		// Admin routes - 需要管理员权限
		adminRoutes := images.Group("/admin")
		adminRoutes.Use(middleware.AdminAuthMiddleware()) // 必须是管理员
		{
			adminRoutes.GET("/", handler.AdminListImages)        // 管理员查看所有图片
			adminRoutes.GET("/:id", handler.AdminGetImage)       // 管理员查看任意图片
			adminRoutes.DELETE("/:id", handler.AdminDeleteImage) // 管理员删除任意图片
			adminRoutes.POST("/batch-delete", handler.AdminBatchDeleteImages) // 批量删除
		}
	}
}