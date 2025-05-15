package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"tp-plugin/internal/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ANSI颜色码
const (
	colorRed    = 31
	colorGreen  = 32
	colorYellow = 33
	colorBlue   = 36
	colorGray   = 37
)

type CustomFormatter struct {
	logrus.TextFormatter
	IsTerminal bool
}

func getColorByLevel(level logrus.Level) int {
	switch level {
	case logrus.ErrorLevel:
		return colorRed
	case logrus.WarnLevel:
		return colorYellow
	case logrus.InfoLevel:
		return colorGreen
	case logrus.DebugLevel:
		return colorBlue
	default:
		return colorGray
	}
}

// 添加颜色包装
func colored(color int, text string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, text)
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")

	// 处理文件路径 - 改为只保留相对路径
	var filePath string
	if entry.Caller != nil {
		file := entry.Caller.File
		// 优先查找项目顶层目录
		if idx := strings.LastIndex(file, "tp-plugin"); idx != -1 {
			// 从tp-plugin开始截取
			filePath = file[idx:]
		} else if idx := strings.Index(file, "internal"); idx != -1 {
			// 备选：从internal开始截取
			filePath = file[idx:]
		} else if idx := strings.Index(file, "cmd"); idx != -1 {
			// 备选：从cmd开始截取
			filePath = file[idx:]
		} else {
			// 如果都找不到，保留原始路径但使用最短的部分
			parts := strings.Split(file, string(os.PathSeparator))
			if len(parts) >= 2 {
				filePath = strings.Join(parts[len(parts)-2:], string(os.PathSeparator))
			} else {
				filePath = file
			}
		}
	}

	// 获取level对应的颜色
	levelColor := getColorByLevel(entry.Level)

	// 构建日志级别部分（带颜色）
	levelText := strings.ToUpper(entry.Level.String())
	if f.IsTerminal {
		levelText = colored(levelColor, levelText)
	}

	// 构建时间戳部分（使用灰色）
	timeText := timestamp
	if f.IsTerminal {
		timeText = colored(colorGray, timestamp)
	}

	// 构建文件路径部分（使用蓝色）
	fileInfo := fmt.Sprintf("%s:%d", filePath, entry.Caller.Line)
	if f.IsTerminal {
		fileInfo = colored(colorBlue, fileInfo)
	}

	// 构建字段信息
	var fields string
	if len(entry.Data) > 0 {
		parts := make([]string, 0, len(entry.Data))
		for k, v := range entry.Data {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		fields = strings.Join(parts, " ")
	}

	// 将文件路径放到消息的末尾
	var logMessage string
	if len(fields) > 0 {
		logMessage = fmt.Sprintf(
			"%s[%s] %s | %s | [%s]",
			levelText,     // 带颜色的日志级别
			timeText,      // 带颜色的时间戳
			entry.Message, // 原始消息
			fields,        // 字段信息
			fileInfo,      // 放到末尾的文件信息
		)
	} else {
		logMessage = fmt.Sprintf(
			"%s[%s] %s [%s]",
			levelText,     // 带颜色的日志级别
			timeText,      // 带颜色的时间戳
			entry.Message, // 原始消息
			fileInfo,      // 放到末尾的文件信息
		)
	}

	return []byte(logMessage + "\n"), nil
}

// InitLogger 初始化日志系统
func InitLogger(cfg *config.LogConfig) {
	var logOutput io.Writer = os.Stdout

	// 如果启用了文件日志，则配置文件输出
	if cfg.EnableFile {
		// 创建文件日志写入器
		fileLogger := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}

		// 创建多重输出 - 同时输出到控制台和文件
		logOutput = io.MultiWriter(os.Stdout, fileLogger)

		// 简化日志输出
		logrus.WithField("path", cfg.FilePath).Info("日志文件已启用")
	} else {
		logrus.Info("日志仅输出到控制台")
	}

	// 设置日志输出
	logrus.SetOutput(logOutput)

	// 启用调用者信息报告
	logrus.SetReportCaller(true)

	// 设置自定义格式化器
	logrus.SetFormatter(&CustomFormatter{
		IsTerminal: true, // 启用终端颜色支持
	})

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
		logrus.Warnf("无效的日志级别: %s, 使用默认级别: INFO", cfg.Level)
	}
	logrus.SetLevel(level)
}
