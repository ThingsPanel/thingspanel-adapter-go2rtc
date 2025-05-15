// internal/bootstrap/app.go
package bootstrap

import (
	"context"
	"tp-plugin/internal/config"

	"github.com/sirupsen/logrus"
)

// AppContext 应用程序上下文，包含所有运行时资源
type AppContext struct {
	Config         *config.Config
	PlatformClient interface{ Close() }
	ctx            context.Context
	cancel         context.CancelFunc
}

// Shutdown 关闭应用程序
func (app *AppContext) Shutdown() {
	if app.cancel != nil {
		app.cancel()
	}

	if app.PlatformClient != nil {
		app.PlatformClient.Close()
	}

	logrus.Info("应用资源已释放")
}

// StartApp 启动应用程序
func StartApp(configPath string) (*AppContext, error) {
	// 1. 加载配置
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// 2. 初始化日志系统
	if err := InitLogger(&cfg.Log); err != nil {
		return nil, err
	}

	// 3. 创建应用上下文
	ctx, cancel := context.WithCancel(context.Background())
	app := &AppContext{
		Config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	// 4. 初始化平台客户端
	platformClient, err := InitPlatformClient(&cfg.Platform)
	if err != nil {
		app.Shutdown()
		return nil, err
	}
	app.PlatformClient = platformClient

	// 5. 启动HTTP服务
	if err := StartHTTPServer(platformClient, cfg.Server.HTTPPort); err != nil {
		app.Shutdown()
		return nil, err
	}

	logrus.Info("服务就绪")
	return app, nil
}
