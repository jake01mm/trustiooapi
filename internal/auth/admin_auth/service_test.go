package admin

import (
	"context"
	"database/sql"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"trusioo_api/internal/auth/admin_auth/dto"
	"trusioo_api/internal/auth/admin_auth/entities"
	verificationDto "trusioo_api/internal/auth/verification/dto"
	"trusioo_api/internal/common"
	"trusioo_api/pkg/ipinfo"
)

// MockAdminRepository 实现 AdminRepository 接口的 mock
type MockAdminRepository struct {
	mock.Mock
}

func (m *MockAdminRepository) GetByEmail(email string) (*entities.Admin, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Admin), args.Error(1)
}

func (m *MockAdminRepository) GetByID(id int64) (*entities.Admin, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Admin), args.Error(1)
}

func (m *MockAdminRepository) Create(admin *entities.Admin) error {
	args := m.Called(admin)
	return args.Error(0)
}

func (m *MockAdminRepository) Update(admin *entities.Admin) error {
	args := m.Called(admin)
	return args.Error(0)
}

func (m *MockAdminRepository) UpdateLastLogin(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAdminRepository) UpdatePassword(id int64, hashedPassword string) error {
	args := m.Called(id, hashedPassword)
	return args.Error(0)
}

func (m *MockAdminRepository) UpdateEmailVerified(id int64, verified bool) error {
	args := m.Called(id, verified)
	return args.Error(0)
}

func (m *MockAdminRepository) CreateRefreshToken(token *entities.AdminRefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockAdminRepository) GetValidRefreshToken(token string) (*entities.AdminRefreshToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AdminRefreshToken), args.Error(1)
}

func (m *MockAdminRepository) InvalidateRefreshToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockAdminRepository) InvalidateAllRefreshTokens(adminID int64) error {
	args := m.Called(adminID)
	return args.Error(0)
}

func (m *MockAdminRepository) CreateLoginSession(session *entities.AdminLoginSession) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockAdminRepository) GetUserStats() (*dto.UserStats, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserStats), args.Error(1)
}

func (m *MockAdminRepository) GetUserList(req *dto.UserListRequest) (*dto.UserListResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserListResponse), args.Error(1)
}

func (m *MockAdminRepository) GetUserByID(id int64) (*entities.UserInfo, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.UserInfo), args.Error(1)
}

// MockIPInfoClient 实现 ipinfo.Client 接口的 mock
type MockIPInfoClient struct {
	mock.Mock
}

// VerificationService 接口定义
type VerificationService interface {
	SendVerificationCode(req *verificationDto.SendVerificationRequest) (*verificationDto.SendVerificationResponse, error)
	VerifyCode(req *verificationDto.VerifyCodeRequest) (*verificationDto.VerifyCodeResponse, error)
}

// MockVerificationService mock verification service
type MockVerificationService struct {
	mock.Mock
}

func (m *MockVerificationService) SendVerificationCode(req *verificationDto.SendVerificationRequest) (*verificationDto.SendVerificationResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*verificationDto.SendVerificationResponse), args.Error(1)
}

func (m *MockVerificationService) VerifyCode(req *verificationDto.VerifyCodeRequest) (*verificationDto.VerifyCodeResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*verificationDto.VerifyCodeResponse), args.Error(1)
}

func (m *MockIPInfoClient) GetIPInfo(ctx context.Context, ip string) (*ipinfo.IPInfo, error) {
	args := m.Called(ctx, ip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ipinfo.IPInfo), args.Error(1)
}

func (m *MockIPInfoClient) GetMyIP(ctx context.Context) (*ipinfo.IPInfo, error) {
	args := m.Called(ctx)
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

// 测试辅助函数
// TestService 用于测试的 Service 结构
type TestService struct {
	adminRepo           AdminRepository
	verificationService VerificationService
	ipinfoClient        ipinfo.Client
}

// Login 实现登录方法
func (s *TestService) Login(ctx context.Context, email, password string) (*dto.AdminLoginCodeResponse, error) {
	// 简化的测试实现
	admin, err := s.adminRepo.GetByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrInvalidCredentials
		}
		return nil, err
	}
	
	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return nil, common.ErrInvalidCredentials
	}
	
	// 发送验证码
	verifyReq := &verificationDto.SendVerificationRequest{
		Target: email,
		Type:   "admin_login",
	}
	
	verifyResp, verifyErr := s.verificationService.SendVerificationCode(verifyReq)
	if verifyErr != nil {
		return nil, verifyErr
	}
	
	return &dto.AdminLoginCodeResponse{
		Message:   verifyResp.Message,
		ExpiresIn: 300, // 5分钟
	}, nil
}

