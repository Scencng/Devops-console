package kafka

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"devops-console-backend/internal/dal"
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/configs"
	cryptoutil "devops-console-backend/pkg/utils/crypto"
)

func (s *Service) ListAlertSilences(req reqKafka.AlertSilenceListRequest) ([]response.KafkaAlertSilenceVO, error) {
	list, err := configs.NewKafkaAlertSilenceRepository().List(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaAlertSilenceVO, 0, len(list))
	for _, item := range list {
		result = append(result, toAlertSilenceVO(item))
	}
	return result, nil
}

func (s *Service) CreateAlertSilence(req reqKafka.AlertSilenceUpsertRequest) (*response.KafkaAlertSilenceVO, error) {
	startsAt, ok := parseFlexibleTime(req.StartsAt)
	if !ok {
		return nil, errors.New("无效的静默开始时间")
	}
	endsAt, ok := parseFlexibleTime(req.EndsAt)
	if !ok {
		return nil, errors.New("无效的静默结束时间")
	}
	item := &dal.KafkaAlertSilence{
		ClusterID:  req.ClusterID,
		Name:       req.Name,
		MetricType: req.MetricType,
		Severity:   req.Severity,
		StartsAt:   startsAt,
		EndsAt:     endsAt,
		Enabled:    req.Enabled,
		Comment:    req.Comment,
	}
	if err := configs.NewKafkaAlertSilenceRepository().Create(item); err != nil {
		return nil, err
	}
	vo := toAlertSilenceVO(*item)
	return &vo, nil
}

func (s *Service) UpdateAlertSilence(id uint, req reqKafka.AlertSilenceUpsertRequest) (*response.KafkaAlertSilenceVO, error) {
	repo := configs.NewKafkaAlertSilenceRepository()
	item, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	startsAt, ok := parseFlexibleTime(req.StartsAt)
	if !ok {
		return nil, errors.New("无效的静默开始时间")
	}
	endsAt, ok := parseFlexibleTime(req.EndsAt)
	if !ok {
		return nil, errors.New("无效的静默结束时间")
	}
	item.ClusterID = req.ClusterID
	item.Name = req.Name
	item.MetricType = req.MetricType
	item.Severity = req.Severity
	item.StartsAt = startsAt
	item.EndsAt = endsAt
	item.Enabled = req.Enabled
	item.Comment = req.Comment
	if err = repo.Update(item); err != nil {
		return nil, err
	}
	vo := toAlertSilenceVO(*item)
	return &vo, nil
}

func (s *Service) DeleteAlertSilence(id uint) error {
	return configs.NewKafkaAlertSilenceRepository().Delete(id)
}

func (s *Service) ListTraceLinks(req reqKafka.TraceLinkListRequest) ([]response.KafkaTraceLinkVO, error) {
	list, err := configs.NewKafkaTraceLinkRepository().List(req.ClusterID, req.Keyword)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaTraceLinkVO, 0, len(list))
	for _, item := range list {
		result = append(result, toTraceLinkVO(item))
	}
	return result, nil
}

func (s *Service) CreateTraceLink(req reqKafka.TraceLinkCreateRequest) (*response.KafkaTraceLinkVO, error) {
	item := &dal.KafkaTraceLink{
		ClusterID:     req.ClusterID,
		TraceID:       req.TraceID,
		SpanID:        req.SpanID,
		ServiceName:   req.ServiceName,
		Topic:         req.Topic,
		MessageKey:    req.MessageKey,
		ConsumerGroup: req.ConsumerGroup,
		Headers:       req.Headers,
		Description:   req.Description,
	}
	if err := configs.NewKafkaTraceLinkRepository().Create(item); err != nil {
		return nil, err
	}
	vo := toTraceLinkVO(*item)
	return &vo, nil
}

func (s *Service) DeleteTraceLink(id uint) error {
	return configs.NewKafkaTraceLinkRepository().Delete(id)
}

