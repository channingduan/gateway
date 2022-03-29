package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type Server struct {
	addr string
	g    *gin.Engine
}

func New(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

func NewWithGin(addr string, g *gin.Engine) *Server {
	return &Server{
		addr: addr,
		g:    g,
	}
}

func (s *Server) RegisterHandler(base string, handler ServiceHandler) {

	g := s.g
	if g == nil {
		g = gin.Default()
	}
	fun := wrapServiceHandler(handler)

	fmt.Println("base-------:", base)
	g.POST(base, fun)
	g.GET(base, fun)
	g.PUT(base, fun)
	g.POST("/service/hello/world", fun)
	s.g = g
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
		if err == nil {
			ctx.Data(http.StatusOK, "application/octet-stream", payload)
			return
		}

		rh := r.Header
		for k, v := range rh {
			if strings.HasPrefix(k, "x-RPCX-") && len(v) > 0 {
				wh.Set(k, v[0])
			}
		}

		wh.Set(XMessageStatusType, "Error")
		wh.Set(XErrorMessage, err.Error())

		ctx.String(http.StatusOK, err.Error())
	}
}

func (s *Server) Serve() error {
	return s.g.Run(s.addr)
}
