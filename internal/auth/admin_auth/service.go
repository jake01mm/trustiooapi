package admin

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"trusioo_api/config"
	"trusioo_api/internal/auth/admin_auth/dto"
	"trusioo_api/internal/auth/admin_auth/entities"
	"trusioo_api/internal/auth/verification"
	verificationDto "trusioo_api/internal/auth/verification/dto"
	"trusioo_api/internal/common"
	"trusioo_api/pkg/auth"
	"trusioo_api/pkg/ipinfo"

	"golang.org/x/crypto/bcrypt"
)

// Service 管理员业务逻辑服务
type Service struct {
	adminRepo           AdminRepository
	verificationService *verification.Service
	ipinfoClient        ipinfo.Client
}

// NewService 创建新的Service实例
func NewService() *Service {
	ipinfoConfig := ipinfo.LoadConfigFromEnv()
	ipinfoClient := ipinfo.NewClient(ipinfoConfig)

	return &Service{
		adminRepo:           NewAdminRepository(),
		verificationService: verification.NewService(),
		ipinfoClient:        ipinfoClient,
	}
}

// Login 管理员登录第一步 - 验证email+password并发送登录验证码
func (s *Service) Login(req *dto.AdminLoginRequest, clientIP, userAgent string) (*dto.AdminLoginCodeResponse, error) {
	// 1. 验证email+password
	admin, err := s.adminRepo.GetByEmail(req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			s.recordLoginSession(0, clientIP, userAgent, "failed", "管理员不存在")
			return nil, common.ErrAdminNotFound
		}
		return nil, err
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password))
	if err != nil {
		s.recordLoginSession(admin.ID, clientIP, userAgent, "failed", "密码错误")
		return nil, common.ErrInvalidAdminCredentials
	}

	// 检查管理员状态 - 必须是激活状态才能登录
	if admin.Status != "active" {
		s.recordLoginSession(admin.ID, clientIP, userAgent, "failed", "管理员账户未激活")
		return nil, common.ErrAdminInactive
	}

	// 2. 发送登录验证码
	sendReq := &verificationDto.SendVerificationRequest{
		Target: req.Email,
		Type:   req.VerificationType(),
	}
	_, err = s.verificationService.SendVerificationCode(sendReq)
	if err != nil {
		return nil, err
	}

	return &dto.AdminLoginCodeResponse{
		Message:   "管理员登录验证码已发送",
		LoginCode: "已发送到邮箱",
		ExpiresIn: 600, // 10分钟
	}, nil
}

// LoginVerify 管理员登录第二步 - 验证登录验证码并返回token
func (s *Service) LoginVerify(req *dto.AdminLoginVerifyRequest, clientIP, userAgent string) (*dto.AdminLoginResponse, error) {
	// 1. 获取管理员信息
	admin, err := s.adminRepo.GetByEmail(req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrAdminNotFound
		}
		return nil, err
	}

	// 2. 验证验证码
	verifyReq := &verificationDto.VerifyCodeRequest{
		Target: req.Email,
		Code:   req.Code,
		Type:   req.VerificationType(),
	}
	verifyResp, err := s.verificationService.VerifyCode(verifyReq)
	if err != nil {
		s.recordLoginSession(admin.ID, clientIP, userAgent, "failed", "验证码错误")
		return nil, err
	}
	if !verifyResp.Valid {
		s.recordLoginSession(admin.ID, clientIP, userAgent, "failed", "验证码无效")
		return nil, common.ErrInvalidCode
	}

	// 3. 生成访问令牌
	accessToken, err := auth.GenerateAccessToken(admin.ID, admin.Email, admin.Role, "admin")
	if err != nil {
		return nil, err
	}

	// 4. 生成刷新令牌
	refreshTokenStr, err := auth.GenerateRefreshToken(admin.ID, admin.Email, admin.Role, "admin")
	if err != nil {
		return nil, err
	}

	// 5. 保存刷新令牌到数据库
	refreshToken := &entities.AdminRefreshToken{
		AdminID:    admin.ID,
		Token:      refreshTokenStr,
		IsValid:    true,
		ExpiresAt:  time.Now().Add(time.Duration(config.AppConfig.JWT.RefreshExpire) * time.Second),
		DeviceInfo: userAgent,
		CreatedAt:  time.Now(),
	}

	err = s.adminRepo.CreateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 6. 更新最后登录时间
	err = s.adminRepo.UpdateLastLogin(admin.ID)
	if err != nil {
		return nil, err
	}

	// 7. 标记邮箱为已验证
	err = s.adminRepo.UpdateEmailVerified(admin.ID, true)
	if err != nil {
		return nil, err
	}

	// 7. 记录登录会话并获取位置信息
	sessionInfo := s.recordLoginSessionWithIPInfo(admin.ID, clientIP, userAgent, "success", "登录成功")

	return &dto.AdminLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		ExpiresIn:    int64(config.AppConfig.JWT.AccessExpire),
		TokenType:    "Bearer",
		Admin:        *admin,
		LoginSession: sessionInfo,
	}, nil
}

