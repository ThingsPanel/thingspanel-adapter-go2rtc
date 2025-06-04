package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// DeviceLogger 设备日志管理器
// 为每个设备创建独立的日志文件，文件名使用设备编号（device_number）
type DeviceLogger struct {
	loggers  map[string]*logrus.Logger // 设备编号 -> logger映射
	mu       sync.RWMutex
	baseDir  string // 日志基础目录
	enabled  bool   // 是否启用设备日志
	maxSize  int    // 单个日志文件最大大小(MB)
	maxAge   int    // 日志文件保留天数
	compress bool   // 是否压缩旧日志
}

// DeviceLoggerConfig 设备日志配置
type DeviceLoggerConfig struct {
	Enabled  bool   `mapstructure:"enabled"`  // 是否启用设备独立日志
	BaseDir  string `mapstructure:"base_dir"` // 设备日志基础目录
	MaxSize  int    `mapstructure:"max_size"` // 单个设备日志文件最大大小(MB)
	MaxAge   int    `mapstructure:"max_age"`  // 设备日志文件保留天数
	Compress bool   `mapstructure:"compress"` // 是否压缩设备日志文件
}

// globalDeviceLogger 全局设备日志管理器实例
var globalDeviceLogger *DeviceLogger
var deviceLoggerOnce sync.Once

// InitDeviceLogger 初始化设备日志管理器
func InitDeviceLogger(config DeviceLoggerConfig) error {
	var initErr error

	deviceLoggerOnce.Do(func() {
		if !config.Enabled {
			logrus.Info("设备独立日志功能已禁用")
			globalDeviceLogger = &DeviceLogger{enabled: false}
			return
		}

		// 设置默认值
		if config.BaseDir == "" {
			config.BaseDir = "logs/devices"
		}
		if config.MaxSize <= 0 {
			config.MaxSize = 10 // 默认10MB
		}
		if config.MaxAge <= 0 {
			config.MaxAge = 7 // 默认保留7天
		}

		// 确保日志目录存在
		if err := os.MkdirAll(config.BaseDir, 0755); err != nil {
			initErr = fmt.Errorf("创建设备日志目录失败: %v", err)
			return
		}

		globalDeviceLogger = &DeviceLogger{
			loggers:  make(map[string]*logrus.Logger),
			baseDir:  config.BaseDir,
			enabled:  true,
			maxSize:  config.MaxSize,
			maxAge:   config.MaxAge,
			compress: config.Compress,
		}

		logrus.WithField("base_dir", config.BaseDir).Info("设备独立日志管理器已启用")
	})

	return initErr
}

// GetDeviceLogger 获取设备日志记录器
// deviceNumber: 设备编号（从设备数据中提取的唯一标识符）
func GetDeviceLogger(deviceNumber string) *logrus.Logger {
	if globalDeviceLogger == nil || !globalDeviceLogger.enabled {
		return logrus.StandardLogger() // 返回标准logger作为fallback
	}

	return globalDeviceLogger.getOrCreateLogger(deviceNumber)
}

// getOrCreateLogger 获取或创建设备专用logger
func (dl *DeviceLogger) getOrCreateLogger(deviceNumber string) *logrus.Logger {
	dl.mu.RLock()
	if logger, exists := dl.loggers[deviceNumber]; exists {
		dl.mu.RUnlock()
		return logger
	}
	dl.mu.RUnlock()

	// 需要创建新的logger
	dl.mu.Lock()
	defer dl.mu.Unlock()

	// 双重检查，防止并发创建
	if logger, exists := dl.loggers[deviceNumber]; exists {
		return logger
	}

	// 创建设备专用日志文件路径（使用设备编号作为文件名）
	logFilePath := filepath.Join(dl.baseDir, fmt.Sprintf("%s.log", deviceNumber))

	// 创建lumberjack轮转写入器
	fileWriter := &lumberjack.Logger{
		Filename:  logFilePath,
		MaxSize:   dl.maxSize,
		MaxAge:    dl.maxAge,
		Compress:  dl.compress,
		LocalTime: true,
	}

	// 创建新的logger实例
	deviceLogger := logrus.New()

	// 设置输出 - 只输出到文件，不输出到控制台
	deviceLogger.SetOutput(fileWriter)

	// 设置日志级别
	deviceLogger.SetLevel(logrus.DebugLevel)

	// 设置格式化器 - 专门为设备日志优化
	deviceLogger.SetFormatter(&DeviceLogFormatter{
		DeviceNumber: deviceNumber,
	})

	// 不启用调用者信息，减少日志开销
	deviceLogger.SetReportCaller(false)

	// 缓存logger
	dl.loggers[deviceNumber] = deviceLogger

	logrus.WithFields(logrus.Fields{
		"device_number": deviceNumber,
		"log_file":      logFilePath,
	}).Debug("设备日志记录器已创建")

	return deviceLogger
}

