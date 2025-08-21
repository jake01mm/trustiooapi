package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// 系统监控器
type SystemMonitor struct {
	redis       *redis.Client
	collectors  map[string]MetricsCollector
	config      MonitorConfig
	
	// 指标存储
	metrics     *SystemMetrics
	mu          sync.RWMutex
	
	// 运行控制
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

type MonitorConfig struct {
	CollectionInterval time.Duration
	RetentionPeriod   time.Duration
	AlertThresholds   AlertThresholds
	EnableAlerts      bool
}

type AlertThresholds struct {
	HighCPU        float64 // CPU使用率告警阈值
	HighMemory     float64 // 内存使用率告警阈值
	HighLatency    time.Duration // 延迟告警阈值
	LowHitRate     float64 // 缓存命中率告警阈值
	HighErrorRate  float64 // 错误率告警阈值
}

type SystemMetrics struct {
	Timestamp       time.Time `json:"timestamp"`
	
	// API指标
	TotalRequests   int64     `json:"total_requests"`
	SuccessRequests int64     `json:"success_requests"`
	ErrorRequests   int64     `json:"error_requests"`
	AverageLatency  time.Duration `json:"average_latency"`
	
	// 上传指标
	TotalUploads    int64     `json:"total_uploads"`
	SuccessUploads  int64     `json:"success_uploads"`
	FailedUploads   int64     `json:"failed_uploads"`
	UploadLatency   time.Duration `json:"upload_latency"`
	
	// 缓存指标
	CacheHitRate    float64   `json:"cache_hit_rate"`
	CacheHits       int64     `json:"cache_hits"`
	CacheMisses     int64     `json:"cache_misses"`
	
	// 队列指标
	QueueLength     int64     `json:"queue_length"`
	ProcessedTasks  int64     `json:"processed_tasks"`
	FailedTasks     int64     `json:"failed_tasks"`
	
	// CDN指标
	CDNRequests     int64     `json:"cdn_requests"`
	CDNLatency      time.Duration `json:"cdn_latency"`
	PrefetchTasks   int64     `json:"prefetch_tasks"`
	
	// 系统指标
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     float64   `json:"memory_usage"`
	DiskUsage       float64   `json:"disk_usage"`
	ActiveConns     int64     `json:"active_connections"`
}

// 指标收集器接口
type MetricsCollector interface {
	CollectMetrics(ctx context.Context) (map[string]interface{}, error)
	Name() string
}

// API指标收集器
type APIMetricsCollector struct {
	name string
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewAPIMetricsCollector() *APIMetricsCollector {
	return &APIMetricsCollector{
		name: "api",
		data: make(map[string]interface{}),
	}
}

func (c *APIMetricsCollector) Name() string {
	return c.name
}

func (c *APIMetricsCollector) CollectMetrics(ctx context.Context) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// 复制数据避免并发问题
	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	
	return result, nil
}

func (c *APIMetricsCollector) RecordRequest(success bool, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.incrementCounter("total_requests")
	if success {
		c.incrementCounter("success_requests")
	} else {
		c.incrementCounter("error_requests")
	}
	
	c.updateAverageLatency("average_latency", duration)
}

func (c *APIMetricsCollector) incrementCounter(key string) {
	if val, exists := c.data[key]; exists {
		if count, ok := val.(int64); ok {
			c.data[key] = count + 1
		}
	} else {
		c.data[key] = int64(1)
	}
}

func (c *APIMetricsCollector) updateAverageLatency(key string, duration time.Duration) {
	if val, exists := c.data[key]; exists {
		if avg, ok := val.(time.Duration); ok {
			c.data[key] = (avg + duration) / 2
		}
	} else {
		c.data[key] = duration
	}
}

// 上传指标收集器
type UploadMetricsCollector struct {
	name string
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewUploadMetricsCollector() *UploadMetricsCollector {
	return &UploadMetricsCollector{
		name: "upload",
		data: make(map[string]interface{}),
	}
}

func (c *UploadMetricsCollector) Name() string {
	return c.name
}

func (c *UploadMetricsCollector) CollectMetrics(ctx context.Context) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	
	return result, nil
}

func (c *UploadMetricsCollector) RecordUpload(success bool, duration time.Duration, fileSize int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.incrementCounter("total_uploads")
	if success {
		c.incrementCounter("success_uploads")
		c.addFileSize("total_bytes_uploaded", fileSize)
	} else {
		c.incrementCounter("failed_uploads")
	}
	
	c.updateAverageLatency("upload_latency", duration)
}

func (c *UploadMetricsCollector) incrementCounter(key string) {
	if val, exists := c.data[key]; exists {
		if count, ok := val.(int64); ok {
			c.data[key] = count + 1
		}
	} else {
		c.data[key] = int64(1)
	}
}

func (c *UploadMetricsCollector) updateAverageLatency(key string, duration time.Duration) {
	if val, exists := c.data[key]; exists {
		if avg, ok := val.(time.Duration); ok {
			c.data[key] = (avg + duration) / 2
		}
	} else {
		c.data[key] = duration
	}
}

func (c *UploadMetricsCollector) addFileSize(key string, size int64) {
	if val, exists := c.data[key]; exists {
		if total, ok := val.(int64); ok {
			c.data[key] = total + size
		}
	} else {
		c.data[key] = size
	}
}

func NewSystemMonitor(redis *redis.Client, config MonitorConfig) *SystemMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	if config.CollectionInterval <= 0 {
		config.CollectionInterval = 30 * time.Second
	}
	if config.RetentionPeriod <= 0 {
		config.RetentionPeriod = 24 * time.Hour
	}
	
	monitor := &SystemMonitor{
		redis:      redis,
		collectors: make(map[string]MetricsCollector),
		config:     config,
		metrics:    &SystemMetrics{},
		ctx:        ctx,
		cancel:     cancel,
	}
	
	// 注册默认收集器
	monitor.RegisterCollector(NewAPIMetricsCollector())
	monitor.RegisterCollector(NewUploadMetricsCollector())
	
	return monitor
}

