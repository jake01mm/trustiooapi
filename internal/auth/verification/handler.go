package verification

import (
	"trusioo_api/internal/auth/verification/dto"
	"trusioo_api/internal/common"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler() *Handler {
	return &Handler{
		service: NewService(),
	}
}

// SendVerificationCode 发送验证码
func (h *Handler) SendVerificationCode(c *gin.Context) {
	var req dto.SendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, err.Error())
		return
	}

	resp, err := h.service.SendVerificationCode(&req)
	if err != nil {
		common.ServerError(c, err)
		return
	}

	common.SuccessWithMessage(c, "Verification code sent successfully", resp)
}

// VerifyCode 验证验证码
func (h *Handler) VerifyCode(c *gin.Context) {
	var req dto.VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, err.Error())
		return
	}

	resp, err := h.service.VerifyCode(&req)
	if err != nil {
		common.ServerError(c, err)
		return
	}

	if !resp.Valid {
		common.ValidationError(c, resp.Message)
		return
	}

	common.SuccessWithMessage(c, resp.Message, resp)
}
