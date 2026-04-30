package response

import "time"

type KafkaClusterVO struct {
	ID                 uint       `json:"id"`
	Name               string     `json:"name"`
	BootstrapServers   string     `json:"bootstrapServers"`
	Version            string     `json:"version"`
	AuthType           string     `json:"authType"`
	Username           string     `json:"username"`
	TLSEnabled         bool       `json:"tlsEnabled"`
	InsecureSkipVerify bool       `json:"insecureSkipVerify"`
	HasCACert          bool       `json:"hasCACert"`
	HasClientCert      bool       `json:"hasClientCert"`
	HasClientKey       bool       `json:"hasClientKey"`
	Description        string     `json:"description"`
	Environment        string     `json:"environment"`
	Tenant             string     `json:"tenant"`
	Status             string     `json:"status"`
	LastErrorMessage   string     `json:"lastErrorMessage"`
	LastTestedAt       *time.Time `json:"lastTestedAt"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type KafkaClusterDetailVO struct {
	KafkaClusterVO
	CACert     string `json:"caCert"`
	ClientCert string `json:"clientCert"`
}

type KafkaClusterListVO struct {
	Total int64            `json:"total"`
	List  []KafkaClusterVO `json:"list"`
}

type KafkaClusterOptionVO struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type KafkaClusterTestVO struct {
	ClusterID    uint      `json:"clusterId"`
	ClusterName  string    `json:"clusterName"`
	BrokerCount  int       `json:"brokerCount"`
	ControllerID int32     `json:"controllerId"`
	ResponseTime int64     `json:"responseTime"`
	TestedAt     time.Time `json:"testedAt"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"errorMessage"`
}

type KafkaBrokerVO struct {
	ID                    int32    `json:"id"`
	Address               string   `json:"address"`
	IsController          bool     `json:"isController"`
	Connected             bool     `json:"connected"`
	LeaderPartitionCount  int      `json:"leaderPartitionCount"`
	ReplicaPartitionCount int      `json:"replicaPartitionCount"`
	Topics                []string `json:"topics"`
}

type KafkaTopicVO struct {
	Name              string            `json:"name"`
	Partitions        int32             `json:"partitions"`
	ReplicationFactor int16             `json:"replicationFactor"`
	Internal          bool              `json:"internal"`
	CleanupPolicy     string            `json:"cleanupPolicy"`
	RetentionMs       string            `json:"retentionMs"`
	MinInSyncReplicas string            `json:"minInSyncReplicas"`
	ConfigEntries     map[string]string `json:"configEntries"`
}

type KafkaTopicCreateVO struct {
	Name              string `json:"name"`
	Partitions        int32  `json:"partitions"`
	ReplicationFactor int16  `json:"replicationFactor"`
}

type KafkaTopicPartitionsUpdateVO struct {
	Name               string `json:"name"`
	PreviousPartitions int32  `json:"previousPartitions"`
	CurrentPartitions  int32  `json:"currentPartitions"`
}

type KafkaTopicPartitionVO struct {
	Partition            int32   `json:"partition"`
	Leader               int32   `json:"leader"`
	Replicas             []int32 `json:"replicas"`
	ISR                  []int32 `json:"isr"`
	OfflineReplicas      []int32 `json:"offlineReplicas"`
	OutOfSyncReplicas    []int32 `json:"outOfSyncReplicas"`
	UnderReplicated      bool    `json:"underReplicated"`
	OldestOffset         int64   `json:"oldestOffset"`
	LatestOffset         int64   `json:"latestOffset"`
	MessageCountEstimate int64   `json:"messageCountEstimate"`
}

type KafkaTopicPartitionDetailVO struct {
	Topic                string                  `json:"topic"`
	PartitionCount       int                     `json:"partitionCount"`
	UnderReplicatedCount int                     `json:"underReplicatedCount"`
	Partitions           []KafkaTopicPartitionVO `json:"partitions"`
}

type KafkaConsumerGroupVO struct {
	GroupID           string   `json:"groupId"`
	ProtocolType      string   `json:"protocolType"`
	State             string   `json:"state"`
	MemberCount       int      `json:"memberCount"`
	Topics            []string `json:"topics"`
	PartitionCount    int      `json:"partitionCount"`
	CommittedLag      int64    `json:"committedLag"`
	LagAvailable      bool     `json:"lagAvailable"`
	LagPartial        bool     `json:"lagPartial"`
	LagWarningMessage string   `json:"lagWarningMessage"`
}

type KafkaConsumerGroupMemberVO struct {
	MemberID   string   `json:"memberId"`
	ClientID   string   `json:"clientId"`
	ClientHost string   `json:"clientHost"`
	Topics     []string `json:"topics"`
}

type KafkaConsumerGroupPartitionLagVO struct {
	Topic           string `json:"topic"`
	Partition       int32  `json:"partition"`
	CommittedOffset int64  `json:"committedOffset"`
	LatestOffset    int64  `json:"latestOffset"`
	OldestOffset    int64  `json:"oldestOffset"`
	Lag             int64  `json:"lag"`
	MemberID        string `json:"memberId"`
	ClientID        string `json:"clientId"`
	ClientHost      string `json:"clientHost"`
}

type KafkaConsumerGroupDetailVO struct {
	GroupID        string                             `json:"groupId"`
	ProtocolType   string                             `json:"protocolType"`
	State          string                             `json:"state"`
	MemberCount    int                                `json:"memberCount"`
	PartitionCount int                                `json:"partitionCount"`
	TotalLag       int64                              `json:"totalLag"`
	Topics         []string                           `json:"topics"`
	Members        []KafkaConsumerGroupMemberVO       `json:"members"`
	Partitions     []KafkaConsumerGroupPartitionLagVO `json:"partitions"`
}

type KafkaConsumerGroupOffsetResetPartitionVO struct {
	Partition int32 `json:"partition"`
	Offset    int64 `json:"offset"`
}

type KafkaConsumerGroupOffsetResetVO struct {
	GroupID       string                                     `json:"groupId"`
	Topic         string                                     `json:"topic"`
	AllPartitions bool                                       `json:"allPartitions"`
	ResetType     string                                     `json:"resetType"`
	Partitions    []KafkaConsumerGroupOffsetResetPartitionVO `json:"partitions"`
}

type KafkaDashboardVO struct {
	BrokerCount        int                    `json:"brokerCount"`
	TopicCount         int                    `json:"topicCount"`
	ConsumerGroupCount int                    `json:"consumerGroupCount"`
	TotalPartitions    int                    `json:"totalPartitions"`
	TotalLag           int64                  `json:"totalLag"`
	TopLagGroups       []KafkaConsumerGroupVO `json:"topLagGroups"`
}
