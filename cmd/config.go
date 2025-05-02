package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/beijian128/minisocks/core"
	"github.com/sirupsen/logrus"
)

const (
	defaultConfigPath = "./minisocks.json"
	defaultListenAddr = ":7448"
	defaultRemoteAddr = "ip:7448"
)

// Config 定义了 minisocks 的配置信息
type Config struct {
	ListenAddr string `json:"listen"`   // 本地监听地址
	RemoteAddr string `json:"remote"`   // 远程服务地址
	Password   string `json:"password"` // 连接使用的密码
}

var (
	logger = logrus.WithField("component", "cmd")
)

// Save 将配置信息保存到配置文件
func (c *Config) Save() error {
	configJSON, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		logger.WithError(err).Error("配置序列化失败")
		return fmt.Errorf("配置序列化失败: %w", err)
	}

	if err := os.WriteFile(defaultConfigPath, configJSON, 0644); err != nil {
		logger.WithFields(logrus.Fields{
			"path":  defaultConfigPath,
			"error": err,
		}).Error("保存配置文件失败")
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	logger.WithField("path", defaultConfigPath).Info("配置文件保存成功")
	return nil
}

// LoadConfig 读取配置文件中的配置信息，如果文件不存在则使用默认配置
func LoadConfig() (*Config, error) {
	config := &Config{
		ListenAddr: defaultListenAddr,
		RemoteAddr: defaultRemoteAddr,
		Password:   core.GenerateCipherTable(),
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(defaultConfigPath); os.IsNotExist(err) {
		logger.WithField("path", defaultConfigPath).Info("配置文件不存在，使用默认配置")
		if err := config.Save(); err != nil {
			return nil, err
		}
		return config, nil
	}

	file, err := os.Open(defaultConfigPath)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"path":  defaultConfigPath,
			"error": err,
		}).Error("打开配置文件失败")
		return nil, fmt.Errorf("打开配置文件失败: %w", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(config); err != nil {
		logger.WithFields(logrus.Fields{
			"path":  defaultConfigPath,
			"error": err,
		}).Error("解析配置文件失败")
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	logger.WithField("path", defaultConfigPath).Info("配置文件加载成功")
	return config, nil
}
