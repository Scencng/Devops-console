package kafka

import (
	"errors"
	"time"

	"devops-console-backend/internal/dal"
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/configs"
)

func (s *Service) ListChangeRequests(req reqKafka.ChangeRequestListRequest) ([]response.KafkaChangeRequestVO, error) {
	list, err := configs.NewKafkaChangeRequestRepository().List(req.ClusterID, req.Status)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaChangeRequestVO, 0, len(list))
	for _, item := range list {
		result = append(result, toChangeRequestVO(item))
	}
	return result, nil
}

func (s *Service) CreateChangeRequest(req reqKafka.ChangeRequestCreateRequest, requesterUserID uint64, requesterUsername string) (*response.KafkaChangeRequestVO, error) {
	cluster, err := s.repo.GetByID(req.ClusterID)
	if err != nil {
		return nil, err
	}
	item := &dal.KafkaChangeRequest{
		ClusterID:         req.ClusterID,
		ChangeType:        req.ChangeType,
		ResourceType:      req.ResourceType,
		ResourceName:      req.ResourceName,
		Payload:           req.Payload,
		Reason:            req.Reason,
		Status:            "pending",
		RequesterUserID:   requesterUserID,
		RequesterUsername: requesterUsername,
		Environment:       cluster.Environment,
		Tenant:            cluster.Tenant,
	}
	if err = configs.NewKafkaChangeRequestRepository().Create(item); err != nil {
		return nil, err
	}
	vo := toChangeRequestVO(*item)
	return &vo, nil
}

func (s *Service) ReviewChangeRequest(id uint, req reqKafka.ChangeRequestReviewRequest, approverUserID uint64, approverUsername string) (*response.KafkaChangeRequestVO, error) {
	repo := configs.NewKafkaChangeRequestRepository()
	item, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	item.Status = req.Status
	item.ApproverUserID = approverUserID
	item.ApproverUsername = approverUsername
	item.ApprovalComment = req.Comment
	if err = repo.Update(item); err != nil {
		return nil, err
	}
	vo := toChangeRequestVO(*item)
	return &vo, nil
}

func (s *Service) ExecuteChangeRequest(id uint) (*response.KafkaChangeRequestVO, error) {
	repo := configs.NewKafkaChangeRequestRepository()
	item, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if item.Status != "approved" {
		return nil, errors.New("只有已审批通过的变更单才能执行")
	}
	if err = s.executeTask(item.ChangeType, item.Payload, item.ClusterID, item.ApproverUsername); err != nil {
		return nil, err
	}
	now := time.Now()
	item.Status = "executed"
	item.ExecutedAt = &now
	if err = repo.Update(item); err != nil {
		return nil, err
	}
	vo := toChangeRequestVO(*item)
	return &vo, nil
}

func toChangeRequestVO(item dal.KafkaChangeRequest) response.KafkaChangeRequestVO {
	return response.KafkaChangeRequestVO{
		ID:                item.ID,
		ClusterID:         item.ClusterID,
		ChangeType:        item.ChangeType,
		ResourceType:      item.ResourceType,
		ResourceName:      item.ResourceName,
		Payload:           item.Payload,
		Reason:            item.Reason,
		Status:            item.Status,
		RequesterUserID:   item.RequesterUserID,
		RequesterUsername: item.RequesterUsername,
		ApproverUserID:    item.ApproverUserID,
		ApproverUsername:  item.ApproverUsername,
		ApprovalComment:   item.ApprovalComment,
		Environment:       item.Environment,
		Tenant:            item.Tenant,
		ExecutedAt:        item.ExecutedAt,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}