func (m *SystemMonitor) RegisterCollector(collector MetricsCollector) {
	m.collectors[collector.Name()] = collector
}

func (m *SystemMonitor) Start() {
	m.wg.Add(1)
	go m.collectLoop()
}

func (m *SystemMonitor) Stop() {
	m.cancel()
	m.wg.Wait()
}

func (m *SystemMonitor) collectLoop() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(m.config.CollectionInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.collectMetrics()
		case <-m.ctx.Done():
			return
		}
	}
}

func (m *SystemMonitor) collectMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.metrics.Timestamp = time.Now()
	
	// 收集各个组件的指标
	for name, collector := range m.collectors {
		data, err := collector.CollectMetrics(m.ctx)
		if err != nil {
			continue
		}
		
		m.mergeMetrics(name, data)
	}
	
	// 存储到Redis
	m.storeMetrics()
	
	// 检查告警
	if m.config.EnableAlerts {
		m.checkAlerts()
	}
}

func (m *SystemMonitor) mergeMetrics(collectorName string, data map[string]interface{}) {
	switch collectorName {
	case "api":
		if val, ok := data["total_requests"].(int64); ok {
			m.metrics.TotalRequests = val
		}
		if val, ok := data["success_requests"].(int64); ok {
			m.metrics.SuccessRequests = val
		}
		if val, ok := data["error_requests"].(int64); ok {
			m.metrics.ErrorRequests = val
		}
		if val, ok := data["average_latency"].(time.Duration); ok {
			m.metrics.AverageLatency = val
		}
		
	case "upload":
		if val, ok := data["total_uploads"].(int64); ok {
			m.metrics.TotalUploads = val
		}
		if val, ok := data["success_uploads"].(int64); ok {
			m.metrics.SuccessUploads = val
		}
		if val, ok := data["failed_uploads"].(int64); ok {
			m.metrics.FailedUploads = val
		}
		if val, ok := data["upload_latency"].(time.Duration); ok {
			m.metrics.UploadLatency = val
		}
	}
}

func (m *SystemMonitor) storeMetrics() {
	data, err := json.Marshal(m.metrics)
	if err != nil {
		return
	}
	
	key := "system_metrics:" + time.Now().Format("2006-01-02:15:04")
	m.redis.Set(m.ctx, key, data, m.config.RetentionPeriod).Err()
	
	// 保存最新指标
	m.redis.Set(m.ctx, "system_metrics:latest", data, 0).Err()
}

