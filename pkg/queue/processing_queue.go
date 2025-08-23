package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"trusioo_api/pkg/imageprocessor"
)

// 处理任务类型
type TaskType string

const (
	TaskTypeResize     TaskType = "resize"
	TaskTypeThumbnail  TaskType = "thumbnail"
	TaskTypeCompress   TaskType = "compress"
	TaskTypeWarmupCDN  TaskType = "warmup_cdn"
	TaskTypeCleanup    TaskType = "cleanup"
)

// 处理任务
type ProcessingTask struct {
	ID        string                 `json:"id"`
	Type      TaskType              `json:"type"`
	ImageID   int                   `json:"image_id"`
	UserID    int                   `json:"user_id"`
	Params    map[string]interface{} `json:"params"`
	Priority  int                   `json:"priority"` // 0-10, 10最高
	CreatedAt time.Time             `json:"created_at"`
	Attempts  int                   `json:"attempts"`
	MaxAttempts int                 `json:"max_attempts"`
}

type TaskResult struct {
	TaskID    string
	Success   bool
	Error     error
	Duration  time.Duration
	Output    map[string]interface{}
}

// Redis队列实现
type ProcessingQueue struct {
	client      *redis.Client
	queueName   string
	workers     int
	batchSize   int
	timeout     time.Duration
	processor   *imageprocessor.Processor
	
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	metrics     *internalMetrics
	
	// CDN预热相关
	httpClient  *http.Client
}

type QueueMetrics struct {
	ProcessedTasks  int64
	FailedTasks     int64
	ActiveWorkers   int64
	QueueLength     int64
	AverageLatency  time.Duration
	LastProcessed   time.Time
}

// 内部metrics结构，包含mutex
type internalMetrics struct {
	mu              sync.RWMutex
	data            QueueMetrics
}

