package local

import (
	"fmt"
	"net"
	"time"

	"github.com/beijian128/minisocks/core"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// LsLocal 表示本地代理服务端，负责处理本地浏览器的代理请求
type LsLocal struct {
	*core.SecureSocket      // 嵌入 SecureSocket 结构体，用于数据的加密和解密传输
	running            bool // 标识本地代理服务是否正在运行
	logger             *logrus.Entry
	// AfterListen 是一个回调函数，在本地代理开始监听后被调用，传入监听地址
	AfterListen func(listenAddr net.Addr)
}

// New 新建一个本地端实例
func New(secret string, localAddr, serverAddr *net.TCPAddr) *LsLocal {
	logger := logrus.WithFields(logrus.Fields{
		"component":  "LsLocal",
		"localAddr":  localAddr.String(),
		"serverAddr": serverAddr.String(),
	})
	logger.Debug("创建新的本地代理实例")

	ci, _ := core.NewSimple(secret)
	return &LsLocal{
		SecureSocket: core.NewSecureSocket(ci, localAddr, serverAddr),
		logger:       logger,
	}
}

// Listen 本地端启动监听，等待本地浏览器的代理请求
func (l *LsLocal) Listen() error {
	l.logger.Info("开始监听本地地址")

	listener, err := net.ListenTCP("tcp", l.LocalAddr)
	if err != nil {
		l.logger.WithError(err).Error("监听失败")
		return fmt.Errorf("监听失败: %w", err)
	}
	defer listener.Close()

	l.logger.WithField("address", listener.Addr()).Info("监听成功")
	l.running = true

	if l.AfterListen != nil {
		l.AfterListen(listener.Addr())
	}

	for l.running {
		l.logger.Debug("等待新连接")
		userConn, err := listener.AcceptTCP()
		if err != nil {
			l.logger.WithError(err).Warn("接受连接失败")
			continue
		}

		l.logger.WithField("remoteAddr", userConn.RemoteAddr()).Debug("接受新连接")
		userConn.SetLinger(0)
		go l.handleConn(userConn)
	}

	return nil
}

// Close 停止运行当前本地代理服务
func (l *LsLocal) Close() {
	l.logger.Info("关闭本地代理服务")
	l.running = false
	l.SecureSocket = nil
}

// handleConn 处理与用户浏览器建立的 TCP 连接
func (l *LsLocal) handleConn(userConn *net.TCPConn) {
	connID := uuid.New().String()
	logger := l.logger.WithFields(logrus.Fields{
		"connID":     connID,
		"remoteAddr": userConn.RemoteAddr(),
	})
	logger.Debug("开始处理连接")

	defer func() {
		if err := userConn.Close(); err != nil {
			logger.WithError(err).Warn("关闭用户连接失败")
		}
		logger.Debug("连接处理完成")
	}()

	// 连接远程服务端
	logger.Debug("连接远程服务端")
	server, err := l.DialServer()
	if err != nil {
		logger.WithError(err).Error("连接服务端失败")
		return
	}
	defer func() {
		if err := server.Close(); err != nil {
			logger.WithError(err).Warn("关闭服务端连接失败")
		}
	}()

	server.SetLinger(0)
	if err := server.SetDeadline(time.Now().Add(core.TIMEOUT)); err != nil {
		logger.WithError(err).Warn("设置截止时间失败")
	}

	// 启动数据转发
	l.startForwarding(logger, userConn, server)
}

func (l *LsLocal) startForwarding(logger *logrus.Entry, userConn, server *net.TCPConn) {
	logger.WithFields(logrus.Fields{
		"userAddr":   userConn.RemoteAddr(),
		"serverAddr": server.RemoteAddr(),
	}).Debug("开始数据转发")

	// 启动加密转发协程
	go func() {
		if err := l.EncodeCopy(server, userConn); err != nil {
			logger.WithError(err).Debug("加密转发结束")
		}
	}()

	// 执行解密转发
	if err := l.DecodeCopy(userConn, server); err != nil {
		logger.WithError(err).Debug("解密转发结束")
	}

	logger.Debug("数据转发完成")
}
