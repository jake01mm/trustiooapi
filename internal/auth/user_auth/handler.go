package user_auth

import (
	"trusioo_api/internal/auth/user_auth/dto"
	"trusioo_api/internal/common"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册请求参数"
// @Success 200 {object} common.Response{data=RegisterResponse} "注册成功"
// @Failure 400 {object} common.Response "参数错误或邮箱已存在"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, err.Error())
		return
	}

	resp, err := h.service.Register(&req)
	if err != nil {
		switch err {
		case common.ErrEmailExists:
			common.ValidationError(c, "Email already registered")
		case common.ErrPhoneExists:
			common.ValidationError(c, "Phone number already registered")
		default:
			common.ServerError(c, err)
		}
		return
	}

	common.Success(c, resp)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户使用邮箱和密码登录系统
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求参数"
// @Success 200 {object} common.Response{data=LoginResponse} "登录成功"
// @Failure 400 {object} common.Response "参数错误或登录失败"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, err.Error())
		return
	}

	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	resp, err := h.service.Login(&req, clientIP, userAgent)
	if err != nil {
		switch err {
		case common.ErrUserNotFound:
			common.ValidationError(c, "User does not exist")
		case common.ErrInvalidCredentials:
			common.ValidationError(c, "Email or password is incorrect")
		case common.ErrEmailNotVerified:
			common.ValidationError(c, "Email not verified")
		case common.ErrUserInactive:
			common.ValidationError(c, "Account not activated")
		default:
			common.ServerError(c, err)
		}
		return
	}

	common.Success(c, resp)
}

// RefreshToken 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌请求参数"
// @Success 200 {object} common.Response{data=LoginResponse} "刷新成功"
// @Failure 400 {object} common.Response "参数错误或令牌无效"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, err.Error())
		return
	}

	resp, err := h.service.RefreshToken(&req)
	if err != nil {
		switch err {
		case common.ErrTokenInvalid, common.ErrRefreshTokenInvalid:
			common.Unauthorized(c, "Invalid refresh token")
		case common.ErrUserNotFound:
			common.ValidationError(c, "User not found")
		case common.ErrUserInactive:
			common.ValidationError(c, "Account not activated")
		default:
			common.ServerError(c, err)
		}
		return
	}

	common.Success(c, resp)
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取当前登录用户的个人资料
// @Tags 认证
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} common.Response{data=User} "获取成功"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/auth/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "User not authenticated")
		return
	}

	user, err := h.service.GetUserByID(userID.(int64))
	if err != nil {
		switch err {
		case common.ErrUserNotFound:
			common.NotFound(c, "User not found")
		default:
			common.ServerError(c, err)
		}
		return
	}

	common.Success(c, user)
}