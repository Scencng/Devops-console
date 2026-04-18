package configs

import "devops-console-backend/internal/dal"

type KafkaAuditLogRepository struct{}

func NewKafkaAuditLogRepository() *KafkaAuditLogRepository {
	return &KafkaAuditLogRepository{}
}

func (r *KafkaAuditLogRepository) Create(log *dal.KafkaAuditLog) error {
	return GORMDB.Create(log).Error
}

func (r *KafkaAuditLogRepository) List(clusterID uint, action, result string, page, pageSize int) ([]dal.KafkaAuditLog, int64, error) {
	var (
		list []dal.KafkaAuditLog
		total int64
	)
	query := GORMDB.Model(&dal.KafkaAuditLog{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if result != "" {
		query = query.Where("result = ?", result)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
