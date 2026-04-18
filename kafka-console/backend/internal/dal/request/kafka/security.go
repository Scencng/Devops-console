package kafka

type ACLListRequest struct {
	ClusterID      uint   `form:"clusterId" json:"clusterId" binding:"required"`
	ResourceType   string `form:"resourceType" json:"resourceType" binding:"omitempty,max=64"`
	ResourceName   string `form:"resourceName" json:"resourceName" binding:"omitempty,max=255"`
	PatternType    string `form:"patternType" json:"patternType" binding:"omitempty,max=64"`
	Principal      string `form:"principal" json:"principal" binding:"omitempty,max=255"`
	Host           string `form:"host" json:"host" binding:"omitempty,max=255"`
	Operation      string `form:"operation" json:"operation" binding:"omitempty,max=64"`
	PermissionType string `form:"permissionType" json:"permissionType" binding:"omitempty,max=64"`
}

type ACLUpsertRequest struct {
	ClusterID      uint   `json:"clusterId" binding:"required"`
	ResourceType   string `json:"resourceType" binding:"required,max=64"`
	ResourceName   string `json:"resourceName" binding:"required,max=255"`
	PatternType    string `json:"patternType" binding:"omitempty,max=64"`
	Principal      string `json:"principal" binding:"required,max=255"`
	Host           string `json:"host" binding:"omitempty,max=255"`
	Operation      string `json:"operation" binding:"required,max=64"`
	PermissionType string `json:"permissionType" binding:"required,max=64"`
}

type ACLDeleteRequest struct {
	ClusterID      uint   `json:"clusterId" binding:"required"`
	ResourceType   string `json:"resourceType" binding:"required,max=64"`
	ResourceName   string `json:"resourceName" binding:"required,max=255"`
	PatternType    string `json:"patternType" binding:"omitempty,max=64"`
	Principal      string `json:"principal" binding:"required,max=255"`
	Host           string `json:"host" binding:"omitempty,max=255"`
	Operation      string `json:"operation" binding:"required,max=64"`
	PermissionType string `json:"permissionType" binding:"required,max=64"`
}

type ScramUserListRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId" binding:"required"`
	Keyword   string `form:"keyword" json:"keyword" binding:"omitempty,max=255"`
}

type ScramUserUpsertRequest struct {
	ClusterID  uint   `json:"clusterId" binding:"required"`
	Username   string `json:"username" binding:"required,max=255"`
	Mechanism  string `json:"mechanism" binding:"required,oneof=sha256 sha512"`
	Password   string `json:"password" binding:"required"`
	Iterations int32  `json:"iterations" binding:"omitempty,min=4096"`
}

type ScramUserDeleteRequest struct {
	ClusterID uint   `json:"clusterId" binding:"required"`
	Username  string `json:"username" binding:"required,max=255"`
	Mechanism string `json:"mechanism" binding:"required,oneof=sha256 sha512"`
}
