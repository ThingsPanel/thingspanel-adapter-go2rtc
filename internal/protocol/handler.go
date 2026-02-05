package protocol

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// SingleProtocolHandler 单协议处理器
type SingleProtocolHandler struct {
	handler    ProtocolHandler
	tcpHandler *TCPHandler
	platform   PlatformInterface
	logger     *logrus.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewSingleProtocolHandler 创建单协议处理器
func NewSingleProtocolHandler(handler ProtocolHandler, platform PlatformInterface, logger *logrus.Logger) *SingleProtocolHandler {
	ctx, cancel := context.WithCancel(context.Background())

	return &SingleProtocolHandler{
		handler:  handler,
		platform: platform,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start 启动协议处理器
func (s *SingleProtocolHandler) Start() error {
	// 启动协议
	if err := s.handler.Start(); err != nil {
		return fmt.Errorf("启动协议 %s 失败: %w", s.handler.Name(), err)
	}

	// 创建并启动TCP处理器
	s.tcpHandler = NewTCPHandler(s.handler.Port(), s.handler, s.platform, s.logger)
	if err := s.tcpHandler.Start(); err != nil {
		s.handler.Stop()
		return fmt.Errorf("启动TCP服务器失败: %w", err)
	}

	s.logger.Infof("协议 %s (v%s) 已启动，端口: %d", s.handler.Name(), s.handler.Version(), s.handler.Port())
	return nil
}

// Stop 停止协议处理器
func (s *SingleProtocolHandler) Stop() error {
	var lastError error

	// 取消上下文
	s.cancel()

	// 停止TCP处理器
	if s.tcpHandler != nil {
		if err := s.tcpHandler.Stop(); err != nil {
			s.logger.WithError(err).Error("停止TCP服务器失败")
			lastError = err
		}
	}

	// 停止协议
	if err := s.handler.Stop(); err != nil {
		s.logger.WithError(err).Error("停止协议失败")
		lastError = err
	}

	s.logger.Infof("协议 %s 已停止", s.handler.Name())
	return lastError
}

// GetInfo 获取协议信息
func (s *SingleProtocolHandler) GetInfo() ProtocolInfo {
	status := "running"
	if s.tcpHandler == nil {
		status = "stopped"
	}

	return ProtocolInfo{
		Name:    s.handler.Name(),
		Version: s.handler.Version(),
		Port:    s.handler.Port(),
		Status:  status,
	}
}

// SendCommand 向指定设备发送指令
func (s *SingleProtocolHandler) SendCommand(deviceNumber string, cmd *Command) error {
	if s.tcpHandler == nil {
		return fmt.Errorf("TCP处理器未启动")
	}

	return s.tcpHandler.SendCommand(deviceNumber, cmd)
}

// GetConnectedDevices 获取已连接设备列表
func (s *SingleProtocolHandler) GetConnectedDevices() []string {
	if s.tcpHandler == nil {
		return []string{}
	}

	return s.tcpHandler.GetConnectedDevices()
}

// IsRunning 检查协议是否正在运行
func (s *SingleProtocolHandler) IsRunning() bool {
	return s.tcpHandler != nil
}

// GetHandler 获取底层协议处理器
func (s *SingleProtocolHandler) GetHandler() ProtocolHandler {
	return s.handler
}

// Name 协议名称
func (s *SingleProtocolHandler) Name() string {
	return s.handler.Name()
}

// Version 协议版本
func (s *SingleProtocolHandler) Version() string {
	return s.handler.Version()
}

// Port 协议端口
func (s *SingleProtocolHandler) Port() int {
	return s.handler.Port()
}

// ExtractDeviceNumber 提取设备编号
func (s *SingleProtocolHandler) ExtractDeviceNumber(data []byte) (string, error) {
	return s.handler.ExtractDeviceNumber(data)
}

// ParseData 解析数据
func (s *SingleProtocolHandler) ParseData(data []byte) (*Message, error) {
	return s.handler.ParseData(data)
}

// EncodeCommand 编码指令
func (s *SingleProtocolHandler) EncodeCommand(cmd *Command) ([]byte, error) {
	return s.handler.EncodeCommand(cmd)
}
