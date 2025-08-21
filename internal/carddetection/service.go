package carddetection

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"trusioo_api/internal/carddetection/dto"
	"trusioo_api/internal/carddetection/entities"
	"trusioo_api/pkg/carddetection"
	"trusioo_api/pkg/logger"
	"trusioo_api/pkg/utils"
)

// Service 卡片检测服务接口
type Service interface {
	// 核心检测功能
	CheckCard(ctx context.Context, userID int64, req *carddetection.CheckCardRequest) (*carddetection.CheckCardResponse, error)
	CheckCardResult(ctx context.Context, userID int64, req *carddetection.CheckCardResultRequest) (*carddetection.CardResult, error)
	
	// CD产品和区域查询
	GetCDProducts() (*dto.CDProductsResponse, error)
	GetCDRegions(productMark string) (*dto.CDRegionsResponse, error)
	
	// 历史记录查询
	GetUserHistory(userID int64, req *dto.CardDetectionHistoryRequest) (*dto.CardDetectionHistoryResponse, error)
	GetRecordDetail(userID int64, recordID int64) (*dto.CardDetectionRecordResponse, error)
	GetUserStats(userID int64) (*dto.CardDetectionStatsResponse, error)
	GetUserSummary(userID int64) (*dto.CardDetectionSummaryResponse, error)
	
	// 服务状态
	GetServiceStatus() map[string]interface{}
}

// service Service接口的实现
type service struct {
	client     *carddetection.Client
	repository Repository
}

// NewService 创建新的服务实例
func NewService(client *carddetection.Client, repository Repository) Service {
	return &service{
		client:     client,
		repository: repository,
	}
}

// CheckCard 执行卡片检测并记录到数据库
func (s *service) CheckCard(ctx context.Context, userID int64, req *carddetection.CheckCardRequest) (*carddetection.CheckCardResponse, error) {
	if s.client == nil {
		return nil, fmt.Errorf("card detection service is not available")
	}
	
	requestID := utils.GenerateUUID()
	startTime := time.Now()
	
	// 为每张卡创建记录
	var records []*entities.CardDetectionRecord
	for _, cardNumber := range req.Cards {
		record := &entities.CardDetectionRecord{
			UserID:      userID,
			RequestID:   requestID,
			CardNumber:  cardNumber,
			ProductMark: req.ProductMark,
			CheckStatus: "pending",
			CreatedAt:   startTime,
			UpdatedAt:   startTime,
		}
		
		// 设置可选字段
		if req.RegionID != 0 {
			record.RegionID = &req.RegionID
		}
		if req.RegionName != "" {
			record.RegionName = &req.RegionName
		}
		if req.AutoType != 0 {
			record.AutoType = &req.AutoType
		}
		
		records = append(records, record)
	}
	
	// 批量创建记录
	if err := s.repository.CreateBatch(records); err != nil {
		logger.WithError(err).Error("Failed to create detection records")
		return nil, fmt.Errorf("failed to create detection records: %w", err)
	}
	
	logger.WithFields(map[string]interface{}{
		"request_id":   requestID,
		"user_id":      userID,
		"product_mark": req.ProductMark,
		"cards_count":  len(req.Cards),
	}).Info("Card detection records created")
	
	// 执行检测
	resp, err := s.client.CheckCard(ctx, req)
	responseTime := int(time.Since(startTime).Milliseconds())
	
	// 更新记录状态
	for _, record := range records {
		var status string
		var errorMessage string
		var responseCode int
		
		if err != nil {
			status = "failed"
			errorMessage = err.Error()
			if carddetection.IsCardDetectionError(err) {
				cdErr := err.(*carddetection.CardDetectionError)
				responseCode = cdErr.Code
			} else {
				responseCode = 500
			}
		} else {
			status = "completed"
			responseCode = resp.Code
		}
		
		updateErr := s.repository.UpdateStatus(record.ID, status, resp, errorMessage)
		if updateErr != nil {
			logger.WithError(updateErr).WithFields(map[string]interface{}{
				"record_id":  record.ID,
				"request_id": requestID,
			}).Error("Failed to update detection record status")
		}
		
		// 更新响应时间
		if responseTime > 0 {
			record.ResponseTime = &responseTime
			record.ResponseCode = &responseCode
			s.repository.Update(record)
		}
	}
	
	if err != nil {
		logger.WithError(err).WithFields(map[string]interface{}{
			"request_id": requestID,
			"user_id":    userID,
		}).Error("Card detection failed")
		return nil, err
	}
	
	logger.WithFields(map[string]interface{}{
		"request_id":     requestID,
		"user_id":        userID,
		"response_code":  resp.Code,
		"response_time":  responseTime,
	}).Info("Card detection completed successfully")
	
	return resp, nil
}

