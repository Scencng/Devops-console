package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apimiddleware "mydeploy-project/internal/api/middleware"
	"mydeploy-project/internal/service"
	"mydeploy-project/pkg/response"
)

type DataHandler struct {
	service *service.DataService
}

func NewDataHandler(service *service.DataService) *DataHandler {
	return &DataHandler{service: service}
}

func (h *DataHandler) GetTableData(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	databaseName := ctx.Query("db")
	tableName := ctx.Query("table")
	if databaseName == "" || tableName == "" {
		response.Error(ctx, http.StatusBadRequest, "query parameters db and table are required")
		return
	}

	limit := 50
	if rawLimit := ctx.DefaultQuery("limit", "50"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit <= 0 {
			response.Error(ctx, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		limit = parsedLimit
	}

	offset := 0
	if rawOffset := ctx.DefaultQuery("offset", "0"); rawOffset != "" {
		parsedOffset, err := strconv.Atoi(rawOffset)
		if err != nil || parsedOffset < 0 {
			response.Error(ctx, http.StatusBadRequest, "offset must be a non-negative integer")
			return
		}
		offset = parsedOffset
	}

	keyword := ctx.Query("keyword")
	sortBy := ctx.Query("sortBy")
	sortOrder := ctx.DefaultQuery("sortOrder", "asc")

	result, err := h.service.GetTableData(ctx.Request.Context(), db, databaseName, tableName, limit, offset, keyword, sortBy, sortOrder)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}
