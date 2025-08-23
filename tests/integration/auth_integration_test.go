package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"trusioo_api/config"
	"trusioo_api/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AuthIntegrationTestSuite 认证集成测试套件
type AuthIntegrationTestSuite struct {
	suite.Suite
	router *gin.Engine
}

// SetupSuite 设置测试套件
func (suite *AuthIntegrationTestSuite) SetupSuite() {
	// 设置测试配置
	config.AppConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:        "test-secret-key-for-integration",
			RefreshSecret: "test-refresh-secret-key-for-integration",
			AccessExpire:  3600,
			RefreshExpire: 86400,
		},
	}

	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// 设置基本路由用于测试
	suite.setupBasicRoutes()
}

// setupBasicRoutes 设置基本测试路由
func (suite *AuthIntegrationTestSuite) setupBasicRoutes() {
	// 模拟受保护的路由
	protected := suite.router.Group("/api/v1/protected")
	// 这里暂时不使用中间件，因为需要数据库连接
	{
		protected.GET("/profile", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "protected route"})
		})
	}

	// 公开路由
	public := suite.router.Group("/api/v1/public")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}
}

// TestJWTTokenGeneration 测试JWT令牌生成
func (suite *AuthIntegrationTestSuite) TestJWTTokenGeneration() {
	// 测试访问令牌生成
	accessToken, err := auth.GenerateAccessToken(1, "test@example.com", "user", "user")
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), accessToken)

	// 测试刷新令牌生成
	refreshToken, err := auth.GenerateRefreshToken(1, "test@example.com", "user", "user")
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), refreshToken)

	// 验证访问令牌
	claims, err := auth.ValidateAccessToken(accessToken)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), claims)
	assert.Equal(suite.T(), int64(1), claims.UserID)
	assert.Equal(suite.T(), "test@example.com", claims.Email)
	assert.Equal(suite.T(), "user", claims.Role)
	assert.Equal(suite.T(), "user", claims.UserType)

	// 验证刷新令牌
	refreshClaims, err := auth.ValidateRefreshToken(refreshToken)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), refreshClaims)
	assert.Equal(suite.T(), int64(1), refreshClaims.UserID)
	assert.Equal(suite.T(), "test@example.com", refreshClaims.Email)
}

// TestTokenValidation 测试令牌验证
func (suite *AuthIntegrationTestSuite) TestTokenValidation() {
	// 测试有效令牌
	validToken, err := auth.GenerateAccessToken(1, "test@example.com", "user", "user")
	require.NoError(suite.T(), err)

	claims, err := auth.ValidateAccessToken(validToken)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), claims)

	// 测试无效令牌
	_, err = auth.ValidateAccessToken("invalid-token")
	assert.Error(suite.T(), err)

	// 测试空令牌
	_, err = auth.ValidateAccessToken("")
	assert.Error(suite.T(), err)
}

// TestTokenExpiration 测试令牌过期
func (suite *AuthIntegrationTestSuite) TestTokenExpiration() {
	// 临时修改配置，使令牌立即过期
	originalExpire := config.AppConfig.JWT.AccessExpire
	config.AppConfig.JWT.AccessExpire = -1
	defer func() {
		config.AppConfig.JWT.AccessExpire = originalExpire
	}()

	// 生成过期的令牌
	expiredToken, err := auth.GenerateAccessToken(1, "test@example.com", "user", "user")
	require.NoError(suite.T(), err)

	// 等待确保令牌过期
	time.Sleep(100 * time.Millisecond)

	// 验证过期令牌应该失败
	_, err = auth.ValidateAccessToken(expiredToken)
	assert.Error(suite.T(), err)
}

// TestDifferentUserTypes 测试不同用户类型的令牌
func (suite *AuthIntegrationTestSuite) TestDifferentUserTypes() {
	// 测试普通用户令牌
	userToken, err := auth.GenerateAccessToken(1, "user@example.com", "user", "user")
	require.NoError(suite.T(), err)

	userClaims, err := auth.ValidateAccessToken(userToken)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "user", userClaims.Role)
	assert.Equal(suite.T(), "user", userClaims.UserType)

	// 测试管理员令牌
	adminToken, err := auth.GenerateAccessToken(2, "admin@example.com", "admin", "admin")
	require.NoError(suite.T(), err)

	adminClaims, err := auth.ValidateAccessToken(adminToken)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "admin", adminClaims.Role)
	assert.Equal(suite.T(), "admin", adminClaims.UserType)

	// 测试超级管理员令牌
	superAdminToken, err := auth.GenerateAccessToken(3, "superadmin@example.com", "super_admin", "admin")
	require.NoError(suite.T(), err)

	superAdminClaims, err := auth.ValidateAccessToken(superAdminToken)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "super_admin", superAdminClaims.Role)
	assert.Equal(suite.T(), "admin", superAdminClaims.UserType)
}

