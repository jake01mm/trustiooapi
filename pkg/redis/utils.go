package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Set 设置键值对
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if Client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	// 如果value是结构体或map，序列化为JSON
	var val interface{}
	switch v := value.(type) {
	case string, int, int64, float64, bool:
		val = v
	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		val = string(jsonData)
	}

	return Client.Set(ctx, key, val, expiration).Err()
}

// Get 获取值
func Get(ctx context.Context, key string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("Redis client is not initialized")
	}

	val, err := Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key does not exist")
	}
	return val, err
}

// GetJSON 获取JSON值并反序列化
func GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// SetJSON 设置JSON值
func SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return Set(ctx, key, string(jsonData), expiration)
}

// Delete 删除键
func Delete(ctx context.Context, keys ...string) error {
	if Client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	return Client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func Exists(ctx context.Context, key string) (bool, error) {
	if Client == nil {
		return false, fmt.Errorf("Redis client is not initialized")
	}

	count, err := Client.Exists(ctx, key).Result()
	return count > 0, err
}

// Expire 设置键的过期时间
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	if Client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	return Client.Expire(ctx, key, expiration).Err()
}

// TTL 获取键的剩余生存时间
func TTL(ctx context.Context, key string) (time.Duration, error) {
	if Client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	return Client.TTL(ctx, key).Result()
}

// Increment 原子递增
func Increment(ctx context.Context, key string) (int64, error) {
	if Client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	return Client.Incr(ctx, key).Result()
}

// IncrementBy 原子递增指定值
func IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	if Client == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	return Client.IncrBy(ctx, key, value).Result()
}

// SetNX 仅当键不存在时设置
func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if Client == nil {
		return false, fmt.Errorf("Redis client is not initialized")
	}

	// 序列化值
	var val interface{}
	switch v := value.(type) {
	case string, int, int64, float64, bool:
		val = v
	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return false, fmt.Errorf("failed to marshal value: %w", err)
		}
		val = string(jsonData)
	}

	return Client.SetNX(ctx, key, val, expiration).Result()
}

// HSet 设置哈希字段
func HSet(ctx context.Context, key string, field string, value interface{}) error {
	if Client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	// 序列化值
	var val interface{}
	switch v := value.(type) {
	case string, int, int64, float64, bool:
		val = v
	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		val = string(jsonData)
	}

	return Client.HSet(ctx, key, field, val).Err()
}

// HGet 获取哈希字段值
func HGet(ctx context.Context, key, field string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("Redis client is not initialized")
	}

	val, err := Client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("field does not exist")
	}
	return val, err
}

// HGetAll 获取哈希的所有字段和值
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if Client == nil {
		return nil, fmt.Errorf("Redis client is not initialized")
	}

	return Client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func HDel(ctx context.Context, key string, fields ...string) error {
	if Client == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	return Client.HDel(ctx, key, fields...).Err()
}