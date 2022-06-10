package config

import (
	"github.com/quanxiang-cloud/process/pkg/misc/client"
	"github.com/quanxiang-cloud/process/pkg/misc/mysql2"
	"github.com/quanxiang-cloud/process/pkg/misc/redis2"
	"io/ioutil"
	"time"

	"github.com/quanxiang-cloud/process/pkg/misc/logger"
	"gopkg.in/yaml.v2"
)

// Config 全局配置对象
var Config *Configs

// Configs 总配置结构体
type Configs struct {
	Model       string        `yaml:"model"`
	Port        string        `yaml:"port"`
	Mysql       mysql2.Config `yaml:"mysql"`
	Log         logger.Config `yaml:"log"`
	InternalNet client.Config `yaml:"internalNet"`
	Redis       redis2.Config `yaml:"redis"`
}

// HTTPServer http服务配置
type HTTPServer struct {
	Port              string        `yaml:"port"`
	ReadHeaderTimeOut time.Duration `yaml:"readHeaderTimeOut"`
	WriteTimeOut      time.Duration `yaml:"writeTimeOut"`
	MaxHeaderBytes    int           `yaml:"maxHeaderBytes"`
}

// Init 初始化
func Init(configPath string) error {
	if configPath == "" {
		configPath = "../configs/configs.yml"
	}
	Config = new(Configs)
	err := read(configPath, Config)
	if err != nil {
		return err
	}
	return nil
}

// read 读取配置文件
func read(yamlPath string, v interface{}) error {
	// Read config file
	buf, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(buf, v)
	if err != nil {
		return err
	}
	return nil
}