func (s *Service) GenerateScalingRecommendations(req reqKafka.ScalingRecommendationRequest) ([]response.KafkaScalingRecommendationVO, error) {
	groups, err := s.ListConsumerGroups(req.ClusterID, "")
	if err != nil {
		return nil, err
	}
	topics, err := s.ListTopics(req.ClusterID, "")
	if err != nil {
		return nil, err
	}
	recommendations := make([]dal.KafkaScalingRecommendation, 0)
	for _, group := range groups {
		if group.CommittedLag <= 1000 {
			continue
		}
		recommended := float64(maxInt(group.MemberCount+1, int(math.Ceil(float64(group.CommittedLag)/1000))))
		recommendations = append(recommendations, dal.KafkaScalingRecommendation{
			ClusterID:        req.ClusterID,
			ResourceType:     "consumer_group",
			ResourceName:     group.GroupID,
			CurrentValue:     float64(group.MemberCount),
			RecommendedValue: recommended,
			Reason:           fmt.Sprintf("当前 Lag=%d，建议提升消费者并发到 %.0f", group.CommittedLag, recommended),
			Status:           "open",
		})
	}
	for _, topic := range topics {
		if topic.Partitions <= 0 {
			continue
		}
		if topic.Partitions < 6 && topic.RetentionMs != "" {
			recommendations = append(recommendations, dal.KafkaScalingRecommendation{
				ClusterID:        req.ClusterID,
				ResourceType:     "topic",
				ResourceName:     topic.Name,
				CurrentValue:     float64(topic.Partitions),
				RecommendedValue: float64(topic.Partitions * 2),
				Reason:           "分区数较低，建议结合吞吐和消费者并发评估扩分区",
				Status:           "open",
			})
		}
	}
	repo := configs.NewKafkaScalingRecommendationRepository()
	if err := repo.ReplaceForCluster(req.ClusterID, recommendations); err != nil {
		return nil, err
	}
	list, err := repo.List(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaScalingRecommendationVO, 0, len(list))
	for _, item := range list {
		result = append(result, toScalingRecommendationVO(item))
	}
	return result, nil
}

func (s *Service) ListSelfHealingPolicies(req reqKafka.SelfHealingPolicyListRequest) ([]response.KafkaSelfHealingPolicyVO, error) {
	list, err := configs.NewKafkaSelfHealingPolicyRepository().List(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaSelfHealingPolicyVO, 0, len(list))
	for _, item := range list {
		result = append(result, toSelfHealingPolicyVO(item))
	}
	return result, nil
}

func (s *Service) CreateSelfHealingPolicy(req reqKafka.SelfHealingPolicyUpsertRequest) (*response.KafkaSelfHealingPolicyVO, error) {
	item := &dal.KafkaSelfHealingPolicy{
		ClusterID:   req.ClusterID,
		Name:        req.Name,
		TriggerType: req.TriggerType,
		ActionType:  req.ActionType,
		Config:      req.Config,
		Enabled:     req.Enabled,
	}
	if err := configs.NewKafkaSelfHealingPolicyRepository().Create(item); err != nil {
		return nil, err
	}
	vo := toSelfHealingPolicyVO(*item)
	return &vo, nil
}

func (s *Service) UpdateSelfHealingPolicy(id uint, req reqKafka.SelfHealingPolicyUpsertRequest) (*response.KafkaSelfHealingPolicyVO, error) {
	repo := configs.NewKafkaSelfHealingPolicyRepository()
	item, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	item.ClusterID = req.ClusterID
	item.Name = req.Name
	item.TriggerType = req.TriggerType
	item.ActionType = req.ActionType
	item.Config = req.Config
	item.Enabled = req.Enabled
	if err = repo.Update(item); err != nil {
		return nil, err
	}
	vo := toSelfHealingPolicyVO(*item)
	return &vo, nil
}

func (s *Service) DeleteSelfHealingPolicy(id uint) error {
	return configs.NewKafkaSelfHealingPolicyRepository().Delete(id)
}

func (s *Service) ExecuteSelfHealingPolicy(id uint) (*response.KafkaSelfHealingExecutionVO, error) {
	repo := configs.NewKafkaSelfHealingPolicyRepository()
	policy, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	execution := &dal.KafkaSelfHealingExecution{
		PolicyID:  policy.ID,
		ClusterID: policy.ClusterID,
		Status:    "running",
		Summary:   "开始执行自愈动作",
		StartedAt: time.Now(),
	}
	if err = repo.CreateExecution(execution); err != nil {
		return nil, err
	}
	status := "success"
	summary := "自愈动作执行成功"
	result := ""
	if execErr := s.executeTask(policy.ActionType, policy.Config, policy.ClusterID, "self-healing"); execErr != nil {
		status = "failed"
		summary = execErr.Error()
		result = execErr.Error()
	}
	now := time.Now()
	execution.Status = status
	execution.Summary = summary
	execution.Result = result
	execution.CompletedAt = &now
	_ = repo.UpdateExecution(execution)
	vo := toSelfHealingExecutionVO(*execution)
	return &vo, nil
}

