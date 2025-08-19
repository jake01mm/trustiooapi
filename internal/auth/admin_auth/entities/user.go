package entities

import "time"

// UserInfo 用户信息实体（在管理员模块中用于用户管理）
type UserInfo struct {
	ID               int64      `json:"id" db:"id"`
	Name             string     `json:"name" db:"name"`
	Email            string     `json:"email" db:"email"`
	Phone            *string    `json:"phone,omitempty" db:"phone"`
	ImageKey         string     `json:"image_key" db:"image_key"`
	Status           string     `json:"status" db:"status"`
	EmailVerified    bool       `json:"email_verified" db:"email_verified"`
	PhoneVerified    bool       `json:"phone_verified" db:"phone_verified"`
	AutoRegistered   bool       `json:"auto_registered" db:"auto_registered"`
	ProfileCompleted bool       `json:"profile_completed" db:"profile_completed"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}