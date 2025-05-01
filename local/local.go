package local

import (
	"log"
	"net"
	"time"

	"github.com/beijian128/minisocks/core"
)

// LsLocal 表示本地代理服务端，负责处理本地浏览器的代理请求。
type LsLocal struct {
	*core.SecureSocket      // 嵌入 SecureSocket 结构体，用于数据的加密和解密传输
	running            bool // 标识本地代理服务是否正在运行
	// AfterListen 是一个回调函数，在本地代理开始监听后被调用，传入监听地址
	AfterListen func(listenAddr net.Addr)
}

// New 新建一个本地端实例。
// 本地端的主要职责如下：
// 0. 监听来自本地浏览器的代理请求。
// 1. 转发前对数据进行加密。
// 2. 将 socket 数据转发到服务端。
// 3. 把服务端返回的数据转发给用户的浏览器。
// 参数 encodePassword 是用于加密的密码。
// 参数 localAddr 是本地监听地址。
// 参数 serverAddr 是远程服务端地址。
// 返回一个指向 LsLocal 实例的指针。
func New(encodePassword *core.Password, localAddr, serverAddr *net.TCPAddr) *LsLocal {
	log.Printf("DEBUG: 创建新的本地代理实例，本地地址: %s, 服务端地址: %s", localAddr.String(), serverAddr.String())
	return &LsLocal{
		SecureSocket: &core.SecureSocket{
			Cipher:     core.NewSimpleCipher(encodePassword), // 创建编解码器实例
			LocalAddr:  localAddr,
			ServerAddr: serverAddr,
		},
	}
}

// Listen 本地端启动监听，等待本地浏览器的代理请求。
// 返回可能出现的错误。
func (local *LsLocal) Listen() error {
	log.Printf("DEBUG: 尝试监听本地地址 %s", local.LocalAddr.String())
	// 开始监听指定的本地 TCP 地址
	listener, err := net.ListenTCP("tcp", local.LocalAddr)
	if err != nil {
		log.Printf("DEBUG: 监听本地地址 %s 失败: %v", local.LocalAddr.String(), err)
		return err
	}
	log.Printf("DEBUG: 成功监听本地地址 %s", local.LocalAddr.String())

	// 函数结束时关闭监听器，注意此处未处理关闭时可能出现的错误
	defer listener.Close()
	// 标记本地代理服务正在运行
	local.running = true

	// 若 AfterListen 回调函数存在，则调用该函数并传入监听地址
	if local.AfterListen != nil {
		local.AfterListen(listener.Addr())
	}

	// 持续监听，直到服务停止
	for local.running {
		log.Println("DEBUG: 等待新的连接...")
		// 接受新的 TCP 连接
		userConn, err := listener.AcceptTCP()
		if err != nil {
			// 若接受连接出错，跳过本次循环继续监听
			log.Printf("DEBUG: 接受新连接时出错: %v", err)
			continue
		}
		log.Printf("DEBUG: 接受新连接来自 %s", userConn.RemoteAddr().String())
		// 设置 userConn 关闭时直接清除所有数据，不等待未发送的数据
		userConn.SetLinger(0)
		// 启动一个新的 goroutine 处理该连接
		go local.handleConn(userConn)
	}
	return nil
}

// Close 停止运行当前本地代理服务，并释放对应资源。
// 目前仅停止服务运行，释放资源部分待实现。
func (local *LsLocal) Close() {
	log.Println("DEBUG: 尝试关闭本地代理服务")
	// TODO 释放所有资源
	// 标记本地代理服务停止运行
	local.running = false
	// 置空 SecureSocket 指针
	local.SecureSocket = nil
	log.Println("DEBUG: 本地代理服务已关闭")
}

// handleConn 处理与用户浏览器建立的 TCP 连接。
// 参数 userConn 是与用户浏览器建立的 TCP 连接。
func (local *LsLocal) handleConn(userConn *net.TCPConn) {
	log.Printf("DEBUG: 开始处理来自 %s 的连接", userConn.RemoteAddr().String())
	// 函数结束时关闭与用户浏览器的连接，处理关闭时可能出现的错误
	defer func() {
		if err := userConn.Close(); err != nil {
			log.Printf("ERROR: 关闭用户浏览器连接出错: %v", err)
		}
	}()

	// 与远程服务端建立 TCP 连接
	server, err := local.DialServer()
	if err != nil {
		// 若连接失败，记录错误日志并返回
		log.Printf("ERROR: 连接远程服务端出错: %v", err)
		return
	}

	// 函数结束时关闭与远程服务端的连接，处理关闭时可能出现的错误
	defer func() {
		if err := server.Close(); err != nil {
			log.Printf("ERROR: 关闭远程服务端连接出错: %v", err)
		}
	}()

	// 设置 server 关闭时直接清除所有数据，不等待未发送的数据
	server.SetLinger(0)
	// 设置 server 连接的截止时间
	server.SetDeadline(time.Now().Add(core.TIMEOUT))

	// 启动一个新的 goroutine 对数据进行加密并从用户连接转发到服务端连接
	go local.EncodeCopy(server, userConn)
	// 对数据进行解密并从服务端连接转发到用户连接
	local.DecodeCopy(userConn, server)
	log.Printf("DEBUG: 完成从 %s 到 %s 的解密转发", local.ServerAddr.String(), userConn.RemoteAddr().String())
}

// ... 已有代码 ...
