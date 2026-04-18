package kafka

import (
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"devops-console-backend/internal/dal"
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/configs"
	cryptoutil "devops-console-backend/pkg/utils/crypto"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"
	"gorm.io/gorm"
)

type Service struct {
	repo *configs.KafkaClusterRepository
}

func NewService(repo *configs.KafkaClusterRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListClusters(req reqKafka.ClusterListRequest) (*response.KafkaClusterListVO, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	list, total, err := s.repo.List(page, pageSize, req.Keyword, req.Status, req.Environment, req.Tenant)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaClusterVO, 0, len(list))
	for _, item := range list {
		result = append(result, toClusterVO(item))
	}
	return &response.KafkaClusterListVO{Total: total, List: result}, nil
}

func (s *Service) ListClusterOptions() ([]response.KafkaClusterOptionVO, error) {
	list, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	options := make([]response.KafkaClusterOptionVO, 0, len(list))
	for _, item := range list {
		options = append(options, response.KafkaClusterOptionVO{ID: item.ID, Name: item.Name, Status: item.Status})
	}
	return options, nil
}

func (s *Service) GetCluster(id uint) (*response.KafkaClusterVO, error) {
	cluster, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	vo := toClusterVO(*cluster)
	return &vo, nil
}

func (s *Service) CreateCluster(req reqKafka.ClusterUpsertRequest) (*response.KafkaClusterVO, error) {
	if _, err := s.repo.GetByName(req.Name); err == nil {
		return nil, errors.New("Kafka 集群名称已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	cluster, err := buildClusterModel(nil, req)
	if err != nil {
		return nil, err
	}
	if err = s.repo.Create(cluster); err != nil {
		return nil, err
	}
	vo := toClusterVO(*cluster)
	return &vo, nil
}

func (s *Service) UpdateCluster(id uint, req reqKafka.ClusterUpsertRequest) (*response.KafkaClusterVO, error) {
	cluster, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing, err := s.repo.GetByName(req.Name); err == nil && existing.ID != id {
		return nil, errors.New("Kafka 集群名称已存在")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	cluster, err = buildClusterModel(cluster, req)
	if err != nil {
		return nil, err
	}
	if err = s.repo.Update(cluster); err != nil {
		return nil, err
	}
	vo := toClusterVO(*cluster)
	return &vo, nil
}

func (s *Service) DeleteCluster(id uint) error {
	return s.repo.Delete(id)
}

func (s *Service) TestCluster(id uint) (*response.KafkaClusterTestVO, error) {
	cluster, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	client, admin, err := s.openClients(cluster)
	testedAt := time.Now()
	if err != nil {
		_ = s.repo.UpdateTestStatus(id, dal.KafkaClusterStatusError, err.Error(), &testedAt)
		_ = s.repo.SaveTestRecord(&dal.ConnectionTest{ResourceType: dal.KafkaResourceTypeCluster, ResourceID: id, TestResult: "failure", ErrorMessage: err.Error(), TestedAt: testedAt})
		return &response.KafkaClusterTestVO{ClusterID: id, ClusterName: cluster.Name, ResponseTime: time.Since(start).Milliseconds(), TestedAt: testedAt, Status: dal.KafkaClusterStatusError, ErrorMessage: err.Error()}, err
	}
	defer admin.Close()
	defer client.Close()
	brokers, controllerID, err := admin.DescribeCluster()
	if err != nil {
		_ = s.repo.UpdateTestStatus(id, dal.KafkaClusterStatusError, err.Error(), &testedAt)
		return nil, err
	}
	responseTime := time.Since(start).Milliseconds()
	_ = s.repo.UpdateTestStatus(id, dal.KafkaClusterStatusActive, "", &testedAt)
	responseTimeInt := int(responseTime)
	_ = s.repo.SaveTestRecord(&dal.ConnectionTest{ResourceType: dal.KafkaResourceTypeCluster, ResourceID: id, TestResult: "success", ResponseTime: &responseTimeInt, TestedAt: testedAt})
	return &response.KafkaClusterTestVO{ClusterID: id, ClusterName: cluster.Name, BrokerCount: len(brokers), ControllerID: controllerID, ResponseTime: responseTime, TestedAt: testedAt, Status: dal.KafkaClusterStatusActive}, nil
}

func (s *Service) ListTopics(clusterID uint, keyword string) ([]response.KafkaTopicVO, error) {
	cluster, err := s.repo.GetByID(clusterID)
	if err != nil {
		return nil, err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return nil, err
	}
	defer admin.Close()
	defer client.Close()
	topicDetails, err := admin.ListTopics()
	if err != nil {
		return nil, err
	}
	metadataTopics, err := client.Topics()
	if err != nil {
		return nil, err
	}
	internalTopics := make(map[string]bool, len(metadataTopics))
	for _, topic := range metadataTopics {
		internalTopics[topic] = strings.HasPrefix(topic, "__")
	}
	result := make([]response.KafkaTopicVO, 0, len(topicDetails))
	for name, detail := range topicDetails {
		if keyword != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(keyword)) {
			continue
		}
		configEntries := make(map[string]string, len(detail.ConfigEntries))
		for key, value := range detail.ConfigEntries {
			if value != nil {
				configEntries[key] = *value
			}
		}
		result = append(result, response.KafkaTopicVO{Name: name, Partitions: detail.NumPartitions, ReplicationFactor: detail.ReplicationFactor, Internal: internalTopics[name], CleanupPolicy: configEntries["cleanup.policy"], RetentionMs: configEntries["retention.ms"], MinInSyncReplicas: configEntries["min.insync.replicas"], ConfigEntries: configEntries})
	}
	return result, nil
}

func (s *Service) DeleteTopic(clusterID uint, topic string) error {
	if strings.TrimSpace(topic) == "" {
		return errors.New("Topic 名称不能为空")
	}
	if strings.HasPrefix(topic, "__") {
		return errors.New("不允许删除 Kafka 内部 Topic")
	}
	cluster, err := s.repo.GetByID(clusterID)
	if err != nil {
		return err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return err
	}
	defer admin.Close()
	defer client.Close()
	return admin.DeleteTopic(topic)
}

func (s *Service) UpdateTopicConfig(topic string, req reqKafka.TopicConfigUpdateRequest) error {
	if strings.TrimSpace(topic) == "" {
		return errors.New("Topic 名称不能为空")
	}
	cluster, err := s.repo.GetByID(req.ClusterID)
	if err != nil {
		return err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return err
	}
	defer admin.Close()
	defer client.Close()
	version, err := parseKafkaVersion(cluster.Version)
	if err != nil {
		return err
	}
	incrementalEntries := make(map[string]sarama.IncrementalAlterConfigsEntry, len(req.Entries))
	alterEntries := make(map[string]*string, len(req.Entries))
	for _, entry := range req.Entries {
		key := strings.TrimSpace(entry.Key)
		if key == "" {
			return errors.New("Topic 配置项 key 不能为空")
		}
		switch entry.Operation {
		case "set":
			if entry.Value == nil {
				return fmt.Errorf("Topic 配置项 %s 缺少 value", key)
			}
			value := *entry.Value
			incrementalEntries[key] = sarama.IncrementalAlterConfigsEntry{Operation: sarama.IncrementalAlterConfigsOperationSet, Value: &value}
			alterEntries[key] = &value
		case "delete":
			if !version.IsAtLeast(sarama.V2_3_0_0) {
				return errors.New("当前 Kafka 版本低于 2.3.0，不支持删除 Topic 配置项")
			}
			incrementalEntries[key] = sarama.IncrementalAlterConfigsEntry{Operation: sarama.IncrementalAlterConfigsOperationDelete}
		default:
			return fmt.Errorf("不支持的 Topic 配置操作: %s", entry.Operation)
		}
	}
	if version.IsAtLeast(sarama.V2_3_0_0) {
		return admin.IncrementalAlterConfig(sarama.TopicResource, topic, incrementalEntries, false)
	}
	return admin.AlterConfig(sarama.TopicResource, topic, alterEntries, false)
}

func (s *Service) ListBrokers(clusterID uint) ([]response.KafkaBrokerVO, error) {
	cluster, err := s.repo.GetByID(clusterID)
	if err != nil {
		return nil, err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return nil, err
	}
	defer admin.Close()
	defer client.Close()
	brokers, controllerID, err := admin.DescribeCluster()
	if err != nil {
		return nil, err
	}
	leaderCounts := map[int32]int{}
	replicaCounts := map[int32]int{}
	brokerTopics := map[int32]map[string]struct{}{}
	topics, err := client.Topics()
	if err == nil {
		for _, topic := range topics {
			partitions, partErr := client.Partitions(topic)
			if partErr != nil {
				continue
			}
			for _, partition := range partitions {
				leader, leaderErr := client.Leader(topic, partition)
				if leaderErr == nil && leader != nil {
					leaderCounts[leader.ID()]++
					if brokerTopics[leader.ID()] == nil {
						brokerTopics[leader.ID()] = map[string]struct{}{}
					}
					brokerTopics[leader.ID()][topic] = struct{}{}
				}
				replicas, replicaErr := client.Replicas(topic, partition)
				if replicaErr == nil {
					for _, brokerID := range replicas {
						replicaCounts[brokerID]++
						if brokerTopics[brokerID] == nil {
							brokerTopics[brokerID] = map[string]struct{}{}
						}
						brokerTopics[brokerID][topic] = struct{}{}
					}
				}
			}
		}
	}
	result := make([]response.KafkaBrokerVO, 0, len(brokers))
	for _, broker := range brokers {
		connected, _ := broker.Connected()
		brokerTopicList := make([]string, 0, len(brokerTopics[broker.ID()]))
		for topic := range brokerTopics[broker.ID()] {
			brokerTopicList = append(brokerTopicList, topic)
		}
		sort.Strings(brokerTopicList)
		result = append(result, response.KafkaBrokerVO{ID: broker.ID(), Address: broker.Addr(), IsController: broker.ID() == controllerID, Connected: connected, LeaderPartitionCount: leaderCounts[broker.ID()], ReplicaPartitionCount: replicaCounts[broker.ID()], Topics: brokerTopicList})
	}
	return result, nil
}

func (s *Service) ListConsumerGroups(clusterID uint, keyword string) ([]response.KafkaConsumerGroupVO, error) {
	cluster, err := s.repo.GetByID(clusterID)
	if err != nil {
		return nil, err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return nil, err
	}
	defer admin.Close()
	defer client.Close()
	groupMap, err := admin.ListConsumerGroups()
	if err != nil {
		return nil, err
	}
	groupIDs := make([]string, 0, len(groupMap))
	for groupID := range groupMap {
		if keyword != "" && !strings.Contains(strings.ToLower(groupID), strings.ToLower(keyword)) {
			continue
		}
		groupIDs = append(groupIDs, groupID)
	}
	if len(groupIDs) == 0 {
		return []response.KafkaConsumerGroupVO{}, nil
	}
	groups, err := admin.DescribeConsumerGroups(groupIDs)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaConsumerGroupVO, 0, len(groups))
	for _, group := range groups {
		offsets, err := admin.ListConsumerGroupOffsets(group.GroupId, nil)
		if err != nil {
			return nil, err
		}
		topics := map[string]struct{}{}
		partitionCount := 0
		var totalLag int64
		for topic, partitions := range offsets.Blocks {
			topics[topic] = struct{}{}
			for partitionID, block := range partitions {
				partitionCount++
				if block == nil || block.Offset < 0 {
					continue
				}
				latest, latestErr := client.GetOffset(topic, partitionID, sarama.OffsetNewest)
				if latestErr == nil && latest > block.Offset {
					totalLag += latest - block.Offset
				}
			}
		}
		topicList := make([]string, 0, len(topics))
		for topic := range topics {
			topicList = append(topicList, topic)
		}
		sort.Strings(topicList)
		result = append(result, response.KafkaConsumerGroupVO{GroupID: group.GroupId, ProtocolType: group.ProtocolType, State: group.State, MemberCount: len(group.Members), Topics: topicList, PartitionCount: partitionCount, CommittedLag: totalLag})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].GroupID < result[j].GroupID })
	return result, nil
}

