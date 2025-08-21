package middleware

import (
	"net/http"
	"sync"
	"time"

	"trusioo_api/internal/common"

	"github.com/gin-gonic/gin"
)

// RateLimiter 简单的内存率限制器
type RateLimiter struct {
	visitors map[string]*Visitor
	mutex    sync.RWMutex
}

type Visitor struct {
	limiter  *TokenBucket
	lastSeen time.Time
}

type TokenBucket struct {
	tokens    int
	capacity  int
	rate      time.Duration
	lastRefill time.Time
	mutex     sync.Mutex
}

// NewTokenBucket 创建新的令牌桶
func NewTokenBucket(capacity int, rate time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// Allow 检查是否允许请求
func (tb *TokenBucket) Allow() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	
	// 根据时间间隔添加令牌
	tokensToAdd := int(elapsed / tb.rate)
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
	}

	// 启动清理 goroutine
	go rl.cleanupVisitors()
	return rl
}

// cleanupVisitors 清理过期的访问者
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		for ip, visitor := range rl.visitors {
			if time.Since(visitor.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mutex.Unlock()
	}
}

// getVisitor 获取或创建访问者
func (rl *RateLimiter) getVisitor(ip string) *Visitor {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	visitor, exists := rl.visitors[ip]
	if !exists {
		// 创建新访问者: 每分钟100个请求
		visitor = &Visitor{
			limiter: NewTokenBucket(100, time.Minute/100),
		}
		rl.visitors[ip] = visitor
	}

	visitor.lastSeen = time.Now()
	return visitor
}

// RateLimitMiddleware 速率限制中间件
func RateLimitMiddleware(rateLimiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		visitor := rateLimiter.getVisitor(ip)

		if !visitor.limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, common.Response{
				Code:    429,
				Message: "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthRateLimitMiddleware 认证接口的严格速率限制
func AuthRateLimitMiddleware() gin.HandlerFunc {
	limiter := NewRateLimiter()
	
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		// 为认证相关接口创建更严格的限制
		visitor, exists := limiter.visitors[ip]
		if !exists {
			// 每分钟只允许10次认证请求
			visitor = &Visitor{
				limiter: NewTokenBucket(10, time.Minute/10),
			}
			limiter.visitors[ip] = visitor
		}
		
		visitor.lastSeen = time.Now()

		if !visitor.limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, common.Response{
				Code:    429,
				Message: "Too many authentication attempts. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}