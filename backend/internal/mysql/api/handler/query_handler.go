package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apimiddleware "devops-console-backend/internal/mysql/api/middleware"
	"devops-console-backend/internal/mysql/model"
	"devops-console-backend/internal/mysql/service"
	"devops-console-backend/pkg/mysqlresponse"
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
		mysqlresponse.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.ExecuteQueryRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	result, err := h.service.Execute(ctx.Request.Context(), db, req)
	if err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, result)
}

func (h *QueryHandler) ExecuteBatch(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.ExecuteBatchRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	result, err := h.service.ExecuteBatch(ctx.Request.Context(), db, req)
	if err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, result)
}


