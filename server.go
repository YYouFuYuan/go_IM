package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex //锁

	//消息广播的channel
	Message chan string
}

//创建一个server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

//监听message广播消息的goroutine
func (this *Server) LinstenMessage() {
	for {
		//从channel取出消息
		msg := <-this.Message
		//所有用户广播，发送该消息
		this.mapLock.Lock()
		for _, client := range this.OnlineMap {
			client.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//服务器广播消息
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//处理链接

	//用户上线，将用户加入map
	user := NewUser(conn, this) //创建用户
	user.Online()

	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err : ", err)
				return
			}
			//提取用户消息(去除“\n”)
			msg := string(buf[:n-1])
			//将得到的消息进行处理
			user.DoMessage(msg)
			//更新活跃通道
			user.isLive <- true
		}
	}()
	//先阻塞
	for {
		select {
		case <-user.isLive:
			//当前用户是活跃的，应该重置定时器
		case <-time.After(time.Second * 300):
			//已经超时
			//当前用户强制下线
			user.SendMessage("force offline\n")
			//删除用户
			delete(this.OnlineMap, user.Name)
			//关闭用户监听的通道
			close(user.C)
			close(user.isLive)
			//关闭链接
			conn.Close()

			return //关闭该协程
		}
	}

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

	//启动监听message
	go this.LinstenMessage()
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
