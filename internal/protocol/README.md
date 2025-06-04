# 协议插件开发指南

## 概述

本框架基于**端口隔离**的设计思想，为IoT协议开发提供极简的开发体验。90%的协议只需要实现2个核心方法即可完成开发。

## 重要概念区分

**设备编号 vs 设备ID**
- `device_number` - 设备本身的唯一标识符，从设备数据中提取
- `device_id` - 平台分配给设备的内部ID，用于平台内部标识

开发协议时，你需要实现的是从设备数据中提取 `device_number`，框架会自动通过平台接口获取对应的 `device_id`。

## 核心设计原则

- **极简优先** - 90%的协议只需要2个核心方法
- **端口隔离** - 每个协议使用独立端口，避免协议识别问题  
- **渐进增强** - 简单协议保持简单，复杂需求可选扩展
- **完全通用** - 架构不绑定任何特定协议实现

## 快速开始

### 1. 选择协议类型

#### 简单协议（推荐，90%的情况）

适用于：传感器数据采集、简单设备控制等无状态协议

**特点：**
- 无需维护连接状态
- 数据包相对简单
- 设备ID容易提取
- 消息类型单一

#### 复杂协议（5%的情况）

适用于：网关协议、多消息类型、有状态协议

**特点：**
- 需要连接认证/握手
- 多种消息类型（心跳、数据、状态等）
- 维护连接状态和会话信息
- 复杂的交互逻辑

### 2. 创建协议实现

#### 简单协议开发

```bash
# 1. 复制模板
cp -r internal/protocol/plugins/template/simple internal/protocol/plugins/your_protocol

# 2. 修改文件名
mv internal/protocol/plugins/your_protocol/handler.go internal/protocol/plugins/your_protocol/your_protocol.go
```

**核心步骤：**

1. **修改基本信息**
```go
func (h *YourProtocolHandler) Name() string {
    return "YourProtocol" // 协议名称
}

func (h *YourProtocolHandler) Version() string {
    return "1.0.0" // 协议版本
}
```

2. **实现设备编号提取**
```go
func (h *YourProtocolHandler) ExtractDeviceNumber(data []byte) (string, error) {
    // 根据你的协议格式提取设备编号
    // 这个方法必须能正确提取设备编号，否则设备无法上线
    
    if len(data) < 4 {
        return "", errors.New("数据包太短")
    }
    
    // 示例：从前4字节提取设备编号
    deviceNumber := binary.BigEndian.Uint32(data[0:4])
    return fmt.Sprintf("%d", deviceNumber), nil
}
```

3. **实现数据解析**
```go
func (h *YourProtocolHandler) ParseData(data []byte) (*protocol.Message, error) {
    // 1. 验证数据包格式
    if len(data) < expectedLength {
        return nil, errors.New("数据包长度不正确")
    }
    
    // 2. 提取设备编号
    deviceNumber, err := h.ExtractDeviceNumber(data)
    if err != nil {
        return nil, err
    }
    
    // 3. 解析传感器数据
    sensorData := map[string]interface{}{
        "temperature": extractTemperature(data),
        "humidity":    extractHumidity(data),
        // ... 其他数据字段
    }
    
    return &protocol.Message{
        DeviceNumber: deviceNumber, // 设备编号（从设备提取）
        DeviceID:     "",          // 设备ID（框架自动填充）
        MessageType:  "data",
        Timestamp:    time.Now(),
        Data:         sensorData,
        Quality:      1,
    }, nil
}
```

4. **实现指令编码**（可选）
```go
func (h *YourProtocolHandler) EncodeCommand(cmd *protocol.Command) ([]byte, error) {
    // 如果设备不支持控制，返回错误即可
    // return nil, errors.New("设备不支持控制指令")
    
    switch cmd.Action {
    case "sleep":
        return h.buildSleepCommand(cmd.DeviceNumber, cmd.Parameters)
    case "config":
        return h.buildConfigCommand(cmd.DeviceNumber, cmd.Parameters)
    default:
        return nil, fmt.Errorf("不支持的指令: %s", cmd.Action)
    }
}
```

