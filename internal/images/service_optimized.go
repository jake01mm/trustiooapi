package images

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"sync"
	"time"

	"trusioo_api/internal/images/dto"
	"trusioo_api/internal/images/entities"
	"trusioo_api/pkg/cache"
	"trusioo_api/pkg/queue"
	"trusioo_api/pkg/r2storage"
	"trusioo_api/pkg/workerpool"
	"trusioo_api/config"
	"github.com/redis/go-redis/v9"
)

// 优化版图片服务
type OptimizedImageService struct {
	repo             Repository
	r2Client         *r2storage.Client
	cache            *cache.ImageCache
	uploadPool       *workerpool.UploadPool
	processingQueue  *queue.ProcessingQueue
	config           *config.Config
	
	// 性能监控
	metrics          *ServiceMetrics
	mu               sync.RWMutex
}

type ServiceMetrics struct {
	TotalRequests     int64     `json:"total_requests"`
	CacheHits         int64     `json:"cache_hits"`
	CacheMisses       int64     `json:"cache_misses"`
	AverageLatency    time.Duration `json:"average_latency"`
	ActiveUploads     int64     `json:"active_uploads"`
	QueuedTasks       int64     `json:"queued_tasks"`
	LastRequestTime   time.Time `json:"last_request_time"`
}

func NewOptimizedImageService(
	repo Repository,
	r2Client *r2storage.Client,
	redisClient *redis.Client,
	config *config.Config,
) *OptimizedImageService {
	// 创建缓存层
	cacheExpiration := time.Duration(config.Performance.CacheExpiration) * time.Second
	imageCache := cache.NewImageCache(redisClient, cacheExpiration)
	
	// 创建上传池
	uploadPool := workerpool.NewUploadPool(
		config.Performance.UploadWorkerPool,
		config.Performance.QueueSize,
		r2Client,
		config.R2Storage.MaxRetries,
		time.Duration(config.R2Storage.RetryDelay)*time.Second,
	)
	
	// 创建处理队列
	processingQueue := queue.NewProcessingQueue(
		redisClient,
		"image_processing_queue",
		config.Performance.ProcessingWorkerPool,
		config.Performance.BatchSize,
		time.Duration(config.Performance.ProcessingTimeout)*time.Second,
	)
	
	service := &OptimizedImageService{
		repo:            repo,
		r2Client:        r2Client,
		cache:          imageCache,
		uploadPool:     uploadPool,
		processingQueue: processingQueue,
		config:         config,
		metrics:        &ServiceMetrics{},
	}
	
	// 启动工作池和队列
	uploadPool.Start()
	processingQueue.Start()
	
	return service
}

