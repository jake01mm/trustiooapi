package verification

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"trusioo_api/internal/auth/verification/dto"
	"trusioo_api/internal/auth/verification/entities"
)

type Service struct {
	repo *Repository
}

func NewService() *Service {
	return &Service{
		repo: NewRepository(),
	}
}



// SendVerificationCode 发送验证码
func (s *Service) SendVerificationCode(req *dto.SendVerificationRequest) (*dto.SendVerificationResponse, error) {


	// 生成6位数字验证码
	code, err := s.generateVerificationCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification code: %w", err)
	}

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

	err = s.repo.CreateVerification(verification)
	if err != nil {
		return nil, fmt.Errorf("failed to save verification: %w", err)
	}

	// 发送验证码邮件（这里暂时只记录日志，实际项目中需要集成邮件服务）
	err = s.sendEmail(req.Target, code, req.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to send verification email: %w", err)
	}

	return &dto.SendVerificationResponse{
		Message:   "Verification code sent successfully",
		ExpiredAt: expiredAt.Format(time.RFC3339),
		Code:      code, // 仅用于测试环境
	}, nil
}

// VerifyCode 验证验证码
func (s *Service) VerifyCode(req *dto.VerifyCodeRequest) (*dto.VerifyCodeResponse, error) {
	// 获取有效的验证码
	verification, err := s.repo.GetValidVerification(req.Target, req.Type, req.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification: %w", err)
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
		return nil, fmt.Errorf("failed to mark verification as used: %w", err)
	}


	return &dto.VerifyCodeResponse{
		Message: "Verification code is valid",
		Valid:   true,
	}, nil
}

// generateVerificationCode 生成6位数字验证码
func (s *Service) generateVerificationCode() (string, error) {
	code := ""
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += n.String()
	}
	return code, nil
}

// sendEmail 发送验证码邮件（模拟实现）
func (s *Service) sendEmail(target, code, verifyType string) error {
	// 将验证码打印到控制台，方便测试
	log.Printf("=== 验证码发送 ===")
	log.Printf("邮箱: %s", target)
	log.Printf("验证码: %s", code)
	log.Printf("类型: %s", verifyType)
	log.Printf("==================")

	// TODO: 集成真实的邮件服务
	// 例如：SendGrid, AWS SES, 阿里云邮件推送等

	// 模拟邮件发送可能的错误
	if target == "" {
		return fmt.Errorf("invalid email address")
	}

	return nil
}

// CleanupExpiredVerifications 清理过期的验证码
func (s *Service) CleanupExpiredVerifications() error {
	return s.repo.DeleteExpiredVerifications()
}
