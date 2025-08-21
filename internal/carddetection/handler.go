package carddetection

import (
	"context"
	"net/http"
	"time"

	"trusioo_api/config"
	"trusioo_api/internal/carddetection/dto"
	"trusioo_api/internal/common"
	"trusioo_api/pkg/carddetection"
	"trusioo_api/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Handler 卡片检测处理器
type Handler struct {
	service Service
}

// NewHandler 创建新的处理器
func NewHandler() *Handler {
	// 从应用配置创建客户端
	cfg := carddetection.NewConfigFromApp(config.AppConfig)
	var client *carddetection.Client
	if cfg == nil {
		logger.Warn("Card detection is disabled or not configured")
		client = nil
	} else {
		client = carddetection.NewClient(cfg)
	}

	// 创建repository和service
	repository := NewRepository()
	service := NewService(client, repository)
	
	return &Handler{service: service}
}

// CheckCard 检测卡片
func (h *Handler) CheckCard(c *gin.Context) {
	// 获取用户ID (从JWT token中获取)
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "User not authenticated")
		return
	}

	var req carddetection.CheckCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Failed to bind check card request")
		common.ValidationError(c, err.Error())
		return
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	logger.WithFields(map[string]interface{}{
		"user_id":      userID,
		"product_mark": req.ProductMark,
		"cards_count":  len(req.Cards),
		"region_id":    req.RegionID,
		"region_name":  req.RegionName,
	}).Info("Processing card detection request")

	// 执行卡片检测
	resp, err := h.service.CheckCard(ctx, userID.(int64), &req)
	if err != nil {
		logger.WithError(err).Error("Card detection failed")
		if carddetection.IsCardDetectionError(err) {
			cdErr := err.(*carddetection.CardDetectionError)
			c.JSON(http.StatusBadRequest, common.Response{
				Code:    cdErr.Code,
				Message: cdErr.Message,
			})
		} else {
			common.ServerError(c, err)
		}
		return
	}

	logger.WithFields(map[string]interface{}{
		"user_id":       userID,
		"response_code": resp.Code,
		"response_msg":  resp.Msg,
	}).Info("Card detection completed")

	common.Success(c, resp)
}

// CheckCardResult 查询卡片检测结果
func (h *Handler) CheckCardResult(c *gin.Context) {
	// 获取用户ID (从JWT token中获取)
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "User not authenticated")
		return
	}

	var req carddetection.CheckCardResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Failed to bind check card result request")
		common.ValidationError(c, err.Error())
		return
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	logger.WithFields(map[string]interface{}{
		"user_id":      userID,
		"product_mark": req.ProductMark,
		"card_no":      req.CardNo,
		"has_pin_code": req.PinCode != "",
	}).Info("Processing card result query")

	// 查询卡片检测结果
	result, err := h.service.CheckCardResult(ctx, userID.(int64), &req)
	if err != nil {
		logger.WithError(err).Error("Card result query failed")
		if carddetection.IsCardDetectionError(err) {
			cdErr := err.(*carddetection.CardDetectionError)
			c.JSON(http.StatusBadRequest, common.Response{
				Code:    cdErr.Code,
				Message: cdErr.Message,
			})
		} else {
			common.ServerError(c, err)
		}
		return
	}

	logger.WithFields(map[string]interface{}{
		"user_id":      userID,
		"card_no":      result.CardNo,
		"status":       result.Status,
		"check_time":   result.GetCheckTimeString(),
		"region_name":  result.RegionName,
	}).Info("Card result query completed")

	common.Success(c, result)
}

// GetCDProducts 获取CD产品列表
func (h *Handler) GetCDProducts(c *gin.Context) {
	// 获取用户ID (从JWT token中获取)
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "User not authenticated")
		return
	}
	
	// 记录访问日志
	logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"endpoint": "GetCDProducts",
		"method":   c.Request.Method,
		"path":     c.Request.URL.Path,
	}).Info("CD Products API accessed")
	
	products, err := h.service.GetCDProducts()
	if err != nil {
		logger.WithError(err).WithField("user_id", userID).Error("Failed to get CD products")
		common.ServerError(c, err)
		return
	}
	
	logger.WithFields(map[string]interface{}{
		"user_id":        userID,
		"products_count": len(products.Products),
	}).Info("CD Products retrieved successfully")
	
	common.Success(c, products)
}

