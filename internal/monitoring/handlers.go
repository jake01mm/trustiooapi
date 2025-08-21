package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"trusioo_api/pkg/monitoring"
)

type Handler struct {
	monitor *monitoring.SystemMonitor
}

func NewHandler(monitor *monitoring.SystemMonitor) *Handler {
	return &Handler{
		monitor: monitor,
	}
}

// 获取当前系统指标
func (h *Handler) GetMetrics(c *gin.Context) {
	metrics := h.monitor.GetMetrics()
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    metrics,
	})
}

// 获取历史指标
func (h *Handler) GetHistoricalMetrics(c *gin.Context) {
	hoursStr := c.DefaultQuery("hours", "1")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		hours = 1
	}
	
	if hours > 24 {
		hours = 24 // 限制最多24小时
	}
	
	metrics, err := h.monitor.GetHistoricalMetrics(c.Request.Context(), hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get historical metrics",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"hours":   hours,
			"metrics": metrics,
		},
	})
}

// 获取系统告警
func (h *Handler) GetAlerts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	
	if limit > 100 {
		limit = 100 // 限制最多100条
	}
	
	alerts, err := h.monitor.GetAlerts(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get alerts",
			"error":   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"limit":  limit,
			"alerts": alerts,
		},
	})
}

// 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	health := h.monitor.HealthCheck()
	
	// 根据健康状态设置HTTP状态码
	status := http.StatusOK
	if healthStatus, ok := health["status"].(string); ok {
		if healthStatus == "warning" {
			status = http.StatusPartialContent // 206
		} else if healthStatus == "critical" {
			status = http.StatusServiceUnavailable // 503
		}
	}
	
	c.JSON(status, gin.H{
		"code":    0,
		"message": "success",
		"data":    health,
	})
}

// 系统状态仪表板数据
func (h *Handler) GetDashboard(c *gin.Context) {
	metrics := h.monitor.GetMetrics()
	
	// 计算各种指标
	var errorRate, successRate float64
	if metrics.TotalRequests > 0 {
		errorRate = float64(metrics.ErrorRequests) / float64(metrics.TotalRequests) * 100
		successRate = float64(metrics.SuccessRequests) / float64(metrics.TotalRequests) * 100
	}
	
	var uploadSuccessRate float64
	if metrics.TotalUploads > 0 {
		uploadSuccessRate = float64(metrics.SuccessUploads) / float64(metrics.TotalUploads) * 100
	}
	
	dashboard := gin.H{
		"summary": gin.H{
			"total_requests":      metrics.TotalRequests,
			"success_rate":        successRate,
			"error_rate":          errorRate,
			"average_latency":     metrics.AverageLatency.Milliseconds(),
			"total_uploads":       metrics.TotalUploads,
			"upload_success_rate": uploadSuccessRate,
			"cache_hit_rate":      metrics.CacheHitRate * 100,
		},
		"performance": gin.H{
			"api_latency":     metrics.AverageLatency.Milliseconds(),
			"upload_latency":  metrics.UploadLatency.Milliseconds(),
			"cdn_latency":     metrics.CDNLatency.Milliseconds(),
			"active_connections": metrics.ActiveConns,
		},
		"resources": gin.H{
			"cpu_usage":    metrics.CPUUsage,
			"memory_usage": metrics.MemoryUsage,
			"disk_usage":   metrics.DiskUsage,
		},
		"queues": gin.H{
			"queue_length":    metrics.QueueLength,
			"processed_tasks": metrics.ProcessedTasks,
			"failed_tasks":    metrics.FailedTasks,
		},
		"cache": gin.H{
			"hit_rate":  metrics.CacheHitRate * 100,
			"hits":      metrics.CacheHits,
			"misses":    metrics.CacheMisses,
		},
		"timestamp": metrics.Timestamp,
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    dashboard,
	})
}

// 性能报告
func (h *Handler) GetPerformanceReport(c *gin.Context) {
	hoursStr := c.DefaultQuery("hours", "1")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		hours = 1
	}
	
	metrics, err := h.monitor.GetHistoricalMetrics(c.Request.Context(), hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to generate performance report",
			"error":   err.Error(),
		})
		return
	}
	
	if len(metrics) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
			"data": gin.H{
				"report": "No data available for the specified period",
			},
		})
		return
	}
	
	// 计算性能报告
	report := generatePerformanceReport(metrics)
	
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    report,
	})
}

