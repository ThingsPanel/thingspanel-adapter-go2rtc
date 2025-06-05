package simple

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"tp-plugin/internal/protocol" // 使用正确的模块路径

	"github.com/sirupsen/logrus"
)

// SimpleProtocolHandler 简单协议处理器
// 适用于：传感器数据采集、简单设备控制等无状态协议
//
// 开发步骤：
// 1. 复制此模板到 internal/protocol/plugins/your_protocol/ 目录
// 2. 根据你的协议格式实现 ExtractDeviceNumber 和 ParseData 方法
// 3. 如果设备支持控制指令，实现 EncodeCommand 方法
// 4. 在 bootstrap/app.go 中初始化你的协议
type SimpleProtocolHandler struct {
	port int
}

// NewSimpleProtocolHandler 创建简单协议处理器
func NewSimpleProtocolHandler(port int) *SimpleProtocolHandler {
	return &SimpleProtocolHandler{port: port}
}

// ============================================================================
// 基本信息 - 必须实现
// ============================================================================

func (h *SimpleProtocolHandler) Name() string {
	return "SimpleProtocol" // TODO: 修改为你的协议名称
}

func (h *SimpleProtocolHandler) Version() string {
	return "1.0.0" // TODO: 修改为你的协议版本
}

func (h *SimpleProtocolHandler) Port() int {
	return h.port
}

// ============================================================================
// 核心方法1：提取设备编号 - 必须实现！
// ============================================================================

// ExtractDeviceNumber 从数据包中提取设备编号（重要：不是平台的device_id）
func (h *SimpleProtocolHandler) ExtractDeviceNumber(data []byte) (string, error) {
	// TODO: 根据你的协议格式提取设备编号

	// 示例：设备编号在数据包的前4个字节
	if len(data) < 4 {
		return "", errors.New("数据包太短，无法提取设备编号")
	}

	// 方式1：二进制格式的设备编号
	deviceNumber := binary.BigEndian.Uint32(data[0:4])
	return fmt.Sprintf("%d", deviceNumber), nil

	// 方式2：字符串格式的设备编号（如果协议使用字符串）
	// if len(data) < 8 {
	//     return "", errors.New("数据包太短，无法提取设备编号")
	// }
	// return string(data[0:8]), nil

	// 方式3：从特定位置提取设备编号
	// deviceNumberBytes := data[startPos:endPos]
	// return string(deviceNumberBytes), nil
}

// ============================================================================
// 核心方法2：解析数据 - 必须实现！
// ============================================================================

