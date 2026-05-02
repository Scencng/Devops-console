package kafka

import (
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
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

const (
	latestOffsetUnavailable     int64 = -1
	consumerGroupMetricsWorkers       = 6
	defaultKafkaVersion               = "3.6.0"
)

type latestOffsetCache struct {
	mu       sync.Mutex
	data     map[string]map[int32]int64
	inflight map[string]chan struct{}
}

func newLatestOffsetCache() *latestOffsetCache {
	return &latestOffsetCache{
		data:     make(map[string]map[int32]int64),
		inflight: make(map[string]chan struct{}),
	}
}

type Service struct {
	repo      *configs.KafkaClusterRepository
	auditRepo *configs.KafkaAuditLogRepository
}

func NewService(repo *configs.KafkaClusterRepository, auditRepo *configs.KafkaAuditLogRepository) *Service {
	return &Service{repo: repo, auditRepo: auditRepo}
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

func (s *Service) GetCluster(id uint) (*response.KafkaClusterDetailVO, error) {
	cluster, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	vo := toClusterDetailVO(*cluster)
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
	if err != nil {
		testedAt := time.Now()
		_ = s.repo.UpdateTestStatus(id, dal.KafkaClusterStatusError, err.Error(), &testedAt)
		_ = s.repo.SaveTestRecord(&dal.ConnectionTest{ResourceType: dal.KafkaResourceTypeCluster, ResourceID: id, TestResult: "failure", ErrorMessage: err.Error(), TestedAt: testedAt})
		return &response.KafkaClusterTestVO{ClusterID: id, ClusterName: cluster.Name, ResponseTime: time.Since(start).Milliseconds(), TestedAt: testedAt, Status: dal.KafkaClusterStatusError, ErrorMessage: err.Error()}, err
	}
	defer client.Close()
	defer admin.Close()
	brokers, controllerID, err := admin.DescribeCluster()
	if err != nil {
		testedAt := time.Now()
		_ = s.repo.UpdateTestStatus(id, dal.KafkaClusterStatusError, err.Error(), &testedAt)
		_ = s.repo.SaveTestRecord(&dal.ConnectionTest{ResourceType: dal.KafkaResourceTypeCluster, ResourceID: id, TestResult: "failure", ErrorMessage: err.Error(), TestedAt: testedAt})
		return &response.KafkaClusterTestVO{ClusterID: id, ClusterName: cluster.Name, ResponseTime: time.Since(start).Milliseconds(), TestedAt: testedAt, Status: dal.KafkaClusterStatusError, ErrorMessage: err.Error()}, err
	}
	testedAt := time.Now()
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
	defer client.Close()
	defer admin.Close()
	return s.listTopicsWithClients(client, admin, keyword)
}

func (s *Service) listTopicsWithClients(client sarama.Client, admin sarama.ClusterAdmin, keyword string) ([]response.KafkaTopicVO, error) {
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
	defer client.Close()
	defer admin.Close()
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
	defer client.Close()
	defer admin.Close()
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

func (s *Service) UpdateBrokerConfig(clusterID uint, brokerID int32, configs map[string]string) error {
	if len(configs) == 0 {
		return errors.New("至少需要提供一项 Broker 配置")
	}
	cluster, err := s.repo.GetByID(clusterID)
	if err != nil {
		return err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return err
	}
	defer client.Close()
	defer admin.Close()

	brokers, _, err := admin.DescribeCluster()
	if err != nil {
		return err
	}
	brokerExists := false
	for _, broker := range brokers {
		if broker != nil && broker.ID() == brokerID {
			brokerExists = true
			break
		}
	}
	if !brokerExists {
		return fmt.Errorf("Broker %d 不存在", brokerID)
	}

	alterEntries := make(map[string]*string, len(configs))
	for rawKey, rawValue := range configs {
		key := strings.TrimSpace(rawKey)
		if key == "" {
			return errors.New("Broker 配置项 key 不能为空")
		}
		value := rawValue
		alterEntries[key] = &value
	}

	return admin.AlterConfig(sarama.BrokerResource, strconv.FormatInt(int64(brokerID), 10), alterEntries, false)
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
	defer client.Close()
	defer admin.Close()
	return s.listBrokersWithClients(client, admin)
}

func (s *Service) listBrokersWithClients(client sarama.Client, admin sarama.ClusterAdmin) ([]response.KafkaBrokerVO, error) {
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
	defer client.Close()
	defer admin.Close()
	return s.listConsumerGroupsWithClients(client, admin, keyword)
}

func (s *Service) listConsumerGroupsWithClients(client sarama.Client, admin sarama.ClusterAdmin, keyword string) ([]response.KafkaConsumerGroupVO, error) {
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
	cache := newLatestOffsetCache()
	workerCount := consumerGroupMetricsWorkers
	if len(groups) < workerCount {
		workerCount = len(groups)
	}
	type workerJob struct {
		group *sarama.GroupDescription
	}
	type workerResult struct {
		group response.KafkaConsumerGroupVO
	}
	jobs := make(chan workerJob, len(groups))
	results := make(chan workerResult, len(groups))
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			workerAdmin, workerErr := sarama.NewClusterAdminFromClient(client)
			if workerErr == nil {
				defer workerAdmin.Close()
			}
			for job := range jobs {
				results <- workerResult{group: buildConsumerGroupSummary(workerAdmin, workerErr, client, job.group, cache)}
			}
		}()
	}
	for _, group := range groups {
		jobs <- workerJob{group: group}
	}
	close(jobs)
	go func() {
		wg.Wait()
		close(results)
	}()
	for item := range results {
		result = append(result, item.group)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].GroupID < result[j].GroupID })
	return result, nil
}

