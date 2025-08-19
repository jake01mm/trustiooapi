package main

import (
	"log"

	"trusioo_api/config"
	"trusioo_api/pkg/database"
)

func main() {
	log.Println("Checking table data...")

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDatabase()

	// 检查用户刷新令牌表
	var tokenCount int
	err := database.DB.Get(&tokenCount, "SELECT COUNT(*) FROM user_refresh_tokens")
	if err != nil {
		log.Printf("Failed to check user_refresh_tokens: %v", err)
	} else {
		log.Printf("user_refresh_tokens: %d records", tokenCount)
	}

	// 检查用户登录会话表
	var sessionCount int
	err = database.DB.Get(&sessionCount, "SELECT COUNT(*) FROM user_login_sessions")
	if err != nil {
		log.Printf("Failed to check user_login_sessions: %v", err)
	} else {
		log.Printf("user_login_sessions: %d records", sessionCount)
	}

	// 检查管理员刷新令牌表
	var adminTokenCount int
	err = database.DB.Get(&adminTokenCount, "SELECT COUNT(*) FROM admin_refresh_tokens")
	if err != nil {
		log.Printf("Failed to check admin_refresh_tokens: %v", err)
	} else {
		log.Printf("admin_refresh_tokens: %d records", adminTokenCount)
	}

	// 检查管理员登录会话表
	var adminSessionCount int
	err = database.DB.Get(&adminSessionCount, "SELECT COUNT(*) FROM admin_login_sessions")
	if err != nil {
		log.Printf("Failed to check admin_login_sessions: %v", err)
	} else {
		log.Printf("admin_login_sessions: %d records", adminSessionCount)
	}

	log.Println("Table data check completed!")
}