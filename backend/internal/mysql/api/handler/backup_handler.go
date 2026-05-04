package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	apimiddleware "devops-console-backend/internal/mysql/api/middleware"
	"devops-console-backend/internal/mysql/model"
	"devops-console-backend/internal/mysql/service"
	"devops-console-backend/pkg/mysqlresponse"
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
		mysqlresponse.Error(ctx, http.StatusBadRequest, "query parameter database is required")
		return
	}

	records, err := h.service.ListBackups(database)
	if err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, model.BackupListResponse{Records: records})
}

func (h *BackupHandler) Create(ctx *gin.Context) {
	profile, ok := apimiddleware.GetProfileFromContext(ctx)
	if !ok {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, "connection profile not found in context")
		return
	}

	var req model.CreateBackupRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	taskID, err := h.service.CreateBackupAsync(ctx.Request.Context(), profile, req)
	if err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, model.CreateBackupResponse{TaskID: taskID})
}

func (h *BackupHandler) Restore(ctx *gin.Context) {
	profile, ok := apimiddleware.GetProfileFromContext(ctx)
	if !ok {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, "connection profile not found in context")
		return
	}

	var req model.RestoreBackupRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	taskID, err := h.service.RestoreBackupAsync(ctx.Request.Context(), profile, req)
	if err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, model.CreateBackupResponse{TaskID: taskID})
}

func (h *BackupHandler) Rename(ctx *gin.Context) {
	var req model.RenameBackupRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	record, err := h.service.RenameBackup(req)
	if err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, record)
}

func (h *BackupHandler) Delete(ctx *gin.Context) {
	var req model.DeleteBackupRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.DeleteBackup(req); err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, gin.H{"success": true})
}

func (h *BackupHandler) Download(ctx *gin.Context) {
	database := strings.TrimSpace(ctx.Query("database"))
	fileName := strings.TrimSpace(ctx.Query("fileName"))
	if database == "" || fileName == "" {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "database and fileName are required")
		return
	}

	reader, record, err := h.service.OpenBackupReadCloser(database, fileName)
	if err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
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
		mysqlresponse.Error(ctx, http.StatusBadRequest, "query parameter id is required")
		return
	}

	task, err := h.service.GetTask(taskID)
	if err != nil {
		mysqlresponse.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	mysqlresponse.Success(ctx, task)
}

func (h *BackupHandler) ListSchedules(ctx *gin.Context) {
	database := strings.TrimSpace(ctx.Query("database"))
	if database == "" {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "query parameter database is required")
		return
	}

	mysqlresponse.Success(ctx, h.service.ListSchedules(database))
}

func (h *BackupHandler) CreateSchedule(ctx *gin.Context) {
	profile, ok := apimiddleware.GetProfileFromContext(ctx)
	if !ok {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, "connection profile not found in context")
		return
	}

	var req model.CreateBackupScheduleRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	schedule, err := h.service.CreateSchedule(profile, req)
	if err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, schedule)
}

func (h *BackupHandler) DeleteSchedule(ctx *gin.Context) {
	var req model.DeleteBackupScheduleRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.DeleteSchedule(req.ID); err != nil {
		mysqlresponse.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	mysqlresponse.Success(ctx, gin.H{"success": true})
}