// ResetConsumerGroupOffset validates the target group and partitions first, then uses
// sarama.NewOffsetManagerFromClient + ManagePartition + ResetOffset to overwrite the
// committed offsets stored in Kafka for the specified consumer group.
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
	defer client.Close()
	defer admin.Close()
	if err = ensureConsumerGroupCanReset(admin, groupID, req.Force); err != nil {
		return nil, err
	}
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
			if closeErr := closeOffsetManagers(partitionManagers, offsetManager); closeErr != nil {
				return nil, fmt.Errorf("%w；资源清理失败：%v", manageErr, closeErr)
			}
			return nil, manageErr
		}
		partitionOffsetManager.ResetOffset(targetOffsets[partition], fmt.Sprintf("reset by kafka-console at %s", time.Now().Format(time.RFC3339)))
		partitionManagers = append(partitionManagers, partitionOffsetManager)
	}
	if closeErr := closeOffsetManagers(partitionManagers, offsetManager); closeErr != nil {
		return nil, closeErr
	}
	return &response.KafkaConsumerGroupOffsetResetVO{GroupID: groupID, Topic: req.Topic, AllPartitions: req.AllPartitions, ResetType: req.ResetType, Partitions: resultPartitions}, nil
}

func (s *Service) GetDashboard(clusterID uint) (*response.KafkaDashboardVO, error) {
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

	brokers, err := s.listBrokersWithClients(client, admin)
	if err != nil {
		return nil, err
	}
	topics, err := s.listTopicsWithClients(client, admin, "")
	if err != nil {
		return nil, err
	}
	groups, err := s.listConsumerGroupsWithClients(client, admin, "")
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
	clientCert := strings.TrimSpace(cluster.ClientCert)
	clientKeyCiphertext := strings.TrimSpace(cluster.ClientKeyCiphertext)
	if clientCert != "" || clientKeyCiphertext != "" {
		if clientCert == "" || clientKeyCiphertext == "" {
			return nil, errors.New("Kafka 客户端证书和私钥必须同时配置")
		}
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
		version = defaultKafkaVersion
	}
	normalizedBootstrapServers := strings.Join(normalizeBootstrapServers(req.BootstrapServers), ",")
	clientCert := strings.TrimSpace(req.ClientCert)
	clientKey := strings.TrimSpace(req.ClientKey)
	if clientKey != "" && clientCert == "" {
		return nil, errors.New("提供客户端私钥时必须同时提供客户端证书")
	}
	existingClientCert := ""
	if existing != nil {
		existingClientCert = strings.TrimSpace(existing.ClientCert)
	}
	if clientKey == "" {
		switch {
		case clientCert == "":
		case existing == nil:
			return nil, errors.New("提供客户端证书时必须同时提供客户端私钥")
		case clientCert != existingClientCert:
			return nil, errors.New("修改客户端证书时请同时提供客户端私钥，或清空客户端证书")
		}
	}
	cluster := &dal.KafkaCluster{Status: dal.KafkaClusterStatusUnknown}
	if existing != nil {
		copyCluster := *existing
		cluster = &copyCluster
	}
	cluster.Name = strings.TrimSpace(req.Name)
	cluster.BootstrapServers = normalizedBootstrapServers
	cluster.Version = version
	cluster.AuthType = authType
	cluster.Username = strings.TrimSpace(req.Username)
	cluster.TLSEnabled = req.TLSEnabled
	cluster.InsecureSkipVerify = req.InsecureSkipVerify
	cluster.CACert = strings.TrimSpace(req.CACert)
	cluster.ClientCert = clientCert
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
	if clientKey != "" {
		cipherText, err := cryptoutil.EncryptString(req.ClientKey)
		if err != nil {
			return nil, err
		}
		cluster.ClientKeyCiphertext = cipherText
	} else if clientCert == "" {
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
		version = defaultKafkaVersion
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

func (c *latestOffsetCache) get(client sarama.Client, topic string, partition int32) (int64, error) {
	cacheKey := fmt.Sprintf("%s#%d", topic, partition)
	for {
		c.mu.Lock()
		if partitions, ok := c.data[topic]; ok {
			if latest, exists := partitions[partition]; exists {
				c.mu.Unlock()
				if latest == latestOffsetUnavailable {
					return 0, errors.New("latest offset unavailable")
				}
				return latest, nil
			}
		}
		if waitCh, exists := c.inflight[cacheKey]; exists {
			c.mu.Unlock()
			<-waitCh
			continue
		}
		waitCh := make(chan struct{})
		c.inflight[cacheKey] = waitCh
		c.mu.Unlock()

		latest, err := client.GetOffset(topic, partition, sarama.OffsetNewest)

		c.mu.Lock()
		if c.data[topic] == nil {
			c.data[topic] = map[int32]int64{}
		}
		if err != nil {
			c.data[topic][partition] = latestOffsetUnavailable
		} else {
			c.data[topic][partition] = latest
		}
		close(c.inflight[cacheKey])
		delete(c.inflight, cacheKey)
		c.mu.Unlock()

		if err != nil {
			return 0, err
		}
		return latest, nil
	}
}

func buildConsumerGroupSummary(admin sarama.ClusterAdmin, adminErr error, client sarama.Client, group *sarama.GroupDescription, cache *latestOffsetCache) response.KafkaConsumerGroupVO {
	if group == nil {
		return response.KafkaConsumerGroupVO{
			Topics:            []string{},
			LagAvailable:      false,
			LagWarningMessage: "消费组描述为空，无法计算 Lag",
		}
	}
	summary := response.KafkaConsumerGroupVO{
		GroupID:      group.GroupId,
		ProtocolType: group.ProtocolType,
		State:        group.State,
		MemberCount:  len(group.Members),
		Topics:       []string{},
		LagAvailable: true,
	}
	if adminErr != nil {
		summary.LagAvailable = false
		summary.LagWarningMessage = "Lag 查询初始化失败，当前未能获取消费组 offsets"
		return summary
	}

	offsets, err := admin.ListConsumerGroupOffsets(group.GroupId, nil)
	if err != nil {
		summary.LagAvailable = false
		summary.LagWarningMessage = fmt.Sprintf("消费组 offsets 查询失败：%s", err.Error())
		return summary
	}

	topics := map[string]struct{}{}
	partitionCount := 0
	var totalLag int64
	partialLag := false
	offsetFailureCount := 0
	latestFailureCount := 0

	for topic, partitions := range offsets.Blocks {
		topics[topic] = struct{}{}
		for partitionID, block := range partitions {
			partitionCount++
			if block == nil {
				partialLag = true
				offsetFailureCount++
				continue
			}
			if block.Err != sarama.ErrNoError {
				partialLag = true
				offsetFailureCount++
				continue
			}
			if block.Offset < 0 {
				continue
			}
			latest, latestErr := cache.get(client, topic, partitionID)
			if latestErr != nil {
				partialLag = true
				latestFailureCount++
				continue
			}
			if latest > block.Offset {
				totalLag += latest - block.Offset
			}
		}
	}

	topicList := make([]string, 0, len(topics))
	for topic := range topics {
		topicList = append(topicList, topic)
	}
	sort.Strings(topicList)

	summary.Topics = topicList
	summary.PartitionCount = partitionCount
	summary.CommittedLag = totalLag
	summary.LagPartial = partialLag
	if partialLag {
		summary.LagWarningMessage = buildConsumerGroupLagWarning(offsetFailureCount, latestFailureCount)
	}
	return summary
}

func buildConsumerGroupLagWarning(offsetFailureCount, latestFailureCount int) string {
	parts := make([]string, 0, 2)
	if offsetFailureCount > 0 {
		parts = append(parts, fmt.Sprintf("%d 个分区 offset 查询失败", offsetFailureCount))
	}
	if latestFailureCount > 0 {
		parts = append(parts, fmt.Sprintf("%d 个分区最新 offset 查询失败", latestFailureCount))
	}
	if len(parts) == 0 {
		return ""
	}
	return "Lag 为部分结果：" + strings.Join(parts, "，")
}

func closeOffsetManagers(partitionManagers []sarama.PartitionOffsetManager, offsetManager sarama.OffsetManager) error {
	errorsList := make([]string, 0, len(partitionManagers)+1)
	for _, manager := range partitionManagers {
		if err := manager.Close(); err != nil {
			errorsList = append(errorsList, err.Error())
		}
	}
	if err := offsetManager.Close(); err != nil {
		errorsList = append(errorsList, err.Error())
	}
	if len(errorsList) == 0 {
		return nil
	}
	return fmt.Errorf("提交消费组 offset 后关闭管理器失败: %s", strings.Join(errorsList, "; "))
}

func ensureConsumerGroupCanReset(admin sarama.ClusterAdmin, groupID string, force bool) error {
	groups, err := admin.DescribeConsumerGroups([]string{groupID})
	if err != nil {
		return err
	}
	if len(groups) == 0 || groups[0] == nil {
		return errors.New("消费组不存在")
	}
	group := groups[0]
	activeMembers := len(group.Members)
	if activeMembers > 0 && !force {
		return fmt.Errorf("消费组当前仍有 %d 个活跃成员（状态：%s），请先暂停消费者，或显式选择强制重置", activeMembers, group.State)
	}
	return nil
}

func toClusterVO(cluster dal.KafkaCluster) response.KafkaClusterVO {
	return response.KafkaClusterVO{
		ID:                 cluster.ID,
		Name:               cluster.Name,
		BootstrapServers:   cluster.BootstrapServers,
		Version:            cluster.Version,
		AuthType:           cluster.AuthType,
		Username:           cluster.Username,
		TLSEnabled:         cluster.TLSEnabled,
		InsecureSkipVerify: cluster.InsecureSkipVerify,
		HasCACert:          strings.TrimSpace(cluster.CACert) != "",
		HasClientCert:      strings.TrimSpace(cluster.ClientCert) != "",
		HasClientKey:       strings.TrimSpace(cluster.ClientKeyCiphertext) != "",
		Description:        cluster.Description,
		Environment:        cluster.Environment,
		Tenant:             cluster.Tenant,
		Status:             cluster.Status,
		LastErrorMessage:   cluster.LastErrorMessage,
		LastTestedAt:       cluster.LastTestedAt,
		CreatedAt:          cluster.CreatedAt,
		UpdatedAt:          cluster.UpdatedAt,
	}
}

func toClusterDetailVO(cluster dal.KafkaCluster) response.KafkaClusterDetailVO {
	return response.KafkaClusterDetailVO{
		KafkaClusterVO: toClusterVO(cluster),
		CACert:         cluster.CACert,
		ClientCert:     cluster.ClientCert,
	}
}
