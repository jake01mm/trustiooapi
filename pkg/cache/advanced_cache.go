package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"trusioo_api/internal/images/entities"
)

// 高级缓存管理器
type AdvancedCacheManager struct {
	client      *redis.Client
	imageCache  *ImageCache
	config      CacheConfig
	
	// 缓存层级
	l1Cache     *LRUCache // 内存缓存
	l2Cache     *redis.Client // Redis缓存
	
	// 预热管理
	warmupChan  chan WarmupTask
	warmupWg    sync.WaitGroup
	
	// 指标
	metrics     *CacheMetrics
	mu          sync.RWMutex
}

type CacheConfig struct {
	L1MaxSize        int
	L1TTL            time.Duration
	L2TTL            time.Duration
	WarmupWorkers    int
	WarmupQueueSize  int
	EnablePreload    bool
	PreloadInterval  time.Duration
}

type CacheMetrics struct {
	L1Hits       int64 `json:"l1_hits"`
	L1Misses     int64 `json:"l1_misses"`
	L2Hits       int64 `json:"l2_hits"`
	L2Misses     int64 `json:"l2_misses"`
	TotalHits    int64 `json:"total_hits"`
	TotalMisses  int64 `json:"total_misses"`
	HitRate      float64 `json:"hit_rate"`
	WarmupTasks  int64 `json:"warmup_tasks"`
	LastWarmup   time.Time `json:"last_warmup"`
}

type WarmupTask struct {
	Key     string
	ImageID int
	UserID  *int
	Priority int
}

// LRU内存缓存实现
type LRUCache struct {
	maxSize   int
	ttl       time.Duration
	items     map[string]*CacheItem
	access    map[string]time.Time
	mu        sync.RWMutex
}

type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

func NewLRUCache(maxSize int, ttl time.Duration) *LRUCache {
	cache := &LRUCache{
		maxSize: maxSize,
		ttl:     ttl,
		items:   make(map[string]*CacheItem),
		access:  make(map[string]time.Time),
	}
	
	// 启动清理协程
	go cache.cleanup()
	
	return cache
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, exists := c.items[key]
	if !exists {
		return nil, false
	}
	
	if time.Now().After(item.ExpiresAt) {
		delete(c.items, key)
		delete(c.access, key)
		return nil, false
	}
	
	c.access[key] = time.Now()
	return item.Value, true
}

func (c *LRUCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 检查是否需要驱逐
	if len(c.items) >= c.maxSize {
		c.evict()
	}
	
	c.items[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	c.access[key] = time.Now()
}

func (c *LRUCache) evict() {
	// 找到最久未访问的项目
	var oldestKey string
	var oldestTime time.Time = time.Now()
	
	for key, accessTime := range c.access {
		if accessTime.Before(oldestTime) {
			oldestTime = accessTime
			oldestKey = key
		}
	}
	
	if oldestKey != "" {
		delete(c.items, oldestKey)
		delete(c.access, oldestKey)
	}
}

func (c *LRUCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.ExpiresAt) {
				delete(c.items, key)
				delete(c.access, key)
			}
		}
		c.mu.Unlock()
	}
}

func NewAdvancedCacheManager(client *redis.Client, config CacheConfig) *AdvancedCacheManager {
	if config.L1MaxSize <= 0 {
		config.L1MaxSize = 1000
	}
	if config.L1TTL <= 0 {
		config.L1TTL = 5 * time.Minute
	}
	if config.L2TTL <= 0 {
		config.L2TTL = time.Hour
	}
	if config.WarmupWorkers <= 0 {
		config.WarmupWorkers = 3
	}
	if config.WarmupQueueSize <= 0 {
		config.WarmupQueueSize = 1000
	}
	
	manager := &AdvancedCacheManager{
		client:     client,
		imageCache: NewImageCache(client, config.L2TTL),
		config:     config,
		l1Cache:    NewLRUCache(config.L1MaxSize, config.L1TTL),
		l2Cache:    client,
		warmupChan: make(chan WarmupTask, config.WarmupQueueSize),
		metrics:    &CacheMetrics{},
	}
	
	// 启动预热工作者
	for i := 0; i < config.WarmupWorkers; i++ {
		manager.warmupWg.Add(1)
		go manager.warmupWorker()
	}
	
	// 启动预加载
	if config.EnablePreload {
		go manager.preloadScheduler()
	}
	
	return manager
}

func (m *AdvancedCacheManager) GetImage(ctx context.Context, imageID int) (*entities.Image, error) {
	key := fmt.Sprintf("image:%d", imageID)
	
	// 尝试L1缓存
	if value, hit := m.l1Cache.Get(key); hit {
		m.updateMetrics(true, false)
		if image, ok := value.(*entities.Image); ok {
			return image, nil
		}
	}
	
	// 尝试L2缓存
	image, err := m.imageCache.GetImage(ctx, imageID)
	if err != nil {
		return nil, err
	}
	
	if image != nil {
		m.updateMetrics(false, true)
		// 回填L1缓存
		m.l1Cache.Set(key, image)
		return image, nil
	}
	
	// 缓存未命中
	m.updateMetrics(false, false)
	return nil, nil
}

