package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apimiddleware "mydeploy-project/internal/api/middleware"
	"mydeploy-project/internal/model"
	"mydeploy-project/internal/service"
	"mydeploy-project/pkg/response"
)

type QueryHandler struct {
	service *service.QueryService
}

func NewQueryHandler(service *service.QueryService) *QueryHandler {
	return &QueryHandler{service: service}
}

func (h *QueryHandler) Execute(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.ExecuteQueryRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	result, err := h.service.Execute(ctx.Request.Context(), db, req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

func (h *QueryHandler) ExecuteBatch(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.ExecuteBatchRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	result, err := h.service.ExecuteBatch(ctx.Request.Context(), db, req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}
