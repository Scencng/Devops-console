package kafka

import (
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (c *Controller) GetHealthOverview(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.HealthOverviewRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.GetHealthOverview(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) GetTrendSeries(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.TrendRangeRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.GetTrendSeries(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) ListAlertRules(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.AlertRuleListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListAlertRules(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateAlertRule(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.AlertRuleUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.CreateAlertRule(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "alert:rule:create", "alert_rule", req.Name, req, "success", "")
	helper.SuccessWithData("创建成功", "data", data)
}

func (c *Controller) UpdateAlertRule(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的规则 ID")
		return
	}
	var req reqKafka.AlertRuleUpsertRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.UpdateAlertRule(id, req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "alert:rule:update", "alert_rule", req.Name, req, "success", "")
	helper.SuccessWithData("更新成功", "data", data)
}

func (c *Controller) DeleteAlertRule(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的规则 ID")
		return
	}
	if err = c.service.DeleteAlertRule(id); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("删除成功", "data", nil)
}

func (c *Controller) EvaluateAlertRules(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.HealthOverviewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.EvaluateAlertRules(req.ClusterID)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "alert:evaluate", "alert_rule", "cluster", req, "success", "")
	helper.SuccessWithData("评估完成", "data", data)
}

func (c *Controller) ListAlertEvents(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.AlertEventListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListAlertEvents(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) UpdateAlertEventStatus(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的告警 ID")
		return
	}
	var req reqKafka.AlertEventStatusRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.UpdateAlertEventStatus(id, req.Status, utils.GetUserNameFromContext(ctx))
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("状态更新成功", "data", data)
}

func (c *Controller) RunInspection(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.InspectionRunRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.RunInspection(req, utils.GetUserNameFromContext(ctx))
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "inspection:run", "inspection", data.Name, req, "success", "")
	helper.SuccessWithData("巡检完成", "data", data)
}

func (c *Controller) ListInspectionReports(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.InspectionListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListInspectionReports(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) GetInspectionReport(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的报告 ID")
		return
	}
	data, err := c.service.GetInspectionReport(id)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) ListTasks(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.TaskListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListTasks(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateTask(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.TaskUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.CreateTask(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "task:create", "task", req.Name, req, "success", "")
	helper.SuccessWithData("创建成功", "data", data)
}

func (c *Controller) UpdateTask(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的任务 ID")
		return
	}
	var req reqKafka.TaskUpsertRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.UpdateTask(id, req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "task:update", "task", req.Name, req, "success", "")
	helper.SuccessWithData("更新成功", "data", data)
}

func (c *Controller) DeleteTask(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的任务 ID")
		return
	}
	if err = c.service.DeleteTask(id); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("删除成功", "data", nil)
}

func (c *Controller) RunTask(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的任务 ID")
		return
	}
	var req reqKafka.TaskRunRequest
	_ = ctx.ShouldBindJSON(&req)
	data, err := c.service.RunTask(id, req.TriggerMode, utils.GetUserNameFromContext(ctx))
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("任务执行完成", "data", data)
}

func (c *Controller) ListTaskRuns(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.TaskRunListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListTaskRuns(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) ListChangeRequests(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ChangeRequestListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListChangeRequests(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateChangeRequest(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ChangeRequestCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.CreateChangeRequest(req, uint64(utils.GetUserIdFromContext(ctx)), utils.GetUserNameFromContext(ctx))
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "change_request:create", req.ResourceType, req.ResourceName, req, "success", "")
	helper.SuccessWithData("变更单已创建", "data", data)
}

func (c *Controller) ReviewChangeRequest(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的变更单 ID")
		return
	}
	var req reqKafka.ChangeRequestReviewRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ReviewChangeRequest(id, req, uint64(utils.GetUserIdFromContext(ctx)), utils.GetUserNameFromContext(ctx))
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("审批完成", "data", data)
}

func (c *Controller) ExecuteChangeRequest(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的变更单 ID")
		return
	}
	data, err := c.service.ExecuteChangeRequest(id)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("变更执行完成", "data", data)
}

func (c *Controller) ListTopicMetadata(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.TopicMetadataListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListTopicMetadata(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateTopicMetadata(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.TopicMetadataUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.CreateTopicMetadata(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("创建成功", "data", data)
}

func (c *Controller) UpdateTopicMetadata(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的元数据 ID")
		return
	}
	var req reqKafka.TopicMetadataUpsertRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.UpdateTopicMetadata(id, req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("更新成功", "data", data)
}

func (c *Controller) DeleteTopicMetadata(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的元数据 ID")
		return
	}
	if err = c.service.DeleteTopicMetadata(id); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("删除成功", "data", nil)
}

func (c *Controller) ListSchemaRegistries(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.SchemaRegistryListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListSchemaRegistries(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateSchemaRegistry(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.SchemaRegistryUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.CreateSchemaRegistry(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("创建成功", "data", data)
}

func (c *Controller) UpdateSchemaRegistry(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的 Schema Registry ID")
		return
	}
	var req reqKafka.SchemaRegistryUpsertRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.UpdateSchemaRegistry(id, req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("更新成功", "data", data)
}

func (c *Controller) DeleteSchemaRegistry(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的 Schema Registry ID")
		return
	}
	if err = c.service.DeleteSchemaRegistry(id); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("删除成功", "data", nil)
}

func (c *Controller) ListSchemaSubjects(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.SchemaSubjectListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListSchemaSubjects(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) ListSchemaVersions(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.SchemaVersionListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListSchemaVersions(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) GetSchemaDetail(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.SchemaDetailRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.GetSchemaDetail(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CheckSchemaCompatibility(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.SchemaCompatibilityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.CheckSchemaCompatibility(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("校验完成", "data", data)
}
