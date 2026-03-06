package config

import (
	"errors"

	"github.com/spf13/viper"
)

type QQConfig struct {
	QQ struct {
		Number   string `mapstructure:"number"`
		Password string `mapstructure:"password"`
	} `mapstructure:"qq"`
}

func LoadConfig(path string) (*QQConfig, error) {
	// 初始化 viper
	vp := viper.New()
	vp.AddConfigPath(path)
	vp.SetConfigName("application")
	vp.SetConfigType("yaml")

	// 自动环境变量
	vp.AutomaticEnv()

	// 读取配置文件
	if err := vp.ReadInConfig(); err != nil {
		return nil, errors.New("读取配置文件失败: " + err.Error())
	}

	// 映射到结构体
	var config QQConfig
	if err := vp.Unmarshal(&config); err != nil {
		return nil, errors.New("解析配置文件失败: " + err.Error())
	}
	return &config, nil
}
