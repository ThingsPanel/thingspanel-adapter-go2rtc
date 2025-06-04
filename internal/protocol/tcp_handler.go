package protocol

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
	"tp-plugin/internal/pkg/logger"

	"github.com/ThingsPanel/tp-protocol-sdk-go/types"
	"github.com/sirupsen/logrus"
)

// PlatformInterface 平台接口，避免循环依赖
type PlatformInterface interface {
	SendTelemetry(deviceID string, values map[string]interface{}) error
	SendDeviceStatus(deviceID string, status int) error
	GetDevice(deviceNumber string) (*types.Device, error)
}

// TCPHandler TCP连接处理器 - 框架提供，协议开发者无需关心
type TCPHandler struct {
	port        int
	handler     ProtocolHandler
	platform    PlatformInterface
	logger      *logrus.Logger
	deviceCache map[net.Conn]string // 连接到设备编号的映射
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	listener    net.Listener
}

// NewTCPHandler 创建TCP处理器
func NewTCPHandler(port int, handler ProtocolHandler, platform PlatformInterface, logger *logrus.Logger) *TCPHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &TCPHandler{
		port:        port,
		handler:     handler,
		platform:    platform,
		logger:      logger,
		deviceCache: make(map[net.Conn]string),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动TCP服务器
func (h *TCPHandler) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", h.port))
	if err != nil {
		return fmt.Errorf("启动TCP服务器失败: %w", err)
	}

	h.listener = listener
	h.logger.Infof("协议 %s 在端口 %d 启动成功", h.handler.Name(), h.port)

	go h.acceptConnections()
	return nil
}

// acceptConnections 接受连接
func (h *TCPHandler) acceptConnections() {
	for {
		select {
		case <-h.ctx.Done():
			return
		default:
			conn, err := h.listener.Accept()
			if err != nil {
				if h.ctx.Err() != nil {
					return // 正在关闭
				}
				h.logger.WithError(err).Error("接受连接失败")
				continue
			}

			// 为每个连接启动处理协程
			go h.handleConnection(conn)
		}
	}
}

// handleConnection 处理连接
func (h *TCPHandler) handleConnection(conn net.Conn) {
	// 检查是否为增强协议
	if enhanced, ok := h.handler.(EnhancedProtocolHandler); ok {
		// 使用增强协议的自定义连接处理
		if err := enhanced.OnConnectionEstablished(conn); err != nil {
			h.logger.WithError(err).Error("连接建立事件处理失败")
		}

		if err := enhanced.HandleConnection(conn); err != nil {
			h.logger.WithError(err).Error("增强协议连接处理失败")
		}

		if err := enhanced.OnConnectionClosed(conn); err != nil {
			h.logger.WithError(err).Error("连接关闭事件处理失败")
		}
	} else {
		// 使用默认的简单协议处理
		h.handleSimpleProtocol(conn)
	}
}

