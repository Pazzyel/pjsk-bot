package config

import (
	"errors"

	"github.com/spf13/viper"
)

type PJSKConfig struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
	PJSK struct {
		Charts struct {
			RequestPath string `mapstructure:"request_path"`
			SavePath    string `mapstructure:"save_path"`
		} `mapstructure:"charts"`
		Jackets struct {
			RequestPath string `mapstructure:"request_path"`
			SavePath    string `mapstructure:"save_path"`
		} `mapstructure:"jackets"`
	} `mapstructure:"pjsk"`
}

func LoadConfig(path string) (*PJSKConfig, error) {
	// 初始化 viper
	vp := viper.New()
	vp.AddConfigPath(path)
	vp.SetConfigName("application")
	vp.SetConfigType("yaml")

	// 设置默认值
	vp.SetDefault("server.host", "127.0.0.1")
	vp.SetDefault("server.port", 9470)
	vp.SetDefault("pjsk.charts.request-path", "https://sdvx.in/prsk/obj/data")
	vp.SetDefault("pjsk.charts.save-path", "resources/images/charts")
	vp.SetDefault("pjsk.jackets.request-path", "https://sdvx.in/prsk/jacket/")
	vp.SetDefault("pjsk.jackets.save-path", "resources/images/jackets")

	// 自动环境变量
	vp.AutomaticEnv()

	// 读取配置文件
	if err := vp.ReadInConfig(); err != nil {
		return nil, errors.New("读取配置文件失败: " + err.Error())
	}

	// 映射到结构体
	var config PJSKConfig
	if err := vp.Unmarshal(&config); err != nil {
		return nil, errors.New("解析配置文件失败: " + err.Error())
	}
	return &config, nil
}