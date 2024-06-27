# 开发帮助
## go版本
>1.22
## SDK升级

go get -u github.com/ThingsPanel/tp-protocol-sdk-go@latest
go get -u github.com/ThingsPanel/tp-protocol-sdk-go@v1.2.0

## 开发说明
-  查看**services/开发说明.md**


## 其他
### 测试命令


mosquitto_pub -h 47.115.210.16 -p 1883 -t "devices/telemetry" -m "{\"temp\":12.5}" -u "c55d8498" -P "c55d8498-e01e" -i "0"
mosquitto_pub -h 47.115.210.16 -p 1883 -t "devices/telemetry" -m "{\"temp\":12.5}" -u "c55d8498" -P "c55d8498-e01e" -i "0"
