package main

import (
	"fmt"
	"log"
	"net"

	"github.com/beijian128/minisocks/cmd"
	"github.com/beijian128/minisocks/core"
	"github.com/beijian128/minisocks/local"
)

var version = "master"

func main() {
	var err error
	config := cmd.ReadConfig()
	password, err := core.ParsePassword(config.Password)
	if err != nil {
		log.Fatalln(err)
	}
	localAddr, err := net.ResolveTCPAddr("tcp", config.ListenAddr)
	if err != nil {
		log.Fatalln(err)
	}
	serverAddr, err := net.ResolveTCPAddr("tcp", config.RemoteAddr)
	if err != nil {
		log.Fatalln(err)
	}
	lsLocal := local.New(password, localAddr, serverAddr)
	lsLocal.AfterListen = func(listenAddr net.Addr) {
		log.Printf("minisocks-client:%s 启动成功 监听在 %s\n", version, listenAddr.String())
		log.Println("使用配置：", fmt.Sprintf(`
本地监听地址 listen：
%s
远程服务地址 remote：
%s
密码 password：
%s
	`, config.ListenAddr, config.RemoteAddr, config.Password))
	}
	log.Fatalln(lsLocal.Listen())
}
