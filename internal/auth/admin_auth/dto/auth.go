package dto

import "trusioo_api/internal/auth/admin_auth/entities"

// AdminLoginRequest 管理员登录请求 - 第一步：验证email+password并发送验证码
type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AdminLoginVerifyRequest 管理员登录验证请求 - 第二步：验证登录验证码
type AdminLoginVerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

// AdminLoginCodeResponse 管理员登录验证码响应
type AdminLoginCodeResponse struct {
	Message   string `json:"message"`
	LoginCode string `json:"login_code"`
	ExpiresIn int    `json:"expires_in"` // 秒
}

// AdminLoginResponse 管理员登录响应
type AdminLoginResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresIn    int64          `json:"expires_in"`
	TokenType    string         `json:"token_type"`
	Admin        entities.Admin `json:"admin"`
	LoginSession *AdminLoginSessionInfo `json:"login_session,omitempty"`
}

// AdminLoginSessionInfo 管理员登录会话信息
type AdminLoginSessionInfo struct {
	IP           string `json:"ip"`
	Country      string `json:"country"`
	City         string `json:"city"`
	Region       string `json:"region"`
	Timezone     string `json:"timezone"`
	Organization string `json:"organization"`
	Location     string `json:"location"`
	IsTrusted    bool   `json:"is_trusted"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AdminForgotPasswordRequest 管理员忘记密码请求
type AdminForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// AdminForgotPasswordResponse 管理员忘记密码响应
type AdminForgotPasswordResponse struct {
	Message   string `json:"message"`
	ExpiresIn int    `json:"expires_in"` // 秒
}

// AdminResetPasswordRequest 管理员重置密码请求
type AdminResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required,len=6"`
	Password string `json:"password" binding:"required,min=6"`
}

// AdminResetPasswordResponse 管理员重置密码响应
type AdminResetPasswordResponse struct {
	Message string `json:"message"`
}
