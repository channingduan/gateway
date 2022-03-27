package service

import (
	"errors"
	"fmt"
	"github.com/smallnest/rpcx/client"
	"strings"
)

func CreateServiceDiscovery(basePath, addr string) (client.ServiceDiscovery, error) {

	i := strings.Index(addr, "://")
	if i < 0 {
		return nil, errors.New("wrong format registry address, the right format is registry_type://address")
	}

	regType := addr[:i]
	addr = addr[i+3:]
	switch regType {
	case "peer2peer":
		return client.NewPeer2PeerDiscovery(fmt.Sprintf("tcp@%s", addr), "")
	case "multiple":
		var pairs []*client.KVPair
		regs := strings.Split(addr, ",")
		for _, v := range regs {
			pairs = append(pairs, &client.KVPair{Key: v})
		}
		return client.NewMultipleServersDiscovery(pairs)
	case "zookeeper":
		return client.NewZookeeperDiscoveryTemplate(basePath, []string{addr}, nil)
	case "consul":
		return client.NewConsulDiscoveryTemplate(basePath, []string{addr}, nil)
	case "redis":
		return client.NewRedisDiscoveryTemplate(basePath, []string{addr}, nil)
	default:
		return nil, fmt.Errorf("wrong registry type %s, only support peer2peer, multiple zookeeper consul redis", regType)
	}
}
