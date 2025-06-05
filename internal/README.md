# internal包结构说明

本文档描述了协议插件模板的internal包结构，基于**单协议专用**设计理念。

## 📁 目录结构

```
internal/
├── bootstrap/          # 应用引导和初始化
│   ├── app.go         # 应用上下文和生命周期管理
│   ├── config.go      # 配置加载
│   ├── http.go        # HTTP服务启动
│   ├── logger.go      # 日志系统初始化
│   └── platform.go    # 平台客户端初始化
├── config/            # 配置结构定义
│   └── config.go      # 配置结构体定义
├── protocol/          # 协议处理核心模块 ⭐
│   ├── interface.go   # 协议接口定义
│   ├── handler.go     # 单协议处理器
│   ├── tcp_handler.go # TCP连接处理器
│   └── plugins/       # 协议实现插件
├── platform/          # 平台客户端
│   └── platform.go   # ThingsPanel平台通信
├── handler/           # HTTP处理器
│   └── handler.go     # HTTP请求处理
├── form_json/         # 表单配置
│   ├── types.go       # 表单类型定义
│   └── *.json         # 表单配置文件
└── pkg/               # 通用包
    └── logger/        # 日志系统
```

## 🎯 设计原则

### 1. 单一职责
每个包都有明确的职责：
- `bootstrap` - 负责应用启动流程
- `config` - 配置管理
- `protocol` - 协议处理核心逻辑
- `platform` - 平台通信
- `handler` - HTTP接口处理
- `pkg` - 通用工具

### 2. 依赖关系清晰
```
bootstrap → config + platform + protocol + handler
protocol → platform (接口依赖)
handler → platform
pkg → 独立模块，被其他模块使用
```

### 3. 单协议专用
- 整个架构专为单协议设计
- 避免多协议管理的复杂性
- 配置简单，一个端口，一个协议

## 📝 模块详解

### bootstrap - 应用引导
应用的启动中心，负责：
- 配置加载和验证
- 各个组件的初始化顺序
- 应用生命周期管理
- 优雅关闭处理

**关键文件：**
- `app.go` - 应用上下文，需要在`initializeProtocol`函数中实现你的协议

### protocol - 协议处理核心 ⭐
框架的核心模块，提供：
- 极简的协议开发接口
- 自动TCP连接管理
- 设备上下线检测
- 数据自动转发

**开发者只需要关注：**
- 实现`ProtocolHandler`接口的3个核心方法
- 设备编号提取逻辑
- 数据解析逻辑
- 指令编码逻辑（可选）

### platform - 平台客户端
与ThingsPanel平台的通信桥梁：
- 设备管理和缓存
- 遥测数据发送
- 设备状态管理
- 心跳上报

### config - 配置管理
简洁的配置结构：
- 服务器配置（端口、超时等）
- 平台配置（MQTT、API等）
- 日志配置（文件、级别等）

### handler - HTTP处理
HTTP管理接口：
- **表单配置获取** - 支持平台表单渲染器，提供动态配置界面
- 设备管理接口
- 通知处理
- 健康检查

**表单系统说明：**
- 平台的表单渲染器会读取协议中间件提供的JSON配置
- 用户可以通过动态生成的表单填写协议特定参数（如Modbus寄存器地址）
- 中间件通过平台接口查询设备配置，获取用户填写的参数
- 这个机制对于需要用户配置的协议（如Modbus、OPC-UA等）非常重要

### form_json - 表单配置 ⭐
协议与平台表单渲染器的接口：
- **JSON配置文件** - 定义表单结构，供平台渲染器使用
- **类型定义** - 用于解析用户填写的表单数据  
- **动态配置支持** - 支持不同协议的个性化配置需求

**重要性：**
- 这是实现协议灵活配置的核心机制
- 不同协议可以定义不同的配置表单
- 用户可以通过Web界面配置协议参数，无需修改代码

### pkg - 通用包
可复用的工具模块：
- 高级日志系统
- 设备独立日志
- 工具函数

## 🚀 开发流程

### 1. 实现协议处理器
在 `internal/protocol/plugins/your_protocol/` 下创建你的协议实现。

### 2. 修改应用初始化
在 `bootstrap/app.go` 的 `initializeProtocol` 函数中：
```go
// 替换为你的协议实现
protocolHandler := your_protocol.NewHandler(cfg.Server.Port)
```

### 3. 配置表单（重要！）
在 `form_json/` 目录下配置你的设备和服务表单：
- **设备配置表单** (`form_config.json`) - 定义设备参数配置界面
- **设备凭证表单** (`form_voucher.json`) - 定义设备认证信息界面  
- **服务凭证表单** (`form_service_voucher.json`) - 定义服务接入点配置界面

**表单配置示例：**
```json
[
  {
    "dataKey": "register_address",
    "label": "寄存器地址", 
    "placeholder": "请输入Modbus寄存器地址",
    "type": "input",
    "validate": {
      "message": "寄存器地址不能为空",
      "required": true,
      "rules": "/^\\d+$/",
      "type": "number"
    }
  }
]
```

### 4. 启动应用
```bash
go run cmd/main.go
```

## ✅ 优化亮点

### 1. 极简化
- 移除了多协议的复杂性
- 配置文件简洁明了
- 开发流程标准化

### 2. 模板友好
- 清晰的TODO注释指导
- 标准化的目录结构
- 完整的示例代码

### 3. 维护性好
- 单一职责原则
- 清晰的依赖关系
- 完善的错误处理

### 4. 扩展性强
- 接口设计灵活
- 组件可独立替换
- 支持复杂协议需求

## 📚 下一步

1. **实现你的协议** - 在`protocol/plugins/`下创建协议实现
2. **配置表单** - 根据协议需求配置表单文件
3. **测试验证** - 使用TCP工具测试协议功能
4. **部署运行** - 容器化部署到生产环境

更多详细开发指南请参考：`internal/protocol/README.md` 