func (s *Service) ListSelfHealingExecutions(req reqKafka.SelfHealingExecutionListRequest) ([]response.KafkaSelfHealingExecutionVO, error) {
	list, err := configs.NewKafkaSelfHealingPolicyRepository().ListExecutions(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaSelfHealingExecutionVO, 0, len(list))
	for _, item := range list {
		result = append(result, toSelfHealingExecutionVO(item))
	}
	return result, nil
}

func (s *Service) ListGitOpsProfiles(req reqKafka.GitOpsProfileListRequest) ([]response.KafkaGitOpsProfileVO, error) {
	list, err := configs.NewKafkaGitOpsRepository().ListProfiles(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaGitOpsProfileVO, 0, len(list))
	for _, item := range list {
		result = append(result, toGitOpsProfileVO(item))
	}
	return result, nil
}

func (s *Service) CreateGitOpsProfile(req reqKafka.GitOpsProfileUpsertRequest) (*response.KafkaGitOpsProfileVO, error) {
	item := &dal.KafkaGitOpsProfile{
		ClusterID:       req.ClusterID,
		Name:            req.Name,
		RepoURL:         req.RepoURL,
		Branch:          req.Branch,
		BasePath:        req.BasePath,
		ManifestFormat:  req.ManifestFormat,
		AuthType:        req.AuthType,
		Enabled:         req.Enabled,
	}
	if req.Token != "" {
		cipher, err := cryptoutil.EncryptString(req.Token)
		if err != nil {
			return nil, err
		}
		item.TokenCiphertext = cipher
	}
	if err := configs.NewKafkaGitOpsRepository().CreateProfile(item); err != nil {
		return nil, err
	}
	vo := toGitOpsProfileVO(*item)
	return &vo, nil
}

func (s *Service) UpdateGitOpsProfile(id uint, req reqKafka.GitOpsProfileUpsertRequest) (*response.KafkaGitOpsProfileVO, error) {
	repo := configs.NewKafkaGitOpsRepository()
	item, err := repo.GetProfileByID(id)
	if err != nil {
		return nil, err
	}
	item.ClusterID = req.ClusterID
	item.Name = req.Name
	item.RepoURL = req.RepoURL
	item.Branch = req.Branch
	item.BasePath = req.BasePath
	item.ManifestFormat = req.ManifestFormat
	item.AuthType = req.AuthType
	item.Enabled = req.Enabled
	if req.Token != "" {
		cipher, cipherErr := cryptoutil.EncryptString(req.Token)
		if cipherErr != nil {
			return nil, cipherErr
		}
		item.TokenCiphertext = cipher
	}
	if err = repo.UpdateProfile(item); err != nil {
		return nil, err
	}
	vo := toGitOpsProfileVO(*item)
	return &vo, nil
}

func (s *Service) DeleteGitOpsProfile(id uint) error {
	return configs.NewKafkaGitOpsRepository().DeleteProfile(id)
}

func (s *Service) RunGitOpsSync(profileID uint) (*response.KafkaGitOpsSyncRecordVO, error) {
	repo := configs.NewKafkaGitOpsRepository()
	profile, err := repo.GetProfileByID(profileID)
	if err != nil {
		return nil, err
	}
	record := &dal.KafkaGitOpsSyncRecord{
		ProfileID: profile.ID,
		Status:    "running",
		Summary:   "开始生成 GitOps 清单",
		StartedAt: time.Now(),
	}
	if err = repo.CreateSync(record); err != nil {
		return nil, err
	}
	metadata, _ := configs.NewKafkaTopicMetadataRepository().List(profile.ClusterID, "")
	manifest := buildGitOpsManifest(profile, metadata)
	now := time.Now()
	record.Status = "success"
	record.Summary = "清单生成完成，请将输出提交到 Git 仓库"
	record.CommitSHA = now.Format("20060102150405")
	record.Output = manifest
	record.FinishedAt = &now
	_ = repo.UpdateSync(record)
	profile.LastSyncStatus = "success"
	profile.LastSyncAt = &now
	_ = repo.UpdateProfile(profile)
	vo := toGitOpsSyncRecordVO(*record)
	return &vo, nil
}

func (s *Service) ListGitOpsSyncs(req reqKafka.GitOpsSyncListRequest) ([]response.KafkaGitOpsSyncRecordVO, error) {
	list, err := configs.NewKafkaGitOpsRepository().ListSyncs(req.ProfileID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaGitOpsSyncRecordVO, 0, len(list))
	for _, item := range list {
		result = append(result, toGitOpsSyncRecordVO(item))
	}
	return result, nil
}

