package config

import (
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

type Config struct {
	Proxy ProxyConfig `yaml:"proxy"`
}

type ProxyConfig struct {
	Socks5 string `yaml:"socks5"`
}

func LoadConfig() (*Config, error) {
	// 获取程序执行路径
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// 读取配置文件
	data, err := os.ReadFile(filepath.Join(currentDir, "config.yml"))
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
