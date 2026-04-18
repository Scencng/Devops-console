package kafka

import (
	"errors"
	"fmt"
	"time"

	"devops-console-backend/internal/dal"
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/configs"

	"gorm.io/gorm"
)

func (s *Service) ListAlertRules(req reqKafka.AlertRuleListRequest) ([]response.KafkaAlertRuleVO, error) {
	repo := configs.NewKafkaAlertRuleRepository()
	list, err := repo.List(req.ClusterID, req.Environment, req.Tenant)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaAlertRuleVO, 0, len(list))
	for _, item := range list {
		result = append(result, toAlertRuleVO(item))
	}
	return result, nil
}

func (s *Service) CreateAlertRule(req reqKafka.AlertRuleUpsertRequest) (*response.KafkaAlertRuleVO, error) {
	item := &dal.KafkaAlertRule{
		ClusterID:   req.ClusterID,
		Name:        req.Name,
		MetricType:  req.MetricType,
		Severity:    req.Severity,
		Threshold:   req.Threshold,
		Operator:    normalizeAlertOperator(req.Operator),
		Enabled:     req.Enabled,
		Runbook:     req.Runbook,
		Environment: req.Environment,
		Tenant:      req.Tenant,
	}
	repo := configs.NewKafkaAlertRuleRepository()
	if err := repo.Create(item); err != nil {
		return nil, err
	}
	vo := toAlertRuleVO(*item)
	return &vo, nil
}

func (s *Service) UpdateAlertRule(id uint, req reqKafka.AlertRuleUpsertRequest) (*response.KafkaAlertRuleVO, error) {
	repo := configs.NewKafkaAlertRuleRepository()
	item, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	item.ClusterID = req.ClusterID
	item.Name = req.Name
	item.MetricType = req.MetricType
	item.Severity = req.Severity
	item.Threshold = req.Threshold
	item.Operator = normalizeAlertOperator(req.Operator)
	item.Enabled = req.Enabled
	item.Runbook = req.Runbook
	item.Environment = req.Environment
	item.Tenant = req.Tenant
	if err = repo.Update(item); err != nil {
		return nil, err
	}
	vo := toAlertRuleVO(*item)
	return &vo, nil
}

func (s *Service) DeleteAlertRule(id uint) error {
	return configs.NewKafkaAlertRuleRepository().Delete(id)
}

func (s *Service) EvaluateAlertRules(clusterID uint) ([]response.KafkaAlertEventVO, error) {
	ruleRepo := configs.NewKafkaAlertRuleRepository()
	eventRepo := configs.NewKafkaAlertEventRepository()
	silenceRepo := configs.NewKafkaAlertSilenceRepository()
	rules, err := ruleRepo.ListEnabled(clusterID)
	if err != nil {
		return nil, err
	}
	silences, _ := silenceRepo.List(clusterID)
	result := make([]response.KafkaAlertEventVO, 0)
	for _, rule := range rules {
		value, message, evalErr := s.evaluateAlertMetric(clusterID, rule.MetricType)
		if evalErr != nil {
			message = evalErr.Error()
			value = 0
		}
		breached := compareAlertValue(value, rule.Operator, rule.Threshold)
		existing, existingErr := eventRepo.GetLatestOpenByRule(rule.ID)
		if breached {
			if alertSilenced(rule, silences, time.Now()) {
				continue
			}
			if existingErr == nil && existing != nil {
				existing.MetricValue = value
				existing.LastTriggeredAt = time.Now()
				existing.Message = message
				existing.Runbook = rule.Runbook
				_ = eventRepo.Update(existing)
				result = append(result, toAlertEventVO(*existing))
				continue
			}
			if existingErr != nil && !errors.Is(existingErr, gorm.ErrRecordNotFound) {
				return nil, existingErr
			}
			now := time.Now()
			item := &dal.KafkaAlertEvent{
				ClusterID:        clusterID,
				RuleID:           &rule.ID,
				Title:            fmt.Sprintf("%s 触发阈值", rule.Name),
				Severity:         rule.Severity,
				Status:           "open",
				MetricType:       rule.MetricType,
				MetricValue:      value,
				Threshold:        rule.Threshold,
				Message:          message,
				Runbook:          rule.Runbook,
				Environment:      rule.Environment,
				Tenant:           rule.Tenant,
				FirstTriggeredAt: now,
				LastTriggeredAt:  now,
			}
			if err = eventRepo.Create(item); err != nil {
				return nil, err
			}
			result = append(result, toAlertEventVO(*item))
			continue
		}
		if existingErr == nil && existing != nil {
			now := time.Now()
			existing.Status = "resolved"
			existing.ResolvedAt = &now
			existing.ResolvedBy = "system"
			existing.Message = message
			_ = eventRepo.Update(existing)
			result = append(result, toAlertEventVO(*existing))
		}
	}
	return result, nil
}

