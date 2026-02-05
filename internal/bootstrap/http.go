// internal/bootstrap/http.go
package bootstrap

import (
	"fmt"
	"tp-plugin/internal/handler"
	"tp-plugin/internal/platform"
	"tp-plugin/internal/protocol"

	"github.com/sirupsen/logrus"
)

// StartHTTPServer 启动HTTP服务
func StartHTTPServer(platformClient *platform.PlatformClient, httpPort int, ph protocol.ProtocolHandler) error {
	// 创建HTTP处理器
	httpHandler := handler.NewHTTPHandler(platformClient, logrus.StandardLogger(), ph)
	handlers := httpHandler.RegisterHandlers()

	// 启动HTTP服务
	go func() {
		logrus.Infof("启动HTTP服务 [:%d]", httpPort)
		if err := handlers.Start(fmt.Sprintf(":%d", httpPort)); err != nil {
			logrus.Errorf("HTTP服务启动失败: %v", err)
		}
	}()

	return nil
}
