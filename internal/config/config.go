// internal/config/config.go
package config

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Platform PlatformConfig `mapstructure:"platform"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port             int `mapstructure:"port"`
	HTTPPort         int `mapstructure:"http_port"`
	HeartbeatTimeout int `mapstructure:"heartbeatTimeout"`
}

type PlatformConfig struct {
	URL               string `mapstructure:"url"`           // 平台API地址
	MQTTBroker        string `mapstructure:"mqtt_broker"`   // MQTT服务器地址
	MQTTUsername      string `mapstructure:"mqtt_username"` // MQTT用户名
	MQTTPassword      string `mapstructure:"mqtt_password"` // MQTT密码
	ServiceIdentifier string `mapstructure:"service_identifier"`
	TemplateSecret    string `mapstructure:"template_secret"` // 模板密钥，用于动态注册
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	FilePath   string `mapstructure:"filePath"`
	EnableFile bool   `mapstructure:"enableFile"` // 是否启用文件日志
	MaxSize    int    `mapstructure:"maxSize"`    // 每个日志文件的最大大小（MB）
	MaxBackups int    `mapstructure:"maxBackups"` // 保留的旧日志文件的最大数量
	MaxAge     int    `mapstructure:"maxAge"`     // 保留日志文件的最大天数
	Compress   bool   `mapstructure:"compress"`   // 是否压缩旧日志文件
}
