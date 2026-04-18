package configs

import (
	"devops-console-backend/internal/dal"

	"gorm.io/gorm"
)

type KafkaAlertSilenceRepository struct{}

func NewKafkaAlertSilenceRepository() *KafkaAlertSilenceRepository { return &KafkaAlertSilenceRepository{} }
func (r *KafkaAlertSilenceRepository) List(clusterID uint) ([]dal.KafkaAlertSilence, error) {
	var list []dal.KafkaAlertSilence
	query := GORMDB.Model(&dal.KafkaAlertSilence{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}
func (r *KafkaAlertSilenceRepository) GetByID(id uint) (*dal.KafkaAlertSilence, error) {
	var item dal.KafkaAlertSilence
	if err := GORMDB.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
func (r *KafkaAlertSilenceRepository) Create(item *dal.KafkaAlertSilence) error { return GORMDB.Create(item).Error }
func (r *KafkaAlertSilenceRepository) Update(item *dal.KafkaAlertSilence) error { return GORMDB.Save(item).Error }
func (r *KafkaAlertSilenceRepository) Delete(id uint) error                      { return GORMDB.Delete(&dal.KafkaAlertSilence{}, id).Error }

type KafkaTraceLinkRepository struct{}

func NewKafkaTraceLinkRepository() *KafkaTraceLinkRepository { return &KafkaTraceLinkRepository{} }
func (r *KafkaTraceLinkRepository) List(clusterID uint, keyword string) ([]dal.KafkaTraceLink, error) {
	var list []dal.KafkaTraceLink
	query := GORMDB.Model(&dal.KafkaTraceLink{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if keyword != "" {
		query = query.Where("trace_id LIKE ? OR topic LIKE ? OR service_name LIKE ? OR message_key LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}
func (r *KafkaTraceLinkRepository) Create(item *dal.KafkaTraceLink) error { return GORMDB.Create(item).Error }
func (r *KafkaTraceLinkRepository) Delete(id uint) error                   { return GORMDB.Delete(&dal.KafkaTraceLink{}, id).Error }

type KafkaScalingRecommendationRepository struct{}

func NewKafkaScalingRecommendationRepository() *KafkaScalingRecommendationRepository {
	return &KafkaScalingRecommendationRepository{}
}
func (r *KafkaScalingRecommendationRepository) ReplaceForCluster(clusterID uint, items []dal.KafkaScalingRecommendation) error {
	return GORMDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cluster_id = ?", clusterID).Delete(&dal.KafkaScalingRecommendation{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}
func (r *KafkaScalingRecommendationRepository) List(clusterID uint) ([]dal.KafkaScalingRecommendation, error) {
	var list []dal.KafkaScalingRecommendation
	query := GORMDB.Model(&dal.KafkaScalingRecommendation{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}

type KafkaSelfHealingPolicyRepository struct{}

func NewKafkaSelfHealingPolicyRepository() *KafkaSelfHealingPolicyRepository {
	return &KafkaSelfHealingPolicyRepository{}
}
func (r *KafkaSelfHealingPolicyRepository) List(clusterID uint) ([]dal.KafkaSelfHealingPolicy, error) {
	var list []dal.KafkaSelfHealingPolicy
	query := GORMDB.Model(&dal.KafkaSelfHealingPolicy{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}
func (r *KafkaSelfHealingPolicyRepository) GetByID(id uint) (*dal.KafkaSelfHealingPolicy, error) {
	var item dal.KafkaSelfHealingPolicy
	if err := GORMDB.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
func (r *KafkaSelfHealingPolicyRepository) Create(item *dal.KafkaSelfHealingPolicy) error { return GORMDB.Create(item).Error }
func (r *KafkaSelfHealingPolicyRepository) Update(item *dal.KafkaSelfHealingPolicy) error { return GORMDB.Save(item).Error }
func (r *KafkaSelfHealingPolicyRepository) Delete(id uint) error                          { return GORMDB.Delete(&dal.KafkaSelfHealingPolicy{}, id).Error }
func (r *KafkaSelfHealingPolicyRepository) CreateExecution(item *dal.KafkaSelfHealingExecution) error {
	return GORMDB.Create(item).Error
}
func (r *KafkaSelfHealingPolicyRepository) UpdateExecution(item *dal.KafkaSelfHealingExecution) error {
	return GORMDB.Save(item).Error
}
func (r *KafkaSelfHealingPolicyRepository) ListExecutions(clusterID uint) ([]dal.KafkaSelfHealingExecution, error) {
	var list []dal.KafkaSelfHealingExecution
	query := GORMDB.Model(&dal.KafkaSelfHealingExecution{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}

type KafkaGitOpsRepository struct{}

func NewKafkaGitOpsRepository() *KafkaGitOpsRepository { return &KafkaGitOpsRepository{} }
func (r *KafkaGitOpsRepository) ListProfiles(clusterID uint) ([]dal.KafkaGitOpsProfile, error) {
	var list []dal.KafkaGitOpsProfile
	query := GORMDB.Model(&dal.KafkaGitOpsProfile{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}
func (r *KafkaGitOpsRepository) GetProfileByID(id uint) (*dal.KafkaGitOpsProfile, error) {
	var item dal.KafkaGitOpsProfile
	if err := GORMDB.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
func (r *KafkaGitOpsRepository) CreateProfile(item *dal.KafkaGitOpsProfile) error { return GORMDB.Create(item).Error }
func (r *KafkaGitOpsRepository) UpdateProfile(item *dal.KafkaGitOpsProfile) error { return GORMDB.Save(item).Error }
func (r *KafkaGitOpsRepository) DeleteProfile(id uint) error                       { return GORMDB.Delete(&dal.KafkaGitOpsProfile{}, id).Error }
func (r *KafkaGitOpsRepository) CreateSync(item *dal.KafkaGitOpsSyncRecord) error { return GORMDB.Create(item).Error }
func (r *KafkaGitOpsRepository) UpdateSync(item *dal.KafkaGitOpsSyncRecord) error { return GORMDB.Save(item).Error }
func (r *KafkaGitOpsRepository) ListSyncs(profileID uint) ([]dal.KafkaGitOpsSyncRecord, error) {
	var list []dal.KafkaGitOpsSyncRecord
	query := GORMDB.Model(&dal.KafkaGitOpsSyncRecord{})
	if profileID > 0 {
		query = query.Where("profile_id = ?", profileID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}

type KafkaCloudAdapterRepository struct{}

func NewKafkaCloudAdapterRepository() *KafkaCloudAdapterRepository { return &KafkaCloudAdapterRepository{} }
func (r *KafkaCloudAdapterRepository) List(clusterID uint) ([]dal.KafkaCloudAdapter, error) {
	var list []dal.KafkaCloudAdapter
	query := GORMDB.Model(&dal.KafkaCloudAdapter{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}
func (r *KafkaCloudAdapterRepository) GetByID(id uint) (*dal.KafkaCloudAdapter, error) {
	var item dal.KafkaCloudAdapter
	if err := GORMDB.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
func (r *KafkaCloudAdapterRepository) Create(item *dal.KafkaCloudAdapter) error { return GORMDB.Create(item).Error }
func (r *KafkaCloudAdapterRepository) Update(item *dal.KafkaCloudAdapter) error { return GORMDB.Save(item).Error }
func (r *KafkaCloudAdapterRepository) Delete(id uint) error                     { return GORMDB.Delete(&dal.KafkaCloudAdapter{}, id).Error }

type KafkaLineageRepository struct{}

func NewKafkaLineageRepository() *KafkaLineageRepository { return &KafkaLineageRepository{} }
func (r *KafkaLineageRepository) List(clusterID uint, keyword string) ([]dal.KafkaLineageRelation, error) {
	var list []dal.KafkaLineageRelation
	query := GORMDB.Model(&dal.KafkaLineageRelation{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if keyword != "" {
		query = query.Where("source_topic LIKE ? OR target_topic LIKE ? OR producer_service LIKE ? OR consumer_service LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}
func (r *KafkaLineageRepository) Create(item *dal.KafkaLineageRelation) error { return GORMDB.Create(item).Error }
func (r *KafkaLineageRepository) Update(item *dal.KafkaLineageRelation) error { return GORMDB.Save(item).Error }
func (r *KafkaLineageRepository) GetByID(id uint) (*dal.KafkaLineageRelation, error) {
	var item dal.KafkaLineageRelation
	if err := GORMDB.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
func (r *KafkaLineageRepository) Delete(id uint) error { return GORMDB.Delete(&dal.KafkaLineageRelation{}, id).Error }

type KafkaLifecycleRepository struct{}

func NewKafkaLifecycleRepository() *KafkaLifecycleRepository { return &KafkaLifecycleRepository{} }
func (r *KafkaLifecycleRepository) ReplaceForCluster(clusterID uint, items []dal.KafkaLifecyclePolicy) error {
	return GORMDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cluster_id = ?", clusterID).Delete(&dal.KafkaLifecyclePolicy{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}
func (r *KafkaLifecycleRepository) List(clusterID uint) ([]dal.KafkaLifecyclePolicy, error) {
	var list []dal.KafkaLifecyclePolicy
	query := GORMDB.Model(&dal.KafkaLifecyclePolicy{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}

type KafkaMeshGatewayRepository struct{}

func NewKafkaMeshGatewayRepository() *KafkaMeshGatewayRepository { return &KafkaMeshGatewayRepository{} }
func (r *KafkaMeshGatewayRepository) List(clusterID uint) ([]dal.KafkaMeshGatewayConfig, error) {
	var list []dal.KafkaMeshGatewayConfig
	query := GORMDB.Model(&dal.KafkaMeshGatewayConfig{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}
func (r *KafkaMeshGatewayRepository) GetByID(id uint) (*dal.KafkaMeshGatewayConfig, error) {
	var item dal.KafkaMeshGatewayConfig
	if err := GORMDB.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
func (r *KafkaMeshGatewayRepository) Create(item *dal.KafkaMeshGatewayConfig) error { return GORMDB.Create(item).Error }
func (r *KafkaMeshGatewayRepository) Update(item *dal.KafkaMeshGatewayConfig) error { return GORMDB.Save(item).Error }
func (r *KafkaMeshGatewayRepository) Delete(id uint) error                          { return GORMDB.Delete(&dal.KafkaMeshGatewayConfig{}, id).Error }

type KafkaCostRepository struct{}

func NewKafkaCostRepository() *KafkaCostRepository { return &KafkaCostRepository{} }
func (r *KafkaCostRepository) Create(item *dal.KafkaCostRecord) error { return GORMDB.Create(item).Error }
func (r *KafkaCostRepository) List(clusterID uint) ([]dal.KafkaCostRecord, error) {
	var list []dal.KafkaCostRecord
	query := GORMDB.Model(&dal.KafkaCostRecord{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("metric_date DESC").Find(&list).Error
	return list, err
}

type KafkaSensitiveRepository struct{}

func NewKafkaSensitiveRepository() *KafkaSensitiveRepository { return &KafkaSensitiveRepository{} }
func (r *KafkaSensitiveRepository) ListRules(clusterID uint) ([]dal.KafkaSensitiveScanRule, error) {
	var list []dal.KafkaSensitiveScanRule
	query := GORMDB.Model(&dal.KafkaSensitiveScanRule{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}
func (r *KafkaSensitiveRepository) GetRuleByID(id uint) (*dal.KafkaSensitiveScanRule, error) {
	var item dal.KafkaSensitiveScanRule
	if err := GORMDB.Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
func (r *KafkaSensitiveRepository) CreateRule(item *dal.KafkaSensitiveScanRule) error { return GORMDB.Create(item).Error }
func (r *KafkaSensitiveRepository) UpdateRule(item *dal.KafkaSensitiveScanRule) error { return GORMDB.Save(item).Error }
func (r *KafkaSensitiveRepository) DeleteRule(id uint) error                          { return GORMDB.Delete(&dal.KafkaSensitiveScanRule{}, id).Error }
func (r *KafkaSensitiveRepository) CreateResults(items []dal.KafkaSensitiveScanResult) error {
	if len(items) == 0 {
		return nil
	}
	return GORMDB.Create(&items).Error
}
func (r *KafkaSensitiveRepository) ListResults(clusterID uint) ([]dal.KafkaSensitiveScanResult, error) {
	var list []dal.KafkaSensitiveScanResult
	query := GORMDB.Model(&dal.KafkaSensitiveScanResult{})
	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	err := query.Order("id DESC").Find(&list).Error
	return list, err
}
