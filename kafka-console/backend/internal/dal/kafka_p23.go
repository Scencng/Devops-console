package dal

import "time"

type KafkaAlertSilence struct {
	ID          uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID   uint       `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	Name        string     `gorm:"not null;column:name;size:191" json:"name"`
	MetricType  string     `gorm:"column:metric_type;size:64" json:"metric_type"`
	Severity    string     `gorm:"column:severity;size:32" json:"severity"`
	StartsAt    time.Time  `gorm:"column:starts_at" json:"starts_at"`
	EndsAt      time.Time  `gorm:"column:ends_at" json:"ends_at"`
	Enabled     bool       `gorm:"not null;default:true;column:enabled" json:"enabled"`
	Comment     string     `gorm:"column:comment;type:text" json:"comment"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaAlertSilence) TableName() string { return "kafka_alert_silences" }

type KafkaTraceLink struct {
	ID            uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID     uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	TraceID       string    `gorm:"not null;column:trace_id;size:191" json:"trace_id"`
	SpanID        string    `gorm:"column:span_id;size:191" json:"span_id"`
	ServiceName   string    `gorm:"column:service_name;size:191" json:"service_name"`
	Topic         string    `gorm:"column:topic;size:255" json:"topic"`
	MessageKey    string    `gorm:"column:message_key;size:255" json:"message_key"`
	ConsumerGroup string    `gorm:"column:consumer_group;size:255" json:"consumer_group"`
	Headers       string    `gorm:"column:headers;type:longtext" json:"headers"`
	Description   string    `gorm:"column:description;type:text" json:"description"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaTraceLink) TableName() string { return "kafka_trace_links" }

type KafkaScalingRecommendation struct {
	ID               uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID        uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	ResourceType     string    `gorm:"not null;column:resource_type;size:64" json:"resource_type"`
	ResourceName     string    `gorm:"not null;column:resource_name;size:255" json:"resource_name"`
	CurrentValue     float64   `gorm:"not null;default:0;column:current_value" json:"current_value"`
	RecommendedValue float64   `gorm:"not null;default:0;column:recommended_value" json:"recommended_value"`
	Reason           string    `gorm:"column:reason;type:text" json:"reason"`
	Status           string    `gorm:"not null;default:'open';column:status;size:32" json:"status"`
	CreatedAt        time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaScalingRecommendation) TableName() string { return "kafka_scaling_recommendations" }

type KafkaSelfHealingPolicy struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID   uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	Name        string    `gorm:"not null;column:name;size:191" json:"name"`
	TriggerType string    `gorm:"not null;column:trigger_type;size:64" json:"trigger_type"`
	ActionType  string    `gorm:"not null;column:action_type;size:64" json:"action_type"`
	Config      string    `gorm:"column:config;type:longtext" json:"config"`
	Enabled     bool      `gorm:"not null;default:true;column:enabled" json:"enabled"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaSelfHealingPolicy) TableName() string { return "kafka_self_healing_policies" }

type KafkaSelfHealingExecution struct {
	ID          uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	PolicyID    uint       `gorm:"index;not null;column:policy_id" json:"policy_id"`
	ClusterID   uint       `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	Status      string     `gorm:"not null;default:'running';column:status;size:32" json:"status"`
	Summary     string     `gorm:"column:summary;type:text" json:"summary"`
	Result      string     `gorm:"column:result;type:longtext" json:"result"`
	StartedAt   time.Time  `gorm:"column:started_at" json:"started_at"`
	CompletedAt *time.Time `gorm:"column:completed_at" json:"completed_at"`
}

func (KafkaSelfHealingExecution) TableName() string { return "kafka_self_healing_executions" }

type KafkaGitOpsProfile struct {
	ID             uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID      uint       `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	Name           string     `gorm:"not null;column:name;size:191" json:"name"`
	RepoURL        string     `gorm:"not null;column:repo_url;size:500" json:"repo_url"`
	Branch         string     `gorm:"column:branch;size:128" json:"branch"`
	BasePath       string     `gorm:"column:base_path;size:255" json:"base_path"`
	ManifestFormat string     `gorm:"column:manifest_format;size:64" json:"manifest_format"`
	AuthType       string     `gorm:"column:auth_type;size:32" json:"auth_type"`
	TokenCiphertext string    `gorm:"column:token_ciphertext;type:text" json:"-"`
	Enabled        bool       `gorm:"not null;default:true;column:enabled" json:"enabled"`
	LastSyncStatus string     `gorm:"column:last_sync_status;size:32" json:"last_sync_status"`
	LastSyncAt     *time.Time `gorm:"column:last_sync_at" json:"last_sync_at"`
	CreatedAt      time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaGitOpsProfile) TableName() string { return "kafka_gitops_profiles" }

type KafkaGitOpsSyncRecord struct {
	ID         uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ProfileID  uint       `gorm:"index;not null;column:profile_id" json:"profile_id"`
	Status     string     `gorm:"not null;default:'running';column:status;size:32" json:"status"`
	Summary    string     `gorm:"column:summary;type:text" json:"summary"`
	CommitSHA  string     `gorm:"column:commit_sha;size:128" json:"commit_sha"`
	Output     string     `gorm:"column:output;type:longtext" json:"output"`
	StartedAt  time.Time  `gorm:"column:started_at" json:"started_at"`
	FinishedAt *time.Time `gorm:"column:finished_at" json:"finished_at"`
}

