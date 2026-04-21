package kafka

import (
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (c *Controller) ScanKafkaNetwork(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.DiscoveryScanRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ScanKafkaNetwork(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("扫描完成", "data", data)
}

func (c *Controller) ProbeKafkaBootstrapServers(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.DiscoveryProbeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ProbeKafkaBootstrapServers(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("探测完成", "data", data)
}

func (c *Controller) ImportDiscoveredKafka(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.DiscoveryImportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ImportDiscoveredKafka(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, data.ID, "cluster:discovery:import", "cluster", data.Name, sanitizeDiscoveryImportPayload(req), "success", "")
	helper.SuccessWithData("导入成功", "data", data)
}

func sanitizeDiscoveryImportPayload(req reqKafka.DiscoveryImportRequest) map[string]interface{} {
	return map[string]interface{}{
		"name":               req.Name,
		"address":            req.Address,
		"environment":        req.Environment,
		"tenant":             req.Tenant,
		"description":        req.Description,
		"version":            req.Auth.Version,
		"authType":           req.Auth.AuthType,
		"username":           req.Auth.Username,
		"tlsEnabled":         req.Auth.TLSEnabled,
		"insecureSkipVerify": req.Auth.InsecureSkipVerify,
		"hasPassword":        req.Auth.Password != "",
		"hasClientKey":       req.Auth.ClientKey != "",
	}
}
