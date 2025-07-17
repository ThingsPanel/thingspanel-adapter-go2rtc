module tp-plugin

go 1.22

require (
	github.com/spf13/viper v1.20.1
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.5 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
)

require (
	github.com/eclipse/paho.mqtt.golang v1.4.3 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
)

require (
	github.com/ThingsPanel/tp-protocol-sdk-go v1.2.6
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sirupsen/logrus v1.9.3
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/urfave/cli/v2 v2.27.5
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// 本地开发时使用本地SDK包，取消下面这行的注释
//replace github.com/ThingsPanel/tp-protocol-sdk-go => D:\B-work\thingspanel-work\1-code\1-tp\tp-protocol-sdk-go

// 切换回线上包时，重新注释上面这行
