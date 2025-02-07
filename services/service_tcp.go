package services

import (
	"net"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// 定义全局的conn管道

func StartTCP() {
	logrus.Println("Launching tcp server...")
	host := viper.GetString("server.address")
	if host == "" {
		host = ":9111"
	}
	// set up listener on localhost port 8080
	ln, err := net.Listen("tcp", host)
	if err != nil {
		logrus.Println("Error listening:", err.Error())
		panic(err)
	}
	logrus.Println("Listening on " + host)

	for {
		// accept connection
		conn, err := ln.Accept()
		if err != nil {
			logrus.Println("Error accepting: ", err.Error())
			continue
		}
		logrus.Println("Connection established")

		// start a new goroutine to handle the connection

		go dealTCPData(conn)

		logrus.Println("New connection")
	}
}

// 处理tcp数据
func dealTCPData(conn net.Conn) {
	NewTCPObject(conn)
}
