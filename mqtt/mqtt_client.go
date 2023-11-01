package mqtt

import (
	"log"

	tpprotocolsdkgo "github.com/ThingsPanel/tp-protocol-sdk-go"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"
)

var MqttClient *tpprotocolsdkgo.MQTTClient

func InitClient() {
	log.Println("创建mqtt客户端")
	// 创建新的MQTT客户端实例
	addr := viper.GetString("mqtt.broker")
	username := viper.GetString("mqtt.username")
	password := viper.GetString("mqtt.password")
	client := tpprotocolsdkgo.NewMQTTClient(addr, username, password)
	// 尝试连接到MQTT代理
	if err := client.Connect(); err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	log.Println("连接成功")
	MqttClient = client
}

// 发布设备消息{"token":device_token,"values":{sub_device_addr1:{key:value...},sub_device_add2r:{key:value...}}}
func Publish(payload string) error {
	// 主题
	topic := viper.GetString("mqtt.topic_to_publish")
	qos := viper.GetUint("mqtt.qos")
	// 发布消息
	if err := MqttClient.Publish(topic, string(payload), uint8(qos)); err != nil {
		log.Printf("发布消息失败: %v", err)
		return err
	}
	log.Println("发布消息成功:", payload, "主题:", topic)
	return nil
}

// 订阅
func Subscribe() {
	// 主题
	topic := viper.GetString("mqtt.topic_to_subscribe")
	qos := viper.GetUint("mqtt.qos")
	// 订阅主题
	if err := MqttClient.Subscribe(topic, messageHandler, uint8(qos)); err != nil {
		log.Printf("订阅主题失败: %v", err)
	}
	log.Println("订阅主题成功:", topic)

}

// 设备下发消息的回调函数：主题plugin/modbus/# payload：{sub_device_addr:{key:value...},sub_device_addr:{key:value...}}
func messageHandler(client MQTT.Client, msg MQTT.Message) {
	log.Printf("收到消息: %s", msg.Payload())
}
