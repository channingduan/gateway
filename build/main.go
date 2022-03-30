package main

import (
	"github.com/channingduan/gateway/service"
	"github.com/channingduan/rpc/config"
	"github.com/gin-gonic/gin"
	"github.com/smallnest/rpcx/client"
	"log"
)

func main() {

	conf, err := config.Register("./server.json")
	if err != nil {
		log.Fatal("config file error: ", err)
	}

	discovery, err := service.CreateServiceDiscovery(conf.BasePath, conf.RegistryConfig.Addr)
	if err != nil {
		log.Fatal("service discovery error: ", err)
	}

	httpServer := service.NewWithGin(conf, gin.Default())
	gw := service.NewGateway("/", httpServer, discovery, client.Failover, client.RandomSelect, client.DefaultOption)
	err = gw.Serve()
	if err != nil {
		log.Fatal("serve error", err)
	}
}
