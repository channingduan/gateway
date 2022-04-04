package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

func (m *Middleware) Authorization() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		bearer := ctx.GetHeader("Authorization")

		token := strings.Split(bearer, " ")
		if bearer == "" || len(token) != 2 {
			ctx.Next()
			return
		}
		err := m.auth.ValidToken(token[1])
		fmt.Println("error: --- ", err)

		ctx.Next()
	}
}