func (s *Service) ListCloudAdapters(req reqKafka.CloudAdapterListRequest) ([]response.KafkaCloudAdapterVO, error) {
	list, err := configs.NewKafkaCloudAdapterRepository().List(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaCloudAdapterVO, 0, len(list))
	for _, item := range list {
		result = append(result, toCloudAdapterVO(item))
	}
	return result, nil
}

func (s *Service) CreateCloudAdapter(req reqKafka.CloudAdapterUpsertRequest) (*response.KafkaCloudAdapterVO, error) {
	item := &dal.KafkaCloudAdapter{
		ClusterID:         req.ClusterID,
		Provider:          req.Provider,
		ServiceName:       req.ServiceName,
		Region:            req.Region,
		ClusterIdentifier: req.ClusterIdentifier,
		EndpointMode:      req.EndpointMode,
		Notes:             req.Notes,
	}
	if err := configs.NewKafkaCloudAdapterRepository().Create(item); err != nil {
		return nil, err
	}
	vo := toCloudAdapterVO(*item)
	return &vo, nil
}

func (s *Service) UpdateCloudAdapter(id uint, req reqKafka.CloudAdapterUpsertRequest) (*response.KafkaCloudAdapterVO, error) {
	repo := configs.NewKafkaCloudAdapterRepository()
	item, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	item.ClusterID = req.ClusterID
	item.Provider = req.Provider
	item.ServiceName = req.ServiceName
	item.Region = req.Region
	item.ClusterIdentifier = req.ClusterIdentifier
	item.EndpointMode = req.EndpointMode
	item.Notes = req.Notes
	if err = repo.Update(item); err != nil {
		return nil, err
	}
	vo := toCloudAdapterVO(*item)
	return &vo, nil
}

func (s *Service) DeleteCloudAdapter(id uint) error {
	return configs.NewKafkaCloudAdapterRepository().Delete(id)
}

func (s *Service) ListLineageRelations(req reqKafka.LineageListRequest) ([]response.KafkaLineageRelationVO, error) {
	list, err := configs.NewKafkaLineageRepository().List(req.ClusterID, req.Keyword)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaLineageRelationVO, 0, len(list))
	for _, item := range list {
		result = append(result, toLineageVO(item))
	}
	return result, nil
}

func (s *Service) CreateLineageRelation(req reqKafka.LineageUpsertRequest) (*response.KafkaLineageRelationVO, error) {
	item := &dal.KafkaLineageRelation{
		ClusterID:       req.ClusterID,
		SourceTopic:     req.SourceTopic,
		TargetTopic:     req.TargetTopic,
		RelationType:    req.RelationType,
		ProducerService: req.ProducerService,
		ConsumerService: req.ConsumerService,
		Description:     req.Description,
	}
	if err := configs.NewKafkaLineageRepository().Create(item); err != nil {
		return nil, err
	}
	vo := toLineageVO(*item)
	return &vo, nil
}

func (s *Service) UpdateLineageRelation(id uint, req reqKafka.LineageUpsertRequest) (*response.KafkaLineageRelationVO, error) {
	repo := configs.NewKafkaLineageRepository()
	item, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	item.ClusterID = req.ClusterID
	item.SourceTopic = req.SourceTopic
	item.TargetTopic = req.TargetTopic
	item.RelationType = req.RelationType
	item.ProducerService = req.ProducerService
	item.ConsumerService = req.ConsumerService
	item.Description = req.Description
	if err = repo.Update(item); err != nil {
		return nil, err
	}
	vo := toLineageVO(*item)
	return &vo, nil
}

func (s *Service) DeleteLineageRelation(id uint) error {
	return configs.NewKafkaLineageRepository().Delete(id)
}

