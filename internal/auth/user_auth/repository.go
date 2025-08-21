package user_auth

import (
	"database/sql"
	"log"

	"trusioo_api/internal/auth/user_auth/entities"
	"trusioo_api/pkg/database"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	// User相关
	GetByEmail(email string) (*entities.User, error)
	GetByID(id int64) (*entities.User, error)
	GetByPhone(phone string) (*entities.User, error)
	Create(user *entities.User) error
	Update(user *entities.User) error
	UpdateLastLogin(id int64) error
	UpdatePassword(id int64, password string) error
	UpdateEmailVerified(id int64, verified bool) error

	// RefreshToken相关
	CreateRefreshToken(token *entities.RefreshToken) error
	GetValidRefreshToken(token string) (*entities.RefreshToken, error)
	InvalidateRefreshToken(token string) error
	InvalidateAllRefreshTokens(userID int64) error

	// LoginSession相关
	CreateLoginSession(session *entities.LoginSession) error
}

// userRepository Repository接口的实现
type userRepository struct{}

// Repository 类型别名，用于向后兼容
type Repository = UserRepository

// NewRepository 创建新的Repository实例
func NewRepository() Repository {
	return &userRepository{}
}

// NewUserRepository 创建新的UserRepository实例
func NewUserRepository() UserRepository {
	return &userRepository{}
}

// User相关方法
func (r *userRepository) GetByEmail(email string) (*entities.User, error) {
	var user entities.User
	err := database.DB.Get(&user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(id int64) (*entities.User, error) {
	var user entities.User
	err := database.DB.Get(&user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByPhone(phone string) (*entities.User, error) {
	var user entities.User
	err := database.DB.Get(&user, "SELECT * FROM users WHERE phone = $1", phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *entities.User) error {
	// 调试信息
	log.Printf("Creating user with status: %s", user.Status)
	
	query := `
		INSERT INTO users (name, email, password, phone, image_key, role, status,
			email_verified, phone_verified, auto_registered, profile_completed, password_set)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at, status`
	
	var returnedStatus string
	err := database.DB.QueryRow(query,
		user.Name,
		user.Email,
		user.Password,
		user.Phone,
		user.ImageKey,
		user.Role,
		user.Status,
		user.EmailVerified,
		user.PhoneVerified,
		user.AutoRegistered,
		user.ProfileCompleted,
		user.PasswordSet,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &returnedStatus)
	
	if err != nil {
		return err
	}
	
	// 调试信息
	log.Printf("User created with ID: %d, status: %s", user.ID, returnedStatus)
	user.Status = returnedStatus
	
	return nil
}

func (r *userRepository) Update(user *entities.User) error {
	query := `
		UPDATE users 
		SET name = $2, email = $3, phone = $4, image_key = $5, role = $6,
			status = $7, email_verified = $8, phone_verified = $9, auto_registered = $10,
			profile_completed = $11, password_set = $12, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`
	
	return database.DB.QueryRow(query,
		user.ID,
		user.Name,
		user.Email,
		user.Phone,
		user.ImageKey,
		user.Role,
		user.Status,
		user.EmailVerified,
		user.PhoneVerified,
		user.AutoRegistered,
		user.ProfileCompleted,
		user.PasswordSet,
	).Scan(&user.UpdatedAt)
}

func (r *userRepository) UpdateLastLogin(id int64) error {
	_, err := database.DB.Exec("UPDATE users SET last_login_at = NOW() WHERE id = $1", id)
	return err
}

func (r *userRepository) UpdatePassword(id int64, password string) error {
	_, err := database.DB.Exec("UPDATE users SET password = $2, updated_at = NOW() WHERE id = $1", id, password)
	return err
}

func (r *userRepository) UpdateEmailVerified(id int64, verified bool) error {
	_, err := database.DB.Exec("UPDATE users SET email_verified = $2, updated_at = NOW() WHERE id = $1", id, verified)
	return err
}

// RefreshToken相关方法
func (r *userRepository) CreateRefreshToken(token *entities.RefreshToken) error {
	query := `
		INSERT INTO user_refresh_tokens (user_id, token, is_valid, expires_at, device_info)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	
	return database.DB.QueryRow(query,
		token.UserID,
		token.Token,
		token.IsValid,
		token.ExpiresAt,
		token.DeviceInfo,
	).Scan(&token.ID, &token.CreatedAt)
}

func (r *userRepository) GetValidRefreshToken(token string) (*entities.RefreshToken, error) {
	var refreshToken entities.RefreshToken
	query := `
		SELECT * FROM user_refresh_tokens 
		WHERE token = $1 AND is_valid = true AND expires_at > NOW()`
	
	err := database.DB.Get(&refreshToken, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &refreshToken, nil
}

func (r *userRepository) InvalidateRefreshToken(token string) error {
	_, err := database.DB.Exec(
		"UPDATE user_refresh_tokens SET is_valid = false WHERE token = $1",
		token,
	)
	return err
}

func (r *userRepository) InvalidateAllRefreshTokens(userID int64) error {
	_, err := database.DB.Exec(
		"UPDATE user_refresh_tokens SET is_valid = false WHERE user_id = $1",
		userID,
	)
	return err
}

// LoginSession相关方法
func (r *userRepository) CreateLoginSession(session *entities.LoginSession) error {
	query := `
		INSERT INTO user_login_sessions (
			user_id, ip, country, city, region, timezone, organization, location,
			user_agent, device_type, os, browser, is_trusted, login_method, platform, status, reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at`
	
	return database.DB.QueryRow(query,
		session.UserID,
		session.IP,
		session.Country,
		session.City,
		session.Region,
		session.Timezone,
		session.Organization,
		session.Location,
		session.UserAgent,
		session.DeviceType,
		session.OS,
		session.Browser,
		session.IsTrusted,
		session.LoginMethod,
		session.Platform,
		session.Status,
		session.Reason,
	).Scan(&session.ID, &session.CreatedAt)
}