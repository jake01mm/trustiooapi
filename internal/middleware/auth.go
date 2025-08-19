package middleware

import (
	"strings"

	"trusioo_api/internal/common"
	"trusioo_api/pkg/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 用户认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			common.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := tokenParts[1]
		claims, err := auth.ValidateAccessToken(token)
		if err != nil {
			common.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 检查用户类型
		if claims.UserType != "user" {
			common.Forbidden(c, "Access denied")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("user_type", claims.UserType)

		c.Next()
	}
}

// AdminAuthMiddleware 管理员认证中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			common.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := tokenParts[1]
		claims, err := auth.ValidateAccessToken(token)
		if err != nil {
			common.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 检查用户类型
		if claims.UserType != "admin" {
			common.Forbidden(c, "Admin access required")
			c.Abort()
			return
		}

		// 将管理员信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("user_type", claims.UserType)

		c.Next()
	}
}

// SuperAdminMiddleware 超级管理员认证中间件
func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			common.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := tokenParts[1]
		claims, err := auth.ValidateAccessToken(token)
		if err != nil {
			common.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 检查用户类型和角色
		if claims.UserType != "admin" || claims.Role != "super_admin" {
			common.Forbidden(c, "Super admin access required")
			c.Abort()
			return
		}

		// 将管理员信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("user_type", claims.UserType)

		c.Next()
	}
}