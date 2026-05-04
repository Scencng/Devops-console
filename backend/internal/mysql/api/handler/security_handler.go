package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apimiddleware "devops-console-backend/internal/mysql/api/middleware"
	"devops-console-backend/internal/mysql/model"
	"devops-console-backend/internal/mysql/service"
	"devops-console-backend/pkg/mysqlresponse"
)

type SecurityHandler struct {
	service *service.SecurityService
}

func NewSecurityHandler(service *service.SecurityService) *SecurityHandler {
	return &SecurityHandler{service: service}
}

func (h *SecurityHandler) Overview(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	result, err := h.service.GetOverview(ctx.Request.Context(), db)
	if err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, result)
}

func (h *SecurityHandler) GetPrincipal(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.GetSecurityPrincipalRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request query: "+err.Error())
		return
	}

	detail, err := h.service.GetPrincipalDetail(ctx.Request.Context(), db, req.User, req.Host, model.SecurityPrincipalKind(req.Kind))
	if err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, detail)
}

func (h *SecurityHandler) CreatePrincipal(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.UpsertSecurityPrincipalRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.CreatePrincipal(ctx.Request.Context(), db, req); err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, gin.H{"success": true})
}

func (h *SecurityHandler) UpdatePrincipal(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.UpsertSecurityPrincipalRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.UpdatePrincipal(ctx.Request.Context(), db, req); err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, gin.H{"success": true})
}

func (h *SecurityHandler) DeletePrincipal(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.DeleteSecurityPrincipalRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.DeletePrincipal(ctx.Request.Context(), db, req); err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, gin.H{"success": true})
}

func (h *SecurityHandler) ClonePrincipal(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.CloneSecurityPrincipalRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.ClonePrincipal(ctx.Request.Context(), db, req); err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, gin.H{"success": true})
}

func (h *SecurityHandler) RevokeAll(ctx *gin.Context) {
	db, ok := apimiddleware.GetDBFromContext(ctx)
	if !ok {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, "database connection not found in context")
		return
	}

	var req model.RevokeAllSecurityPrincipalRequest
	if err := bindJSON(ctx, &req); err != nil {
		mysqlresponse.Error(ctx, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := h.service.RevokeAll(ctx.Request.Context(), db, req); err != nil {
		mysqlresponse.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	mysqlresponse.Success(ctx, gin.H{"success": true})
}


