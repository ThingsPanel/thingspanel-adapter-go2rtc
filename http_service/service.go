package httpservice

import (
	"encoding/json"
	"errors"
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
	// service_identifier := r.URL.Query()["protocol_type"][0]
	// 根据需要对服务标识符进行验证，可不验证
	// if service_identifier != "xxxx" {
	// 	RspError(w, fmt.Errorf("not support protocol type: %s", service_identifier))
	// 	return
	// }
	//CFG配置表单 VCR凭证表单 SVCR服务凭证表单
	switch form_type {
	case "VCR":
		if device_type == "1" {
			// 设备凭证表单
			RspSuccess(w, readFormConfigByPath("./form_voucher.json"))
		} else {
			RspSuccess(w, nil)
		}
	case "SVCR":
		if device_type == "1" {
			//服务凭证类型表单
			RspSuccess(w, readFormConfigByPath("./form_service_voucher.json"))
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
