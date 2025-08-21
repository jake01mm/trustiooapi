package monitoring

import (
	"github.com/gin-gonic/gin"
	"trusioo_api/internal/middleware"
)

func SetupRoutes(r *gin.RouterGroup, handler *Handler) {
	// 所有监控接口都需要管理员权限
	monitoringRoutes := r.Group("/monitoring")
	monitoringRoutes.Use(middleware.AdminAuthMiddleware())
	
	{
		// 基础指标
		monitoringRoutes.GET("/metrics", handler.GetMetrics)
		monitoringRoutes.GET("/metrics/history", handler.GetHistoricalMetrics)
		
		// 健康检查 - 不需要认证，供负载均衡器使用
		monitoringRoutes.GET("/health", handler.HealthCheck)
		
		// 告警
		monitoringRoutes.GET("/alerts", handler.GetAlerts)
		
		// 仪表板
		monitoringRoutes.GET("/dashboard", handler.GetDashboard)
		
		// 性能报告
		monitoringRoutes.GET("/report", handler.GetPerformanceReport)
	}
}

// 设置公开的健康检查路由（不需要认证）
func SetupPublicHealthRoutes(r *gin.Engine, handler *Handler) {
	r.GET("/health", handler.HealthCheck)
	r.GET("/healthz", handler.HealthCheck) // Kubernetes风格的健康检查
}