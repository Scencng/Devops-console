package dal

import "time"

type KafkaAuditLog struct {
	ID               uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClusterID        uint      `gorm:"column:cluster_id;not null;index" json:"clusterId"`
	Action           string    `gorm:"column:action;size:64;not null;index" json:"action"`
	ResourceType     string    `gorm:"column:resource_type;size:64;not null;index" json:"resourceType"`
	ResourceName     string    `gorm:"column:resource_name;size:255;not null" json:"resourceName"`
	OperatorUserID   uint64    `gorm:"column:operator_user_id;not null;index" json:"operatorUserId"`
	OperatorUsername string    `gorm:"column:operator_username;size:128;not null" json:"operatorUsername"`
	RequestPayload   string    `gorm:"column:request_payload;type:longtext" json:"requestPayload"`
	Result           string    `gorm:"column:result;size:32;not null;index" json:"result"`
	ErrorMessage     string    `gorm:"column:error_message;type:text" json:"errorMessage"`
	CreatedAt        time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (KafkaAuditLog) TableName() string {
	return "kafka_audit_logs"
}
