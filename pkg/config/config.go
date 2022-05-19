package config

import (
	"git.internal.yunify.com/qxp/misc/client"
	"git.internal.yunify.com/qxp/misc/mysql2"
	"git.internal.yunify.com/qxp/misc/redis2"
	"io/ioutil"
	"time"

	"git.internal.yunify.com/qxp/misc/logger"
	"gopkg.in/yaml.v2"
)

// Config 全局配置对象
var Config *Configs

// Configs 总配置结构体
type Configs struct {
	Model         string        `yaml:"model"`
	Port          string        `yaml:"port"`
	Mysql         mysql2.Config `yaml:"mysql"`
	Log           logger.Config `yaml:"log"`
	InternalNet   client.Config `yaml:"internalNet"`
	Redis         redis2.Config `yaml:"redis"`
	APIHost       APIHost       `yaml:"api"`
	FlowRPCServer string        `yaml:"flowRpcServer"`
}

// HTTPServer http服务配置
type HTTPServer struct {
	Port              string        `yaml:"port"`
	ReadHeaderTimeOut time.Duration `yaml:"readHeaderTimeOut"`
	WriteTimeOut      time.Duration `yaml:"writeTimeOut"`
	MaxHeaderBytes    int           `yaml:"maxHeaderBytes"`
}

// APIHost api host
type APIHost struct {
	OrgHost  string `yaml:"orgHost" validate:"required"`
	FlowHost string `yaml:"flowHost" validate:"required"`
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
