package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// FileHook 文件钩子，用于同时写入文件
type FileHook struct {
	fileLogger *logrus.Logger
}

// Levels 返回支持的日志级别
func (hook *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 执行钩子
func (hook *FileHook) Fire(entry *logrus.Entry) error {
	// 将日志条目写入文件
	hook.fileLogger.WithFields(entry.Data).Log(entry.Level, entry.Message)
	return nil
}

// PrettyFormatter 自定义美观的日志格式化器
type PrettyFormatter struct {
	TimestampFormat string
	Colorize        bool
}

// Format 实现 logrus.Formatter 接口
func (f *PrettyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format(f.TimestampFormat)
	message := entry.Message

	// 获取调用者信息
	caller := ""
	if entry.HasCaller() {
		fileParts := strings.Split(entry.Caller.File, "/")
		if len(fileParts) > 0 {
			caller = fmt.Sprintf("%s:%d", fileParts[len(fileParts)-1], entry.Caller.Line)
		}
	}

	// 颜色和图标配置
	var levelIcon, levelColor, resetColor string
	if f.Colorize {
		resetColor = "\033[0m"
		switch entry.Level {
		case logrus.DebugLevel:
			levelIcon = "🔍"
			levelColor = "\033[36m" // 青色
		case logrus.InfoLevel:
			levelIcon = "ℹ️"
			levelColor = "\033[32m" // 绿色
		case logrus.WarnLevel:
			levelIcon = "⚠️"
			levelColor = "\033[33m" // 黄色
		case logrus.ErrorLevel:
			levelIcon = "❌"
			levelColor = "\033[31m" // 红色
		case logrus.FatalLevel, logrus.PanicLevel:
			levelIcon = "💀"
			levelColor = "\033[35m" // 紫色
		}
	} else {
		switch entry.Level {
		case logrus.DebugLevel:
			levelIcon = "[DEBUG]"
		case logrus.InfoLevel:
			levelIcon = "[INFO]"
		case logrus.WarnLevel:
			levelIcon = "[WARN]"
		case logrus.ErrorLevel:
			levelIcon = "[ERROR]"
		case logrus.FatalLevel, logrus.PanicLevel:
			levelIcon = "[FATAL]"
		}
	}

	// 构建日志行
	var logLine string
	if f.Colorize {
		logLine = fmt.Sprintf("%s%s %s%s %s%s %s%s\n",
			"\033[90m", timestamp, // 灰色时间戳
			levelColor, levelIcon,
			"\033[94m", message, // 蓝色消息
			resetColor, caller)
	} else {
		logLine = fmt.Sprintf("%s %s %s [%s]\n", timestamp, levelIcon, message, caller)
	}

	// 添加字段信息
	if len(entry.Data) > 0 {
		for key, value := range entry.Data {
			if f.Colorize {
				logLine += fmt.Sprintf("  \033[96m%s\033[0m: %v\n", key, value) // 青色字段名
			} else {
				logLine += fmt.Sprintf("  %s: %v\n", key, value)
			}
		}
	}

	return []byte(logLine), nil
}

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

	// 为控制台和文件设置不同的格式化器
	// 控制台使用美观格式（带颜色）
	consoleFormatter := &PrettyFormatter{
		TimestampFormat: "15:04:05",
		Colorize:        true,
	}

	// 文件使用JSON格式（便于日志分析）
	fileFormatter := &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFunc:  "caller",
		},
	}

	// 创建自定义的多重写入器
	consoleLogger := logrus.New()
	consoleLogger.SetOutput(os.Stdout)
	consoleLogger.SetFormatter(consoleFormatter)
	consoleLogger.SetLevel(logrus.InfoLevel)
	consoleLogger.SetReportCaller(true)

	fileLogger := logrus.New()
	fileLogger.SetOutput(logFile)
	fileLogger.SetFormatter(fileFormatter)
	fileLogger.SetLevel(logrus.InfoLevel)
	fileLogger.SetReportCaller(true)

	// 设置主日志器使用控制台格式
	Log.SetOutput(os.Stdout)
	Log.SetFormatter(consoleFormatter)
	Log.SetLevel(logrus.InfoLevel)
	Log.SetReportCaller(true)

	// 添加文件钩子
	Log.AddHook(&FileHook{fileLogger: fileLogger})

	Log.Info("🚀 Trusioo API Logger initialized successfully")
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