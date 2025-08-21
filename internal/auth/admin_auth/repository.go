package admin

import (
	"database/sql"
	"fmt"

	"trusioo_api/internal/auth/admin_auth/dto"
	"trusioo_api/internal/auth/admin_auth/entities"
	"trusioo_api/pkg/database"
)

// AdminRepository 管理员数据访问接口
type AdminRepository interface {
	// Admin相关
	GetByEmail(email string) (*entities.Admin, error)
	GetByID(id int64) (*entities.Admin, error)
	Create(admin *entities.Admin) error
	Update(admin *entities.Admin) error
	UpdateLastLogin(id int64) error
	UpdatePassword(id int64, password string) error

	// RefreshToken相关
	CreateRefreshToken(token *entities.AdminRefreshToken) error
	GetValidRefreshToken(token string) (*entities.AdminRefreshToken, error)
	InvalidateRefreshToken(token string) error
	InvalidateAllRefreshTokens(adminID int64) error

	// LoginSession相关
	CreateLoginSession(session *entities.AdminLoginSession) error

	// 用户管理相关
	GetUserStats() (*dto.UserStats, error)
	GetUserList(req *dto.UserListRequest) (*dto.UserListResponse, error)
	GetUserByID(id int64) (*entities.UserInfo, error)
}

// adminRepository Repository接口的实现
type adminRepository struct{}

// NewAdminRepository 创建新的AdminRepository实例
func NewAdminRepository() AdminRepository {
	return &adminRepository{}
}

