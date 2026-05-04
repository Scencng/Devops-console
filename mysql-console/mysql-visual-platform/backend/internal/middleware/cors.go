package middleware

import "github.com/gin-gonic/gin"

func CORS() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		headers := ctx.Writer.Header()
		headers.Set("Access-Control-Allow-Origin", "*")
		headers.Set("Access-Control-Allow-Credentials", "true")
		headers.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept, Accept-Encoding, Authorization, X-Requested-With, X-Connection-Token")
		headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		headers.Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}

		ctx.Next()
	}
}
