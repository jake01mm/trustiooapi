package cdn

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"trusioo_api/config"
)

// CDN预热服务
type PrefetchService struct {
	httpClient *http.Client
	config     *config.Config
	workers    int
	
	// 工作池
	taskChan   chan PrefetchTask
	resultChan chan PrefetchResult
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	
	// 指标
	metrics    *PrefetchMetrics
	mu         sync.RWMutex
}

type PrefetchTask struct {
	ID       string
	URL      string
	Priority int
	Timeout  time.Duration
}

type PrefetchResult struct {
	TaskID     string
	URL        string
	Success    bool
	StatusCode int
	Duration   time.Duration
	Error      error
}

type PrefetchMetrics struct {
	TotalTasks    int64         `json:"total_tasks"`
	SuccessTasks  int64         `json:"success_tasks"`
	FailedTasks   int64         `json:"failed_tasks"`
	ActiveTasks   int64         `json:"active_tasks"`
	AverageLatency time.Duration `json:"average_latency"`
	LastPrefetch   time.Time     `json:"last_prefetch"`
}

func NewPrefetchService(config *config.Config) *PrefetchService {
	ctx, cancel := context.WithCancel(context.Background())
	
	workers := config.Performance.PrefetchWorkers
	if workers <= 0 {
		workers = 5
	}
	
	return &PrefetchService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxConnsPerHost:     10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		config:     config,
		workers:    workers,
		taskChan:   make(chan PrefetchTask, 1000),
		resultChan: make(chan PrefetchResult, 1000),
		ctx:        ctx,
		cancel:     cancel,
		metrics:    &PrefetchMetrics{},
	}
}

func (s *PrefetchService) Start() {
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
	
	// 启动结果处理器
	go s.resultProcessor()
}

func (s *PrefetchService) Stop() {
	s.cancel()
	close(s.taskChan)
	s.wg.Wait()
	close(s.resultChan)
}

func (s *PrefetchService) worker(id int) {
	defer s.wg.Done()
	
	for {
		select {
		case task, ok := <-s.taskChan:
			if !ok {
				return
			}
			
			result := s.processPrefetchTask(task)
			
			select {
			case s.resultChan <- result:
			case <-s.ctx.Done():
				return
			}
			
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *PrefetchService) processPrefetchTask(task PrefetchTask) PrefetchResult {
	start := time.Now()
	
	s.mu.Lock()
	s.metrics.ActiveTasks++
	s.mu.Unlock()
	
	defer func() {
		s.mu.Lock()
		s.metrics.ActiveTasks--
		s.mu.Unlock()
	}()
	
	timeout := task.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "HEAD", task.URL, nil)
	if err != nil {
		return PrefetchResult{
			TaskID:   task.ID,
			URL:      task.URL,
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("failed to create request: %w", err),
		}
	}
	
	// 添加预热相关的头部
	req.Header.Set("User-Agent", "TrusiooAPI-CDN-Prefetch/1.0")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return PrefetchResult{
			TaskID:   task.ID,
			URL:      task.URL,
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("request failed: %w", err),
		}
	}
	defer resp.Body.Close()
	
	success := resp.StatusCode >= 200 && resp.StatusCode < 400
	
	return PrefetchResult{
		TaskID:     task.ID,
		URL:        task.URL,
		Success:    success,
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start),
	}
}

func (s *PrefetchService) resultProcessor() {
	for result := range s.resultChan {
		s.updateMetrics(result)
	}
}

func (s *PrefetchService) updateMetrics(result PrefetchResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.metrics.TotalTasks++
	if result.Success {
		s.metrics.SuccessTasks++
	} else {
		s.metrics.FailedTasks++
	}
	
	// 更新平均延迟
	if s.metrics.TotalTasks == 1 {
		s.metrics.AverageLatency = result.Duration
	} else {
		s.metrics.AverageLatency = (s.metrics.AverageLatency + result.Duration) / 2
	}
	
	s.metrics.LastPrefetch = time.Now()
}

// 提交单个预热任务
func (s *PrefetchService) SubmitTask(task PrefetchTask) error {
	select {
	case s.taskChan <- task:
		return nil
	case <-s.ctx.Done():
		return fmt.Errorf("prefetch service is shutting down")
	default:
		return fmt.Errorf("prefetch queue is full")
	}
}

// 批量预热
func (s *PrefetchService) BatchPrefetch(urls []string, priority int) error {
	for i, url := range urls {
		task := PrefetchTask{
			ID:       fmt.Sprintf("batch_%d_%d", time.Now().Unix(), i),
			URL:      url,
			Priority: priority,
			Timeout:  30 * time.Second,
		}
		
		if err := s.SubmitTask(task); err != nil {
			return fmt.Errorf("failed to submit task for URL %s: %w", url, err)
		}
	}
	
	return nil
}

// 预热图片的所有变体
func (s *PrefetchService) PrefetchImageVariants(baseURL string, variants []string) error {
	urls := make([]string, len(variants))
	for i, variant := range variants {
		urls[i] = fmt.Sprintf("%s/%s", baseURL, variant)
	}
	
	return s.BatchPrefetch(urls, 5)
}

// 获取指标
func (s *PrefetchService) GetMetrics() PrefetchMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return *s.metrics
}

// 健康检查
func (s *PrefetchService) HealthCheck() map[string]interface{} {
	metrics := s.GetMetrics()
	
	health := map[string]interface{}{
		"status":           "healthy",
		"workers":          s.workers,
		"total_tasks":      metrics.TotalTasks,
		"success_tasks":    metrics.SuccessTasks,
		"failed_tasks":     metrics.FailedTasks,
		"active_tasks":     metrics.ActiveTasks,
		"average_latency":  metrics.AverageLatency.String(),
		"last_prefetch":    metrics.LastPrefetch.Format(time.RFC3339),
		"queue_capacity":   cap(s.taskChan),
		"queue_length":     len(s.taskChan),
	}
	
	// 健康状态判断
	if metrics.TotalTasks > 0 {
		failureRate := float64(metrics.FailedTasks) / float64(metrics.TotalTasks)
		if failureRate > 0.2 {
			health["status"] = "warning"
			health["message"] = "High failure rate detected"
		}
	}
	
	queueUsage := float64(len(s.taskChan)) / float64(cap(s.taskChan))
	if queueUsage > 0.8 {
		health["status"] = "warning"
		health["message"] = "Queue is nearly full"
	}
	
	return health
}

// 便捷方法：预热公开图片
func (s *PrefetchService) PrefetchPublicImage(imageKey string) error {
	if !s.config.Performance.EnableCDNPrefetch {
		return nil // CDN预热被禁用
	}
	
	baseURL := s.config.R2Storage.PublicCDNURL
	variants := []string{
		imageKey,                    // 原图
		fmt.Sprintf("thumb_%s", imageKey), // 缩略图
		fmt.Sprintf("small_%s", imageKey), // 小图
		fmt.Sprintf("medium_%s", imageKey), // 中图
	}
	
	return s.PrefetchImageVariants(baseURL, variants)
}

// 便捷方法：预热热门图片
func (s *PrefetchService) PrefetchPopularImages(imageKeys []string) error {
	if !s.config.Performance.EnableCDNPrefetch {
		return nil
	}
	
	for _, imageKey := range imageKeys {
		if err := s.PrefetchPublicImage(imageKey); err != nil {
			continue // 继续处理其他图片
		}
	}
	
	return nil
}