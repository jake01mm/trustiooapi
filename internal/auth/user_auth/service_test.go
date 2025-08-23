package user_auth

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"trusioo_api/internal/auth/user_auth/dto"
	"trusioo_api/internal/auth/user_auth/entities"
	verificationDto "trusioo_api/internal/auth/verification/dto"
	"trusioo_api/internal/common"
	"trusioo_api/internal/testutil"
	"trusioo_api/pkg/ipinfo"
)

// MockUserRepository 模拟用户仓库
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByEmail(email string) (*entities.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id int64) (*entities.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByPhone(phone string) (*entities.User, error) {
	args := m.Called(phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Create(user *entities.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(user *entities.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(id int64, password string) error {
	args := m.Called(id, password)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateEmailVerified(id int64, verified bool) error {
	args := m.Called(id, verified)
	return args.Error(0)
}

func (m *MockUserRepository) CreateRefreshToken(token *entities.RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockUserRepository) GetValidRefreshToken(token string) (*entities.RefreshToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.RefreshToken), args.Error(1)
}

func (m *MockUserRepository) InvalidateRefreshToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockUserRepository) InvalidateAllRefreshTokens(userID int64) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserRepository) CreateLoginSession(session *entities.LoginSession) error {
	args := m.Called(session)
	return args.Error(0)
}

// MockVerificationService 模拟验证服务，实现VerificationService接口
type MockVerificationService struct {
	mock.Mock
}

func (m *MockVerificationService) SendVerificationCode(req *verificationDto.SendVerificationRequest) (*verificationDto.SendVerificationResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*verificationDto.SendVerificationResponse), args.Error(1)
}

func (m *MockVerificationService) VerifyCode(req *verificationDto.VerifyCodeRequest) (*verificationDto.VerifyCodeResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*verificationDto.VerifyCodeResponse), args.Error(1)
}

// MockIPInfoClient 实现 ipinfo.Client 接口的 mock
type MockIPInfoClient struct {
	mock.Mock
}

func (m *MockIPInfoClient) GetIPInfo(ctx context.Context, ip string) (*ipinfo.IPInfo, error) {
	args := m.Called(ctx, ip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ipinfo.IPInfo), args.Error(1)
}

func (m *MockIPInfoClient) BatchGetIPInfo(ctx context.Context, ips []string) (map[string]*ipinfo.IPInfo, error) {
	args := m.Called(ctx, ips)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*ipinfo.IPInfo), args.Error(1)
}

func (m *MockIPInfoClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockIPInfoClient) GetMyIP(ctx context.Context) (*ipinfo.IPInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ipinfo.IPInfo), args.Error(1)
}