func (s *Service) GenerateLifecyclePolicies(req reqKafka.LifecycleReportRequest) ([]response.KafkaLifecyclePolicyVO, error) {
	topics, err := s.ListTopics(req.ClusterID, "")
	if err != nil {
		return nil, err
	}
	metadata, _ := configs.NewKafkaTopicMetadataRepository().List(req.ClusterID, "")
	metadataMap := make(map[string]dal.KafkaTopicMetadata, len(metadata))
	for _, item := range metadata {
		metadataMap[item.TopicName] = item
	}
	policies := make([]dal.KafkaLifecyclePolicy, 0, len(topics))
	for _, topic := range topics {
		action := "keep"
		status := "healthy"
		recommendation := "保留当前生命周期配置"
		targetHours := 0
		owner := ""
		meta, ok := metadataMap[topic.Name]
		if ok {
			owner = meta.Owner
		}
		if !ok {
			action = "annotate"
			status = "warning"
			recommendation = "缺少 Topic 元数据，建议补齐 Owner、生命周期和敏感级别"
		} else if strings.EqualFold(meta.Lifecycle, "temporary") || strings.EqualFold(meta.Lifecycle, "临时") {
			action = "reduce_retention"
			status = "warning"
			targetHours = 24
			recommendation = "临时 Topic 建议缩短保留时间并定期清理"
		}
		policies = append(policies, dal.KafkaLifecyclePolicy{
			ClusterID:            req.ClusterID,
			TopicName:            topic.Name,
			Action:               action,
			TargetRetentionHours: targetHours,
			Owner:                owner,
			Status:               status,
			Recommendation:       recommendation,
		})
	}
	repo := configs.NewKafkaLifecycleRepository()
	if err := repo.ReplaceForCluster(req.ClusterID, policies); err != nil {
		return nil, err
	}
	list, err := repo.List(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaLifecyclePolicyVO, 0, len(list))
	for _, item := range list {
		result = append(result, toLifecycleVO(item))
	}
	return result, nil
}

func (s *Service) ListMeshGatewayConfigs(req reqKafka.MeshGatewayListRequest) ([]response.KafkaMeshGatewayConfigVO, error) {
	list, err := configs.NewKafkaMeshGatewayRepository().List(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaMeshGatewayConfigVO, 0, len(list))
	for _, item := range list {
		result = append(result, toMeshGatewayVO(item))
	}
	return result, nil
}

func (s *Service) CreateMeshGatewayConfig(req reqKafka.MeshGatewayUpsertRequest) (*response.KafkaMeshGatewayConfigVO, error) {
	item := &dal.KafkaMeshGatewayConfig{
		ClusterID:   req.ClusterID,
		GatewayType: req.GatewayType,
		Endpoint:    req.Endpoint,
		AuthMode:    req.AuthMode,
		Config:      req.Config,
		Enabled:     req.Enabled,
	}
	if err := configs.NewKafkaMeshGatewayRepository().Create(item); err != nil {
		return nil, err
	}
	vo := toMeshGatewayVO(*item)
	return &vo, nil
}

func (s *Service) UpdateMeshGatewayConfig(id uint, req reqKafka.MeshGatewayUpsertRequest) (*response.KafkaMeshGatewayConfigVO, error) {
	repo := configs.NewKafkaMeshGatewayRepository()
	item, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	item.ClusterID = req.ClusterID
	item.GatewayType = req.GatewayType
	item.Endpoint = req.Endpoint
	item.AuthMode = req.AuthMode
	item.Config = req.Config
	item.Enabled = req.Enabled
	if err = repo.Update(item); err != nil {
		return nil, err
	}
	vo := toMeshGatewayVO(*item)
	return &vo, nil
}

func (s *Service) DeleteMeshGatewayConfig(id uint) error {
	return configs.NewKafkaMeshGatewayRepository().Delete(id)
}

func (s *Service) GenerateCostRecord(req reqKafka.CostRecordGenerateRequest) (*response.KafkaCostRecordVO, error) {
	topics, err := s.ListTopics(req.ClusterID, "")
	if err != nil {
		return nil, err
	}
	storageBytes := float64(0)
	ingressBytes := float64(0)
	egressBytes := float64(0)
	if value, err := s.queryPrometheusInstantValue("sum(kafka_log_log_size)"); err == nil {
		storageBytes = value
	}
	if value, err := s.queryPrometheusInstantValue("sum(rate(kafka_server_brokertopicmetrics_bytesin_total[5m])) * 3600 * 24"); err == nil {
		ingressBytes = value
	}
	if value, err := s.queryPrometheusInstantValue("sum(rate(kafka_server_brokertopicmetrics_bytesout_total[5m])) * 3600 * 24"); err == nil {
		egressBytes = value
	}
	partitionCount := 0
	for _, topic := range topics {
		partitionCount += int(topic.Partitions)
	}
	estimatedCost := storageBytes/(1024*1024*1024)*0.12 + ingressBytes/(1024*1024*1024)*0.02 + egressBytes/(1024*1024*1024)*0.03 + float64(partitionCount)*0.01
	record := &dal.KafkaCostRecord{
		ClusterID:      req.ClusterID,
		MetricDate:     time.Now(),
		StorageBytes:   storageBytes,
		IngressBytes:   ingressBytes,
		EgressBytes:    egressBytes,
		PartitionCount: partitionCount,
		EstimatedCost:  estimatedCost,
		Currency:       "USD",
	}
	if err := configs.NewKafkaCostRepository().Create(record); err != nil {
		return nil, err
	}
	vo := toCostRecordVO(*record)
	return &vo, nil
}