// RefreshToken 刷新访问令牌
func (s *Service) RefreshToken(req *dto.RefreshTokenRequest) (*dto.AdminLoginResponse, error) {
	// 验证刷新令牌
	claims, err := auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, common.ErrTokenInvalid
	}

	// 检查刷新令牌是否在数据库中且有效
	refreshToken, err := s.adminRepo.GetValidRefreshToken(req.RefreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrRefreshTokenInvalid
		}
		return nil, err
	}

	// 验证token所属用户ID
	if refreshToken.AdminID != claims.UserID {
		return nil, common.ErrRefreshTokenInvalid
	}

	// 获取管理员信息
	admin, err := s.adminRepo.GetByID(claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrAdminNotFound
		}
		return nil, err
	}

	// 检查管理员状态
	if admin.Status != "active" {
		return nil, common.ErrAdminInactive
	}

	// 生成新的访问令牌
	accessToken, err := auth.GenerateAccessToken(admin.ID, admin.Email, admin.Role, "admin")
	if err != nil {
		return nil, err
	}

	return &dto.AdminLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
		ExpiresIn:    int64(config.AppConfig.JWT.AccessExpire),
		TokenType:    "Bearer",
		Admin:        *admin,
	}, nil
}

// GetAdminByID 根据ID获取管理员信息
func (s *Service) GetAdminByID(adminID int64) (*entities.Admin, error) {
	admin, err := s.adminRepo.GetByID(adminID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrAdminNotFound
		}
		return nil, err
	}
	return admin, nil
}

// GetUserStats 获取用户统计
func (s *Service) GetUserStats() (*dto.UserStats, error) {
	return s.adminRepo.GetUserStats()
}

// GetUserList 获取用户列表
func (s *Service) GetUserList(req *dto.UserListRequest) (*dto.UserListResponse, error) {
	return s.adminRepo.GetUserList(req)
}

// GetUserByID 根据ID获取用户信息
func (s *Service) GetUserByID(userID int64) (*entities.UserInfo, error) {
	user, err := s.adminRepo.GetUserByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// recordLoginSession 记录登录会话
func (s *Service) recordLoginSession(adminID int64, ip, userAgent, status, reason string) {
	s.recordLoginSessionWithIPInfo(adminID, ip, userAgent, status, reason)
}

func (s *Service) recordLoginSessionWithIPInfo(adminID int64, ip, userAgent, status, reason string) *dto.AdminLoginSessionInfo {
	session := &entities.AdminLoginSession{
		AdminID:   adminID,
		IP:        ip,
		UserAgent: userAgent,
		Status:    status,
		Reason:    reason,
		CreatedAt: time.Now(),
	}

	// 获取IP地理位置信息
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if ipInfo, err := s.ipinfoClient.GetIPInfo(ctx, ip); err == nil {
		session.Country = ipInfo.Country
		session.City = ipInfo.City
		session.Region = ipInfo.Region
		session.Timezone = ipInfo.Timezone
		session.Organization = ipInfo.Org
		session.Location = ipInfo.Loc
	} else {
		log.Printf("获取IP信息失败 %s: %v", ip, err)
	}

	// 解析User-Agent获取设备信息
	s.parseUserAgent(session, userAgent)

	err := s.adminRepo.CreateLoginSession(session)
	if err != nil {
		log.Printf("记录管理员登录会话失败: %v", err)
		return nil
	}

	// 返回会话信息给客户端
	if status == "success" {
		return &dto.AdminLoginSessionInfo{
			IP:           session.IP,
			Country:      session.Country,
			City:         session.City,
			Region:       session.Region,
			Timezone:     session.Timezone,
			Organization: session.Organization,
			Location:     session.Location,
			IsTrusted:    session.IsTrusted,
		}
	}
	return nil
}

func (s *Service) parseUserAgent(session *entities.AdminLoginSession, userAgent string) {
	ua := strings.ToLower(userAgent)

	// 检测设备类型
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		session.DeviceType = "mobile"
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		session.DeviceType = "tablet"
	} else {
		session.DeviceType = "desktop"
	}

	// 检测操作系统
	if strings.Contains(ua, "windows") {
		session.OS = "Windows"
	} else if strings.Contains(ua, "mac") || strings.Contains(ua, "darwin") {
		session.OS = "macOS"
	} else if strings.Contains(ua, "linux") {
		session.OS = "Linux"
	} else if strings.Contains(ua, "android") {
		session.OS = "Android"
	} else if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ios") {
		session.OS = "iOS"
	}

	// 检测浏览器
	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edge") {
		session.Browser = "Chrome"
	} else if strings.Contains(ua, "firefox") {
		session.Browser = "Firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		session.Browser = "Safari"
	} else if strings.Contains(ua, "edge") {
		session.Browser = "Edge"
	} else if strings.Contains(ua, "opera") {
		session.Browser = "Opera"
	}

	// 检测平台
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		session.Platform = "mobile"
	} else {
		session.Platform = "web"
	}
}