// Admin相关方法
func (r *adminRepository) GetByEmail(email string) (*entities.Admin, error) {
	var admin entities.Admin
	err := database.DB.Get(&admin, "SELECT * FROM admins WHERE email = $1", email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) GetByID(id int64) (*entities.Admin, error) {
	var admin entities.Admin
	err := database.DB.Get(&admin, "SELECT * FROM admins WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) Create(admin *entities.Admin) error {
	query := `
		INSERT INTO admins (name, email, password, phone, image_key, role, is_super, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`
	
	return database.DB.QueryRow(query,
		admin.Name,
		admin.Email,
		admin.Password,
		admin.Phone,
		admin.ImageKey,
		admin.Role,
		admin.IsSuper,
		admin.Status,
	).Scan(&admin.ID, &admin.CreatedAt, &admin.UpdatedAt)
}

func (r *adminRepository) Update(admin *entities.Admin) error {
	query := `
		UPDATE admins 
		SET name = $2, email = $3, phone = $4, image_key = $5, role = $6, 
			is_super = $7, status = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`
	
	return database.DB.QueryRow(query,
		admin.ID,
		admin.Name,
		admin.Email,
		admin.Phone,
		admin.ImageKey,
		admin.Role,
		admin.IsSuper,
		admin.Status,
	).Scan(&admin.UpdatedAt)
}

func (r *adminRepository) UpdateLastLogin(id int64) error {
	_, err := database.DB.Exec("UPDATE admins SET last_login_at = NOW() WHERE id = $1", id)
	return err
}

func (r *adminRepository) UpdatePassword(id int64, password string) error {
	_, err := database.DB.Exec("UPDATE admins SET password = $2, updated_at = NOW() WHERE id = $1", id, password)
	return err
}

// RefreshToken相关方法
func (r *adminRepository) CreateRefreshToken(token *entities.AdminRefreshToken) error {
	query := `
		INSERT INTO admin_refresh_tokens (admin_id, token, is_valid, expires_at, device_info)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	
	return database.DB.QueryRow(query,
		token.AdminID,
		token.Token,
		token.IsValid,
		token.ExpiresAt,
		token.DeviceInfo,
	).Scan(&token.ID, &token.CreatedAt)
}

func (r *adminRepository) GetValidRefreshToken(token string) (*entities.AdminRefreshToken, error) {
	var refreshToken entities.AdminRefreshToken
	query := `
		SELECT * FROM admin_refresh_tokens 
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

func (r *adminRepository) InvalidateRefreshToken(token string) error {
	_, err := database.DB.Exec(
		"UPDATE admin_refresh_tokens SET is_valid = false WHERE token = $1",
		token,
	)
	return err
}

func (r *adminRepository) InvalidateAllRefreshTokens(adminID int64) error {
	_, err := database.DB.Exec(
		"UPDATE admin_refresh_tokens SET is_valid = false WHERE admin_id = $1",
		adminID,
	)
	return err
}

// LoginSession相关方法
func (r *adminRepository) CreateLoginSession(session *entities.AdminLoginSession) error {
	query := `
		INSERT INTO admin_login_sessions (
			admin_id, ip, country, city, region, timezone, organization, location,
			user_agent, device_type, os, browser, is_trusted, platform, status, reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at`
	
	return database.DB.QueryRow(query,
		session.AdminID,
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
		session.Platform,
		session.Status,
		session.Reason,
	).Scan(&session.ID, &session.CreatedAt)
}

// 用户管理相关方法
func (r *adminRepository) GetUserStats() (*dto.UserStats, error) {
	var stats dto.UserStats
	
	// 获取总用户数
	err := database.DB.Get(&stats.TotalUsers, "SELECT COUNT(*) FROM users")
	if err != nil {
		return nil, err
	}
	
	// 获取活跃用户数
	err = database.DB.Get(&stats.ActiveUsers, "SELECT COUNT(*) FROM users WHERE status = 'active'")
	if err != nil {
		return nil, err
	}
	
	// 获取非活跃用户数
	err = database.DB.Get(&stats.InactiveUsers, "SELECT COUNT(*) FROM users WHERE status = 'inactive'")
	if err != nil {
		return nil, err
	}
	
	// 获取今日注册用户数
	err = database.DB.Get(&stats.RegisteredToday, 
		"SELECT COUNT(*) FROM users WHERE DATE(created_at) = CURRENT_DATE")
	if err != nil {
		return nil, err
	}
	
	// 获取本周注册用户数
	err = database.DB.Get(&stats.RegisteredThisWeek, 
		"SELECT COUNT(*) FROM users WHERE created_at >= DATE_TRUNC('week', NOW())")
	if err != nil {
		return nil, err
	}
	
	// 获取本月注册用户数
	err = database.DB.Get(&stats.RegisteredThisMonth, 
		"SELECT COUNT(*) FROM users WHERE created_at >= DATE_TRUNC('month', NOW())")
	if err != nil {
		return nil, err
	}
	
	return &stats, nil
}

func (r *adminRepository) GetUserList(req *dto.UserListRequest) (*dto.UserListResponse, error) {
	// 设置默认值
	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	
	// 构建查询条件
	var whereConditions []string
	var args []interface{}
	argIndex := 1
	
	if req.Status != "" && req.Status != "all" {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, req.Status)
		argIndex++
	}
	
	if req.Email != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("email ILIKE $%d", argIndex))
		args = append(args, "%"+req.Email+"%")
		argIndex++
	}
	
	if req.Phone != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("phone ILIKE $%d", argIndex))
		args = append(args, "%"+req.Phone+"%")
		argIndex++
	}
	
	// 构建完整查询
	baseQuery := `SELECT id, name, email, phone, image_key, status, 
					email_verified, phone_verified, auto_registered, profile_completed, 
					last_login_at, created_at FROM users`
	
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = " WHERE " + whereConditions[0]
		for i := 1; i < len(whereConditions); i++ {
			whereClause += " AND " + whereConditions[i]
		}
	}
	
	// 获取总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM users" + whereClause
	err := database.DB.Get(&total, countQuery, args...)
	if err != nil {
		return nil, err
	}
	
	// 获取用户列表
	query := baseQuery + whereClause + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)
	
	var users []entities.UserInfo
	err = database.DB.Select(&users, query, args...)
	if err != nil {
		return nil, err
	}
	
	return &dto.UserListResponse{
		Total: total,
		Page:  page,
		Size:  pageSize,
		Users: users,
	}, nil
}

func (r *adminRepository) GetUserByID(id int64) (*entities.UserInfo, error) {
	var user entities.UserInfo
	query := `SELECT id, name, email, phone, image_key, status, 
				email_verified, phone_verified, auto_registered, profile_completed, 
				last_login_at, created_at 
			  FROM users WHERE id = $1`
	
	err := database.DB.Get(&user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &user, nil
}