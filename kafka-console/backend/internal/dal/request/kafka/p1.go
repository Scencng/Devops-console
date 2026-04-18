package kafka

type HealthOverviewRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId" binding:"required"`
}

type TrendRangeRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId" binding:"required"`
	Preset    string `form:"preset" json:"preset" binding:"omitempty,max=64"`
	Start     string `form:"start" json:"start" binding:"omitempty"`
	End       string `form:"end" json:"end" binding:"omitempty"`
	Step      string `form:"step" json:"step" binding:"omitempty,max=32"`
}

type AlertRuleListRequest struct {
	ClusterID   uint   `form:"clusterId" json:"clusterId"`
	Environment string `form:"environment" json:"environment" binding:"omitempty,max=64"`
	Tenant      string `form:"tenant" json:"tenant" binding:"omitempty,max=64"`
}

type AlertRuleUpsertRequest struct {
	ClusterID   uint    `json:"clusterId" binding:"required"`
	Name        string  `json:"name" binding:"required,max=191"`
	MetricType  string  `json:"metricType" binding:"required,max=64"`
	Severity    string  `json:"severity" binding:"required,max=32"`
	Threshold   float64 `json:"threshold" binding:"required"`
	Operator    string  `json:"operator" binding:"omitempty,max=8"`
	Enabled     bool    `json:"enabled"`
	Runbook     string  `json:"runbook"`
	Environment string  `json:"environment" binding:"omitempty,max=64"`
	Tenant      string  `json:"tenant" binding:"omitempty,max=64"`
}

type AlertEventListRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId"`
	Status    string `form:"status" json:"status" binding:"omitempty,max=32"`
	Severity  string `form:"severity" json:"severity" binding:"omitempty,max=32"`
}

type AlertEventStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=acked resolved"`
}

type InspectionRunRequest struct {
	ClusterID uint   `json:"clusterId" binding:"required"`
	Name      string `json:"name" binding:"omitempty,max=191"`
}

type InspectionListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId"`
}

type TaskListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId"`
}

type TaskUpsertRequest struct {
	ClusterID   uint   `json:"clusterId" binding:"required"`
	Name        string `json:"name" binding:"required,max=191"`
	TaskType    string `json:"taskType" binding:"required,max=64"`
	Payload     string `json:"payload"`
	CronExpr    string `json:"cronExpr" binding:"omitempty,max=128"`
	Enabled     bool   `json:"enabled"`
	Environment string `json:"environment" binding:"omitempty,max=64"`
	Tenant      string `json:"tenant" binding:"omitempty,max=64"`
}

type TaskRunRequest struct {
	TriggerMode string `json:"triggerMode" binding:"omitempty,max=32"`
}

type TaskRunListRequest struct {
	TaskID uint `form:"taskId" json:"taskId"`
}

type ChangeRequestListRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId"`
	Status    string `form:"status" json:"status" binding:"omitempty,max=32"`
}

type ChangeRequestCreateRequest struct {
	ClusterID    uint   `json:"clusterId" binding:"required"`
	ChangeType   string `json:"changeType" binding:"required,max=64"`
	ResourceType string `json:"resourceType" binding:"required,max=64"`
	ResourceName string `json:"resourceName" binding:"required,max=255"`
	Payload      string `json:"payload"`
	Reason       string `json:"reason"`
}

type ChangeRequestReviewRequest struct {
	Status  string `json:"status" binding:"required,oneof=approved rejected"`
	Comment string `json:"comment"`
}

type TopicMetadataListRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId"`
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=255"`
}

type TopicMetadataUpsertRequest struct {
	ClusterID   uint   `json:"clusterId" binding:"required"`
	TopicName   string `json:"topicName" binding:"required,max=255"`
	SystemName  string `json:"systemName" binding:"omitempty,max=128"`
	Owner       string `json:"owner" binding:"omitempty,max=128"`
	OwnerEmail  string `json:"ownerEmail" binding:"omitempty,max=191"`
	Environment string `json:"environment" binding:"omitempty,max=64"`
	Tenant      string `json:"tenant" binding:"omitempty,max=64"`
	Lifecycle   string `json:"lifecycle" binding:"omitempty,max=64"`
	Sensitivity string `json:"sensitivity" binding:"omitempty,max=64"`
	Description string `json:"description"`
	Labels      string `json:"labels"`
}

type SchemaRegistryListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId"`
}

type SchemaRegistryUpsertRequest struct {
	ClusterID   uint   `json:"clusterId" binding:"required"`
	Name        string `json:"name" binding:"required,max=191"`
	Endpoint    string `json:"endpoint" binding:"required,max=500"`
	AuthType    string `json:"authType" binding:"omitempty,max=32"`
	Username    string `json:"username" binding:"omitempty,max=191"`
	Password    string `json:"password"`
	VerifySSL   bool   `json:"verifySsl"`
	Environment string `json:"environment" binding:"omitempty,max=64"`
	Tenant      string `json:"tenant" binding:"omitempty,max=64"`
	Description string `json:"description"`
}

type SchemaSubjectListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId" binding:"required"`
}

type SchemaVersionListRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId" binding:"required"`
	Subject   string `form:"subject" json:"subject" binding:"required,max=255"`
}

type SchemaDetailRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId" binding:"required"`
	Subject   string `form:"subject" json:"subject" binding:"required,max=255"`
	Version   string `form:"version" json:"version" binding:"required,max=64"`
}

type SchemaCompatibilityRequest struct {
	ClusterID uint   `json:"clusterId" binding:"required"`
	Subject   string `json:"subject" binding:"required,max=255"`
	Version   string `json:"version" binding:"required,max=64"`
	Schema    string `json:"schema" binding:"required"`
}
