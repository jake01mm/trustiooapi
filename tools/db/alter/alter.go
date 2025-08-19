package main

import (
	"log"
	"os"

	"trusioo_api/config"
	"trusioo_api/pkg/database"
)

func main() {
	log.Println("Starting database table alteration...")

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDatabase()

	// 读取 SQL 文件
	sqlContent, err := os.ReadFile("scripts/alter_db.sql")
	if err != nil {
		log.Fatalf("Failed to read SQL file: %v", err)
	}

	// 直接执行整个 SQL 文件
	log.Println("Executing database schema alteration...")
	_, err = database.DB.Exec(string(sqlContent))
	if err != nil {
		log.Fatalf("Failed to execute alteration: %v", err)
	}

	log.Println("Database table alteration completed successfully!")
}
