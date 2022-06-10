package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/quanxiang-cloud/process/api/restful"
	"github.com/quanxiang-cloud/process/pkg/config"

	"github.com/quanxiang-cloud/process/pkg/misc/logger"
)

var (
	configPath = flag.String("config", "configs/config.yml", "-config 配置文件地址")
)

func main() {
	flag.Parse()

	err := config.Init(*configPath)
	if err != nil {
		panic(err)
	}

	err = logger.New(&config.Config.Log)
	if err != nil {
		panic(err)
	}

	// 启动路由
	router, err := restful.NewRouter(config.Config)
	if err != nil {
		panic(err)
	}
	go router.Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			router.Close()
			logger.Sync()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
