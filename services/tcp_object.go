package services

import (
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

type TCPObject struct {
	Mutex     *sync.Mutex //互斥锁
	Conn      net.Conn    //连接
	DeviceID  string      //设备在平台的唯一标识
	UID       string      //硬件设备唯一标识
	ReplyFlag bool        //是否需要回复
}

// 当设备与平台连接的时候，开启一个goroutine接管连接
func NewTCPObject(conn net.Conn) {

	w := &TCPObject{
		Conn:  conn,
		Mutex: &sync.Mutex{},
	}
	w.Start()
}

func (w *TCPObject) Start() {
	// 鉴权
	// make a buffer to hold incoming data
	buf := make([]byte, 1024)
	// read the incoming connection into the buffer
	reqLen, err := w.Conn.Read(buf)
	if err != nil {
		logrus.Println("Error reading:", err.Error())
	}
	logrus.Println("Received data:", string(buf[:reqLen]))
	// 处理数据
}
