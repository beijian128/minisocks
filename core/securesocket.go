// core 包包含 minisocks 的核心功能实现
package core

import (
	"errors" // 提供错误处理功能
	"fmt"    // 提供格式化输入输出功能
	"io"     // 提供基础的 I/O 接口
	"net"    // 提供网络相关功能
	"time"   // 提供时间相关功能
)

// BufSize 定义读写操作时缓冲区的大小
const BufSize = 1024

// TIMEOUT 定义网络操作的超时时间
const TIMEOUT = 10 * time.Second

// SecureSocket 结构体表示一个安全的网络套接字，用于加密传输数据
type SecureSocket struct {
	Cipher     *Cipher      // 编解码器实例，用于数据的加密和解密
	LocalAddr  *net.TCPAddr // 本地 TCP 地址
	ServerAddr *net.TCPAddr // 远程服务器 TCP 地址
}

// DecodeRead 从指定的 TCP 连接中读取加密数据，解密后将原始数据存入指定的字节切片中。
// 参数 conn 是要读取数据的 TCP 连接。
// 参数 bs 是用于存储解密后数据的字节切片。
// 返回读取的字节数和可能出现的错误。
func (secureSocket *SecureSocket) DecodeRead(conn *net.TCPConn, bs []byte) (n int, err error) {
	// 从连接中读取数据
	n, err = conn.Read(bs)
	if err != nil {
		return
	}
	// 对读取到的数据进行解密
	secureSocket.Cipher.decode(bs[:n])
	return
}

// EncodeWrite 对指定字节切片中的数据进行加密，并将加密后的数据全部写入指定的 TCP 连接。
// 参数 conn 是要写入数据的 TCP 连接。
// 参数 bs 是要加密并写入的数据。
// 返回写入的字节数和可能出现的错误。
func (secureSocket *SecureSocket) EncodeWrite(conn *net.TCPConn, bs []byte) (int, error) {
	// 对数据进行加密
	secureSocket.Cipher.encode(bs)
	// 将加密后的数据写入连接
	return conn.Write(bs)
}

// EncodeCopy 从源 TCP 连接中持续读取原始数据，加密后写入目标 TCP 连接，直到源连接没有更多数据可读。
// 参数 dst 是目标 TCP 连接。
// 参数 src 是源 TCP 连接。
// 返回可能出现的错误。
func (secureSocket *SecureSocket) EncodeCopy(dst *net.TCPConn, src *net.TCPConn) error {
	// 创建缓冲区
	buf := make([]byte, BufSize)
	for {
		// 从源连接读取数据
		nr, er := src.Read(buf)
		if nr > 0 {
			// 对读取的数据进行加密并写入目标连接
			nw, ew := secureSocket.EncodeWrite(dst, buf[0:nr])
			if ew != nil {
				return ew
			}
			// 检查写入的字节数是否与读取的字节数一致
			if nr != nw {
				return io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return er
			} else {
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
	// 创建缓冲区
	buf := make([]byte, BufSize)
	for {
		// 从源连接读取加密数据并解密
		nr, er := secureSocket.DecodeRead(src, buf)
		if nr > 0 {
			// 将解密后的数据写入目标连接
			nw, ew := dst.Write(buf[0:nr])
			if ew != nil {
				return ew
			}
			// 检查写入的字节数是否与读取的字节数一致
			if nr != nw {
				return io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return er
			} else {
				return nil
			}
		}
	}
}

// DialServer 与远程服务器建立 TCP 连接，该连接上的数据将进行加密传输。
// 返回建立好的 TCP 连接和可能出现的错误。
func (secureSocket *SecureSocket) DialServer() (*net.TCPConn, error) {
	// 尝试与远程服务器建立 TCP 连接
	remoteConn, err := net.DialTCP("tcp", nil, secureSocket.ServerAddr)
	if err != nil {
		// 若连接失败，返回错误信息
		return nil, errors.New(fmt.Sprintf("dail remote %s fail:%s", secureSocket.ServerAddr, err))
	}
	return remoteConn, nil
}
