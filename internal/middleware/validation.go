package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"trusioo_api/internal/common"

	"github.com/gin-gonic/gin"
)

// ValidationRule 验证规则
type ValidationRule struct {
	Field     string
	Required  bool
	MinLength int
	MaxLength int
	Pattern   *regexp.Regexp
	Custom    func(interface{}) error
}

// Validator 验证器
type Validator struct {
	rules []ValidationRule
}

// NewValidator 创建新验证器
func NewValidator() *Validator {
	return &Validator{
		rules: make([]ValidationRule, 0),
	}
}

// AddRule 添加验证规则
func (v *Validator) AddRule(rule ValidationRule) *Validator {
	v.rules = append(v.rules, rule)
	return v
}

// Validate 执行验证
func (v *Validator) Validate(data map[string]interface{}) []string {
	var errors []string

	for _, rule := range v.rules {
		value, exists := data[rule.Field]

		// 检查必填字段
		if rule.Required && (!exists || value == nil || value == "") {
			errors = append(errors, fmt.Sprintf("%s is required", rule.Field))
			continue
		}

		// 如果字段不存在且不是必填，跳过其他验证
		if !exists {
			continue
		}

		// 转换为字符串进行验证
		strValue, ok := value.(string)
		if !ok {
			if value != nil {
				strValue = fmt.Sprintf("%v", value)
			}
		}

		// 长度验证
		if rule.MinLength > 0 && len(strValue) < rule.MinLength {
			errors = append(errors, fmt.Sprintf("%s must be at least %d characters", rule.Field, rule.MinLength))
		}

		if rule.MaxLength > 0 && len(strValue) > rule.MaxLength {
			errors = append(errors, fmt.Sprintf("%s must not exceed %d characters", rule.Field, rule.MaxLength))
		}

		// 正则验证
		if rule.Pattern != nil && strValue != "" && !rule.Pattern.MatchString(strValue) {
			errors = append(errors, fmt.Sprintf("%s format is invalid", rule.Field))
		}

		// 自定义验证
		if rule.Custom != nil {
			if err := rule.Custom(value); err != nil {
				errors = append(errors, err.Error())
			}
		}
	}

	return errors
}

// RequestValidationMiddleware 请求验证中间件
func RequestValidationMiddleware(validator *Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只验证 JSON 请求
		if !strings.Contains(c.GetHeader("Content-Type"), "application/json") {
			c.Next()
			return
		}

		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			common.ValidationError(c, "Failed to read request body")
			c.Abort()
			return
		}

		// 恢复请求体供后续使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// 解析 JSON
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			common.ValidationError(c, "Invalid JSON format")
			c.Abort()
			return
		}

		// 执行验证
		errors := validator.Validate(data)
		if len(errors) > 0 {
			common.ValidationError(c, strings.Join(errors, "; "))
			c.Abort()
			return
		}

		c.Next()
	}
}

// 常用验证规则
var (
	EmailPattern    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	PhonePattern    = regexp.MustCompile(`^(\+\d{1,3}[- ]?)?\d{10,11}$`)
	PasswordPattern = regexp.MustCompile(`^[a-zA-Z\d@$!%*?&]{8,}$`)  // 简化的密码模式
	UUIDPattern     = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

// 预定义验证器

// LoginValidator 登录验证器
func LoginValidator() *Validator {
	return NewValidator().
		AddRule(ValidationRule{
			Field:    "email",
			Required: true,
			Pattern:  EmailPattern,
		}).
		AddRule(ValidationRule{
			Field:     "password",
			Required:  true,
			MinLength: 1,
		})
}

// RegisterValidator 注册验证器
func RegisterValidator() *Validator {
	return NewValidator().
		AddRule(ValidationRule{
			Field:    "email",
			Required: true,
			Pattern:  EmailPattern,
		}).
		AddRule(ValidationRule{
			Field:     "password",
			Required:  true,
			MinLength: 8,
			Pattern:   PasswordPattern,
		}).
		AddRule(ValidationRule{
			Field:     "name",
			Required:  true,
			MinLength: 2,
			MaxLength: 50,
		})
}

// VerificationCodeValidator 验证码验证器
func VerificationCodeValidator() *Validator {
	return NewValidator().
		AddRule(ValidationRule{
			Field:    "email",
			Required: true,
			Pattern:  EmailPattern,
		}).
		AddRule(ValidationRule{
			Field:     "code",
			Required:  true,
			MinLength: 4,
			MaxLength: 8,
		})
}

