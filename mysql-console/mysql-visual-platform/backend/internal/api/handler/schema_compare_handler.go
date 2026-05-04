package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apimiddleware "mydeploy-project/internal/api/middleware"
	"mydeploy-project/internal/model"
	"mydeploy-project/internal/service"
	"mydeploy-project/pkg/response"
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
		response.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.SchemaCompareRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	result, err := h.service.Compare(ctx.Request.Context(), db, req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}
