package response

import "time"

type KafkaAlertSilenceVO struct {
	ID         uint      `json:"id"`
	ClusterID  uint      `json:"clusterId"`
	Name       string    `json:"name"`
	MetricType string    `json:"metricType"`
	Severity   string    `json:"severity"`
	StartsAt   time.Time `json:"startsAt"`
	EndsAt     time.Time `json:"endsAt"`
	Enabled    bool      `json:"enabled"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type KafkaTraceLinkVO struct {
	ID            uint      `json:"id"`
	ClusterID     uint      `json:"clusterId"`
	TraceID       string    `json:"traceId"`
	SpanID        string    `json:"spanId"`
	ServiceName   string    `json:"serviceName"`
	Topic         string    `json:"topic"`
	MessageKey    string    `json:"messageKey"`
	ConsumerGroup string    `json:"consumerGroup"`
	Headers       string    `json:"headers"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type KafkaScalingRecommendationVO struct {
	ID               uint      `json:"id"`
	ClusterID        uint      `json:"clusterId"`
	ResourceType     string    `json:"resourceType"`
	ResourceName     string    `json:"resourceName"`
	CurrentValue     float64   `json:"currentValue"`
	RecommendedValue float64   `json:"recommendedValue"`
	Reason           string    `json:"reason"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type KafkaSelfHealingPolicyVO struct {
	ID          uint      `json:"id"`
	ClusterID   uint      `json:"clusterId"`
	Name        string    `json:"name"`
	TriggerType string    `json:"triggerType"`
	ActionType  string    `json:"actionType"`
	Config      string    `json:"config"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type KafkaSelfHealingExecutionVO struct {
	ID          uint       `json:"id"`
	PolicyID    uint       `json:"policyId"`
	ClusterID   uint       `json:"clusterId"`
	Status      string     `json:"status"`
	Summary     string     `json:"summary"`
	Result      string     `json:"result"`
	StartedAt   time.Time  `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
}

type KafkaGitOpsProfileVO struct {
	ID              uint       `json:"id"`
	ClusterID       uint       `json:"clusterId"`
	Name            string     `json:"name"`
	RepoURL         string     `json:"repoUrl"`
	Branch          string     `json:"branch"`
	BasePath        string     `json:"basePath"`
	ManifestFormat  string     `json:"manifestFormat"`
	AuthType        string     `json:"authType"`
	Enabled         bool       `json:"enabled"`
	LastSyncStatus  string     `json:"lastSyncStatus"`
	LastSyncAt      *time.Time `json:"lastSyncAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type KafkaGitOpsSyncRecordVO struct {
	ID         uint       `json:"id"`
	ProfileID  uint       `json:"profileId"`
	Status     string     `json:"status"`
	Summary    string     `json:"summary"`
	CommitSHA  string     `json:"commitSha"`
	Output     string     `json:"output"`
	StartedAt  time.Time  `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
}

type KafkaCloudAdapterVO struct {
	ID                uint      `json:"id"`
	ClusterID         uint      `json:"clusterId"`
	Provider          string    `json:"provider"`
	ServiceName       string    `json:"serviceName"`
	Region            string    `json:"region"`
	ClusterIdentifier string    `json:"clusterIdentifier"`
	EndpointMode      string    `json:"endpointMode"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type KafkaLineageRelationVO struct {
	ID              uint      `json:"id"`
	ClusterID       uint      `json:"clusterId"`
	SourceTopic     string    `json:"sourceTopic"`
	TargetTopic     string    `json:"targetTopic"`
	RelationType    string    `json:"relationType"`
	ProducerService string    `json:"producerService"`
	ConsumerService string    `json:"consumerService"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type KafkaLifecyclePolicyVO struct {
	ID                   uint      `json:"id"`
	ClusterID            uint      `json:"clusterId"`
	TopicName            string    `json:"topicName"`
	Action               string    `json:"action"`
	TargetRetentionHours int       `json:"targetRetentionHours"`
	Owner                string    `json:"owner"`
	Status               string    `json:"status"`
	Recommendation       string    `json:"recommendation"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type KafkaMeshGatewayConfigVO struct {
	ID          uint      `json:"id"`
	ClusterID   uint      `json:"clusterId"`
	GatewayType string    `json:"gatewayType"`
	Endpoint    string    `json:"endpoint"`
	AuthMode    string    `json:"authMode"`
	Config      string    `json:"config"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type KafkaCostRecordVO struct {
	ID             uint      `json:"id"`
	ClusterID      uint      `json:"clusterId"`
	MetricDate     time.Time `json:"metricDate"`
	StorageBytes   float64   `json:"storageBytes"`
	IngressBytes   float64   `json:"ingressBytes"`
	EgressBytes    float64   `json:"egressBytes"`
	PartitionCount int       `json:"partitionCount"`
	EstimatedCost  float64   `json:"estimatedCost"`
	Currency       string    `json:"currency"`
	CreatedAt      time.Time `json:"createdAt"`
}

type KafkaSensitiveScanRuleVO struct {
	ID           uint      `json:"id"`
	ClusterID    uint      `json:"clusterId"`
	Name         string    `json:"name"`
	PatternType  string    `json:"patternType"`
	PatternValue string    `json:"patternValue"`
	Severity     string    `json:"severity"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type KafkaSensitiveScanResultVO struct {
	ID          uint      `json:"id"`
	ClusterID   uint      `json:"clusterId"`
	Topic       string    `json:"topic"`
	Partition   int32     `json:"partition"`
	Offset      int64     `json:"offset"`
	RuleName    string    `json:"ruleName"`
	Severity    string    `json:"severity"`
	MatchedText string    `json:"matchedText"`
	Summary     string    `json:"summary"`
	CreatedAt   time.Time `json:"createdAt"`
}
