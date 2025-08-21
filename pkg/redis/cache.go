package redis

import (
	"context"
	"fmt"
	"time"
)

// CacheService Redis缓存服务
type CacheService struct {
	prefix string
}

// NewCacheService 创建新的缓存服务实例
func NewCacheService(prefix string) *CacheService {
	return &CacheService{
		prefix: prefix,
	}
}

// buildKey 构建带前缀的键名
func (c *CacheService) buildKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

// Set 设置缓存
func (c *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return Set(ctx, c.buildKey(key), value, expiration)
}

// Get 获取缓存
func (c *CacheService) Get(ctx context.Context, key string) (string, error) {
	return Get(ctx, c.buildKey(key))
}

// GetJSON 获取JSON缓存
func (c *CacheService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	return GetJSON(ctx, c.buildKey(key), dest)
}

// Delete 删除缓存
func (c *CacheService) Delete(ctx context.Context, key string) error {
	return Delete(ctx, c.buildKey(key))
}

// Exists 检查缓存是否存在
func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	return Exists(ctx, c.buildKey(key))
}

// SetNX 仅当键不存在时设置缓存
func (c *CacheService) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return SetNX(ctx, c.buildKey(key), value, expiration)
}

// Remember 记忆模式：如果缓存存在则返回，否则执行函数并缓存结果
func (c *CacheService) Remember(ctx context.Context, key string, expiration time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// 尝试从缓存获取
	val, err := c.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	// 缓存不存在，执行函数
	result, err := fn()
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if setErr := c.Set(ctx, key, result, expiration); setErr != nil {
		// 记录错误但不影响返回结果
		fmt.Printf("Failed to cache result for key %s: %v\n", key, setErr)
	}

	return result, nil
}

// RememberJSON 记忆模式的JSON版本
func (c *CacheService) RememberJSON(ctx context.Context, key string, expiration time.Duration, dest interface{}, fn func() (interface{}, error)) error {
	// 尝试从缓存获取
	err := c.GetJSON(ctx, key, dest)
	if err == nil {
		return nil
	}

	// 缓存不存在，执行函数
	result, err := fn()
	if err != nil {
		return err
	}

	// 缓存结果
	if setErr := c.Set(ctx, key, result, expiration); setErr != nil {
		// 记录错误但不影响返回结果
		fmt.Printf("Failed to cache result for key %s: %v\n", key, setErr)
	}

	// 将结果复制到dest
	switch v := result.(type) {
	case string:
		return GetJSON(ctx, c.buildKey(key), dest)
	default:
		// 如果result不是字符串，直接赋值
		*dest.(*interface{}) = v
		return nil
	}
}

// Flush 清空指定前缀的所有缓存
func (c *CacheService) Flush(ctx context.Context) error {
	if Client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	pattern := c.buildKey("*")
	keys, err := Client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return Client.Del(ctx, keys...).Err()
	}

	return nil
}

// 预定义的缓存服务实例
var (
	// UserCache 用户相关缓存
	UserCache = NewCacheService("user")
	// AdminCache 管理员相关缓存
	AdminCache = NewCacheService("admin")
	// SessionCache 会话相关缓存
	SessionCache = NewCacheService("session")
	// TokenCache 令牌相关缓存
	TokenCache = NewCacheService("token")
	// RateLimitCache 限流相关缓存
	RateLimitCache = NewCacheService("ratelimit")
	// TempCache 临时缓存
	TempCache = NewCacheService("temp")
)