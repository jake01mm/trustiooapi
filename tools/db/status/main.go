package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"trusioo_api/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	showMigrationStatus()
}

func showMigrationStatus() {
	cfg := config.AppConfig.Database

	// Create database connection string
	dbURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=require",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create postgres driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create postgres driver: %v", err)
	}

	// Create migrate instance
	migrationsPath := "file://migrations"
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	// Get current version
	currentVersion, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			fmt.Println("📊 数据库迁移状态:")
			fmt.Println("   当前版本: 无 (数据库未初始化)")
		} else {
			log.Fatalf("Failed to get migration version: %v", err)
		}
	} else {
		fmt.Println("📊 数据库迁移状态:")
		fmt.Printf("   当前版本: %d", currentVersion)
		if dirty {
			fmt.Println(" ⚠️  (脏状态 - 迁移可能失败)")
		} else {
			fmt.Println(" ✅ (干净状态)")
		}
	}

	// List available migrations
	fmt.Println("\n📋 可用的迁移文件:")
	migrationsDir := "migrations"
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		log.Printf("Failed to read migrations directory: %v", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("   无迁移文件")
		return
	}

	sort.Strings(files)
	for _, file := range files {
		basename := filepath.Base(file)
		parts := strings.Split(basename, "_")
		if len(parts) >= 2 {
			versionStr := parts[0]
			name := strings.Join(parts[1:], "_")
			name = strings.TrimSuffix(name, ".up.sql")
			
			status := "📋 待执行"
			if err == nil && !dirty {
				if version, parseErr := parseInt(versionStr); parseErr == nil && version <= int(currentVersion) {
					status = "✅ 已执行"
				}
			}
			
			fmt.Printf("   %s %s - %s\n", status, versionStr, name)
		}
	}

	fmt.Println("\n💡 使用命令:")
	fmt.Println("   make db-migrate          # 执行待处理的迁移")
	fmt.Println("   make db-rollback         # 回滚最后一个迁移")
	fmt.Println("   make db-create NAME=xxx  # 创建新迁移")
}

func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}