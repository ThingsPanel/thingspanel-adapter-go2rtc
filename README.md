# ThingsPanel go2rtc Adapter

这是一个 ThingsPanel 的**三方接入服务插件**，用于集成 [go2rtc](https://github.com/AlexxIT/go2rtc) 流媒体服务器。

## 功能特性

- **自动同步**: 自动从go2rtc获取streams列表，同步到ThingsPanel
- **三方接入**: 使用服务接入模式，无需手动创建设备
- **流媒体集成**: 支持 RTSP, RTMP, WebRTC, HLS 等多种协议
- **设备模拟**: 支持使用 ffmpeg 模拟摄像头流，方便无实物开发测试

---

## 🔧 完整接入/模拟流程

### 1. 启动 go2rtc 服务

确保 `go2rtc` 已在宿主机运行，并且 API端口为 `1984`。

### 2. 添加流媒体设备 (模拟或真实)

如果你没有真实摄像头，可以使用模拟流。

#### 方案 A: 添加模拟流 (Virtual Camera)
我们使用 `ffmpeg` 生成一个测试流。你可以通过 go2rtc 的 API 添加：

```bash
# 添加一个名为 virtual_cam 的虚拟流，显示测试图案和时间
curl -X PUT "http://localhost:1984/api/streams?src=exec:ffmpeg+-re+-f+lavfi+-i+testsrc=size=1920x1080:rate=30+-c:v+libx264+-preset+ultrafast+-tune+zerolatency+-f+rtsp+{output}&name=virtual_cam"
```

> **提示**: 这个命令会让 go2rtc 启动 ffmpeg 进程，生成一个实时的 RTSP 流。

#### 方案 B: 添加真实摄像头 (RTSP/ONVIF)
如果有真实摄像头，直接添加其 RTSP 地址：

```bash
curl -X PUT "http://localhost:1984/api/streams?src=rtsp://admin:password@192.168.1.100:554/stream&name=living_room"
```

---

## 一、超级管理员：配置插件

### 1.1 进入插件管理
**路径**: 应用管理 → 插件管理

### 1.2 配置插件 (关键步骤)
找到 `GO2RTC` 插件，点击 **配置**。

⚠️ **注意：HTTP服务地址千万不要加 http:// 前缀！**

| 配置项 | 正确填写示例 | 错误示例 (不要这样填) |
|--------|-------------|----------------------|
| **HTTP服务地址** | `172.17.0.1:12000` | `http://172.17.0.1:12000` |
| **服务订阅主题前缀** | `service/go2rtc` | (留空) |
| **设备类型** | `直连设备` | - |

> **说明**: 平台会自动添加协议头，如果这里填了 `http://`，会导致请求变成 `http://http://...` 从而失败。

---

## 二、租户账户：接入 go2rtc

### 2.1 进入三方接入
**路径**: 设备管理 → **三方接入**

### 2.2 新增接入点
1. 点击 **新增接入**
2. 选择 `GO2RTC` 插件
3. 填写配置：

| 配置项 | 填写内容 | 说明 |
|--------|---------|------|
| 接入点名称 | `本地go2rtc` | 任意名称 |
| go2rtc API地址 | `http://localhost:1984` | 指向 go2rtc 的 API |
| 同步间隔 | `30` | 自动同步周期(秒) |
| 启用自动同步 | `开启` | |

4. 点击 **确认**
   - 如果配置正确，会提示成功。
   - 如果提示"插件请求失败"，请检查超级管理员的HTTP服务地址配置。

---

## 三、同步与查看

### 3.1 同步设备
接入点创建成功后，点击 **设备同步** (或等待自动同步)。
- 平台会从 go2rtc 拉取所有流信息 (`virtual_cam`, `living_room` 等)。

### 3.2 查看设备
进入 **设备列表**，你会看到：
- 设备名称: `virtual_cam`
- 设备状态: 在线 (如果 go2rtc 中流是活跃的)

### 3.3 观看视频
进入 **设备详情** → **实时视频** (需平台支持播放 go2rtc 流)。

---

## 常见问题排查

### Q1: 新增接入点时提示 "插件请求失败"
**原因**: 超级管理员插件配置中的 API 地址填错了。
**解决**: 去掉 `http://` 前缀。只填 `172.17.0.1:12000`。

### Q2: 提示 "404 Not Found"
**原因**: 插件返回了 `null` 而不是空数组 `[]` (已在 v1.0.1 修复)。
**解决**: 确保使用最新的插件版本。

### Q3: 设备列表为空
**原因**: go2rtc 中没有任何流。
**解决**: 参考本文的 [添加模拟流](#方案-a-添加模拟流-virtual-camera) 章节添加一个测试流。

---

## 快速部署命令

```bash
# 1. 编译 (Mac上编译Linux版)
GOOS=linux GOARCH=amd64 go build -o tp-adapter-linux cmd/main.go

# 2. 部署到服务器
./deploy.exp
```