#### 复杂协议开发

```bash
# 1. 复制模板
cp -r internal/protocol/plugins/template/complex internal/protocol/plugins/your_protocol
```

**额外需要实现：**

1. **连接处理逻辑**
```go
func (h *YourProtocolHandler) HandleConnection(conn net.Conn) error {
    // 1. 连接认证/握手
    deviceID, err := h.authenticateConnection(conn)
    if err != nil {
        return err
    }
    
    // 2. 创建会话
    session := h.createSession(deviceID, conn)
    defer h.cleanupSession(deviceID)
    
    // 3. 消息处理循环
    for {
        // 读取数据
        // 处理不同类型的消息
        // 发送响应
    }
}
```

2. **会话管理**
```go
type Session struct {
    DeviceID      string
    Conn          net.Conn
    LastHeartbeat time.Time
    Authenticated bool
    // ... 其他状态信息
}
```

### 3. 注册协议

在 `internal/bootstrap/app.go` 中注册协议：

```go
func initializeProtocols(app *AppContext, cfg *config.Config) error {
    manager := protocol.NewManager(app.PlatformClient, logrus.StandardLogger())
    
    // 注册你的协议
    if cfg.Protocols.YourProtocol.Enabled {
        handler := your_protocol.NewHandler(cfg.Protocols.YourProtocol.Port)
        if err := manager.RegisterProtocol(handler); err != nil {
            return fmt.Errorf("注册协议失败: %w", err)
        }
    }
    
    app.ProtocolManager = manager
    return nil
}
```

### 4. 添加配置

在配置文件中添加协议配置：

```yaml
protocols:
  your_protocol:
    enabled: true
    port: 15001  # 协议专用端口
```

## 框架功能

### 自动处理的功能

框架会自动处理以下功能，协议开发者无需关心：

1. **TCP服务器管理**
   - 自动创建TCP监听服务器
   - 处理连接接入和断开
   - 并发连接处理

2. **设备上下线通知**
   - 首次收到数据包并成功提取设备ID后发送上线事件
   - 连接断开时自动发送下线事件
   - 事件自动发送到ThingsPanel平台

3. **数据转发**
   - 解析后的数据自动发送到ThingsPanel平台
   - 支持遥测数据、状态数据等多种类型

4. **错误处理**
   - 连接超时处理
   - 数据解析错误恢复
   - 资源清理

5. **并发安全**
   - 线程安全的设备连接管理
   - 并发数据处理

### 开发者需要实现的功能

1. **协议解析逻辑**
   - 数据包格式解析
   - 设备编号提取
   - 传感器数据提取

2. **指令编码逻辑**（可选）
   - 控制指令格式构建
   - 参数验证和处理

3. **协议特定逻辑**（复杂协议）
   - 连接认证和握手
   - 会话状态管理
   - 多消息类型处理

## 示例参考

### 完整示例

查看 `internal/protocol/plugins/examples/sensor_protocol.go` 获取完整的协议实现示例。

该示例实现了一个传感器协议，包含：
- 温度、湿度、电压、电量数据采集
- 休眠、配置、查询指令支持
- 完整的数据验证和错误处理

### 测试数据

传感器协议测试数据包：
```
设备ID=1, 温度=25.6°C, 湿度=60.5%, 电压=3.30V, 电量=85%
原始数据: 00000001 0100 025D 014A 0055 A8

解析结果：
- 设备ID: 1
- 温度: 25.6°C
- 湿度: 60.5%
- 电压: 3.30V
- 电量: 85%
```

## 调试和测试

### 1. 启动服务

```bash
go run cmd/main.go
```

### 2. 查看日志

协议启动后会看到类似日志：
```
INFO[0001] 协议 SensorProtocol (v1.0.0) 已注册并启动，端口: 15001
INFO[0001] 协议 SensorProtocol 在端口 15001 启动成功
```

