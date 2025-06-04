package protocol

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// ProtocolManager 协议管理器
type ProtocolManager struct {
	protocols   map[string]ProtocolHandler
	tcpHandlers map[string]*TCPHandler
	platform    PlatformInterface
	logger      *logrus.Logger
	mu          sync.RWMutex
}

// NewManager 创建协议管理器
func NewManager(platform PlatformInterface, logger *logrus.Logger) *ProtocolManager {
	return &ProtocolManager{
		protocols:   make(map[string]ProtocolHandler),
		tcpHandlers: make(map[string]*TCPHandler),
		platform:    platform,
		logger:      logger,
	}
}

// RegisterProtocol 注册协议 - 框架自动处理TCP服务器创建和生命周期管理
func (m *ProtocolManager) RegisterProtocol(handler ProtocolHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := handler.Name()
	port := handler.Port()

	// 检查协议名称冲突
	if _, exists := m.protocols[name]; exists {
		return fmt.Errorf("协议 %s 已注册", name)
	}

	// 检查端口冲突
	for _, existingHandler := range m.protocols {
		if existingHandler.Port() == port {
			return fmt.Errorf("端口 %d 已被协议 %s 使用", port, existingHandler.Name())
		}
	}

	// 启动协议
	if err := handler.Start(); err != nil {
		return fmt.Errorf("启动协议 %s 失败: %w", name, err)
	}

	// 创建TCP处理器
	tcpHandler := NewTCPHandler(port, handler, m.platform, m.logger)

	// 启动TCP服务器
	if err := tcpHandler.Start(); err != nil {
		handler.Stop() // 清理协议资源
		return fmt.Errorf("启动协议 %s 的TCP服务器失败: %w", name, err)
	}

	// 注册到管理器
	m.protocols[name] = handler
	m.tcpHandlers[name] = tcpHandler

	m.logger.Infof("协议 %s (v%s) 已注册并启动，端口: %d", name, handler.Version(), port)
	return nil
}

// UnregisterProtocol 注销协议
func (m *ProtocolManager) UnregisterProtocol(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	handler, exists := m.protocols[name]
	if !exists {
		return fmt.Errorf("协议 %s 未注册", name)
	}

	tcpHandler := m.tcpHandlers[name]

	// 停止TCP服务器
	if err := tcpHandler.Stop(); err != nil {
		m.logger.WithError(err).Errorf("停止协议 %s 的TCP服务器失败", name)
	}

	// 停止协议
	if err := handler.Stop(); err != nil {
		m.logger.WithError(err).Errorf("停止协议 %s 失败", name)
	}

	// 从管理器中移除
	delete(m.protocols, name)
	delete(m.tcpHandlers, name)

	m.logger.Infof("协议 %s 已注销", name)
	return nil
}

// GetProtocolInfo 获取所有已注册协议信息
func (m *ProtocolManager) GetProtocolInfo() []ProtocolInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]ProtocolInfo, 0, len(m.protocols))
	for name, handler := range m.protocols {
		status := "running"
		// 这里可以添加更复杂的状态检查逻辑

		infos = append(infos, ProtocolInfo{
			Name:    name,
			Version: handler.Version(),
			Port:    handler.Port(),
			Status:  status,
		})
	}

	return infos
}

// GetProtocol 获取指定协议处理器
func (m *ProtocolManager) GetProtocol(name string) (ProtocolHandler, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	handler, exists := m.protocols[name]
	return handler, exists
}

// SendCommand 向指定设备发送指令
func (m *ProtocolManager) SendCommand(protocolName, deviceID string, cmd *Command) error {
	m.mu.RLock()
	tcpHandler, exists := m.tcpHandlers[protocolName]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("协议 %s 未注册", protocolName)
	}

	return tcpHandler.SendCommand(deviceID, cmd)
}

// GetConnectedDevices 获取指定协议的已连接设备
func (m *ProtocolManager) GetConnectedDevices(protocolName string) ([]string, error) {
	m.mu.RLock()
	tcpHandler, exists := m.tcpHandlers[protocolName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("协议 %s 未注册", protocolName)
	}

	return tcpHandler.GetConnectedDevices(), nil
}

// GetAllConnectedDevices 获取所有协议的已连接设备
func (m *ProtocolManager) GetAllConnectedDevices() map[string][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]string)
	for name, tcpHandler := range m.tcpHandlers {
		result[name] = tcpHandler.GetConnectedDevices()
	}

	return result
}

// StopAll 停止所有协议
func (m *ProtocolManager) StopAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastError error

	// 停止所有TCP处理器
	for name, tcpHandler := range m.tcpHandlers {
		if err := tcpHandler.Stop(); err != nil {
			m.logger.WithError(err).Errorf("停止协议 %s 的TCP服务器失败", name)
			lastError = err
		}
	}

	// 停止所有协议
	for name, handler := range m.protocols {
		if err := handler.Stop(); err != nil {
			m.logger.WithError(err).Errorf("停止协议 %s 失败", name)
			lastError = err
		}
	}

	// 清空管理器
	m.protocols = make(map[string]ProtocolHandler)
	m.tcpHandlers = make(map[string]*TCPHandler)

	m.logger.Info("所有协议已停止")
	return lastError
}

// IsRunning 检查管理器是否有运行中的协议
func (m *ProtocolManager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.protocols) > 0
}
