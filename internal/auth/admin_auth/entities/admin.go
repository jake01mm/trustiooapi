package entities

import "time"

// Admin 管理员实体
type Admin struct {
	ID               int64      `json:"id" db:"id"`
	Name             string     `json:"name" db:"name"`
	Email            string     `json:"email" db:"email"`
	Password         string     `json:"-" db:"password"`
	Phone            *string    `json:"phone,omitempty" db:"phone"`
	ImageKey         string     `json:"image_key" db:"image_key"`
	Role             string     `json:"role" db:"role"`
	IsSuper          bool       `json:"is_super" db:"is_super"`
	Status           string     `json:"status" db:"status"`
	EmailVerified    bool       `json:"email_verified" db:"email_verified"`
	PhoneVerified    bool       `json:"phone_verified" db:"phone_verified"`
	ProfileCompleted bool       `json:"profile_completed" db:"profile_completed"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}