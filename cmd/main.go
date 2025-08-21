package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"trusioo_api/config"
	"trusioo_api/internal/router"
	"trusioo_api/pkg/database"
	"trusioo_api/pkg/logger"
	"trusioo_api/pkg/redis"
)

func main() {
	// 初始化日志
	logger.InitLogger()
	logger.Info("Starting Trusioo API...")

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}
	logger.Info("Configuration loaded successfully")

	// 初始化数据库
	if err := database.InitDatabase(); err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDatabase()

	// 初始化Redis
	if err := redis.InitRedis(); err != nil {
		logger.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer redis.CloseRedis()

	// 设置路由
	r := router.SetupRouter()

	// 创建 HTTP 服务器
	port := config.AppConfig.Server.Port
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
		// 安全相关的超时设置
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		// 设置最大请求头大小
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// 在 goroutine 中启动服务器
	go func() {
		logger.Infof("Server starting on port %s", port)
		logger.Infof("Server URL: http://localhost:%s", port)
		logger.Info("Press Ctrl+C to gracefully shutdown the server")
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 设置优雅关闭
	gracefulShutdown(srv)
}

// gracefulShutdown 优雅关闭服务器
func gracefulShutdown(srv *http.Server) {
	// 创建一个接收系统信号的通道
	quit := make(chan os.Signal, 1)
	
	// 监听中断信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// 等待信号
	sig := <-quit
	logger.Infof("Received signal: %v. Initiating graceful shutdown...", sig)

	// 设置关闭超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	// 清理资源
	logger.Info("Cleaning up resources...")
	
	// 关闭数据库连接
	database.CloseDatabase()
	
	// 关闭Redis连接
	if err := redis.CloseRedis(); err != nil {
		logger.Errorf("Failed to close Redis connection: %v", err)
	}
	
	// 清理日志
	if err := logger.RotateLogFile(); err != nil {
		logger.Errorf("Failed to rotate log file: %v", err)
	}

	logger.Info("Server gracefully stopped")
}