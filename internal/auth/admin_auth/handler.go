package admin

import (
	"strconv"

	"trusioo_api/internal/auth/admin_auth/dto"
	"trusioo_api/internal/common"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Login 管理员登录第一步
// @Summary 管理员登录第一步
// @Description 验证管理员邮箱和密码，成功后发送登录验证码
// @Tags 管理员
// @Accept json
// @Produce json
// @Param request body AdminLoginRequest true "登录请求参数"
// @Success 200 {object} common.Response{data=dto.AdminLoginCodeResponse} "验证码发送成功"
// @Failure 400 {object} common.Response "参数错误或登录失败"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/admin/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, err.Error())
		return
	}

	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	resp, err := h.service.Login(&req, clientIP, userAgent)
	if err != nil {
		switch err {
		case common.ErrAdminNotFound:
			common.ValidationError(c, "Admin does not exist")
		case common.ErrInvalidAdminCredentials:
			common.ValidationError(c, "Email or password is incorrect")
		case common.ErrAdminInactive:
			common.ValidationError(c, "Account not activated")
		default:
			common.ServerError(c, err)
		}
		return
	}

	common.Success(c, resp)
}

// LoginVerify 管理员登录第二步
// @Summary 管理员登录第二步
// @Description 验证管理员登录验证码并返回访问令牌
// @Tags 管理员
// @Accept json
// @Produce json
// @Param request body AdminLoginVerifyRequest true "登录验证请求参数"
// @Success 200 {object} common.Response{data=dto.AdminLoginResponse} "登录成功"
// @Failure 400 {object} common.Response "参数错误或验证失败"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/admin/auth/login/verify [post]
func (h *Handler) LoginVerify(c *gin.Context) {
	var req dto.AdminLoginVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, err.Error())
		return
	}

	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	resp, err := h.service.LoginVerify(&req, clientIP, userAgent)
	if err != nil {
		switch err {
		case common.ErrAdminNotFound:
			common.ValidationError(c, "Admin does not exist")
		case common.ErrInvalidCode:
			common.ValidationError(c, "Invalid or expired verification code")
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
// @Tags 管理员
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌请求参数"
// @Success 200 {object} common.Response{data=dto.AdminLoginResponse} "刷新成功"
// @Failure 400 {object} common.Response "参数错误或令牌无效"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/admin/auth/refresh [post]
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
		case common.ErrAdminNotFound:
			common.ValidationError(c, "Admin not found")
		case common.ErrAdminInactive:
			common.ValidationError(c, "Account not activated")
		default:
			common.ServerError(c, err)
		}
		return
	}

	common.Success(c, resp)
}

// GetProfile 获取管理员资料
// @Summary 获取管理员资料
// @Description 获取当前登录管理员的个人资料
// @Tags 管理员
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} common.Response{data=entities.Admin} "获取成功"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/admin/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "Admin not authenticated")
		return
	}

	admin, err := h.service.GetAdminByID(adminID.(int64))
	if err != nil {
		switch err {
		case common.ErrAdminNotFound:
			common.NotFound(c, "Admin not found")
		default:
			common.ServerError(c, err)
		}
		return
	}

	common.Success(c, admin)
}

// GetUserStats 获取用户统计
// @Summary 获取用户统计
// @Description 获取用户统计数据
// @Tags 管理员-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} common.Response{data=dto.UserStats} "获取成功"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/admin/users/stats [get]
func (h *Handler) GetUserStats(c *gin.Context) {
	stats, err := h.service.GetUserStats()
	if err != nil {
		common.ServerError(c, err)
		return
	}

	common.Success(c, stats)
}

// GetUserList 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表（支持分页和筛选）
// @Tags 管理员-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Param status query string false "用户状态筛选" Enums(active, inactive, all)
// @Param email query string false "邮箱筛选"
// @Param phone query string false "手机号筛选"
// @Success 200 {object} common.Response{data=dto.UserListResponse} "获取成功"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/admin/users [get]
func (h *Handler) GetUserList(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.ValidationError(c, err.Error())
		return
	}

	resp, err := h.service.GetUserList(&req)
	if err != nil {
		common.ServerError(c, err)
		return
	}

	common.Success(c, resp)
}

// GetUserDetail 获取用户详情
// @Summary 获取用户详情
// @Description 根据用户ID获取用户详细信息
// @Tags 管理员-用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "用户ID"
// @Success 200 {object} common.Response{data=entities.UserInfo} "获取成功"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 404 {object} common.Response "用户不存在"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/admin/users/{id} [get]
func (h *Handler) GetUserDetail(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		common.ValidationError(c, "Invalid user ID")
		return
	}

	user, err := h.service.GetUserByID(userID)
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

// ForgotPassword 管理员忘记密码第一步 - 发送重置密码验证码
// @Summary 管理员忘记密码第一步
// @Description 发送重置密码验证码到管理员邮箱
// @Tags 管理员
// @Accept json
// @Produce json
// @Param request body AdminForgotPasswordRequest true "忘记密码请求参数"
// @Success 200 {object} common.Response{data=dto.AdminForgotPasswordResponse} "验证码发送成功"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/admin/auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req dto.AdminForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, err.Error())
		return
	}

	resp, err := h.service.ForgotPassword(&req)
	if err != nil {
		common.ServerError(c, err)
		return
	}

	common.Success(c, resp)
}

// ResetPassword 管理员忘记密码第二步 - 验证验证码并重置密码
// @Summary 管理员忘记密码第二步
// @Description 验证重置密码验证码并设置新密码
// @Tags 管理员
// @Accept json
// @Produce json
// @Param request body AdminResetPasswordRequest true "重置密码请求参数"
// @Success 200 {object} common.Response{data=dto.AdminResetPasswordResponse} "密码重置成功"
// @Failure 400 {object} common.Response "参数错误或验证失败"
// @Failure 500 {object} common.Response "服务器错误"
// @Router /api/v1/admin/auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req dto.AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, err.Error())
		return
	}

	resp, err := h.service.ResetPassword(&req)
	if err != nil {
		switch err {
		case common.ErrAdminNotFound:
			common.ValidationError(c, "Admin not found")
		case common.ErrInvalidCode:
			common.ValidationError(c, "Invalid or expired verification code")
		default:
			common.ServerError(c, err)
		}
		return
	}

	common.Success(c, resp)
}