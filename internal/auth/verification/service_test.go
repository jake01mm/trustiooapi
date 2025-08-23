package verification

import (
	"testing"
	"time"

	"trusioo_api/internal/auth/verification/dto"
	"trusioo_api/internal/auth/verification/entities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository 模拟 Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateVerification(verification *entities.Verification) error {
	args := m.Called(verification)
	return args.Error(0)
}

func (m *MockRepository) GetValidVerification(target, verifyType, code string) (*entities.Verification, error) {
	args := m.Called(target, verifyType, code)
	return args.Get(0).(*entities.Verification), args.Error(1)
}

func (m *MockRepository) MarkVerificationAsUsed(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepository) DeleteExpiredVerifications() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRepository) GetRecentVerification(target, verifyType string, duration time.Duration) (*entities.Verification, error) {
	args := m.Called(target, verifyType, duration)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Verification), args.Error(1)
}

// TestService 测试用的服务结构体
type TestService struct {
	repo *MockRepository
}

func (s *TestService) SendVerificationCode(req *dto.SendVerificationRequest) (*dto.SendVerificationResponse, error) {
	// 生成6位数字验证码
	code := "123456" // 测试用固定验证码

	// 创建验证码记录
	now := time.Now()
	expiredAt := now.Add(10 * time.Minute) // 10分钟有效期

	verification := &entities.Verification{
		Target:    req.Target,
		Type:      req.Type,
		Action:    entities.ActionEmailVerification,
		SentAt:    now,
		Code:      code,
		IsUsed:    false,
		ExpiredAt: expiredAt,
		CreatedAt: now,
	}

	err := s.repo.CreateVerification(verification)
	if err != nil {
		return nil, err
	}

	return &dto.SendVerificationResponse{
		Message:   "Verification code sent successfully",
		ExpiredAt: expiredAt.Format(time.RFC3339),
		Code:      code,
	}, nil
}

func (s *TestService) VerifyCode(req *dto.VerifyCodeRequest) (*dto.VerifyCodeResponse, error) {
	// 获取有效的验证码
	verification, err := s.repo.GetValidVerification(req.Target, req.Type, req.Code)
	if err != nil {
		return nil, err
	}

	if verification == nil {
		return &dto.VerifyCodeResponse{
			Message: "Invalid or expired verification code",
			Valid:   false,
		}, nil
	}

	// 标记验证码为已使用
	err = s.repo.MarkVerificationAsUsed(verification.ID)
	if err != nil {
		return nil, err
	}

	return &dto.VerifyCodeResponse{
		Message: "Verification code is valid",
		Valid:   true,
	}, nil
}

func (s *TestService) CleanupExpiredVerifications() error {
	return s.repo.DeleteExpiredVerifications()
}

// setupVerificationService 设置测试服务
func setupVerificationService() (*TestService, *MockRepository) {
	mockRepo := new(MockRepository)
	service := &TestService{
		repo: mockRepo,
	}
	return service, mockRepo
}

func TestService_SendVerificationCode(t *testing.T) {
	tests := []struct {
		name          string
		request       *dto.SendVerificationRequest
		mockSetup     func(*MockRepository)
		expectedError string
		expectedCode  string
	}{
		{
			name: "成功发送验证码",
			request: &dto.SendVerificationRequest{
				Target: "test@example.com",
				Type:   entities.TypeRegister,
			},
			mockSetup: func(mockRepo *MockRepository) {
				mockRepo.On("CreateVerification", mock.AnythingOfType("*entities.Verification")).Return(nil)
			},
			expectedError: "",
			expectedCode:  "123456",
		},
		{
			name: "数据库错误",
			request: &dto.SendVerificationRequest{
				Target: "test@example.com",
				Type:   entities.TypeRegister,
			},
			mockSetup: func(mockRepo *MockRepository) {
				mockRepo.On("CreateVerification", mock.AnythingOfType("*entities.Verification")).Return(assert.AnError)
			},
			expectedError: "assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo := setupVerificationService()
			tt.mockSetup(mockRepo)

			response, err := service.SendVerificationCode(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, "Verification code sent successfully", response.Message)
				assert.Equal(t, tt.expectedCode, response.Code)
				assert.NotEmpty(t, response.ExpiredAt)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_VerifyCode(t *testing.T) {
	now := time.Now()
	validVerification := &entities.Verification{
		ID:        1,
		Target:    "test@example.com",
		Type:      entities.TypeRegister,
		Action:    entities.ActionEmailVerification,
		Code:      "123456",
		IsUsed:    false,
		ExpiredAt: now.Add(10 * time.Minute),
		CreatedAt: now,
	}

	tests := []struct {
		name          string
		request       *dto.VerifyCodeRequest
		mockSetup     func(*MockRepository)
		expectedError string
		expectedValid bool
	}{
		{
			name: "验证码有效",
			request: &dto.VerifyCodeRequest{
				Target: "test@example.com",
				Type:   entities.TypeRegister,
				Code:   "123456",
			},
			mockSetup: func(mockRepo *MockRepository) {
				mockRepo.On("GetValidVerification", "test@example.com", entities.TypeRegister, "123456").Return(validVerification, nil)
				mockRepo.On("MarkVerificationAsUsed", int64(1)).Return(nil)
			},
			expectedError: "",
			expectedValid: true,
		},
		{
			name: "验证码无效或过期",
			request: &dto.VerifyCodeRequest{
				Target: "test@example.com",
				Type:   entities.TypeRegister,
				Code:   "999999",
			},
			mockSetup: func(mockRepo *MockRepository) {
				mockRepo.On("GetValidVerification", "test@example.com", entities.TypeRegister, "999999").Return((*entities.Verification)(nil), nil)
			},
			expectedError: "",
			expectedValid: false,
		},
		{
			name: "数据库错误",
			request: &dto.VerifyCodeRequest{
				Target: "test@example.com",
				Type:   entities.TypeRegister,
				Code:   "123456",
			},
			mockSetup: func(mockRepo *MockRepository) {
				mockRepo.On("GetValidVerification", "test@example.com", entities.TypeRegister, "123456").Return((*entities.Verification)(nil), assert.AnError)
			},
			expectedError: "assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo := setupVerificationService()
			tt.mockSetup(mockRepo)

			response, err := service.VerifyCode(tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tt.expectedValid, response.Valid)
				if tt.expectedValid {
					assert.Equal(t, "Verification code is valid", response.Message)
				} else {
					assert.Equal(t, "Invalid or expired verification code", response.Message)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_CleanupExpiredVerifications(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockRepository)
		expectedError string
	}{
		{
			name: "成功清理过期验证码",
			mockSetup: func(mockRepo *MockRepository) {
				mockRepo.On("DeleteExpiredVerifications").Return(nil)
			},
			expectedError: "",
		},
		{
			name: "数据库错误",
			mockSetup: func(mockRepo *MockRepository) {
				mockRepo.On("DeleteExpiredVerifications").Return(assert.AnError)
			},
			expectedError: "assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo := setupVerificationService()
			tt.mockSetup(mockRepo)

			err := service.CleanupExpiredVerifications()

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}