// CheckCardResult 查询卡片检测结果并记录到数据库
func (s *service) CheckCardResult(ctx context.Context, userID int64, req *carddetection.CheckCardResultRequest) (*carddetection.CardResult, error) {
	if s.client == nil {
		return nil, fmt.Errorf("card detection service is not available")
	}
	
	requestID := utils.GenerateUUID()
	startTime := time.Now()
	
	// 创建查询记录
	record := &entities.CardDetectionRecord{
		UserID:      userID,
		RequestID:   requestID,
		CardNumber:  req.CardNo,
		ProductMark: req.ProductMark,
		CheckStatus: "pending",
		CreatedAt:   startTime,
		UpdatedAt:   startTime,
	}
	
	if req.PinCode != "" {
		record.PinCode = &req.PinCode
	}
	
	// 创建记录
	if err := s.repository.Create(record); err != nil {
		logger.WithError(err).Error("Failed to create result query record")
		return nil, fmt.Errorf("failed to create result query record: %w", err)
	}
	
	// 执行查询
	result, err := s.client.CheckCardResult(ctx, req)
	responseTime := int(time.Since(startTime).Milliseconds())
	
	// 更新记录状态
	var status string
	var errorMessage string
	var responseCode int
	
	if err != nil {
		status = "failed"
		errorMessage = err.Error()
		if carddetection.IsCardDetectionError(err) {
			cdErr := err.(*carddetection.CardDetectionError)
			responseCode = cdErr.Code
		} else {
			responseCode = 500
		}
	} else {
		status = "completed"
		responseCode = 200
	}
	
	updateErr := s.repository.UpdateStatus(record.ID, status, result, errorMessage)
	if updateErr != nil {
		logger.WithError(updateErr).WithFields(map[string]interface{}{
			"record_id":  record.ID,
			"request_id": requestID,
		}).Error("Failed to update result query record status")
	}
	
	// 更新响应时间
	if responseTime > 0 {
		record.ResponseTime = &responseTime
		record.ResponseCode = &responseCode
		s.repository.Update(record)
	}
	
	if err != nil {
		logger.WithError(err).WithFields(map[string]interface{}{
			"request_id": requestID,
			"user_id":    userID,
		}).Error("Card result query failed")
		return nil, err
	}
	
	logger.WithFields(map[string]interface{}{
		"request_id":    requestID,
		"user_id":       userID,
		"card_no":       result.CardNo,
		"response_time": responseTime,
	}).Info("Card result query completed successfully")
	
	return result, nil
}

