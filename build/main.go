package main

import (
	"github.com/channingduan/gateway/config"
	"github.com/channingduan/gateway/service"
	"github.com/smallnest/rpcx/client"
	"log"
)

func main() {

	conf := config.Config{
		Addr:         "127.0.0.1:9090",
		RegistryAddr: "consul://127.0.0.1:8500",
		BasePath:     "rpc",
		FailMode:     int(client.Failover),
		SelectMode:   int(client.RandomSelect),
	}

	d, err := service.CreateServiceDiscovery(conf.BasePath, conf.RegistryAddr)
	if err != nil {
		log.Fatal(err)
	}
	httpServer := service.New(conf.Addr)
	gw := service.NewGateway("/", httpServer, d, client.FailMode(conf.FailMode), client.SelectMode(conf.SelectMode), client.DefaultOption)
	err = gw.Serve()
	if err != nil {
		log.Fatal("Serve", err)
	}
}
