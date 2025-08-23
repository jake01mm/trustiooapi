package auth

import (
	"testing"
	"time"

	"trusioo_api/config"

	"github.com/golang-jwt/jwt/v5"
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

func TestGenerateAccessToken(t *testing.T) {
	setupTestConfig()

	tests := []struct {
		name     string
		userID   int64
		email    string
		role     string
		userType string
		wantErr  bool
	}{
		{
			name:     "成功生成用户访问令牌",
			userID:   1,
			email:    "user@example.com",
			role:     "user",
			userType: "user",
			wantErr:  false,
		},
		{
			name:     "成功生成管理员访问令牌",
			userID:   2,
			email:    "admin@example.com",
			role:     "admin",
			userType: "admin",
			wantErr:  false,
		},
		{
			name:     "成功生成超级管理员访问令牌",
			userID:   3,
			email:    "superadmin@example.com",
			role:     "super_admin",
			userType: "admin",
			wantErr:  false,
		},
		{
			name:     "空邮箱",
			userID:   1,
			email:    "",
			role:     "user",
			userType: "user",
			wantErr:  false, // 应该仍然能生成 token
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateAccessToken(tt.userID, tt.email, tt.role, tt.userType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// 验证生成的 token 是否有效
				claims, err := ValidateAccessToken(token)
				require.NoError(t, err)
				assert.Equal(t, tt.userID, claims.UserID)
				assert.Equal(t, tt.email, claims.Email)
				assert.Equal(t, tt.role, claims.Role)
				assert.Equal(t, tt.userType, claims.UserType)
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	setupTestConfig()

	tests := []struct {
		name     string
		userID   int64
		email    string
		role     string
		userType string
		wantErr  bool
	}{
		{
			name:     "成功生成用户刷新令牌",
			userID:   1,
			email:    "user@example.com",
			role:     "user",
			userType: "user",
			wantErr:  false,
		},
		{
			name:     "成功生成管理员刷新令牌",
			userID:   2,
			email:    "admin@example.com",
			role:     "admin",
			userType: "admin",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateRefreshToken(tt.userID, tt.email, tt.role, tt.userType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// 验证生成的 token 是否有效
				claims, err := ValidateRefreshToken(token)
				require.NoError(t, err)
				assert.Equal(t, tt.userID, claims.UserID)
				assert.Equal(t, tt.email, claims.Email)
				assert.Equal(t, tt.role, claims.Role)
				assert.Equal(t, tt.userType, claims.UserType)
			}
		})
	}
}

func TestValidateAccessToken(t *testing.T) {
	setupTestConfig()

	tests := []struct {
		name      string
		setupFunc func() string
		wantErr   bool
		errorMsg  string
	}{
		{
			name: "有效的访问令牌",
			setupFunc: func() string {
				token, _ := GenerateAccessToken(1, "user@example.com", "user", "user")
				return token
			},
			wantErr: false,
		},
		{
			name: "空令牌",
			setupFunc: func() string {
				return ""
			},
			wantErr:  true,
			errorMsg: "token contains an invalid number of segments",
		},
		{
			name: "无效格式的令牌",
			setupFunc: func() string {
				return "invalid.token.format"
			},
			wantErr:  true,
			errorMsg: "invalid character",
		},
		{
			name: "错误签名的令牌",
			setupFunc: func() string {
				// 使用错误的密钥生成 token
				originalSecret := config.AppConfig.JWT.Secret
				config.AppConfig.JWT.Secret = "wrong-secret"
				token, _ := GenerateAccessToken(1, "user@example.com", "user", "user")
				config.AppConfig.JWT.Secret = originalSecret
				return token
			},
			wantErr:  true,
			errorMsg: "signature is invalid",
		},
		{
			name: "过期的令牌",
			setupFunc: func() string {
				// 临时设置过期时间为负数
				originalExpire := config.AppConfig.JWT.AccessExpire
				config.AppConfig.JWT.AccessExpire = -1
				token, _ := GenerateAccessToken(1, "user@example.com", "user", "user")
				config.AppConfig.JWT.AccessExpire = originalExpire
				return token
			},
			wantErr:  true,
			errorMsg: "token is expired",
		},
		{
			name: "刷新令牌用于访问验证",
			setupFunc: func() string {
				token, _ := GenerateRefreshToken(1, "user@example.com", "user", "user")
				return token
			},
			wantErr:  true,
			errorMsg: "signature is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupFunc()
			claims, err := ValidateAccessToken(token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, int64(1), claims.UserID)
				assert.Equal(t, "user@example.com", claims.Email)
				assert.Equal(t, "user", claims.Role)
				assert.Equal(t, "user", claims.UserType)
			}
		})
	}
}

func TestValidateRefreshToken(t *testing.T) {
	setupTestConfig()

	tests := []struct {
		name      string
		setupFunc func() string
		wantErr   bool
		errorMsg  string
	}{
		{
			name: "有效的刷新令牌",
			setupFunc: func() string {
				token, _ := GenerateRefreshToken(1, "user@example.com", "user", "user")
				return token
			},
			wantErr: false,
		},
		{
			name: "空令牌",
			setupFunc: func() string {
				return ""
			},
			wantErr:  true,
			errorMsg: "token contains an invalid number of segments",
		},
		{
			name: "无效格式的令牌",
			setupFunc: func() string {
				return "invalid.token.format"
			},
			wantErr:  true,
			errorMsg: "invalid character",
		},
		{
			name: "错误签名的令牌",
			setupFunc: func() string {
				// 使用错误的密钥生成 token
				originalSecret := config.AppConfig.JWT.RefreshSecret
				config.AppConfig.JWT.RefreshSecret = "wrong-secret"
				token, _ := GenerateRefreshToken(1, "user@example.com", "user", "user")
				config.AppConfig.JWT.RefreshSecret = originalSecret
				return token
			},
			wantErr:  true,
			errorMsg: "signature is invalid",
		},
		{
			name: "过期的令牌",
			setupFunc: func() string {
				// 临时设置过期时间为负数
				originalExpire := config.AppConfig.JWT.RefreshExpire
				config.AppConfig.JWT.RefreshExpire = -1
				token, _ := GenerateRefreshToken(1, "user@example.com", "user", "user")
				config.AppConfig.JWT.RefreshExpire = originalExpire
				return token
			},
			wantErr:  true,
			errorMsg: "token is expired",
		},
		{
			name: "访问令牌用于刷新验证",
			setupFunc: func() string {
				token, _ := GenerateAccessToken(1, "user@example.com", "user", "user")
				return token
			},
			wantErr:  true,
			errorMsg: "signature is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupFunc()
			claims, err := ValidateRefreshToken(token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, int64(1), claims.UserID)
				assert.Equal(t, "user@example.com", claims.Email)
				assert.Equal(t, "user", claims.Role)
				assert.Equal(t, "user", claims.UserType)
			}
		})
	}
}

