package configs

import (
	"devops-console-backend/internal/dal"

	"gorm.io/gorm"
)

type KafkaAlertRuleRepository struct{}

func NewKafkaAlertRuleRepository() *KafkaAlertRuleRepository { return &KafkaAlertRuleRepository{} }

func (r *KafkaAlertRuleRepository) List(clusterID uint, environment, tenant string) ([]dal.KafkaAlertRule, error) {
	var list []dal.KafkaAlertRule
	query := GORMDB.Model(&dal.KafkaAlertRule{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if environment != "" {
		query = query.Where("environment = ?", environment)
	}
	if tenant != "" {
		query = query.Where("tenant = ?", tenant)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *KafkaAlertRuleRepository) ListEnabled(clusterID uint) ([]dal.KafkaAlertRule, error) {
	var list []dal.KafkaAlertRule
	err := GORMDB.Where("cluster_id = ? AND enabled = ?", clusterID, true).Order("id DESC").Find(&list).Error
	return list, err
}

func (r *KafkaAlertRuleRepository) GetByID(id uint) (*dal.KafkaAlertRule, error) {
	var item dal.KafkaAlertRule
	err := GORMDB.Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *KafkaAlertRuleRepository) Create(item *dal.KafkaAlertRule) error { return GORMDB.Create(item).Error }
func (r *KafkaAlertRuleRepository) Update(item *dal.KafkaAlertRule) error { return GORMDB.Save(item).Error }
func (r *KafkaAlertRuleRepository) Delete(id uint) error                  { return GORMDB.Delete(&dal.KafkaAlertRule{}, id).Error }

type KafkaAlertEventRepository struct{}

func NewKafkaAlertEventRepository() *KafkaAlertEventRepository { return &KafkaAlertEventRepository{} }

func (r *KafkaAlertEventRepository) List(clusterID uint, status, severity string) ([]dal.KafkaAlertEvent, error) {
	var list []dal.KafkaAlertEvent
	query := GORMDB.Model(&dal.KafkaAlertEvent{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *KafkaAlertEventRepository) CountOpen(clusterID uint) (int64, error) {
	var total int64
	err := GORMDB.Model(&dal.KafkaAlertEvent{}).Where("cluster_id = ? AND status IN ?", clusterID, []string{"open", "acked"}).Count(&total).Error
	return total, err
}

func (r *KafkaAlertEventRepository) GetByID(id uint) (*dal.KafkaAlertEvent, error) {
	var item dal.KafkaAlertEvent
	err := GORMDB.Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *KafkaAlertEventRepository) Create(item *dal.KafkaAlertEvent) error { return GORMDB.Create(item).Error }
func (r *KafkaAlertEventRepository) Update(item *dal.KafkaAlertEvent) error { return GORMDB.Save(item).Error }
func (r *KafkaAlertEventRepository) GetLatestOpenByRule(ruleID uint) (*dal.KafkaAlertEvent, error) {
	var item dal.KafkaAlertEvent
	err := GORMDB.Where("rule_id = ? AND status IN ?", ruleID, []string{"open", "acked"}).Order("id DESC").First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

type KafkaInspectionRepository struct{}

func NewKafkaInspectionRepository() *KafkaInspectionRepository { return &KafkaInspectionRepository{} }

func (r *KafkaInspectionRepository) CreateReportWithItems(report *dal.KafkaInspectionReport, items []dal.KafkaInspectionItem) error {
	return GORMDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(report).Error; err != nil {
			return err
		}
		if len(items) > 0 {
			for index := range items {
				items[index].ReportID = report.ID
			}
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *KafkaInspectionRepository) ListReports(clusterID uint) ([]dal.KafkaInspectionReport, error) {
	var list []dal.KafkaInspectionReport
	query := GORMDB.Model(&dal.KafkaInspectionReport{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *KafkaInspectionRepository) GetReportByID(id uint) (*dal.KafkaInspectionReport, error) {
	var item dal.KafkaInspectionReport
	err := GORMDB.Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *KafkaInspectionRepository) GetItemsByReportID(reportID uint) ([]dal.KafkaInspectionItem, error) {
	var list []dal.KafkaInspectionItem
	err := GORMDB.Where("report_id = ?", reportID).Order("id ASC").Find(&list).Error
	return list, err
}

type KafkaTaskRepository struct{}

func NewKafkaTaskRepository() *KafkaTaskRepository { return &KafkaTaskRepository{} }

func (r *KafkaTaskRepository) List(clusterID uint) ([]dal.KafkaTask, error) {
	var list []dal.KafkaTask
	query := GORMDB.Model(&dal.KafkaTask{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *KafkaTaskRepository) GetByID(id uint) (*dal.KafkaTask, error) {
	var item dal.KafkaTask
	err := GORMDB.Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *KafkaTaskRepository) Create(item *dal.KafkaTask) error { return GORMDB.Create(item).Error }
func (r *KafkaTaskRepository) Update(item *dal.KafkaTask) error { return GORMDB.Save(item).Error }
func (r *KafkaTaskRepository) Delete(id uint) error             { return GORMDB.Delete(&dal.KafkaTask{}, id).Error }

func (r *KafkaTaskRepository) CreateRun(item *dal.KafkaTaskRun) error { return GORMDB.Create(item).Error }
func (r *KafkaTaskRepository) UpdateRun(item *dal.KafkaTaskRun) error { return GORMDB.Save(item).Error }

func (r *KafkaTaskRepository) ListRuns(taskID uint) ([]dal.KafkaTaskRun, error) {
	var list []dal.KafkaTaskRun
	query := GORMDB.Model(&dal.KafkaTaskRun{})
	if taskID > 0 {
		query = query.Where("task_id = ?", taskID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}

type KafkaChangeRequestRepository struct{}

func NewKafkaChangeRequestRepository() *KafkaChangeRequestRepository { return &KafkaChangeRequestRepository{} }

func (r *KafkaChangeRequestRepository) List(clusterID uint, status string) ([]dal.KafkaChangeRequest, error) {
	var list []dal.KafkaChangeRequest
	query := GORMDB.Model(&dal.KafkaChangeRequest{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *KafkaChangeRequestRepository) GetByID(id uint) (*dal.KafkaChangeRequest, error) {
	var item dal.KafkaChangeRequest
	err := GORMDB.Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *KafkaChangeRequestRepository) Create(item *dal.KafkaChangeRequest) error { return GORMDB.Create(item).Error }
func (r *KafkaChangeRequestRepository) Update(item *dal.KafkaChangeRequest) error { return GORMDB.Save(item).Error }

type KafkaTopicMetadataRepository struct{}

func NewKafkaTopicMetadataRepository() *KafkaTopicMetadataRepository { return &KafkaTopicMetadataRepository{} }

func (r *KafkaTopicMetadataRepository) List(clusterID uint, keyword string) ([]dal.KafkaTopicMetadata, error) {
	var list []dal.KafkaTopicMetadata
	query := GORMDB.Model(&dal.KafkaTopicMetadata{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if keyword != "" {
		query = query.Where("topic_name LIKE ? OR system_name LIKE ? OR owner LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	err := query.Order("topic_name ASC").Find(&list).Error
	return list, err
}

func (r *KafkaTopicMetadataRepository) GetByID(id uint) (*dal.KafkaTopicMetadata, error) {
	var item dal.KafkaTopicMetadata
	err := GORMDB.Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *KafkaTopicMetadataRepository) GetByClusterTopic(clusterID uint, topic string) (*dal.KafkaTopicMetadata, error) {
	var item dal.KafkaTopicMetadata
	err := GORMDB.Where("cluster_id = ? AND topic_name = ?", clusterID, topic).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *KafkaTopicMetadataRepository) Create(item *dal.KafkaTopicMetadata) error { return GORMDB.Create(item).Error }
func (r *KafkaTopicMetadataRepository) Update(item *dal.KafkaTopicMetadata) error { return GORMDB.Save(item).Error }
func (r *KafkaTopicMetadataRepository) Delete(id uint) error                      { return GORMDB.Delete(&dal.KafkaTopicMetadata{}, id).Error }

type KafkaSchemaRegistryRepository struct{}

func NewKafkaSchemaRegistryRepository() *KafkaSchemaRegistryRepository {
	return &KafkaSchemaRegistryRepository{}
}

func (r *KafkaSchemaRegistryRepository) List(clusterID uint) ([]dal.KafkaSchemaRegistry, error) {
	var list []dal.KafkaSchemaRegistry
	query := GORMDB.Model(&dal.KafkaSchemaRegistry{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *KafkaSchemaRegistryRepository) GetByID(id uint) (*dal.KafkaSchemaRegistry, error) {
	var item dal.KafkaSchemaRegistry
	err := GORMDB.Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *KafkaSchemaRegistryRepository) GetFirstByCluster(clusterID uint) (*dal.KafkaSchemaRegistry, error) {
	var item dal.KafkaSchemaRegistry
	err := GORMDB.Where("cluster_id = ?", clusterID).Order("id DESC").First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *KafkaSchemaRegistryRepository) Create(item *dal.KafkaSchemaRegistry) error { return GORMDB.Create(item).Error }
func (r *KafkaSchemaRegistryRepository) Update(item *dal.KafkaSchemaRegistry) error { return GORMDB.Save(item).Error }
func (r *KafkaSchemaRegistryRepository) Delete(id uint) error                       { return GORMDB.Delete(&dal.KafkaSchemaRegistry{}, id).Error }
