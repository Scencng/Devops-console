package configs

import "devops-console-backend/internal/dal"

func AutoMigrateCustomTables() error {
	if Config == nil || !Config.Database.AutoMigrate || GORMDB == nil {
		return nil
	}
	return GORMDB.AutoMigrate(
		&dal.KafkaCluster{},
		&dal.KafkaAuditLog{},
	)
}
