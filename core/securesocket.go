// core 包包含 minisocks 的核心功能实现
package core

import (
	"errors" // 提供错误处理功能
	"fmt"    // 提供格式化输入输出功能
	"io"     // 提供基础的 I/O 接口
	"log"    // 新增日志包
	"net"    // 提供网络相关功能
	"time"
)

// BufSize 定义读写操作时缓冲区的大小
const BufSize = 1024

// TIMEOUT 定义网络操作的超时时间
const TIMEOUT = 10 * time.Second

// SecureSocket 结构体表示一个安全的网络套接字，用于加密传输数据
type SecureSocket struct {
	Cipher     Cipher       // 编解码器实例，用于数据的加密和解密
	LocalAddr  *net.TCPAddr // 本地 TCP 地址
	ServerAddr *net.TCPAddr // 远程服务器 TCP 地址
}

// EncodeCopy 从源 TCP 连接中持续读取原始数据，加密后写入目标 TCP 连接，直到源连接没有更多数据可读。
// 参数 dst 是目标 TCP 连接。
// 参数 src 是源 TCP 连接。
// 返回可能出现的错误。
func (secureSocket *SecureSocket) EncodeCopy(dst *net.TCPConn, src *net.TCPConn) error {
	log.Printf("DEBUG: 开始从 %s 读取原始数据并加密写入 %s", src.RemoteAddr(), dst.RemoteAddr())
	// 创建缓冲区
	buf := make([]byte, BufSize)
	for {
		// 从源连接读取数据
		nr, er := src.Read(buf)
		if nr > 0 {
			log.Printf("DEBUG: 从 %s 读取到 %d 字节原始数据", src.RemoteAddr(), nr)
			// 对读取的数据进行加密并写入目标连接
			data, err := secureSocket.Cipher.Encrypt(buf[:nr])
			if err != nil {
				log.Printf("ERROR: <UNK>: %v", err)
				return err
			}
			_, ew := dst.Write(data)
			if ew != nil {
				log.Printf("DEBUG: 向 %s 写入加密数据时出错: %v", dst.RemoteAddr(), ew)
				return ew
			}

			log.Printf("DEBUG: 已成功将 %d 字节原始数据加密写入 %s", nr, dst.RemoteAddr())
		}
		if er != nil {
			if er != io.EOF {
				log.Printf("DEBUG: 从 %s 读取数据时出错: %v", src.RemoteAddr(), er)
				return er
			} else {
				log.Printf("DEBUG: 从 %s 读取数据结束（EOF）", src.RemoteAddr())
				return nil
			}
		}
	}
}

// DecodeCopy 从源 TCP 连接中持续读取加密数据，解密后写入目标 TCP 连接，直到源连接没有更多数据可读。
// 参数 dst 是目标 TCP 连接。
// 参数 src 是源 TCP 连接。
// 返回可能出现的错误。
func (secureSocket *SecureSocket) DecodeCopy(dst *net.TCPConn, src *net.TCPConn) error {
	log.Printf("DEBUG: 开始从 %s 读取加密数据并解密写入 %s", src.RemoteAddr(), dst.RemoteAddr())
	// 创建缓冲区
	buf := make([]byte, BufSize)
	for {
		// 从源连接读取加密数据并解密

		nr, er := src.Read(buf)
		if nr > 0 {
			log.Printf("DEBUG: 从 %s 读取并解密 %d 字节数据", src.RemoteAddr(), nr)
			data, err := secureSocket.Cipher.Decrypt(buf[0:nr])
			if err != nil {
				log.Printf("DEBUG: <UNK> %s <UNK>: %v", dst.RemoteAddr(), err)
				return err
			}
			// 将解密后的数据写入目标连接
			_, ew := dst.Write(data)
			if ew != nil {
				log.Printf("DEBUG: 向 %s 写入解密数据时出错: %v", dst.RemoteAddr(), ew)
				return ew
			}
			log.Printf("DEBUG: 已成功将 %d 字节解密数据写入 %s", nr, dst.RemoteAddr())
		}
		if er != nil {
			if er != io.EOF {
				log.Printf("DEBUG: 从 %s 读取加密数据时出错: %v", src.RemoteAddr(), er)
				return er
			} else {
				log.Printf("DEBUG: 从 %s 读取加密数据结束（EOF）", src.RemoteAddr())
				return nil
			}
		}
	}
}

// DialServer 与远程服务器建立 TCP 连接，该连接上的数据将进行加密传输。
// 返回建立好的 TCP 连接和可能出现的错误。
func (secureSocket *SecureSocket) DialServer() (*net.TCPConn, error) {
	log.Printf("DEBUG: 尝试连接远程服务器 %s", secureSocket.ServerAddr)
	// 尝试与远程服务器建立 TCP 连接
	remoteConn, err := net.DialTCP("tcp", nil, secureSocket.ServerAddr)
	if err != nil {
		// 若连接失败，返回错误信息
		errMsg := fmt.Sprintf("dail remote %s fail:%s", secureSocket.ServerAddr, err)
		log.Printf("DEBUG: 连接远程服务器 %s 失败: %v", secureSocket.ServerAddr, err)
		return nil, errors.New(errMsg)
	}
	log.Printf("DEBUG: 成功连接到远程服务器 %s", secureSocket.ServerAddr)
	return remoteConn, nil
}
