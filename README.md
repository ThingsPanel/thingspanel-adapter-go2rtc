# ThingsPanel go2rtc Adapter

[English](README_EN.md) | [中文](README.md)

本项目是 **ThingsPanel** 物联网平台的官方官方适配器，用于无缝集成 [**go2rtc**](https://github.com/AlexxIT/go2rtc) 这一强大的流媒体服务器。

### 🚀 为什么要结合 Go2RTC？
Go2RTC 是目前最先进的开源流媒体服务器之一，支持 RTSP, RTMP, WebRTC, HTTP-FLV, HLS 等几乎所有主流协议。
通过本适配器，您可以获得：
- 📺 **极低延迟**: 在 ThingsPanel 仪表盘中实现毫秒级 (WebRTC/MSE) 视频监控。
- 🔄 **自动同步**: 自动发现 go2rtc 中的所有视频流，无需在平台手动重复创建设备。
- ⚡ **统一管理**: 在 ThingsPanel 中统一管理视频设备、查看状态、接收告警，同时享受 go2rtc 强大的转码和分发能力。

---

这是一个 ThingsPanel 的**三方接入服务插件**。


## 功能特性

- **自动同步**: 自动从go2rtc获取streams列表，同步到ThingsPanel
- **三方接入**: 使用服务接入模式，无需手动创建设备
- **流媒体集成**: 支持 RTSP, RTMP, WebRTC, HLS 等多种协议
- **设备模拟**: 支持使用 ffmpeg 模拟摄像头流，方便无实物开发测试

---

## 🔧 完整接入/模拟流程

## 🔧 Go2RTC 部署与安装

在运行适配器之前，您需要先安装并启动 `go2rtc` 流媒体服务。

### 1. 下载与安装

请前往 [go2rtc Releases](https://github.com/AlexxIT/go2rtc/releases) 下载适合您架构的二进制文件 (如 Linux amd64)。

```bash
# 下载 (以 v1.9.8 linux_amd64 为例)
wget https://github.com/AlexxIT/go2rtc/releases/download/v1.9.8/go2rtc_linux_amd64 -O go2rtc
chmod +x go2rtc
```

### 2. 配置文件 (go2rtc.yaml)

创建 `/etc/go2rtc/go2rtc.yaml`，填入以下基础配置：

```yaml
api:
  listen: "0.0.0.0:1984" # 开放 API 端口

rtsp:
  listen: ":8554"        # RTSP 端口

streams:
  # 可选：预置一些测试流
  camera_demo: exec:ffmpeg -re -stream_loop -1 -i https://media.w3.org/2010/05/sintel/trailer.mp4 -c copy -rtsp_transport tcp -f rtsp {output}
```

### 3. 启动服务

```bash
# 前台启动测试
./go2rtc -c /etc/go2rtc/go2rtc.yaml

# 或后台运行
nohup ./go2rtc -c /etc/go2rtc/go2rtc.yaml > go2rtc.log 2>&1 &
```

确保访问 `http://<服务器IP>:1984` 能看到 go2rtc 的 Web 界面。

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

## 🧪 自动化测试脚本

项目内置了测试脚本，可以快速添加模拟设备：

```bash
# 在服务器上执行
chmod +x tests/simulate_device.sh
./tests/simulate_device.sh
```

执行后，go2rtc 会新增一个名为 `simulated_cam_v2` 的流。等待 30 秒后，ThingsPanel 应自动发现该设备，并在 **属性** 页签中显示 `stream_url`。

---

## 📹 OBS 推流测试 (直播场景)

除了使用脚本模拟，您也可以使用 **OBS Studio** 进行真实的推流测试。

### 1. 准备工作 (关键)
在开始推流前，您必须先在 `go2rtc.yaml` 中定义这个流名称（留空即可），否则 go2rtc 会拒绝推流。

**修改 `/etc/go2rtc/go2rtc.yaml`**:
```yaml
streams:
  # ... 其他流 ...
  obs_demo:  # 👈 必须添加这一行，表示允许接收名为 obs_demo 的推流
```
修改后记得重启 go2rtc 服务。

### 2. 配置 OBS
1. 打开 OBS -> **设置** -> **推流**。
2. **服务**: 选择 `自定义`。
3. **服务器**: `rtmp://192.168.31.205:1935` (请替换为您的实际服务器IP)。
4. **推流码**: `obs_demo` (必须与配置文件中的名称一致)。

### 3. 开始推流
点击 **"开始直播"**。如果连接成功，OBS 底部状态栏会出现绿色的比特率提示。

### 4. 验证
等待 30 秒左右，ThingsPanel 设备列表中会自动出现一个名为 `obs_demo` 的新设备，且属性中包含推流地址。

---

## 🔧 适配器服务端配置 (重要)

在启动适配器前，需要修改 `configs/config.yaml` 文件。

### 1. 获取 Template Secret (关键)

设备自动注册需要使用 **设备模板密钥**。请按以下步骤获取：

1. 进入 **ThingsPanel 常规管理** -> **设备模板**。
2. 找到（或创建）`go2rtc` 模板，点击**详情**。
3. 进入 **设备设置** -> **自动创建设备**。
4. 找到 **一型一密 (One-Type-One-Secret)** 配置项。
5. 点击复选框 **"允许设备自动创建"**。
6. 复制显示的 **"设备密码"** (即 Template Secret)。

### 2. 修改配置文件

打开 `configs/config.yaml` (或服务器上的 `~/tp-adapter/configs/config.yaml`)：

```yaml
# ...
platform:
  # ...
  # 将刚才复制的密钥填入此处
  template_secret: "ff3267c2-e0f5-6615-ba15-99a50a89600f" 
```

---

## 🔧 完整接入/模拟流程


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
