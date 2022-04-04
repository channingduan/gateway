package middleware

import (
	"github.com/channingduan/rpc/auth"
	"github.com/channingduan/rpc/cache"
	"github.com/channingduan/rpc/config"
)

type Middleware struct {
	auth *auth.Auth
}

func NewMiddleware(config *config.Config, cache *cache.Cache) *Middleware {
	return &Middleware{
		auth: auth.NewAuth(config, cache),
	}
}
