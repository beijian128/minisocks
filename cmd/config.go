// cmd 包提供命令行相关的功能，主要负责配置文件的读取和保存
package cmd

import (
	"encoding/json" // 提供 JSON 编码和解码功能
	"log"           // 提供简单的日志记录功能
	"os"            // 提供与操作系统交互的功能
	"path"          // 提供对路径进行操作的功能

	"github.com/beijian128/minisocks/core" // 引入 minisocks 核心功能包
	"github.com/mitchellh/go-homedir"      // 提供获取用户主目录的功能
)

// Config 结构体定义了 minisocks 的配置信息
type Config struct {
	// ListenAddr 是本地监听地址，对应 JSON 中的 "listen" 字段
	ListenAddr string `json:"listen"`
	// RemoteAddr 是远程服务地址，对应 JSON 中的 "remote" 字段
	RemoteAddr string `json:"remote"`
	// Password 是连接使用的密码，对应 JSON 中的 "password" 字段
	Password string `json:"password"`
}

// 配置文件路径，用于存储 minisocks 的配置信息
var configPath string

// init 函数在包被加载时自动执行，用于初始化配置文件路径
func init() {
	// 获取用户主目录，忽略可能的错误
	home, _ := homedir.Dir()
	// 拼接配置文件的完整路径
	configPath = path.Join(home, ".minisocks.json")
}

// SaveConfig 将配置信息保存到配置文件中
func (config *Config) SaveConfig() {
	// 将配置信息编码为格式化的 JSON 字符串，忽略可能的错误
	configJson, _ := json.MarshalIndent(config, "", "	")
	// 将 JSON 字符串写入配置文件
	err := os.WriteFile(configPath, configJson, 0644)
	if err != nil {
		// 若写入失败，记录错误日志并终止程序
		log.Fatalf("保存配置到文件 %s 出错: %v", configPath, err)
	}
	// 若写入成功，记录成功日志
	log.Printf("保存配置到文件 %s 成功\n", configPath)
}

// ReadConfig 读取配置文件中的配置信息，如果文件不存在则使用默认配置
func ReadConfig() *Config {
	// 创建一个默认配置实例
	config := &Config{
		ListenAddr: ":7448",
		RemoteAddr: ":7448",
		Password:   core.RandPassword().String(),
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		// 若文件存在，记录读取日志
		log.Printf("从文件 %s 中读取配置\n", configPath)
		// 打开配置文件
		file, err := os.Open(configPath)
		if err != nil {
			// 若打开失败，记录错误日志并终止程序
			log.Fatalf("打开文件 %s 出错:%s", configPath, err)
		}
		// 函数结束时关闭文件
		defer file.Close()

		// 从文件中解析 JSON 数据到配置实例
		err = json.NewDecoder(file).Decode(config)
		if err != nil {
			// 若解析失败，记录错误日志并终止程序
			log.Fatalf("格式不合法的 JSON 配置文件:\n%s", file.Name())
		}
	}
	// 保存配置信息到文件
	config.SaveConfig()
	// 返回配置实例
	return config
}