func (s *Service) ResetConsumerGroupOffset(groupID string, req reqKafka.ResetConsumerGroupOffsetRequest) (*response.KafkaConsumerGroupOffsetResetVO, error) {
	if strings.TrimSpace(groupID) == "" {
		return nil, errors.New("消费组不能为空")
	}
	cluster, err := s.repo.GetByID(req.ClusterID)
	if err != nil {
		return nil, err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return nil, err
	}
	defer admin.Close()
	defer client.Close()
	partitions, err := client.Partitions(req.Topic)
	if err != nil {
		return nil, err
	}

	targetPartitions := make([]int32, 0, len(partitions))
	if req.AllPartitions {
		targetPartitions = append(targetPartitions, partitions...)
	} else {
		if req.Partition == nil {
			return nil, errors.New("请选择需要重置的分区，或勾选全部分区")
		}
		exists := false
		for _, partition := range partitions {
			if partition == *req.Partition {
				exists = true
				break
			}
		}
		if !exists {
			return nil, fmt.Errorf("Topic %s 不存在分区 %d", req.Topic, *req.Partition)
		}
		targetPartitions = append(targetPartitions, *req.Partition)
	}

	sort.Slice(targetPartitions, func(i, j int) bool { return targetPartitions[i] < targetPartitions[j] })
	targetOffsets := make(map[int32]int64, len(targetPartitions))
	resultPartitions := make([]response.KafkaConsumerGroupOffsetResetPartitionVO, 0, len(targetPartitions))
	for _, partition := range targetPartitions {
		targetOffset, resolveErr := resolveTargetOffset(client, req.Topic, partition, req)
		if resolveErr != nil {
			return nil, resolveErr
		}
		targetOffsets[partition] = targetOffset
		resultPartitions = append(resultPartitions, response.KafkaConsumerGroupOffsetResetPartitionVO{Partition: partition, Offset: targetOffset})
	}

	offsetManager, err := sarama.NewOffsetManagerFromClient(groupID, client)
	if err != nil {
		return nil, err
	}

	partitionManagers := make([]sarama.PartitionOffsetManager, 0, len(targetPartitions))
	for _, partition := range targetPartitions {
		partitionOffsetManager, manageErr := offsetManager.ManagePartition(req.Topic, partition)
		if manageErr != nil {
			for _, manager := range partitionManagers {
				_ = manager.Close()
			}
			_ = offsetManager.Close()
			return nil, manageErr
		}
		partitionOffsetManager.ResetOffset(targetOffsets[partition], fmt.Sprintf("reset by kafka-console at %s", time.Now().Format(time.RFC3339)))
		partitionManagers = append(partitionManagers, partitionOffsetManager)
	}
	offsetManager.Commit()
	for _, manager := range partitionManagers {
		_ = manager.Close()
	}
	_ = offsetManager.Close()
	return &response.KafkaConsumerGroupOffsetResetVO{GroupID: groupID, Topic: req.Topic, AllPartitions: req.AllPartitions, ResetType: req.ResetType, Partitions: resultPartitions}, nil
}

