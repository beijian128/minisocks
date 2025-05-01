// main 包表示该程序是一个可执行程序。
package main

import (
	"fmt" // 提供格式化输入输出功能
	"log" // 提供简单的日志记录功能
	"net" // 提供网络相关的功能

	"github.com/beijian128/minisocks/cmd"
	"github.com/beijian128/minisocks/local"
)

// 定义全局变量，用于存储程序的版本信息、提交哈希和构建日期
var (
	// version 表示程序的版本号，当前为开发版本
	version = "dev"
	// commit 表示代码的提交哈希，默认值为 "none"
	commit = "none"
	// date 表示程序的构建日期，默认值为 "unknown"
	date = "unknown"
)

// main 函数是程序的入口点
func main() {
	var err error
	// 从配置文件中读取配置信息
	config := cmd.ReadConfig()

	// 将本地监听地址解析为 TCP 地址
	localAddr, err := net.ResolveTCPAddr("tcp", config.ListenAddr)
	// 若解析失败，记录错误日志并终止程序
	if err != nil {
		log.Fatalln(err)
	}
	// 将远程服务地址解析为 TCP 地址
	serverAddr, err := net.ResolveTCPAddr("tcp", config.RemoteAddr)
	// 若解析失败，记录错误日志并终止程序
	if err != nil {
		log.Fatalln(err)
	}
	// 创建一个本地代理实例
	lsLocal := local.New(localAddr, serverAddr)
	// 设置本地代理监听成功后的回调函数
	lsLocal.AfterListen = func(listenAddr net.Addr) {
		// 记录程序启动成功的日志，包含版本号和监听地址
		log.Printf("minisocks-client:%s 启动成功 监听在 %s\n", version, listenAddr.String())
		// 记录程序使用的配置信息
		log.Println("使用配置：", fmt.Sprintf(`
本地监听地址 listen：
%s
远程服务地址 remote：
%s
密码 password：
%s
	`, config.ListenAddr, config.RemoteAddr, config.Password))
	}
	// 启动本地代理监听，并记录错误日志（若有），终止程序
	log.Fatalln(lsLocal.Listen())
}