func (s *Service) ListCostRecords(req reqKafka.CostRecordListRequest) ([]response.KafkaCostRecordVO, error) {
	list, err := configs.NewKafkaCostRepository().List(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaCostRecordVO, 0, len(list))
	for _, item := range list {
		result = append(result, toCostRecordVO(item))
	}
	return result, nil
}

func (s *Service) ListSensitiveRules(req reqKafka.SensitiveRuleListRequest) ([]response.KafkaSensitiveScanRuleVO, error) {
	list, err := configs.NewKafkaSensitiveRepository().ListRules(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaSensitiveScanRuleVO, 0, len(list))
	for _, item := range list {
		result = append(result, toSensitiveRuleVO(item))
	}
	return result, nil
}

func (s *Service) CreateSensitiveRule(req reqKafka.SensitiveRuleUpsertRequest) (*response.KafkaSensitiveScanRuleVO, error) {
	item := &dal.KafkaSensitiveScanRule{
		ClusterID:    req.ClusterID,
		Name:         req.Name,
		PatternType:  req.PatternType,
		PatternValue: req.PatternValue,
		Severity:     req.Severity,
		Enabled:      req.Enabled,
	}
	if err := configs.NewKafkaSensitiveRepository().CreateRule(item); err != nil {
		return nil, err
	}
	vo := toSensitiveRuleVO(*item)
	return &vo, nil
}

func (s *Service) UpdateSensitiveRule(id uint, req reqKafka.SensitiveRuleUpsertRequest) (*response.KafkaSensitiveScanRuleVO, error) {
	repo := configs.NewKafkaSensitiveRepository()
	item, err := repo.GetRuleByID(id)
	if err != nil {
		return nil, err
	}
	item.ClusterID = req.ClusterID
	item.Name = req.Name
	item.PatternType = req.PatternType
	item.PatternValue = req.PatternValue
	item.Severity = req.Severity
	item.Enabled = req.Enabled
	if err = repo.UpdateRule(item); err != nil {
		return nil, err
	}
	vo := toSensitiveRuleVO(*item)
	return &vo, nil
}

func (s *Service) DeleteSensitiveRule(id uint) error {
	return configs.NewKafkaSensitiveRepository().DeleteRule(id)
}

func (s *Service) RunSensitiveScan(req reqKafka.SensitiveScanRunRequest) ([]response.KafkaSensitiveScanResultVO, error) {
	rules, err := configs.NewKafkaSensitiveRepository().ListRules(req.ClusterID)
	if err != nil {
		return nil, err
	}
	browseResult, err := s.BrowseMessages(reqKafka.MessageBrowseRequest{
		ClusterID: req.ClusterID,
		Topic:     req.Topic,
		Partition: req.Partition,
		Mode:      "latest",
		Limit:     maxInt(req.Limit, 20),
	})
	if err != nil {
		return nil, err
	}
	results := make([]dal.KafkaSensitiveScanResult, 0)
	for _, message := range browseResult.Messages {
		content := strings.Join([]string{message.KeyPreview, message.ValuePreview}, "\n")
		for _, rule := range rules {
			if !rule.Enabled {
				continue
			}
			matched := matchSensitiveRule(content, rule.PatternType, rule.PatternValue)
			if matched == "" {
				continue
			}
			results = append(results, dal.KafkaSensitiveScanResult{
				ClusterID:   req.ClusterID,
				Topic:       req.Topic,
				Partition:   req.Partition,
				Offset:      message.Offset,
				RuleName:    rule.Name,
				Severity:    rule.Severity,
				MatchedText: matched,
				Summary:     fmt.Sprintf("命中规则 %s", rule.Name),
			})
		}
	}
	if err := configs.NewKafkaSensitiveRepository().CreateResults(results); err != nil {
		return nil, err
	}
	result := make([]response.KafkaSensitiveScanResultVO, 0, len(results))
	for _, item := range results {
		result = append(result, toSensitiveResultVO(item))
	}
	return result, nil
}

func (s *Service) ListSensitiveResults(req reqKafka.SensitiveResultListRequest) ([]response.KafkaSensitiveScanResultVO, error) {
	list, err := configs.NewKafkaSensitiveRepository().ListResults(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaSensitiveScanResultVO, 0, len(list))
	for _, item := range list {
		result = append(result, toSensitiveResultVO(item))
	}
	return result, nil
}

