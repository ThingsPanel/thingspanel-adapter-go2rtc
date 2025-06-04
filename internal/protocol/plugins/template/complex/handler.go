package complex

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"tp-plugin/internal/protocol"

	"github.com/sirupsen/logrus"
)

// ComplexProtocolHandler 复杂协议处理器
// 适用于：网关协议、多消息类型、有状态协议
//
// 与简单协议的区别：
// - 需要自定义连接处理逻辑
// - 支持多种消息类型（心跳、数据、状态等）
// - 维护连接状态和会话信息
// - 处理连接认证/握手
//
// 开发步骤：
// 1. 复制此模板到 internal/protocol/plugins/your_protocol/ 目录
// 2. 实现基础协议接口（ParseData, ExtractDeviceID, EncodeCommand）
// 3. 实现增强协议接口（HandleConnection, OnConnectionEstablished, OnConnectionClosed）
// 4. 根据协议需求自定义连接处理逻辑
type ComplexProtocolHandler struct {
	port     int
	sessions map[string]*Session // 设备会话管理
	mu       sync.RWMutex
	logger   *logrus.Logger
}

// Session 设备会话信息
type Session struct {
	DeviceID      string
	Conn          net.Conn
	LastHeartbeat time.Time
	Authenticated bool
	MessageBuffer []byte
	Context       context.Context
	Cancel        context.CancelFunc
	Metadata      map[string]interface{}
}

// NewComplexProtocolHandler 创建复杂协议处理器
func NewComplexProtocolHandler(port int) *ComplexProtocolHandler {
	return &ComplexProtocolHandler{
		port:     port,
		sessions: make(map[string]*Session),
		logger:   logrus.StandardLogger(),
	}
}

// ============================================================================
// 基础协议接口实现 - 必须实现
// ============================================================================

func (h *ComplexProtocolHandler) Name() string {
	return "ComplexProtocol" // TODO: 修改为你的协议名称
}

func (h *ComplexProtocolHandler) Version() string {
	return "1.0.0" // TODO: 修改为你的协议版本
}

func (h *ComplexProtocolHandler) Port() int {
	return h.port
}

func (h *ComplexProtocolHandler) ParseData(data []byte) (*protocol.Message, error) {
	// TODO: 实现基础数据解析逻辑
	// 注意：复杂协议通常在HandleConnection中处理数据，这个方法可能不会被直接调用

	// 示例实现
	if len(data) < 6 {
		return nil, errors.New("数据包长度不足")
	}

	deviceID, err := h.ExtractDeviceNumber(data)
	if err != nil {
		return nil, err
	}

	// 基础数据解析
	messageData := map[string]interface{}{
		"raw_data": fmt.Sprintf("%x", data),
		"length":   len(data),
	}

	return &protocol.Message{
		DeviceID:    deviceID,
		MessageType: "data",
		Timestamp:   time.Now(),
		Data:        messageData,
		Quality:     1,
	}, nil
}

func (h *ComplexProtocolHandler) EncodeCommand(cmd *protocol.Command) ([]byte, error) {
	// TODO: 实现指令编码逻辑
	switch cmd.Action {
	case "heartbeat":
		return h.buildHeartbeatCommand(cmd.DeviceID)
	case "config":
		return h.buildConfigCommand(cmd.DeviceID, cmd.Parameters)
	case "query":
		return h.buildQueryCommand(cmd.DeviceID, cmd.Parameters)
	default:
		return nil, fmt.Errorf("不支持的指令: %s", cmd.Action)
	}
}

func (h *ComplexProtocolHandler) ExtractDeviceNumber(data []byte) (string, error) {
	// TODO: 根据你的协议格式提取设备编号
	if len(data) < 4 {
		return "", errors.New("数据包太短，无法提取设备编号")
	}

	// 示例：设备编号在数据包的前4个字节
	deviceNumber := binary.BigEndian.Uint32(data[0:4])
	return fmt.Sprintf("%08d", deviceNumber), nil
}

func (h *ComplexProtocolHandler) Start() error {
	h.logger.Infof("复杂协议 %s 启动，端口: %d", h.Name(), h.port)
	// TODO: 协议特定的初始化逻辑
	return nil
}

func (h *ComplexProtocolHandler) Stop() error {
	h.logger.Infof("复杂协议 %s 停止", h.Name())

	// 清理所有会话
	h.mu.Lock()
	for deviceID, session := range h.sessions {
		session.Cancel()
		session.Conn.Close()
		h.logger.Infof("会话已清理: %s", deviceID)
	}
	h.sessions = make(map[string]*Session)
	h.mu.Unlock()

	return nil
}

// ============================================================================
// 增强协议接口实现 - 复杂协议的核心
// ============================================================================

