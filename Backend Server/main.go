package main

import (
	"backendserver/service"
	"backendserver/utility"
	"fmt"
	"net"
	"net/rpc"
	"os"

	"github.com/murasakiakari/pathlib"
)

func main() {
	var err error
	defer func() {
		if err != nil {
			fmt.Printf("exit with error: %v\n", err.Error())
		}
	}()

	logPath := pathlib.CurrentExecutablePath.Dir().Join(utility.Config.Log.FileName)
	logFile, err := logPath.OpenFile(os.O_APPEND|os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return
	}

	logger := utility.NewLogger(logFile)
	defer func() {
		if err != nil {
			logger.Errorln(err.Error())
		}
		logFile.Close()
	}()

	service, err := service.New()
	if err != nil {
		return
	}

	rpc.RegisterName("Service", service)
	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		return
	}

	fmt.Println("backend server running...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
					logger.Errorln(err)
				}
			}()
			rpc.ServeConn(conn)
		}()
	}
}