func buildGitOpsManifest(profile *dal.KafkaGitOpsProfile, metadata []dal.KafkaTopicMetadata) string {
	items := make([]map[string]interface{}, 0, len(metadata))
	for _, item := range metadata {
		items = append(items, map[string]interface{}{
			"apiVersion": "platform.kafka.console/v1alpha1",
			"kind":       "TopicMetadata",
			"metadata": map[string]interface{}{
				"name": item.TopicName,
			},
			"spec": map[string]interface{}{
				"systemName":  item.SystemName,
				"owner":       item.Owner,
				"ownerEmail":  item.OwnerEmail,
				"environment": item.Environment,
				"tenant":      item.Tenant,
				"lifecycle":   item.Lifecycle,
				"sensitivity": item.Sensitivity,
				"description": item.Description,
				"labels":      item.Labels,
			},
		})
	}
	payload := map[string]interface{}{
		"profile": map[string]interface{}{
			"name":           profile.Name,
			"repoUrl":        profile.RepoURL,
			"branch":         profile.Branch,
			"basePath":       profile.BasePath,
			"manifestFormat": profile.ManifestFormat,
		},
		"resources": items,
	}
	data, _ := json.MarshalIndent(payload, "", "  ")
	return string(data)
}

func matchSensitiveRule(content, patternType, patternValue string) string {
	switch patternType {
	case "contains":
		if strings.Contains(strings.ToLower(content), strings.ToLower(patternValue)) {
			return patternValue
		}
	case "regex":
		if re, err := regexp.Compile(patternValue); err == nil {
			return re.FindString(content)
		}
	default:
		if re, err := regexp.Compile(patternValue); err == nil {
			return re.FindString(content)
		}
	}
	return ""
}

