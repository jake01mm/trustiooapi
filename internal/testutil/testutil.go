package testutil

import (
	"fmt"
	"testing"
	"time"

	"trusioo_api/config"
	"trusioo_api/pkg/auth"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestDB 测试数据库连接
type TestDB struct {
	*sqlx.DB
	testConfig *config.TestConfig
}

// NewTestDB 创建测试数据库连接
func NewTestDB(t *testing.T) *TestDB {
	testConfig := config.LoadTestConfig()
	
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		testConfig.DB.Host,
		testConfig.DB.Port,
		testConfig.DB.User,
		testConfig.DB.Password,
		testConfig.DB.Name,
		testConfig.DB.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err, "Failed to connect to test database")

	return &TestDB{
		DB: db,
		testConfig: testConfig,
	}
}

// Close 关闭数据库连接
func (tdb *TestDB) Close() {
	if tdb.DB != nil {
		tdb.DB.Close()
	}
}

// CleanupTables 清理测试表数据
func (tdb *TestDB) CleanupTables(t *testing.T, tables ...string) {
	for _, table := range tables {
		_, err := tdb.Exec(fmt.Sprintf("DELETE FROM %s", table))
		require.NoError(t, err, "Failed to cleanup table %s", table)
	}
}

// CleanupAllAuthTables 清理所有认证相关表
func (tdb *TestDB) CleanupAllAuthTables(t *testing.T) {
	tables := []string{
		"user_login_sessions",
		"admin_login_sessions", 
		"user_refresh_tokens",
		"admin_refresh_tokens",
		"verification_codes",
		"users",
		"admins",
	}
	tdb.CleanupTables(t, tables...)
}

// CreateTestUser 创建测试用户
func (tdb *TestDB) CreateTestUser(t *testing.T, email, password string) int64 {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	var userID int64
	err = tdb.QueryRow(`
		INSERT INTO users (email, password_hash, email_verified, phone_verified, profile_completed, created_at, updated_at)
		VALUES ($1, $2, false, false, false, $3, $3)
		RETURNING id
	`, email, hashedPassword, time.Now()).Scan(&userID)
	require.NoError(t, err)

	return userID
}

// CreateTestAdmin 创建测试管理员
func (tdb *TestDB) CreateTestAdmin(t *testing.T, email, password, role string) int64 {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	var adminID int64
	err = tdb.QueryRow(`
		INSERT INTO admins (email, password_hash, role, email_verified, phone_verified, profile_completed, created_at, updated_at)
		VALUES ($1, $2, $3, false, false, false, $4, $4)
		RETURNING id
	`, email, hashedPassword, role, time.Now()).Scan(&adminID)
	require.NoError(t, err)

	return adminID
}

// CreateVerificationCode 创建测试验证码
func (tdb *TestDB) CreateVerificationCode(t *testing.T, email, code, codeType string) {
	_, err := tdb.Exec(`
		INSERT INTO verification_codes (email, code, code_type, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, email, code, codeType, time.Now().Add(10*time.Minute), time.Now())
	require.NoError(t, err)
}

// TestRedis 测试Redis连接
type TestRedis struct {
	*redis.Client
	testConfig *config.TestConfig
}

// NewTestRedis 创建测试Redis连接
func NewTestRedis(t *testing.T) *TestRedis {
	testConfig := config.LoadTestConfig()
	
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", testConfig.Redis.Host, testConfig.Redis.Port),
		Password: testConfig.Redis.Password,
		DB:       testConfig.Redis.DB,
	})

	return &TestRedis{
		Client: rdb,
		testConfig: testConfig,
	}
}

// Close 关闭Redis连接
func (tr *TestRedis) Close() {
	if tr.Client != nil {
		tr.Client.Close()
	}
}

// MockJWTConfig 模拟JWT配置
func MockJWTConfig() {
	// 确保AppConfig已初始化
	if config.AppConfig == nil {
		config.LoadConfig()
	}
	testConfig := config.LoadTestConfig()
	config.AppConfig.JWT.Secret = testConfig.JWT.Secret
	config.AppConfig.JWT.RefreshSecret = testConfig.JWT.RefreshSecret
	config.AppConfig.JWT.AccessExpire = testConfig.JWT.AccessExpire
	config.AppConfig.JWT.RefreshExpire = testConfig.JWT.RefreshExpire
}

// GenerateTestToken 生成测试用的JWT token
func GenerateTestToken(userID int64, email, role, userType string) (string, error) {
	MockJWTConfig()
	return auth.GenerateAccessToken(userID, email, role, userType)
}