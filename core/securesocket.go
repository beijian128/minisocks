package core

import (
	"net"
	"errors"
	"fmt"
	"io"
	"time"
)

const (
	BufSize = 1024
	TIMEOUT = 10 * time.Second
)

type SecureSocket struct {
	Cipher     *Cipher
	LocalAddr  *net.TCPAddr
	ServerAddr *net.TCPAddr
}

//从输入流里读取加密过的数据，解密后把原数据放到bs里
func (secureSocket *SecureSocket) DecodeRead(conn *net.TCPConn, bs []byte) (n int, err error) {
	n, err = conn.Read(bs)
	if err != nil {
		return
	}
	secureSocket.Cipher.decode(bs[:n])
	return
}

//把放在bs里的数据加密后立即全部写入输出流
func (secureSocket *SecureSocket) EncodeWrite(conn *net.TCPConn, bs []byte) (int, error) {
	secureSocket.Cipher.encode(bs)
	return conn.Write(bs)
}

//从src中源源不断的读取原数据加密后写入到dst，直到src中没有数据可以再读取
func (secureSocket *SecureSocket) EncodeCopy(dst *net.TCPConn, src *net.TCPConn) error {
	buf := make([]byte, BufSize)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := secureSocket.EncodeWrite(dst, buf[0:nr])
			if ew != nil {
				return ew
			}
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

//从src中源源不断的读取加密后的数据解密后写入到dst，直到src中没有数据可以再读取
func (secureSocket *SecureSocket) DecodeCopy(dst *net.TCPConn, src *net.TCPConn) error {
	buf := make([]byte, BufSize)
	for {
		nr, er := secureSocket.DecodeRead(src, buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if ew != nil {
				return ew
			}
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

//和远程的socket建立连接，他们直接的数据会加密传输
func (secureSocket *SecureSocket) DialServer() (*net.TCPConn, error) {
	remoteConn, err := net.DialTCP("tcp", nil, secureSocket.ServerAddr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("dail remote %s fail:%s", secureSocket.ServerAddr, err))
	}
	return remoteConn, nil
}
