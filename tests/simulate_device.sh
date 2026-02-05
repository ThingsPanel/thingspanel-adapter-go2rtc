#!/bin/bash

# simulate_device.sh
# 模拟一个摄像头流并添加到 go2rtc，用于测试设备自动发现和属性上报

GO2RTC_API="http://localhost:1984"
STREAM_NAME="simulated_cam_v2"
# 使用 ffmpeg 生成测试流 (testsrc)
# 注意: URL encode spaces with +
# STREAM_SRC="exec:ffmpeg+-re+-f+lavfi+-i+testsrc=size=1920x1080:rate=30+-c:v+libx264+-preset+ultrafast+-tune+zerolatency+-f+rtsp+{output}"

# 也可以使用简单的 RTSP 地址 (如果不需真实画面)
STREAM_SRC="rtsp://127.0.0.1:8554/fake"

echo "Adding stream '$STREAM_NAME' to go2rtc at $GO2RTC_API..."

# Use -G to ensure data is sent as query params if needed, but here we just construct URL
curl -v -X PUT "${GO2RTC_API}/api/streams?src=${STREAM_SRC}&name=${STREAM_NAME}"

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Success! Stream '$STREAM_NAME' added."
    echo "👉 Now check ThingsPanel Device List."
    echo "👉 Wait 30s. A new device '$STREAM_NAME' should appear."
    echo "👉 Check 'Attributes' tab for 'stream_url'."
else
    echo ""
    echo "❌ Failed to add stream."
fi
