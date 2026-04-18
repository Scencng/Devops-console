package configs

import (
	"time"

	"devops-console-backend/internal/dal"
)

type KafkaClusterRepository struct{}

func NewKafkaClusterRepository() *KafkaClusterRepository {
	return &KafkaClusterRepository{}
}

func (r *KafkaClusterRepository) List(page, pageSize int, keyword, status, environment, tenant string) ([]dal.KafkaCluster, int64, error) {
	var (
		list  []dal.KafkaCluster
		total int64
	)

	query := GORMDB.Model(&dal.KafkaCluster{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR bootstrap_servers LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if environment != "" {
		query = query.Where("environment = ?", environment)
	}
	if tenant != "" {
		query = query.Where("tenant = ?", tenant)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *KafkaClusterRepository) ListAll() ([]dal.KafkaCluster, error) {
	var list []dal.KafkaCluster
	err := GORMDB.Order("name ASC").Find(&list).Error
	return list, err
}

func (r *KafkaClusterRepository) GetByID(id uint) (*dal.KafkaCluster, error) {
	var item dal.KafkaCluster
	err := GORMDB.Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *KafkaClusterRepository) GetByName(name string) (*dal.KafkaCluster, error) {
	var item dal.KafkaCluster
	err := GORMDB.Where("name = ?", name).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *KafkaClusterRepository) Create(item *dal.KafkaCluster) error {
	return GORMDB.Create(item).Error
}

func (r *KafkaClusterRepository) Update(item *dal.KafkaCluster) error {
	return GORMDB.Save(item).Error
}

func (r *KafkaClusterRepository) Delete(id uint) error {
	return GORMDB.Delete(&dal.KafkaCluster{}, id).Error
}

func (r *KafkaClusterRepository) UpdateTestStatus(id uint, status, message string, testedAt *time.Time) error {
	return GORMDB.Model(&dal.KafkaCluster{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":             status,
		"last_error_message": message,
		"last_tested_at":     testedAt,
	}).Error
}

func (r *KafkaClusterRepository) SaveTestRecord(record *dal.ConnectionTest) error {
	return GORMDB.Create(record).Error
}