// 实现Service接口的优化版本
func (s *OptimizedImageService) UploadImage(ctx context.Context, userID *int, file *multipart.FileHeader, req dto.UploadImageRequest) (*dto.UploadImageResponse, error) {
	start := time.Now()
	defer s.updateMetrics(start)
	
	// 生成文件名
	fileName := req.FileName
	if fileName == "" {
		ext := filepath.Ext(file.Filename)
		fileName = fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	}
	
	// 创建上传任务
	uploadTask := workerpool.UploadTask{
		ID:     fmt.Sprintf("upload_%s_%d", fileName, time.Now().UnixNano()),
		UserID: userID,
		File:   file,
		Options: r2storage.UploadOptions{
			IsPublic: req.IsPublic,
			Folder:   req.Folder,
			FileName: fileName,
		},
		ResultCh:  make(chan workerpool.UploadResult, 1),
		CreatedAt: time.Now(),
	}
	
	// 提交到上传池
	if err := s.uploadPool.Submit(uploadTask); err != nil {
		return nil, fmt.Errorf("failed to submit upload task: %w", err)
	}
	
	s.mu.Lock()
	s.metrics.ActiveUploads++
	s.mu.Unlock()
	
	defer func() {
		s.mu.Lock()
		s.metrics.ActiveUploads--
		s.mu.Unlock()
	}()
	
	// 等待上传结果
	select {
	case result := <-uploadTask.ResultCh:
		if !result.Success {
			return nil, fmt.Errorf("upload failed: %w", result.Error)
		}
		
		// 创建图片记录
		image := &entities.Image{
			UserID:       userID,
			FileName:     fileName,
			OriginalName: file.Filename,
			Key:          result.Result.Key,
			Bucket:       result.Result.Bucket,
			URL:          result.Result.URL,
			ContentType:  file.Header.Get("Content-Type"),
			Size:         result.Result.Size,
			IsPublic:     req.IsPublic,
		}
		
		if result.Result.PublicURL != "" {
			image.PublicURL = &result.Result.PublicURL
		}
		
		if req.Folder != "" {
			image.Folder = &req.Folder
		}
		
		// 保存到数据库
		err := s.repo.Create(ctx, image)
		if err != nil {
			// 上传成功但数据库保存失败，清理R2文件
			s.r2Client.DeleteFile(ctx, result.Result.Bucket, result.Result.Key)
			return nil, fmt.Errorf("failed to save image metadata: %w", err)
		}
		
		// 异步处理任务
		go s.schedulePostProcessing(image)
		
		// 使缓存失效
		if userID != nil {
			s.cache.InvalidateUserImageLists(ctx, *userID)
		}
		s.cache.InvalidateAdminImageLists(ctx)
		
		return &dto.UploadImageResponse{
			ID:           image.ID,
			FileName:     image.FileName,
			OriginalName: image.OriginalName,
			Key:          image.Key,
			URL:          image.URL,
			PublicURL:    image.PublicURL,
			ContentType:  image.ContentType,
			Size:         image.Size,
			IsPublic:     image.IsPublic,
			Folder:       image.Folder,
		}, nil
		
	case <-ctx.Done():
		return nil, ctx.Err()
		
	case <-time.After(time.Duration(s.config.R2Storage.UploadTimeout) * time.Second):
		return nil, fmt.Errorf("upload timeout after %d seconds", s.config.R2Storage.UploadTimeout)
	}
}

// 优化版图片获取（带缓存）
func (s *OptimizedImageService) GetUserImage(ctx context.Context, userID int, imageID int) (*dto.GetImageResponse, error) {
	start := time.Now()
	defer s.updateMetrics(start)
	
	// 先尝试从缓存获取
	if s.config.Performance.EnableCache {
		cachedImage, err := s.cache.GetImage(ctx, imageID)
		if err == nil && cachedImage != nil {
			// 验证所有权
			if cachedImage.UserID != nil && *cachedImage.UserID == userID {
				s.mu.Lock()
				s.metrics.CacheHits++
				s.mu.Unlock()
				
				return s.imageToResponse(ctx, cachedImage), nil
			}
		}
	}
	
	// 缓存未命中，从数据库获取
	s.mu.Lock()
	s.metrics.CacheMisses++
	s.mu.Unlock()
	
	image, err := s.repo.GetByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}
	
	// 验证所有权
	if image.UserID == nil || *image.UserID != userID {
		return nil, fmt.Errorf("access denied: image belongs to another user")
	}
	
	// 更新缓存
	if s.config.Performance.EnableCache {
		s.cache.SetImage(ctx, image)
	}
	
	return s.imageToResponse(ctx, image), nil
}

