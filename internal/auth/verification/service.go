package verification

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"time"

	"trusioo_api/internal/auth/verification/dto"
	"trusioo_api/internal/auth/verification/entities"
	"trusioo_api/pkg/database"
)

type Service struct {
	repo *Repository
}

func NewService() *Service {
	return &Service{
		repo: NewRepository(),
	}
}

// checkUserExists 检查用户是否存在于数据库中
func (s *Service) checkUserExists(email, userType string) (bool, int64, error) {
	var userID int64
	var query string
	
	if userType == "admin" {
		query = "SELECT id FROM admins WHERE email = $1"
	} else {
		query = "SELECT id FROM users WHERE email = $1"
	}
	
	err := database.DB.QueryRow(query, email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0, nil // 用户不存在
		}
		return false, 0, err // 数据库错误
	}
	
	return true, userID, nil // 用户存在
}

// activateUserAccount 激活用户账户
func (s *Service) activateUserAccount(email string) error {
	query := `
		UPDATE users 
		SET status = 'active', email_verified = true, updated_at = NOW()
		WHERE email = $1 AND status = 'inactive'
	`
	
	result, err := database.DB.Exec(query, email)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already active")
	}
	
	return nil
}

// SendVerificationCode 发送验证码
func (s *Service) SendVerificationCode(req *dto.SendVerificationRequest) (*dto.SendVerificationResponse, error) {
	// 对于注册类型的验证码，需要检查用户是否存在于数据库中
	if req.Type == "register" {
		exists, _, err := s.checkUserExists(req.Target, "user")
		if err != nil {
			return nil, fmt.Errorf("failed to check user existence: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("user not found in database, please register first")
		}
	}

	// 检查发送频率限制（60秒内不能重复发送）
	recentVerification, err := s.repo.GetRecentVerification(req.Target, req.Type, time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to check recent verification: %w", err)
	}

	if recentVerification != nil {
		return nil, fmt.Errorf("verification code was sent recently, please wait before requesting again")
	}

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

	// 如果是注册类型的验证码，激活用户账户
	if req.Type == "register" {
		err = s.activateUserAccount(req.Target)
		if err != nil {
			return nil, fmt.Errorf("failed to activate user account: %w", err)
		}
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
