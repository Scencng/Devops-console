package response

import "time"

type KafkaMessageHeaderVO struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KafkaMessageVO struct {
	Offset       int64                  `json:"offset"`
	Partition    int32                  `json:"partition"`
	Timestamp    time.Time              `json:"timestamp"`
	KeyPreview   string                 `json:"keyPreview"`
	ValuePreview string                 `json:"valuePreview"`
	KeyBase64    string                 `json:"keyBase64"`
	ValueBase64  string                 `json:"valueBase64"`
	Headers      []KafkaMessageHeaderVO `json:"headers"`
}

type KafkaMessageBrowseVO struct {
	Topic          string           `json:"topic"`
	Partition      int32            `json:"partition"`
	StartOffset    int64            `json:"startOffset"`
	Count          int              `json:"count"`
	Messages       []KafkaMessageVO `json:"messages"`
	Partial        bool             `json:"partial"`
	TimedOut       bool             `json:"timedOut"`
	WarningMessage string           `json:"warningMessage"`
}

type KafkaMessageProduceVO struct {
	Topic     string    `json:"topic"`
	Partition int32     `json:"partition"`
	Offset    int64     `json:"offset"`
	Timestamp time.Time `json:"timestamp"`
}
