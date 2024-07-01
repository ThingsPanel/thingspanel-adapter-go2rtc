package sub

import (
	"encoding/json"
	"strings"

	"plugin_template/mqtt"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// 订阅
func SubscribeCommands() {
	// 主题
	topic := viper.GetString("mqtt.command_topic_to_subscribe")
	qos := viper.GetUint("mqtt.qos")
	// 订阅主题
	if err := mqtt.MqttClient.Subscribe(topic, commandMessageHandler, uint8(qos)); err != nil {
		logrus.Infof("订阅主题失败: %v", err)
	}
	logrus.Info("订阅主题成功:", topic)

}

// 设备下发消息的回调函数,payload示例{"method":"FindAnimal","params":{"count":2,"animalType":"cat"}}
// 主题示例service/alarm/devices/command/{device_id}/{message_id}，其中device_id为设备ID，message_id为消息ID
// 设备下发消息的回调函数
func commandMessageHandler(client MQTT.Client, msg MQTT.Message) {
	logrus.Debugf("收到消息: %s", msg.Payload())

	// 解析主题
	topicParts := strings.Split(msg.Topic(), "/")
	if len(topicParts) < 6 {
		logrus.Warnf("无效的主题格式: %s", msg.Topic())
		return
	}

	// 解析payload
	var payload struct {
		Method string                 `json:"method"`
		Params map[string]interface{} `json:"params"`
	}
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		logrus.Warnf("解析消息payload失败: %v", err)
		return
	}

	logrus.Infof("收到命令: %+v", payload)

}
