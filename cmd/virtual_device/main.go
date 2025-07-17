package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

// VirtualSensor 虚拟传感器设备
type VirtualSensor struct {
	deviceID     uint32        // 设备ID
	serverAddr   string        // 服务器地址
	conn         net.Conn      // TCP连接
	temperature  float64       // 当前温度
	humidity     float64       // 当前湿度
	voltage      float64       // 当前电压
	battery      int           // 当前电量
	reportPeriod time.Duration // 上报周期
}

// NewVirtualSensor 创建虚拟传感器
func NewVirtualSensor(deviceID uint32, serverAddr string, reportPeriod time.Duration) *VirtualSensor {
	return &VirtualSensor{
		deviceID:     deviceID,
		serverAddr:   serverAddr,
		temperature:  20.0 + rand.Float64()*20.0, // 20-40°C
		humidity:     40.0 + rand.Float64()*40.0, // 40-80%
		voltage:      3.0 + rand.Float64()*1.0,   // 3.0-4.0V
		battery:      80 + rand.Intn(20),         // 80-100%
		reportPeriod: reportPeriod,
	}
}

// Connect 连接到服务器
func (v *VirtualSensor) Connect() error {
	conn, err := net.Dial("tcp", v.serverAddr)
	if err != nil {
		return fmt.Errorf("连接服务器失败: %w", err)
	}
	v.conn = conn
	log.Printf("设备 %d 已连接到服务器 %s", v.deviceID, v.serverAddr)
	return nil
}

// Disconnect 断开连接
func (v *VirtualSensor) Disconnect() {
	if v.conn != nil {
		v.conn.Close()
		v.conn = nil
		log.Printf("设备 %d 已断开连接", v.deviceID)
	}
}

// Start 启动设备，开始发送数据
func (v *VirtualSensor) Start() {
	ticker := time.NewTicker(v.reportPeriod)
	defer ticker.Stop()

	log.Printf("设备 %d 开始运行，上报周期: %v", v.deviceID, v.reportPeriod)

	for {
		select {
		case <-ticker.C:
			if err := v.sendData(); err != nil {
				log.Printf("设备 %d 发送数据失败: %v", v.deviceID, err)
				// 尝试重连
				v.Disconnect()
				if err := v.Connect(); err != nil {
					log.Printf("设备 %d 重连失败: %v", v.deviceID, err)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}
}

// sendData 发送传感器数据
func (v *VirtualSensor) sendData() error {
	if v.conn == nil {
		return fmt.Errorf("连接未建立")
	}

	// 模拟传感器数据变化
	v.updateSensorValues()

	// 构建数据包
	data := v.buildDataPacket()

	// 发送数据
	_, err := v.conn.Write(data)
	if err != nil {
		return fmt.Errorf("发送数据失败: %w", err)
	}

	log.Printf("设备 %d 发送数据: 温度=%.1f°C, 湿度=%.1f%%, 电压=%.2fV, 电量=%d%%",
		v.deviceID, v.temperature, v.humidity, v.voltage, v.battery)

	return nil
}

// updateSensorValues 更新传感器数值（模拟真实传感器的变化）
func (v *VirtualSensor) updateSensorValues() {
	// 温度缓慢变化 ±0.5°C
	v.temperature += (rand.Float64() - 0.5)
	if v.temperature < -40.0 {
		v.temperature = -40.0
	} else if v.temperature > 85.0 {
		v.temperature = 85.0
	}

	// 湿度缓慢变化 ±2%
	v.humidity += (rand.Float64() - 0.5) * 4.0
	if v.humidity < 0.0 {
		v.humidity = 0.0
	} else if v.humidity > 100.0 {
		v.humidity = 100.0
	}

	// 电压缓慢变化 ±0.05V
	v.voltage += (rand.Float64() - 0.5) * 0.1
	if v.voltage < 0.0 {
		v.voltage = 0.0
	} else if v.voltage > 5.0 {
		v.voltage = 5.0
	}

	// 电量缓慢下降
	if rand.Intn(10) == 0 { // 10%概率下降1%
		v.battery--
		if v.battery < 0 {
			v.battery = 0
		}
	}
}

// buildDataPacket 构建数据包
// 格式: [4字节设备ID][2字节温度][2字节湿度][2字节电压][1字节电量][1字节校验]
func (v *VirtualSensor) buildDataPacket() []byte {
	data := make([]byte, 12)

	// 设备ID (4字节，大端序)
	binary.BigEndian.PutUint32(data[0:4], v.deviceID)

	// 温度 (2字节，大端序，单位0.1°C)
	tempValue := int16(v.temperature * 10)
	binary.BigEndian.PutUint16(data[4:6], uint16(tempValue))

	// 湿度 (2字节，大端序，单位0.1%)
	humidityValue := uint16(v.humidity * 10)
	binary.BigEndian.PutUint16(data[6:8], humidityValue)

	// 电压 (2字节，大端序，单位0.01V)
	voltageValue := uint16(v.voltage * 100)
	binary.BigEndian.PutUint16(data[8:10], voltageValue)

	// 电量 (1字节，单位1%)
	data[10] = byte(v.battery)

	// 校验和 (1字节，前11字节异或)
	data[11] = v.calculateChecksum(data[0:11])

	return data
}

// calculateChecksum 计算异或校验和
func (v *VirtualSensor) calculateChecksum(data []byte) byte {
	var checksum byte
	for _, b := range data {
		checksum ^= b
	}
	return checksum
}

func main() {
	var (
		serverAddr  = flag.String("server", "localhost:15001", "服务器地址")
		deviceID    = flag.Uint("device", 1, "设备ID")
		period      = flag.Duration("period", 10*time.Second, "数据上报周期")
		deviceCount = flag.Int("count", 1, "虚拟设备数量")
	)
	flag.Parse()

	log.Printf("启动虚拟传感器设备")
	log.Printf("服务器地址: %s", *serverAddr)
	log.Printf("设备数量: %d", *deviceCount)
	log.Printf("上报周期: %v", *period)

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 创建多个虚拟设备
	for i := 0; i < *deviceCount; i++ {
		deviceNum := uint32(*deviceID) + uint32(i)
		sensor := NewVirtualSensor(deviceNum, *serverAddr, *period)

		// 连接到服务器
		if err := sensor.Connect(); err != nil {
			log.Fatalf("设备 %d 连接失败: %v", deviceNum, err)
		}

		// 启动设备（每个设备在单独的goroutine中运行）
		go sensor.Start()

		// 错开启动时间，避免同时发送
		time.Sleep(time.Duration(i) * time.Second)
	}

	// 保持主程序运行
	select {}
}
