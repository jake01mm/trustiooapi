package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// InitLogger 初始化结构化日志
func InitLogger() {
	Log = logrus.New()

	// 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		logrus.Errorf("Failed to create logs directory: %v", err)
		return
	}

	// 创建日志文件
	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Errorf("Failed to open log file: %v", err)
		return
	}

	// 同时输出到控制台和文件
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	Log.SetOutput(multiWriter)

	// 设置日志格式为JSON
	Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFunc:  "caller",
		},
	})

	// 设置日志级别
	Log.SetLevel(logrus.InfoLevel)

	// 设置报告调用者
	Log.SetReportCaller(true)

	Log.Info("Logger initialized successfully")
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	if Log == nil {
		InitLogger()
	}
	return Log
}

// WithFields 创建带字段的日志条目
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

// WithField 创建带单个字段的日志条目
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

// WithError 创建带错误的日志条目
func WithError(err error) *logrus.Entry {
	return GetLogger().WithError(err)
}

// WithRequest 创建带请求信息的日志条目
func WithRequest(method, path, ip string) *logrus.Entry {
	return GetLogger().WithFields(logrus.Fields{
		"method": method,
		"path":   path,
		"ip":     ip,
	})
}

// Info 记录信息日志
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Warn 记录警告日志
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Error 记录错误日志
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Fatal 记录致命错误日志并退出
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Debug 记录调试日志
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Infof 格式化信息日志
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warnf 格式化警告日志
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Errorf 格式化错误日志
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatalf 格式化致命错误日志并退出
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// Debugf 格式化调试日志
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// 自动轮转日志文件（可选实现）
func RotateLogFile() error {
	logPath := "logs/app.log"
	
	// 检查文件大小，如果超过10MB则轮转
	if info, err := os.Stat(logPath); err == nil {
		if info.Size() > 10*1024*1024 { // 10MB
			// 重命名当前日志文件
			newName := "logs/app_" + info.ModTime().Format("20060102_150405") + ".log"
			if err := os.Rename(logPath, newName); err != nil {
				return err
			}
			
			// 重新初始化日志
			InitLogger()
		}
	}
	
	return nil
}

// CleanupOldLogs 清理旧日志文件
func CleanupOldLogs() error {
	logsDir := "logs"
	
	return filepath.Walk(logsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// 删除7天前的日志文件
		if info.ModTime().Before(time.Now().AddDate(0, 0, -7)) && 
		   filepath.Ext(path) == ".log" && 
		   path != "logs/app.log" {
			return os.Remove(path)
		}
		
		return nil
	})
}