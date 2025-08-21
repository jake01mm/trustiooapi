package entities

import (
	"time"
	"trusioo_api/pkg/carddetection"
)

// CardDetectionRecord 卡片检测记录实体
type CardDetectionRecord struct {
	ID            int64                      `json:"id" db:"id"`
	UserID        int64                      `json:"user_id" db:"user_id"`
	RequestID     string                     `json:"request_id" db:"request_id"`
	CardNumber    string                     `json:"card_number" db:"card_number"`
	PinCode       *string                    `json:"pin_code,omitempty" db:"pin_code"`
	ProductMark   carddetection.ProductMark  `json:"product_mark" db:"product_mark"`
	RegionID      *int                       `json:"region_id,omitempty" db:"region_id"`
	RegionName    *string                    `json:"region_name,omitempty" db:"region_name"`
	AutoType      *int                       `json:"auto_type,omitempty" db:"auto_type"`
	CheckStatus   string                     `json:"check_status" db:"check_status"` // pending, completed, failed
	CheckResult   *string                    `json:"check_result,omitempty" db:"check_result"` // JSON格式的检测结果
	ErrorMessage  *string                    `json:"error_message,omitempty" db:"error_message"`
	ResponseCode  *int                       `json:"response_code,omitempty" db:"response_code"`
	ResponseTime  *int                       `json:"response_time,omitempty" db:"response_time"` // 响应时间(毫秒)
	CheckedAt     *time.Time                 `json:"checked_at,omitempty" db:"checked_at"`
	CreatedAt     time.Time                  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at" db:"updated_at"`
}

// CardDetectionSummary 卡片检测汇总统计
type CardDetectionSummary struct {
	UserID        int64  `json:"user_id" db:"user_id"`
	TotalChecks   int    `json:"total_checks" db:"total_checks"`
	SuccessChecks int    `json:"success_checks" db:"success_checks"`
	FailedChecks  int    `json:"failed_checks" db:"failed_checks"`
	LastCheckAt   *time.Time `json:"last_check_at,omitempty" db:"last_check_at"`
}

// TableName 返回表名
func (CardDetectionRecord) TableName() string {
	return "card_detection_records"
}

// CDProduct CD产品实体
type CDProduct struct {
	ID                int       `json:"id" db:"id"`
	ProductMark       string    `json:"product_mark" db:"product_mark"`
	ProductName       string    `json:"product_name" db:"product_name"`
	RequiresRegion    bool      `json:"requires_region" db:"requires_region"`
	RequiresPin       bool      `json:"requires_pin" db:"requires_pin"`
	CardFormat        string    `json:"card_format" db:"card_format"`
	CardLengthMin     int       `json:"card_length_min" db:"card_length_min"`
	CardLengthMax     int       `json:"card_length_max" db:"card_length_max"`
	PinLength         int       `json:"pin_length" db:"pin_length"`
	ValidationPattern *string   `json:"validation_pattern,omitempty" db:"validation_pattern"`
	SupportsAutoType  bool      `json:"supports_auto_type" db:"supports_auto_type"`
	Status            string    `json:"status" db:"status"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// TableName 返回表名
func (CDProduct) TableName() string {
	return "cd_products"
}

// CDRegion CD区域实体
type CDRegion struct {
	ID           int       `json:"id" db:"id"`
	ProductMark  string    `json:"product_mark" db:"product_mark"`
	RegionID     string    `json:"region_id" db:"region_id"`
	RegionName   string    `json:"region_name" db:"region_name"`
	RegionNameEn string    `json:"region_name_en" db:"region_name_en"`
	Status       string    `json:"status" db:"status"`
	SortOrder    int       `json:"sort_order" db:"sort_order"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// TableName 返回表名
func (CDRegion) TableName() string {
	return "cd_regions"
}