func (KafkaGitOpsSyncRecord) TableName() string { return "kafka_gitops_sync_records" }

type KafkaCloudAdapter struct {
	ID                uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID         uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	Provider          string    `gorm:"not null;column:provider;size:64" json:"provider"`
	ServiceName       string    `gorm:"not null;column:service_name;size:128" json:"service_name"`
	Region            string    `gorm:"column:region;size:64" json:"region"`
	ClusterIdentifier string    `gorm:"column:cluster_identifier;size:191" json:"cluster_identifier"`
	EndpointMode      string    `gorm:"column:endpoint_mode;size:64" json:"endpoint_mode"`
	Notes             string    `gorm:"column:notes;type:text" json:"notes"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaCloudAdapter) TableName() string { return "kafka_cloud_adapters" }

type KafkaLineageRelation struct {
	ID              uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID       uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	SourceTopic     string    `gorm:"not null;column:source_topic;size:255" json:"source_topic"`
	TargetTopic     string    `gorm:"not null;column:target_topic;size:255" json:"target_topic"`
	RelationType    string    `gorm:"not null;column:relation_type;size:64" json:"relation_type"`
	ProducerService string    `gorm:"column:producer_service;size:191" json:"producer_service"`
	ConsumerService string    `gorm:"column:consumer_service;size:191" json:"consumer_service"`
	Description     string    `gorm:"column:description;type:text" json:"description"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaLineageRelation) TableName() string { return "kafka_lineage_relations" }

type KafkaLifecyclePolicy struct {
	ID                   uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID            uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	TopicName            string    `gorm:"not null;column:topic_name;size:255" json:"topic_name"`
	Action               string    `gorm:"column:action;size:64" json:"action"`
	TargetRetentionHours int       `gorm:"column:target_retention_hours" json:"target_retention_hours"`
	Owner                string    `gorm:"column:owner;size:128" json:"owner"`
	Status               string    `gorm:"column:status;size:32" json:"status"`
	Recommendation       string    `gorm:"column:recommendation;type:text" json:"recommendation"`
	CreatedAt            time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaLifecyclePolicy) TableName() string { return "kafka_lifecycle_policies" }

type KafkaMeshGatewayConfig struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID   uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	GatewayType string    `gorm:"not null;column:gateway_type;size:64" json:"gateway_type"`
	Endpoint    string    `gorm:"column:endpoint;size:500" json:"endpoint"`
	AuthMode    string    `gorm:"column:auth_mode;size:64" json:"auth_mode"`
	Config      string    `gorm:"column:config;type:longtext" json:"config"`
	Enabled     bool      `gorm:"not null;default:true;column:enabled" json:"enabled"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaMeshGatewayConfig) TableName() string { return "kafka_mesh_gateway_configs" }

type KafkaCostRecord struct {
	ID             uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID      uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	MetricDate     time.Time `gorm:"column:metric_date" json:"metric_date"`
	StorageBytes   float64   `gorm:"column:storage_bytes" json:"storage_bytes"`
	IngressBytes   float64   `gorm:"column:ingress_bytes" json:"ingress_bytes"`
	EgressBytes    float64   `gorm:"column:egress_bytes" json:"egress_bytes"`
	PartitionCount int       `gorm:"column:partition_count" json:"partition_count"`
	EstimatedCost  float64   `gorm:"column:estimated_cost" json:"estimated_cost"`
	Currency       string    `gorm:"column:currency;size:16" json:"currency"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
}

func (KafkaCostRecord) TableName() string { return "kafka_cost_records" }

type KafkaSensitiveScanRule struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID    uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	Name         string    `gorm:"not null;column:name;size:191" json:"name"`
	PatternType  string    `gorm:"not null;column:pattern_type;size:64" json:"pattern_type"`
	PatternValue string    `gorm:"not null;column:pattern_value;size:500" json:"pattern_value"`
	Severity     string    `gorm:"column:severity;size:32" json:"severity"`
	Enabled      bool      `gorm:"not null;default:true;column:enabled" json:"enabled"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaSensitiveScanRule) TableName() string { return "kafka_sensitive_scan_rules" }

type KafkaSensitiveScanResult struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID   uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	Topic       string    `gorm:"not null;column:topic;size:255" json:"topic"`
	Partition   int32     `gorm:"column:partition" json:"partition"`
	Offset      int64     `gorm:"column:offset_value" json:"offset"`
	RuleName    string    `gorm:"column:rule_name;size:191" json:"rule_name"`
	Severity    string    `gorm:"column:severity;size:32" json:"severity"`
	MatchedText string    `gorm:"column:matched_text;size:500" json:"matched_text"`
	Summary     string    `gorm:"column:summary;type:text" json:"summary"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

func (KafkaSensitiveScanResult) TableName() string { return "kafka_sensitive_scan_results" }
