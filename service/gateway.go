package service

import (
	"context"
	"fmt"
	"github.com/smallnest/rpcx/client"
	"net/http"
	"sync"
	"sync/atomic"
)

// ServiceHandler 转换处理
type ServiceHandler func(*http.Request, string) (map[string]string, []byte, error)

type HTTPServer interface {
	RegisterHandler(base string, handler ServiceHandler)
	Serve() error
}

type Gateway struct {
	base             string
	httpServer       HTTPServer
	serviceDiscovery client.ServiceDiscovery
	FailMode         client.FailMode
	SelectMode       client.SelectMode
	Option           client.Option
	rwMux            sync.RWMutex
	xClients         map[string]client.XClient
	seq              uint64
}

func NewGateway(base string, srv HTTPServer, dis client.ServiceDiscovery, failMode client.FailMode, selectMode client.SelectMode, option client.Option) *Gateway {

	if base == "" {
		base = "/"
	}

	if base[0] != '/' {
		base = "/" + base
	}

	g := &Gateway{
		base:             base,
		httpServer:       srv,
		serviceDiscovery: dis,
		FailMode:         failMode,
		SelectMode:       selectMode,
		Option:           option,
		xClients:         make(map[string]client.XClient),
	}

	srv.RegisterHandler(base, g.handler)
	return g
}

func (g *Gateway) Serve() error {
	return g.httpServer.Serve()
}

func (g *Gateway) handler(r *http.Request, servicePath string) (map[string]string, []byte, error) {

	req, err := HttpRequest2RPCRequest(r)
	if err != nil {
		return nil, nil, err
	}
	seq := atomic.AddUint64(&g.seq, 1)
	req.SetSeq(seq)

	var xc client.XClient
	g.rwMux.Lock()
	xc, err = getXClient(g, servicePath)
	g.rwMux.Unlock()

	if err != nil {
		return nil, nil, err
	}

	// 直接处理
	//err = xc.Call(context.Background(), "hello", config.Request{}, config.Response{})
	//if err != nil {
	//	fmt.Println("Call err: ", err)
	//}

	return xc.SendRaw(context.Background(), req)
}

func getXClient(g *Gateway, servicePath string) (xClient client.XClient, err error) {

	defer func() {
		if e := recover(); e != nil {
			if ee, ok := e.(error); ok {
				err = ee
				return
			}
			err = fmt.Errorf("failed to get xclient: %v", e)
		}
	}()

	if g.xClients[servicePath] == nil {
		d, err := g.serviceDiscovery.Clone(servicePath)
		if err != nil {
			return nil, err
		}
		g.xClients[servicePath] = client.NewXClient(servicePath, g.FailMode, g.SelectMode, d, g.Option)
	}
	xClient = g.xClients[servicePath]

	return xClient, err
}
