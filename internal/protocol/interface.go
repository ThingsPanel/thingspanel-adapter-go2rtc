package protocol

import (
	"net"
	"time"
)

// Message 设备消息结构
type Message struct {
	DeviceNumber string                 `json:"device_number"` // 设备编号(从设备中提取)
	DeviceID     string                 `json:"device_id"`     // 设备ID(平台分配的ID)
	MessageType  string                 `json:"message_type"`  // data/heartbeat/status
	Timestamp    time.Time              `json:"timestamp"`
	Data         map[string]interface{} `json:"data"`    // 设备数据
	Quality      int                    `json:"quality"` // 数据质量，1=正常
}

// Command 设备指令结构
type Command struct {
	DeviceNumber string        `json:"device_number"` // 设备编号(从设备中提取)
	DeviceID     string        `json:"device_id"`     // 设备ID(平台分配的ID)
	CommandID    string        `json:"command_id"`
	Action       string        `json:"action"` // sleep/config/query等
	Parameters   interface{}   `json:"parameters"`
	Timeout      time.Duration `json:"timeout"`
}

// ProtocolHandler 基础协议处理器接口
// 适用于90%的无状态、简单数据交互协议
type ProtocolHandler interface {
	// 协议基本信息
	Name() string
	Version() string
	Port() int // 协议专用端口

	// 核心功能 - 只需要这两个方法！
	ParseData(data []byte) (*Message, error)    // 解析设备数据
	EncodeCommand(cmd *Command) ([]byte, error) // 编码控制指令

	// 设备编号提取（从数据包中提取设备编号，注意：不是平台的device_id）
	ExtractDeviceNumber(data []byte) (string, error) // 从数据包提取设备编号

	// 生命周期管理（通常只需要返回nil）
	Start() error // 启动协议
	Stop() error  // 停止协议
}

// EnhancedProtocolHandler 增强协议处理器接口
// 适用于有状态、多消息类型的复杂协议（5%的情况）
type EnhancedProtocolHandler interface {
	ProtocolHandler // 嵌入基础接口

	// 自定义连接处理（替代默认的简单处理逻辑）
	HandleConnection(conn net.Conn) error // 处理完整连接生命周期

	// 连接事件处理（可选实现）
	OnConnectionEstablished(conn net.Conn) error // 连接建立时调用
	OnConnectionClosed(conn net.Conn) error      // 连接关闭时调用
}

// ProtocolInfo 协议信息
type ProtocolInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Port    int    `json:"port"`
	Status  string `json:"status"` // running/stopped/error
}
