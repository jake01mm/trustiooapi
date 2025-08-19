package entities

import "time"

// User 用户实体
type User struct {
	ID               int64      `json:"id" db:"id"`
	Name             string     `json:"name" db:"name"`
	Email            string     `json:"email" db:"email"`
	Password         string     `json:"-" db:"password"`
	Phone            *string    `json:"phone,omitempty" db:"phone"`
	ImageKey         string     `json:"image_key" db:"image_key"`
	Role             string     `json:"role" db:"role"`
	Status           string     `json:"status" db:"status"`
	EmailVerified    bool       `json:"email_verified" db:"email_verified"`
	PhoneVerified    bool       `json:"phone_verified" db:"phone_verified"`
	AutoRegistered   bool       `json:"auto_registered" db:"auto_registered"`
	ProfileCompleted bool       `json:"profile_completed" db:"profile_completed"`
	PasswordSet      bool       `json:"password_set" db:"password_set"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}