package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mydeploy-project/internal/model"
	"mydeploy-project/internal/service"
	"mydeploy-project/pkg/response"
)

type ConnectionHandler struct {
	manager *service.ConnectionManager
}

func NewConnectionHandler(manager *service.ConnectionManager) *ConnectionHandler {
	return &ConnectionHandler{manager: manager}
}

func (h *ConnectionHandler) Open(ctx *gin.Context) {
	var req model.OpenConnectionRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	token, err := h.manager.Open(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"connectionToken": token,
	})
}

func (h *ConnectionHandler) Close(ctx *gin.Context) {
	var req model.CloseConnectionRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.manager.Close(req.ConnectionToken); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"connectionToken": req.ConnectionToken,
	})
}
