# Go2RTC Adapter 迭代日志

> 本文档记录部署测试过程中的每次迭代，包括遇到的问题及解决方案。

## 迭代 1: 初始化部署 (2026-02-06)

### 目标
- 构建并部署 go2rtc adapter 到测试服务器
- 验证服务正常运行

### 执行步骤

#### 1.1 本地构建
- [x] 项目分析完成
- [ ] Cross-compile for Linux

#### 1.2 部署到服务器
- [ ] SSH 部署 (deploy.exp)
- [ ] 验证进程状态
- [ ] 验证 API 响应

### 发现的问题

#### Bug #1: MQTT连接失败
- **现象**: 服务启动后立即退出，Exit code 1
- **日志**: `MQTT连接失败: network Error : EOF`
- **原因**: config.yaml 中 MQTT broker 使用 `127.0.0.1:1883`，但 ThingsPanel 运行在 Docker 容器中
- **参考**: [go2rtc GitHub](https://github.com/AlexxIT/go2rtc), [ThingsPanel开发文档](https://docs.thingspanel.cn/zh-Hans/docs/developer-guide/developing-plug-in/customProtocol)

### 解决方案
修改 `configs/config.yaml`:
- `mqtt_broker`: `tcp://127.0.0.1:1883` → `tcp://172.17.0.1:1883` (Docker桥接网关)
- `url`: `http://127.0.0.1:9999` → `http://172.17.0.1:9999`

---

## 迭代 2: 服务启动失败排查 (2026-02-06)

### 发现的问题

#### Bug #2: Text file busy
- **现象**: `nohup: failed to run command './tp-adapter-linux': Text file busy`
- **原因**: 旧进程仍在运行或文件句柄未释放，上传覆盖的文件无法执行
- **参考**: [Linux Text file busy error](https://stackoverflow.com/questions/16764946)

### 解决方案
1. 修改 `deploy.exp` - 在上传前先杀死旧进程
2. 如果仍有问题，删除旧二进制文件后重新上传

### 解决状态: ✅ 已解决

---

## 迭代 3: MQTT 连接问题 (2026-02-06)

### 发现的问题

#### Bug #3: MQTT 服务器不可达
- **现象**: `MQTT连接失败: network Error : EOF`
- **原因**: 192.168.31.205 服务器上 ThingsPanel/MQTT 可能未运行
- **检查项**: 
  - Docker 是否运行
  - 172.17.0.1:1883 是否可访问
  - ThingsPanel 服务是否正常

### 解决方案
修改 `configs/config.yaml`:
- `mqtt_password`: `change_me` → `root` (从 gmqtt 配置中获取)

### 解决状态: ✅ 已修复

---

## 🎉 最终验证结果 (2026-02-06)

### 服务状态
- ✅ 进程运行: PID 4351
- ✅ HTTP 端口 12000 监听中
- ✅ 协议端口 12001 监听中
- ✅ MQTT 连接成功
- ✅ 心跳发送成功

### API 测试
1. **设备列表**: `curl http://127.0.0.1:12000/api/v1/plugin/device/list?...`
   - 响应: `{"code":200,"message":"success","data":{"list":[],"total":0}}`
   - ✅ 返回空数组 `[]` (不是 null)

2. **表单配置**: `curl http://127.0.0.1:12000/api/v1/form/config?form_type=SVCR&...`
   - 响应: `{"code":200,"message":"success","data":[...表单配置...]}`
   - ✅ 正确返回 SVCR 表单

---

## 迭代 4: 修复配置修改时的 404 错误 (2026-02-06)

### 发现的问题

#### Bug #4: 点击"下一步"报 404 错误
- **现象**: 修改配置 -> 选择自动 -> 下一步，报错 `protocol plugin response message: 404 Not Found`
- **原因**: 平台发送的通知接口地址是 `/api/v1/notify/event`，但 SDK (v1.2.6) 注册的路由是 `/api/v1/plugin/notification`。路径不匹配导致 404。
- **排查**:
  - `inspect_sdk` 工具显示 SDK 确实没有 `/api/v1/notify/event` 的 handler。
  - WVP 插件使用了旧版 SDK 并手动处理了该路由。

### 解决方案
在 `internal/bootstrap/http.go` 中增加 HTTP 中间件，进行 **URL 重写**:
```go
// URL重写: /api/v1/notify/event -> /api/v1/plugin/notification
if r.URL.Path == "/api/v1/notify/event" {
    r.URL.Path = "/api/v1/plugin/notification"
}
```

### 验证结果
- ✅ 模拟请求: `curl -X POST .../api/v1/notify/event` 返回 200 OK
- ✅ 日志确认: `DEBUG ... URL Rewriting: /api/v1/notify/event -> /api/v1/plugin/notification`

### 解决状态: ✅ 已修复

---

## 迭代 5: 属性上报功能实现与文档更新 (2026-02-06)

### 目标
- 实现视频流地址 (`stream_url`) 自动上报到 ThingsPanel 设备属性
- 完善 `README.md`，增加 Go2RTC 部署及完整测试流程

### 发现的问题

#### Bug #5: 属性上报无数据
- **现象**: Adapter 日志显示 `SendAttributes` 调用成功，但平台页面"属性"栏为空。
- **原因**: 
  1. **Topic 格式错误**: 之前使用 `devices/attributes`，平台要求必须带 unique ID 后缀（如 `devices/attributes/<timestamp>`）。
  2. **逻辑缺陷**: 当设备已存在时，`registerDevice` 返回 "Device exists" 错误并提前退出，导致属性上报逻辑被跳过。
- **排查手段**: 
  - `grep` 日志发现 "上报属性成功" 缺失。
  - 分析 `sync.go` 逻辑发现错误处理分支直接 return。

### 解决方案
1. **修正 Topic**: 修改 `platform.go`，在属性 Topic 后追加时间戳 ID。
   ```go
   topic := "devices/attributes/" + messageID
   ```
2. **优化注册逻辑**: 修改 `sync.go`，当捕获 "设备已存在" 错误时，不退出，而是调用 `GetDevice` 获取 ID，并继续执行属性上报。
3. **新增测试脚本**: 创建 `tests/simulate_device.sh`，方便快速添加模拟流进行验证。

### 验证结果
- ✅ 执行 `./simulate_device.sh` 添加流
- ✅ 日志显示: `上报属性成功: stream_url=rtsp://...`
- ✅ 平台 UI: 属性页正确显示 `stream_url`

### 解决状态: ✅ 已完成

---

## 参考资料
1. [go2rtc GitHub](https://github.com/AlexxIT/go2rtc)
2. [ThingsPanel 协议插件开发文档](https://docs.thingspanel.cn/zh-Hans/docs/developer-guide/developing-plug-in/customProtocol)
3. [tp-protocol-sdk-go](https://github.com/ThingsPanel/tp-protocol-sdk-go)

