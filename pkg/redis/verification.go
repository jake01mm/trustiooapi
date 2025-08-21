package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// VerificationCache Redis验证码缓存服务
type VerificationCache struct {
	client *redis.Client
}

// NewVerificationCache 创建验证码缓存服务
func NewVerificationCache() *VerificationCache {
	return &VerificationCache{
		client: GetClient(),
	}
}

// StoreVerificationCode 存储验证码到Redis
func (vc *VerificationCache) StoreVerificationCode(target, vType, code string, expiry time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("vc:%s:%s", target, vType)
	return vc.client.SetEx(ctx, key, code, expiry).Err()
}

// GetVerificationCode 从Redis获取验证码
func (vc *VerificationCache) GetVerificationCode(target, vType string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("vc:%s:%s", target, vType)
	return vc.client.Get(ctx, key).Result()
}

// DeleteVerificationCode 删除验证码（验证成功后）
func (vc *VerificationCache) DeleteVerificationCode(target, vType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("vc:%s:%s", target, vType)
	return vc.client.Del(ctx, key).Err()
}

// CheckSendFrequency 检查发送频率限制
func (vc *VerificationCache) CheckSendFrequency(target, vType string, cooldown time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("vc_rate:%s:%s", target, vType)
	exists, err := vc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil // true表示在冷却期内，不能发送
}

// SetSendFrequency 设置发送频率限制
func (vc *VerificationCache) SetSendFrequency(target, vType string, cooldown time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("vc_rate:%s:%s", target, vType)
	return vc.client.SetEx(ctx, key, "1", cooldown).Err()
}

// GetAttemptCount 获取验证失败次数
func (vc *VerificationCache) GetAttemptCount(target, vType string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("vc_attempts:%s:%s", target, vType)
	val, err := vc.client.Get(ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, nil // 键不存在，返回0
		}
		return 0, err
	}

	var count int
	if _, err := fmt.Sscanf(val, "%d", &count); err != nil {
		return 0, err
	}

	return count, nil
}

// IncrementAttemptCount 增加验证失败次数
func (vc *VerificationCache) IncrementAttemptCount(target, vType string, expiry time.Duration) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("vc_attempts:%s:%s", target, vType)

	// 使用原子操作增加计数
	count, err := vc.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// 如果是第一次设置，添加过期时间
	if count == 1 {
		vc.client.Expire(ctx, key, expiry)
	}

	return int(count), nil
}

// ClearAttemptCount 清除验证失败次数
func (vc *VerificationCache) ClearAttemptCount(target, vType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("vc_attempts:%s:%s", target, vType)
	return vc.client.Del(ctx, key).Err()
}

// IsBlocked 检查是否因失败次数过多而被阻止
func (vc *VerificationCache) IsBlocked(target, vType string, maxAttempts int) (bool, error) {
	count, err := vc.GetAttemptCount(target, vType)
	if err != nil {
		return false, err
	}

	return count >= maxAttempts, nil
}
