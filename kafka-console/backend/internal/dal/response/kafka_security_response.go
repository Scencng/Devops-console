package response

type KafkaACLVO struct {
	ResourceType   string `json:"resourceType"`
	ResourceName   string `json:"resourceName"`
	PatternType    string `json:"patternType"`
	Principal      string `json:"principal"`
	Host           string `json:"host"`
	Operation      string `json:"operation"`
	PermissionType string `json:"permissionType"`
}

type KafkaACLDeleteResultVO struct {
	DeletedCount int          `json:"deletedCount"`
	Entries      []KafkaACLVO `json:"entries"`
}

type KafkaScramCredentialVO struct {
	Mechanism  string `json:"mechanism"`
	Iterations int32  `json:"iterations"`
}

type KafkaScramUserVO struct {
	Username    string                   `json:"username"`
	Credentials []KafkaScramCredentialVO `json:"credentials"`
}
