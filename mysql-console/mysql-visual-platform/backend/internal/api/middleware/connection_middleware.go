package middleware

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"mydeploy-project/internal/model"
	"mydeploy-project/internal/service"
	"mydeploy-project/pkg/response"
)

const dbContextKey = "connection_db"
const profileContextKey = "connection_profile"
const tokenContextKey = "connection_token"

func ConnectionToken(manager *service.ConnectionManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("X-Connection-Token")
		if token == "" {
			response.Error(ctx, http.StatusUnauthorized, "missing X-Connection-Token header")
			ctx.Abort()
			return
		}

		session, err := manager.GetSession(token)
		if err != nil {
			response.Error(ctx, http.StatusUnauthorized, err.Error())
			ctx.Abort()
			return
		}

		ctx.Set(dbContextKey, session.DB)
		ctx.Set(profileContextKey, session.Profile)
		ctx.Set(tokenContextKey, token)
		ctx.Next()
	}
}

func GetDBFromContext(ctx *gin.Context) (*sql.DB, bool) {
	value, ok := ctx.Get(dbContextKey)
	if !ok {
		return nil, false
	}

	db, ok := value.(*sql.DB)
	return db, ok
}

func GetProfileFromContext(ctx *gin.Context) (model.OpenConnectionRequest, bool) {
	value, ok := ctx.Get(profileContextKey)
	if !ok {
		return model.OpenConnectionRequest{}, false
	}

	profile, ok := value.(model.OpenConnectionRequest)
	return profile, ok
}

func GetConnectionTokenFromContext(ctx *gin.Context) (string, bool) {
	value, ok := ctx.Get(tokenContextKey)
	if !ok {
		return "", false
	}

	token, ok := value.(string)
	return token, ok
}