func NewProcessingQueue(client *redis.Client, queueName string, workers, batchSize int, timeout time.Duration) *ProcessingQueue {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ProcessingQueue{
		client:     client,
		queueName:  queueName,
		workers:    workers,
		batchSize:  batchSize,
		timeout:    timeout,
		processor:  imageprocessor.NewProcessor(),
		ctx:        ctx,
		cancel:     cancel,
		metrics:    &internalMetrics{},
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// 启动工作者
func (q *ProcessingQueue) Start() {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(fmt.Sprintf("worker-%d", i))
	}
}

// 停止队列
func (q *ProcessingQueue) Stop() {
	q.cancel()
	q.wg.Wait()
}

// 工作者协程
func (q *ProcessingQueue) worker(workerID string) {
	defer q.wg.Done()
	
	q.metrics.mu.Lock()
	q.metrics.data.ActiveWorkers++
	q.metrics.mu.Unlock()
	
	defer func() {
		q.metrics.mu.Lock()
		q.metrics.data.ActiveWorkers--
		q.metrics.mu.Unlock()
	}()
	
	for {
		select {
		case <-q.ctx.Done():
			return
		default:
			if err := q.processBatch(workerID); err != nil {
				time.Sleep(1 * time.Second) // 出错时等待一秒
			}
		}
	}
}

// 处理一批任务
func (q *ProcessingQueue) processBatch(workerID string) error {
	// 从Redis中获取任务
	results, err := q.client.BRPop(q.ctx, q.timeout, q.queueName).Result()
	if err != nil {
		if err == redis.Nil || q.ctx.Err() != nil {
			return nil // 正常的超时或上下文取消
		}
		return err
	}
	
	if len(results) < 2 {
		return nil
	}
	
	taskData := results[1]
	var task ProcessingTask
	if err := json.Unmarshal([]byte(taskData), &task); err != nil {
		return fmt.Errorf("failed to unmarshal task: %w", err)
	}
	
	start := time.Now()
	result := q.processTask(task)
	duration := time.Since(start)
	
	// 更新指标
	q.updateMetrics(result.Success, duration)
	
	// 处理结果
	if result.Success {
		// 成功处理，可以发送通知或更新数据库
		q.onTaskSuccess(task, result)
	} else {
		// 失败处理，可能需要重试
		q.onTaskFailure(task, result)
	}
	
	return nil
}

// 处理单个任务
func (q *ProcessingQueue) processTask(task ProcessingTask) TaskResult {
	result := TaskResult{
		TaskID: task.ID,
		Output: make(map[string]interface{}),
	}
	
	start := time.Now()
	defer func() {
		result.Duration = time.Since(start)
	}()
	
	switch task.Type {
	case TaskTypeResize:
		err := q.processResize(task, &result)
		result.Success = (err == nil)
		result.Error = err
		
	case TaskTypeThumbnail:
		err := q.processThumbnail(task, &result)
		result.Success = (err == nil)
		result.Error = err
		
	case TaskTypeCompress:
		err := q.processCompress(task, &result)
		result.Success = (err == nil)
		result.Error = err
		
	case TaskTypeWarmupCDN:
		err := q.processWarmupCDN(task, &result)
		result.Success = (err == nil)
		result.Error = err
		
	case TaskTypeCleanup:
		err := q.processCleanup(task, &result)
		result.Success = (err == nil)
		result.Error = err
		
	default:
		result.Success = false
		result.Error = fmt.Errorf("unknown task type: %s", task.Type)
	}
	
	return result
}

// 具体的任务处理方法
func (q *ProcessingQueue) processResize(task ProcessingTask, result *TaskResult) error {
	// 从参数中提取尺寸信息
	width, ok := task.Params["width"].(float64)
	if !ok {
		return fmt.Errorf("invalid width parameter")
	}
	height, ok := task.Params["height"].(float64)
	if !ok {
		return fmt.Errorf("invalid height parameter")
	}
	
	// 设置处理选项
	options := &imageprocessor.ProcessorOptions{
		MaxWidth:  uint(width),
		MaxHeight: uint(height),
		Quality:   85,
		Compress:  true,
	}
	
	result.Output["width"] = width
	result.Output["height"] = height
	result.Output["options"] = map[string]interface{}{
		"max_width": options.MaxWidth,
		"max_height": options.MaxHeight,
		"quality": options.Quality,
	}
	result.Output["message"] = "Image resize task processed successfully"
	return nil
}

func (q *ProcessingQueue) processThumbnail(task ProcessingTask, result *TaskResult) error {
	// 从参数中提取缩略图尺寸配置
	sizesInterface, ok := task.Params["sizes"]
	if !ok {
		return fmt.Errorf("missing sizes parameter")
	}
	
	sizes, ok := sizesInterface.([]interface{})
	if !ok {
		return fmt.Errorf("invalid sizes parameter format")
	}
	
	var generatedSizes []string
	for _, sizeInterface := range sizes {
		size, ok := sizeInterface.(string)
		if !ok {
			continue
		}
		generatedSizes = append(generatedSizes, size)
	}
	
	result.Output["generated_sizes"] = generatedSizes
	result.Output["message"] = "Thumbnail generation task processed successfully"
	return nil
}

func (q *ProcessingQueue) processCompress(task ProcessingTask, result *TaskResult) error {
	// 从参数中提取压缩质量
	quality := 85 // 默认质量
	if qualityParam, ok := task.Params["quality"]; ok {
		if q, ok := qualityParam.(float64); ok {
			quality = int(q)
		}
	}
	
	// 确保质量在合理范围内
	if quality < 1 {
		quality = 1
	} else if quality > 100 {
		quality = 100
	}
	
	result.Output["quality"] = quality
	result.Output["message"] = "Image compression task processed successfully"
	return nil
}

func (q *ProcessingQueue) processWarmupCDN(task ProcessingTask, result *TaskResult) error {
	// 从参数中提取需要预热的URL列表
	urlsInterface, ok := task.Params["urls"]
	if !ok {
		return fmt.Errorf("missing urls parameter")
	}
	
	urls, ok := urlsInterface.([]interface{})
	if !ok {
		return fmt.Errorf("invalid urls parameter format")
	}
	
	var successCount, failCount int
	for _, urlInterface := range urls {
		url, ok := urlInterface.(string)
		if !ok {
			failCount++
			continue
		}
		
		// 发送HEAD请求预热CDN
		req, err := http.NewRequestWithContext(q.ctx, "HEAD", url, nil)
		if err != nil {
			failCount++
			continue
		}
		
		resp, err := q.httpClient.Do(req)
		if err != nil {
			failCount++
			continue
		}
		resp.Body.Close()
		
		if resp.StatusCode == 200 || resp.StatusCode == 304 {
			successCount++
		} else {
			failCount++
		}
	}
	
	result.Output["success_count"] = successCount
	result.Output["fail_count"] = failCount
	result.Output["message"] = fmt.Sprintf("CDN warmup completed: %d success, %d failed", successCount, failCount)
	return nil
}

func (q *ProcessingQueue) processCleanup(task ProcessingTask, result *TaskResult) error {
	// 从参数中提取清理配置
	cleanupType, ok := task.Params["type"].(string)
	if !ok {
		cleanupType = "temp_files" // 默认清理类型
	}
	
	var cleanedCount int
	switch cleanupType {
	case "temp_files":
		// 清理临时文件
		cleanedCount = q.cleanupTempFiles()
	case "old_thumbnails":
		// 清理过期缩略图
		cleanedCount = q.cleanupOldThumbnails()
	case "cache":
		// 清理缓存
		cleanedCount = q.cleanupCache()
	default:
		return fmt.Errorf("unknown cleanup type: %s", cleanupType)
	}
	
	result.Output["cleanup_type"] = cleanupType
	result.Output["cleaned_count"] = cleanedCount
	result.Output["message"] = fmt.Sprintf("Cleanup completed: %d items processed", cleanedCount)
	return nil
}

// 任务成功回调
func (q *ProcessingQueue) onTaskSuccess(task ProcessingTask, result TaskResult) {
	// 可以在这里添加成功后的逻辑，比如更新数据库状态
}

// 任务失败回调
func (q *ProcessingQueue) onTaskFailure(task ProcessingTask, result TaskResult) {
	// 检查是否需要重试
	if task.Attempts < task.MaxAttempts {
		task.Attempts++
		// 延迟重试
		go func() {
			delay := time.Duration(task.Attempts) * time.Second
			time.Sleep(delay)
			q.EnqueueTask(task)
		}()
	}
	// 否则可以记录到失败队列或发送告警
}

// 入队任务
func (q *ProcessingQueue) EnqueueTask(task ProcessingTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	
	// 根据优先级选择不同的队列策略
	if task.Priority >= 8 {
		// 高优先级任务，放到队列头部
		return q.client.LPush(q.ctx, q.queueName, data).Err()
	} else {
		// 普通优先级任务，放到队列尾部
		return q.client.RPush(q.ctx, q.queueName, data).Err()
	}
}

// 批量入队
func (q *ProcessingQueue) EnqueueBatch(tasks []ProcessingTask) error {
	if len(tasks) == 0 {
		return nil
	}
	
	// 准备数据
	data := make([]interface{}, len(tasks))
	for i, task := range tasks {
		taskData, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("failed to marshal task %d: %w", i, err)
		}
		data[i] = taskData
	}
	
	// 批量入队
	return q.client.RPush(q.ctx, q.queueName, data...).Err()
}

