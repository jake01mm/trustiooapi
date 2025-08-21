package entities

import "time"

// LoginCode 登录验证码实体
type LoginCode struct {
	ID        int64     `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Code      string    `json:"code" db:"code"`
	IsUsed    bool      `json:"is_used" db:"is_used"`
	ExpiredAt time.Time `json:"expired_at" db:"expired_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}