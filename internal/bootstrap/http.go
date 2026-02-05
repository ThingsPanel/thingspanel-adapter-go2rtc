// internal/bootstrap/http.go
package bootstrap

import (
	"fmt"
	"net/http"
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

		// 创建自定义处理器来处理URL重写
		mux := http.NewServeMux()

		// 注册根处理器，并在其中处理URL重写
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// URL重写: /api/v1/notify/event -> /api/v1/plugin/notification
			if r.URL.Path == "/api/v1/notify/event" {
				logrus.Debugf("URL Rewriting: %s -> /api/v1/plugin/notification", r.URL.Path)
				r.URL.Path = "/api/v1/plugin/notification"
			}

			// 调用SDK处理器
			handlers.ServeHTTP(w, r)
		})

		if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), mux); err != nil {
			logrus.Errorf("HTTP服务启动失败: %v", err)
		}
	}()

	return nil
}
