package httpservice

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	tpprotocolsdkgo "github.com/ThingsPanel/tp-protocol-sdk-go"
	"github.com/spf13/viper"
)

var HttpClient *tpprotocolsdkgo.Client

func Init() {
	go start()
}

func start() {
	var handler tpprotocolsdkgo.Handler = tpprotocolsdkgo.Handler{
		OnCreateDevice: OnCreateDevice,
		OnUpdateDevice: OnUpdateDevice,
		OnDeleteDevice: OnDeleteDevice,
		OnGetForm:      OnGetForm,
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
	log.Println("OnGetForm")
	r.ParseForm() //解析参数，默认是不会解析的
	log.Println("【收到api请求】path", r.URL.Path)
	log.Println("scheme", r.URL.Scheme)
	var rsp = make(map[string]interface{})
	rsp["data"] = readFormConfig()
	data, err := json.Marshal(rsp)
	if err != nil {
		log.Println(err.Error())
	}
	fmt.Fprint(w, string(data)) //这个写入到w的是输出到客户端的
}

func readFormConfig() interface{} {
	filePtr, err := os.Open("./form_config.json")
	if err != nil {
		log.Println("文件打开失败...", err.Error())
		return nil
	}
	defer filePtr.Close()
	var info interface{}
	// 创建json解码器
	decoder := json.NewDecoder(filePtr)
	err = decoder.Decode(&info)
	if err != nil {
		log.Println("解码失败", err.Error())
		return info
	} else {
		log.Println("读取文件[form_config.json]成功...")
		return info
	}
}

// OnCreateDevice 创建设备
func OnCreateDevice(w http.ResponseWriter, r *http.Request) {
	log.Println("OnCreateDevice")
	r.ParseForm() //解析参数，默认是不会解析的
	log.Println("【收到api请求】path", r.URL.Path)
	log.Println("scheme", r.URL.Scheme)
	// 读取客户端发送的数据
	var reqDataMap = make(map[string]interface{})
	if err := json.NewDecoder(r.Body).Decode(&reqDataMap); err != nil {
		r.Body.Close()
		w.WriteHeader(400)
		return
	}
	// 逻辑处理。。。
	// 返回成功
	var rspdata = make(map[string]interface{})
	w.Header().Set("Content-Type", "application/json")
	rspdata["code"] = 200
	rspdata["message"] = "success"
	data, err := json.Marshal(rspdata)
	if err != nil {
		log.Println(err.Error())
	}
	fmt.Fprint(w, string(data))
}

// OnUpdateDevice 更新设备
func OnUpdateDevice(w http.ResponseWriter, r *http.Request) {
	log.Println("OnUpdateDevice")
	r.ParseForm() //解析参数，默认是不会解析的
	log.Println("【收到api请求】path", r.URL.Path)
	log.Println("scheme", r.URL.Scheme)
	// 读取客户端发送的数据
	var reqDataMap = make(map[string]interface{})
	if err := json.NewDecoder(r.Body).Decode(&reqDataMap); err != nil {
		r.Body.Close()
		w.WriteHeader(400)
		return
	}
	// 逻辑处理。。。
	// 返回成功
	var rspdata = make(map[string]interface{})
	w.Header().Set("Content-Type", "application/json")
	rspdata["code"] = 200
	rspdata["message"] = "success"
	data, err := json.Marshal(rspdata)
	if err != nil {
		log.Println(err.Error())
	}
	fmt.Fprint(w, string(data))
}

// OnDeleteDevice 删除设备
func OnDeleteDevice(w http.ResponseWriter, r *http.Request) {
	log.Println("OnDeleteDevice")
	r.ParseForm() //解析参数，默认是不会解析的
	log.Println("【收到api请求】path", r.URL.Path)
	log.Println("scheme", r.URL.Scheme)
	// 读取客户端发送的数据
	var reqDataMap = make(map[string]interface{})
	if err := json.NewDecoder(r.Body).Decode(&reqDataMap); err != nil {
		r.Body.Close()
		w.WriteHeader(400)
		return
	}
	// 逻辑处理。。。
	// 返回成功
	var rspdata = make(map[string]interface{})
	w.Header().Set("Content-Type", "application/json")
	rspdata["code"] = 200
	rspdata["message"] = "success"
	data, err := json.Marshal(rspdata)
	if err != nil {
		log.Println(err.Error())
	}
	fmt.Fprint(w, string(data))
}
