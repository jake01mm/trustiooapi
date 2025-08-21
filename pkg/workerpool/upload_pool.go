package workerpool

import (
	"context"
	"fmt"
	"mime/multipart"
	"sync"
	"time"

	"trusioo_api/pkg/r2storage"
)

// 上传任务结构
type UploadTask struct {
	ID        string
	UserID    *int
	File      *multipart.FileHeader
	Options   r2storage.UploadOptions
	ResultCh  chan UploadResult
	CreatedAt time.Time
}

type UploadResult struct {
	TaskID    string
	Success   bool
	Result    *r2storage.UploadResult
	Error     error
	Duration  time.Duration
}

// 并发上传池
type UploadPool struct {
	workers    int
	queue      chan UploadTask
	r2Client   *r2storage.Client
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	metrics    *PoolMetrics
	maxRetries int
	retryDelay time.Duration
}

type PoolMetrics struct {
	mu                sync.RWMutex
	TotalTasks        int64
	CompletedTasks    int64
	FailedTasks       int64
	ActiveTasks       int64
	AverageProcessing time.Duration
	QueueSize         int64
	MaxQueueSize      int64
}

func NewUploadPool(workers int, queueSize int, r2Client *r2storage.Client, maxRetries int, retryDelay time.Duration) *UploadPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &UploadPool{
		workers:    workers,
		queue:      make(chan UploadTask, queueSize),
		r2Client:   r2Client,
		ctx:        ctx,
		cancel:     cancel,
		metrics:    &PoolMetrics{MaxQueueSize: int64(queueSize)},
		maxRetries: maxRetries,
		retryDelay: time.Duration(retryDelay) * time.Second,
	}
}

func (p *UploadPool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *UploadPool) Stop() {
	p.cancel()
	close(p.queue)
	p.wg.Wait()
}

func (p *UploadPool) worker(id int) {
	defer p.wg.Done()
	
	for {
		select {
		case task, ok := <-p.queue:
			if !ok {
				return
			}
			
			p.processTask(task)
			
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *UploadPool) processTask(task UploadTask) {
	start := time.Now()
	
	// 更新指标
	p.metrics.mu.Lock()
	p.metrics.ActiveTasks++
	p.metrics.mu.Unlock()
	
	defer func() {
		duration := time.Since(start)
		
		p.metrics.mu.Lock()
		p.metrics.ActiveTasks--
		p.metrics.CompletedTasks++
		// 计算平均处理时间
		if p.metrics.CompletedTasks == 1 {
			p.metrics.AverageProcessing = duration
		} else {
			p.metrics.AverageProcessing = (p.metrics.AverageProcessing + duration) / 2
		}
		p.metrics.mu.Unlock()
	}()
	
	// 执行上传任务，带重试机制
	result := p.uploadWithRetry(task)
	
	// 发送结果
	select {
	case task.ResultCh <- result:
	case <-time.After(5 * time.Second):
		// 避免阻塞，5秒后丢弃结果
	}
}

func (p *UploadPool) uploadWithRetry(task UploadTask) UploadResult {
	var lastErr error
	
	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(p.retryDelay * time.Duration(attempt)) // 指数退避
		}
		
		start := time.Now()
		result, err := p.r2Client.UploadFile(p.ctx, task.File, task.Options)
		duration := time.Since(start)
		
		if err == nil {
			return UploadResult{
				TaskID:   task.ID,
				Success:  true,
				Result:   result,
				Duration: duration,
			}
		}
		
		lastErr = err
		
		// 如果是上下文取消，直接退出
		if p.ctx.Err() != nil {
			break
		}
	}
	
	// 所有重试都失败了
	p.metrics.mu.Lock()
	p.metrics.FailedTasks++
	p.metrics.mu.Unlock()
	
	return UploadResult{
		TaskID:  task.ID,
		Success: false,
		Error:   fmt.Errorf("upload failed after %d attempts: %w", p.maxRetries+1, lastErr),
	}
}

// 提交上传任务
func (p *UploadPool) Submit(task UploadTask) error {
	p.metrics.mu.Lock()
	queueSize := int64(len(p.queue))
	p.metrics.QueueSize = queueSize
	p.metrics.TotalTasks++
	p.metrics.mu.Unlock()
	
	select {
	case p.queue <- task:
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("upload pool is shutting down")
	default:
		return fmt.Errorf("upload queue is full (capacity: %d)", cap(p.queue))
	}
}

// 获取池状态
func (p *UploadPool) GetMetrics() PoolMetrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()
	
	// 创建副本避免复制锁
	metrics := PoolMetrics{
		TotalTasks:        p.metrics.TotalTasks,
		CompletedTasks:    p.metrics.CompletedTasks,
		FailedTasks:       p.metrics.FailedTasks,
		ActiveTasks:       p.metrics.ActiveTasks,
		AverageProcessing: p.metrics.AverageProcessing,
		MaxQueueSize:      p.metrics.MaxQueueSize,
	}
	metrics.QueueSize = int64(len(p.queue))
	
	return metrics
}

// 批量上传接口
type BatchUploadRequest struct {
	Tasks []UploadTask
}

type BatchUploadResponse struct {
	TotalTasks     int
	SubmittedTasks int
	FailedTasks    int
	Results        []UploadResult
}

func (p *UploadPool) BatchSubmit(ctx context.Context, tasks []UploadTask, timeout time.Duration) *BatchUploadResponse {
	response := &BatchUploadResponse{
		TotalTasks: len(tasks),
		Results:    make([]UploadResult, 0, len(tasks)),
	}
	
	// 创建结果收集器
	resultCollector := make(chan UploadResult, len(tasks))
	
	// 提交所有任务
	for _, task := range tasks {
		task.ResultCh = resultCollector
		
		if err := p.Submit(task); err != nil {
			response.FailedTasks++
			response.Results = append(response.Results, UploadResult{
				TaskID:  task.ID,
				Success: false,
				Error:   err,
			})
		} else {
			response.SubmittedTasks++
		}
	}
	
	// 等待所有结果
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	for i := 0; i < response.SubmittedTasks; i++ {
		select {
		case result := <-resultCollector:
			response.Results = append(response.Results, result)
		case <-timeoutCtx.Done():
			// 超时，返回已收集的结果
			return response
		}
	}
	
	return response
}

// 健康检查
func (p *UploadPool) HealthCheck() map[string]interface{} {
	metrics := p.GetMetrics()
	
	health := map[string]interface{}{
		"status":            "healthy",
		"workers":           p.workers,
		"queue_size":        metrics.QueueSize,
		"max_queue_size":    metrics.MaxQueueSize,
		"total_tasks":       metrics.TotalTasks,
		"completed_tasks":   metrics.CompletedTasks,
		"failed_tasks":      metrics.FailedTasks,
		"active_tasks":      metrics.ActiveTasks,
		"average_duration":  metrics.AverageProcessing.String(),
	}
	
	// 判断健康状态
	if metrics.QueueSize >= metrics.MaxQueueSize*8/10 {
		health["status"] = "warning"
		health["message"] = "Queue is nearly full"
	}
	
	if metrics.FailedTasks > 0 && float64(metrics.FailedTasks)/float64(metrics.TotalTasks) > 0.1 {
		health["status"] = "critical"
		health["message"] = "High failure rate detected"
	}
	
	return health
}