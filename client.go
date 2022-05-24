package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	//链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error :", err)
	}
	client.conn = conn
	//返回对象
	return client
}

//处理server回应的消息，显示到控制台
func (this *Client) DealServerResponse() {
	//一旦conn有数据发过来，就直接拷贝到标准输出流，永久阻塞
	io.Copy(os.Stdout, this.conn) //等价于下面代码

	//for {
	//	buf := make([]byte, 4096)
	//	this.conn.Read(buf)
	//	fmt.Println(string(buf))
	//}
}

//客户端菜单
func (this *Client) menu() bool {
	var flag int
	fmt.Println("1.public chat")
	fmt.Println("2.private chat")
	fmt.Println("3.update username")
	fmt.Println("0.exit")
	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		this.flag = flag
		return true
	} else {
		fmt.Println(">>>>>please enter a value number<<<<<<")
		return false
	}
}

// 查询在线用户
func (this *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn write err:", err)
		return
	}
}

//私聊模式
func (this *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	this.SelectUsers()
	fmt.Println(">>>>please enter [username],\"exit\" exit model")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>please enter message，\"exit\" exit model")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			//发给服务器

			//消息不为空就发
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := this.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>please enter message，\"exit\" exit model")
			fmt.Scanln(&chatMsg)
		}
		this.SelectUsers()
		fmt.Println(">>>>please enter [username],\"exit\" exit model")
		fmt.Scanln(&remoteName)
	}
}

//公聊模式具体处理函数
func (this *Client) PublicChat() {
	//提示用户输入信息
	var chatMsg string
	fmt.Println(">>>please enter message，\"exit\" exit model")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//发给服务器

		//消息不为空就发
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn write err:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>please enter message，\"exit\" exit model")
		fmt.Scanln(&chatMsg)
	}
}

//更改用户名具体处理函数
func (this *Client) Updatename() bool {
	fmt.Println(">>> please enter username <<<<")
	fmt.Scanln(&this.Name)
	sendMsg := "rename|" + this.Name + "\n"
	//发送消息
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (this *Client) Run() {
	for this.flag != 0 {
		for this.menu() != true {

		}
		// 根据不同的模式处理不同的业务
		switch this.flag {
		case 1:
			//启动公聊模式
			this.PublicChat()
			break
		case 2:
			//私聊模式
			this.PrivateChat()
			break
		case 3:
			//更新用户名
			this.Updatename()
			break
		}
	}
}

var serverIp string
var serverPort int

//解析命令行 ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址，默认(127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(8888)")
}

func main() {
	//命令行解析
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>>>link server error.........")
		return
	}
	//接收服务端消息
	go client.DealServerResponse()
	fmt.Println(">>>>>>>>>link server success........")

	//启动客户端业务
	client.Run()

}
