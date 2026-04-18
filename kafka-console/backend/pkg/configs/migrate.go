package configs

import "devops-console-backend/internal/dal"

func AutoMigrateCustomTables() error {
	if Config == nil || !Config.Database.AutoMigrate || GORMDB == nil {
		return nil
	}
	return GORMDB.AutoMigrate(
		&dal.KafkaCluster{},
		&dal.KafkaAuditLog{},
		&dal.KafkaAlertRule{},
		&dal.KafkaAlertEvent{},
		&dal.KafkaInspectionReport{},
		&dal.KafkaInspectionItem{},
		&dal.KafkaTask{},
		&dal.KafkaTaskRun{},
		&dal.KafkaChangeRequest{},
		&dal.KafkaTopicMetadata{},
		&dal.KafkaSchemaRegistry{},
		&dal.KafkaAlertSilence{},
		&dal.KafkaTraceLink{},
		&dal.KafkaScalingRecommendation{},
		&dal.KafkaSelfHealingPolicy{},
		&dal.KafkaSelfHealingExecution{},
		&dal.KafkaGitOpsProfile{},
		&dal.KafkaGitOpsSyncRecord{},
		&dal.KafkaCloudAdapter{},
		&dal.KafkaLineageRelation{},
		&dal.KafkaLifecyclePolicy{},
		&dal.KafkaMeshGatewayConfig{},
		&dal.KafkaCostRecord{},
		&dal.KafkaSensitiveScanRule{},
		&dal.KafkaSensitiveScanResult{},
	)
}
