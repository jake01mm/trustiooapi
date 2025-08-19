package verification

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	verification := router.Group("/verification")
	{
		// 公开路由 - 不需要认证
		verification.POST("/send", handler.SendVerificationCode)
		verification.POST("/verify", handler.VerifyCode)
	}
}
