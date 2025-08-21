package dto

// SendVerificationRequest 发送验证码请求
type SendVerificationRequest struct {
	Target string `json:"target" binding:"required,email" validate:"required,email"`
	Type   string `json:"type" binding:"required" validate:"required,oneof=register reset_password forgot_password"`
}

// VerifyCodeRequest 验证验证码请求
type VerifyCodeRequest struct {
	Target string `json:"target" binding:"required,email" validate:"required,email"`
	Type   string `json:"type" binding:"required" validate:"required,oneof=register reset_password forgot_password"`
	Code   string `json:"code" binding:"required,len=6" validate:"required,len=6"`
}

// SendVerificationResponse 发送验证码响应
type SendVerificationResponse struct {
	Message   string `json:"message"`
	ExpiredAt string `json:"expired_at"`
	Code      string `json:"code,omitempty"` // 仅用于测试环境
}

// VerifyCodeResponse 验证验证码响应
type VerifyCodeResponse struct {
	Message string `json:"message"`
	Valid   bool   `json:"valid"`
}