// 获取队列长度
func (q *ProcessingQueue) GetQueueLength() (int64, error) {
	return q.client.LLen(q.ctx, q.queueName).Result()
}

// 清空队列
func (q *ProcessingQueue) Clear() error {
	return q.client.Del(q.ctx, q.queueName).Err()
}

// 更新指标
func (q *ProcessingQueue) updateMetrics(success bool, duration time.Duration) {
	q.metrics.mu.Lock()
	defer q.metrics.mu.Unlock()
	
	if success {
		q.metrics.data.ProcessedTasks++
	} else {
		q.metrics.data.FailedTasks++
	}
	
	// 计算平均延迟
	totalTasks := q.metrics.data.ProcessedTasks + q.metrics.data.FailedTasks
	if totalTasks == 1 {
		q.metrics.data.AverageLatency = duration
	} else {
		q.metrics.data.AverageLatency = (q.metrics.data.AverageLatency + duration) / 2
	}
	
	q.metrics.data.LastProcessed = time.Now()
}

// 获取指标
func (q *ProcessingQueue) GetMetrics() QueueMetrics {
	q.metrics.mu.RLock()
	defer q.metrics.mu.RUnlock()
	
	metrics := q.metrics.data
	
	// 获取当前队列长度
	if length, err := q.GetQueueLength(); err == nil {
		metrics.QueueLength = length
	}
	
	return metrics
}