func (s *Service) GetDashboard(clusterID uint) (*response.KafkaDashboardVO, error) {
	brokers, err := s.ListBrokers(clusterID)
	if err != nil {
		return nil, err
	}
	topics, err := s.ListTopics(clusterID, "")
	if err != nil {
		return nil, err
	}
	groups, err := s.ListConsumerGroups(clusterID, "")
	if err != nil {
		return nil, err
	}
	partitions := 0
	var totalLag int64
	for _, topic := range topics {
		partitions += int(topic.Partitions)
	}
	for _, group := range groups {
		totalLag += group.CommittedLag
	}
	groupCount := len(groups)
	sort.Slice(groups, func(i, j int) bool { return groups[i].CommittedLag > groups[j].CommittedLag })
	topLagGroups := groups
	if len(topLagGroups) > 5 {
		topLagGroups = topLagGroups[:5]
	}
	return &response.KafkaDashboardVO{BrokerCount: len(brokers), TopicCount: len(topics), ConsumerGroupCount: groupCount, TotalPartitions: partitions, TotalLag: totalLag, TopLagGroups: topLagGroups}, nil
}

func (s *Service) openClients(cluster *dal.KafkaCluster) (sarama.Client, sarama.ClusterAdmin, error) {
	addrs := normalizeBootstrapServers(cluster.BootstrapServers)
	if len(addrs) == 0 {
		return nil, nil, errors.New("Kafka bootstrap servers 不能为空")
	}
	config, err := s.buildKafkaConfig(cluster)
	if err != nil {
		return nil, nil, err
	}
	client, err := sarama.NewClient(addrs, config)
	if err != nil {
		return nil, nil, err
	}
	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return client, admin, nil
}

