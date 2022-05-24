package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string //用户发送消息的通道
	isLive chan bool   //用户活跃通道
	conn   net.Conn
	server *Server //该用户所对应的server
}

//创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		isLive: make(chan bool),
		conn:   conn,
		server: server,
	}
	//启动监听当前user channel消息的协程
	go user.ListenMessage()
	return user
}

//监听当前User channel的方法，一旦有消息，就直接发送给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}

//用户上线
func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播当前用户上线消息
	this.server.BroadCast(this, "already online")
}

//用户下线
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播当前用户下线消息
	this.server.BroadCast(this, "already offline")
}

//用户处理消息的业务
func (this *User) DoMessage(msg string) {

	if msg == "who" {
		//查询当前用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "online...\n"
			this.SendMessage(onlineMsg)
		}
		this.server.mapLock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" { //更新用户名
		//取出名字
		newName := strings.Split(msg, "|")[1]
		//判断name是否已经存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMessage("current username already exist\n")
		} else {
			//把新的用户名更新一下
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMessage("your username is update:" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//私聊消息格式 ： to|张三|消息内容
		//1. 获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMessage("message format is error,please use \"to|name|msg\" \n")
			return
		}

		//2. 根据用户名得到目标对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMessage("remoteUser is not exist\n")
			return

		}
		//3. 获取消息，并给目标用户发送消息
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMessage("no message,please send again\n")
			return
		}
		remoteUser.SendMessage(this.Name + " send for you:" + content + "\n")

	} else {
		this.server.BroadCast(this, msg)
	}
}

//给当前用户发送消息
func (this *User) SendMessage(msg string) {
	this.conn.Write([]byte(msg))
}
