package dto

import (
	"time"
	"trusioo_api/pkg/carddetection"
)

// CardDetectionHistoryRequest 查询检测历史请求
type CardDetectionHistoryRequest struct {
	Page        int                       `form:"page" binding:"min=1" json:"page"`
	PageSize    int                       `form:"page_size" binding:"min=1,max=100" json:"page_size"`
	Status      string                    `form:"status" json:"status"` // pending, completed, failed
	ProductMark carddetection.ProductMark `form:"product_mark" json:"product_mark"`
	CardNumber  string                    `form:"card_number" json:"card_number"`
	StartDate   string                    `form:"start_date" json:"start_date"` // YYYY-MM-DD
	EndDate     string                    `form:"end_date" json:"end_date"`     // YYYY-MM-DD
}

// CardDetectionRecordResponse 检测记录响应
type CardDetectionRecordResponse struct {
	ID           int64                     `json:"id"`
	RequestID    string                    `json:"request_id"`
	CardNumber   string                    `json:"card_number"`
	PinCode      *string                   `json:"pin_code,omitempty"`
	ProductMark  carddetection.ProductMark `json:"product_mark"`
	RegionID     *int                      `json:"region_id,omitempty"`
	RegionName   *string                   `json:"region_name,omitempty"`
	AutoType     *int                      `json:"auto_type,omitempty"`
	CheckStatus  string                    `json:"check_status"`
	CheckResult  interface{}               `json:"check_result,omitempty"`
	ErrorMessage *string                   `json:"error_message,omitempty"`
	ResponseCode *int                      `json:"response_code,omitempty"`
	ResponseTime *int                      `json:"response_time,omitempty"`
	CheckedAt    *time.Time                `json:"checked_at,omitempty"`
	CreatedAt    time.Time                 `json:"created_at"`
	UpdatedAt    time.Time                 `json:"updated_at"`
}

// CardDetectionHistoryResponse 检测历史响应
type CardDetectionHistoryResponse struct {
	Records    []*CardDetectionRecordResponse `json:"records"`
	Pagination PaginationResponse             `json:"pagination"`
	Summary    *CardDetectionSummaryResponse  `json:"summary"`
}

// CardDetectionSummaryResponse 检测汇总响应
type CardDetectionSummaryResponse struct {
	TotalChecks   int        `json:"total_checks"`
	SuccessChecks int        `json:"success_checks"`
	FailedChecks  int        `json:"failed_checks"`
	PendingChecks int        `json:"pending_checks"`
	SuccessRate   float64    `json:"success_rate"`
	LastCheckAt   *time.Time `json:"last_check_at,omitempty"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Page        int `json:"page"`
	PageSize    int `json:"page_size"`
	Total       int `json:"total"`
	TotalPages  int `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

// CardDetectionDetailRequest 查询单个检测记录详情请求
type CardDetectionDetailRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// CardDetectionStatsResponse 用户检测统计响应
type CardDetectionStatsResponse struct {
	Summary      *CardDetectionSummaryResponse       `json:"summary"`
	ProductStats []*ProductStatsResponse            `json:"product_stats"`
	MonthlyStats []*MonthlyStatsResponse            `json:"monthly_stats"`
	RecentChecks []*CardDetectionRecordResponse     `json:"recent_checks"`
}

// ProductStatsResponse 产品统计响应
type ProductStatsResponse struct {
	ProductMark   carddetection.ProductMark `json:"product_mark"`
	TotalChecks   int                      `json:"total_checks"`
	SuccessChecks int                      `json:"success_checks"`
	FailedChecks  int                      `json:"failed_checks"`
	SuccessRate   float64                  `json:"success_rate"`
	LastCheckAt   *time.Time               `json:"last_check_at,omitempty"`
}

// MonthlyStatsResponse 月度统计响应
type MonthlyStatsResponse struct {
	Month         string `json:"month"` // YYYY-MM
	TotalChecks   int    `json:"total_checks"`
	SuccessChecks int    `json:"success_checks"`
	FailedChecks  int    `json:"failed_checks"`
	SuccessRate   float64 `json:"success_rate"`
}

// CDProduct CD产品信息
type CDProduct struct {
	ID                int     `json:"id"`
	ProductMark       string  `json:"product_mark"`
	ProductName       string  `json:"product_name"`
	RequiresRegion    bool    `json:"requires_region"`
	RequiresPin       bool    `json:"requires_pin"`
	CardFormat        string  `json:"card_format"`
	CardLengthMin     int     `json:"card_length_min"`
	CardLengthMax     int     `json:"card_length_max"`
	PinLength         *int    `json:"pin_length,omitempty"`
	ValidationPattern *string `json:"validation_pattern,omitempty"`
	SupportsAutoType  bool    `json:"supports_auto_type"`
	Status            string  `json:"status"`
}

// CDProductsResponse CD产品列表响应
type CDProductsResponse struct {
	Products []*CDProduct `json:"products"`
	Total    int          `json:"total"`
}

// CDRegion CD区域信息
type CDRegion struct {
	ID           int    `json:"id"`
	ProductMark  string `json:"product_mark"`
	RegionID     string `json:"region_id"`
	RegionName   string `json:"region_name"`
	RegionNameEn string `json:"region_name_en"`
	Status       string `json:"status"`
	SortOrder    int    `json:"sort_order"`
}

// CDRegionsResponse CD区域列表响应
type CDRegionsResponse struct {
	Regions     []*CDRegion `json:"regions"`
	Total       int         `json:"total"`
	ProductMark string      `json:"product_mark,omitempty"`
}