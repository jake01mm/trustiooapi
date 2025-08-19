package main

import (
	"log"

	"trusioo_api/config"
	"trusioo_api/pkg/database"
)

func main() {
	log.Println("Testing database connection...")

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDatabase()

	log.Println("Database connection successful!")

	// 测试简单查询
	var version string
	err := database.DB.Get(&version, "SELECT version()")
	if err != nil {
		log.Fatalf("Failed to query database: %v", err)
	}

	log.Printf("PostgreSQL version: %s", version)
}