package carddetection

import (
	"crypto/md5"
	"crypto/des"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// CryptoUtils 加密工具
type CryptoUtils struct {
	appSecret string
}

// NewCryptoUtils 创建加密工具实例
func NewCryptoUtils(appSecret string) *CryptoUtils {
	return &CryptoUtils{
		appSecret: appSecret,
	}
}

// Sign 生成签名
func (c *CryptoUtils) Sign(params map[string]interface{}) (string, error) {
	// 移除sign字段
	delete(params, "sign")
	
	// 获取所有键并排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	// 构建查询字符串
	var query strings.Builder
	query.WriteString(c.appSecret)
	
	for _, key := range keys {
		value := c.formatValue(params[key])
		query.WriteString(key)
		query.WriteString(value)
	}
	query.WriteString(c.appSecret)
	
	// 计算MD5
	hash := md5.Sum([]byte(query.String()))
	return hex.EncodeToString(hash[:]), nil
}

// formatValue 格式化参数值，与Java版本保持一致
func (c *CryptoUtils) formatValue(value interface{}) string {
	switch v := value.(type) {
	case []string:
		// 数组格式化为JSON字符串，与Java的JSONArray.toString()一致
		if len(v) == 0 {
			return "[]"
		}
		var builder strings.Builder
		builder.WriteString("[")
		for i, item := range v {
			if i > 0 {
				builder.WriteString(",")
			}
			builder.WriteString("\"")
			builder.WriteString(item)
			builder.WriteString("\"")
		}
		builder.WriteString("]")
		return builder.String()
	case []interface{}:
		// 处理interface{}数组
		if len(v) == 0 {
			return "[]"
		}
		var builder strings.Builder
		builder.WriteString("[")
		for i, item := range v {
			if i > 0 {
				builder.WriteString(",")
			}
			builder.WriteString("\"")
			builder.WriteString(fmt.Sprintf("%v", item))
			builder.WriteString("\"")
		}
		builder.WriteString("]")
		return builder.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// VerifySign 验证签名
func (c *CryptoUtils) VerifySign(params map[string]interface{}, signToVerify string) bool {
	sign, err := c.Sign(params)
	if err != nil {
		return false
	}
	return sign == signToVerify
}

// DESEncrypt DES加密
func (c *CryptoUtils) DESEncrypt(plainText string) (string, error) {
	key := c.generateDESKey()
	
	block, err := des.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher failed: %w", err)
	}
	
	// PKCS7填充
	plainTextBytes := []byte(plainText)
	paddedData := c.pkcs7Pad(plainTextBytes, block.BlockSize())
	
	cipherText := make([]byte, len(paddedData))
	
	// ECB模式加密
	for i := 0; i < len(paddedData); i += block.BlockSize() {
		block.Encrypt(cipherText[i:i+block.BlockSize()], paddedData[i:i+block.BlockSize()])
	}
	
	return hex.EncodeToString(cipherText), nil
}

// DESDecrypt DES解密
func (c *CryptoUtils) DESDecrypt(cipherText string) (string, error) {
	key := c.generateDESKey()
	
	block, err := des.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher failed: %w", err)
	}
	
	// 解码十六进制
	cipherBytes, err := hex.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("decode hex failed: %w", err)
	}
	
	if len(cipherBytes)%block.BlockSize() != 0 {
		return "", fmt.Errorf("cipher text length is not multiple of block size")
	}
	
	plainText := make([]byte, len(cipherBytes))
	
	// ECB模式解密
	for i := 0; i < len(cipherBytes); i += block.BlockSize() {
		block.Decrypt(plainText[i:i+block.BlockSize()], cipherBytes[i:i+block.BlockSize()])
	}
	
	// 去除填充
	result, err := c.pkcs7Unpad(plainText)
	if err != nil {
		return "", fmt.Errorf("unpad failed: %w", err)
	}
	
	return string(result), nil
}

// generateDESKey 根据appSecret生成DES密钥
func (c *CryptoUtils) generateDESKey() []byte {
	secret := []byte(c.appSecret)
	key := make([]byte, 8)
	
	// 如果secret长度小于8，重复填充
	for i := 0; i < 8; i++ {
		key[i] = secret[i%len(secret)]
	}
	
	return key
}

// pkcs7Pad PKCS7填充
func (c *CryptoUtils) pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

// pkcs7Unpad PKCS7去填充
func (c *CryptoUtils) pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	
	// 验证填充是否正确
	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	
	return data[:len(data)-padding], nil
}

// AESModeExists 检查是否支持AES模式（根据文档，这里实际使用的是DES）
// 保留这个方法以便将来可能的升级
func (c *CryptoUtils) AESModeExists() bool {
	return false // 当前使用DES加密
}