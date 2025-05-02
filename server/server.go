package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/beijian128/minisocks/core"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// LsServer 表示 minisocks 服务端，负责处理来自本地端的请求
type LsServer struct {
	*core.SecureSocket      // 嵌入 SecureSocket 结构体，用于数据的加密和解密
	running            bool // 标识服务端是否正在运行
	logger             *logrus.Entry
	// AfterListen 是一个回调函数，在服务端开始监听后被调用，传入监听地址
	AfterListen func(listenAddr net.Addr)
}

// New 新建一个服务端实例
func New(secret string, localAddr *net.TCPAddr) *LsServer {
	logger := logrus.WithFields(logrus.Fields{
		"component":  "LsServer",
		"listenAddr": localAddr.String(),
	})
	logger.Debug("创建新的服务端实例")

	ci, _ := core.NewSimple(secret)
	return &LsServer{
		SecureSocket: core.NewSecureSocket(ci, localAddr, nil),
		logger:       logger,
	}
}

// Listen 启动服务端并监听来自本地端的请求
func (s *LsServer) Listen() error {
	s.logger.Info("开始监听")

	listener, err := net.ListenTCP("tcp", s.LocalAddr)
	if err != nil {
		s.logger.WithError(err).Error("监听失败")
		return fmt.Errorf("监听失败: %w", err)
	}
	defer listener.Close()

	s.logger.WithField("address", listener.Addr()).Info("监听成功")
	s.running = true

	if s.AfterListen != nil {
		s.AfterListen(listener.Addr())
	}

	for s.running {
		s.logger.Debug("等待新连接")
		localConn, err := listener.AcceptTCP()
		if err != nil {
			s.logger.WithError(err).Error("接受连接失败")
			continue
		}

		s.logger.WithField("remoteAddr", localConn.RemoteAddr()).Debug("接受新连接")
		localConn.SetLinger(0)
		go s.handleConn(localConn)
	}

	return nil
}

// Close 停止运行当前服务端并释放对应资源
func (s *LsServer) Close() {
	s.logger.Info("关闭服务端")
	s.running = false
	s.SecureSocket = nil
}

// handleConn 处理来自本地端的连接，实现 socks5 协议
func (s *LsServer) handleConn(localConn *net.TCPConn) {
	connID := uuid.New().String()
	logger := s.logger.WithFields(logrus.Fields{
		"connID":     connID,
		"remoteAddr": localConn.RemoteAddr(),
	})
	logger.Debug("开始处理连接")
	defer localConn.Close()

	buf := make([]byte, 256)

	// 处理 SOCKS5 握手
	if err := s.handleHandshake(logger, localConn, buf); err != nil {
		logger.WithError(err).Error("握手失败")
		return
	}

	// 处理 SOCKS5 请求
	dstServer, err := s.handleRequest(logger, localConn, buf)
	if err != nil {
		logger.WithError(err).Error("请求处理失败")
		return
	}
	defer dstServer.Close()

	// 开始转发数据
	s.startForwarding(logger, localConn, dstServer)
}

func (s *LsServer) handleHandshake(logger *logrus.Entry, conn *net.TCPConn, buf []byte) error {
	logger.Debug("开始握手")

	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("读取握手数据失败: %w", err)
	}

	data, err := s.Cipher.Decrypt(buf[:n])
	if err != nil || data[0] != 0x05 {
		if err != nil {
			return fmt.Errorf("解密握手数据失败: %w", err)
		}
		return errors.New("不支持的协议版本，仅支持 Socks5")
	}

	if data[1] != 0x01 {
		return fmt.Errorf("不支持的请求类型: 0x%x，仅支持 CONNECT(0x01)", data[1])
	}

	// 发送验证通过响应
	response, _ := s.Cipher.Encrypt([]byte{0x05, 0x00})
	if _, err := conn.Write(response); err != nil {
		return fmt.Errorf("发送验证响应失败: %w", err)
	}

	logger.Debug("握手成功")
	return nil
}

func (s *LsServer) handleRequest(logger *logrus.Entry, conn *net.TCPConn, buf []byte) (*net.TCPConn, error) {
	logger.Debug("处理请求")

	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("读取请求数据失败: %w", err)
	}

	data, err := s.Cipher.Decrypt(buf[:n])
	if err != nil || len(data) < 7 {
		if err != nil {
			return nil, fmt.Errorf("解密请求数据失败: %w", err)
		}
		return nil, fmt.Errorf("请求数据长度不足，期望至少 7 字节，实际 %d 字节", len(data))
	}

	var dIP []byte
	switch data[3] {
	case 0x01:
		dIP = data[4 : 4+net.IPv4len]
	case 0x03:
		domain := string(data[5 : len(data)-2])
		ipAddr, err := net.ResolveIPAddr("ip", domain)
		if err != nil {
			return nil, fmt.Errorf("解析域名 %s 失败: %w", domain, err)
		}
		dIP = ipAddr.IP
	case 0x04:
		dIP = data[4 : 4+net.IPv6len]
	default:
		return nil, fmt.Errorf("不支持的目标地址类型: 0x%x", data[3])
	}

	dPort := data[len(data)-2:]
	dstAddr := &net.TCPAddr{
		IP:   dIP,
		Port: int(binary.BigEndian.Uint16(dPort)),
	}

	logger.WithField("targetAddr", dstAddr.String()).Debug("连接目标服务器")
	dstServer, err := net.DialTCP("tcp", nil, dstAddr)
	if err != nil {
		return nil, fmt.Errorf("连接目标服务器失败: %w", err)
	}

	// 发送成功响应
	successResp, _ := s.Cipher.Encrypt([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if _, err := conn.Write(successResp); err != nil {
		dstServer.Close()
		return nil, fmt.Errorf("发送成功响应失败: %w", err)
	}

	if err := dstServer.SetLinger(0); err != nil {
		logger.WithError(err).Warn("设置 Linger 失败")
	}
	if err := dstServer.SetDeadline(time.Now().Add(core.TIMEOUT)); err != nil {
		logger.WithError(err).Warn("设置 Deadline 失败")
	}

	logger.Debug("请求处理成功")
	return dstServer, nil
}

func (s *LsServer) startForwarding(logger *logrus.Entry, localConn, dstServer *net.TCPConn) {
	logger.WithFields(logrus.Fields{
		"localAddr":  localConn.RemoteAddr(),
		"targetAddr": dstServer.RemoteAddr(),
	}).Debug("开始数据转发")

	// 启动解密转发协程
	go func() {
		if err := s.DecodeCopy(dstServer, localConn); err != nil {
			logger.WithError(err).Debug("解密转发结束")
		}
	}()

	// 执行加密转发
	if err := s.EncodeCopy(localConn, dstServer); err != nil {
		logger.WithError(err).Debug("加密转发结束")
	}

	logger.Debug("数据转发完成")
}
