// cmd/main.go
package main

import (
	"fmt"
	"os"
	"tp-plugin/internal/bootstrap"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	// 设置日志初始格式
	bootstrap.SetupInitialLogger()

	logrus.Info("=================== Demo插件服务启动 ===================")

	app := &cli.App{
		Name:    "tp-plugin",
		Usage:   "tp-plugin Demo protocol plugin",
		Version: "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "./configs/config.yaml",
				Usage:   "config file path",
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Fatal("程序运行失败")
	}
}

func run(c *cli.Context) error {
	// 获取配置文件路径
	configPath := c.String("config")

	// 启动应用
	appContext, err := bootstrap.StartApp(configPath)
	if err != nil {
		return fmt.Errorf("启动应用失败: %v", err)
	}

	// 确保应用退出时清理资源
	defer appContext.Shutdown()

	// 阻塞主goroutine，等待信号
	select {}
}
