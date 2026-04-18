package kafka

import (
	"devops-console-backend/internal/dal"
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/configs"
)

func (s *Service) ListAuditLogs(req reqKafka.AuditLogListRequest) (*response.KafkaAuditLogListVO, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	repo := configs.NewKafkaAuditLogRepository()
	list, total, err := repo.List(req.ClusterID, req.Action, req.Result, page, pageSize)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaAuditLogVO, 0, len(list))
	for _, item := range list {
		result = append(result, toAuditVO(item))
	}
	return &response.KafkaAuditLogListVO{Total: total, List: result}, nil
}

func toAuditVO(item dal.KafkaAuditLog) response.KafkaAuditLogVO {
	return response.KafkaAuditLogVO{ID: item.ID, ClusterID: item.ClusterID, Action: item.Action, ResourceType: item.ResourceType, ResourceName: item.ResourceName, OperatorUserID: item.OperatorUserID, OperatorUsername: item.OperatorUsername, RequestPayload: item.RequestPayload, Result: item.Result, ErrorMessage: item.ErrorMessage, CreatedAt: item.CreatedAt}
}
