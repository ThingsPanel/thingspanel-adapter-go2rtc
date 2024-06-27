package services

import (
	httpclient "earth/http_client"
	"fmt"

	"github.com/ThingsPanel/tp-protocol-sdk-go/api"
	"github.com/sirupsen/logrus"
)

// 认证设备并获取设备信息
func AuthDevice(deviceSecret string) (deviceInfo *api.DeviceConfigResponse, err error) {
	voucher := AssembleVoucher(deviceSecret)
	// 读取设备信息
	deviceInfo, err = httpclient.GetDeviceConfig(voucher, deviceSecret)
	if err != nil {
		// 获取设备信息失败，请检查连接包是否正确
		logrus.Error(err)
		return
	}
	return
}

// 凭证信息组装
func AssembleVoucher(deviceSecret string) (voucher string) {
	return fmt.Sprintf(`{"UID":"%s"}`, deviceSecret)
}