func alertSilenced(rule dal.KafkaAlertRule, silences []dal.KafkaAlertSilence, now time.Time) bool {
	for _, silence := range silences {
		if !silence.Enabled {
			continue
		}
		if now.Before(silence.StartsAt) || now.After(silence.EndsAt) {
			continue
		}
		if silence.MetricType != "" && silence.MetricType != rule.MetricType {
			continue
		}
		if silence.Severity != "" && silence.Severity != rule.Severity {
			continue
		}
		return true
	}
	return false
}

func (s *Service) ListAlertEvents(req reqKafka.AlertEventListRequest) ([]response.KafkaAlertEventVO, error) {
	list, err := configs.NewKafkaAlertEventRepository().List(req.ClusterID, req.Status, req.Severity)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaAlertEventVO, 0, len(list))
	for _, item := range list {
		result = append(result, toAlertEventVO(item))
	}
	return result, nil
}

func (s *Service) UpdateAlertEventStatus(id uint, status string, operator string) (*response.KafkaAlertEventVO, error) {
	repo := configs.NewKafkaAlertEventRepository()
	item, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	switch status {
	case "acked":
		item.Status = "acked"
		item.AckedBy = operator
		item.AckedAt = &now
	case "resolved":
		item.Status = "resolved"
		item.ResolvedBy = operator
		item.ResolvedAt = &now
	default:
		return nil, errors.New("不支持的告警状态")
	}
	if err = repo.Update(item); err != nil {
		return nil, err
	}
	vo := toAlertEventVO(*item)
	return &vo, nil
}

func (s *Service) evaluateAlertMetric(clusterID uint, metricType string) (float64, string, error) {
	switch metricType {
	case "consumer_lag":
		dashboard, err := s.GetDashboard(clusterID)
		if err != nil {
			return 0, "", err
		}
		return float64(dashboard.TotalLag), fmt.Sprintf("当前总 Lag=%.0f", float64(dashboard.TotalLag)), nil
	case "under_replicated_partitions":
		value, err := s.queryPrometheusInstantValue("sum(kafka_topic_partition_under_replicated_partition)")
		return value, fmt.Sprintf("当前未同步分区=%.0f", value), err
	case "broker_down", "unhealthy_brokers":
		brokers, err := s.ListBrokers(clusterID)
		if err != nil {
			return 0, "", err
		}
		count := 0.0
		for _, broker := range brokers {
			if !broker.Connected {
				count++
			}
		}
		return count, fmt.Sprintf("当前异常 Broker=%.0f", count), nil
	default:
		return 0, "", fmt.Errorf("不支持的告警指标: %s", metricType)
	}
}

func normalizeAlertOperator(operator string) string {
	switch operator {
	case ">=", ">", "<", "<=", "==":
		return operator
	default:
		return ">"
	}
}

func compareAlertValue(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	default:
		return value > threshold
	}
}

func toAlertRuleVO(item dal.KafkaAlertRule) response.KafkaAlertRuleVO {
	return response.KafkaAlertRuleVO{
		ID:          item.ID,
		ClusterID:   item.ClusterID,
		Name:        item.Name,
		MetricType:  item.MetricType,
		Severity:    item.Severity,
		Threshold:   item.Threshold,
		Operator:    item.Operator,
		Enabled:     item.Enabled,
		Runbook:     item.Runbook,
		Environment: item.Environment,
		Tenant:      item.Tenant,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func toAlertEventVO(item dal.KafkaAlertEvent) response.KafkaAlertEventVO {
	return response.KafkaAlertEventVO{
		ID:               item.ID,
		ClusterID:        item.ClusterID,
		RuleID:           item.RuleID,
		Title:            item.Title,
		Severity:         item.Severity,
		Status:           item.Status,
		MetricType:       item.MetricType,
		MetricValue:      item.MetricValue,
		Threshold:        item.Threshold,
		Message:          item.Message,
		Runbook:          item.Runbook,
		Environment:      item.Environment,
		Tenant:           item.Tenant,
		AckedBy:          item.AckedBy,
		ResolvedBy:       item.ResolvedBy,
		FirstTriggeredAt: item.FirstTriggeredAt,
		LastTriggeredAt:  item.LastTriggeredAt,
		AckedAt:          item.AckedAt,
		ResolvedAt:       item.ResolvedAt,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}
