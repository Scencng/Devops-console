package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mydeploy-project/pkg/response"
)

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) == 0 || ctx.Writer.Written() {
			return
		}

		response.Error(ctx, http.StatusInternalServerError, ctx.Errors.Last().Error())
	}
}