func (m *SystemMonitor) checkAlerts() {
	thresholds := m.config.AlertThresholds
	
	// 检查延迟告警
	if thresholds.HighLatency > 0 && m.metrics.AverageLatency > thresholds.HighLatency {
		m.sendAlert("HIGH_LATENCY", "Average latency is too high", map[string]interface{}{
			"current_latency": m.metrics.AverageLatency.String(),
			"threshold": thresholds.HighLatency.String(),
		})
	}
	
	// 检查错误率告警
	if thresholds.HighErrorRate > 0 && m.metrics.TotalRequests > 0 {
		errorRate := float64(m.metrics.ErrorRequests) / float64(m.metrics.TotalRequests)
		if errorRate > thresholds.HighErrorRate {
			m.sendAlert("HIGH_ERROR_RATE", "Error rate is too high", map[string]interface{}{
				"current_error_rate": errorRate,
				"threshold": thresholds.HighErrorRate,
			})
		}
	}
	
	// 检查缓存命中率告警
	if thresholds.LowHitRate > 0 && m.metrics.CacheHitRate < thresholds.LowHitRate {
		m.sendAlert("LOW_CACHE_HIT_RATE", "Cache hit rate is too low", map[string]interface{}{
			"current_hit_rate": m.metrics.CacheHitRate,
			"threshold": thresholds.LowHitRate,
		})
	}
}

func (m *SystemMonitor) sendAlert(alertType, message string, data map[string]interface{}) {
	alert := map[string]interface{}{
		"type":      alertType,
		"message":   message,
		"timestamp": time.Now(),
		"data":      data,
	}
	
	alertData, _ := json.Marshal(alert)
	
	// 发送到告警队列
	m.redis.LPush(m.ctx, "system_alerts", alertData).Err()
	
	// 保留最近100条告警
	m.redis.LTrim(m.ctx, "system_alerts", 0, 99).Err()
}

// 获取当前指标
func (m *SystemMonitor) GetMetrics() SystemMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return *m.metrics
}

// 获取历史指标
func (m *SystemMonitor) GetHistoricalMetrics(ctx context.Context, hours int) ([]SystemMetrics, error) {
	if hours <= 0 {
		hours = 1
	}
	
	var metrics []SystemMetrics
	now := time.Now()
	
	for i := 0; i < hours*2; i++ { // 每30分钟一个数据点
		t := now.Add(-time.Duration(i*30) * time.Minute)
		key := "system_metrics:" + t.Format("2006-01-02:15:04")
		
		data, err := m.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		
		var metric SystemMetrics
		if json.Unmarshal([]byte(data), &metric) == nil {
			metrics = append(metrics, metric)
		}
	}
	
	return metrics, nil
}

// 获取告警列表
func (m *SystemMonitor) GetAlerts(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}
	
	alertsData, err := m.redis.LRange(ctx, "system_alerts", 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}
	
	var alerts []map[string]interface{}
	for _, data := range alertsData {
		var alert map[string]interface{}
		if json.Unmarshal([]byte(data), &alert) == nil {
			alerts = append(alerts, alert)
		}
	}
	
	return alerts, nil
}

// 健康检查
func (m *SystemMonitor) HealthCheck() map[string]interface{} {
	metrics := m.GetMetrics()
	
	health := map[string]interface{}{
		"status":           "healthy",
		"total_requests":   metrics.TotalRequests,
		"error_requests":   metrics.ErrorRequests,
		"average_latency":  metrics.AverageLatency.String(),
		"total_uploads":    metrics.TotalUploads,
		"failed_uploads":   metrics.FailedUploads,
		"cache_hit_rate":   fmt.Sprintf("%.2f%%", metrics.CacheHitRate*100),
		"last_collection":  metrics.Timestamp.Format(time.RFC3339),
	}
	
	// 健康状态判断
	if metrics.TotalRequests > 0 {
		errorRate := float64(metrics.ErrorRequests) / float64(metrics.TotalRequests)
		if errorRate > 0.1 {
			health["status"] = "warning"
			health["message"] = "High error rate detected"
		}
	}
	
	if metrics.AverageLatency > 5*time.Second {
		health["status"] = "warning"
		health["message"] = "High latency detected"
	}
	
	return health
}