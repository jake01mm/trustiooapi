package carddetection

import "fmt"

// Error codes
const (
	ErrCodeInvalidConfig     = 1001
	ErrCodeInvalidRequest    = 1002
	ErrCodeEncryptionFailed  = 1003
	ErrCodeDecryptionFailed  = 1004
	ErrCodeSignatureFailed   = 1005
	ErrCodeAPIRequest        = 1006
	ErrCodeAPIResponse       = 1007
	ErrCodeTimeout           = 1008
	ErrCodeUnsupportedRegion = 1009
	ErrCodeInvalidCardFormat = 1010
)

// CardDetectionError 卡片检测错误
type CardDetectionError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

// Error 实现error接口
func (e *CardDetectionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("CardDetection Error %d: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("CardDetection Error %d: %s", e.Code, e.Message)
}

// Unwrap 支持errors.Unwrap
func (e *CardDetectionError) Unwrap() error {
	return e.Cause
}

// NewError 创建新的错误
func NewError(code int, message string, cause error) *CardDetectionError {
	return &CardDetectionError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// 预定义错误
var (
	// 配置错误
	ErrInvalidConfig = &CardDetectionError{
		Code:    ErrCodeInvalidConfig,
		Message: "invalid configuration",
	}
	
	ErrMissingHost = &CardDetectionError{
		Code:    ErrCodeInvalidConfig,
		Message: "missing host configuration",
	}
	
	ErrMissingAppID = &CardDetectionError{
		Code:    ErrCodeInvalidConfig,
		Message: "missing app ID configuration",
	}
	
	ErrMissingAppSecret = &CardDetectionError{
		Code:    ErrCodeInvalidConfig,
		Message: "missing app secret configuration",
	}
	
	// 请求错误
	ErrInvalidRequest = &CardDetectionError{
		Code:    ErrCodeInvalidRequest,
		Message: "invalid request parameters",
	}
	
	ErrInvalidProductMark = &CardDetectionError{
		Code:    ErrCodeInvalidRequest,
		Message: "invalid product mark",
	}
	
	ErrInvalidCardFormat = &CardDetectionError{
		Code:    ErrCodeInvalidCardFormat,
		Message: "invalid card format",
	}
	
	ErrUnsupportedRegion = &CardDetectionError{
		Code:    ErrCodeUnsupportedRegion,
		Message: "unsupported region for this product",
	}
	
	// 加密错误
	ErrEncryptionFailed = &CardDetectionError{
		Code:    ErrCodeEncryptionFailed,
		Message: "encryption failed",
	}
	
	ErrDecryptionFailed = &CardDetectionError{
		Code:    ErrCodeDecryptionFailed,
		Message: "decryption failed",
	}
	
	ErrSignatureFailed = &CardDetectionError{
		Code:    ErrCodeSignatureFailed,
		Message: "signature generation/verification failed",
	}
	
	// API错误
	ErrAPIRequest = &CardDetectionError{
		Code:    ErrCodeAPIRequest,
		Message: "API request failed",
	}
	
	ErrAPIResponse = &CardDetectionError{
		Code:    ErrCodeAPIResponse,
		Message: "invalid API response",
	}
	
	ErrTimeout = &CardDetectionError{
		Code:    ErrCodeTimeout,
		Message: "request timeout",
	}
)

// WrapError 包装错误
func WrapError(err error, code int, message string) *CardDetectionError {
	return &CardDetectionError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// IsCardDetectionError 检查是否是卡片检测错误
func IsCardDetectionError(err error) bool {
	_, ok := err.(*CardDetectionError)
	return ok
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) int {
	if cdErr, ok := err.(*CardDetectionError); ok {
		return cdErr.Code
	}
	return 0
}