package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	apimiddleware "mydeploy-project/internal/api/middleware"
	"mydeploy-project/internal/model"
	"mydeploy-project/internal/service"
	"mydeploy-project/pkg/response"
)

type BackupHandler struct {
	service *service.BackupService
}

func NewBackupHandler(service *service.BackupService) *BackupHandler {
	return &BackupHandler{service: service}
}

func (h *BackupHandler) List(ctx *gin.Context) {
	database := strings.TrimSpace(ctx.Query("database"))
	if database == "" {
		response.Error(ctx, http.StatusBadRequest, "query parameter database is required")
		return
	}

	records, err := h.service.ListBackups(database)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, model.BackupListResponse{Records: records})
}

func (h *BackupHandler) Create(ctx *gin.Context) {
	profile, ok := apimiddleware.GetProfileFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "connection profile not found in context")
		return
	}

	var req model.CreateBackupRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	taskID, err := h.service.CreateBackupAsync(ctx.Request.Context(), profile, req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, model.CreateBackupResponse{TaskID: taskID})
}

func (h *BackupHandler) Restore(ctx *gin.Context) {
	profile, ok := apimiddleware.GetProfileFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "connection profile not found in context")
		return
	}

	var req model.RestoreBackupRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	taskID, err := h.service.RestoreBackupAsync(ctx.Request.Context(), profile, req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, model.CreateBackupResponse{TaskID: taskID})
}

func (h *BackupHandler) Rename(ctx *gin.Context) {
	var req model.RenameBackupRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	record, err := h.service.RenameBackup(req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, record)
}

func (h *BackupHandler) Delete(ctx *gin.Context) {
	var req model.DeleteBackupRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.DeleteBackup(req); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, gin.H{"success": true})
}

func (h *BackupHandler) Download(ctx *gin.Context) {
	database := strings.TrimSpace(ctx.Query("database"))
	fileName := strings.TrimSpace(ctx.Query("fileName"))
	if database == "" || fileName == "" {
		response.Error(ctx, http.StatusBadRequest, "database and fileName are required")
		return
	}

	reader, record, err := h.service.OpenBackupReadCloser(database, fileName)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer reader.Close()

	ctx.Header("Content-Disposition", `attachment; filename="`+record.FileName+`"`)
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Cache-Control", "no-store")
	_, _ = io.Copy(ctx.Writer, reader)
}

func (h *BackupHandler) Task(ctx *gin.Context) {
	taskID := strings.TrimSpace(ctx.Query("id"))
	if taskID == "" {
		response.Error(ctx, http.StatusBadRequest, "query parameter id is required")
		return
	}

	task, err := h.service.GetTask(taskID)
	if err != nil {
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	response.Success(ctx, task)
}

func (h *BackupHandler) ListSchedules(ctx *gin.Context) {
	database := strings.TrimSpace(ctx.Query("database"))
	if database == "" {
		response.Error(ctx, http.StatusBadRequest, "query parameter database is required")
		return
	}

	response.Success(ctx, h.service.ListSchedules(database))
}

func (h *BackupHandler) CreateSchedule(ctx *gin.Context) {
	profile, ok := apimiddleware.GetProfileFromContext(ctx)
	if !ok {
		response.Error(ctx, http.StatusInternalServerError, "connection profile not found in context")
		return
	}

	var req model.CreateBackupScheduleRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	schedule, err := h.service.CreateSchedule(profile, req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, schedule)
}

func (h *BackupHandler) DeleteSchedule(ctx *gin.Context) {
	var req model.DeleteBackupScheduleRequest
	if err := bindJSON(ctx, &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.DeleteSchedule(req.ID); err != nil {
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	response.Success(ctx, gin.H{"success": true})
}