### 3. 连接测试

使用TCP客户端工具连接到协议端口，发送测试数据包：

```bash
# 使用nc命令测试
echo -ne '\x00\x00\x00\x01\x01\x00\x02\x5D\x01\x4A\x55\xA8' | nc localhost 15001
```

### 4. 查看设备事件

设备连接后会看到：
```
INFO[0010] 设备上线: 00000001 (127.0.0.1:54321) - 协议: SensorProtocol
INFO[0010] 设备数据: 00000001 - map[battery:85 humidity:60.5 temperature:25.6 voltage:3.3]
```

## 常见问题

### Q: 设备无法上线？
**A:** 检查 `ExtractDeviceNumber` 方法是否能正确提取设备编号。这是最常见的问题。

### Q: 数据解析失败？
**A:** 检查数据包格式是否与 `ParseData` 方法中的解析逻辑匹配。

### Q: 端口冲突？
**A:** 确保每个协议使用不同的端口，检查配置文件中的端口设置。

### Q: 设备频繁上下线？
**A:** 检查网络连接稳定性，或者调整连接超时时间。

### Q: 何时使用复杂协议？
**A:** 只有当协议需要连接认证、多消息类型处理、会话状态管理时才使用。90%的情况下简单协议就足够了。

## 最佳实践

1. **数据验证**
   - 验证数据包长度
   - 验证数据范围合理性
   - 添加校验和验证

2. **错误处理**
   - 详细的错误信息
   - 优雅的错误恢复
   - 避免因单个错误导致连接断开

3. **日志记录**
   - 记录关键操作
   - 使用合适的日志级别
   - 包含足够的调试信息

4. **性能考虑**
   - 避免内存泄漏
   - 高效的数据解析
   - 合理的缓冲区大小

5. **安全考虑**
   - 输入数据验证
   - 防止缓冲区溢出
   - 连接数限制（如需要）

## 架构优势

1. **极其简单** - 大多数协议只需要实现2个方法
2. **完全通用** - 不绑定任何特定协议
3. **端口隔离** - 避免所有协议识别复杂性
4. **自动监控** - 设备上下线自动通知平台
5. **渐进增强** - 简单协议保持简单，复杂需求可扩展
6. **故障隔离** - 协议独立运行，互不影响

## 总结

这个架构的核心思想是：**让简单的事情保持简单，让复杂的事情成为可能**

对于大多数IoT协议开发，你只需要关注数据解析逻辑，其他一切都由框架自动处理。

## 核心接口

每个协议必须实现 `ProtocolHandler` 接口：

```go
type ProtocolHandler interface {
    Name() string                                      // 协议名称
    Version() string                                   // 协议版本
    Port() int                                         // 监听端口
    
    // 核心方法
    ParseData(data []byte) (*Message, error)           // 解析设备数据
    EncodeCommand(cmd *Command) ([]byte, error)        // 编码控制指令
    ExtractDeviceNumber(data []byte) (string, error)   // 提取设备编号
    
    // 生命周期
    Start() error                                       // 启动协议
    Stop() error                                        // 停止协议
}
```

## 数据结构

### Message 消息结构
```go
type Message struct {
    DeviceNumber string                 // 设备编号(从设备中提取)
    DeviceID     string                 // 设备ID(平台分配的ID)
    MessageType  string                 // data/heartbeat/status
    Timestamp    time.Time              // 时间戳
    Data         map[string]interface{} // 设备数据
    Quality      int                    // 数据质量，1=正常
}
```

### Command 指令结构
```go
type Command struct {
    DeviceNumber string        // 设备编号(从设备中提取)
    DeviceID     string        // 设备ID(平台分配的ID)
    CommandID    string        // 指令ID
    Action       string        // 指令动作
    Parameters   interface{}   // 指令参数
    Timeout      time.Duration // 超时时间
} 