// TestTokenClaims 测试 Claims 结构体
func TestTokenClaims(t *testing.T) {
	setupTestConfig()

	// 测试 Claims 结构体的字段
	claims := &Claims{
		UserID:   123,
		Email:    "test@example.com",
		Role:     "admin",
		UserType: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	assert.Equal(t, int64(123), claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, "admin", claims.UserType)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
}

// TestTokenIntegration 集成测试：生成和验证完整流程
func TestTokenIntegration(t *testing.T) {
	setupTestConfig()

	userID := int64(42)
	email := "integration@example.com"
	role := "user"
	userType := "user"

	// 生成访问令牌
	accessToken, err := GenerateAccessToken(userID, email, role, userType)
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)

	// 生成刷新令牌
	refreshToken, err := GenerateRefreshToken(userID, email, role, userType)
	require.NoError(t, err)
	require.NotEmpty(t, refreshToken)

	// 验证访问令牌
	accessClaims, err := ValidateAccessToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, accessClaims.UserID)
	assert.Equal(t, email, accessClaims.Email)
	assert.Equal(t, role, accessClaims.Role)
	assert.Equal(t, userType, accessClaims.UserType)

	// 验证刷新令牌
	refreshClaims, err := ValidateRefreshToken(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
	assert.Equal(t, email, refreshClaims.Email)
	assert.Equal(t, role, refreshClaims.Role)
	assert.Equal(t, userType, refreshClaims.UserType)

	// 交叉验证应该失败（使用错误的密钥）
	_, err = ValidateRefreshToken(accessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature is invalid")

	_, err = ValidateAccessToken(refreshToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature is invalid")
}

// TestConfigMissing 测试配置缺失的情况
func TestConfigMissing(t *testing.T) {
	// 保存原始配置
	originalConfig := config.AppConfig
	defer func() {
		config.AppConfig = originalConfig
	}()

	// 设置为 nil 配置
	config.AppConfig = nil

	// 应该 panic 或返回错误
	assert.Panics(t, func() {
		GenerateAccessToken(1, "test@example.com", "user", "user")
	})

	assert.Panics(t, func() {
		GenerateRefreshToken(1, "test@example.com", "user", "user")
	})
}