// 优化版公开图片获取
func (s *OptimizedImageService) GetPublicImageByKey(ctx context.Context, key string) (*dto.GetImageResponse, error) {
	start := time.Now()
	defer s.updateMetrics(start)
	
	// 先尝试从缓存获取
	if s.config.Performance.EnableCache {
		cachedImage, err := s.cache.GetImageByKey(ctx, key)
		if err == nil && cachedImage != nil && cachedImage.IsPublic {
			s.mu.Lock()
			s.metrics.CacheHits++
			s.mu.Unlock()
			
			return s.imageToResponse(ctx, cachedImage), nil
		}
	}
	
	// 缓存未命中，从数据库获取
	s.mu.Lock()
	s.metrics.CacheMisses++
	s.mu.Unlock()
	
	image, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}
	
	// 只允许访问公开图片
	if !image.IsPublic {
		return nil, fmt.Errorf("access denied: this is a private image")
	}
	
	// 更新缓存
	if s.config.Performance.EnableCache {
		s.cache.SetImageByKey(ctx, key, image)
	}
	
	return s.imageToResponse(ctx, image), nil
}

// 优化版图片列表（带缓存）
func (s *OptimizedImageService) ListImages(ctx context.Context, userID *int, req dto.ListImagesRequest) (*dto.ListImagesResponse, error) {
	start := time.Now()
	defer s.updateMetrics(start)
	
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	
	// 先尝试从缓存获取
	if s.config.Performance.EnableCache && userID != nil {
		cachedResponse, err := s.cache.GetUserImageList(ctx, *userID, req.Folder, req.IsPublic, req.Page, req.PageSize)
		if err == nil && cachedResponse != nil {
			s.mu.Lock()
			s.metrics.CacheHits++
			s.mu.Unlock()
			
			return cachedResponse, nil
		}
	}
	
	// 缓存未命中，从数据库获取
	s.mu.Lock()
	s.metrics.CacheMisses++
	s.mu.Unlock()
	
	offset := (req.Page - 1) * req.PageSize
	images, total, err := s.repo.List(ctx, userID, req.Folder, req.IsPublic, offset, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	
	// 并发处理URL生成
	imageResponses := make([]dto.GetImageResponse, len(images))
	var wg sync.WaitGroup
	
	for i, image := range images {
		wg.Add(1)
		go func(idx int, img *entities.Image) {
			defer wg.Done()
			imageResponses[idx] = *s.imageToResponse(ctx, img)
		}(i, image)
	}
	
	wg.Wait()
	
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	
	response := &dto.ListImagesResponse{
		Images:     imageResponses,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}
	
	// 更新缓存
	if s.config.Performance.EnableCache && userID != nil {
		s.cache.SetUserImageList(ctx, *userID, req.Folder, req.IsPublic, req.Page, req.PageSize, response)
	}
	
	return response, nil
}

// 管理员接口的优化版本
func (s *OptimizedImageService) AdminListAllImages(ctx context.Context, req dto.ListImagesRequest) (*dto.ListImagesResponse, error) {
	start := time.Now()
	defer s.updateMetrics(start)
	
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	
	// 先尝试从缓存获取
	if s.config.Performance.EnableCache {
		cachedResponse, err := s.cache.GetAdminImageList(ctx, req.Folder, req.IsPublic, req.Page, req.PageSize)
		if err == nil && cachedResponse != nil {
			s.mu.Lock()
			s.metrics.CacheHits++
			s.mu.Unlock()
			
			return cachedResponse, nil
		}
	}
	
	// 缓存未命中，从数据库获取
	s.mu.Lock()
	s.metrics.CacheMisses++
	s.mu.Unlock()
	
	offset := (req.Page - 1) * req.PageSize
	images, total, err := s.repo.List(ctx, nil, req.Folder, req.IsPublic, offset, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	
	// 并发处理URL生成
	imageResponses := make([]dto.GetImageResponse, len(images))
	var wg sync.WaitGroup
	
	for i, image := range images {
		wg.Add(1)
		go func(idx int, img *entities.Image) {
			defer wg.Done()
			imageResponses[idx] = *s.imageToResponse(ctx, img)
		}(i, image)
	}
	
	wg.Wait()
	
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	
	response := &dto.ListImagesResponse{
		Images:     imageResponses,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}
	
	// 更新缓存
	if s.config.Performance.EnableCache {
		s.cache.SetAdminImageList(ctx, req.Folder, req.IsPublic, req.Page, req.PageSize, response)
	}
	
	return response, nil
}

// 辅助方法
func (s *OptimizedImageService) imageToResponse(ctx context.Context, image *entities.Image) *dto.GetImageResponse {
	// 如果是私有图片且没有公开URL，异步生成预签名URL
	if !image.IsPublic && image.PublicURL == nil {
		go func() {
			newURL, err := s.r2Client.GeneratePresignedURL(context.Background(), image.Bucket, image.Key, 24*time.Hour)
			if err == nil {
				image.URL = newURL
				s.repo.Update(context.Background(), image)
				
				// 更新缓存
				if s.config.Performance.EnableCache {
					s.cache.SetImage(context.Background(), image)
				}
			}
		}()
	}
	
	return &dto.GetImageResponse{
		ID:           image.ID,
		FileName:     image.FileName,
		OriginalName: image.OriginalName,
		Key:          image.Key,
		URL:          image.URL,
		PublicURL:    image.PublicURL,
		ContentType:  image.ContentType,
		Size:         image.Size,
		IsPublic:     image.IsPublic,
		Folder:       image.Folder,
		CreatedAt:    image.CreatedAt.Format(time.RFC3339),
	}
}

// 调度后处理任务
func (s *OptimizedImageService) schedulePostProcessing(image *entities.Image) {
	if !s.config.Performance.EnableQueue {
		return
	}
	
	userID := 0
	if image.UserID != nil {
		userID = *image.UserID
	}
	
	// 如果是公开图片，预热CDN
	if image.IsPublic && s.config.Performance.EnableCDNPrefetch {
		urls := []string{image.URL}
		if image.PublicURL != nil {
			urls = append(urls, *image.PublicURL)
		}
		s.processingQueue.EnqueueCDNWarmup(image.ID, userID, urls, 5)
	}
	
	// 生成缩略图
	thumbnailSizes := []string{"small", "medium", "large"}
	s.processingQueue.EnqueueThumbnail(image.ID, userID, thumbnailSizes, 3)
}

// 更新性能指标
func (s *OptimizedImageService) updateMetrics(start time.Time) {
	duration := time.Since(start)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.metrics.TotalRequests++
	s.metrics.LastRequestTime = time.Now()
	
	// 计算平均延迟
	if s.metrics.TotalRequests == 1 {
		s.metrics.AverageLatency = duration
	} else {
		s.metrics.AverageLatency = (s.metrics.AverageLatency + duration) / 2
	}
}

// 获取服务指标
func (s *OptimizedImageService) GetMetrics() ServiceMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	metrics := *s.metrics
	
	// 获取实时队列状态
	uploadMetrics := s.uploadPool.GetMetrics()
	queueMetrics := s.processingQueue.GetMetrics()
	
	metrics.ActiveUploads = uploadMetrics.ActiveTasks
	metrics.QueuedTasks = queueMetrics.QueueLength
	
	return metrics
}

// 健康检查
func (s *OptimizedImageService) HealthCheck() map[string]interface{} {
	metrics := s.GetMetrics()
	uploadHealth := s.uploadPool.HealthCheck()
	queueHealth := s.processingQueue.HealthCheck()
	cacheStats := s.cache.GetStats(context.Background())
	
	health := map[string]interface{}{
		"status":             "healthy",
		"service_metrics":    metrics,
		"upload_pool":        uploadHealth,
		"processing_queue":   queueHealth,
		"cache":             cacheStats,
		"cache_hit_rate":    float64(metrics.CacheHits) / float64(metrics.CacheHits+metrics.CacheMisses),
	}
	
	// 综合判断健康状态
	if uploadHealth["status"] != "healthy" || queueHealth["status"] != "healthy" {
		health["status"] = "warning"
	}
	
	return health
}

// 关闭服务
func (s *OptimizedImageService) Shutdown() {
	s.uploadPool.Stop()
	s.processingQueue.Stop()
}