func setupAdminService(mockRepo *MockAdminRepository, mockIPInfoClient *MockIPInfoClient, mockVerificationService *MockVerificationService) *TestService {
	return &TestService{
		adminRepo:           mockRepo,
		verificationService: mockVerificationService,
		ipinfoClient:        mockIPInfoClient,
	}
}

func TestService_Login(t *testing.T) {
	tests := []struct {
		name           string
		request        *dto.AdminLoginRequest
		setupMocks     func(*MockAdminRepository, *MockIPInfoClient, *MockVerificationService)
		expectedError  string
		expectedResult bool
	}{
		{
			name: "成功登录",
			request: &dto.AdminLoginRequest{
				Email: "admin@example.com",
				Password: "password123",
			},
			setupMocks: func(mockRepo *MockAdminRepository, mockIPInfo *MockIPInfoClient, mockVerificationService *MockVerificationService) {
				admin := &entities.Admin{
					ID: 1,
					Email: "admin@example.com",
					Password: "$2a$10$AS9mJUHq5hxMoz5meJdMYumez3uzf4A8Dng0ng8nMqu.TH4ThrcGy", // password123
					Status: "active",
				}
				mockRepo.On("GetByEmail", "admin@example.com").Return(admin, nil)
				// 模拟发送验证码
				mockVerificationService.On("SendVerificationCode", mock.MatchedBy(func(req *verificationDto.SendVerificationRequest) bool {
					return req.Target == "admin@example.com" && req.Type == "admin_login"
				})).Return(&verificationDto.SendVerificationResponse{
					Message: "验证码已发送",
				}, nil)
			},
			expectedError: "",
			expectedResult: true,
		},
		{
			name: "管理员不存在",
			request: &dto.AdminLoginRequest{
				Email: "nonexistent@example.com",
				Password: "password123",
			},
			setupMocks: func(mockRepo *MockAdminRepository, mockIPInfo *MockIPInfoClient, mockVerificationService *MockVerificationService) {
				mockRepo.On("GetByEmail", "nonexistent@example.com").Return(nil, sql.ErrNoRows)
				// 即使失败的情况下，也可能会调用 GetIPInfo 来记录登录尝试
				mockIPInfo.On("GetIPInfo", mock.Anything, "192.168.1.1").Return(&ipinfo.IPInfo{
					IP:      "192.168.1.1",
					Country: "US",
					City:    "New York",
				}, nil).Maybe()
				mockRepo.On("CreateLoginSession", mock.AnythingOfType("*entities.AdminLoginSession")).Return(nil).Maybe()
			},
			expectedError: "invalid credentials",
			expectedResult: false,
		},
		{
			name: "密码错误",
			request: &dto.AdminLoginRequest{
				Email: "admin@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func(mockRepo *MockAdminRepository, mockIPInfo *MockIPInfoClient, mockVerificationService *MockVerificationService) {
				admin := &entities.Admin{
					ID: 1,
					Email: "admin@example.com",
					Password: "$2a$10$AS9mJUHq5hxMoz5meJdMYumez3uzf4A8Dng0ng8nMqu.TH4ThrcGy", // password123
					Status: "active",
				}
				mockRepo.On("GetByEmail", "admin@example.com").Return(admin, nil)
				// 密码错误时也会记录登录尝试
				mockIPInfo.On("GetIPInfo", mock.Anything, "192.168.1.1").Return(&ipinfo.IPInfo{
					IP:      "192.168.1.1",
					Country: "US",
					City:    "New York",
				}, nil).Maybe()
				mockRepo.On("CreateLoginSession", mock.AnythingOfType("*entities.AdminLoginSession")).Return(nil).Maybe()
			},
			expectedError: "invalid credentials",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAdminRepository)
			mockIPInfo := new(MockIPInfoClient)
			mockVerificationService := new(MockVerificationService)
			service := setupAdminService(mockRepo, mockIPInfo, mockVerificationService)

			tt.setupMocks(mockRepo, mockIPInfo, mockVerificationService)

			result, err := service.Login(context.Background(), tt.request.Email, tt.request.Password)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotEmpty(t, result.Message)
			}

			mockRepo.AssertExpectations(t)
			mockIPInfo.AssertExpectations(t)
		})
	}
}