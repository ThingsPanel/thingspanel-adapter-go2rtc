package examples

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"tp-plugin/internal/protocol"

	"github.com/sirupsen/logrus"
)

// SensorProtocolHandler 传感器协议处理器
// 这是一个具体的协议实现示例，基于简单协议模板
//
// 协议格式：
// 数据包: [4字节设备ID][2字节温度][2字节湿度][2字节电压][1字节电量][1字节校验]
// 设备ID: 大端序32位整数
// 温度: 大端序16位整数，单位0.1°C，范围-40.0~85.0°C
// 湿度: 大端序16位整数，单位0.1%，范围0.0~100.0%
// 电压: 大端序16位整数，单位0.01V，范围0.00~5.00V
// 电量: 8位整数，单位1%，范围0~100%
// 校验: 前面所有字节的异或校验
type SensorProtocolHandler struct {
	port int
}

// NewSensorProtocolHandler 创建传感器协议处理器
func NewSensorProtocolHandler(port int) *SensorProtocolHandler {
	return &SensorProtocolHandler{port: port}
}

// ============================================================================
// 基本信息
// ============================================================================

func (h *SensorProtocolHandler) Name() string {
	return "SensorProtocol"
}

func (h *SensorProtocolHandler) Version() string {
	return "1.0.0"
}

func (h *SensorProtocolHandler) Port() int {
	return h.port
}

// ============================================================================
// 核心方法实现
// ============================================================================

func (h *SensorProtocolHandler) ParseData(data []byte) (*protocol.Message, error) {
	// 验证数据包长度
	if len(data) != 12 {
		return nil, fmt.Errorf("数据包长度错误，期望12字节，实际%d字节", len(data))
	}

	// 验证校验和
	calculatedChecksum := h.calculateChecksum(data[:11])
	if calculatedChecksum != data[11] {
		return nil, fmt.Errorf("校验和错误，期望0x%02x，实际0x%02x", calculatedChecksum, data[11])
	}

	// 提取设备编号（注意：这里提取的是设备编号，不是平台的device_id）
	deviceNumber, err := h.ExtractDeviceNumber(data)
	if err != nil {
		return nil, fmt.Errorf("提取设备编号失败: %w", err)
	}

	// 解析传感器数据
	temperature := float64(int16(binary.BigEndian.Uint16(data[4:6]))) / 10.0 // 有符号温度
	humidity := float64(binary.BigEndian.Uint16(data[6:8])) / 10.0           // 湿度
	voltage := float64(binary.BigEndian.Uint16(data[8:10])) / 100.0          // 电压
	battery := int(data[10])                                                 // 电量百分比

	// 数据验证
	if temperature < -40.0 || temperature > 85.0 {
		return nil, fmt.Errorf("温度值超出范围: %.1f°C", temperature)
	}
	if humidity < 0.0 || humidity > 100.0 {
		return nil, fmt.Errorf("湿度值超出范围: %.1f%%", humidity)
	}
	if voltage < 0.0 || voltage > 5.0 {
		return nil, fmt.Errorf("电压值超出范围: %.2fV", voltage)
	}
	if battery < 0 || battery > 100 {
		return nil, fmt.Errorf("电量值超出范围: %d%%", battery)
	}

	// 构造消息
	sensorData := map[string]interface{}{
		"temperature": temperature,
		"humidity":    humidity,
		"voltage":     voltage,
		"battery":     battery,
	}

	return &protocol.Message{
		DeviceNumber: deviceNumber, // 设备编号（从设备提取）
		DeviceID:     "",           // 设备ID（由框架根据设备编号从平台获取）
		MessageType:  "data",
		Timestamp:    time.Now(),
		Data:         sensorData,
		Quality:      1,
	}, nil
}

func (h *SensorProtocolHandler) EncodeCommand(cmd *protocol.Command) ([]byte, error) {
	switch cmd.Action {
	case "sleep":
		return h.buildSleepCommand(cmd.DeviceNumber, cmd.Parameters)
	case "config":
		return h.buildConfigCommand(cmd.DeviceNumber, cmd.Parameters)
	case "query":
		return h.buildQueryCommand(cmd.DeviceNumber)
	default:
		return nil, fmt.Errorf("不支持的指令: %s", cmd.Action)
	}
}

// ExtractDeviceNumber 从数据包中提取设备编号（重要：不是平台的device_id）
func (h *SensorProtocolHandler) ExtractDeviceNumber(data []byte) (string, error) {
	if len(data) < 4 {
		return "", errors.New("数据包太短，无法提取设备编号")
	}

	// 从数据包前4字节提取设备编号（大端序）
	deviceNumber := binary.BigEndian.Uint32(data[0:4])
	return fmt.Sprintf("%08d", deviceNumber), nil
}

func (h *SensorProtocolHandler) Start() error {
	logrus.Infof("传感器协议启动，端口: %d", h.port)
	return nil
}

func (h *SensorProtocolHandler) Stop() error {
	logrus.Info("传感器协议停止")
	return nil
}

// ============================================================================
// 私有辅助方法
// ============================================================================