func (s *Service) buildKafkaConfig(cluster *dal.KafkaCluster) (*sarama.Config, error) {
	version, err := parseKafkaVersion(cluster.Version)
	if err != nil {
		return nil, err
	}
	config := sarama.NewConfig()
	config.Version = version
	config.ClientID = "kafka-console"
	config.Metadata.Full = true
	config.Metadata.AllowAutoTopicCreation = false
	config.Admin.Timeout = 15 * time.Second
	config.Net.DialTimeout = 10 * time.Second
	config.Net.ReadTimeout = 15 * time.Second
	config.Net.WriteTimeout = 15 * time.Second
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = false
	if cluster.TLSEnabled {
		tlsConfig, err := s.buildTLSConfig(cluster)
		if err != nil {
			return nil, err
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}
	if cluster.AuthType != "" && cluster.AuthType != dal.KafkaAuthTypeNone {
		password, err := cryptoutil.DecryptString(cluster.PasswordCiphertext)
		if err != nil {
			return nil, err
		}
		config.Net.SASL.Enable = true
		config.Net.SASL.User = cluster.Username
		config.Net.SASL.Password = password
		config.Net.SASL.Handshake = true
		config.Net.SASL.Version = sarama.SASLHandshakeV1
		switch cluster.AuthType {
		case dal.KafkaAuthTypePlain:
			config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		case dal.KafkaAuthTypeSCRAMSHA256:
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: scram.HashGeneratorFcn(sha256.New)}
			}
		case dal.KafkaAuthTypeSCRAMSHA512:
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: scram.HashGeneratorFcn(sha512.New)}
			}
		default:
			return nil, fmt.Errorf("不支持的 Kafka 认证类型: %s", cluster.AuthType)
		}
	}
	return config, config.Validate()
}

