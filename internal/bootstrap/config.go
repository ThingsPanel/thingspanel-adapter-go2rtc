// internal/bootstrap/config.go
package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"tp-plugin/internal/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*config.Config, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取并解析配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"port":            cfg.Server.Port,
		"max_connections": cfg.Server.MaxConnections,
		"heartbeat":       cfg.Server.HeartbeatTimeout,
		"log_level":       cfg.Log.Level,
		"log_path":        cfg.Log.FilePath,
		"enable_file_log": cfg.Log.EnableFile,
	}).Info("配置加载成功")

	return &cfg, nil
}

// EnsureLogDir 确保日志目录存在
func EnsureLogDir(logPath string) error {
	dir := filepath.Dir(logPath)
	return os.MkdirAll(dir, 0755)
}