func (m *AdvancedCacheManager) SetImage(ctx context.Context, image *entities.Image) error {
	key := fmt.Sprintf("image:%d", image.ID)
	
	// 设置L1缓存
	m.l1Cache.Set(key, image)
	
	// 设置L2缓存
	return m.imageCache.SetImage(ctx, image)
}

func (m *AdvancedCacheManager) InvalidateImage(ctx context.Context, imageID int) error {
	key := fmt.Sprintf("image:%d", imageID)
	
	// 清除L1缓存
	m.l1Cache.mu.Lock()
	delete(m.l1Cache.items, key)
	delete(m.l1Cache.access, key)
	m.l1Cache.mu.Unlock()
	
	// 清除L2缓存
	return m.imageCache.InvalidateImage(ctx, imageID)
}

func (m *AdvancedCacheManager) WarmupImage(imageID int, userID *int, priority int) {
	task := WarmupTask{
		Key:     fmt.Sprintf("warmup:%d", imageID),
		ImageID: imageID,
		UserID:  userID,
		Priority: priority,
	}
	
	select {
	case m.warmupChan <- task:
	default:
		// 队列满了，丢弃低优先级任务
	}
}

func (m *AdvancedCacheManager) warmupWorker() {
	defer m.warmupWg.Done()
	
	for task := range m.warmupChan {
		m.processWarmupTask(task)
	}
}

func (m *AdvancedCacheManager) processWarmupTask(task WarmupTask) {
	ctx := context.Background()
	
	// 检查是否已经在缓存中
	if _, hit := m.l1Cache.Get(fmt.Sprintf("image:%d", task.ImageID)); hit {
		return
	}
	
	// 这里应该从数据库加载图片并缓存
	// 为了示例，我们模拟一个预热操作
	m.mu.Lock()
	m.metrics.WarmupTasks++
	m.metrics.LastWarmup = time.Now()
	m.mu.Unlock()
	
	// 在实际实现中，这里会从Repository加载图片数据
	_ = ctx
}

func (m *AdvancedCacheManager) preloadScheduler() {
	if m.config.PreloadInterval <= 0 {
		m.config.PreloadInterval = 10 * time.Minute
	}
	
	ticker := time.NewTicker(m.config.PreloadInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		m.preloadPopularImages()
	}
}

func (m *AdvancedCacheManager) preloadPopularImages() {
	// 在实际实现中，这里会查询最近访问频率高的图片
	// 然后预加载到缓存中
	
	// 示例：预热一些图片ID
	popularImageIDs := []int{1, 2, 3, 4, 5}
	for _, imageID := range popularImageIDs {
		m.WarmupImage(imageID, nil, 3)
	}
}

func (m *AdvancedCacheManager) updateMetrics(l1Hit, l2Hit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if l1Hit {
		m.metrics.L1Hits++
		m.metrics.TotalHits++
	} else if l2Hit {
		m.metrics.L2Hits++
		m.metrics.TotalHits++
		m.metrics.L1Misses++
	} else {
		m.metrics.L1Misses++
		m.metrics.L2Misses++
		m.metrics.TotalMisses++
	}
	
	total := m.metrics.TotalHits + m.metrics.TotalMisses
	if total > 0 {
		m.metrics.HitRate = float64(m.metrics.TotalHits) / float64(total)
	}
}

func (m *AdvancedCacheManager) GetMetrics() CacheMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return *m.metrics
}

func (m *AdvancedCacheManager) HealthCheck() map[string]interface{} {
	metrics := m.GetMetrics()
	
	health := map[string]interface{}{
		"status":        "healthy",
		"l1_cache_size": len(m.l1Cache.items),
		"l1_max_size":   m.l1Cache.maxSize,
		"hit_rate":      fmt.Sprintf("%.2f%%", metrics.HitRate*100),
		"total_hits":    metrics.TotalHits,
		"total_misses":  metrics.TotalMisses,
		"warmup_tasks":  metrics.WarmupTasks,
		"last_warmup":   metrics.LastWarmup.Format(time.RFC3339),
	}
	
	// 健康状态判断
	if metrics.HitRate < 0.7 && metrics.TotalHits+metrics.TotalMisses > 100 {
		health["status"] = "warning"
		health["message"] = "Low cache hit rate"
	}
	
	l1Usage := float64(len(m.l1Cache.items)) / float64(m.l1Cache.maxSize)
	if l1Usage > 0.9 {
		health["status"] = "warning"
		health["message"] = "L1 cache nearly full"
	}
	
	return health
}

func (m *AdvancedCacheManager) Stop() {
	close(m.warmupChan)
	m.warmupWg.Wait()
}