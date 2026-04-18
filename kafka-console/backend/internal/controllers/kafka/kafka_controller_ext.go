package kafka

import (
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (c *Controller) CreateTopic(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.CreateTopicRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.CreateTopic(req)
	if err != nil {
		c.writeAuditLog(ctx, req.ClusterID, "topic:create", "topic", req.Name, req, "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "topic:create", "topic", data.Name, req, "success", "")
	helper.SuccessWithData("Topic 创建成功", "data", data)
}

func (c *Controller) IncreaseTopicPartitions(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.IncreaseTopicPartitionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	topic := ctx.Param("topic")
	data, err := c.service.IncreaseTopicPartitions(topic, req)
	if err != nil {
		c.writeAuditLog(ctx, req.ClusterID, "topic:partitions:increase", "topic", topic, req, "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "topic:partitions:increase", "topic", topic, req, "success", "")
	helper.SuccessWithData("Topic 分区扩容成功", "data", data)
}

func (c *Controller) GetTopicPartitions(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.TopicPartitionsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	topic := ctx.Param("topic")
	data, err := c.service.DescribeTopicPartitions(req.ClusterID, topic)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) GetConsumerGroupDetail(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ConsumerGroupDetailRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	groupID := ctx.Param("groupId")
	data, err := c.service.GetConsumerGroupDetail(groupID, req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) ProduceMessage(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.MessageProduceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ProduceMessage(req)
	if err != nil {
		c.writeAuditLog(ctx, req.ClusterID, "message:produce", "topic", req.Topic, sanitizeProduceMessagePayload(req), "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "message:produce", "topic", req.Topic, sanitizeProduceMessagePayload(req), "success", "")
	helper.SuccessWithData("消息发送成功", "data", data)
}

func (c *Controller) ListACLs(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ACLListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListACLs(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) CreateACL(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ACLUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.CreateACL(req)
	if err != nil {
		c.writeAuditLog(ctx, req.ClusterID, "acl:create", "acl", req.ResourceName, req, "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "acl:create", "acl", req.ResourceName, req, "success", "")
	helper.SuccessWithData("ACL 创建成功", "data", data)
}

func (c *Controller) DeleteACL(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ACLDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.DeleteACL(req)
	if err != nil {
		c.writeAuditLog(ctx, req.ClusterID, "acl:delete", "acl", req.ResourceName, req, "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "acl:delete", "acl", req.ResourceName, req, "success", "")
	helper.SuccessWithData("ACL 删除成功", "data", data)
}

func (c *Controller) ListScramUsers(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ScramUserListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	data, err := c.service.ListScramUsers(req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", data)
}

func (c *Controller) UpsertScramUser(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ScramUserUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	if err := c.service.UpsertScramUser(req); err != nil {
		c.writeAuditLog(ctx, req.ClusterID, "scram:user:upsert", "scram_user", req.Username, sanitizeScramUserPayload(req), "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "scram:user:upsert", "scram_user", req.Username, sanitizeScramUserPayload(req), "success", "")
	helper.SuccessWithData("SCRAM 用户已保存", "data", nil)
}

func (c *Controller) DeleteScramUser(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqKafka.ScramUserDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}
	if err := c.service.DeleteScramUser(req); err != nil {
		c.writeAuditLog(ctx, req.ClusterID, "scram:user:delete", "scram_user", req.Username, req, "failed", err.Error())
		helper.InternalError(err.Error())
		return
	}
	c.writeAuditLog(ctx, req.ClusterID, "scram:user:delete", "scram_user", req.Username, req, "success", "")
	helper.SuccessWithData("SCRAM 用户已删除", "data", nil)
}

func sanitizeClusterPayload(req reqKafka.ClusterUpsertRequest) map[string]interface{} {
	return map[string]interface{}{
		"name":               req.Name,
		"bootstrapServers":   req.BootstrapServers,
		"version":            req.Version,
		"authType":           req.AuthType,
		"username":           req.Username,
		"tlsEnabled":         req.TLSEnabled,
		"insecureSkipVerify": req.InsecureSkipVerify,
		"caCert":             req.CACert,
		"clientCert":         req.ClientCert,
		"description":        req.Description,
		"environment":        req.Environment,
		"tenant":             req.Tenant,
		"hasPassword":        req.Password != "",
		"hasClientKey":       req.ClientKey != "",
	}
}

func sanitizeResetOffsetPayload(req reqKafka.ResetConsumerGroupOffsetRequest) map[string]interface{} {
	return map[string]interface{}{
		"topic":         req.Topic,
		"partition":     req.Partition,
		"allPartitions": req.AllPartitions,
		"resetType":     req.ResetType,
		"offset":        req.Offset,
		"timestampMs":   req.TimestampMs,
	}
}

func sanitizeProduceMessagePayload(req reqKafka.MessageProduceRequest) map[string]interface{} {
	return map[string]interface{}{
		"topic":         req.Topic,
		"partition":     req.Partition,
		"keyEncoding":   req.KeyEncoding,
		"valueEncoding": req.ValueEncoding,
		"headerCount":   len(req.Headers),
		"hasKey":        req.Key != "",
		"valueBytes":    len(req.Value),
	}
}

func sanitizeScramUserPayload(req reqKafka.ScramUserUpsertRequest) map[string]interface{} {
	return map[string]interface{}{
		"username":   req.Username,
		"mechanism":  req.Mechanism,
		"iterations": req.Iterations,
	}
}
