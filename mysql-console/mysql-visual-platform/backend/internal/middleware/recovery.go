package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"mydeploy-project/pkg/response"
)

func Recovery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v", rec)
				response.Error(ctx, http.StatusInternalServerError, "internal server error")
				ctx.Abort()
			}
		}()

		ctx.Next()
	}
}
