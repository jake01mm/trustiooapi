package main

import (
	"database/sql"
	"fmt"
	"log"

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

	fmt.Println("🔧 初始化 Neon Database 迁移系统...")

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

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Printf("✅ 数据库连接成功: %s:%s/%s\n", cfg.Host, cfg.Port, cfg.Name)

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

	// Check current status
	currentVersion, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			fmt.Println("📋 数据库为空，准备执行初始迁移...")
			
			// Run initial migration
			err = m.Up()
			if err != nil && err != migrate.ErrNoChange {
				log.Fatalf("Initial migration failed: %v", err)
			}
			
			newVersion, _, err := m.Version()
			if err != nil {
				log.Fatalf("Failed to get new version: %v", err)
			}
			
			fmt.Printf("🎉 初始化完成！当前数据库版本: %d\n", newVersion)
		} else {
			log.Fatalf("Failed to get migration version: %v", err)
		}
	} else {
		fmt.Printf("📊 数据库已初始化，当前版本: %d", currentVersion)
		if dirty {
			fmt.Println(" ⚠️  (脏状态)")
			fmt.Println("💡 使用 make db-force N=<version> 来修复脏状态")
		} else {
			fmt.Println(" ✅ (干净状态)")
			
			// Check for pending migrations
			err = m.Up()
			if err == migrate.ErrNoChange {
				fmt.Println("📋 没有待处理的迁移")
			} else if err != nil {
				log.Fatalf("Migration failed: %v", err)
			} else {
				finalVersion, _, _ := m.Version()
				fmt.Printf("🚀 执行了待处理的迁移，当前版本: %d\n", finalVersion)
			}
		}
	}

	fmt.Println("\n✨ Neon Database 迁移系统已就绪！")
	fmt.Println("💡 使用 'make help' 查看所有可用命令")
}