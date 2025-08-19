package dto

import "trusioo_api/internal/auth/user_auth/entities"

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required,min=6"`
	VerificationCode string `json:"verification_code" binding:"required,len=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required"`
	VerificationCode string `json:"verification_code" binding:"required,len=6"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresIn    int64          `json:"expires_in"`
	TokenType    string         `json:"token_type"`
	User         *entities.User `json:"user"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User *entities.User `json:"user"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// PreAuthRequest 预验证请求（验证email+password后才能获取验证码）
type PreAuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// PreAuthResponse 预验证响应
type PreAuthResponse struct {
	Message   string `json:"message"`
	Verified  bool   `json:"verified"`
	Email     string `json:"email,omitempty"`
	ExpiredAt string `json:"expired_at,omitempty"`
}
