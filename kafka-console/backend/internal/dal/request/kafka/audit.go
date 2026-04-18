package kafka

type AuditLogListRequest struct {
	ClusterID uint   `form:"clusterId" json:"clusterId"`
	Action    string `form:"action" json:"action" binding:"omitempty,max=64"`
	Result    string `form:"result" json:"result" binding:"omitempty,oneof=success failed"`
	Page      int    `form:"page" json:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"pageSize" json:"pageSize" binding:"omitempty,min=1,max=100"`
}
