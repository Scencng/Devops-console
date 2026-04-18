package kafka

type AlertSilenceListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId"`
}

type AlertSilenceUpsertRequest struct {
	ClusterID  uint   `json:"clusterId" binding:"required"`
	Name       string `json:"name" binding:"required,max=191"`
	MetricType string `json:"metricType" binding:"omitempty,max=64"`
	Severity   string `json:"severity" binding:"omitempty,max=32"`
	StartsAt   string `json:"startsAt" binding:"required"`
	EndsAt     string `json:"endsAt" binding:"required"`
	Enabled    bool   `json:"enabled"`
	Comment    string `json:"comment"`
}

type TraceLinkListRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId"`
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=255"`
}

type TraceLinkCreateRequest struct {
	ClusterID     uint   `json:"clusterId" binding:"required"`
	TraceID       string `json:"traceId" binding:"required,max=191"`
	SpanID        string `json:"spanId" binding:"omitempty,max=191"`
	ServiceName   string `json:"serviceName" binding:"omitempty,max=191"`
	Topic         string `json:"topic" binding:"omitempty,max=255"`
	MessageKey    string `json:"messageKey" binding:"omitempty,max=255"`
	ConsumerGroup string `json:"consumerGroup" binding:"omitempty,max=255"`
	Headers       string `json:"headers"`
	Description   string `json:"description"`
}

type ScalingRecommendationRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId" binding:"required"`
}

type SelfHealingPolicyListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId"`
}

type SelfHealingPolicyUpsertRequest struct {
	ClusterID   uint   `json:"clusterId" binding:"required"`
	Name        string `json:"name" binding:"required,max=191"`
	TriggerType string `json:"triggerType" binding:"required,max=64"`
	ActionType  string `json:"actionType" binding:"required,max=64"`
	Config      string `json:"config"`
	Enabled     bool   `json:"enabled"`
}

type SelfHealingExecutionListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId"`
}

type GitOpsProfileListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId"`
}

type GitOpsProfileUpsertRequest struct {
	ClusterID      uint   `json:"clusterId" binding:"required"`
	Name           string `json:"name" binding:"required,max=191"`
	RepoURL        string `json:"repoUrl" binding:"required,max=500"`
	Branch         string `json:"branch" binding:"omitempty,max=128"`
	BasePath       string `json:"basePath" binding:"omitempty,max=255"`
	ManifestFormat string `json:"manifestFormat" binding:"omitempty,max=64"`
	AuthType       string `json:"authType" binding:"omitempty,max=32"`
	Token          string `json:"token"`
	Enabled        bool   `json:"enabled"`
}

type GitOpsSyncListRequest struct {
	ProfileID uint `form:"profileId" json:"profileId"`
}

type CloudAdapterListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId"`
}

type CloudAdapterUpsertRequest struct {
	ClusterID         uint   `json:"clusterId" binding:"required"`
	Provider          string `json:"provider" binding:"required,max=64"`
	ServiceName       string `json:"serviceName" binding:"required,max=128"`
	Region            string `json:"region" binding:"omitempty,max=64"`
	ClusterIdentifier string `json:"clusterIdentifier" binding:"omitempty,max=191"`
	EndpointMode      string `json:"endpointMode" binding:"omitempty,max=64"`
	Notes             string `json:"notes"`
}

type LineageListRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId"`
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=255"`
}

type LineageUpsertRequest struct {
	ClusterID       uint   `json:"clusterId" binding:"required"`
	SourceTopic     string `json:"sourceTopic" binding:"required,max=255"`
	TargetTopic     string `json:"targetTopic" binding:"required,max=255"`
	RelationType    string `json:"relationType" binding:"required,max=64"`
	ProducerService string `json:"producerService" binding:"omitempty,max=191"`
	ConsumerService string `json:"consumerService" binding:"omitempty,max=191"`
	Description     string `json:"description"`
}

type LifecycleReportRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId" binding:"required"`
}

type MeshGatewayListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId"`
}

type MeshGatewayUpsertRequest struct {
	ClusterID   uint   `json:"clusterId" binding:"required"`
	GatewayType string `json:"gatewayType" binding:"required,max=64"`
	Endpoint    string `json:"endpoint" binding:"omitempty,max=500"`
	AuthMode    string `json:"authMode" binding:"omitempty,max=64"`
	Config      string `json:"config"`
	Enabled     bool   `json:"enabled"`
}

type CostRecordListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId" binding:"required"`
}

type CostRecordGenerateRequest struct {
	ClusterID uint `json:"clusterId" binding:"required"`
}

type SensitiveRuleListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId"`
}

type SensitiveRuleUpsertRequest struct {
	ClusterID    uint   `json:"clusterId" binding:"required"`
	Name         string `json:"name" binding:"required,max=191"`
	PatternType  string `json:"patternType" binding:"required,max=64"`
	PatternValue string `json:"patternValue" binding:"required,max=500"`
	Severity     string `json:"severity" binding:"omitempty,max=32"`
	Enabled      bool   `json:"enabled"`
}

type SensitiveScanRunRequest struct {
	ClusterID uint   `json:"clusterId" binding:"required"`
	Topic     string `json:"topic" binding:"required,max=255"`
	Partition int32  `json:"partition" binding:"required,min=0"`
	Limit     int    `json:"limit" binding:"omitempty,min=1,max=200"`
}

type SensitiveResultListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId"`
}
