package main

import (
	"log"

	"trusioo_api/config"
	"trusioo_api/internal/router"
	"trusioo_api/pkg/database"
	"trusioo_api/pkg/logger"
)

func main() {
	// 初始化日志
	logger.InitLogger()
	log.Println("Starting Trusioo API...")

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Println("Configuration loaded successfully")

	// 初始化数据库
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDatabase()

	// 设置路由
	r := router.SetupRouter()

	// 启动服务器
	port := config.AppConfig.Server.Port
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}