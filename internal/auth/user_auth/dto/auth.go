package dto

import "trusioo_api/internal/auth/user_auth/entities"

// RegisterRequest 注册请求 - 简化为email+password
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求 - 第一步：验证email+password并发送验证码
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// VerificationType 返回用户登录的验证类型
func (r *LoginRequest) VerificationType() string {
	return "user_login"
}

// LoginVerifyRequest 登录验证请求 - 第二步：验证登录验证码
// 注意：此请求在Service层会自动添加 type: "user_login" 字段调用verification服务
type LoginVerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

// VerificationType 返回此DTO对应的验证类型
func (r *LoginVerifyRequest) VerificationType() string {
	return "user_login"
}

// LoginCodeResponse 登录验证码响应
type LoginCodeResponse struct {
	Message   string `json:"message"`
	LoginCode string `json:"login_code"`
	ExpiresIn int    `json:"expires_in"` // 秒
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	ExpiresIn    int64             `json:"expires_in"`
	TokenType    string            `json:"token_type"`
	User         *entities.User    `json:"user"`
	LoginSession *LoginSessionInfo `json:"login_session,omitempty"`
}

// LoginSessionInfo 登录会话信息
type LoginSessionInfo struct {
	IP           string `json:"ip"`
	Country      string `json:"country"`
	City         string `json:"city"`
	Region       string `json:"region"`
	Timezone     string `json:"timezone"`
	Organization string `json:"organization"`
	Location     string `json:"location"`
	IsTrusted    bool   `json:"is_trusted"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User *entities.User `json:"user"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPasswordResponse 忘记密码响应
type ForgotPasswordResponse struct {
	Message   string `json:"message"`
	ExpiresIn int    `json:"expires_in"` // 秒
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required,len=6"`
	Password string `json:"password" binding:"required,min=6"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Message string `json:"message"`
}