func generatePerformanceReport(metrics []monitoring.SystemMetrics) gin.H {
	if len(metrics) == 0 {
		return gin.H{"error": "no data"}
	}
	
	latest := metrics[0]
	oldest := metrics[len(metrics)-1]
	
	// 计算平均值和趋势
	var totalLatency time.Duration
	var totalRequests, totalErrors int64
	var totalCacheHitRate float64
	
	for _, m := range metrics {
		totalLatency += m.AverageLatency
		totalRequests += m.TotalRequests
		totalErrors += m.ErrorRequests
		totalCacheHitRate += m.CacheHitRate
	}
	
	avgLatency := totalLatency / time.Duration(len(metrics))
	avgCacheHitRate := totalCacheHitRate / float64(len(metrics))
	
	// 计算增长率
	var requestGrowth, errorGrowth float64
	if oldest.TotalRequests > 0 {
		requestGrowth = float64(latest.TotalRequests-oldest.TotalRequests) / float64(oldest.TotalRequests) * 100
	}
	if oldest.ErrorRequests > 0 {
		errorGrowth = float64(latest.ErrorRequests-oldest.ErrorRequests) / float64(oldest.ErrorRequests) * 100
	}
	
	return gin.H{
		"period": gin.H{
			"start": oldest.Timestamp,
			"end":   latest.Timestamp,
			"data_points": len(metrics),
		},
		"performance": gin.H{
			"average_latency":     avgLatency.Milliseconds(),
			"latest_latency":      latest.AverageLatency.Milliseconds(),
			"latency_trend":       getLatencyTrend(metrics),
			"average_cache_hit_rate": avgCacheHitRate * 100,
		},
		"traffic": gin.H{
			"total_requests":   latest.TotalRequests,
			"request_growth":   requestGrowth,
			"error_rate":       getErrorRate(latest),
			"error_growth":     errorGrowth,
		},
		"uploads": gin.H{
			"total_uploads":    latest.TotalUploads,
			"success_rate":     getUploadSuccessRate(latest),
		},
		"recommendations": generateRecommendations(latest, avgLatency, avgCacheHitRate),
	}
}

func getLatencyTrend(metrics []monitoring.SystemMetrics) string {
	if len(metrics) < 2 {
		return "stable"
	}
	
	recent := metrics[0].AverageLatency
	older := metrics[len(metrics)-1].AverageLatency
	
	if recent > older*110/100 {
		return "increasing"
	} else if recent < older*90/100 {
		return "decreasing"
	}
	return "stable"
}

func getErrorRate(metrics monitoring.SystemMetrics) float64 {
	if metrics.TotalRequests == 0 {
		return 0
	}
	return float64(metrics.ErrorRequests) / float64(metrics.TotalRequests) * 100
}

func getUploadSuccessRate(metrics monitoring.SystemMetrics) float64 {
	if metrics.TotalUploads == 0 {
		return 0
	}
	return float64(metrics.SuccessUploads) / float64(metrics.TotalUploads) * 100
}

func generateRecommendations(metrics monitoring.SystemMetrics, avgLatency time.Duration, avgCacheHitRate float64) []string {
	var recommendations []string
	
	// 延迟相关建议
	if avgLatency > 2*time.Second {
		recommendations = append(recommendations, "Consider optimizing API response times - average latency is high")
	}
	
	// 错误率相关建议
	errorRate := getErrorRate(metrics)
	if errorRate > 5 {
		recommendations = append(recommendations, "High error rate detected - investigate error causes")
	}
	
	// 缓存相关建议
	if avgCacheHitRate < 0.7 {
		recommendations = append(recommendations, "Low cache hit rate - consider cache warming or optimization")
	}
	
	// 队列相关建议
	if metrics.QueueLength > 1000 {
		recommendations = append(recommendations, "High queue length - consider scaling workers")
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "System performance is optimal")
	}
	
	return recommendations
}