// HandleConnection 处理完整连接生命周期 - 这是复杂协议的核心方法
func (h *ComplexProtocolHandler) HandleConnection(conn net.Conn) error {
	defer conn.Close()

	// 1. 连接认证/握手
	deviceID, err := h.authenticateConnection(conn)
	if err != nil {
		h.logger.WithError(err).Error("连接认证失败")
		return err
	}

	// 2. 创建会话
	session := h.createSession(deviceID, conn)
	defer h.cleanupSession(deviceID)

	h.logger.Infof("设备连接已建立: %s (%s)", deviceID, conn.RemoteAddr())

	// 3. 连接处理循环
	buffer := make([]byte, 4096)
	for {
		select {
		case <-session.Context.Done():
			h.logger.Infof("会话已取消: %s", deviceID)
			return nil
		default:
			// 设置读取超时
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))

			n, err := conn.Read(buffer)
			if err != nil {
				if err != io.EOF {
					h.logger.WithError(err).Error("读取数据失败")
				}
				return err
			}

			data := buffer[:n]

			// 处理接收到的数据
			if err := h.handleReceivedData(session, data); err != nil {
				h.logger.WithError(err).Error("处理数据失败")
				// 不要因为单个数据包错误就断开连接
			}
		}
	}
}

// OnConnectionEstablished 连接建立时调用
func (h *ComplexProtocolHandler) OnConnectionEstablished(conn net.Conn) error {
	h.logger.Infof("新连接建立: %s", conn.RemoteAddr())
	// TODO: 连接建立时的特殊处理逻辑
	return nil
}

// OnConnectionClosed 连接关闭时调用
func (h *ComplexProtocolHandler) OnConnectionClosed(conn net.Conn) error {
	h.logger.Infof("连接已关闭: %s", conn.RemoteAddr())
	// TODO: 连接关闭时的清理逻辑
	return nil
}

// ============================================================================
// 私有方法 - 协议特定的实现逻辑
// ============================================================================

// authenticateConnection 连接认证/握手
func (h *ComplexProtocolHandler) authenticateConnection(conn net.Conn) (string, error) {
	// TODO: 实现协议特定的认证逻辑

	// 示例：简单的设备ID交换
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// 1. 读取认证数据包
	buffer := make([]byte, 256)
	n, err := conn.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("读取认证数据失败: %w", err)
	}

	authData := buffer[:n]

	// 2. 验证认证数据包格式
	if len(authData) < 8 {
		return "", errors.New("认证数据包格式错误")
	}

	// 3. 提取设备ID
	deviceID, err := h.ExtractDeviceNumber(authData)
	if err != nil {
		return "", fmt.Errorf("提取设备ID失败: %w", err)
	}

	// 4. 发送认证响应
	response := h.buildAuthResponse(deviceID, true)
	if _, err := conn.Write(response); err != nil {
		return "", fmt.Errorf("发送认证响应失败: %w", err)
	}

	return deviceID, nil
}

// createSession 创建设备会话
func (h *ComplexProtocolHandler) createSession(deviceID string, conn net.Conn) *Session {
	ctx, cancel := context.WithCancel(context.Background())

	session := &Session{
		DeviceID:      deviceID,
		Conn:          conn,
		LastHeartbeat: time.Now(),
		Authenticated: true,
		Context:       ctx,
		Cancel:        cancel,
		Metadata:      make(map[string]interface{}),
	}

	h.mu.Lock()
	h.sessions[deviceID] = session
	h.mu.Unlock()

	// 启动心跳检测
	go h.heartbeatMonitor(session)

	return session
}

// cleanupSession 清理会话
func (h *ComplexProtocolHandler) cleanupSession(deviceID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if session, exists := h.sessions[deviceID]; exists {
		session.Cancel()
		delete(h.sessions, deviceID)
		h.logger.Infof("会话已清理: %s", deviceID)
	}
}

// handleReceivedData 处理接收到的数据
func (h *ComplexProtocolHandler) handleReceivedData(session *Session, data []byte) error {
	// TODO: 根据协议实现数据处理逻辑

	// 示例：根据消息类型分发处理
	if len(data) < 2 {
		return errors.New("数据包太短")
	}

	messageType := data[1] // 假设第2个字节是消息类型

	switch messageType {
	case 0x01: // 心跳消息
		return h.handleHeartbeat(session, data)
	case 0x02: // 数据消息
		return h.handleDataMessage(session, data)
	case 0x03: // 状态消息
		return h.handleStatusMessage(session, data)
	default:
		return fmt.Errorf("未知消息类型: 0x%02x", messageType)
	}
}

// handleHeartbeat 处理心跳消息
func (h *ComplexProtocolHandler) handleHeartbeat(session *Session, data []byte) error {
	h.logger.Debugf("收到心跳: %s", session.DeviceID)

	// 更新心跳时间
	session.LastHeartbeat = time.Now()

	// 发送心跳响应
	response := h.buildHeartbeatResponse(session.DeviceID)
	if _, err := session.Conn.Write(response); err != nil {
		return fmt.Errorf("发送心跳响应失败: %w", err)
	}

	return nil
}

