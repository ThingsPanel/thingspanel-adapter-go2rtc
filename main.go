package main

import (
	"log"
	httpclient "plugin_template/http_client"
	httpservice "plugin_template/http_service"
	"plugin_template/mqtt"
	"strings"

	"github.com/spf13/viper"
)

func main() {
	conf()
	LogInIt()
	log.Println("Starting the application...")
	// 启动mqtt客户端
	mqtt.InitClient()
	// 启动http客户端
	httpclient.Init()
	// 启动服务
	// go deviceconfig.Start()
	// 启动http服务
	httpservice.Init()
	// 订阅平台下发的消息
	mqtt.Subscribe()
	select {}
}
func conf() {
	log.Println("加载配置文件...")
	// 设置环境变量前缀
	viper.SetEnvPrefix("plugin_template")
	// 使 Viper 能够读取环境变量
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigType("yaml")
	viper.SetConfigFile("./config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err.Error())
	}
	log.Println("加载配置文件完成...")
}