func (h *SimpleProtocolHandler) ParseData(data []byte) (*protocol.Message, error) {
	// TODO: 根据你的协议格式实现数据解析逻辑

	// 1. 数据包基本验证
	if len(data) < 6 {
		return nil, errors.New("数据包长度不足")
	}

	// 2. 提取设备编号（注意：这里提取的是设备编号，不是平台的device_id）
	deviceNumber, err := h.ExtractDeviceNumber(data)
	if err != nil {
		return nil, fmt.Errorf("提取设备编号失败: %w", err)
	}

	// 3. 根据协议格式解析数据
	// TODO: 根据你的协议实现数据解析
	sensorData := map[string]interface{}{
		"temperature": 25.6,                    // 示例数据
		"humidity":    60.2,                    // 示例数据
		"voltage":     3.7,                     // 示例数据
		"raw_data":    fmt.Sprintf("%x", data), // 原始数据（十六进制）
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

// ============================================================================
// 核心方法3：编码指令 - 如果设备不支持控制，返回错误即可
// ============================================================================

func (h *SimpleProtocolHandler) EncodeCommand(cmd *protocol.Command) ([]byte, error) {
	// TODO: 如果你的设备不支持控制指令，返回错误即可
	// return nil, errors.New("设备不支持控制指令")

	// TODO: 根据你的协议格式实现指令编码
	switch cmd.Action {
	case "sleep":
		// 示例：休眠指令
		return h.buildSleepCommand(cmd.DeviceNumber, cmd.Parameters)
	case "config":
		// 示例：配置指令
		return h.buildConfigCommand(cmd.DeviceNumber, cmd.Parameters)
	case "query":
		// 示例：查询指令
		return h.buildQueryCommand(cmd.DeviceNumber)
	default:
		return nil, fmt.Errorf("不支持的指令: %s", cmd.Action)
	}
}

// ============================================================================
// 生命周期管理 - 通常只需要打印日志
// ============================================================================

func (h *SimpleProtocolHandler) Start() error {
	logrus.Infof("协议 %s 启动，端口: %d", h.Name(), h.port)
	// TODO: 如果需要额外的初始化工作，在这里实现
	return nil
}

func (h *SimpleProtocolHandler) Stop() error {
	logrus.Infof("协议 %s 停止", h.Name())
	// TODO: 如果需要清理资源，在这里实现
	return nil
}

// ============================================================================
// 私有辅助方法 - 根据需要实现
// ============================================================================

// buildSleepCommand 构建休眠指令
func (h *SimpleProtocolHandler) buildSleepCommand(deviceNumber string, params interface{}) ([]byte, error) {
	// TODO: 实现休眠指令构建逻辑
	// 示例代码：
	cmd := make([]byte, 8)

	// 解析设备编号
	id, err := parseDeviceNumber(deviceNumber)
	if err != nil {
		return nil, err
	}

	// 设备编号 (4字节)
	binary.BigEndian.PutUint32(cmd[0:4], id)

	// 指令类型：休眠 (1字节)
	cmd[4] = 0x01

	// 休眠时间 (2字节) - 从参数中获取
	sleepTime := uint16(3600) // 默认1小时
	if params != nil {
		if p, ok := params.(map[string]interface{}); ok {
			if t, ok := p["sleep_time"].(float64); ok {
				sleepTime = uint16(t)
			}
		}
	}
	binary.BigEndian.PutUint16(cmd[5:7], sleepTime)

	// 校验和 (1字节)
	cmd[7] = calculateChecksum(cmd[0:7])

	return cmd, nil
}

// buildConfigCommand 构建配置指令
func (h *SimpleProtocolHandler) buildConfigCommand(deviceNumber string, params interface{}) ([]byte, error) {
	// TODO: 实现配置指令构建逻辑
	return nil, errors.New("配置指令暂未实现")
}

// buildQueryCommand 构建查询指令
func (h *SimpleProtocolHandler) buildQueryCommand(deviceNumber string) ([]byte, error) {
	// TODO: 实现查询指令构建逻辑
	return nil, errors.New("查询指令暂未实现")
}

// parseDeviceNumber 解析设备编号字符串为数字
func parseDeviceNumber(deviceNumber string) (uint32, error) {
	// TODO: 根据你的设备编号格式实现解析逻辑
	var id uint32
	_, err := fmt.Sscanf(deviceNumber, "%d", &id)
	return id, err
}

// calculateChecksum 计算校验和
func calculateChecksum(data []byte) byte {
	// TODO: 根据你的协议实现校验和算法
	var sum byte
	for _, b := range data {
		sum ^= b // 示例：简单异或校验
	}
	return sum
}

// ============================================================================
// 开发提示
// ============================================================================

/*
开发简单协议的步骤：

1. 修改基本信息
   - Name(): 协议名称
   - Version(): 协议版本
   - Port(): 协议端口

2. 实现ExtractDeviceNumber方法
   - 这是最重要的方法，必须能从数据包中提取设备编号
   - 提取失败时，设备无法上线

3. 实现ParseData方法
   - 根据协议格式解析数据包
   - 返回标准的Message结构
   - MessageType通常设为"data"

4. 实现EncodeCommand方法（可选）
   - 如果设备不支持控制，返回错误即可
   - 如果支持，根据cmd.Action构建相应指令

5. 在bootstrap/app.go中初始化协议：
   ```go
   // 创建你的协议处理器
   protocolHandler := your_protocol.NewHandler(cfg.Server.Port)

   // 创建单协议处理器
   singleHandler := protocol.NewSingleProtocolHandler(
       protocolHandler,
       app.PlatformClient,
       logrus.StandardLogger(),
   )

   // 启动协议
   singleHandler.Start()
   ```

6. 在配置文件中设置协议端口：
   ```yaml
   server:
     port: 15001  # 协议端口
   ```

注意事项：
- 设备编号提取失败时，设备无法正常上线
- 数据解析失败时，只会记录日志，不会断开连接
- 框架会自动处理设备上下线通知
- 框架会自动发送数据到ThingsPanel平台
- 一个中间件只处理一个协议，配置简单
*/
