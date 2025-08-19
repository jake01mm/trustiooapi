package verification

import (
	"database/sql"
	"fmt"
	"time"

	"trusioo_api/internal/auth/verification/entities"
	"trusioo_api/pkg/database"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository() *Repository {
	return &Repository{
		db: database.DB,
	}
}

// CreateVerification 创建验证码记录
func (r *Repository) CreateVerification(verification *entities.Verification) error {
	query := `
		INSERT INTO verifications (user_id, target, type, action, sent_at, code, is_used, expired_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		verification.UserID,
		verification.Target,
		verification.Type,
		verification.Action,
		verification.SentAt,
		verification.Code,
		verification.IsUsed,
		verification.ExpiredAt,
		verification.CreatedAt,
	).Scan(&verification.ID)

	if err != nil {
		return fmt.Errorf("failed to create verification: %w", err)
	}

	return nil
}

// GetValidVerification 获取有效的验证码
func (r *Repository) GetValidVerification(target, verifyType, code string) (*entities.Verification, error) {
	query := `
		SELECT id, user_id, target, type, action, sent_at, code, is_used, expired_at, created_at
		FROM verifications
		WHERE target = $1 AND type = $2 AND code = $3 AND is_used = false AND expired_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	verification := &entities.Verification{}
	err := r.db.QueryRow(query, target, verifyType, code).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.Target,
		&verification.Type,
		&verification.Action,
		&verification.SentAt,
		&verification.Code,
		&verification.IsUsed,
		&verification.ExpiredAt,
		&verification.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get verification: %w", err)
	}

	return verification, nil
}

// MarkVerificationAsUsed 标记验证码为已使用
func (r *Repository) MarkVerificationAsUsed(id int64) error {
	query := `UPDATE verifications SET is_used = true WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to mark verification as used: %w", err)
	}
	return nil
}

// DeleteExpiredVerifications 删除过期的验证码
func (r *Repository) DeleteExpiredVerifications() error {
	query := `DELETE FROM verifications WHERE expired_at < NOW()`
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete expired verifications: %w", err)
	}
	return nil
}

// GetRecentVerification 获取最近发送的验证码（用于限制发送频率）
func (r *Repository) GetRecentVerification(target, verifyType string, duration time.Duration) (*entities.Verification, error) {
	query := `
		SELECT id, user_id, target, type, action, sent_at, code, is_used, expired_at, created_at
		FROM verifications
		WHERE target = $1 AND type = $2 AND sent_at > $3
		ORDER BY sent_at DESC
		LIMIT 1
	`

	verification := &entities.Verification{}
	limitTime := time.Now().Add(-duration)

	err := r.db.QueryRow(query, target, verifyType, limitTime).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.Target,
		&verification.Type,
		&verification.Action,
		&verification.SentAt,
		&verification.Code,
		&verification.IsUsed,
		&verification.ExpiredAt,
		&verification.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get recent verification: %w", err)
	}

	return verification, nil
}