func (s *Service) buildTLSConfig(cluster *dal.KafkaCluster) (*tls.Config, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: cluster.InsecureSkipVerify}
	if strings.TrimSpace(cluster.CACert) != "" {
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM([]byte(cluster.CACert)); !ok {
			return nil, errors.New("Kafka CA 证书解析失败")
		}
		tlsConfig.RootCAs = pool
	}
	if strings.TrimSpace(cluster.ClientCert) != "" && strings.TrimSpace(cluster.ClientKeyCiphertext) != "" {
		clientKey, err := cryptoutil.DecryptString(cluster.ClientKeyCiphertext)
		if err != nil {
			return nil, err
		}
		cert, err := tls.X509KeyPair([]byte(cluster.ClientCert), []byte(clientKey))
		if err != nil {
			return nil, errors.New("Kafka 客户端证书解析失败")
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	return tlsConfig, nil
}

func buildClusterModel(existing *dal.KafkaCluster, req reqKafka.ClusterUpsertRequest) (*dal.KafkaCluster, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("Kafka 集群名称不能为空")
	}
	if strings.TrimSpace(req.BootstrapServers) == "" {
		return nil, errors.New("Kafka bootstrap servers 不能为空")
	}
	authType := req.AuthType
	if authType == "" {
		authType = dal.KafkaAuthTypeNone
	}
	version := strings.TrimSpace(req.Version)
	if version == "" {
		version = "3.6.0"
	}
	cluster := existing
	if cluster == nil {
		cluster = &dal.KafkaCluster{Status: dal.KafkaClusterStatusUnknown}
	}
	cluster.Name = strings.TrimSpace(req.Name)
	cluster.BootstrapServers = strings.Join(normalizeBootstrapServers(req.BootstrapServers), ",")
	cluster.Version = version
	cluster.AuthType = authType
	cluster.Username = strings.TrimSpace(req.Username)
	cluster.TLSEnabled = req.TLSEnabled
	cluster.InsecureSkipVerify = req.InsecureSkipVerify
	cluster.CACert = req.CACert
	cluster.ClientCert = req.ClientCert
	cluster.Description = req.Description
	cluster.Environment = strings.TrimSpace(req.Environment)
	cluster.Tenant = strings.TrimSpace(req.Tenant)
	if req.Password != "" {
		cipherText, err := cryptoutil.EncryptString(req.Password)
		if err != nil {
			return nil, err
		}
		cluster.PasswordCiphertext = cipherText
	} else if authType == dal.KafkaAuthTypeNone {
		cluster.PasswordCiphertext = ""
		cluster.Username = ""
	}
	if req.ClientKey != "" {
		cipherText, err := cryptoutil.EncryptString(req.ClientKey)
		if err != nil {
			return nil, err
		}
		cluster.ClientKeyCiphertext = cipherText
	} else if strings.TrimSpace(req.ClientCert) == "" {
		cluster.ClientKeyCiphertext = ""
	}
	return cluster, nil
}