// handleSimpleProtocol 简单协议的默认处理逻辑
func (h *TCPHandler) handleSimpleProtocol(conn net.Conn) {
	defer func() {
		h.notifyDeviceOffline(conn)
		conn.Close()
	}()

	// 设置连接超时
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	buffer := make([]byte, 4096)
	var deviceNumber string
	var deviceID string
	deviceNumberExtracted := false

	for {
		// 重置读取超时
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				h.logger.WithError(err).Error("读取连接数据失败")
				// 记录连接错误到设备日志
				if deviceNumberExtracted {
					logger.LogDeviceEvent(deviceNumber, "connection_error", map[string]interface{}{
						"error":       err.Error(),
						"remote_addr": conn.RemoteAddr().String(),
					})
				}
			}
			break
		}

		data := buffer[:n]

		// 首次提取设备编号
		if !deviceNumberExtracted {
			extractedNumber, err := h.handler.ExtractDeviceNumber(data)
			if err != nil {
				h.logger.WithError(err).Debug("提取设备编号失败，继续尝试")
				// 记录设备编号提取失败的原始数据
				logger.LogDeviceData("unknown", "received", data, map[string]interface{}{
					"extract_error": err.Error(),
					"remote_addr":   conn.RemoteAddr().String(),
					"protocol":      h.handler.Name(),
				})
				continue
			}
			deviceNumber = extractedNumber
			deviceNumberExtracted = true

			// 获取设备信息，包括平台分配的device_id
			device, err := h.platform.GetDevice(deviceNumber)
			if err != nil {
				h.logger.WithError(err).Error("获取设备信息失败")
				// 记录获取设备信息失败
				logger.LogDeviceEvent(deviceNumber, "get_device_failed", map[string]interface{}{
					"error":         err.Error(),
					"device_number": deviceNumber,
					"protocol":      h.handler.Name(),
				})
				break
			}
			deviceID = device.ID

			// 缓存设备编号
			h.mu.Lock()
			h.deviceCache[conn] = deviceNumber
			h.mu.Unlock()

			// 记录设备连接建立事件
			logger.LogDeviceEvent(deviceNumber, "connection_established", map[string]interface{}{
				"remote_addr":   conn.RemoteAddr().String(),
				"protocol":      h.handler.Name(),
				"port":          h.port,
				"device_id":     deviceID,
				"device_number": deviceNumber,
			})

			// 发送设备上线通知
			h.notifyDeviceOnline(conn, deviceNumber, deviceID)
		}

		// 记录接收到的设备数据
		logger.LogDeviceData(deviceNumber, "received", data, map[string]interface{}{
			"remote_addr": conn.RemoteAddr().String(),
			"protocol":    h.handler.Name(),
		})

		// 调用协议的ParseData方法
		message, err := h.handler.ParseData(data)
		if err != nil {
			h.logger.WithError(err).Debug("解析数据失败")
			// 记录数据解析失败事件
			logger.LogDeviceEvent(deviceNumber, "parse_data_failed", map[string]interface{}{
				"error":    err.Error(),
				"data_hex": fmt.Sprintf("%x", data),
				"protocol": h.handler.Name(),
			})
			continue
		}

		// 确保消息包含正确的设备信息
		if message.DeviceNumber == "" {
			message.DeviceNumber = deviceNumber
		}
		if message.DeviceID == "" {
			message.DeviceID = deviceID
		}

		// 记录解析成功的消息内容
		logger.LogDeviceEvent(deviceNumber, "data_parsed", map[string]interface{}{
			"message_type": message.MessageType,
			"data_fields":  len(message.Data),
			"quality":      message.Quality,
			"protocol":     h.handler.Name(),
		})

		// 自动发送到ThingsPanel平台
		if err := h.sendToPlatform(message); err != nil {
			h.logger.WithError(err).Error("发送数据到平台失败")
			// 记录平台发送失败事件
			logger.LogDeviceEvent(deviceNumber, "platform_send_failed", map[string]interface{}{
				"error":        err.Error(),
				"message_type": message.MessageType,
				"protocol":     h.handler.Name(),
			})
		} else {
			// 记录平台发送成功事件
			logger.LogDeviceEvent(deviceNumber, "platform_sent", map[string]interface{}{
				"message_type": message.MessageType,
				"data_fields":  len(message.Data),
				"protocol":     h.handler.Name(),
			})
		}
	}
}

// notifyDeviceOnline 设备上线通知
func (h *TCPHandler) notifyDeviceOnline(conn net.Conn, deviceNumber string, deviceID string) {
	// 记录设备上线状态变化
	logger.LogDeviceStatus(deviceNumber, "online", map[string]interface{}{
		"remote_addr":   conn.RemoteAddr().String(),
		"protocol":      h.handler.Name(),
		"port":          h.port,
		"device_id":     deviceID,
		"device_number": deviceNumber,
	})

	// 发送设备上线状态（使用平台分配的device_id）
	if err := h.platform.SendDeviceStatus(deviceID, 1); err != nil {
		h.logger.WithError(err).Error("发送设备上线状态失败")
		// 记录状态发送失败事件
		logger.LogDeviceEvent(deviceNumber, "status_send_failed", map[string]interface{}{
			"error":         err.Error(),
			"status":        "online",
			"protocol":      h.handler.Name(),
			"device_id":     deviceID,
			"device_number": deviceNumber,
		})
	} else {
		h.logger.Infof("设备上线: %s (设备ID: %s, 地址: %s) - 协议: %s", deviceNumber, deviceID, conn.RemoteAddr(), h.handler.Name())
		// 记录状态发送成功事件
		logger.LogDeviceEvent(deviceNumber, "status_sent", map[string]interface{}{
			"status":        "online",
			"protocol":      h.handler.Name(),
			"device_id":     deviceID,
			"device_number": deviceNumber,
		})
	}
}

