package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apimiddleware "mydeploy-project/internal/api/middleware"
	"mydeploy-project/internal/model"
	"mydeploy-project/internal/service"
	"mydeploy-project/pkg/response"
)

type MetadataHandler struct {
	service *service.MetadataService
}

func NewMetadataHandler(service *service.MetadataService) *MetadataHandler {
	return &MetadataHandler{service: service}
}

func (h *MetadataHandler) ListDatabases(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	databases, err := h.service.ListDatabases(ctx.Request.Context(), db)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, databases)
}

func (h *MetadataHandler) ListTables(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	databaseName := ctx.Query("db")
	if databaseName == "" {
		response.Error(ctx, http.StatusBadRequest, "query parameter db is required")
		return
	}

	tables, err := h.service.ListTables(ctx.Request.Context(), db, databaseName)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, tables)
}

func (h *MetadataHandler) CreateDatabase(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.CreateDatabaseRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.CreateDatabase(ctx.Request.Context(), db, req.Name); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, gin.H{"success": true})
}

func (h *MetadataHandler) RenameDatabase(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.RenameDatabaseRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.RenameDatabase(ctx.Request.Context(), db, req.OldName, req.NewName); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, gin.H{"success": true})
}

func (h *MetadataHandler) DeleteDatabase(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.DeleteDatabaseRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.DeleteDatabase(ctx.Request.Context(), db, req.Name); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, gin.H{"success": true})
}

func (h *MetadataHandler) CreateTable(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.CreateTableRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.CreateTable(ctx.Request.Context(), db, req.Database, req.Name, req.Columns, req.Options); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, gin.H{"success": true})
}

func (h *MetadataHandler) AutoImportTable(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.AutoImportTableRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	result, err := h.service.AutoImportTable(ctx.Request.Context(), db, req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

func (h *MetadataHandler) RenameTable(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.RenameTableRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.RenameTable(ctx.Request.Context(), db, req.Database, req.OldName, req.NewName); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, gin.H{"success": true})
}

func (h *MetadataHandler) DeleteTable(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.DeleteTableRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.DeleteTable(ctx.Request.Context(), db, req.Database, req.Name); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, gin.H{"success": true})
}
