package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		method := ctx.Request.Method

		ctx.Next()

		log.Printf(
			"method=%s path=%s status=%d duration=%s client_ip=%s",
			method,
			path,
			ctx.Writer.Status(),
			time.Since(start),
			ctx.ClientIP(),
		)
	}
}
