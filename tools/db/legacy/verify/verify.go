package main

import (
	"log"

	"trusioo_api/config"
	"trusioo_api/pkg/database"
)

type TableInfo struct {
	TableName string `db:"table_name"`
}

func main() {
	log.Println("Verifying database tables...")

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDatabase()

	// 查询所有表
	var tables []TableInfo
	err := database.DB.Select(&tables, `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`)
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}

	log.Printf("Found %d tables:", len(tables))
	for _, table := range tables {
		log.Printf("  - %s", table.TableName)
	}

	// 验证特定表是否存在
	expectedTables := []string{
		"users",
		"admins",
		"verifications",
		"user_refresh_tokens",
		"admin_refresh_tokens",
		"user_login_sessions",
		"admin_login_sessions",
	}

	log.Println("\nChecking expected tables:")
	for _, expectedTable := range expectedTables {
		found := false
		for _, table := range tables {
			if table.TableName == expectedTable {
				found = true
				break
			}
		}
		if found {
			log.Printf("  ✓ %s - exists", expectedTable)
		} else {
			log.Printf("  ✗ %s - missing", expectedTable)
		}
	}

	// 检查管理员账户是否存在
	var adminCount int
	err = database.DB.Get(&adminCount, "SELECT COUNT(*) FROM admins WHERE email = 'admin@trusioo.com'")
	if err != nil {
		log.Printf("Failed to check admin account: %v", err)
	} else {
		if adminCount > 0 {
			log.Println("\n✓ Default admin account exists")
		} else {
			log.Println("\n✗ Default admin account not found")
		}
	}

	log.Println("\nDatabase verification completed!")
}