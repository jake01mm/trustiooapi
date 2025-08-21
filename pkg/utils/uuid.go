package utils

import (
	"github.com/google/uuid"
)

// GenerateUUID 生成新的 UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// ParseUUID 解析 UUID 字符串
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// ValidateUUID 验证 UUID 格式
func ValidateUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// GenerateUUIDWithoutHyphens 生成不带连字符的 UUID
func GenerateUUIDWithoutHyphens() string {
	id := uuid.New()
	return id.String()[:8] + id.String()[9:13] + id.String()[14:18] + id.String()[19:23] + id.String()[24:]
}

// MustParseUUID 解析 UUID，如果失败则 panic
func MustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}