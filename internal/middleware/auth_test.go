package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"trusioo_api/config"
	"trusioo_api/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestConfig 设置测试配置
func setupTestConfig() {
	config.AppConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:        "test-secret-key",
			RefreshSecret: "test-refresh-secret-key",
			AccessExpire:  3600,
			RefreshExpire: 86400,
		},
	}
}

// generateTestToken 生成测试用的 JWT token
func generateTestToken(userID int64, email, role, userType string) (string, error) {
	return auth.GenerateAccessToken(userID, email, role, userType)
}

// setupTestRouter 设置测试路由
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestAuthMiddleware(t *testing.T) {
	setupTestConfig()

	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
		expectedBody   string
		checkContext   func(*gin.Context)
	}{
		{
			name: "成功认证用户",
			setupAuth: func() string {
				token, _ := generateTestToken(1, "user@example.com", "user", "user")
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			checkContext: func(c *gin.Context) {
				userID, exists := c.Get("user_id")
				assert.True(t, exists)
				assert.Equal(t, int64(1), userID)

				userEmail, exists := c.Get("user_email")
				assert.True(t, exists)
				assert.Equal(t, "user@example.com", userEmail)

				userType, exists := c.Get("user_type")
				assert.True(t, exists)
				assert.Equal(t, "user", userType)
			},
		},
		{
			name: "缺少 Authorization header",
			setupAuth: func() string {
				return ""
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authorization header required",
		},
		{
			name: "无效的 Authorization header 格式",
			setupAuth: func() string {
				return "InvalidFormat"
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid authorization header format",
		},
		{
			name: "无效的 token",
			setupAuth: func() string {
				return "Bearer invalid-token"
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid or expired token",
		},
		{
			name: "管理员用户被拒绝",
			setupAuth: func() string {
				token, _ := generateTestToken(1, "admin@example.com", "admin", "admin")
				return "Bearer " + token
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.Use(AuthMiddleware())
			router.GET("/test", func(c *gin.Context) {
				if tt.checkContext != nil {
					tt.checkContext(c)
				}
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if authHeader := tt.setupAuth(); authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestOptionalAuth(t *testing.T) {
	setupTestConfig()

	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
		checkContext   func(*gin.Context)
	}{
		{
			name: "有效 token",
			setupAuth: func() string {
				token, _ := generateTestToken(1, "user@example.com", "user", "user")
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			checkContext: func(c *gin.Context) {
				userID, exists := c.Get("user_id")
				assert.True(t, exists)
				assert.Equal(t, int64(1), userID)
			},
		},
		{
			name: "无 Authorization header",
			setupAuth: func() string {
				return ""
			},
			expectedStatus: http.StatusOK,
			checkContext: func(c *gin.Context) {
				_, exists := c.Get("user_id")
				assert.False(t, exists)
			},
		},
		{
			name: "无效 token 格式",
			setupAuth: func() string {
				return "InvalidFormat"
			},
			expectedStatus: http.StatusOK,
			checkContext: func(c *gin.Context) {
				_, exists := c.Get("user_id")
				assert.False(t, exists)
			},
		},
		{
			name: "无效 token",
			setupAuth: func() string {
				return "Bearer invalid-token"
			},
			expectedStatus: http.StatusOK,
			checkContext: func(c *gin.Context) {
				_, exists := c.Get("user_id")
				assert.False(t, exists)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.Use(OptionalAuth())
			router.GET("/test", func(c *gin.Context) {
				if tt.checkContext != nil {
					tt.checkContext(c)
				}
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if authHeader := tt.setupAuth(); authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAdminAuthMiddleware(t *testing.T) {
	setupTestConfig()

	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
		expectedBody   string
		checkContext   func(*gin.Context)
	}{
		{
			name: "成功认证管理员",
			setupAuth: func() string {
				token, _ := generateTestToken(1, "admin@example.com", "admin", "admin")
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			checkContext: func(c *gin.Context) {
				userID, exists := c.Get("user_id")
				assert.True(t, exists)
				assert.Equal(t, int64(1), userID)

				userType, exists := c.Get("user_type")
				assert.True(t, exists)
				assert.Equal(t, "admin", userType)
			},
		},
		{
			name: "普通用户被拒绝",
			setupAuth: func() string {
				token, _ := generateTestToken(1, "user@example.com", "user", "user")
				return "Bearer " + token
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Admin access required",
		},
		{
			name: "缺少 Authorization header",
			setupAuth: func() string {
				return ""
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authorization header required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.Use(AdminAuthMiddleware())
			router.GET("/test", func(c *gin.Context) {
				if tt.checkContext != nil {
					tt.checkContext(c)
				}
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if authHeader := tt.setupAuth(); authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestSuperAdminMiddleware(t *testing.T) {
	setupTestConfig()

	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
		expectedBody   string
		checkContext   func(*gin.Context)
	}{
		{
			name: "成功认证超级管理员",
			setupAuth: func() string {
				token, _ := generateTestToken(1, "superadmin@example.com", "super_admin", "admin")
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			checkContext: func(c *gin.Context) {
				userRole, exists := c.Get("user_role")
				assert.True(t, exists)
				assert.Equal(t, "super_admin", userRole)

				userType, exists := c.Get("user_type")
				assert.True(t, exists)
				assert.Equal(t, "admin", userType)
			},
		},
		{
			name: "普通管理员被拒绝",
			setupAuth: func() string {
				token, _ := generateTestToken(1, "admin@example.com", "admin", "admin")
				return "Bearer " + token
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Super admin access required",
		},
		{
			name: "普通用户被拒绝",
			setupAuth: func() string {
				token, _ := generateTestToken(1, "user@example.com", "user", "user")
				return "Bearer " + token
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Super admin access required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.Use(SuperAdminMiddleware())
			router.GET("/test", func(c *gin.Context) {
				if tt.checkContext != nil {
					tt.checkContext(c)
				}
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if authHeader := tt.setupAuth(); authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

// TestTokenExpiration 测试 token 过期
func TestTokenExpiration(t *testing.T) {
	setupTestConfig()

	// 临时修改配置，使 token 立即过期
	originalExpire := config.AppConfig.JWT.AccessExpire
	config.AppConfig.JWT.AccessExpire = -1 // 设置为负数，使 token 立即过期
	defer func() {
		config.AppConfig.JWT.AccessExpire = originalExpire
	}()

	token, err := generateTestToken(1, "user@example.com", "user", "user")
	require.NoError(t, err)

	// 等待一小段时间确保 token 过期
	time.Sleep(100 * time.Millisecond)

	router := setupTestRouter()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}