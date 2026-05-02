package response

type KafkaDiscoveryResultVO struct {
	IP                 string   `json:"ip"`
	Port               int      `json:"port"`
	Address            string   `json:"address"`
	LooksLikeKafka     bool     `json:"looksLikeKafka"`
	AdvertisedBroker   bool     `json:"advertisedBroker"`
	KafkaVersion       string   `json:"kafkaVersion"`
	BrokerID           int32    `json:"brokerId"`
	ClusterID          string   `json:"clusterId"`
	ControllerID       int32    `json:"controllerId"`
	Listeners          []string `json:"listeners"`
	VersionDetectError string   `json:"versionDetectError"`
	ErrorMessage       string   `json:"errorMessage"`
}
