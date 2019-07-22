package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// 创建全局 读写锁 保护 onLineMap
var rwMutex sync.RWMutex		// 有空间

// 定义客户端结构体
type Client struct {
	Name string
	Addr string
	C chan string
}

// 全局在线用户列表 - Map[string]Client
var onLineMap = make(map[string]Client)

// 定义全局 channel -- Message
var Message = make(chan string)

func WriteMsgToClient(clnt Client, conn net.Conn)  {
	/*	for {
		msg := <-clnt.C		// 无数据，阻塞； 有数据，继续
		conn.Write([]byte(msg + "\n"))
	}*/

	for msg := range clnt.C {
		conn.Write([]byte(msg + "\n"))
	}
}

// 封装函数，组织广播消息
func makeMsg(clnt Client, str string) string {

	msg := "[" + clnt.Addr + "]" + ":" + clnt.Name + ":" + str

	return msg
}

// 与客户端进行数据交互的 go程
func handleConnect(conn net.Conn)  {

	defer conn.Close()

	// 获取客户端地址结构
	clntAddr := conn.RemoteAddr().String()

	// 组织客户端 信息 -- 初始化  Name == Addr == Ip + port
	clnt := Client{clntAddr, clntAddr, make(chan string)}

	rwMutex.Lock()
	// 将客户端信息添加至全局map  —— 写
	onLineMap[clntAddr] = clnt
	rwMutex.Unlock()

	// 创建 读用户自带channel的 go程
	go WriteMsgToClient(clnt, conn)

	// 组织用户上线消息
	//msg := "[" + clntAddr + "]" + ":" + clnt.Name + ":" + "Login"
	msg := makeMsg(clnt, "Login")

	// 广播： 将msg 写到 全局 Channel —— Message
	Message <- msg

	// 创建 管理用户下线的channel
	isQuit := make(chan bool)

	// 创建 描述用户为活跃用户的channel （重置After 计时器）
	isLive := make(chan bool)

	// 创建匿名 go程 专门监听客户端发送的聊天信息
	go func() {

		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			if n == 0 {
				//fmt.Println("客户端下线！")
				isQuit <- true		// 写入channel 代表用户下线
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Listen err:", err)
				return
			}
			// fmt.Println("测试----：", buf[:n])
			msg := string(buf[:n-1])

			// 判断 用户是否是一个 查询在线 用户命令
			if "who" == msg {

				rwMutex.RLock()			// 加读锁
				// 遍历在线用户列表
				for _, clnt := range onLineMap {
					// 组织用户现在信息
					onLineMsg := makeMsg(clnt, "[OnLine]\n")
					// 写给当前用户
					conn.Write([]byte(onLineMsg))
				}
				rwMutex.RUnlock()		// 解锁

			} else if len(msg) > 7 && msg[:7] == "rename|" {
				//} else if msg[:7] == "rename|" && len(msg) > 7 {

				// 按管道 拆分命令，得到新用户名
				newName := strings.Split(msg, "|")[1]

				// 新用户名 覆盖旧用户名
				clnt.Name = newName

				rwMutex.Lock()
				// 将 持有新用户名的用户 ，更新到全局map上
				onLineMap[clnt.Addr] = clnt
				rwMutex.Unlock()

				// 提示用户改名成功 —— 不广播
				conn.Write([]byte("rename successfull!!!\n"))

			} else {
				// 组织读到的用户聊天数据
				msg = makeMsg(clnt, msg)

				// 写给全局 Message  —— 广播
				Message <- msg
			}

			isLive <- true 	// 证明当前用户是活跃用户
		}

	}()

	// 添加 for 防止 该处理客户端事件go程 提前退出。
	for {
		select {
		case <-isQuit :			// 说明用户下线
			// 关闭用户自带 channel 写端。—— 导致 WriteMsgToClient go程 终止
			close(clnt.C)

			rwMutex.Lock()
			// 从在线用户列表中 移除当前用户
			delete(onLineMap, clnt.Addr)
			rwMutex.Unlock()

			// 组织用户下线消息
			msg = makeMsg(clnt, "Logout")

			// 写给全局 Message  —— 广播
			Message <- msg

			// 结束 当前 handleConnect go程 —— 不能使用 break
			return

		case <-isLive:		// 能重置After 计时器

			// 什么也不作，目的能重置After 计时器

		case <-time.After(3600*time.Second):
			// 关闭用户自带 channel 写端。—— 导致 WriteMsgToClient go程 终止
			close(clnt.C)

			rwMutex.Lock()
			// 从在线用户列表中 移除当前用户
			delete(onLineMap, clnt.Addr)
			rwMutex.Unlock()

			// 组织用户下线消息
			msg = makeMsg(clnt, "time out To Leave!")

			// 写给全局 Message  —— 广播
			Message <- msg

			// 结束 当前 handleConnect go程 —— 不能使用 break
			return
		}
	}
}


// 全局 Manager
func Manager()  {

	// for 监听全局Message, 无数据，阻塞； 有数据，继续
	for {
		msg := <-Message

		rwMutex.RLock()		// 遍历 在线用户列表之前，加读锁
		// 遍历在线用户列表map，写消息msg给用户自带 channel
		for _, clnt := range onLineMap {
			clnt.C <- msg
		}
		rwMutex.RUnlock()	// 遍历 完成，解锁
	}
}

func main() {
	// 创建监听socket
	listener, err := net.Listen("tcp", "127.0.0.1:8880")
	if err != nil {
		fmt.Println("Listen err:", err)
		return
	}
	defer listener.Close()

	// 创建全局Manager go程
	go Manager()

	// 循环监听客户端连接请求
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept err:", err)
			continue
		}

		// 创建 go程 专门与该客户端进行数据通信
		go handleConnect(conn)
	}
}
