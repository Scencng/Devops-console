package response

import "time"

type KafkaHealthCardVO struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type KafkaHealthIssueVO struct {
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

type KafkaHealthOverviewVO struct {
	Cards  []KafkaHealthCardVO  `json:"cards"`
	Issues []KafkaHealthIssueVO `json:"issues"`
}

type KafkaTrendSeriesVO struct {
	Name   string                 `json:"name"`
	Query  string                 `json:"query"`
	Series []KafkaPrometheusSeriesVO `json:"series"`
}

type KafkaAlertRuleVO struct {
	ID          uint      `json:"id"`
	ClusterID   uint      `json:"clusterId"`
	Name        string    `json:"name"`
	MetricType  string    `json:"metricType"`
	Severity    string    `json:"severity"`
	Threshold   float64   `json:"threshold"`
	Operator    string    `json:"operator"`
	Enabled     bool      `json:"enabled"`
	Runbook     string    `json:"runbook"`
	Environment string    `json:"environment"`
	Tenant      string    `json:"tenant"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type KafkaAlertEventVO struct {
	ID               uint       `json:"id"`
	ClusterID        uint       `json:"clusterId"`
	RuleID           *uint      `json:"ruleId"`
	Title            string     `json:"title"`
	Severity         string     `json:"severity"`
	Status           string     `json:"status"`
	MetricType       string     `json:"metricType"`
	MetricValue      float64    `json:"metricValue"`
	Threshold        float64    `json:"threshold"`
	Message          string     `json:"message"`
	Runbook          string     `json:"runbook"`
	Environment      string     `json:"environment"`
	Tenant           string     `json:"tenant"`
	AckedBy          string     `json:"ackedBy"`
	ResolvedBy       string     `json:"resolvedBy"`
	FirstTriggeredAt time.Time  `json:"firstTriggeredAt"`
	LastTriggeredAt  time.Time  `json:"lastTriggeredAt"`
	AckedAt          *time.Time `json:"ackedAt"`
	ResolvedAt       *time.Time `json:"resolvedAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type KafkaInspectionItemVO struct {
	ID        uint      `json:"id"`
	CheckCode string    `json:"checkCode"`
	Severity  string    `json:"severity"`
	Status    string    `json:"status"`
	Title     string    `json:"title"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}

type KafkaInspectionReportVO struct {
	ID          uint                  `json:"id"`
	ClusterID   uint                  `json:"clusterId"`
	Name        string                `json:"name"`
	Status      string                `json:"status"`
	Summary     string                `json:"summary"`
	IssueCount  int                   `json:"issueCount"`
	Environment string                `json:"environment"`
	Tenant      string                `json:"tenant"`
	TriggeredBy string                `json:"triggeredBy"`
	ExecutedAt  time.Time             `json:"executedAt"`
	CreatedAt   time.Time             `json:"createdAt"`
	Items       []KafkaInspectionItemVO `json:"items,omitempty"`
}

type KafkaTaskVO struct {
	ID            uint      `json:"id"`
	ClusterID     uint      `json:"clusterId"`
	Name          string    `json:"name"`
	TaskType      string    `json:"taskType"`
	Payload       string    `json:"payload"`
	CronExpr      string    `json:"cronExpr"`
	Enabled       bool      `json:"enabled"`
	LastRunStatus string    `json:"lastRunStatus"`
	Environment   string    `json:"environment"`
	Tenant        string    `json:"tenant"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type KafkaTaskRunVO struct {
	ID            uint       `json:"id"`
	TaskID        uint       `json:"taskId"`
	ClusterID     uint       `json:"clusterId"`
	Status        string     `json:"status"`
	TriggerMode   string     `json:"triggerMode"`
	ResultSummary string     `json:"resultSummary"`
	ResultPayload string     `json:"resultPayload"`
	StartedAt     time.Time  `json:"startedAt"`
	FinishedAt    *time.Time `json:"finishedAt"`
}

type KafkaChangeRequestVO struct {
	ID                uint       `json:"id"`
	ClusterID         uint       `json:"clusterId"`
	ChangeType        string     `json:"changeType"`
	ResourceType      string     `json:"resourceType"`
	ResourceName      string     `json:"resourceName"`
	Payload           string     `json:"payload"`
	Reason            string     `json:"reason"`
	Status            string     `json:"status"`
	RequesterUserID   uint64     `json:"requesterUserId"`
	RequesterUsername string     `json:"requesterUsername"`
	ApproverUserID    uint64     `json:"approverUserId"`
	ApproverUsername  string     `json:"approverUsername"`
	ApprovalComment   string     `json:"approvalComment"`
	Environment       string     `json:"environment"`
	Tenant            string     `json:"tenant"`
	ExecutedAt        *time.Time `json:"executedAt"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type KafkaTopicMetadataVO struct {
	ID          uint      `json:"id"`
	ClusterID   uint      `json:"clusterId"`
	TopicName   string    `json:"topicName"`
	SystemName  string    `json:"systemName"`
	Owner       string    `json:"owner"`
	OwnerEmail  string    `json:"ownerEmail"`
	Environment string    `json:"environment"`
	Tenant      string    `json:"tenant"`
	Lifecycle   string    `json:"lifecycle"`
	Sensitivity string    `json:"sensitivity"`
	Description string    `json:"description"`
	Labels      string    `json:"labels"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type KafkaSchemaRegistryVO struct {
	ID          uint      `json:"id"`
	ClusterID   uint      `json:"clusterId"`
	Name        string    `json:"name"`
	Endpoint    string    `json:"endpoint"`
	AuthType    string    `json:"authType"`
	Username    string    `json:"username"`
	VerifySSL   bool      `json:"verifySsl"`
	Environment string    `json:"environment"`
	Tenant      string    `json:"tenant"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type KafkaSchemaSubjectVO struct {
	Subject string `json:"subject"`
}

type KafkaSchemaVersionVO struct {
	Version int `json:"version"`
}

type KafkaSchemaDetailVO struct {
	Subject      string      `json:"subject"`
	Version      int         `json:"version"`
	ID           int         `json:"id"`
	SchemaType   string      `json:"schemaType"`
	Schema       string      `json:"schema"`
	References   interface{} `json:"references"`
	Compatibility string     `json:"compatibility,omitempty"`
}

type KafkaSchemaCompatibilityVO struct {
	IsCompatible bool   `json:"isCompatible"`
	Message      string `json:"message"`
}
