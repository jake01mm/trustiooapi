package health

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"trusioo_api/pkg/database"
	"trusioo_api/pkg/logger"

	"github.com/gin-gonic/gin"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
	Uptime    string            `json:"uptime"`
}

var startTime = time.Now()

// HealthCheck 健康检查处理器
func HealthCheck(c *gin.Context) {
	logger.WithRequest(c.Request.Method, c.Request.URL.Path, c.ClientIP()).
		Info("Health check requested")

	services := make(map[string]string)

	// 检查数据库连接
	dbStatus := checkDatabase()
	services["database"] = dbStatus

	// 计算运行时间
	uptime := time.Since(startTime)

	// 确定整体状态
	status := "healthy"
	for _, serviceStatus := range services {
		if serviceStatus != "healthy" {
			status = "unhealthy"
			break
		}
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0", // 可以从配置或构建信息中获取
		Services:  services,
		Uptime:    uptime.String(),
	}

	if status == "healthy" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// ReadinessCheck 就绪检查处理器
func ReadinessCheck(c *gin.Context) {
	logger.WithRequest(c.Request.Method, c.Request.URL.Path, c.ClientIP()).
		Info("Readiness check requested")

	// 检查所有关键服务
	services := make(map[string]string)
	services["database"] = checkDatabase()

	// 检查是否所有服务都就绪
	ready := true
	for _, serviceStatus := range services {
		if serviceStatus != "healthy" {
			ready = false
			break
		}
	}

	response := map[string]interface{}{
		"ready":     ready,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"services":  services,
	}

	if ready {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// LivenessCheck 存活检查处理器
func LivenessCheck(c *gin.Context) {
	// 简单的存活检查，只要服务能响应就表示存活
	c.JSON(http.StatusOK, map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
	})
}

// MetricsCheck 指标检查处理器
func MetricsCheck(c *gin.Context) {
	logger.WithRequest(c.Request.Method, c.Request.URL.Path, c.ClientIP()).
		Info("Metrics check requested")

	// 这里可以添加更多的系统指标
	metrics := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
		"uptime_seconds": time.Since(startTime).Seconds(),
		"database": map[string]interface{}{
			"status": checkDatabase(),
		},
		"memory": getMemoryUsage(),
		"goroutines": getGoroutineCount(),
	}

	c.JSON(http.StatusOK, metrics)
}

// checkDatabase 检查数据库连接状态
func checkDatabase() string {
	// 使用数据库包的健康检查函数
	if err := database.HealthCheck(); err != nil {
		logger.WithError(err).Error("Database health check failed")
		return "unhealthy"
	}

	// 检查数据库是否可以执行查询
	db := database.GetStdDB()
	if db == nil {
		return "unavailable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		logger.WithError(err).Error("Database query test failed")
		return "unhealthy"
	}

	return "healthy"
}

// getMemoryUsage 获取内存使用情况
func getMemoryUsage() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"alloc_mb":      bToMb(m.Alloc),
		"total_alloc_mb": bToMb(m.TotalAlloc),
		"sys_mb":        bToMb(m.Sys),
		"num_gc":        m.NumGC,
	}
}

// getGoroutineCount 获取协程数量
func getGoroutineCount() int {
	return runtime.NumGoroutine()
}

// bToMb 字节转MB
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}