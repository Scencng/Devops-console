package kafka

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/configs"
)

func (s *Service) GetHealthOverview(req reqKafka.HealthOverviewRequest) (*response.KafkaHealthOverviewVO, error) {
	brokers, err := s.ListBrokers(req.ClusterID)
	if err != nil {
		return nil, err
	}
	topics, err := s.ListTopics(req.ClusterID, "")
	if err != nil {
		return nil, err
	}
	groups, err := s.ListConsumerGroups(req.ClusterID, "")
	if err != nil {
		return nil, err
	}
	alertRepo := configs.NewKafkaAlertEventRepository()
	openAlerts, _ := alertRepo.CountOpen(req.ClusterID)

	healthyBrokers := 0
	unhealthyBrokers := 0
	for _, broker := range brokers {
		if broker.Connected {
			healthyBrokers++
		} else {
			unhealthyBrokers++
		}
	}
	underReplicated := int64(0)
	if value, queryErr := s.queryPrometheusInstantValue("sum(kafka_topic_partition_under_replicated_partition)"); queryErr == nil {
		underReplicated = int64(value)
	}
	totalLag := int64(0)
	for _, group := range groups {
		totalLag += group.CommittedLag
	}
	cards := []response.KafkaHealthCardVO{
		{Name: "Broker 总数", Value: len(brokers)},
		{Name: "健康 Broker", Value: healthyBrokers},
		{Name: "异常 Broker", Value: unhealthyBrokers},
		{Name: "Topic 总数", Value: len(topics)},
		{Name: "消费组总数", Value: len(groups)},
		{Name: "总 Lag", Value: totalLag},
		{Name: "未同步分区", Value: underReplicated},
		{Name: "未恢复告警", Value: openAlerts},
	}

	issues := make([]response.KafkaHealthIssueVO, 0)
	for _, broker := range brokers {
		if !broker.Connected {
			issues = append(issues, response.KafkaHealthIssueVO{
				Severity: "critical",
				Title:    fmt.Sprintf("Broker %d 连接异常", broker.ID),
				Detail:   fmt.Sprintf("Broker %s 当前处于断开状态", broker.Address),
			})
		}
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].CommittedLag > groups[j].CommittedLag })
	for _, group := range groups {
		if group.CommittedLag <= 0 {
			continue
		}
		issues = append(issues, response.KafkaHealthIssueVO{
			Severity: "warning",
			Title:    fmt.Sprintf("消费组 %s 存在积压", group.GroupID),
			Detail:   fmt.Sprintf("当前 Lag=%d，涉及 Topics：%v", group.CommittedLag, group.Topics),
		})
		if len(issues) >= 10 {
			break
		}
	}
	if underReplicated > 0 {
		issues = append([]response.KafkaHealthIssueVO{{
			Severity: "critical",
			Title:    "存在未同步副本分区",
			Detail:   fmt.Sprintf("当前 under replicated partitions=%d", underReplicated),
		}}, issues...)
	}
	return &response.KafkaHealthOverviewVO{Cards: cards, Issues: issues}, nil
}

func (s *Service) GetTrendSeries(req reqKafka.TrendRangeRequest) ([]response.KafkaTrendSeriesVO, error) {
	start, end, step := normalizeTrendRange(req.Start, req.End, req.Step)
	queries := map[string][]struct {
		Name  string
		Query string
	}{
		"default": {
			{Name: "总 Lag", Query: "sum(kafka_consumergroup_lag)"},
			{Name: "未同步分区", Query: "sum(kafka_topic_partition_under_replicated_partition)"},
			{Name: "Broker 数量", Query: "sum(kafka_brokers)"},
		},
		"throughput": {
			{Name: "In Messages/s", Query: "sum(rate(kafka_server_brokertopicmetrics_messagesin_total[5m]))"},
			{Name: "Bytes In/s", Query: "sum(rate(kafka_server_brokertopicmetrics_bytesin_total[5m]))"},
			{Name: "Bytes Out/s", Query: "sum(rate(kafka_server_brokertopicmetrics_bytesout_total[5m]))"},
		},
		"lag": {
			{Name: "总 Lag", Query: "sum(kafka_consumergroup_lag)"},
			{Name: "最大 Lag", Query: "max(kafka_consumergroup_lag)"},
		},
	}
	selected := queries[req.Preset]
	if len(selected) == 0 {
		selected = queries["default"]
	}
	result := make([]response.KafkaTrendSeriesVO, 0, len(selected))
	for _, item := range selected {
		series, err := s.queryPrometheusRangeSeries(item.Query, start, end, step)
		if err != nil {
			series = []response.KafkaPrometheusSeriesVO{}
		}
		result = append(result, response.KafkaTrendSeriesVO{Name: item.Name, Query: item.Query, Series: series})
	}
	return result, nil
}

func normalizeTrendRange(startRaw, endRaw, stepRaw string) (time.Time, time.Time, string) {
	end := time.Now()
	start := end.Add(-6 * time.Hour)
	if parsedEnd, ok := parseFlexibleTime(endRaw); ok {
		end = parsedEnd
	}
	if parsedStart, ok := parseFlexibleTime(startRaw); ok {
		start = parsedStart
	}
	if stepRaw == "" {
		stepRaw = "5m"
	}
	if start.After(end) {
		start = end.Add(-6 * time.Hour)
	}
	return start, end, stepRaw
}

func parseFlexibleTime(raw string) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}
	if seconds, err := strconv.ParseInt(raw, 10, 64); err == nil {
		if len(raw) >= 13 {
			return time.UnixMilli(seconds), true
		}
		return time.Unix(seconds, 0), true
	}
	if value, err := time.Parse(time.RFC3339, raw); err == nil {
		return value, true
	}
	return time.Time{}, false
}
