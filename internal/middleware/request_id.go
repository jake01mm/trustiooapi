package middleware

import (
	"trusioo_api/pkg/logger"
	"trusioo_api/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取现有的请求ID，或生成新的
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = utils.GenerateUUID()
		}

		// 设置请求ID到上下文和响应头
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// LoggingMiddleware 请求日志中间件
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		requestID, _ := param.Keys["request_id"].(string)
		
		// 使用结构化日志记录请求
		logger.WithFields(logrus.Fields{
			"request_id":   requestID,
			"method":       param.Method,
			"path":         param.Path,
			"status":       param.StatusCode,
			"latency":      param.Latency.String(),
			"ip":           param.ClientIP,
			"user_agent":   param.Request.UserAgent(),
			"error":        param.ErrorMessage,
			"body_size":    param.BodySize,
		}).Info("HTTP Request")

		return ""
	})
}

// RecoveryMiddleware 自定义恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID, _ := c.Get("request_id")
		
		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"ip":         c.ClientIP,
			"panic":      recovered,
		}).Error("Panic recovered")

		c.JSON(500, gin.H{
			"code":       500,
			"message":    "Internal server error",
			"request_id": requestID,
		})
	})
}

// CorrelationMiddleware 关联ID中间件（用于微服务追踪）
func CorrelationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取关联ID
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = utils.GenerateUUID()
		}

		// 设置到上下文和响应头
		c.Set("correlation_id", correlationID)
		c.Header("X-Correlation-ID", correlationID)

		c.Next()
	}
}