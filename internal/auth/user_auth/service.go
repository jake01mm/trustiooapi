package user_auth

import (
	"database/sql"
	"fmt"
	"time"

	"trusioo_api/config"
	"trusioo_api/internal/auth/user_auth/dto"
	"trusioo_api/internal/auth/user_auth/entities"
	"trusioo_api/internal/auth/verification"
	verificationDto "trusioo_api/internal/auth/verification/dto"
	"trusioo_api/internal/common"
	"trusioo_api/pkg/auth"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo                Repository
	verificationService *verification.Service
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:                repo,
		verificationService: verification.NewService(),
	}
}

func (s *Service) Register(req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// 验证验证码
	verifyReq := &verificationDto.VerifyCodeRequest{
		Target: req.Email,
		Code:   req.VerificationCode,
		Type:   "register",
	}
	_, err := s.verificationService.VerifyCode(verifyReq)
	if err != nil {
		return nil, err
	}

	// 检查邮箱是否已存在
	_, err = s.repo.GetByEmail(req.Email)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil {
		return nil, common.ErrEmailExists
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &entities.User{
		Name:          "", // 注册时不需要姓名
		Email:         req.Email,
		Password:      string(hashedPassword),
		Phone:         nil, // 注册时不需要手机号
		Role:          "user",
		Status:        "active",
		EmailVerified: true, // 验证码验证通过后设置为true
		PhoneVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = s.repo.Create(user)
	if err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		User: user,
	}, nil
}

// PreAuth 预验证email+password，验证通过后可发送验证码
func (s *Service) PreAuth(req *dto.PreAuthRequest) (*dto.PreAuthResponse, error) {
	// 根据邮箱获取用户
	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return &dto.PreAuthResponse{
				Message:  "邮箱或密码错误",
				Verified: false,
			}, nil
		}
		return nil, err
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return &dto.PreAuthResponse{
			Message:  "邮箱或密码错误",
			Verified: false,
		}, nil
	}

	// 检查用户状态
	if user.Status != "active" {
		return &dto.PreAuthResponse{
			Message:  "账户已禁用",
			Verified: false,
		}, nil
	}

	return &dto.PreAuthResponse{
		Message:  "预验证成功，可以发送验证码",
		Verified: true,
		Email:    req.Email,
	}, nil
}

func (s *Service) Login(req *dto.LoginRequest, clientIP, userAgent string) (*dto.LoginResponse, error) {
	// 1. 首先验证email+password
	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			s.recordLoginSession(0, clientIP, userAgent, "email", "failed", "用户不存在")
			return nil, common.ErrInvalidCredentials
		}
		return nil, err
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		s.recordLoginSession(user.ID, clientIP, userAgent, "email", "failed", "密码错误")
		return nil, common.ErrInvalidCredentials
	}

	// 检查用户状态
	if user.Status != "active" {
		s.recordLoginSession(user.ID, clientIP, userAgent, "email", "failed", "账户已禁用")
		return nil, common.ErrUserInactive
	}

	// 2. 验证验证码
	verifyReq := &verificationDto.VerifyCodeRequest{
		Target: req.Email,
		Code:   req.VerificationCode,
		Type:   "login",
	}
	verifyResp, err := s.verificationService.VerifyCode(verifyReq)
	if err != nil {
		s.recordLoginSession(user.ID, clientIP, userAgent, "email", "failed", "验证码错误")
		return nil, err
	}
	if !verifyResp.Valid {
		s.recordLoginSession(user.ID, clientIP, userAgent, "email", "failed", "验证码无效")
		return nil, common.ErrInvalidCode
	}

	// 生成访问令牌
	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, user.Role, "user")
	if err != nil {
		return nil, err
	}

	// 生成刷新令牌
	refreshTokenStr, err := auth.GenerateRefreshToken(user.ID, user.Email, user.Role, "user")
	if err != nil {
		return nil, err
	}

	// 保存刷新令牌
	refreshToken := &entities.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: time.Now().Add(time.Duration(config.AppConfig.JWT.RefreshExpire) * time.Second),
		CreatedAt: time.Now(),
	}

	err = s.repo.CreateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	err = s.repo.UpdateLastLogin(user.ID)
	if err != nil {
		return nil, err
	}

	// 记录登录会话
	s.recordLoginSession(user.ID, clientIP, userAgent, "email", "success", "登录成功")

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		TokenType:    "Bearer",
		ExpiresIn:    int64(config.AppConfig.JWT.AccessExpire),
		User: &entities.User{
			ID:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			Phone:  user.Phone,
			Role:   user.Role,
			Status: user.Status,
		},
	}, nil
}

func (s *Service) RefreshToken(req *dto.RefreshTokenRequest) (*dto.LoginResponse, error) {
	// 验证刷新令牌
	refreshToken, err := s.repo.GetValidRefreshToken(req.RefreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrRefreshTokenInvalid
		}
		return nil, err
	}

	// 获取用户信息
	user, err := s.repo.GetByID(refreshToken.UserID)
	if err != nil {
		return nil, err
	}

	// 生成新的访问令牌
	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, user.Role, "user")
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(config.AppConfig.JWT.AccessExpire),
		User: &entities.User{
			ID:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			Phone:  user.Phone,
			Role:   user.Role,
			Status: user.Status,
		},
	}, nil
}

func (s *Service) GetUserByID(userID int64) (*entities.User, error) {
	return s.repo.GetByID(userID)
}

func (s *Service) recordLoginSession(userID int64, ip, userAgent, method, status, reason string) {
	session := &entities.LoginSession{
		UserID:      userID,
		IP:          ip,
		UserAgent:   userAgent,
		LoginMethod: method,
		Status:      status,
		Reason:      reason,
		CreatedAt:   time.Now(),
	}

	err := s.repo.CreateLoginSession(session)
	if err != nil {
		fmt.Printf("记录登录会话失败: %v\n", err)
	}
}