// GetUserHistory 获取用户检测历史
func (s *service) GetUserHistory(userID int64, req *dto.CardDetectionHistoryRequest) (*dto.CardDetectionHistoryResponse, error) {
	// 设置默认分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	
	offset := (req.Page - 1) * req.PageSize
	
	// 根据条件查询记录
	var records []*entities.CardDetectionRecord
	var err error
	var total int
	
	if req.Status != "" {
		records, err = s.repository.GetByUserIDAndStatus(userID, req.Status, req.PageSize, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get records by status: %w", err)
		}
		total, err = s.repository.CountByUserIDAndStatus(userID, req.Status)
	} else if req.ProductMark != "" {
		records, err = s.repository.GetByUserIDAndProductMark(userID, req.ProductMark, req.PageSize, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get records by product mark: %w", err)
		}
		total, err = s.repository.CountByUserID(userID)
	} else if req.CardNumber != "" {
		records, err = s.repository.GetByCardNumber(userID, req.CardNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to get records by card number: %w", err)
		}
		total = len(records)
	} else {
		records, err = s.repository.GetByUserID(userID, req.PageSize, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get user records: %w", err)
		}
		total, err = s.repository.CountByUserID(userID)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}
	
	// 转换为响应格式
	recordResponses := make([]*dto.CardDetectionRecordResponse, len(records))
	for i, record := range records {
		recordResponses[i] = s.entityToResponse(record)
	}
	
	// 计算分页信息
	totalPages := (total + req.PageSize - 1) / req.PageSize
	
	pagination := dto.PaginationResponse{
		Page:        req.Page,
		PageSize:    req.PageSize,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     req.Page < totalPages,
		HasPrevious: req.Page > 1,
	}
	
	// 获取汇总信息
	summary, err := s.GetUserSummary(userID)
	if err != nil {
		logger.WithError(err).Error("Failed to get user summary")
		// 不返回错误，汇总信息是可选的
		summary = nil
	}
	
	return &dto.CardDetectionHistoryResponse{
		Records:    recordResponses,
		Pagination: pagination,
		Summary:    summary,
	}, nil
}

// GetRecordDetail 获取检测记录详情
func (s *service) GetRecordDetail(userID int64, recordID int64) (*dto.CardDetectionRecordResponse, error) {
	record, err := s.repository.GetByID(recordID)
	if err != nil {
		return nil, fmt.Errorf("failed to get record: %w", err)
	}
	
	// 验证记录属于该用户
	if record.UserID != userID {
		return nil, fmt.Errorf("record not found")
	}
	
	return s.entityToResponse(record), nil
}

// GetUserStats 获取用户检测统计
func (s *service) GetUserStats(userID int64) (*dto.CardDetectionStatsResponse, error) {
	// 获取汇总信息
	summary, err := s.GetUserSummary(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user summary: %w", err)
	}
	
	// 获取最近的检测记录
	recentRecords, err := s.repository.GetByUserID(userID, 5, 0)
	if err != nil {
		logger.WithError(err).Error("Failed to get recent records")
		recentRecords = []*entities.CardDetectionRecord{}
	}
	
	recentResponses := make([]*dto.CardDetectionRecordResponse, len(recentRecords))
	for i, record := range recentRecords {
		recentResponses[i] = s.entityToResponse(record)
	}
	
	return &dto.CardDetectionStatsResponse{
		Summary:      summary,
		ProductStats: []*dto.ProductStatsResponse{}, // TODO: 实现产品统计
		MonthlyStats: []*dto.MonthlyStatsResponse{},  // TODO: 实现月度统计
		RecentChecks: recentResponses,
	}, nil
}

// GetUserSummary 获取用户检测汇总
func (s *service) GetUserSummary(userID int64) (*dto.CardDetectionSummaryResponse, error) {
	summary, err := s.repository.GetSummaryByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user summary: %w", err)
	}
	
	pendingCount, err := s.repository.CountByUserIDAndStatus(userID, "pending")
	if err != nil {
		pendingCount = 0
	}
	
	var successRate float64
	if summary.TotalChecks > 0 {
		successRate = float64(summary.SuccessChecks) / float64(summary.TotalChecks) * 100
	}
	
	return &dto.CardDetectionSummaryResponse{
		TotalChecks:   summary.TotalChecks,
		SuccessChecks: summary.SuccessChecks,
		FailedChecks:  summary.FailedChecks,
		PendingChecks: pendingCount,
		SuccessRate:   successRate,
		LastCheckAt:   summary.LastCheckAt,
	}, nil
}

// GetServiceStatus 获取服务状态
func (s *service) GetServiceStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled": s.client != nil,
		"service": "Card Detection API",
	}
	
	if s.client != nil {
		if err := s.client.ValidateConfig(); err != nil {
			status["config_valid"] = false
			status["config_error"] = err.Error()
		} else {
			status["config_valid"] = true
		}
	}
	
	return status
}

