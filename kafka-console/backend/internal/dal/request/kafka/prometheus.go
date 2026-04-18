package kafka

type PrometheusQueryRequest struct {
	Query string `form:"query" json:"query" binding:"required"`
	Time  string `form:"time" json:"time"`
}

type PrometheusQueryRangeRequest struct {
	Query string `form:"query" json:"query" binding:"required"`
	Start string `form:"start" json:"start" binding:"required"`
	End   string `form:"end" json:"end" binding:"required"`
	Step  string `form:"step" json:"step" binding:"required"`
}
