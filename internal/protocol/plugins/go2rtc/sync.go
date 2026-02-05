// internal/protocol/plugins/go2rtc/sync.go
package go2rtc

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"tp-plugin/internal/platform"

	"github.com/sirupsen/logrus"
)

// DeviceSyncService 设备同步服务
// 从go2rtc定期获取streams列表并同步到ThingsPanel
type DeviceSyncService struct {
	handler        *Go2RTCProtocolHandler
	platformClient *platform.PlatformClient
	logger         *logrus.Logger

	syncInterval  time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
	syncedDevices map[string]bool // 已同步设备列表 (stream name -> synced)
	syncedMutex   sync.RWMutex
}

// NewDeviceSyncService 创建设备同步服务
func NewDeviceSyncService(
	handler *Go2RTCProtocolHandler,
	platformClient *platform.PlatformClient,
	logger *logrus.Logger,
	syncIntervalSec int,
) *DeviceSyncService {
	if syncIntervalSec <= 0 {
		syncIntervalSec = 30 // 默认30秒
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &DeviceSyncService{
		handler:        handler,
		platformClient: platformClient,
		logger:         logger,
		syncInterval:   time.Duration(syncIntervalSec) * time.Second,
		ctx:            ctx,
		cancel:         cancel,
		syncedDevices:  make(map[string]bool),
	}
}

// Start 启动同步服务
func (s *DeviceSyncService) Start() {
	s.logger.Infof("设备同步服务启动，间隔: %v", s.syncInterval)

	// 立即执行一次同步
	s.syncDevices()

	// 定时同步
	go func() {
		ticker := time.NewTicker(s.syncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.syncDevices()
			case <-s.ctx.Done():
				s.logger.Info("设备同步服务已停止")
				return
			}
		}
	}()
}

// Stop 停止同步服务
func (s *DeviceSyncService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

// syncDevices 执行设备同步
func (s *DeviceSyncService) syncDevices() {
	// 从go2rtc获取streams列表
	streams, err := s.handler.ListStreams()
	if err != nil {
		s.logger.WithError(err).Error("获取go2rtc streams失败")
		return
	}

	s.logger.Debugf("从go2rtc获取到 %d 个streams", len(streams))

	// 记录当前存在的streams
	currentStreams := make(map[string]bool)

	// 同步新设备
	for _, stream := range streams {
		currentStreams[stream.Name] = true

		// 检查是否已同步
		s.syncedMutex.RLock()
		synced := s.syncedDevices[stream.Name]
		s.syncedMutex.RUnlock()

		if !synced {
			// 注册新设备
			if err := s.registerDevice(stream); err != nil {
				s.logger.WithError(err).Errorf("注册设备失败: %s", stream.Name)
			} else {
				s.syncedMutex.Lock()
				s.syncedDevices[stream.Name] = true
				s.syncedMutex.Unlock()
				s.logger.Infof("设备已同步: %s", stream.Name)
			}
		}
	}

	// 检测已删除的streams (发送离线状态)
	s.syncedMutex.Lock()
	for deviceName := range s.syncedDevices {
		if !currentStreams[deviceName] {
			// 设备已从go2rtc移除，发送离线状态
			s.sendDeviceOffline(deviceName)
			delete(s.syncedDevices, deviceName)
			s.logger.Infof("设备已移除: %s", deviceName)
		}
	}
	s.syncedMutex.Unlock()
}

// registerDevice 注册设备到ThingsPanel
func (s *DeviceSyncService) registerDevice(stream StreamInfo) error {
	var deviceID string

	// 使用动态注册API
	result, err := s.platformClient.DynamicRegister(stream.Name)
	if err != nil {
		// 如果是设备已存在错误，则获取设备信息继续往下走
		// 注意：这里的错误匹配需要根据SDK返回的实际错误信息来定
		// SDK error: "直连设备动态注册失败: 设备已存在"
		if strings.Contains(err.Error(), "已存在") || strings.Contains(err.Error(), "exists") {
			s.logger.Debugf("设备 %s 已存在，尝试获取ID并更新属性", stream.Name)
			device, errGet := s.platformClient.GetDevice(stream.Name)
			if errGet != nil {
				return fmt.Errorf("设备已存在但获取信息失败: %v", errGet)
			}
			deviceID = device.ID
		} else {
			return err
		}
	} else {
		deviceID = result.DeviceID
		s.logger.WithFields(logrus.Fields{
			"device_id":     deviceID,
			"device_number": stream.Name,
			"stream_url":    stream.URL,
		}).Info("设备动态注册成功")
	}

	// 发送设备在线状态
	if err := s.platformClient.SendDeviceStatus(deviceID, platform.DeviceStatusOnline); err != nil {
		s.logger.WithError(err).Warn("发送设备在线状态失败")
	}

	// 上报流地址属性
	if stream.URL != "" {
		attrs := map[string]interface{}{
			"stream_url": stream.URL,
		}
		if err := s.platformClient.SendAttributes(deviceID, attrs); err != nil {
			s.logger.WithError(err).Warn("发送流地址属性失败")
		} else {
			s.logger.Infof("上报属性成功: stream_url=%s", stream.URL)
		}
	}

	return nil
}

// sendDeviceOffline 发送设备离线状态
func (s *DeviceSyncService) sendDeviceOffline(streamName string) {
	// 获取设备信息
	device, err := s.platformClient.GetDevice(streamName)
	if err != nil {
		s.logger.WithError(err).Warnf("获取设备信息失败: %s", streamName)
		return
	}

	// 发送离线状态
	if err := s.platformClient.SendDeviceStatus(device.ID, platform.DeviceStatusOffline); err != nil {
		s.logger.WithError(err).Warnf("发送设备离线状态失败: %s", streamName)
	}
}

// GetSyncedDevices 获取已同步设备列表
func (s *DeviceSyncService) GetSyncedDevices() []string {
	s.syncedMutex.RLock()
	defer s.syncedMutex.RUnlock()

	devices := make([]string, 0, len(s.syncedDevices))
	for name := range s.syncedDevices {
		devices = append(devices, name)
	}
	return devices
}
