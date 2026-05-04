package model

type OpenConnectionRequest struct {
	Host     string            `json:"host" binding:"required"`
	Port     int               `json:"port" binding:"required,min=1,max=65535"`
	Username string            `json:"username" binding:"required"`
	Password string            `json:"password"`
	Database string            `json:"database"`
	Params   map[string]string `json:"params"`
}

type CloseConnectionRequest struct {
	ConnectionToken string `json:"connectionToken" binding:"required,uuid"`
}
