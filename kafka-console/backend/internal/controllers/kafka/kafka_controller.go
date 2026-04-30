package kafka

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"time"

	"devops-console-backend/internal/dal"
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	serviceKafka "devops-console-backend/internal/services/kafka"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	logutil "devops-console-backend/pkg/utils/logs"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	service   *serviceKafka.Service
	auditRepo *configs.KafkaAuditLogRepository
}

func NewController() *Controller {
	auditRepo := configs.NewKafkaAuditLogRepository()
	return &Controller{service: serviceKafka.NewService(configs.NewKafkaClusterRepository(), auditRepo), auditRepo: auditRepo}
}

func (c *Controller) ListClusters(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ClusterListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListClusters(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) ListClusterOptions(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	data, err := c.service.ListClusterOptions()
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) GetCluster(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的集群 ID")
		return
	}
	data, err := c.service.GetCluster(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.NotFound("Kafka 集群不存在")
			return
		}
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateCluster(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ClusterUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.CreateCluster(req)
	if err != nil {
		c.writeAuditLog(ctx, 0, "cluster:create", "cluster", req.Name, sanitizeClusterPayload(req), "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, data.ID, "cluster:create", "cluster", data.Name, sanitizeClusterPayload(req), "success", "")
	helper.SuccessWithData("创建成功", "data", data)
}

func (c *Controller) UpdateCluster(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的集群 ID")
		return
	}
	before, beforeErr := c.service.GetCluster(id)
	if beforeErr != nil && !errors.Is(beforeErr, gorm.ErrRecordNotFound) {
		helper.InternalError(beforeErr.Error())
		return
	}
	var req reqKafka.ClusterUpsertRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.UpdateCluster(id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.NotFound("Kafka 集群不存在")
			return
		}
		resourceName := req.Name
		if before != nil {
			resourceName = before.Name
		}
		c.writeAuditLog(ctx, id, "cluster:update", "cluster", resourceName, sanitizeClusterPayload(req), "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, id, "cluster:update", "cluster", data.Name, sanitizeClusterPayload(req), "success", "")
	helper.SuccessWithData("更新成功", "data", data)
}

func (c *Controller) DeleteCluster(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的集群 ID")
		return
	}
	before, beforeErr := c.service.GetCluster(id)
	if beforeErr != nil && !errors.Is(beforeErr, gorm.ErrRecordNotFound) {
		helper.InternalError(beforeErr.Error())
		return
	}
	resourceName := strconv.FormatUint(uint64(id), 10)
	if before != nil {
		resourceName = before.Name
	}
	if err = c.service.DeleteCluster(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.NotFound("Kafka 集群不存在")
			return
		}
		c.writeAuditLog(ctx, id, "cluster:delete", "cluster", resourceName, map[string]interface{}{"id": id}, "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, id, "cluster:delete", "cluster", resourceName, map[string]interface{}{"id": id}, "success", "")
	helper.SuccessWithData("删除成功", "data", nil)
}

func (c *Controller) TestClusterConnection(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的集群 ID")
		return
	}
	data, err := c.service.TestCluster(id)
	if err != nil {
		if data != nil {
			helper.SuccessWithData("连接测试完成", "data", data)
			return
		}
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("连接测试完成", "data", data)
}

func (c *Controller) ListTopics(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.TopicListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListTopics(req.ClusterID, req.Keyword)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) DeleteTopic(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.TopicActionRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	topic := ctx.Param("topic")
	if err := c.service.DeleteTopic(req.ClusterID, topic); err != nil {
		c.writeAuditLog(ctx, req.ClusterID, "topic:delete", "topic", topic, req, "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "topic:delete", "topic", topic, req, "success", "")
	helper.SuccessWithData("Topic 删除成功", "data", nil)
}

func (c *Controller) UpdateTopicConfig(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.TopicConfigUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	topic := ctx.Param("topic")
	if err := c.service.UpdateTopicConfig(topic, req); err != nil {
		c.writeAuditLog(ctx, req.ClusterID, "topic:config:update", "topic", topic, req, "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "topic:config:update", "topic", topic, req, "success", "")
	helper.SuccessWithData("Topic 配置更新成功", "data", nil)
}

func (c *Controller) UpdateBrokerConfig(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ClusterQueryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

	id, err := parseIDParam(ctx, "id")
	if err != nil {
		helper.BadRequest("无效的 Broker ID")
		return
	}
	brokerID := int32(id)
	if uint(brokerID) != id {
		helper.BadRequest("无效的 Broker ID")
		return
	}

	configs := make(map[string]string)
	if err := ctx.ShouldBindJSON(&configs); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	if len(configs) == 0 {
		helper.BadRequest("至少需要提供一项 Broker 配置")
		return
	}

	if err := c.service.UpdateBrokerConfig(req.ClusterID, brokerID, configs); err != nil {
		c.writeAuditLog(ctx, req.ClusterID, "broker:config:update", "broker", strconv.Itoa(int(brokerID)), sanitizeBrokerConfigPayload(req.ClusterID, configs), "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}

	c.writeAuditLog(ctx, req.ClusterID, "broker:config:update", "broker", strconv.Itoa(int(brokerID)), sanitizeBrokerConfigPayload(req.ClusterID, configs), "success", "")
	helper.SuccessWithData("Broker 动态配置更新成功", "data", nil)
}

func (c *Controller) ListBrokers(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.BrokerListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListBrokers(req.ClusterID)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) ListConsumerGroups(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ConsumerGroupListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListConsumerGroups(req.ClusterID, req.Keyword)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) ResetConsumerGroupOffset(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ResetConsumerGroupOffsetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	groupID := ctx.Param("groupId")
	result, err := c.service.ResetConsumerGroupOffset(groupID, req)
	if err != nil {
		c.writeAuditLog(ctx, req.ClusterID, "group:offset:reset", "consumer_group", groupID, sanitizeResetOffsetPayload(req), "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "group:offset:reset", "consumer_group", groupID, sanitizeResetOffsetPayload(req), "success", "")
	helper.SuccessWithData("消费组 Offset 重置成功", "data", result)
}

func (c *Controller) BrowseMessages(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.MessageBrowseRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.BrowseMessages(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) ListAuditLogs(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.AuditLogListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListAuditLogs(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) GetDashboard(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ClusterQueryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.GetDashboard(req.ClusterID)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) writeAuditLog(ctx *gin.Context, clusterID uint, action, resourceType, resourceName string, payload interface{}, result, errorMessage string) {
	payloadBytes, _ := json.Marshal(payload)
	if err := c.auditRepo.Create(&dal.KafkaAuditLog{ClusterID: clusterID, Action: action, ResourceType: resourceType, ResourceName: resourceName, OperatorUserID: uint64(utils.GetUserIdFromContext(ctx)), OperatorUsername: utils.GetUserNameFromContext(ctx), RequestPayload: string(payloadBytes), Result: result, ErrorMessage: errorMessage, CreatedAt: time.Now()}); err != nil {
		logutil.Error(map[string]interface{}{"cluster_id": clusterID, "action": action, "resource_type": resourceType, "resource_name": resourceName, "error": err.Error()}, "写入 Kafka 审计日志失败")
	}
}

func parseIDParam(ctx *gin.Context, name string) (uint, error) {
	value := ctx.Param(name)
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func sanitizeBrokerConfigPayload(clusterID uint, configs map[string]string) map[string]interface{} {
	keys := make([]string, 0, len(configs))
	for key := range configs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return map[string]interface{}{
		"clusterId":   clusterID,
		"configCount": len(configs),
		"configKeys":  keys,
	}
}
