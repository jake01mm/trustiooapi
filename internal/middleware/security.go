package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"trusioo_api/config"
	"trusioo_api/internal/common"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 内容类型嗅探保护
		c.Header("X-Content-Type-Options", "nosniff")
		
		// 点击劫持保护
		c.Header("X-Frame-Options", "DENY")
		
		// XSS 保护
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// HTTPS 严格传输安全
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		
		// 内容安全策略
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"media-src 'self'; " +
			"object-src 'none'; " +
			"child-src 'none'; " +
			"worker-src 'none'; " +
			"frame-ancestors 'none'; " +
			"form-action 'self'; " +
			"base-uri 'self'"
		c.Header("Content-Security-Policy", csp)
		
		// 引用来源策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// 权限策略
		permissions := "geolocation=(), microphone=(), camera=(), " +
			"payment=(), usb=(), magnetometer=(), gyroscope=(), " +
			"speaker=(), vibrate=(), fullscreen=(self)"
		c.Header("Permissions-Policy", permissions)
		
		// 移除服务器标识
		c.Header("Server", "")
		
		// 缓存控制（对于敏感数据）
		if strings.Contains(c.Request.URL.Path, "/api/auth/") ||
		   strings.Contains(c.Request.URL.Path, "/api/admin/") {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.AppConfig.CORS
		
		origin := c.Request.Header.Get("Origin")
		
		// 如果允许所有来源
		if cfg.AllowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			// 检查是否在允许列表中
			allowed := false
			for _, allowedOrigin := range cfg.Origins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}
			
			if allowed {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", 
			"Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, " +
			"Authorization, X-Requested-With, X-Request-ID, X-Correlation-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID, X-Correlation-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24小时
		
		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// HTTPSRedirectMiddleware HTTPS 重定向中间件
func HTTPSRedirectMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只在生产环境强制 HTTPS
		if config.AppConfig.Server.Env == "production" {
			if c.Request.Header.Get("X-Forwarded-Proto") != "https" {
				url := "https://" + c.Request.Host + c.Request.RequestURI
				c.Redirect(http.StatusMovedPermanently, url)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// ContentTypeValidationMiddleware 内容类型验证中间件
func ContentTypeValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对于 POST, PUT, PATCH 请求，验证内容类型
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			
			// 跳过文件上传
			if strings.Contains(contentType, "multipart/form-data") {
				c.Next()
				return
			}
			
			// 要求 JSON 内容类型
			if !strings.Contains(contentType, "application/json") {
				common.ValidationError(c, "Content-Type must be application/json")
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

// RequestSizeLimit 请求大小限制中间件
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, common.Response{
				Code:    413,
				Message: "Request entity too large",
			})
			c.Abort()
			return
		}
		
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// TimeoutMiddleware 请求超时中间件
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		
		// 替换请求上下文
		c.Request = c.Request.WithContext(ctx)
		
		// 创建一个通道来接收处理完成信号
		done := make(chan bool, 1)
		
		// 在 goroutine 中处理请求
		go func() {
			c.Next()
			done <- true
		}()
		
		// 等待处理完成或超时
		select {
		case <-done:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 请求超时
			c.JSON(http.StatusRequestTimeout, common.Response{
				Code:    408,
				Message: "Request timeout",
			})
			c.Abort()
			return
		}
	}
}

// IPWhitelistMiddleware IP 白名单中间件
func IPWhitelistMiddleware(whitelist []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(whitelist) == 0 {
			c.Next()
			return
		}
		
		clientIP := c.ClientIP()
		allowed := false
		
		for _, ip := range whitelist {
			if clientIP == ip {
				allowed = true
				break
			}
		}
		
		if !allowed {
			c.JSON(http.StatusForbidden, common.Response{
				Code:    403,
				Message: "Access denied from your IP address",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}