// notifyDeviceOffline 设备下线通知
func (h *TCPHandler) notifyDeviceOffline(conn net.Conn) {
	h.mu.RLock()
	deviceNumber, exists := h.deviceCache[conn]
	h.mu.RUnlock()

	if !exists {
		return // 设备编号未提取，无需发送下线通知
	}

	// 获取设备信息以获取device_id
	device, err := h.platform.GetDevice(deviceNumber)
	var deviceID string
	if err != nil {
		h.logger.WithError(err).Warn("获取设备信息失败，使用设备编号作为ID")
		deviceID = deviceNumber // 降级处理
	} else {
		deviceID = device.ID
	}

	// 记录设备连接断开事件
	logger.LogDeviceEvent(deviceNumber, "connection_closed", map[string]interface{}{
		"remote_addr":   conn.RemoteAddr().String(),
		"protocol":      h.handler.Name(),
		"device_id":     deviceID,
		"device_number": deviceNumber,
	})

	// 清理缓存
	h.mu.Lock()
	delete(h.deviceCache, conn)
	h.mu.Unlock()

	// 记录设备下线状态变化
	logger.LogDeviceStatus(deviceNumber, "offline", map[string]interface{}{
		"remote_addr":   conn.RemoteAddr().String(),
		"protocol":      h.handler.Name(),
		"device_id":     deviceID,
		"device_number": deviceNumber,
	})

	// 发送设备下线状态（使用平台分配的device_id）
	if err := h.platform.SendDeviceStatus(deviceID, 0); err != nil {
		h.logger.WithError(err).Error("发送设备下线状态失败")
		// 记录状态发送失败事件
		logger.LogDeviceEvent(deviceNumber, "status_send_failed", map[string]interface{}{
			"error":         err.Error(),
			"status":        "offline",
			"protocol":      h.handler.Name(),
			"device_id":     deviceID,
			"device_number": deviceNumber,
		})
	} else {
		h.logger.Infof("设备下线: %s (设备ID: %s, 地址: %s) - 协议: %s", deviceNumber, deviceID, conn.RemoteAddr(), h.handler.Name())
		// 记录状态发送成功事件
		logger.LogDeviceEvent(deviceNumber, "status_sent", map[string]interface{}{
			"status":        "offline",
			"protocol":      h.handler.Name(),
			"device_id":     deviceID,
			"device_number": deviceNumber,
		})
	}

	// 清理设备日志记录器(可选，根据需要决定是否立即清理)
	// logger.CleanupDeviceLogger(deviceNumber)
}

// sendToPlatform 发送数据到平台
func (h *TCPHandler) sendToPlatform(message *Message) error {
	// 根据消息类型发送到不同的主题
	switch message.MessageType {
	case "data":
		// 发送遥测数据（使用平台分配的device_id）
		return h.platform.SendTelemetry(message.DeviceID, message.Data)
	case "heartbeat", "status":
		// 发送状态数据 (需要在平台接口中添加)
		// 暂时也发送到遥测数据
		return h.platform.SendTelemetry(message.DeviceID, message.Data)
	default:
		h.logger.Warnf("未知消息类型: %s", message.MessageType)
		return h.platform.SendTelemetry(message.DeviceID, message.Data)
	}
}

// SendCommand 发送指令到设备
func (h *TCPHandler) SendCommand(deviceNumber string, cmd *Command) error {
	// 记录指令发送开始
	logger.LogDeviceCommand(deviceNumber, cmd.Action, cmd.Parameters, "sending")

	// 查找设备连接
	h.mu.RLock()
	var targetConn net.Conn
	for conn, cachedDeviceNumber := range h.deviceCache {
		if cachedDeviceNumber == deviceNumber {
			targetConn = conn
			break
		}
	}
	h.mu.RUnlock()

	if targetConn == nil {
		err := fmt.Errorf("设备 %s 未连接", deviceNumber)
		// 记录指令发送失败
		logger.LogDeviceCommand(deviceNumber, cmd.Action, cmd.Parameters, map[string]interface{}{
			"result": "failed",
			"error":  err.Error(),
		})
		return err
	}

	// 编码指令
	data, err := h.handler.EncodeCommand(cmd)
	if err != nil {
		// 记录指令编码失败
		logger.LogDeviceCommand(deviceNumber, cmd.Action, cmd.Parameters, map[string]interface{}{
			"result": "encode_failed",
			"error":  err.Error(),
		})
		return fmt.Errorf("编码指令失败: %w", err)
	}

	// 记录发送的指令数据
	logger.LogDeviceData(deviceNumber, "sent", data, map[string]interface{}{
		"command":     cmd.Action,
		"parameters":  cmd.Parameters,
		"remote_addr": targetConn.RemoteAddr().String(),
		"protocol":    h.handler.Name(),
	})

	// 发送指令
	if _, err := targetConn.Write(data); err != nil {
		// 记录指令发送失败
		logger.LogDeviceCommand(deviceNumber, cmd.Action, cmd.Parameters, map[string]interface{}{
			"result": "send_failed",
			"error":  err.Error(),
		})
		return fmt.Errorf("发送指令失败: %w", err)
	}

	// 记录指令发送成功
	logger.LogDeviceCommand(deviceNumber, cmd.Action, cmd.Parameters, map[string]interface{}{
		"result":   "sent",
		"data_len": len(data),
	})

	h.logger.Infof("指令已发送到设备 %s: %s", deviceNumber, cmd.Action)
	return nil
}

// Stop 停止TCP服务器
func (h *TCPHandler) Stop() error {
	h.cancel()
	if h.listener != nil {
		return h.listener.Close()
	}
	return nil
}

// GetConnectedDevices 获取已连接设备列表
func (h *TCPHandler) GetConnectedDevices() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	devices := make([]string, 0, len(h.deviceCache))
	for _, deviceNumber := range h.deviceCache {
		devices = append(devices, deviceNumber)
	}
	return devices
}
