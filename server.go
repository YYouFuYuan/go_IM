package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

//创建一个server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}
	return server
}

func (this *Server) Handler(conn net.Conn) {
	//处理链接
	fmt.Println("处理流程")
}

//启动服务器的接口
func (this *Server) Start() {
	// socket listen
	listener, error := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if error != nil {
		fmt.Println("net.Listener error:", error)
		return
	}
	// close listen socket
	defer listener.Close()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listenner accept err : ", err)
			continue
		}

		// do handler 有连接了，开启一个协程进行处理
		go this.Handler(conn)

	}

}