// DeviceLogFormatter 设备日志专用格式化器
type DeviceLogFormatter struct {
	DeviceNumber string
}

// Format 格式化日志消息
func (f *DeviceLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	level := entry.Level.String()

	// 构建字段信息
	var fields string
	if len(entry.Data) > 0 {
		fieldParts := make([]string, 0, len(entry.Data))
		for k, v := range entry.Data {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		fields = " | " + fmt.Sprintf("[%s]", joinStrings(fieldParts, " "))
	}

	// 格式: [时间] [级别] [设备ID] 消息 [字段]
	logLine := fmt.Sprintf("[%s] [%s] [%s] %s%s\n",
		timestamp,
		level,
		f.DeviceNumber,
		entry.Message,
		fields,
	)

	return []byte(logLine), nil
}

// joinStrings 连接字符串切片
func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}

// LogDeviceData 记录设备数据
// deviceNumber: 设备编号（从设备数据中提取的唯一标识符）
func LogDeviceData(deviceNumber string, direction string, data []byte, extraInfo map[string]interface{}) {
	if globalDeviceLogger == nil || !globalDeviceLogger.enabled {
		return
	}

	logger := GetDeviceLogger(deviceNumber)
	fields := logrus.Fields{
		"direction": direction, // "received" 或 "sent"
		"data_len":  len(data),
		"data_hex":  fmt.Sprintf("%x", data),
	}

	// 添加额外信息
	for k, v := range extraInfo {
		fields[k] = v
	}

	logger.WithFields(fields).Info("设备数据交互")
}

// LogDeviceEvent 记录设备事件
// deviceNumber: 设备编号（从设备数据中提取的唯一标识符）
func LogDeviceEvent(deviceNumber string, event string, details map[string]interface{}) {
	if globalDeviceLogger == nil || !globalDeviceLogger.enabled {
		return
	}

	logger := GetDeviceLogger(deviceNumber)
	fields := logrus.Fields{
		"event": event,
	}

	// 添加详细信息
	for k, v := range details {
		fields[k] = v
	}

	logger.WithFields(fields).Info("设备事件")
}

// LogDeviceCommand 记录设备指令
// deviceNumber: 设备编号（从设备数据中提取的唯一标识符）
func LogDeviceCommand(deviceNumber string, command string, params interface{}, result interface{}) {
	if globalDeviceLogger == nil || !globalDeviceLogger.enabled {
		return
	}

	logger := GetDeviceLogger(deviceNumber)
	logger.WithFields(logrus.Fields{
		"command": command,
		"params":  params,
		"result":  result,
	}).Info("设备指令")
}

// LogDeviceStatus 记录设备状态变化
// deviceNumber: 设备编号（从设备数据中提取的唯一标识符）
func LogDeviceStatus(deviceNumber string, status string, details map[string]interface{}) {
	if globalDeviceLogger == nil || !globalDeviceLogger.enabled {
		return
	}

	logger := GetDeviceLogger(deviceNumber)
	fields := logrus.Fields{
		"status": status,
	}

	// 添加详细信息
	for k, v := range details {
		fields[k] = v
	}

	logger.WithFields(fields).Info("设备状态")
}

// CleanupDeviceLogger 清理指定设备的日志记录器
// deviceNumber: 设备编号（从设备数据中提取的唯一标识符）
func CleanupDeviceLogger(deviceNumber string) {
	if globalDeviceLogger == nil || !globalDeviceLogger.enabled {
		return
	}

	globalDeviceLogger.mu.Lock()
	defer globalDeviceLogger.mu.Unlock()

	if logger, exists := globalDeviceLogger.loggers[deviceNumber]; exists {
		// 关闭日志文件
		if writer, ok := logger.Out.(*lumberjack.Logger); ok {
			writer.Close()
		}
		delete(globalDeviceLogger.loggers, deviceNumber)

		logrus.WithField("device_number", deviceNumber).Debug("设备日志记录器已清理")
	}
}

// GetDeviceLogStats 获取设备日志统计信息
func GetDeviceLogStats() map[string]interface{} {
	if globalDeviceLogger == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	globalDeviceLogger.mu.RLock()
	defer globalDeviceLogger.mu.RUnlock()

	return map[string]interface{}{
		"enabled":        globalDeviceLogger.enabled,
		"base_dir":       globalDeviceLogger.baseDir,
		"active_loggers": len(globalDeviceLogger.loggers),
		"max_size":       globalDeviceLogger.maxSize,
		"max_age":        globalDeviceLogger.maxAge,
		"compress":       globalDeviceLogger.compress,
	}
}