func normalizeBootstrapServers(raw string) []string {
	items := strings.Split(raw, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func parseKafkaVersion(version string) (sarama.KafkaVersion, error) {
	if strings.TrimSpace(version) == "" {
		version = "3.6.0"
	}
	parsed, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		return sarama.KafkaVersion{}, fmt.Errorf("Kafka 版本解析失败: %w", err)
	}
	return parsed, nil
}

func resolveTargetOffset(client sarama.Client, topic string, partition int32, req reqKafka.ResetConsumerGroupOffsetRequest) (int64, error) {
	switch req.ResetType {
	case "earliest":
		return client.GetOffset(topic, partition, sarama.OffsetOldest)
	case "latest":
		return client.GetOffset(topic, partition, sarama.OffsetNewest)
	case "offset":
		oldest, err := client.GetOffset(topic, partition, sarama.OffsetOldest)
		if err != nil {
			return 0, err
		}
		latest, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return 0, err
		}
		if req.Offset < oldest || req.Offset > latest {
			return 0, fmt.Errorf("自定义 offset 超出范围，当前允许范围为 [%d, %d]", oldest, latest)
		}
		return req.Offset, nil
	case "timestamp":
		if req.TimestampMs <= 0 {
			return 0, errors.New("请提供有效的时间戳")
		}
		offset, err := client.GetOffset(topic, partition, req.TimestampMs)
		if err != nil {
			return 0, err
		}
		if offset >= 0 {
			return offset, nil
		}
		return client.GetOffset(topic, partition, sarama.OffsetNewest)
	default:
		return 0, fmt.Errorf("不支持的 offset 重置类型: %s", req.ResetType)
	}
}

func toClusterVO(cluster dal.KafkaCluster) response.KafkaClusterVO {
	return response.KafkaClusterVO{ID: cluster.ID, Name: cluster.Name, BootstrapServers: cluster.BootstrapServers, Version: cluster.Version, AuthType: cluster.AuthType, Username: cluster.Username, TLSEnabled: cluster.TLSEnabled, InsecureSkipVerify: cluster.InsecureSkipVerify, CACert: cluster.CACert, ClientCert: cluster.ClientCert, Description: cluster.Description, Environment: cluster.Environment, Tenant: cluster.Tenant, Status: cluster.Status, LastErrorMessage: cluster.LastErrorMessage, LastTestedAt: cluster.LastTestedAt, CreatedAt: cluster.CreatedAt, UpdatedAt: cluster.UpdatedAt}
}
