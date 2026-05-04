package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apimiddleware "devops-console-backend/internal/mysql/api/middleware"
	"devops-console-backend/internal/mysql/model"
	"devops-console-backend/internal/mysql/service"
	"devops-console-backend/pkg/mysqlresponse"
)

type SchemaCompareHandler struct {
	service *service.SchemaCompareService
}

func NewSchemaCompareHandler(service *service.SchemaCompareService) *SchemaCompareHandler {
	return &SchemaCompareHandler{service: service}
}

func (h *SchemaCompareHandler) Compare(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.SchemaCompareRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	result, err := h.service.Compare(ctx.Request.Context(), db, req)
	if err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, result)
}


