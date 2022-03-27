package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

//监听发送消息
func (server *Server) ListenAndPublishMessage() {
	fmt.Println("监听发送消息启动...")
	for {
		msg := <-server.Message
		server.mapLock.Lock()
		for _, user := range server.OnlineMap {
			user.Channel <- msg
		}
		server.mapLock.Unlock()
	}
}

/**
 * 创建一个Server对象
 */
func NewServer(ip string, port int) *Server {
	// server := &Server{ip: ip, port: port}
	// return server
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

//广播消息
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "user[" + user.Name + "] send : " + msg
	server.Message <- sendMsg
}

func (server *Server) Handler(conn net.Conn) {
	//处理当前连接的业务
	fmt.Println("连接建立成功...")
	user := NewUser(conn, server)
	//用户上线
	user.Online()

	//接收客户端发送的消息
	go user.ListeningSendMsg()
	//当前handler阻塞
	for {
		select {
		case <-user.IsLive:

		case <-time.After(10 * time.Minute):
			user.DelOnline()
			//被踢掉
			user.CallBackMsg("你被踢下线")
			close(user.Channel)
			user.conn.Close()

			return
		}
	}
}

/**
 * 启动服务器的接口
 */
func (server *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}
	//close listen socket
	defer listener.Close()

	go server.ListenAndPublishMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		//do hander
		go server.Handler(conn)
	}
}