func maxInt(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func toAlertSilenceVO(item dal.KafkaAlertSilence) response.KafkaAlertSilenceVO {
	return response.KafkaAlertSilenceVO{
		ID:         item.ID,
		ClusterID:  item.ClusterID,
		Name:       item.Name,
		MetricType: item.MetricType,
		Severity:   item.Severity,
		StartsAt:   item.StartsAt,
		EndsAt:     item.EndsAt,
		Enabled:    item.Enabled,
		Comment:    item.Comment,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}
}

func toTraceLinkVO(item dal.KafkaTraceLink) response.KafkaTraceLinkVO {
	return response.KafkaTraceLinkVO{
		ID:            item.ID,
		ClusterID:     item.ClusterID,
		TraceID:       item.TraceID,
		SpanID:        item.SpanID,
		ServiceName:   item.ServiceName,
		Topic:         item.Topic,
		MessageKey:    item.MessageKey,
		ConsumerGroup: item.ConsumerGroup,
		Headers:       item.Headers,
		Description:   item.Description,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func toScalingRecommendationVO(item dal.KafkaScalingRecommendation) response.KafkaScalingRecommendationVO {
	return response.KafkaScalingRecommendationVO{
		ID:               item.ID,
		ClusterID:        item.ClusterID,
		ResourceType:     item.ResourceType,
		ResourceName:     item.ResourceName,
		CurrentValue:     item.CurrentValue,
		RecommendedValue: item.RecommendedValue,
		Reason:           item.Reason,
		Status:           item.Status,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}

func toSelfHealingPolicyVO(item dal.KafkaSelfHealingPolicy) response.KafkaSelfHealingPolicyVO {
	return response.KafkaSelfHealingPolicyVO{
		ID:          item.ID,
		ClusterID:   item.ClusterID,
		Name:        item.Name,
		TriggerType: item.TriggerType,
		ActionType:  item.ActionType,
		Config:      item.Config,
		Enabled:     item.Enabled,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func toSelfHealingExecutionVO(item dal.KafkaSelfHealingExecution) response.KafkaSelfHealingExecutionVO {
	return response.KafkaSelfHealingExecutionVO{
		ID:          item.ID,
		PolicyID:    item.PolicyID,
		ClusterID:   item.ClusterID,
		Status:      item.Status,
		Summary:     item.Summary,
		Result:      item.Result,
		StartedAt:   item.StartedAt,
		CompletedAt: item.CompletedAt,
	}
}

func toGitOpsProfileVO(item dal.KafkaGitOpsProfile) response.KafkaGitOpsProfileVO {
	return response.KafkaGitOpsProfileVO{
		ID:             item.ID,
		ClusterID:      item.ClusterID,
		Name:           item.Name,
		RepoURL:        item.RepoURL,
		Branch:         item.Branch,
		BasePath:       item.BasePath,
		ManifestFormat: item.ManifestFormat,
		AuthType:       item.AuthType,
		Enabled:        item.Enabled,
		LastSyncStatus: item.LastSyncStatus,
		LastSyncAt:     item.LastSyncAt,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

func toGitOpsSyncRecordVO(item dal.KafkaGitOpsSyncRecord) response.KafkaGitOpsSyncRecordVO {
	return response.KafkaGitOpsSyncRecordVO{
		ID:         item.ID,
		ProfileID:  item.ProfileID,
		Status:     item.Status,
		Summary:    item.Summary,
		CommitSHA:  item.CommitSHA,
		Output:     item.Output,
		StartedAt:  item.StartedAt,
		FinishedAt: item.FinishedAt,
	}
}

func toCloudAdapterVO(item dal.KafkaCloudAdapter) response.KafkaCloudAdapterVO {
	return response.KafkaCloudAdapterVO{
		ID:                item.ID,
		ClusterID:         item.ClusterID,
		Provider:          item.Provider,
		ServiceName:       item.ServiceName,
		Region:            item.Region,
		ClusterIdentifier: item.ClusterIdentifier,
		EndpointMode:      item.EndpointMode,
		Notes:             item.Notes,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

func toLineageVO(item dal.KafkaLineageRelation) response.KafkaLineageRelationVO {
	return response.KafkaLineageRelationVO{
		ID:              item.ID,
		ClusterID:       item.ClusterID,
		SourceTopic:     item.SourceTopic,
		TargetTopic:     item.TargetTopic,
		RelationType:    item.RelationType,
		ProducerService: item.ProducerService,
		ConsumerService: item.ConsumerService,
		Description:     item.Description,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func toLifecycleVO(item dal.KafkaLifecyclePolicy) response.KafkaLifecyclePolicyVO {
	return response.KafkaLifecyclePolicyVO{
		ID:                   item.ID,
		ClusterID:            item.ClusterID,
		TopicName:            item.TopicName,
		Action:               item.Action,
		TargetRetentionHours: item.TargetRetentionHours,
		Owner:                item.Owner,
		Status:               item.Status,
		Recommendation:       item.Recommendation,
		CreatedAt:            item.CreatedAt,
		UpdatedAt:            item.UpdatedAt,
	}
}

func toMeshGatewayVO(item dal.KafkaMeshGatewayConfig) response.KafkaMeshGatewayConfigVO {
	return response.KafkaMeshGatewayConfigVO{
		ID:          item.ID,
		ClusterID:   item.ClusterID,
		GatewayType: item.GatewayType,
		Endpoint:    item.Endpoint,
		AuthMode:    item.AuthMode,
		Config:      item.Config,
		Enabled:     item.Enabled,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func toCostRecordVO(item dal.KafkaCostRecord) response.KafkaCostRecordVO {
	return response.KafkaCostRecordVO{
		ID:             item.ID,
		ClusterID:      item.ClusterID,
		MetricDate:     item.MetricDate,
		StorageBytes:   item.StorageBytes,
		IngressBytes:   item.IngressBytes,
		EgressBytes:    item.EgressBytes,
		PartitionCount: item.PartitionCount,
		EstimatedCost:  item.EstimatedCost,
		Currency:       item.Currency,
		CreatedAt:      item.CreatedAt,
	}
}

func toSensitiveRuleVO(item dal.KafkaSensitiveScanRule) response.KafkaSensitiveScanRuleVO {
	return response.KafkaSensitiveScanRuleVO{
		ID:           item.ID,
		ClusterID:    item.ClusterID,
		Name:         item.Name,
		PatternType:  item.PatternType,
		PatternValue: item.PatternValue,
		Severity:     item.Severity,
		Enabled:      item.Enabled,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

func toSensitiveResultVO(item dal.KafkaSensitiveScanResult) response.KafkaSensitiveScanResultVO {
	return response.KafkaSensitiveScanResultVO{
		ID:          item.ID,
		ClusterID:   item.ClusterID,
		Topic:       item.Topic,
		Partition:   item.Partition,
		Offset:      item.Offset,
		RuleName:    item.RuleName,
		Severity:    item.Severity,
		MatchedText: item.MatchedText,
		Summary:     item.Summary,
		CreatedAt:   item.CreatedAt,
	}
}
