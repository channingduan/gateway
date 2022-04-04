package service

import (
	"context"
	"fmt"
	"github.com/channingduan/gateway/middleware"
	"github.com/channingduan/rpc/cache"
	"github.com/channingduan/rpc/config"
	"github.com/channingduan/rpc/server"
	"github.com/gin-gonic/gin"
	"github.com/oscto/ky3k"
	"github.com/smallnest/rpcx/codec"
	"net/http"
	"strings"
)

type Server struct {
	addr   string
	srv    *gin.Engine
	config *config.Config
	cache  *cache.Cache
}

// New 扩展其他 HTTP Server
func New(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

func NewWithGin(config *config.Config, srv *gin.Engine) *Server {
	return &Server{
		addr:   config.ServiceAddr,
		srv:    srv,
		config: config,
		cache:  cache.Register(&config.CacheConfig),
	}
}

func (s *Server) RegisterHandler(base string, handler ServiceHandler) {

	srv := s.srv
	if srv == nil {
		srv = gin.Default()
	}
	mw := middleware.NewMiddleware(s.config, s.cache)
	srv.Use(mw.Authorization())
	fun := wrapServiceHandler(handler)

	routers := s.cache.NewCache().SMembers(context.TODO(), server.RouterKey).Val()
	for _, router := range routers {
		router = strings.ReplaceAll(router, ".", "/")
		fmt.Println("router:", router)
		srv.POST(router, fun)
	}
	s.srv = srv
}

func wrapServiceHandler(handler ServiceHandler) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		r := ctx.Request
		w := ctx.Writer
		uris := strings.Split(strings.Trim(r.URL.String(), "/"), "/")

		r.Header.Set(XSerializeType, "3")
		if r.Header.Get(XServicePath) == "" {
			servicePath := uris[0]
			if strings.HasPrefix(servicePath, "/") {
				servicePath = servicePath[1:]
			}
			r.Header.Set(XServicePath, servicePath)
		}

		servicePath := r.Header.Get(XServicePath)
		if r.Header.Get(XServiceMethod) == "" {
			r.Header.Set(XServiceMethod, fmt.Sprintf("%s.%s", uris[1], uris[2]))
		}

		messageID := r.Header.Get(XMessageID)
		wh := w.Header()
		if messageID != "" {
			wh.Set(XMessageID, messageID)
		}

		meta, payload, err := handler(r, servicePath)
		for k, v := range meta {
			wh.Set(k, v)
		}

		var response config.HttpResponse
		if err == nil {
			cc := &codec.MsgpackCodec{}
			reply := &config.Response{}
			err = cc.Decode(payload, reply)
			response.Code = http.StatusOK

			var data map[string]interface{}
			if err := ky3k.StringToJson(reply.Message, &data); err != nil {
				return
			}

			response.Data = data

		} else {
			response.Message = err.Error()
		}

		ctx.JSON(http.StatusOK, response)
		return
	}
}

func (s *Server) Serve() error {
	return s.srv.Run(s.addr)
}
