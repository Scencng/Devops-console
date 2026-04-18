package response

import "time"

type KafkaAuditLogVO struct {
	ID               uint      `json:"id"`
	ClusterID        uint      `json:"clusterId"`
	Action           string    `json:"action"`
	ResourceType     string    `json:"resourceType"`
	ResourceName     string    `json:"resourceName"`
	OperatorUserID   uint64    `json:"operatorUserId"`
	OperatorUsername string    `json:"operatorUsername"`
	RequestPayload   string    `json:"requestPayload"`
	Result           string    `json:"result"`
	ErrorMessage     string    `json:"errorMessage"`
	CreatedAt        time.Time `json:"createdAt"`
}

type KafkaAuditLogListVO struct {
	Total int64             `json:"total"`
	List  []KafkaAuditLogVO `json:"list"`
}

type KafkaPrometheusCardVO struct {
	Name  string  `json:"name"`
	Query string  `json:"query"`
	Value float64 `json:"value"`
}

type KafkaPrometheusPointVO struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type KafkaPrometheusSeriesVO struct {
	Metric string                    `json:"metric"`
	Points []KafkaPrometheusPointVO  `json:"points"`
}

type KafkaPrometheusPanelVO struct {
	Cards     []KafkaPrometheusCardVO  `json:"cards"`
	LagSeries []KafkaPrometheusSeriesVO `json:"lagSeries"`
}
