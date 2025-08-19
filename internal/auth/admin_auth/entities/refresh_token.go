package entities

import "time"

// AdminRefreshToken 管理员刷新令牌实体
type AdminRefreshToken struct {
	ID         int64     `json:"id" db:"id"`
	AdminID    int64     `json:"admin_id" db:"admin_id"`
	Token      string    `json:"token" db:"token"`
	IsValid    bool      `json:"is_valid" db:"is_valid"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	DeviceInfo string    `json:"device_info" db:"device_info"`
}