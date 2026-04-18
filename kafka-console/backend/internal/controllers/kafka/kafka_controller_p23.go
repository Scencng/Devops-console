package kafka

import (
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (c *Controller) ListAlertSilences(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.AlertSilenceListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.ListAlertSilences(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateAlertSilence(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.AlertSilenceUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.CreateAlertSilence(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("创建成功", "data", data)
}

func (c *Controller) UpdateAlertSilence(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的静默 ID"); return }
	var req reqKafka.AlertSilenceUpsertRequest
	if err = ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.UpdateAlertSilence(id, req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("更新成功", "data", data)
}

func (c *Controller) DeleteAlertSilence(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的静默 ID"); return }
	if err = c.service.DeleteAlertSilence(id); err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("删除成功", "data", nil)
}

func (c *Controller) ListTraceLinks(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.TraceLinkListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.ListTraceLinks(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateTraceLink(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.TraceLinkCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.CreateTraceLink(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("创建成功", "data", data)
}

func (c *Controller) DeleteTraceLink(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的 Trace ID"); return }
	if err = c.service.DeleteTraceLink(id); err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("删除成功", "data", nil)
}

func (c *Controller) GenerateScalingRecommendations(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ScalingRecommendationRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.GenerateScalingRecommendations(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("生成成功", "data", data)
}

func (c *Controller) ListSelfHealingPolicies(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.SelfHealingPolicyListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.ListSelfHealingPolicies(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateSelfHealingPolicy(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.SelfHealingPolicyUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.CreateSelfHealingPolicy(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("创建成功", "data", data)
}

func (c *Controller) UpdateSelfHealingPolicy(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的策略 ID"); return }
	var req reqKafka.SelfHealingPolicyUpsertRequest
	if err = ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.UpdateSelfHealingPolicy(id, req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("更新成功", "data", data)
}

func (c *Controller) DeleteSelfHealingPolicy(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的策略 ID"); return }
	if err = c.service.DeleteSelfHealingPolicy(id); err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("删除成功", "data", nil)
}

func (c *Controller) ExecuteSelfHealingPolicy(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的策略 ID"); return }
	data, err := c.service.ExecuteSelfHealingPolicy(id)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("执行完成", "data", data)
}

func (c *Controller) ListSelfHealingExecutions(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.SelfHealingExecutionListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.ListSelfHealingExecutions(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) ListGitOpsProfiles(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.GitOpsProfileListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.ListGitOpsProfiles(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateGitOpsProfile(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.GitOpsProfileUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.CreateGitOpsProfile(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("创建成功", "data", data)
}

func (c *Controller) UpdateGitOpsProfile(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的 GitOps ID"); return }
	var req reqKafka.GitOpsProfileUpsertRequest
	if err = ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.UpdateGitOpsProfile(id, req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("更新成功", "data", data)
}

func (c *Controller) DeleteGitOpsProfile(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的 GitOps ID"); return }
	if err = c.service.DeleteGitOpsProfile(id); err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("删除成功", "data", nil)
}

func (c *Controller) RunGitOpsSync(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的 GitOps ID"); return }
	data, err := c.service.RunGitOpsSync(id)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("同步完成", "data", data)
}

func (c *Controller) ListGitOpsSyncs(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.GitOpsSyncListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.ListGitOpsSyncs(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) ListCloudAdapters(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.CloudAdapterListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.ListCloudAdapters(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateCloudAdapter(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.CloudAdapterUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.CreateCloudAdapter(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("创建成功", "data", data)
}

func (c *Controller) UpdateCloudAdapter(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的云适配 ID"); return }
	var req reqKafka.CloudAdapterUpsertRequest
	if err = ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.UpdateCloudAdapter(id, req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("更新成功", "data", data)
}

func (c *Controller) DeleteCloudAdapter(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的云适配 ID"); return }
	if err = c.service.DeleteCloudAdapter(id); err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("删除成功", "data", nil)
}

func (c *Controller) ListLineageRelations(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.LineageListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.ListLineageRelations(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateLineageRelation(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.LineageUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.CreateLineageRelation(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("创建成功", "data", data)
}

func (c *Controller) UpdateLineageRelation(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的血缘 ID"); return }
	var req reqKafka.LineageUpsertRequest
	if err = ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.UpdateLineageRelation(id, req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("更新成功", "data", data)
}

func (c *Controller) DeleteLineageRelation(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的血缘 ID"); return }
	if err = c.service.DeleteLineageRelation(id); err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("删除成功", "data", nil)
}

func (c *Controller) GenerateLifecyclePolicies(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.LifecycleReportRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.GenerateLifecyclePolicies(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("生成成功", "data", data)
}

func (c *Controller) ListMeshGatewayConfigs(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.MeshGatewayListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.ListMeshGatewayConfigs(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateMeshGatewayConfig(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.MeshGatewayUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.CreateMeshGatewayConfig(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("创建成功", "data", data)
}

func (c *Controller) UpdateMeshGatewayConfig(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的网关配置 ID"); return }
	var req reqKafka.MeshGatewayUpsertRequest
	if err = ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.UpdateMeshGatewayConfig(id, req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("更新成功", "data", data)
}

func (c *Controller) DeleteMeshGatewayConfig(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的网关配置 ID"); return }
	if err = c.service.DeleteMeshGatewayConfig(id); err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("删除成功", "data", nil)
}

func (c *Controller) ListCostRecords(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.CostRecordListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.ListCostRecords(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) GenerateCostRecord(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.CostRecordGenerateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.GenerateCostRecord(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("生成成功", "data", data)
}

func (c *Controller) ListSensitiveRules(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.SensitiveRuleListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.ListSensitiveRules(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateSensitiveRule(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.SensitiveRuleUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.CreateSensitiveRule(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("创建成功", "data", data)
}

func (c *Controller) UpdateSensitiveRule(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的规则 ID"); return }
	var req reqKafka.SensitiveRuleUpsertRequest
	if err = ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.UpdateSensitiveRule(id, req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("更新成功", "data", data)
}

func (c *Controller) DeleteSensitiveRule(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil { helper.BadRequest("无效的规则 ID"); return }
	if err = c.service.DeleteSensitiveRule(id); err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("删除成功", "data", nil)
}

func (c *Controller) RunSensitiveScan(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.SensitiveScanRunRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.RunSensitiveScan(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("扫描完成", "data", data)
}

func (c *Controller) ListSensitiveResults(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.SensitiveResultListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil { helper.BadRequest("请求参数错误: " + err.Error()); return }
	data, err := c.service.ListSensitiveResults(req)
	if err != nil { helper.InternalError(err.Error()); return }
	helper.SuccessWithData("查询成功", "data", data)
}
