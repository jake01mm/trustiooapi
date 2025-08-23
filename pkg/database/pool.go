package database

import (
	"database/sql"
	"fmt"
	"time"

	"trusioo_api/config"
	"trusioo_api/pkg/logger"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func InitDatabase() error {
	cfg := config.AppConfig.Database

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	logger.WithFields(map[string]interface{}{
		"host": cfg.Host,
		"port": cfg.Port,
		"database": cfg.Name,
	}).Info("Connecting to database")

	var err error
	DB, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		logger.WithError(err).Error("Failed to connect to database")
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	DB.SetMaxOpenConns(cfg.MaxOpenConns)
	DB.SetMaxIdleConns(cfg.MaxIdleConns)
	DB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// 测试连接
	if err := DB.Ping(); err != nil {
		logger.WithError(err).Error("Failed to ping database")
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"max_open_conns": cfg.MaxOpenConns,
		"max_idle_conns": cfg.MaxIdleConns,
		"conn_max_lifetime": cfg.ConnMaxLifetime,
	}).Info("Database connected successfully")

	return nil
}

func CloseDatabase() error {
	if DB != nil {
		logger.Info("Closing database connection")
		return DB.Close()
	}
	return nil
}

// GetDB 获取数据库连接
func GetDB() *sqlx.DB {
	return DB
}

// GetStdDB 获取标准数据库连接
func GetStdDB() *sql.DB {
	if DB != nil {
		return DB.DB
	}
	return nil
}

// HealthCheck 数据库健康检查
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	return DB.Ping()
}

// Stats 获取数据库连接池统计信息
func Stats() sql.DBStats {
	if DB != nil {
		return DB.Stats()
	}
	return sql.DBStats{}
}