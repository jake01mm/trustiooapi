package dto

import "trusioo_api/internal/auth/admin_auth/entities"

// UserListRequest 获取用户列表请求
type UserListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive all"`
	Email    string `form:"email" binding:"omitempty,email"`
	Phone    string `form:"phone" binding:"omitempty"`
}

// UserStats 用户统计
type UserStats struct {
	TotalUsers         int64 `json:"total_users"`
	ActiveUsers        int64 `json:"active_users"`
	InactiveUsers      int64 `json:"inactive_users"`
	RegisteredToday    int64 `json:"registered_today"`
	RegisteredThisWeek int64 `json:"registered_this_week"`
	RegisteredThisMonth int64 `json:"registered_this_month"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
	Users []entities.UserInfo  `json:"users"`
}