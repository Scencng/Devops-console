package kafka

import (
	"fmt"
	"time"

	"devops-console-backend/internal/dal"
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/configs"
)

func (s *Service) RunInspection(req reqKafka.InspectionRunRequest, triggeredBy string) (*response.KafkaInspectionReportVO, error) {
	cluster, err := s.repo.GetByID(req.ClusterID)
	if err != nil {
		return nil, err
	}
	brokers, err := s.ListBrokers(req.ClusterID)
	if err != nil {
		return nil, err
	}
	groups, err := s.ListConsumerGroups(req.ClusterID, "")
	if err != nil {
		return nil, err
	}
	underReplicatedValue := float64(0)
	if value, queryErr := s.queryPrometheusInstantValue("sum(kafka_topic_partition_under_replicated_partition)"); queryErr == nil {
		underReplicatedValue = value
	}
	items := make([]dal.KafkaInspectionItem, 0)
	issueCount := 0
	addItem := func(code, severity, status, title, detail string) {
		if status != "ok" {
			issueCount++
		}
		items = append(items, dal.KafkaInspectionItem{
			CheckCode: code,
			Severity:  severity,
			Status:    status,
			Title:     title,
			Detail:    detail,
			CreatedAt: time.Now(),
		})
	}

	downBrokers := 0
	for _, broker := range brokers {
		if !broker.Connected {
			downBrokers++
		}
	}
	if downBrokers > 0 {
		addItem("broker.connectivity", "critical", "fail", "存在异常 Broker", fmt.Sprintf("当前有 %d 个 Broker 连接异常", downBrokers))
	} else {
		addItem("broker.connectivity", "info", "ok", "Broker 连接正常", "所有 Broker 当前均可连接")
	}

	if underReplicatedValue > 0 {
		addItem("topic.replicas", "critical", "fail", "存在未同步副本分区", fmt.Sprintf("under replicated partitions=%.0f", underReplicatedValue))
	} else {
		addItem("topic.replicas", "info", "ok", "副本同步正常", "未发现 under replicated partitions")
	}

	maxLag := int64(0)
	maxLagGroup := ""
	for _, group := range groups {
		if group.CommittedLag > maxLag {
			maxLag = group.CommittedLag
			maxLagGroup = group.GroupID
		}
	}
	if maxLag > 0 {
		addItem("consumer.lag", "warning", "warning", "存在消费积压", fmt.Sprintf("最大积压消费组=%s，Lag=%d", maxLagGroup, maxLag))
	} else {
		addItem("consumer.lag", "info", "ok", "消费积压正常", "当前未发现消费积压")
	}

	alertRepo := configs.NewKafkaAlertEventRepository()
	openAlerts, _ := alertRepo.CountOpen(req.ClusterID)
	if openAlerts > 0 {
		addItem("alert.center", "warning", "warning", "存在未恢复告警", fmt.Sprintf("当前未恢复告警数=%d", openAlerts))
	} else {
		addItem("alert.center", "info", "ok", "告警中心无未恢复事件", "当前告警中心无 open/acked 状态事件")
	}

	status := "success"
	if issueCount > 0 {
		status = "warning"
	}
	report := &dal.KafkaInspectionReport{
		ClusterID:   req.ClusterID,
		Name:        req.Name,
		Status:      status,
		Summary:     fmt.Sprintf("巡检完成，共 %d 项检查，发现 %d 个问题", len(items), issueCount),
		IssueCount:  issueCount,
		Environment: cluster.Environment,
		Tenant:      cluster.Tenant,
		TriggeredBy: triggeredBy,
		ExecutedAt:  time.Now(),
		CreatedAt:   time.Now(),
	}
	if report.Name == "" {
		report.Name = fmt.Sprintf("%s 巡检 %s", cluster.Name, time.Now().Format("2006-01-02 15:04:05"))
	}
	repo := configs.NewKafkaInspectionRepository()
	if err = repo.CreateReportWithItems(report, items); err != nil {
		return nil, err
	}
	return s.GetInspectionReport(report.ID)
}

func (s *Service) ListInspectionReports(req reqKafka.InspectionListRequest) ([]response.KafkaInspectionReportVO, error) {
	list, err := configs.NewKafkaInspectionRepository().ListReports(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaInspectionReportVO, 0, len(list))
	for _, item := range list {
		result = append(result, toInspectionReportVO(item, nil))
	}
	return result, nil
}

func (s *Service) GetInspectionReport(id uint) (*response.KafkaInspectionReportVO, error) {
	repo := configs.NewKafkaInspectionRepository()
	report, err := repo.GetReportByID(id)
	if err != nil {
		return nil, err
	}
	items, err := repo.GetItemsByReportID(id)
	if err != nil {
		return nil, err
	}
	voItems := make([]response.KafkaInspectionItemVO, 0, len(items))
	for _, item := range items {
		voItems = append(voItems, response.KafkaInspectionItemVO{
			ID:        item.ID,
			CheckCode: item.CheckCode,
			Severity:  item.Severity,
			Status:    item.Status,
			Title:     item.Title,
			Detail:    item.Detail,
			CreatedAt: item.CreatedAt,
		})
	}
	vo := toInspectionReportVO(*report, voItems)
	return &vo, nil
}

func toInspectionReportVO(item dal.KafkaInspectionReport, items []response.KafkaInspectionItemVO) response.KafkaInspectionReportVO {
	return response.KafkaInspectionReportVO{
		ID:          item.ID,
		ClusterID:   item.ClusterID,
		Name:        item.Name,
		Status:      item.Status,
		Summary:     item.Summary,
		IssueCount:  item.IssueCount,
		Environment: item.Environment,
		Tenant:      item.Tenant,
		TriggeredBy: item.TriggeredBy,
		ExecutedAt:  item.ExecutedAt,
		CreatedAt:   item.CreatedAt,
		Items:       items,
	}
}
