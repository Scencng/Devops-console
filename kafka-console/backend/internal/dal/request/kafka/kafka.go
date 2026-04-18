package kafka

type ClusterListRequest struct {
	Page     int    `form:"page" json:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" json:"pageSize" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" json:"keyword" binding:"omitempty,max=100"`
	Status   string `form:"status" json:"status" binding:"omitempty,oneof=active error unknown"`
	Environment string `form:"environment" json:"environment" binding:"omitempty,max=64"`
	Tenant      string `form:"tenant" json:"tenant" binding:"omitempty,max=64"`
}

type ClusterUpsertRequest struct {
	Name               string `json:"name" binding:"required,max=191"`
	BootstrapServers   string `json:"bootstrapServers" binding:"required,max=2000"`
	Version            string `json:"version" binding:"omitempty,max=50"`
	AuthType           string `json:"authType" binding:"omitempty,oneof=none plain scram_sha256 scram_sha512"`
	Username           string `json:"username" binding:"omitempty,max=255"`
	Password           string `json:"password" binding:"omitempty"`
	TLSEnabled         bool   `json:"tlsEnabled"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
	CACert             string `json:"caCert"`
	ClientCert         string `json:"clientCert"`
	ClientKey          string `json:"clientKey"`
	Description        string `json:"description"`
	Environment        string `json:"environment" binding:"omitempty,max=64"`
	Tenant             string `json:"tenant" binding:"omitempty,max=64"`
}

type ClusterQueryRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId" binding:"required"`
}

type TopicListRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId" binding:"required"`
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=200"`
}

type TopicActionRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId" binding:"required"`
}

type TopicConfigEntryRequest struct {
	Key       string  `json:"key" binding:"required,max=200"`
	Operation string  `json:"operation" binding:"required,oneof=set delete"`
	Value     *string `json:"value"`
}

type TopicConfigUpdateRequest struct {
	ClusterID uint                      `json:"clusterId" binding:"required"`
	Entries   []TopicConfigEntryRequest `json:"entries" binding:"required,min=1,dive"`
}

type TopicCreateConfigEntryRequest struct {
	Key   string `json:"key" binding:"required,max=200"`
	Value string `json:"value"`
}

type CreateTopicRequest struct {
	ClusterID         uint                           `json:"clusterId" binding:"required"`
	Name              string                         `json:"name" binding:"required,max=255"`
	NumPartitions     int32                          `json:"numPartitions" binding:"required,min=1"`
	ReplicationFactor int16                          `json:"replicationFactor" binding:"required,min=1"`
	ConfigEntries     []TopicCreateConfigEntryRequest `json:"configEntries" binding:"omitempty,dive"`
}

type TopicPartitionsRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId" binding:"required"`
}

type IncreaseTopicPartitionsRequest struct {
	ClusterID uint  `json:"clusterId" binding:"required"`
	Count     int32 `json:"count" binding:"required,min=1"`
}

type BrokerListRequest struct {
	ClusterID uint `form:"clusterId" json:"clusterId" binding:"required"`
}

type ConsumerGroupListRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId" binding:"required"`
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=200"`
}

type ConsumerGroupDetailRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId" binding:"required"`
	Topic     string `form:"topic" json:"topic" binding:"omitempty,max=255"`
}

type ResetConsumerGroupOffsetRequest struct {
	ClusterID     uint   `json:"clusterId" binding:"required"`
	Topic         string `json:"topic" binding:"required,max=255"`
	Partition     *int32 `json:"partition" binding:"omitempty,min=0"`
	AllPartitions bool   `json:"allPartitions"`
	ResetType     string `json:"resetType" binding:"required,oneof=earliest latest offset timestamp"`
	Offset        int64  `json:"offset"`
	TimestampMs   int64  `json:"timestampMs"`
}