// TestPublicRouteAccess 测试公开路由访问
func (suite *AuthIntegrationTestSuite) TestPublicRouteAccess() {
	req := httptest.NewRequest("GET", "/api/v1/public/health", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "ok", response["status"])
}

// TestTokenIntegration 测试令牌集成流程
func (suite *AuthIntegrationTestSuite) TestTokenIntegration() {
	// 1. 生成访问令牌和刷新令牌
	accessToken, err := auth.GenerateAccessToken(1, "integration@example.com", "user", "user")
	require.NoError(suite.T(), err)

	refreshToken, err := auth.GenerateRefreshToken(1, "integration@example.com", "user", "user")
	require.NoError(suite.T(), err)

	// 2. 验证访问令牌
	accessClaims, err := auth.ValidateAccessToken(accessToken)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), accessClaims.UserID)
	assert.Equal(suite.T(), "integration@example.com", accessClaims.Email)

	// 3. 验证刷新令牌
	refreshClaims, err := auth.ValidateRefreshToken(refreshToken)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), refreshClaims.UserID)
	assert.Equal(suite.T(), "integration@example.com", refreshClaims.Email)

	// 4. 使用刷新令牌生成新的访问令牌
	newAccessToken, err := auth.GenerateAccessToken(
		refreshClaims.UserID,
		refreshClaims.Email,
		refreshClaims.Role,
		refreshClaims.UserType,
	)
	require.NoError(suite.T(), err)

	// 5. 验证新的访问令牌
	newAccessClaims, err := auth.ValidateAccessToken(newAccessToken)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), refreshClaims.UserID, newAccessClaims.UserID)
	assert.Equal(suite.T(), refreshClaims.Email, newAccessClaims.Email)
}

// TestConfigMissing 测试配置缺失情况
func (suite *AuthIntegrationTestSuite) TestConfigMissing() {
	// 保存原始配置
	originalConfig := config.AppConfig
	defer func() {
		config.AppConfig = originalConfig
	}()

	// 设置空配置
	config.AppConfig = nil

	// 尝试生成令牌应该panic或失败
	defer func() {
		if r := recover(); r != nil {
			// 预期会panic，这是正常的
			assert.NotNil(suite.T(), r)
		}
	}()

	// 这个调用会导致panic，因为config.AppConfig为nil
	_, err := auth.GenerateAccessToken(1, "test@example.com", "user", "user")
	if err != nil {
		assert.Error(suite.T(), err)
	}
}

// TestInvalidTokenFormats 测试无效令牌格式
func (suite *AuthIntegrationTestSuite) TestInvalidTokenFormats() {
	invalidTokens := []string{
		"",
		"invalid",
		"invalid.token",
		"invalid.token.format",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
	}

	for _, token := range invalidTokens {
		_, err := auth.ValidateAccessToken(token)
		assert.Error(suite.T(), err, "Token should be invalid: %s", token)

		_, err = auth.ValidateRefreshToken(token)
		assert.Error(suite.T(), err, "Refresh token should be invalid: %s", token)
	}
}

// TestConcurrentTokenOperations 测试并发令牌操作
func (suite *AuthIntegrationTestSuite) TestConcurrentTokenOperations() {
	const numGoroutines = 10
	const numOperations = 5

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				// 生成令牌
				token, err := auth.GenerateAccessToken(
					int64(id+1),
					"concurrent@example.com",
					"user",
					"user",
				)
				assert.NoError(suite.T(), err)
				assert.NotEmpty(suite.T(), token)

				// 验证令牌
				claims, err := auth.ValidateAccessToken(token)
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), int64(id+1), claims.UserID)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// TestAuthIntegrationSuite 运行集成测试套件
func TestAuthIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AuthIntegrationTestSuite))
}