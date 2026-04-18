package kafka

import (
	"errors"

	"devops-console-backend/internal/dal"
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/configs"
)

func (s *Service) ListTopicMetadata(req reqKafka.TopicMetadataListRequest) ([]response.KafkaTopicMetadataVO, error) {
	list, err := configs.NewKafkaTopicMetadataRepository().List(req.ClusterID, req.Keyword)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaTopicMetadataVO, 0, len(list))
	for _, item := range list {
		result = append(result, toTopicMetadataVO(item))
	}
	return result, nil
}

func (s *Service) CreateTopicMetadata(req reqKafka.TopicMetadataUpsertRequest) (*response.KafkaTopicMetadataVO, error) {
	repo := configs.NewKafkaTopicMetadataRepository()
	if _, err := repo.GetByClusterTopic(req.ClusterID, req.TopicName); err == nil {
		return nil, errors.New("该 Topic 的元数据已存在")
	}
	item := &dal.KafkaTopicMetadata{
		ClusterID:   req.ClusterID,
		TopicName:   req.TopicName,
		SystemName:  req.SystemName,
		Owner:       req.Owner,
		OwnerEmail:  req.OwnerEmail,
		Environment: req.Environment,
		Tenant:      req.Tenant,
		Lifecycle:   req.Lifecycle,
		Sensitivity: req.Sensitivity,
		Description: req.Description,
		Labels:      req.Labels,
	}
	if err := repo.Create(item); err != nil {
		return nil, err
	}
	vo := toTopicMetadataVO(*item)
	return &vo, nil
}

func (s *Service) UpdateTopicMetadata(id uint, req reqKafka.TopicMetadataUpsertRequest) (*response.KafkaTopicMetadataVO, error) {
	repo := configs.NewKafkaTopicMetadataRepository()
	item, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	item.ClusterID = req.ClusterID
	item.TopicName = req.TopicName
	item.SystemName = req.SystemName
	item.Owner = req.Owner
	item.OwnerEmail = req.OwnerEmail
	item.Environment = req.Environment
	item.Tenant = req.Tenant
	item.Lifecycle = req.Lifecycle
	item.Sensitivity = req.Sensitivity
	item.Description = req.Description
	item.Labels = req.Labels
	if err = repo.Update(item); err != nil {
		return nil, err
	}
	vo := toTopicMetadataVO(*item)
	return &vo, nil
}

func (s *Service) DeleteTopicMetadata(id uint) error {
	return configs.NewKafkaTopicMetadataRepository().Delete(id)
}

func toTopicMetadataVO(item dal.KafkaTopicMetadata) response.KafkaTopicMetadataVO {
	return response.KafkaTopicMetadataVO{
		ID:          item.ID,
		ClusterID:   item.ClusterID,
		TopicName:   item.TopicName,
		SystemName:  item.SystemName,
		Owner:       item.Owner,
		OwnerEmail:  item.OwnerEmail,
		Environment: item.Environment,
		Tenant:      item.Tenant,
		Lifecycle:   item.Lifecycle,
		Sensitivity: item.Sensitivity,
		Description: item.Description,
		Labels:      item.Labels,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
