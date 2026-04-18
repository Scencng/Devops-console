package dal

import "time"

type KafkaAlertRule struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID   uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	Name        string    `gorm:"not null;column:name;size:191" json:"name"`
	MetricType  string    `gorm:"not null;column:metric_type;size:64" json:"metric_type"`
	Severity    string    `gorm:"not null;default:'warning';column:severity;size:32" json:"severity"`
	Threshold   float64   `gorm:"not null;default:0;column:threshold" json:"threshold"`
	Operator    string    `gorm:"not null;default:'>';column:operator;size:8" json:"operator"`
	Enabled     bool      `gorm:"not null;default:true;column:enabled" json:"enabled"`
	Runbook     string    `gorm:"column:runbook;type:text" json:"runbook"`
	Environment string    `gorm:"column:environment;size:64" json:"environment"`
	Tenant      string    `gorm:"column:tenant;size:64" json:"tenant"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaAlertRule) TableName() string { return "kafka_alert_rules" }

type KafkaAlertEvent struct {
	ID              uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID       uint       `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	RuleID          *uint      `gorm:"index;column:rule_id" json:"rule_id"`
	Title           string     `gorm:"not null;column:title;size:255" json:"title"`
	Severity        string     `gorm:"not null;default:'warning';column:severity;size:32" json:"severity"`
	Status          string     `gorm:"not null;default:'open';index;column:status;size:32" json:"status"`
	MetricType      string     `gorm:"not null;column:metric_type;size:64" json:"metric_type"`
	MetricValue     float64    `gorm:"not null;default:0;column:metric_value" json:"metric_value"`
	Threshold       float64    `gorm:"not null;default:0;column:threshold" json:"threshold"`
	Message         string     `gorm:"column:message;type:text" json:"message"`
	Runbook         string     `gorm:"column:runbook;type:text" json:"runbook"`
	Environment     string     `gorm:"column:environment;size:64" json:"environment"`
	Tenant          string     `gorm:"column:tenant;size:64" json:"tenant"`
	AckedBy         string     `gorm:"column:acked_by;size:128" json:"acked_by"`
	ResolvedBy      string     `gorm:"column:resolved_by;size:128" json:"resolved_by"`
	FirstTriggeredAt time.Time `gorm:"column:first_triggered_at" json:"first_triggered_at"`
	LastTriggeredAt time.Time  `gorm:"column:last_triggered_at" json:"last_triggered_at"`
	AckedAt         *time.Time `gorm:"column:acked_at" json:"acked_at"`
	ResolvedAt      *time.Time `gorm:"column:resolved_at" json:"resolved_at"`
	CreatedAt       time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaAlertEvent) TableName() string { return "kafka_alert_events" }

type KafkaInspectionReport struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID    uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	Name         string    `gorm:"not null;column:name;size:191" json:"name"`
	Status       string    `gorm:"not null;default:'success';column:status;size:32" json:"status"`
	Summary      string    `gorm:"column:summary;type:text" json:"summary"`
	IssueCount   int       `gorm:"not null;default:0;column:issue_count" json:"issue_count"`
	Environment  string    `gorm:"column:environment;size:64" json:"environment"`
	Tenant       string    `gorm:"column:tenant;size:64" json:"tenant"`
	TriggeredBy  string    `gorm:"column:triggered_by;size:128" json:"triggered_by"`
	ExecutedAt   time.Time `gorm:"column:executed_at" json:"executed_at"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

func (KafkaInspectionReport) TableName() string { return "kafka_inspection_reports" }

type KafkaInspectionItem struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ReportID  uint      `gorm:"index;not null;column:report_id" json:"report_id"`
	CheckCode string    `gorm:"not null;column:check_code;size:64" json:"check_code"`
	Severity  string    `gorm:"not null;default:'info';column:severity;size:32" json:"severity"`
	Status    string    `gorm:"not null;default:'ok';column:status;size:32" json:"status"`
	Title     string    `gorm:"not null;column:title;size:255" json:"title"`
	Detail    string    `gorm:"column:detail;type:text" json:"detail"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (KafkaInspectionItem) TableName() string { return "kafka_inspection_items" }

