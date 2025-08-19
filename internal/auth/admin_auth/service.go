package admin

import (
	"database/sql"
	"time"

	"trusioo_api/config"
	"trusioo_api/internal/auth/admin_auth/dto"
	"trusioo_api/internal/auth/admin_auth/entities"
	"trusioo_api/internal/auth/verification"
	verificationDto "trusioo_api/internal/auth/verification/dto"
	"trusioo_api/internal/common"
	"trusioo_api/pkg/auth"

	"golang.org/x/crypto/bcrypt"
)

// Service 管理员业务逻辑服务
type Service struct {
	adminRepo           AdminRepository
	verificationService *verification.Service
}

// NewService 创建新的Service实例
func NewService() *Service {
	return &Service{
		adminRepo:           NewAdminRepository(),
		verificationService: verification.NewService(),
	}
}

// Login 管理员登录
func (s *Service) Login(req *dto.AdminLoginRequest, clientIP, userAgent string) (*dto.AdminLoginResponse, error) {
	// 查找管理员
	admin, err := s.adminRepo.GetByEmail(req.Email)
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

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password))
	if err != nil {
		// 记录失败的登录会话
		s.recordLoginSession(admin.ID, clientIP, userAgent, "failed", "密码错误")
		return nil, common.ErrInvalidAdminCredentials
	}

	// 验证验证码
	verifyReq := &verificationDto.VerifyCodeRequest{
		Target: req.Email,
		Code:   req.VerificationCode,
		Type:   "admin_login",
	}
	verifyResp, err := s.verificationService.VerifyCode(verifyReq)
	if err != nil {
		// 记录失败的登录会话
		s.recordLoginSession(admin.ID, clientIP, userAgent, "failed", "验证码验证失败")
		return nil, err
	}
	if !verifyResp.Valid {
		// 记录失败的登录会话
		s.recordLoginSession(admin.ID, clientIP, userAgent, "failed", "验证码无效")
		return nil, common.ErrInvalidCode
	}

	// 生成访问令牌
	accessToken, err := auth.GenerateAccessToken(admin.ID, admin.Email, admin.Role, "admin")
	if err != nil {
		return nil, err
	}

	// 生成刷新令牌
	refreshToken, err := auth.GenerateRefreshToken(admin.ID, admin.Email, admin.Role, "admin")
	if err != nil {
		return nil, err
	}

	// 保存刷新令牌到数据库
	refreshTokenEntity := &entities.AdminRefreshToken{
		AdminID:    admin.ID,
		Token:      refreshToken,
		IsValid:    true,
		ExpiresAt:  time.Now().Add(time.Duration(config.AppConfig.JWT.RefreshExpire) * time.Second),
		DeviceInfo: userAgent,
	}
	
	err = s.adminRepo.CreateRefreshToken(refreshTokenEntity)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	err = s.adminRepo.UpdateLastLogin(admin.ID)
	if err != nil {
		return nil, err
	}

	// 记录成功的登录会话
	s.recordLoginSession(admin.ID, clientIP, userAgent, "success", "")

	return &dto.AdminLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(config.AppConfig.JWT.AccessExpire),
		TokenType:    "Bearer",
		Admin:        *admin,
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
	session := &entities.AdminLoginSession{
		AdminID:   adminID,
		IP:        ip,
		UserAgent: userAgent,
		Status:    status,
		Reason:    reason,
		// 其他字段可以后续解析userAgent等获得
		Country:    "",
		City:       "",
		DeviceType: "",
		OS:         "",
		Browser:    "",
		IsTrusted:  false,
		Platform:   "",
	}

	err := s.adminRepo.CreateLoginSession(session)
	if err != nil {
		// 日志记录错误，但不影响主流程
		// TODO: 使用proper logger
		_ = err
	}
}