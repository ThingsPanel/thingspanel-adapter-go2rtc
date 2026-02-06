# ThingsPanel go2rtc Adapter

[English](README_EN.md) | [中文](README.md)

This project is the official adapter for the **ThingsPanel** IoT platform, designed to seamlessly integrate with [**go2rtc**](https://github.com/AlexxIT/go2rtc), a powerful streaming media server.

### 🚀 Why Integrate with Go2RTC?
Go2RTC is one of the most advanced open-source streaming servers available, supporting virtually all mainstream protocols including RTSP, RTMP, WebRTC, HTTP-FLV, and HLS.
By using this adapter, you gain:
- 📺 **Ultra-Low Latency**: Achieve millisecond-level (WebRTC/MSE) video monitoring directly within ThingsPanel dashboards.
- 🔄 **Auto-Synchronization**: Automatically discover all video streams from go2rtc, eliminating the need to manually create devices in the platform.
- ⚡ **Unified Management**: Manage video devices, view status, and receive alerts uniformly within ThingsPanel, while checking the robust transcoding and distribution capabilities of go2rtc.

---

## Features

- **Auto Sync**: Automatically fetch stream lists from go2rtc and sync to ThingsPanel.
- **Third-Party Integration**: Uses the "Service Access" mode, no manual device creation required.
- **Streaming Integration**: Supports RTSP, RTMP, WebRTC, HLS, and more.
- **Device Simulation**: Supports simulating camera streams using ffmpeg for development and testing without physical hardware.

## 🔧 Go2RTC Deployment & Installation

Before running the adapter, you need to install and start the `go2rtc` streaming service.

### 1. Download & Install

Please modify the [go2rtc Releases](https://github.com/AlexxIT/go2rtc/releases) page to download the binary suitable for your architecture (e.g., Linux amd64).

```bash
# Download (example for v1.9.8 linux_amd64)
wget https://github.com/AlexxIT/go2rtc/releases/download/v1.9.8/go2rtc_linux_amd64 -O go2rtc
chmod +x go2rtc
```

### 2. Configuration (go2rtc.yaml)

Create `/etc/go2rtc/go2rtc.yaml` and add the following basic configuration:

```yaml
api:
  listen: "0.0.0.0:1984" # Public API port

rtsp:
  listen: ":8554"        # RTSP port

streams:
  # Optional: Pre-configured test stream
  camera_demo: exec:ffmpeg -re -stream_loop -1 -i https://media.w3.org/2010/05/sintel/trailer.mp4 -c copy -rtsp_transport tcp -f rtsp {output}
```

### 3. Start Service

```bash
# Foreground test
./go2rtc -c /etc/go2rtc/go2rtc.yaml

# Or background run
nohup ./go2rtc -c /etc/go2rtc/go2rtc.yaml > go2rtc.log 2>&1 &
```

Ensure that you can access the go2rtc Web UI at `http://<Server_IP>:1984`.

---

## 🔧 Quick Start / Simulation Flow

### 1. Start go2rtc Service

Ensure `go2rtc` is running on the host and the API port is `1984`.

### 2. Add Media Stream (Simulated or Real)

#### Option A: Add Simulated Stream (Virtual Camera)
Use `ffmpeg` to generate a test stream. Add it via the go2rtc API:

```bash
curl -X PUT "http://localhost:1984/api/streams?src=exec:ffmpeg+-re+-f+lavfi+-i+testsrc=size=1920x1080:rate=30+-c:v+libx264+-preset+ultrafast+-tune+zerolatency+-f+rtsp+{output}&name=virtual_cam"
```

#### Option B: Add Real Camera (RTSP/ONVIF)

```bash
curl -X PUT "http://localhost:1984/api/streams?src=rtsp://admin:password@192.168.1.100:554/stream&name=living_room"
```

### 3. Verification

The adapter will automatically discover these streams and report them to ThingsPanel. Check the device list; you should see the new devices within 30 seconds.

## 🔧 Adapter Server Configuration (Important)

Before starting the adapter, you must modify the `configs/config.yaml` file.

### 1. Get Template Secret (Critical)

Device auto-registration requires a **Device Template Secret**. 
1. Go to **ThingsPanel Management** -> **Device Templates**.
2. Find (or create) the `go2rtc` template -> **Details**.
3. Go to **Device Settings** -> **Auto Create Device**.
4. Enable **"Allow Device Auto Creation"** under **One-Type-One-Secret**.
5. Copy the displayed **"Device Password"** (Template Secret).

### 2. Modify Config File

Open `configs/config.yaml`:

```yaml
platform:
  # Paste the secret here
  template_secret: "YOUR_TEMPLATE_SECRET_HERE" 
```

## 📹 OBS Streaming Test (Live Scenario)

Apart from using scripts, you can use **OBS Studio** for real streaming tests.

### 1. Preparation (Critical)
Before streaming, you MUST define the stream name in `go2rtc.yaml` (can be empty), otherwise go2rtc will reject the stream.

**Edit `/etc/go2rtc/go2rtc.yaml`**:
```yaml
streams:
  # ... other streams ...
  obs_demo:  # 👈 Must add this line to allow "obs_demo" stream
```
Remember to restart the go2rtc service after editing.

### 2. Configure OBS
1. Open OBS -> **Settings** -> **Stream**.
2. **Service**: Select `Custom`.
3. **Server**: `rtmp://192.168.31.205:1935` (Replace with your actual server IP).
4. **Stream Key**: `obs_demo` (Must match the name in config).

### 3. Start Streaming
Click **"Start Streaming"**. A green bitrate indicator should appear at the bottom if successful.

### 4. Verify
Wait for about 30 seconds. A new device named `obs_demo` will automatically appear in ThingsPanel, with the stream URL in its attributes.

## 🧪 Automated Test Script

A test script is included to quickly add simulated devices:

```bash
chmod +x tests/simulate_device.sh
./tests/simulate_device.sh
```

After execution, a new device `simulated_cam_v2` will be added to go2rtc. ThingsPanel should discover it and display the `stream_url` in the **Attributes** tab after 30 seconds.
