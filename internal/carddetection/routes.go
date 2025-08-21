package carddetection

import (
	"trusioo_api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册卡片检测路由
func RegisterRoutes(r *gin.RouterGroup, handler *Handler) {
	// 卡片检测路由组
	cardDetection := r.Group("/card-detection")
	
	// 应用验证中间件
	cardDetection.Use(middleware.RequestValidationMiddleware(middleware.NewValidator()))
	// 所有端点都需要认证和权限验证
	cardDetection.Use(middleware.AuthMiddleware()) // 添加认证中间件
	
	// CD产品端点
	cardDetection.GET("/cd_products", handler.GetCDProducts)
	
	// CD区域端点
	cardDetection.GET("/cd_regions", handler.GetCDRegions)
	
	// 卡片检测端点
	cardDetection.POST("/check", handler.CheckCard)
	cardDetection.POST("/result", handler.CheckCardResult)
	
	// 历史记录和统计端点
	cardDetection.GET("/history", handler.GetUserHistory)
	cardDetection.GET("/records/:id", handler.GetRecordDetail)
	cardDetection.GET("/stats", handler.GetUserStats)
	cardDetection.GET("/summary", handler.GetUserSummary)
}