// buildSleepCommand 构建休眠指令
// 指令格式: [4字节设备编号][1字节指令类型0x01][2字节休眠时间(分钟)][1字节校验]
func (h *SensorProtocolHandler) buildSleepCommand(deviceNumber string, params interface{}) ([]byte, error) {
	cmd := make([]byte, 8)

	// 解析设备编号
	id, err := h.parseDeviceNumber(deviceNumber)
	if err != nil {
		return nil, err
	}

	// 设备编号 (4字节)
	binary.BigEndian.PutUint32(cmd[0:4], id)

	// 指令类型：休眠 (1字节)
	cmd[4] = 0x01

	// 休眠时间 (2字节，单位：分钟)
	sleepTime := uint16(60) // 默认60分钟
	if params != nil {
		if p, ok := params.(map[string]interface{}); ok {
			if t, ok := p["sleep_minutes"].(float64); ok {
				sleepTime = uint16(t)
			}
		}
	}
	binary.BigEndian.PutUint16(cmd[5:7], sleepTime)

	// 校验和 (1字节)
	cmd[7] = h.calculateChecksum(cmd[0:7])

	return cmd, nil
}

// buildConfigCommand 构建配置指令
// 指令格式: [4字节设备编号][1字节指令类型0x02][2字节上报间隔(秒)][1字节校验]
func (h *SensorProtocolHandler) buildConfigCommand(deviceNumber string, params interface{}) ([]byte, error) {
	cmd := make([]byte, 8)

	// 解析设备编号
	id, err := h.parseDeviceNumber(deviceNumber)
	if err != nil {
		return nil, err
	}

	// 设备编号 (4字节)
	binary.BigEndian.PutUint32(cmd[0:4], id)

	// 指令类型：配置 (1字节)
	cmd[4] = 0x02

	// 上报间隔 (2字节，单位：秒)
	interval := uint16(300) // 默认300秒
	if params != nil {
		if p, ok := params.(map[string]interface{}); ok {
			if i, ok := p["report_interval"].(float64); ok {
				interval = uint16(i)
			}
		}
	}
	binary.BigEndian.PutUint16(cmd[5:7], interval)

	// 校验和 (1字节)
	cmd[7] = h.calculateChecksum(cmd[0:7])

	return cmd, nil
}

// buildQueryCommand 构建查询指令
// 指令格式: [4字节设备编号][1字节指令类型0x03][2字节预留][1字节校验]
func (h *SensorProtocolHandler) buildQueryCommand(deviceNumber string) ([]byte, error) {
	cmd := make([]byte, 8)

	// 解析设备编号
	id, err := h.parseDeviceNumber(deviceNumber)
	if err != nil {
		return nil, err
	}

	// 设备编号 (4字节)
	binary.BigEndian.PutUint32(cmd[0:4], id)

	// 指令类型：查询 (1字节)
	cmd[4] = 0x03

	// 预留字段 (2字节)
	binary.BigEndian.PutUint16(cmd[5:7], 0x0000)

	// 校验和 (1字节)
	cmd[7] = h.calculateChecksum(cmd[0:7])

	return cmd, nil
}

// parseDeviceNumber 解析设备编号字符串为数值
func (h *SensorProtocolHandler) parseDeviceNumber(deviceNumber string) (uint32, error) {
	// 解析设备编号字符串（例如："00000001" -> 1）
	var id uint32
	_, err := fmt.Sscanf(deviceNumber, "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("无效的设备编号格式: %s", deviceNumber)
	}
	return id, nil
}

// calculateChecksum 计算异或校验和
func (h *SensorProtocolHandler) calculateChecksum(data []byte) byte {
	var checksum byte
	for _, b := range data {
		checksum ^= b
	}
	return checksum
}

// ============================================================================
// 协议说明
// ============================================================================

/*
传感器协议说明：

1. 数据包格式（12字节）：
   字节0-3: 设备ID（大端序32位整数）
   字节4-5: 温度（大端序16位有符号整数，单位0.1°C）
   字节6-7: 湿度（大端序16位整数，单位0.1%）
   字节8-9: 电压（大端序16位整数，单位0.01V）
   字节10:  电量（8位整数单位1%）
   字节11:  校验和（前11字节异或）

2. 指令格式（8字节）：
   字节0-3: 设备ID（大端序32位整数）
   字节4:   指令类型（0x01=休眠，0x02=配置，0x03=查询）
   字节5-6: 参数（根据指令类型而定）
   字节7:   校验和（前7字节异或）

3. 数据范围：
   - 温度: -40.0°C ~ 85.0°C
   - 湿度: 0.0% ~ 100.0%
   - 电压: 0.00V ~ 5.00V
   - 电量: 0% ~ 100%

4. 使用示例：
   ```go
   // 在bootstrap/app.go中注册
   if cfg.Protocols.SensorProtocol.Enabled {
       handler := examples.NewSensorProtocolHandler(cfg.Protocols.SensorProtocol.Port)
       manager.RegisterProtocol(handler)
   }
   ```

5. 配置示例：
   ```yaml
   protocols:
     sensor_protocol:
       enabled: true
       port: 15001
   ```

6. 测试数据包示例：
   设备ID=1, 温度=25.6°C, 湿度=60.5%, 电压=3.30V, 电量=85%
   原始数据: 00000001 0100 025D 014A 0055 A8

   解析：
   - 00000001: 设备ID=1
   - 0100: 温度=256 -> 25.6°C
   - 025D: 湿度=605 -> 60.5%
   - 014A: 电压=330 -> 3.30V
   - 55: 电量=85%
   - A8: 校验和
*/
