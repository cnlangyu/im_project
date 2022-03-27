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
	Conn       net.Conn
	Flag       int
}

func CreateClient(serverIp string, serverPort int) *Client {
	//创建客户端
	client := Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		Name:       fmt.Sprintf("%s:%d", serverIp, serverPort),
		Flag:       -1,
	}

	//创建服务器连接
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("连接服务器失败，错误信息:", err)
		return nil
	}
	fmt.Println("连接服务器成功...")

	client.Conn = conn

	return &client
}

var _serverIp string
var _serverPort int

//解析命令行
func init() {
	flag.StringVar(&_serverIp, "ip", "127.0.0.1", "设置连接服务器IP,默认127.0.0.1")
	flag.IntVar(&_serverPort, "port", 8888, "设置连接服务器Port,默认8888")
}

//菜单
func (client *Client) menu() {
	fmt.Println("1. 私聊模式")
	fmt.Println("2. 公聊模式")
	fmt.Println("3. 更新用户名")
	fmt.Println("0. 退出")

	fmt.Scanln(&client.Flag)
}

//监听消息
func (client *Client) DealResponse() {
	//一旦client.conn存在消息，直接copy到控制台
	io.Copy(os.Stdin, client.Conn)

}

//修改用户名称
func (client *Client) ModifyName() bool {
	fmt.Println(">>>>> 输入用户名称:")
	var modifyName string
	fmt.Scanln(&modifyName)
	msg := "rename|" + modifyName + "\n"
	_, err := client.Conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("修改用户名出错:", err)
		return false
	}
	client.Name = modifyName
	return true
}

//公聊模式
func (client *Client) PublicChat() {
	for {
		var msg string
		fmt.Println("请输入内容:")
		fmt.Scanln(&msg)
		if msg == "exit" {
			break
		} else if len(msg) > 0 {
			_, err := client.Conn.Write([]byte(msg + "\n"))
			if err != nil {
				fmt.Println("发送消息失败...")
			}
		}
	}
}

//查询用户
func (client *Client) FindUser() bool {
	var sendMsg = "who\n"
	_, err := client.Conn.Write([]byte(sendMsg))
	return err == nil
}

//私聊模式
func (client *Client) PrivateChat() {
	if !client.FindUser() {
		fmt.Println("获取用户列表失败, 请重试...")
		return
	}
	for {
		var chatUser = ""
		fmt.Println("输入私聊用户名(exit退出):")
		fmt.Scanln(&chatUser)
		fmt.Printf("给用户:%s 开始私聊...\n", chatUser)
		if chatUser == "exit" {
			break
		}

		for {
			var sendMsg = ""
			fmt.Println("发送消息(exit退出):")
			fmt.Scanln(&sendMsg)
			if sendMsg == "exit" {
				break
			}
			var _sendMsg = "to|" + chatUser + "|" + sendMsg + "\n"
			_, err := client.Conn.Write([]byte(_sendMsg))
			if err != nil {
				fmt.Printf("向%s发送消息失败\n", chatUser)
			}
		}
	}
}

//运行
func (client *Client) Run() {
	for ; client.Flag != 0; client.menu() {
		switch client.Flag {
		case 1:
			client.PrivateChat()
		case 2:
			client.PublicChat()
		case 3:
			client.ModifyName()
		default:
			fmt.Println("请选择正确的菜单选项")
		}
		client.Flag = -1
	}
}

func main() {
	//命令行解析
	flag.Parse()

	fmt.Printf("获取命令行数据: %s:%d \n", _serverIp, _serverPort)
	client := CreateClient(_serverIp, _serverPort)

	if client == nil {
		fmt.Println(">>>>>> 创建客户端失败...")
		return
	}
	fmt.Println(">>>>> 创建客户端成功...")

	go client.DealResponse()

	//运行
	client.Run()
}
