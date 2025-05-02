package main

import (
	"net"

	"github.com/beijian128/minisocks/cmd"
	"github.com/beijian128/minisocks/server"
	"github.com/sirupsen/logrus"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	logger  = logrus.WithField("component", "minisocks-server")
)

func init() {
	// 设置日志格式
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
	}).Info("启动 minisocks 服务端")

	// 加载配置
	config, err := cmd.LoadConfig()
	if err != nil {
		logger.WithError(err).Fatal("加载配置失败")
	}

	// 解析监听地址
	localAddr, err := net.ResolveTCPAddr("tcp", config.ListenAddr)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"listenAddr": config.ListenAddr,
			"error":      err,
		}).Fatal("解析监听地址失败")
	}

	// 创建服务器实例
	lsServer := server.New(config.Password, localAddr)
	lsServer.AfterListen = func(listenAddr net.Addr) {
		logger.WithFields(logrus.Fields{
			"listenAddr": listenAddr.String(),
			"password":   config.Password,
		}).Info("服务启动成功")
	}

	// 启动服务器
	if err := lsServer.Listen(); err != nil {
		logger.WithError(err).Fatal("服务运行失败")
	}
}
