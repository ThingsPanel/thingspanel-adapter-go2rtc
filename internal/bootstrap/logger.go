// internal/bootstrap/logger.go
package bootstrap

import (
	"fmt"
	"tp-plugin/internal/config"
	"tp-plugin/internal/pkg/logger"

	"github.com/sirupsen/logrus"
)

// InitLogger 初始化日志系统
func InitLogger(cfg *config.LogConfig) error {
	// 只有当启用文件日志时才创建日志目录
	if cfg.EnableFile {
		// 确保日志目录存在
		if err := EnsureLogDir(cfg.FilePath); err != nil {
			logrus.WithError(err).Error("创建日志目录失败")
			return fmt.Errorf("创建日志目录失败: %v", err)
		}
	}

	// 初始化日志配置
	logger.InitLogger(cfg)
	return nil
}

// SetupInitialLogger 配置初始日志格式
func SetupInitialLogger() {
	// 设置在启动初期就使用自定义日志格式
	customFormatter := &logger.CustomFormatter{
		TextFormatter: logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		},
		IsTerminal: true,
	}

	// 启用调用者信息从一开始就追踪文件位置
	logrus.SetReportCaller(true)
	logrus.SetFormatter(customFormatter)
}
