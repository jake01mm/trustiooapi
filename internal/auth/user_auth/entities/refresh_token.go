package entities

import "time"

// RefreshToken 刷新令牌实体
type RefreshToken struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	IsValid    bool      `json:"is_valid" db:"is_valid"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	DeviceInfo string    `json:"device_info" db:"device_info"`
}