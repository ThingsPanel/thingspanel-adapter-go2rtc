package httpservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	tpprotocolsdkgo "github.com/ThingsPanel/tp-protocol-sdk-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var HttpClient *tpprotocolsdkgo.Client

func Init() {
	go start()
}

func start() {
	var handler tpprotocolsdkgo.Handler = tpprotocolsdkgo.Handler{
		OnDisconnectDevice: OnDisconnectDevice,
		OnGetForm:          OnGetForm,
	}
	addr := viper.GetString("http_server.address")
	log.Println("http服务启动：", addr)
	err := handler.ListenAndServe(addr)
	if err != nil {
		log.Println("ListenAndServe() failed, err: ", err)
		return
	}
}

// OnGetForm 获取协议插件的json表单
func OnGetForm(w http.ResponseWriter, r *http.Request) {
	logrus.Info("OnGetForm")
	r.ParseForm() //解析参数，默认是不会解析的
	logrus.Info("【收到api请求】path", r.URL.Path)
	logrus.Info("query", r.URL.Query())

	device_type := r.URL.Query()["device_type"][0]
	form_type := r.URL.Query()["form_type"][0]
	protocol_type := r.URL.Query()["protocol_type"][0]
	// 如果请求参数protocol_type不等于MMindjoy-WM，返回空
	if protocol_type != "Mindjoy-WM" {
		RspError(w, fmt.Errorf("not support protocol type: %s", protocol_type))
		return
	}
	//CFG配置表单 VCR凭证表单 VCRT凭证类型表单
	switch form_type {
	case "VCR":
		if device_type == "1" {
			// 设备凭证表单
			RspSuccess(w, readFormConfigByPath("./form_voucher.json"))
		} else {
			RspSuccess(w, nil)
		}
	case "VCRT":
		if device_type == "1" {
			//设备凭证类型表单
			RspSuccess(w, readFormConfigByPath("./form_voucher_type.json"))
		} else {
			RspSuccess(w, nil)
		}
	default:
		RspError(w, errors.New("not support form type: "+form_type))
	}
}

func OnDisconnectDevice(w http.ResponseWriter, r *http.Request) {
	logrus.Info("OnDisconnectDevice")
	r.ParseForm() //解析参数，默认是不会解析的
	logrus.Info("【收到api请求】path", r.URL.Path)
	logrus.Info("query", r.URL.Query())
	// 断开设备

	//RspSuccess(w, nil)
}

// ./form_config.json
func readFormConfigByPath(path string) interface{} {
	filePtr, err := os.Open(path)
	if err != nil {
		logrus.Warn("文件打开失败...", err.Error())
		return nil
	}
	defer filePtr.Close()
	var info interface{}
	// 创建json解码器
	decoder := json.NewDecoder(filePtr)
	err = decoder.Decode(&info)
	if err != nil {
		logrus.Warn("解码失败", err.Error())
		return info
	} else {
		logrus.Info("读取文件[form_config.json]成功...")
		return info
	}
}