// GetCDRegions 获取CD区域列表
func (h *Handler) GetCDRegions(c *gin.Context) {
	// 获取用户ID (从JWT token中获取)
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "User not authenticated")
		return
	}
	
	// 获取可选的产品标识参数
	productMark := c.Query("product_mark")
	
	// 记录访问日志
	logger.WithFields(map[string]interface{}{
		"user_id":      userID,
		"endpoint":     "GetCDRegions",
		"method":       c.Request.Method,
		"path":         c.Request.URL.Path,
		"product_mark": productMark,
	}).Info("CD Regions API accessed")
	
	regions, err := h.service.GetCDRegions(productMark)
	if err != nil {
		logger.WithError(err).WithFields(map[string]interface{}{
			"user_id":      userID,
			"product_mark": productMark,
		}).Error("Failed to get CD regions")
		common.ServerError(c, err)
		return
	}
	
	logger.WithFields(map[string]interface{}{
		"user_id":       userID,
		"product_mark":  productMark,
		"regions_count": len(regions.Regions),
	}).Info("CD Regions retrieved successfully")
	
	common.Success(c, regions)
}

// GetSupportedRegions 获取支持的地区列表（保留用于向后兼容）
func (h *Handler) GetSupportedRegions(c *gin.Context) {
	productMark := c.Query("productMark")
	if productMark == "" {
		common.ValidationError(c, "productMark parameter is required")
		return
	}

	var regions interface{}
	switch carddetection.ProductMark(productMark) {
	case carddetection.ProductMarkItunes:
		regions = carddetection.ITunesRegions
	case carddetection.ProductMarkAmazon:
		regions = carddetection.AmazonRegions
	case carddetection.ProductMarkRazer:
		regions = carddetection.RazerRegions
	case carddetection.ProductMarkXbox:
		regions = carddetection.XboxRegions
	default:
		common.ValidationError(c, "Unsupported product mark")
		return
	}

	common.Success(c, map[string]interface{}{
		"productMark": productMark,
		"regions":     regions,
	})
}

// GetStatus 获取服务状态
func (h *Handler) GetStatus(c *gin.Context) {
	status := h.service.GetServiceStatus()
	
	// 添加配置信息
	if config.AppConfig.ThirdParty.CardDetectionEnabled {
		status["host"] = config.AppConfig.ThirdParty.CardDetectionHost
		status["timeout"] = config.AppConfig.ThirdParty.CardDetectionTimeout
	}

	common.Success(c, status)
}

// GetUserHistory 获取用户检测历史
func (h *Handler) GetUserHistory(c *gin.Context) {
	// 获取用户ID (从JWT token中获取)
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "User not authenticated")
		return
	}

	var req dto.CardDetectionHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.WithError(err).Error("Failed to bind history request")
		common.ValidationError(c, err.Error())
		return
	}

	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	logger.WithFields(map[string]interface{}{
		"user_id":      userID,
		"page":         req.Page,
		"page_size":    req.PageSize,
		"status":       req.Status,
		"product_mark": req.ProductMark,
	}).Info("Getting user detection history")

	resp, err := h.service.GetUserHistory(userID.(int64), &req)
	if err != nil {
		logger.WithError(err).Error("Failed to get user history")
		common.ServerError(c, err)
		return
	}

	common.Success(c, resp)
}

// GetRecordDetail 获取检测记录详情
func (h *Handler) GetRecordDetail(c *gin.Context) {
	// 获取用户ID (从JWT token中获取)
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "User not authenticated")
		return
	}

	var req dto.CardDetectionDetailRequest
	if err := c.ShouldBindUri(&req); err != nil {
		logger.WithError(err).Error("Failed to bind detail request")
		common.ValidationError(c, err.Error())
		return
	}

	logger.WithFields(map[string]interface{}{
		"user_id":   userID,
		"record_id": req.ID,
	}).Info("Getting detection record detail")

	resp, err := h.service.GetRecordDetail(userID.(int64), req.ID)
	if err != nil {
		logger.WithError(err).Error("Failed to get record detail")
		if err.Error() == "record not found" {
			common.NotFound(c, "Detection record not found")
		} else {
			common.ServerError(c, err)
		}
		return
	}

	common.Success(c, resp)
}

// GetUserStats 获取用户检测统计
func (h *Handler) GetUserStats(c *gin.Context) {
	// 获取用户ID (从JWT token中获取)
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "User not authenticated")
		return
	}

	logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Getting user detection stats")

	resp, err := h.service.GetUserStats(userID.(int64))
	if err != nil {
		logger.WithError(err).Error("Failed to get user stats")
		common.ServerError(c, err)
		return
	}

	common.Success(c, resp)
}

// GetUserSummary 获取用户检测汇总
func (h *Handler) GetUserSummary(c *gin.Context) {
	// 获取用户ID (从JWT token中获取)
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "User not authenticated")
		return
	}

	logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Getting user detection summary")

	resp, err := h.service.GetUserSummary(userID.(int64))
	if err != nil {
		logger.WithError(err).Error("Failed to get user summary")
		common.ServerError(c, err)
		return
	}

	common.Success(c, resp)
}