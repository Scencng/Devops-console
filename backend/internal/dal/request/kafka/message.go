package kafka

type MessageBrowseRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId" binding:"required"`
	Topic     string `form:"topic" json:"topic" binding:"required,max=255"`
	Partition *int32 `form:"partition" json:"partition" binding:"omitempty,min=0"`
	Mode      string `form:"mode" json:"mode" binding:"omitempty,oneof=earliest latest offset"`
	Offset    int64  `form:"offset" json:"offset"`
	Limit     int    `form:"limit" json:"limit" binding:"omitempty,min=1,max=500"`
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=500"`
}

type MessageHeaderRequest struct {
	Key           string `json:"key" binding:"required,max=255"`
	Value         string `json:"value"`
	ValueEncoding string `json:"valueEncoding" binding:"omitempty,oneof=plain base64"`
}

type MessageProduceRequest struct {
	ClusterID     uint                  `json:"clusterId" binding:"required"`
	Topic         string                `json:"topic" binding:"required,max=255"`
	Partition     *int32                `json:"partition" binding:"omitempty,min=0"`
	Key           string                `json:"key"`
	KeyEncoding   string                `json:"keyEncoding" binding:"omitempty,oneof=plain base64"`
	Value         string                `json:"value" binding:"required"`
	ValueEncoding string                `json:"valueEncoding" binding:"omitempty,oneof=plain base64"`
	Headers       []MessageHeaderRequest `json:"headers" binding:"omitempty,dive"`
}
