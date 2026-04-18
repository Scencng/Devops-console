package kafka

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/configs"
)

type prometheusAPIResponse struct {
	Status    string `json:"status"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
	Data      struct {
		ResultType string                   `json:"resultType"`
		Result     []map[string]interface{} `json:"result"`
	} `json:"data"`
}

func (s *Service) GetPrometheusPanels() (*response.KafkaPrometheusPanelVO, error) {
	cards := []response.KafkaPrometheusCardVO{}
	var firstErr error
	successCount := 0
	queries := []struct {
		Name  string
		Query string
	}{
		{Name: "Broker", Query: "sum(kafka_brokers)"},
		{Name: "Topic 分区", Query: "sum(kafka_topic_partitions)"},
		{Name: "消费组 Lag", Query: "sum(kafka_consumergroup_lag)"},
		{Name: "未同步副本", Query: "sum(kafka_topic_partition_under_replicated_partition)"},
	}
	for _, item := range queries {
		value, err := s.queryPrometheusInstantValue(item.Query)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			value = 0
		} else {
			successCount++
		}
		cards = append(cards, response.KafkaPrometheusCardVO{Name: item.Name, Query: item.Query, Value: value})
	}
	series, err := s.queryPrometheusRangeSeries("sum(kafka_consumergroup_lag)", time.Now().Add(-time.Hour), time.Now(), "1m")
	if err != nil {
		if firstErr == nil {
			firstErr = err
		}
	} else {
		successCount++
	}
	if successCount == 0 && firstErr != nil {
		return nil, firstErr
	}
	return &response.KafkaPrometheusPanelVO{Cards: cards, LagSeries: series}, nil
}

func (s *Service) QueryPrometheus(query string, queryTime string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("query", query)
	if queryTime != "" {
		params.Set("time", queryTime)
	}
	return s.prometheusGet("/api/v1/query", params)
}

func (s *Service) QueryPrometheusRange(query, start, end, step string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", start)
	params.Set("end", end)
	params.Set("step", step)
	return s.prometheusGet("/api/v1/query_range", params)
}

func (s *Service) queryPrometheusInstantValue(query string) (float64, error) {
	data, err := s.QueryPrometheus(query, "")
	if err != nil {
		return 0, err
	}
	result, _ := data["result"].([]interface{})
	if len(result) == 0 {
		return 0, nil
	}
	item, _ := result[0].(map[string]interface{})
	value, _ := item["value"].([]interface{})
	if len(value) < 2 {
		return 0, nil
	}
	valueStr, _ := value[1].(string)
	return strconv.ParseFloat(valueStr, 64)
}

func (s *Service) queryPrometheusRangeSeries(query string, start, end time.Time, step string) ([]response.KafkaPrometheusSeriesVO, error) {
	data, err := s.QueryPrometheusRange(query, fmt.Sprintf("%d", start.Unix()), fmt.Sprintf("%d", end.Unix()), step)
	if err != nil {
		return nil, err
	}
	result, _ := data["result"].([]interface{})
	series := make([]response.KafkaPrometheusSeriesVO, 0, len(result))
	for _, raw := range result {
		item, _ := raw.(map[string]interface{})
		metricMap, _ := item["metric"].(map[string]interface{})
		metric := "kafka_consumergroup_lag"
		if name, ok := metricMap["__name__"].(string); ok && name != "" {
			metric = name
		}
		values, _ := item["values"].([]interface{})
		points := make([]response.KafkaPrometheusPointVO, 0, len(values))
		for _, rawPoint := range values {
			pair, _ := rawPoint.([]interface{})
			if len(pair) < 2 {
				continue
			}
			timestampFloat, _ := pair[0].(float64)
			valueStr, _ := pair[1].(string)
			value, _ := strconv.ParseFloat(valueStr, 64)
			points = append(points, response.KafkaPrometheusPointVO{Timestamp: int64(timestampFloat), Value: value})
		}
		series = append(series, response.KafkaPrometheusSeriesVO{Metric: metric, Points: points})
	}
	return series, nil
}

func (s *Service) prometheusGet(path string, params url.Values) (map[string]interface{}, error) {
	cfg := configs.GetConfig()
	if cfg == nil || cfg.Prometheus.BaseURL == "" {
		return nil, fmt.Errorf("Prometheus 未配置，请先设置 prometheus.base_url")
	}
	baseURL, err := url.Parse(strings.TrimRight(cfg.Prometheus.BaseURL, "/"))
	if err != nil {
		return nil, err
	}
	endpointPath := strings.TrimLeft(path, "/")
	basePath := strings.TrimRight(baseURL.Path, "/")
	if basePath == "" {
		baseURL.Path = "/" + endpointPath
	} else {
		baseURL.Path = basePath + "/" + endpointPath
	}
	baseURL.RawQuery = params.Encode()
	client := &http.Client{Timeout: time.Duration(cfg.Prometheus.Timeout) * time.Second}
	resp, err := client.Get(baseURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Prometheus 请求失败: HTTP %d - %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload prometheusAPIResponse
	if err = json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.Status != "success" {
		if payload.Error != "" {
			return nil, fmt.Errorf("Prometheus 查询失败: %s", payload.Error)
		}
		if payload.ErrorType != "" {
			return nil, fmt.Errorf("Prometheus 查询失败: %s", payload.ErrorType)
		}
		return nil, fmt.Errorf("Prometheus 查询失败: status=%s", payload.Status)
	}
	return map[string]interface{}{"resultType": payload.Data.ResultType, "result": payload.Data.Result}, nil
}