// entityToResponse 将实体转换为响应格式
func (s *service) entityToResponse(record *entities.CardDetectionRecord) *dto.CardDetectionRecordResponse {
	resp := &dto.CardDetectionRecordResponse{
		ID:           record.ID,
		RequestID:    record.RequestID,
		CardNumber:   record.CardNumber,
		PinCode:      record.PinCode,
		ProductMark:  record.ProductMark,
		RegionID:     record.RegionID,
		RegionName:   record.RegionName,
		AutoType:     record.AutoType,
		CheckStatus:  record.CheckStatus,
		ErrorMessage: record.ErrorMessage,
		ResponseCode: record.ResponseCode,
		ResponseTime: record.ResponseTime,
		CheckedAt:    record.CheckedAt,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
	}
	
	// 解析JSON结果
	if record.CheckResult != nil && *record.CheckResult != "" {
		var result interface{}
		if err := json.Unmarshal([]byte(*record.CheckResult), &result); err == nil {
			resp.CheckResult = result
		} else {
			// 如果JSON解析失败，直接返回字符串
			resp.CheckResult = *record.CheckResult
		}
	}
	
	return resp
}

// GetCDProducts 获取CD产品列表
func (s *service) GetCDProducts() (*dto.CDProductsResponse, error) {
	products, err := s.repository.GetCDProducts()
	if err != nil {
		logger.WithError(err).Error("Failed to get CD products")
		return nil, fmt.Errorf("failed to get CD products: %w", err)
	}
	
	// 转换为响应格式
	var dtoProducts []*dto.CDProduct
	for _, product := range products {
		dtoProduct := &dto.CDProduct{
			ID:                product.ID,
			ProductMark:       product.ProductMark,
			ProductName:       product.ProductName,
			RequiresRegion:    product.RequiresRegion,
			RequiresPin:       product.RequiresPin,
			CardFormat:        product.CardFormat,
			CardLengthMin:     product.CardLengthMin,
			CardLengthMax:     product.CardLengthMax,
			PinLength:         &product.PinLength,
			ValidationPattern: product.ValidationPattern,
			SupportsAutoType:  product.SupportsAutoType,
			Status:            product.Status,
		}
		dtoProducts = append(dtoProducts, dtoProduct)
	}
	
	response := &dto.CDProductsResponse{
		Products: dtoProducts,
		Total:    len(dtoProducts),
	}
	
	logger.WithFields(map[string]interface{}{
		"products_count": len(dtoProducts),
	}).Info("Retrieved CD products")
	
	return response, nil
}

// GetCDRegions 获取CD区域列表
func (s *service) GetCDRegions(productMark string) (*dto.CDRegionsResponse, error) {
	var productMarkPtr *string
	var productMarkForResponse string
	
	// 如果指定了产品标识，先验证产品是否存在
	if productMark != "" {
		product, err := s.repository.GetCDProductByMark(productMark)
		if err != nil {
			logger.WithError(err).WithField("product_mark", productMark).Error("Failed to get CD product by mark")
			return nil, fmt.Errorf("failed to get CD product by mark: %w", err)
		}
		if product == nil {
			return nil, fmt.Errorf("product not found with mark: %s", productMark)
		}
		productMarkPtr = &productMark
		productMarkForResponse = product.ProductMark
	}
	
	regions, err := s.repository.GetCDRegions(productMarkPtr)
	if err != nil {
		logger.WithError(err).WithField("product_mark", productMark).Error("Failed to get CD regions")
		return nil, fmt.Errorf("failed to get CD regions: %w", err)
	}
	
	// 转换为响应格式
	var dtoRegions []*dto.CDRegion
	for _, region := range regions {
		dtoRegion := &dto.CDRegion{
			ID:            region.ID,
			ProductMark:   region.ProductMark,
			RegionID:      region.RegionID,
			RegionName:    region.RegionName,
			RegionNameEn:  region.RegionNameEn,
			Status:        region.Status,
			SortOrder:     region.SortOrder,
		}
		dtoRegions = append(dtoRegions, dtoRegion)
	}
	
	response := &dto.CDRegionsResponse{
		Regions:     dtoRegions,
		Total:       len(dtoRegions),
		ProductMark: productMarkForResponse,
	}
	
	logger.WithFields(map[string]interface{}{
		"regions_count": len(dtoRegions),
		"product_mark":  productMark,
	}).Info("Retrieved CD regions")
	
	return response, nil
}