package kafka

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"

	"github.com/IBM/sarama"
)

func (s *Service) CreateTopic(req reqKafka.CreateTopicRequest) (*response.KafkaTopicCreateVO, error) {
	topic := strings.TrimSpace(req.Name)
	if topic == "" {
		return nil, errors.New("Topic 名称不能为空")
	}
	cluster, err := s.repo.GetByID(req.ClusterID)
	if err != nil {
		return nil, err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	defer admin.Close()

	configEntries := make(map[string]*string, len(req.ConfigEntries))
	for _, entry := range req.ConfigEntries {
		key := strings.TrimSpace(entry.Key)
		if key == "" {
			return nil, errors.New("Topic 配置项 key 不能为空")
		}
		value := entry.Value
		configEntries[key] = &value
	}

	detail := &sarama.TopicDetail{
		NumPartitions:     req.NumPartitions,
		ReplicationFactor: req.ReplicationFactor,
		ConfigEntries:     configEntries,
	}
	if err = admin.CreateTopic(topic, detail, false); err != nil {
		return nil, err
	}

	return &response.KafkaTopicCreateVO{
		Name:              topic,
		Partitions:        req.NumPartitions,
		ReplicationFactor: req.ReplicationFactor,
	}, nil
}

func (s *Service) IncreaseTopicPartitions(topic string, req reqKafka.IncreaseTopicPartitionsRequest) (*response.KafkaTopicPartitionsUpdateVO, error) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return nil, errors.New("Topic 名称不能为空")
	}
	cluster, err := s.repo.GetByID(req.ClusterID)
	if err != nil {
		return nil, err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	defer admin.Close()

	currentPartitions, err := client.Partitions(topic)
	if err != nil {
		return nil, err
	}
	currentCount := int32(len(currentPartitions))
	if req.Count <= currentCount {
		return nil, fmt.Errorf("目标分区数必须大于当前分区数 %d", currentCount)
	}
	if err = admin.CreatePartitions(topic, req.Count, nil, false); err != nil {
		return nil, err
	}

	return &response.KafkaTopicPartitionsUpdateVO{
		Name:               topic,
		PreviousPartitions: currentCount,
		CurrentPartitions:  req.Count,
	}, nil
}

func (s *Service) DescribeTopicPartitions(clusterID uint, topic string) (*response.KafkaTopicPartitionDetailVO, error) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return nil, errors.New("Topic 名称不能为空")
	}
	cluster, err := s.repo.GetByID(clusterID)
	if err != nil {
		return nil, err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	defer admin.Close()

	metadata, err := admin.DescribeTopics([]string{topic})
	if err != nil {
		return nil, err
	}
	if len(metadata) == 0 {
		return nil, fmt.Errorf("未找到 Topic %s", topic)
	}
	if metadata[0].Err != sarama.ErrNoError {
		return nil, metadata[0].Err
	}

	partitions := make([]response.KafkaTopicPartitionVO, 0, len(metadata[0].Partitions))
	underReplicatedCount := 0
	for _, partition := range metadata[0].Partitions {
		if partition == nil {
			continue
		}
		oldestOffset, oldestErr := client.GetOffset(topic, partition.ID, sarama.OffsetOldest)
		if oldestErr != nil {
			oldestOffset = -1
		}
		latestOffset, latestErr := client.GetOffset(topic, partition.ID, sarama.OffsetNewest)
		if latestErr != nil {
			latestOffset = -1
		}
		outOfSyncReplicas := diffInt32Slice(partition.Replicas, partition.Isr)
		underReplicated := len(outOfSyncReplicas) > 0 || len(partition.OfflineReplicas) > 0
		if underReplicated {
			underReplicatedCount++
		}
		messageCountEstimate := int64(0)
		if oldestOffset >= 0 && latestOffset >= oldestOffset {
			messageCountEstimate = latestOffset - oldestOffset
		}
		partitions = append(partitions, response.KafkaTopicPartitionVO{
			Partition:            partition.ID,
			Leader:               partition.Leader,
			Replicas:             append([]int32(nil), partition.Replicas...),
			ISR:                  append([]int32(nil), partition.Isr...),
			OfflineReplicas:      append([]int32(nil), partition.OfflineReplicas...),
			OutOfSyncReplicas:    outOfSyncReplicas,
			UnderReplicated:      underReplicated,
			OldestOffset:         oldestOffset,
			LatestOffset:         latestOffset,
			MessageCountEstimate: messageCountEstimate,
		})
	}
	sort.Slice(partitions, func(i, j int) bool { return partitions[i].Partition < partitions[j].Partition })

	return &response.KafkaTopicPartitionDetailVO{
		Topic:                topic,
		PartitionCount:       len(partitions),
		UnderReplicatedCount: underReplicatedCount,
		Partitions:           partitions,
	}, nil
}

func diffInt32Slice(full []int32, subset []int32) []int32 {
	subsetSet := make(map[int32]struct{}, len(subset))
	for _, value := range subset {
		subsetSet[value] = struct{}{}
	}
	result := make([]int32, 0, len(full))
	for _, value := range full {
		if _, ok := subsetSet[value]; ok {
			continue
		}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
