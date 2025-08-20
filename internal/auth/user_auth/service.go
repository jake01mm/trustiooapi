package user_auth

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"trusioo_api/config"
	"trusioo_api/internal/auth/user_auth/dto"
	"trusioo_api/internal/auth/user_auth/entities"
	"trusioo_api/internal/auth/verification"
	verificationDto "trusioo_api/internal/auth/verification/dto"
	"trusioo_api/internal/common"
	"trusioo_api/pkg/auth"
	"trusioo_api/pkg/ipinfo"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo                Repository
	verificationService *verification.Service
	ipinfoClient        ipinfo.Client
}

func NewService(repo Repository) *Service {
	ipinfoConfig := ipinfo.LoadConfigFromEnv()
	ipinfoClient := ipinfo.NewClient(ipinfoConfig)
	
	return &Service{
		repo:                repo,
		verificationService: verification.NewService(),
		ipinfoClient:        ipinfoClient,
	}
}

func (s *Service) Register(req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// 检查邮箱是否已存在
	_, err := s.repo.GetByEmail(req.Email)
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

	// 创建用户 - 默认未激活状态
	user := &entities.User{
		Name:          "", // 注册时不需要姓名
		Email:         req.Email,
		Password:      string(hashedPassword),
		Phone:         nil, // 注册时不需要手机号
		Role:          "user",
		Status:        "inactive", // 默认未激活
		EmailVerified: false,      // 默认未验证
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

// Login 第一步登录 - 验证email+password并发送登录验证码
func (s *Service) Login(req *dto.LoginRequest, clientIP, userAgent string) (*dto.LoginCodeResponse, error) {
	// 1. 验证email+password
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

	// 检查用户状态 - 必须是激活状态才能登录
	if user.Status != "active" {
		s.recordLoginSession(user.ID, clientIP, userAgent, "email", "failed", "账户未激活")
		return nil, common.ErrUserInactive
	}

	// 2. 发送登录验证码
	sendReq := &verificationDto.SendVerificationRequest{
		Target: req.Email,
		Type:   "login",
	}
	_, err = s.verificationService.SendVerificationCode(sendReq)
	if err != nil {
		return nil, err
	}

	return &dto.LoginCodeResponse{
		Message:   "登录验证码已发送",
		LoginCode: "已发送到邮箱",
		ExpiresIn: 600, // 10分钟
	}, nil
}

// LoginVerify 第二步登录 - 验证登录验证码并返回token
func (s *Service) LoginVerify(req *dto.LoginVerifyRequest, clientIP, userAgent string) (*dto.LoginResponse, error) {
	// 1. 获取用户信息
	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrInvalidCredentials
		}
		return nil, err
	}

	// 2. 验证验证码
	verifyReq := &verificationDto.VerifyCodeRequest{
		Target: req.Email,
		Code:   req.Code,
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

	// 3. 生成访问令牌
	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, user.Role, "user")
	if err != nil {
		return nil, err
	}

	// 4. 生成刷新令牌
	refreshTokenStr, err := auth.GenerateRefreshToken(user.ID, user.Email, user.Role, "user")
	if err != nil {
		return nil, err
	}

	// 5. 保存刷新令牌
	refreshToken := &entities.RefreshToken{
		UserID:     user.ID,
		Token:      refreshTokenStr,
		IsValid:    true,
		ExpiresAt:  time.Now().Add(time.Duration(config.AppConfig.JWT.RefreshExpire) * time.Second),
		DeviceInfo: userAgent,
		CreatedAt:  time.Now(),
	}

	err = s.repo.CreateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 6. 更新最后登录时间
	err = s.repo.UpdateLastLogin(user.ID)
	if err != nil {
		return nil, err
	}

	// 7. 记录登录会话并获取位置信息
	sessionInfo := s.recordLoginSessionWithIPInfo(user.ID, clientIP, userAgent, "email", "success", "登录成功")

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
		LoginSession: sessionInfo,
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
	s.recordLoginSessionWithIPInfo(userID, ip, userAgent, method, status, reason)
}

func (s *Service) recordLoginSessionWithIPInfo(userID int64, ip, userAgent, method, status, reason string) *dto.LoginSessionInfo {
	session := &entities.LoginSession{
		UserID:      userID,
		IP:          ip,
		UserAgent:   userAgent,
		LoginMethod: method,
		Status:      status,
		Reason:      reason,
		CreatedAt:   time.Now(),
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

	err := s.repo.CreateLoginSession(session)
	if err != nil {
		log.Printf("记录登录会话失败: %v", err)
		return nil
	}

	// 返回会话信息给客户端
	if status == "success" {
		return &dto.LoginSessionInfo{
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

func (s *Service) parseUserAgent(session *entities.LoginSession, userAgent string) {
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
