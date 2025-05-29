// internal/bootstrap/app.go
package bootstrap

import (
	"context"
	"time"
	"tp-plugin/internal/config"
	"tp-plugin/internal/platform"

	"github.com/sirupsen/logrus"
)

// AppContext 应用程序上下文，包含所有运行时资源
type AppContext struct {
	Config          *config.Config
	PlatformClient  *platform.PlatformClient
	ctx             context.Context
	cancel          context.CancelFunc
	heartbeatTicker *time.Ticker // 心跳定时器
}

// Shutdown 关闭应用程序
func (app *AppContext) Shutdown() {
	if app.cancel != nil {
		app.cancel()
	}

	if app.PlatformClient != nil {
		app.PlatformClient.Close()
	}

	// 停止心跳定时器
	if app.heartbeatTicker != nil {
		app.heartbeatTicker.Stop()
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

	// 5. 启动心跳定时器 - 每分钟发送一次心跳
	app.heartbeatTicker = time.NewTicker(time.Second * time.Duration(cfg.Server.HeartbeatTimeout))
	go func() {
		// 启动时立即发送一次心跳
		if err := platformClient.SendHeartbeat(ctx, cfg.Platform.ServiceIdentifier); err != nil {
			logrus.WithError(err).Error("首次发送心跳失败")
		} else {
			logrus.Info("首次心跳发送成功")
		}

		// 定时发送心跳
		for {
			select {
			case <-app.heartbeatTicker.C:
				if err := platformClient.SendHeartbeat(ctx, cfg.Platform.ServiceIdentifier); err != nil {
					logrus.WithError(err).Error("发送心跳失败")
				} else {
					logrus.Debug("心跳发送成功")
				}
			case <-ctx.Done():
				logrus.Info("心跳上报已停止")
				return
			}
		}
	}()

	// 6. 初始化OG-SZ501协议处理组件
	if err := initializeOGSZ501Protocol(app, cfg); err != nil {
		app.Shutdown()
		return nil, err
	}

	// 7. 启动HTTP服务
	if err := StartHTTPServer(platformClient, cfg.Server.HTTPPort); err != nil {
		app.Shutdown()
		return nil, err
	}

	logrus.Info("服务就绪")
	return app, nil
}

// initializeOGSZ501Protocol 初始化OG-SZ501协议处理组件
func initializeOGSZ501Protocol(app *AppContext, cfg *config.Config) error {
	// 不要创建新logger，使用全局配置好的logger
	// logger := logrus.New()

	// 传递全局logger
	// tcpServer := transport.NewTCPServer(cfg.Server.Port, logrus.StandardLogger(), app.PlatformClient)
	// app.TCPServer = tcpServer

	// // 2. 创建设备处理器
	// deviceHandler := handler.NewDeviceHandler(tcpServer, app.PlatformClient, logrus.StandardLogger())
	// app.DeviceHandler = deviceHandler

	// 3. 设置数据包处理回调 - 这里建立了两者的联系
	// tcpServer.SetPacketHandler(deviceHandler.HandleData)

	// // 4. 启动TCP服务器
	// if err := tcpServer.Start(); err != nil {
	// 	return err
	// }

	logrus.Infof("OG-SZ501协议服务器已启动，监听端口: %d", cfg.Server.Port)
	return nil
}