// 测试用户注册功能
func TestService_Register(t *testing.T) {
	testutil.MockJWTConfig()

	tests := []struct {
		name           string
		request        *dto.RegisterRequest
		mockSetup      func(*MockUserRepository, *MockVerificationService, *MockIPInfoClient)
		expectedError  error
		validateResult func(*testing.T, *dto.RegisterResponse)
	}{
		{
			name: "成功注册",
			request: &dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(userRepo *MockUserRepository, verifyService *MockVerificationService, ipinfoClient *MockIPInfoClient) {
				// 用户不存在
				userRepo.On("GetByEmail", "test@example.com").Return(nil, sql.ErrNoRows)
				// 创建用户成功
				userRepo.On("Create", mock.AnythingOfType("*entities.User")).Return(nil)
				// 注册流程不需要发送验证码
			},
			expectedError: nil,
			validateResult: func(t *testing.T, resp *dto.RegisterResponse) {
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.User)
				assert.Equal(t, "test@example.com", resp.User.Email)
			},
		},
		{
			name: "用户已存在",
			request: &dto.RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockSetup: func(userRepo *MockUserRepository, verifyService *MockVerificationService, ipinfoClient *MockIPInfoClient) {
				// 用户已存在
				userRepo.On("GetByEmail", "existing@example.com").Return(&entities.User{
					ID:    1,
					Email: "existing@example.com",
				}, nil)
			},
			expectedError: common.ErrEmailExists,
			validateResult: func(t *testing.T, resp *dto.RegisterResponse) {
				assert.Nil(t, resp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &MockUserRepository{}
			verifyService := &MockVerificationService{}
			ipinfoClient := &MockIPInfoClient{}

			tt.mockSetup(userRepo, verifyService, ipinfoClient)

			service := &Service{
				repo:                userRepo,
				verificationService: verifyService,
				ipinfoClient:        ipinfoClient,
			}

			resp, err := service.Register(tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validateResult(t, resp)
			userRepo.AssertExpectations(t)
			verifyService.AssertExpectations(t)
		})
	}
}

// 测试用户登录功能
func TestService_Login(t *testing.T) {
	testutil.MockJWTConfig()

	// 创建测试用户密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		request        *dto.LoginRequest
		mockSetup      func(*MockUserRepository, *MockVerificationService, *MockIPInfoClient)
		expectedError  error
		validateResult func(*testing.T, *dto.LoginCodeResponse)
	}{
		{
			name: "成功登录",
			request: &dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(userRepo *MockUserRepository, verifyService *MockVerificationService, ipinfoClient *MockIPInfoClient) {
				// 用户存在且密码正确
				userRepo.On("GetByEmail", "test@example.com").Return(&entities.User{
					ID:       1,
					Email:    "test@example.com",
					Password: string(hashedPassword),
					Status:   "active",
				}, nil)
				// 发送验证码成功
				verifyService.On("SendVerificationCode", mock.MatchedBy(func(req *verificationDto.SendVerificationRequest) bool {
					return req.Target == "test@example.com" && req.Type == "user_login"
				})).Return(
					&verificationDto.SendVerificationResponse{
						Message: "验证码已发送",
					}, nil)
				// 注意：Login方法只是发送验证码，不会记录登录会话
				// 只有在LoginVerify中验证成功后才会记录登录会话
			},
			expectedError: nil,
			validateResult: func(t *testing.T, resp *dto.LoginCodeResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, "用户登录验证码已发送", resp.Message)
				assert.Equal(t, 600, resp.ExpiresIn)
			},
		},
		{
			name: "用户不存在",
			request: &dto.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			mockSetup: func(userRepo *MockUserRepository, verifyService *MockVerificationService, ipinfoClient *MockIPInfoClient) {
				// 用户不存在
				userRepo.On("GetByEmail", "nonexistent@example.com").Return(nil, sql.ErrNoRows)
				// 记录登录会话
				userRepo.On("CreateLoginSession", mock.AnythingOfType("*entities.LoginSession")).Return(nil)
				// Mock IP信息获取（即使登录失败也会调用）
				ipinfoClient.On("GetIPInfo", mock.Anything, mock.Anything).Return(&ipinfo.IPInfo{
					IP:      "127.0.0.1",
					City:    "Test City",
					Region:  "Test Region",
					Country: "Test Country",
					Loc:     "0,0",
				}, nil)
			},
			expectedError: common.ErrUserNotFound,
			validateResult: func(t *testing.T, resp *dto.LoginCodeResponse) {
				assert.Nil(t, resp)
			},
		},
		{
			name: "密码错误",
			request: &dto.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(userRepo *MockUserRepository, verifyService *MockVerificationService, ipinfoClient *MockIPInfoClient) {
				// 用户存在但密码错误
				userRepo.On("GetByEmail", "test@example.com").Return(&entities.User{
					ID:       1,
					Email:    "test@example.com",
					Password: "different_hashed_password",
					Status:   "active",
				}, nil)
				// 记录登录会话
				userRepo.On("CreateLoginSession", mock.AnythingOfType("*entities.LoginSession")).Return(nil)
				// Mock IP信息获取（即使登录失败也会调用）
				ipinfoClient.On("GetIPInfo", mock.Anything, mock.Anything).Return(&ipinfo.IPInfo{
					IP:      "127.0.0.1",
					City:    "Test City",
					Region:  "Test Region",
					Country: "Test Country",
					Loc:     "0,0",
				}, nil)
			},
			expectedError: common.ErrInvalidCredentials,
			validateResult: func(t *testing.T, resp *dto.LoginCodeResponse) {
				assert.Nil(t, resp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &MockUserRepository{}
			verifyService := &MockVerificationService{}
			ipinfoClient := &MockIPInfoClient{}

			tt.mockSetup(userRepo, verifyService, ipinfoClient)

			service := &Service{
				repo:                userRepo,
				verificationService: verifyService,
				ipinfoClient:        ipinfoClient,
			}

			resp, err := service.Login(tt.request, "127.0.0.1", "test-agent")

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validateResult(t, resp)
			userRepo.AssertExpectations(t)
			verifyService.AssertExpectations(t)
		})
	}
}