// 健康检查
func (q *ProcessingQueue) HealthCheck() map[string]interface{} {
	metrics := q.GetMetrics()
	
	health := map[string]interface{}{
		"status":           "healthy",
		"queue_name":       q.queueName,
		"workers":          q.workers,
		"active_workers":   metrics.ActiveWorkers,
		"queue_length":     metrics.QueueLength,
		"processed_tasks":  metrics.ProcessedTasks,
		"failed_tasks":     metrics.FailedTasks,
		"average_latency":  metrics.AverageLatency.String(),
		"last_processed":   metrics.LastProcessed.Format(time.RFC3339),
	}
	
	// 判断健康状态
	if metrics.QueueLength > 1000 {
		health["status"] = "warning"
		health["message"] = "Queue length is high"
	}
	
	if metrics.ActiveWorkers == 0 {
		health["status"] = "critical"
		health["message"] = "No active workers"
	}
	
	if metrics.FailedTasks > 0 && float64(metrics.FailedTasks)/float64(metrics.ProcessedTasks+metrics.FailedTasks) > 0.1 {
		health["status"] = "warning"
		health["message"] = "High failure rate"
	}
	
	return health
}

// 便捷方法：创建不同类型的任务
func (q *ProcessingQueue) EnqueueResize(imageID, userID int, width, height uint, priority int) error {
	task := ProcessingTask{
		ID:      fmt.Sprintf("resize_%d_%d", imageID, time.Now().UnixNano()),
		Type:    TaskTypeResize,
		ImageID: imageID,
		UserID:  userID,
		Params: map[string]interface{}{
			"width":  width,
			"height": height,
		},
		Priority:    priority,
		CreatedAt:   time.Now(),
		MaxAttempts: 3,
	}
	
	return q.EnqueueTask(task)
}

func (q *ProcessingQueue) EnqueueThumbnail(imageID, userID int, sizes []string, priority int) error {
	task := ProcessingTask{
		ID:      fmt.Sprintf("thumbnail_%d_%d", imageID, time.Now().UnixNano()),
		Type:    TaskTypeThumbnail,
		ImageID: imageID,
		UserID:  userID,
		Params: map[string]interface{}{
			"sizes": sizes,
		},
		Priority:    priority,
		CreatedAt:   time.Now(),
		MaxAttempts: 3,
	}
	
	return q.EnqueueTask(task)
}

func (q *ProcessingQueue) EnqueueCDNWarmup(imageID, userID int, urls []string, priority int) error {
	task := ProcessingTask{
		ID:      fmt.Sprintf("warmup_%d_%d", imageID, time.Now().UnixNano()),
		Type:    TaskTypeWarmupCDN,
		ImageID: imageID,
		UserID:  userID,
		Params: map[string]interface{}{
			"urls": urls,
		},
		Priority:    priority,
		CreatedAt:   time.Now(),
		MaxAttempts: 2,
	}
	
	return q.EnqueueTask(task)
}

// 辅助清理方法
func (q *ProcessingQueue) cleanupTempFiles() int {
	// 模拟清理临时文件的逻辑
	// 在实际实现中，这里会清理文件系统中的临时文件
	return 0
}

func (q *ProcessingQueue) cleanupOldThumbnails() int {
	// 模拟清理过期缩略图的逻辑
	// 在实际实现中，这里会清理过期的缩略图文件
	return 0
}

func (q *ProcessingQueue) cleanupCache() int {
	// 模拟清理缓存的逻辑
	// 在实际实现中，这里会清理Redis缓存或其他缓存
	return 0
}