// handleDataMessage 处理数据消息
func (h *ComplexProtocolHandler) handleDataMessage(session *Session, data []byte) error {
	// TODO: 解析数据消息并发送到平台

	// 示例实现
	if len(data) < 10 {
		return errors.New("数据消息长度不足")
	}

	// 解析传感器数据（根据协议格式）
	temperature := float64(binary.BigEndian.Uint16(data[4:6])) / 10.0
	humidity := float64(binary.BigEndian.Uint16(data[6:8])) / 10.0

	sensorData := map[string]interface{}{
		"temperature": temperature,
		"humidity":    humidity,
		"timestamp":   time.Now().Unix(),
	}

	// 这里需要通过某种方式发送到平台
	// 在实际的HandleConnection实现中，应该有平台客户端的引用
	h.logger.Infof("设备数据: %s - %+v", session.DeviceID, sensorData)

	return nil
}

// handleStatusMessage 处理状态消息
func (h *ComplexProtocolHandler) handleStatusMessage(session *Session, data []byte) error {
	// TODO: 处理设备状态消息
	h.logger.Infof("设备状态: %s - %x", session.DeviceID, data)
	return nil
}

// heartbeatMonitor 心跳监控
func (h *ComplexProtocolHandler) heartbeatMonitor(session *Session) {
	ticker := time.NewTicker(30 * time.Second) // 30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-session.Context.Done():
			return
		case <-ticker.C:
			// 检查心跳超时
			if time.Since(session.LastHeartbeat) > 60*time.Second {
				h.logger.Warnf("设备心跳超时: %s", session.DeviceID)
				session.Cancel()
				return
			}
		}
	}
}

// ============================================================================
// 指令构建方法
// ============================================================================

// buildAuthResponse 构建认证响应
func (h *ComplexProtocolHandler) buildAuthResponse(deviceID string, success bool) []byte {
	// TODO: 根据协议格式构建认证响应
	response := make([]byte, 8)

	// 设备ID
	var id uint32
	fmt.Sscanf(deviceID, "%d", &id)
	binary.BigEndian.PutUint32(response[0:4], id)

	// 认证结果
	if success {
		response[4] = 0x01
	} else {
		response[4] = 0x00
	}

	return response
}

// buildHeartbeatCommand 构建心跳指令
func (h *ComplexProtocolHandler) buildHeartbeatCommand(deviceID string) ([]byte, error) {
	// TODO: 构建心跳指令
	cmd := make([]byte, 6)

	var id uint32
	_, err := fmt.Sscanf(deviceID, "%d", &id)
	if err != nil {
		return nil, err
	}

	binary.BigEndian.PutUint32(cmd[0:4], id)
	cmd[4] = 0x01 // 心跳指令类型

	return cmd, nil
}

// buildHeartbeatResponse 构建心跳响应
func (h *ComplexProtocolHandler) buildHeartbeatResponse(deviceID string) []byte {
	// TODO: 构建心跳响应
	response := make([]byte, 6)

	var id uint32
	fmt.Sscanf(deviceID, "%d", &id)
	binary.BigEndian.PutUint32(response[0:4], id)
	response[4] = 0x81 // 心跳响应类型

	return response
}

// buildConfigCommand 构建配置指令
func (h *ComplexProtocolHandler) buildConfigCommand(deviceID string, params interface{}) ([]byte, error) {
	// TODO: 构建配置指令
	return nil, errors.New("配置指令暂未实现")
}

// buildQueryCommand 构建查询指令
func (h *ComplexProtocolHandler) buildQueryCommand(deviceID string, params interface{}) ([]byte, error) {
	// TODO: 构建查询指令
	return nil, errors.New("查询指令暂未实现")
}

// ============================================================================
// 开发提示
// ============================================================================

/*
开发复杂协议的步骤：

1. 继承简单协议的所有实现
   - Name(), Version(), Port()
   - ParseData(), EncodeCommand(), ExtractDeviceID()
   - Start(), Stop()

2. 实现增强协议接口
   - HandleConnection(): 核心方法，处理完整连接生命周期
   - OnConnectionEstablished(): 连接建立时的处理
   - OnConnectionClosed(): 连接关闭时的清理

3. 设计会话管理
   - Session结构：存储连接状态、设备信息
   - 会话创建、更新、清理逻辑
   - 多连接并发安全

4. 实现协议特定逻辑
   - 连接认证/握手：authenticateConnection()
   - 消息类型分发：handleReceivedData()
   - 心跳机制：heartbeatMonitor()

5. 错误处理和资源管理
   - 超时处理：连接超时、心跳超时
   - 异常恢复：网络中断、数据异常
   - 资源清理：连接关闭、内存释放

注意事项：
- 复杂协议需要更多的状态管理和错误处理
- 并发安全：多个设备同时连接时的线程安全
- 内存管理：避免会话泄漏和内存泄漏
- 性能考虑：大量设备连接时的性能优化

使用场景：
- 网关协议：一个连接对应多个下级设备
- 有状态协议：需要维护连接状态和认证信息
- 多消息类型：心跳、数据、配置、控制等多种消息
- 复杂交互：需要应答确认、重传机制等
*/
