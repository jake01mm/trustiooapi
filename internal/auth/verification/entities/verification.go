package entities

import "time"

// Verification 验证码实体
type Verification struct {
	ID        int64     `json:"id" db:"id"`
	UserID    *int64    `json:"user_id" db:"user_id"`
	Target    string    `json:"target" db:"target"`
	Type      string    `json:"type" db:"type"`
	Action    string    `json:"action" db:"action"`
	SentAt    time.Time `json:"sent_at" db:"sent_at"`
	Code      string    `json:"code" db:"code"`
	IsUsed    bool      `json:"is_used" db:"is_used"`
	ExpiredAt time.Time `json:"expired_at" db:"expired_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// VerificationType 验证码类型常量
const (
	TypeRegister       = "register"
	TypeResetPassword  = "reset_password"
	TypeForgotPassword = "forgot_password"
)

// VerificationAction 验证码动作常量
const (
	ActionEmailVerification = "email_verification"
	ActionPhoneVerification = "phone_verification"
)