type KafkaTask struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID   uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	Name        string    `gorm:"not null;column:name;size:191" json:"name"`
	TaskType    string    `gorm:"not null;column:task_type;size:64" json:"task_type"`
	Payload     string    `gorm:"column:payload;type:longtext" json:"payload"`
	CronExpr    string    `gorm:"column:cron_expr;size:128" json:"cron_expr"`
	Enabled     bool      `gorm:"not null;default:true;column:enabled" json:"enabled"`
	LastRunStatus string  `gorm:"column:last_run_status;size:32" json:"last_run_status"`
	Environment string    `gorm:"column:environment;size:64" json:"environment"`
	Tenant      string    `gorm:"column:tenant;size:64" json:"tenant"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaTask) TableName() string { return "kafka_tasks" }

type KafkaTaskRun struct {
	ID            uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	TaskID        uint       `gorm:"index;not null;column:task_id" json:"task_id"`
	ClusterID     uint       `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	Status        string     `gorm:"not null;default:'running';column:status;size:32" json:"status"`
	TriggerMode   string     `gorm:"not null;default:'manual';column:trigger_mode;size:32" json:"trigger_mode"`
	ResultSummary string     `gorm:"column:result_summary;type:text" json:"result_summary"`
	ResultPayload string     `gorm:"column:result_payload;type:longtext" json:"result_payload"`
	StartedAt     time.Time  `gorm:"column:started_at" json:"started_at"`
	FinishedAt    *time.Time `gorm:"column:finished_at" json:"finished_at"`
}

func (KafkaTaskRun) TableName() string { return "kafka_task_runs" }

type KafkaChangeRequest struct {
	ID               uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID        uint       `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	ChangeType       string     `gorm:"not null;column:change_type;size:64" json:"change_type"`
	ResourceType     string     `gorm:"not null;column:resource_type;size:64" json:"resource_type"`
	ResourceName     string     `gorm:"not null;column:resource_name;size:255" json:"resource_name"`
	Payload          string     `gorm:"column:payload;type:longtext" json:"payload"`
	Reason           string     `gorm:"column:reason;type:text" json:"reason"`
	Status           string     `gorm:"not null;default:'pending';index;column:status;size:32" json:"status"`
	RequesterUserID  uint64     `gorm:"column:requester_user_id" json:"requester_user_id"`
	RequesterUsername string    `gorm:"column:requester_username;size:128" json:"requester_username"`
	ApproverUserID   uint64     `gorm:"column:approver_user_id" json:"approver_user_id"`
	ApproverUsername string     `gorm:"column:approver_username;size:128" json:"approver_username"`
	ApprovalComment  string     `gorm:"column:approval_comment;type:text" json:"approval_comment"`
	Environment      string     `gorm:"column:environment;size:64" json:"environment"`
	Tenant           string     `gorm:"column:tenant;size:64" json:"tenant"`
	ExecutedAt       *time.Time `gorm:"column:executed_at" json:"executed_at"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaChangeRequest) TableName() string { return "kafka_change_requests" }

type KafkaTopicMetadata struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID   uint      `gorm:"not null;index:uk_cluster_topic,unique;column:cluster_id" json:"cluster_id"`
	TopicName   string    `gorm:"not null;size:255;index:uk_cluster_topic,unique;column:topic_name" json:"topic_name"`
	SystemName  string    `gorm:"column:system_name;size:128" json:"system_name"`
	Owner       string    `gorm:"column:owner;size:128" json:"owner"`
	OwnerEmail  string    `gorm:"column:owner_email;size:191" json:"owner_email"`
	Environment string    `gorm:"column:environment;size:64" json:"environment"`
	Tenant      string    `gorm:"column:tenant;size:64" json:"tenant"`
	Lifecycle   string    `gorm:"column:lifecycle;size:64" json:"lifecycle"`
	Sensitivity string    `gorm:"column:sensitivity;size:64" json:"sensitivity"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	Labels      string    `gorm:"column:labels;type:longtext" json:"labels"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaTopicMetadata) TableName() string { return "kafka_topic_metadata" }

type KafkaSchemaRegistry struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID          uint      `gorm:"index;not null;column:cluster_id" json:"cluster_id"`
	Name               string    `gorm:"not null;column:name;size:191" json:"name"`
	Endpoint           string    `gorm:"not null;column:endpoint;size:500" json:"endpoint"`
	AuthType           string    `gorm:"not null;default:'none';column:auth_type;size:32" json:"auth_type"`
	Username           string    `gorm:"column:username;size:191" json:"username"`
	PasswordCiphertext string    `gorm:"column:password_ciphertext;type:text" json:"-"`
	VerifySSL          bool      `gorm:"not null;default:true;column:verify_ssl" json:"verify_ssl"`
	Environment        string    `gorm:"column:environment;size:64" json:"environment"`
	Tenant             string    `gorm:"column:tenant;size:64" json:"tenant"`
	Description        string    `gorm:"column:description;type:text" json:"description"`
	CreatedAt          time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaSchemaRegistry) TableName() string { return "kafka_schema_registries" }
