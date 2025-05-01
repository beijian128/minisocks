package server

import (
	"encoding/binary" // 提供二进制数据编码解码功能
	"net"             // 提供网络相关功能
	"time"            // 提供时间相关功能

	"github.com/beijian128/minisocks/core" // 引入 minisocks 核心功能包
)

// LsServer 表示 minisocks 服务端，负责处理来自本地端的请求
type LsServer struct {
	*core.SecureSocket      // 嵌入 SecureSocket 结构体，用于数据的加密和解密
	running            bool // 标识服务端是否正在运行
	// AfterListen 是一个回调函数，在服务端开始监听后被调用，传入监听地址
	AfterListen func(listenAddr net.Addr)
}

// New 新建一个服务端实例。
// 服务端的职责是:
// 0. 监听来自本地端的请求
// 1. 解密本地端请求的数据，解析 socks5 协议，连接用户浏览器真正想要连接的远程服务器
// 2. 加密后转发用户浏览器真正想要连接的远程服务器返回的数据到本地端
// 参数 encodePassword 是用于加密的密码
// 参数 localAddr 是服务端监听的本地地址
// 返回一个指向 LsServer 实例的指针
func New(encodePassword *core.Password, localAddr *net.TCPAddr) *LsServer {
	return &LsServer{
		SecureSocket: &core.SecureSocket{
			Cipher:    core.NewCipher(encodePassword), // 创建编解码器实例
			LocalAddr: localAddr,
		},
	}
}

// Listen 启动服务端并监听来自本地端的请求
// 返回可能出现的错误
func (server *LsServer) Listen() error {
	// 开始监听指定的本地 TCP 地址
	listener, err := net.ListenTCP("tcp", server.LocalAddr)
	if err != nil {
		return err
	}

	// 函数结束时关闭监听器，注意此处未处理关闭时可能出现的错误
	defer listener.Close()
	// 标记服务端正在运行
	server.running = true

	// 若 AfterListen 回调函数存在，则调用该函数并传入监听地址
	if server.AfterListen != nil {
		server.AfterListen(listener.Addr())
	}

	// 持续监听，直到服务停止
	for server.running {
		// 接受新的 TCP 连接
		localConn, err := listener.AcceptTCP()
		if err != nil {
			// 若接受连接出错，跳过本次循环继续监听
			continue
		}
		// 设置 localConn 关闭时直接清除所有数据，不等待未发送的数据
		localConn.SetLinger(0)
		// 启动一个新的 goroutine 处理该连接
		go server.handleConn(localConn)
	}
	return nil
}

// Close 停止运行当前服务端并释放对应资源
// 目前仅停止服务运行，释放资源部分待实现
func (server *LsServer) Close() {
	// TODO 释放所有资源
	// 标记服务端停止运行
	server.running = false
	// 置空 SecureSocket 指针
	server.SecureSocket = nil
}

// handleConn 处理来自本地端的连接，实现 socks5 协议
// 参考文档：
// https://www.ietf.org/rfc/rfc1928.txt
// http://www.jianshu.com/p/172810a70fad
// 参数 localConn 是与本地端建立的 TCP 连接
func (server *LsServer) handleConn(localConn *net.TCPConn) {
	// 函数结束时关闭与本地端的连接，注意此处未处理关闭时可能出现的错误
	defer localConn.Close()
	// 创建缓冲区用于接收数据
	buf := make([]byte, 256)

	/**
	The localConn connects to the dstServer, and sends a ver
	identifier/method selection message:
			+----+----------+----------+
			|VER | NMETHODS | METHODS  |
			+----+----------+----------+
			| 1  |    1     | 1 to 255 |
			+----+----------+----------+
	The VER field is set to X'05' for this ver of the protocol.  The
	NMETHODS field contains the number of method identifier octets that
	appear in the METHODS field.
	*/
	// 读取本地端发送的版本和方法选择消息并解密
	_, err := server.DecodeRead(localConn, buf)
	// 只支持 Socks5 协议，若读取出错或版本号不为 0x05 则返回
	if err != nil || buf[0] != 0x05 {
		return
	}

	/**
	The dstServer selects from one of the methods given in METHODS, and
	sends a METHOD selection message:

					+----+--------+
					|VER | METHOD |
					+----+--------+
					| 1  |   1    |
					+----+--------+
	*/
	// 不需要验证，直接发送验证通过消息并加密写入本地连接
	server.EncodeWrite(localConn, []byte{0x05, 0x00})

	/**
	+----+-----+-------+------+----------+----------+
	|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	+----+-----+-------+------+----------+----------+
	| 1  |  1  | X'00' |  1   | Variable |    2     |
	+----+-----+-------+------+----------+----------+
	*/

	// CMD 代表客户端请求的类型，值长度为 1 个字节，有三种类型
	// CONNECT X'01'
	// 目前只支持 CONNECT 类型，若不是则返回
	if buf[1] != 0x01 {
		return
	}

	// 读取客户端请求的目标地址和端口信息并解密
	n, err := server.DecodeRead(localConn, buf)
	// n 最短的长度为 7，情况为 ATYP=3 且 DST.ADDR 占用 1 字节，值为 0x0
	// 若读取出错或长度不足则返回
	if err != nil || n < 7 {
		return
	}
	var dIP []byte
	// aType 代表请求的远程服务器地址类型，值长度 1 个字节，有三种类型
	switch buf[3] {
	case 0x01:
		//	IP V4 address: X'01'
		dIP = buf[4 : 4+net.IPv4len]
	case 0x03:
		//	DOMAINNAME: X'03'
		// 解析域名对应的 IP 地址
		ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:n-2]))
		if err != nil {
			return
		}
		dIP = ipAddr.IP
	case 0x04:
		//	IP V6 address: X'04'
		dIP = buf[4 : 4+net.IPv6len]
	default:
		return
	}
	// 提取目标端口信息
	dPort := buf[n-2:]
	// 构建目标服务器的 TCP 地址
	dstAddr := &net.TCPAddr{
		IP:   dIP,
		Port: int(binary.BigEndian.Uint16(dPort)),
	}
	// 连接目标服务器
	dstServer, err := net.DialTCP("tcp", nil, dstAddr)

	/**
	 +----+-----+-------+------+----------+----------+
	|VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	+----+-----+-------+------+----------+----------+
	| 1  |  1  | X'00' |  1   | Variable |    2     |
	+----+-----+-------+------+----------+----------+
	*/
	if err != nil {
		// 若连接目标服务器失败则返回
		return
	} else {
		// 函数结束时关闭与目标服务器的连接，注意此处未处理关闭时可能出现的错误
		defer dstServer.Close()
		// 发送响应消息给客户端表示连接成功并加密写入本地连接
		server.EncodeWrite(localConn, []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		// 设置 dstServer 关闭时直接清除所有数据，不等待未发送的数据
		dstServer.SetLinger(0)
		// 设置 dstServer 连接的截止时间
		dstServer.SetDeadline(time.Now().Add(core.TIMEOUT))
	}
	// 启动一个新的 goroutine 对数据进行解密并从本地连接转发到目标服务器连接
	go server.DecodeCopy(dstServer, localConn)
	// 对数据进行加密并从目标服务器连接转发到本地连接
	server.EncodeCopy(localConn, dstServer)
}
