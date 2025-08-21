package images

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"trusioo_api/internal/common"
	"trusioo_api/internal/images/dto"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// 获取当前用户ID的辅助函数
func getUserID(c *gin.Context) (int, error) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("user not authenticated")
	}
	
	userID, ok := userIDValue.(int)
	if !ok {
		return 0, fmt.Errorf("invalid user ID format")
	}
	
	return userID, nil
}

// 检查用户类型的辅助函数
func getUserType(c *gin.Context) string {
	if userType, exists := c.Get("user_type"); exists {
		if ut, ok := userType.(string); ok {
			return ut
		}
	}
	return ""
}

// 用户上传图片 - 需要认证
func (h *Handler) UploadImage(c *gin.Context) {
	var req dto.UploadImageRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request parameters",
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:   "NO_FILE",
			Message: "No file provided",
		})
		return
	}

	// 获取当前用户ID
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "User authentication required",
		})
		return
	}

	result, err := h.service.UploadImage(c.Request.Context(), &userID, file, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse{
			Error:   "UPLOAD_FAILED",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, common.SuccessResponse{
		Message: "Image uploaded successfully",
		Data:    result,
	})
}

// 用户查看自己的图片
func (h *Handler) GetImage(c *gin.Context) {
	idParam := c.Param("id")
	imageID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid image ID",
		})
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "User authentication required",
		})
		return
	}

	result, err := h.service.GetUserImage(c.Request.Context(), userID, imageID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, common.ErrorResponse{
				Error:   "IMAGE_NOT_FOUND",
				Message: "Image not found or access denied",
			})
		} else {
			c.JSON(http.StatusInternalServerError, common.ErrorResponse{
				Error:   "GET_FAILED",
				Message: err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse{
		Data: result,
	})
}

// 通过key获取公开图片 - 无需认证
func (h *Handler) GetImageByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:   "INVALID_KEY",
			Message: "Invalid image key",
		})
		return
	}

	result, err := h.service.GetPublicImageByKey(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse{
			Error:   "IMAGE_NOT_FOUND",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse{
		Data: result,
	})
}

// 用户列出自己的图片
func (h *Handler) ListImages(c *gin.Context) {
	var req dto.ListImagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request parameters",
		})
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "User authentication required",
		})
		return
	}

	result, err := h.service.ListImages(c.Request.Context(), &userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse{
			Error:   "LIST_FAILED",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse{
		Data: result,
	})
}

// 用户删除自己的图片
func (h *Handler) DeleteImage(c *gin.Context) {
	idParam := c.Param("id")
	imageID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid image ID",
		})
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "User authentication required",
		})
		return
	}

	err = h.service.DeleteUserImage(c.Request.Context(), userID, imageID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, common.ErrorResponse{
				Error:   "IMAGE_NOT_FOUND",
				Message: "Image not found or access denied",
			})
		} else {
			c.JSON(http.StatusInternalServerError, common.ErrorResponse{
				Error:   "DELETE_FAILED",
				Message: err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse{
		Message: "Image deleted successfully",
	})
}

// 用户刷新自己图片的URL
func (h *Handler) RefreshURL(c *gin.Context) {
	idParam := c.Param("id")
	imageID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid image ID",
		})
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "User authentication required",
		})
		return
	}

	result, err := h.service.RefreshUserImageURL(c.Request.Context(), userID, imageID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, common.ErrorResponse{
				Error:   "IMAGE_NOT_FOUND",
				Message: "Image not found or access denied",
			})
		} else {
			c.JSON(http.StatusInternalServerError, common.ErrorResponse{
				Error:   "REFRESH_FAILED",
				Message: err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse{
		Message: "URL refreshed successfully",
		Data:    result,
	})
}

// ================== 管理员专用接口 ==================

// 管理员查看所有图片
func (h *Handler) AdminListImages(c *gin.Context) {
	var req dto.ListImagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request parameters",
		})
		return
	}

	// 管理员可以查看所有用户的图片
	result, err := h.service.AdminListAllImages(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse{
			Error:   "LIST_FAILED",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse{
		Data: result,
	})
}

// 管理员查看任意图片
func (h *Handler) AdminGetImage(c *gin.Context) {
	idParam := c.Param("id")
	imageID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid image ID",
		})
		return
	}

	result, err := h.service.AdminGetAnyImage(c.Request.Context(), imageID)
	if err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse{
			Error:   "IMAGE_NOT_FOUND",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse{
		Data: result,
	})
}

// 管理员删除任意图片
func (h *Handler) AdminDeleteImage(c *gin.Context) {
	idParam := c.Param("id")
	imageID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid image ID",
		})
		return
	}

	err = h.service.AdminDeleteAnyImage(c.Request.Context(), imageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse{
			Error:   "DELETE_FAILED",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse{
		Message: "Image deleted successfully by admin",
	})
}

// 管理员批量删除图片
func (h *Handler) AdminBatchDeleteImages(c *gin.Context) {
	type BatchDeleteRequest struct {
		ImageIDs []int `json:"image_ids" binding:"required"`
	}

	var req BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request parameters",
		})
		return
	}

	if len(req.ImageIDs) == 0 {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:   "EMPTY_LIST",
			Message: "No image IDs provided",
		})
		return
	}

	deletedCount, err := h.service.AdminBatchDeleteImages(c.Request.Context(), req.ImageIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse{
			Error:   "BATCH_DELETE_FAILED",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse{
		Message: fmt.Sprintf("Successfully deleted %d images", deletedCount),
		Data: map[string]int{
			"deleted_count": deletedCount,
		},
	})
}