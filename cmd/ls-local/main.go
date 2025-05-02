package main

import (
	"net"

	"github.com/beijian128/minisocks/cmd"
	"github.com/beijian128/minisocks/local"
	"github.com/sirupsen/logrus"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	logger  = logrus.WithField("component", "minisocks-client")
)

func init() {
	// 配置日志格式
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

func main() {
	// 打印版本信息
	logger.WithFields(logrus.Fields{
		"version": version,
		"commit":  commit,
		"date":    date,
	}).Info("启动 minisocks 客户端")

	// 加载配置
	config, err := cmd.LoadConfig()
	if err != nil {
		logger.WithError(err).Fatal("加载配置失败")
	}

	// 解析本地监听地址
	localAddr, err := net.ResolveTCPAddr("tcp", config.ListenAddr)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"listenAddr": config.ListenAddr,
			"error":      err,
		}).Fatal("解析本地监听地址失败")
	}

	// 解析远程服务地址
	serverAddr, err := net.ResolveTCPAddr("tcp", config.RemoteAddr)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"remoteAddr": config.RemoteAddr,
			"error":      err,
		}).Fatal("解析远程服务地址失败")
	}

	// 创建本地代理实例
	lsLocal := local.New(config.Password, localAddr, serverAddr)
	lsLocal.AfterListen = func(listenAddr net.Addr) {
		logger.WithFields(logrus.Fields{
			"listenAddr": listenAddr.String(),
			"remoteAddr": config.RemoteAddr,
		}).Info("客户端启动成功")
	}

	// 启动本地代理
	if err := lsLocal.Listen(); err != nil {
		logger.WithError(err).Fatal("客户端运行失败")
	}
}
