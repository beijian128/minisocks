package core

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

// BufSize 定义读写操作时缓冲区的大小
const BufSize = 1024

// TIMEOUT 定义网络操作的超时时间
const TIMEOUT = 30 * time.Second

// SecureSocket 结构体表示一个安全的网络套接字，用于加密传输数据
type SecureSocket struct {
	Cipher     Cipher       // 编解码器实例，用于数据的加密和解密
	LocalAddr  *net.TCPAddr // 本地 TCP 地址
	ServerAddr *net.TCPAddr // 远程服务器 TCP 地址
	logger     *logrus.Entry
}

// NewSecureSocket 创建新的 SecureSocket 实例
func NewSecureSocket(cipher Cipher, localAddr, serverAddr *net.TCPAddr) *SecureSocket {
	return &SecureSocket{
		Cipher:     cipher,
		LocalAddr:  localAddr,
		ServerAddr: serverAddr,
		logger: logrus.WithFields(logrus.Fields{
			"component": "SecureSocket",
			"local":     localAddr,
			"remote":    serverAddr,
		}),
	}
}

// EncodeCopy 从源 TCP 连接中持续读取原始数据，加密后写入目标 TCP 连接
func (s *SecureSocket) EncodeCopy(dst *net.TCPConn, src *net.TCPConn) error {
	s.logger.WithFields(logrus.Fields{
		"src": src.RemoteAddr(),
		"dst": dst.RemoteAddr(),
	}).Debug("开始加密传输数据")

	buf := make([]byte, BufSize)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			s.logger.WithField("bytes", nr).Debug("读取原始数据")

			data, err := s.Cipher.Encrypt(buf[:nr])
			if err != nil {
				s.logger.WithError(err).Error("加密数据失败")
				return fmt.Errorf("加密失败: %w", err)
			}

			if _, ew := dst.Write(data); ew != nil {
				s.logger.WithError(ew).Error("写入加密数据失败")
				return fmt.Errorf("写入失败: %w", ew)
			}
		}

		if er != nil {
			if er != io.EOF {
				s.logger.WithError(er).Error("读取原始数据时出错")
				return fmt.Errorf("读取失败: %w", er)
			}
			s.logger.Debug("读取原始数据结束 (EOF)")
			return nil
		}
	}
}

// DecodeCopy 从源 TCP 连接中持续读取加密数据，解密后写入目标 TCP 连接
func (s *SecureSocket) DecodeCopy(dst *net.TCPConn, src *net.TCPConn) error {
	s.logger.WithFields(logrus.Fields{
		"src": src.RemoteAddr(),
		"dst": dst.RemoteAddr(),
	}).Debug("开始解密传输数据")

	buf := make([]byte, BufSize)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			data, err := s.Cipher.Decrypt(buf[:nr])
			if err != nil {
				s.logger.WithError(err).Error("解密数据失败")
				return fmt.Errorf("解密失败: %w", err)
			}

			if _, ew := dst.Write(data); ew != nil {
				s.logger.WithError(ew).Error("写入解密数据失败")
				return fmt.Errorf("写入失败: %w", ew)
			}
		}

		if er != nil {
			if er != io.EOF {
				s.logger.WithError(er).Error("读取加密数据时出错")
				return fmt.Errorf("读取失败: %w", er)
			}
			s.logger.Debug("读取加密数据结束 (EOF)")
			return nil
		}
	}
}

// DialServer 与远程服务器建立 TCP 连接
func (s *SecureSocket) DialServer() (*net.TCPConn, error) {
	s.logger.Info("尝试连接远程服务器")

	remoteConn, err := net.DialTCP("tcp", nil, s.ServerAddr)
	if err != nil {
		s.logger.WithError(err).Error("连接远程服务器失败")
		return nil, fmt.Errorf("连接 %s 失败: %w", s.ServerAddr, err)
	}

	s.logger.Info("成功连接到远程服务器")
	return remoteConn, nil
}
