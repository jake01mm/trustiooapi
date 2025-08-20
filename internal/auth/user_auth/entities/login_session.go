package entities

import "time"

// LoginSession 登录会话记录实体
type LoginSession struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	IP           string    `json:"ip" db:"ip"`
	Country      string    `json:"country" db:"country"`
	City         string    `json:"city" db:"city"`
	Region       string    `json:"region" db:"region"`
	Timezone     string    `json:"timezone" db:"timezone"`
	Organization string    `json:"organization" db:"organization"`
	Location     string    `json:"location" db:"location"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	DeviceType   string    `json:"device_type" db:"device_type"`
	OS           string    `json:"os" db:"os"`
	Browser      string    `json:"browser" db:"browser"`
	IsTrusted    bool      `json:"is_trusted" db:"is_trusted"`
	LoginMethod  string    `json:"login_method" db:"login_method"`
	Platform     string    `json:"platform" db:"platform"`
	Status       string    `json:"status" db:"status"`
	Reason       string    `json:"reason" db:"reason"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}