// ForgotPassword 管理员忘记密码 - 发送重置密码验证码
func (s *Service) ForgotPassword(req *dto.AdminForgotPasswordRequest) (*dto.AdminForgotPasswordResponse, error) {
	// 1. 验证管理员是否存在
	admin, err := s.adminRepo.GetByEmail(req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// 为了安全，即使管理员不存在也返回成功，避免暴露管理员是否存在
			return &dto.AdminForgotPasswordResponse{
				Message:   "如果该邮箱已注册为管理员，重置密码验证码已发送",
				ExpiresIn: 600,
			}, nil
		}
		return nil, err
	}

	// 2. 检查管理员状态
	if admin.Status != "active" {
		// 为了安全，不暴露账户状态信息
		return &dto.AdminForgotPasswordResponse{
			Message:   "如果该邮箱已注册为管理员，重置密码验证码已发送",
			ExpiresIn: 600,
		}, nil
	}

	// 3. 发送重置密码验证码
	sendReq := &verificationDto.SendVerificationRequest{
		Target: req.Email,
		Type:   "admin_forgot_password",
	}
	_, err = s.verificationService.SendVerificationCode(sendReq)
	if err != nil {
		return nil, err
	}

	return &dto.AdminForgotPasswordResponse{
		Message:   "管理员重置密码验证码已发送到您的邮箱",
		ExpiresIn: 600, // 10分钟
	}, nil
}

// ResetPassword 管理员重置密码 - 验证验证码并重置密码
func (s *Service) ResetPassword(req *dto.AdminResetPasswordRequest) (*dto.AdminResetPasswordResponse, error) {
	// 1. 验证管理员是否存在
	admin, err := s.adminRepo.GetByEmail(req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrAdminNotFound
		}
		return nil, err
	}

	// 2. 验证重置密码验证码
	verifyReq := &verificationDto.VerifyCodeRequest{
		Target: req.Email,
		Code:   req.Code,
		Type:   "admin_forgot_password",
	}
	verifyResp, err := s.verificationService.VerifyCode(verifyReq)
	if err != nil {
		return nil, err
	}
	if !verifyResp.Valid {
		return nil, common.ErrInvalidCode
	}

	// 3. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 4. 更新密码
	err = s.adminRepo.UpdatePassword(admin.ID, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	// 5. 可选：使所有refresh token失效，强制重新登录
	err = s.adminRepo.InvalidateAllRefreshTokens(admin.ID)
	if err != nil {
		log.Printf("Failed to invalidate refresh tokens for admin %d: %v", admin.ID, err)
		// 不返回错误，因为密码已经重置成功
	}

	return &dto.AdminResetPasswordResponse{
		Message: "管理员密码重置成功，请使用新密码登录",
	}, nil
}
