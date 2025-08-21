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

// DetailedHealthCheck 详细健康检查处理器
func DetailedHealthCheck(c *gin.Context) {
	logger.WithRequest(c.Request.Method, c.Request.URL.Path, c.ClientIP()).
		Info("Detailed health check requested")

	services := make(map[string]interface{})
	
	// 数据库详细检查
	dbHealth := getDetailedDatabaseHealth()
	services["database"] = dbHealth
	
	// Redis检查（如果配置了）
	redisHealth := getRedisHealth()
	if redisHealth != nil {
		services["redis"] = redisHealth
	}
	
	// 系统指标
	services["system"] = map[string]interface{}{
		"memory":     getMemoryUsage(),
		"goroutines": getGoroutineCount(),
		"uptime":     time.Since(startTime).String(),
		"uptime_seconds": time.Since(startTime).Seconds(),
	}

	// 计算整体健康状态
	overallStatus := "healthy"
	for _, serviceData := range services {
		if serviceMap, ok := serviceData.(map[string]interface{}); ok {
			if status, exists := serviceMap["status"]; exists && status != "healthy" {
				overallStatus = "degraded"
				if status == "unhealthy" {
					overallStatus = "unhealthy"
					break
				}
			}
		}
	}

	response := map[string]interface{}{
		"status":    overallStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).String(),
		"services":  services,
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if overallStatus == "degraded" {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, response)
}

// DatabaseHealthCheck 数据库专用健康检查处理器
func DatabaseHealthCheck(c *gin.Context) {
	logger.WithRequest(c.Request.Method, c.Request.URL.Path, c.ClientIP()).
		Info("Database health check requested")

	dbHealth := getDetailedDatabaseHealth()
	
	statusCode := http.StatusOK
	if dbHealth["status"] != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, dbHealth)
}

// RedisHealthCheck Redis专用健康检查处理器  
func RedisHealthCheck(c *gin.Context) {
	logger.WithRequest(c.Request.Method, c.Request.URL.Path, c.ClientIP()).
		Info("Redis health check requested")

	redisHealth := getRedisHealth()
	
	if redisHealth == nil {
		c.JSON(http.StatusNotImplemented, map[string]interface{}{
			"status":  "not_configured",
			"message": "Redis is not configured for this application",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	statusCode := http.StatusOK
	if redisHealth["status"] != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, redisHealth)
}

// getDetailedDatabaseHealth 获取详细的数据库健康信息
func getDetailedDatabaseHealth() map[string]interface{} {
	health := map[string]interface{}{
		"status": checkDatabase(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// 添加连接池统计信息
	if DB := database.GetStdDB(); DB != nil {
		stats := DB.Stats()
		health["connection_pool"] = map[string]interface{}{
			"max_open_connections":     stats.MaxOpenConnections,
			"open_connections":         stats.OpenConnections,
			"in_use":                  stats.InUse,
			"idle":                    stats.Idle,
			"wait_count":              stats.WaitCount,
			"wait_duration_ms":        stats.WaitDuration.Milliseconds(),
			"max_idle_closed":         stats.MaxIdleClosed,
			"max_idle_time_closed":    stats.MaxIdleTimeClosed,
			"max_lifetime_closed":     stats.MaxLifetimeClosed,
		}

		// 测试查询响应时间
		start := time.Now()
		var result int
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		err := DB.QueryRowContext(ctx, "SELECT 1").Scan(&result)
		queryDuration := time.Since(start)
		
		health["query_test"] = map[string]interface{}{
			"duration_ms": queryDuration.Milliseconds(),
			"success":    err == nil,
		}
		
		if err != nil {
			health["query_test"].(map[string]interface{})["error"] = err.Error()
			health["status"] = "unhealthy"
		}
	} else {
		health["status"] = "unavailable"
		health["error"] = "Database connection is nil"
	}

	return health
}

// getRedisHealth 获取Redis健康信息（如果配置了Redis）
func getRedisHealth() map[string]interface{} {
	// 注意：这里假设Redis是可选的
	// 如果没有配置Redis，返回nil
	// 实际项目中，你可能需要根据配置来决定是否检查Redis
	
	// TODO: 实现Redis健康检查
	// 目前返回nil表示Redis未配置
	return nil
	
	// 以下是Redis健康检查的示例实现（当你配置了Redis时取消注释）
	/*
	health := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	
	// 这里需要你的Redis客户端实例
	// rdb := redis.GetClient() // 假设你有这样的函数
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	start := time.Now()
	// pong, err := rdb.Ping(ctx).Result()
	// pingDuration := time.Since(start)
	
	health["ping_duration_ms"] = pingDuration.Milliseconds()
	
	if err != nil {
		health["status"] = "unhealthy"
		health["error"] = err.Error()
	} else {
		health["status"] = "healthy"
		health["ping_response"] = pong
	}
	
	return health
	*/
}

// bToMb 字节转MB
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}