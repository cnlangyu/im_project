package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

type User struct {
	Name    string
	Addr    string
	Channel chan string
	conn    net.Conn

	server *Server
	IsLive chan bool
}

//创建一个用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:    userAddr,
		Addr:    userAddr,
		Channel: make(chan string),
		conn:    conn,
		server:  server,
		IsLive:  make(chan bool),
	}

	//启动监听当前user channel消息的goroutine
	go user.ListenReceiveMessage()
	return user
}

//收消息
//监听当前User channel的方法，一旦有消息，就直接发送给客户端
func (user *User) ListenReceiveMessage() {
	for {
		msg := <-user.Channel
		user.conn.Write([]byte(msg + "\n"))
	}
}

//上线
func (user *User) Online() {
	//用户上线，将用户加入到OnlineMap中
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()
	//广播当前用户上线消息
	user.server.BroadCast(user, "我上线啦")
}

//下线
func (user *User) Offline() {
	user.DelOnline()
	user.server.BroadCast(user, "我下线")
}

//清除我的在线记录
func (user *User) DelOnline() {
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()
}

//给自己发消息
func (user *User) CallBackMsg(msg string) {
	user.conn.Write([]byte(msg))
}

//查询用户列表
func (user *User) OnlineUsers() {
	user.server.mapLock.Lock()
	for _, u := range user.server.OnlineMap {
		// sendMsg := "用户[" + u.Addr + "]" + u.Name + "\n"
		sendMsg := "用户[" + u.Name + "]" + " 地址:" + u.Addr + "\n"
		user.CallBackMsg(sendMsg)
	}
	user.server.mapLock.Unlock()
}

//修改名称
func (user *User) ReName(rename string) {
	name := strings.Split(rename, "|")[1]
	_, ok := user.server.OnlineMap[name]
	if ok {
		user.CallBackMsg("名称:" + name + "，已经存在\n")
	} else {
		user.server.mapLock.Lock()
		delete(user.server.OnlineMap, user.Name)
		user.server.OnlineMap[name] = user
		user.server.mapLock.Unlock()
		user.Name = name
		user.CallBackMsg("名字成功修改为：" + name + "\n")
	}
}

//发消息
func (user *User) ListeningSendMsg() {
	buf := make([]byte, 8*1024)
	for {
		n, err := user.conn.Read(buf)
		//下线
		if n == 0 {
			user.Offline()
			return
		}
		//错误
		if err != nil && err != io.EOF {
			fmt.Println("Conn Read err:", err)
			return
		}
		//去除消息中的回车
		msg := string(buf[:n-1])
		if msg == "who" {
			user.OnlineUsers()
		} else if len(msg) > 7 && msg[:7] == "rename|" {
			user.ReName(msg)
		} else if len(msg) > 3 && msg[:3] == "to|" && len(strings.Split(msg, "|")) == 3 {
			privateMsg := strings.Split(msg, "|")
			toUserName := privateMsg[1]
			if len(toUserName) < 1 {
				user.CallBackMsg("用户名错误")
				continue
			}
			toUser, ok := user.server.OnlineMap[toUserName]
			if !ok {
				user.CallBackMsg("没有找到用户")
				continue
			}
			toUser.CallBackMsg("user[" + user.Name + "] to me : " + privateMsg[2] + "\n")

		} else {
			//广播消息
			user.server.BroadCast(user, msg)
		}
		user.IsLive <- true
	}

}
