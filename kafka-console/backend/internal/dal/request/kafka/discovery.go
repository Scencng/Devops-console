package kafka

type DiscoveryAuthTemplateRequest struct {
	Version            string `json:"version" binding:"omitempty,max=50"`
	AuthType           string `json:"authType" binding:"omitempty,oneof=none plain scram_sha256 scram_sha512"`
	Username           string `json:"username" binding:"omitempty,max=255"`
	Password           string `json:"password"`
	TLSEnabled         bool   `json:"tlsEnabled"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
	CACert             string `json:"caCert"`
	ClientCert         string `json:"clientCert"`
	ClientKey          string `json:"clientKey"`
}

type DiscoveryScanRequest struct {
	CIDR        string                       `json:"cidr" binding:"required,max=64"`
	Ports       []int                        `json:"ports" binding:"required,min=1,dive,min=1,max=65535"`
	TimeoutMs   int                          `json:"timeoutMs" binding:"omitempty,min=200,max=30000"`
	Concurrency int                          `json:"concurrency" binding:"omitempty,min=1,max=1024"`
	Auth        DiscoveryAuthTemplateRequest `json:"auth"`
}

type DiscoveryImportRequest struct {
	Name        string                       `json:"name" binding:"required,max=191"`
	Address     string                       `json:"address" binding:"required,max=255"`
	Environment string                       `json:"environment" binding:"omitempty,max=64"`
	Tenant      string                       `json:"tenant" binding:"omitempty,max=64"`
	Description string                       `json:"description"`
	Auth        DiscoveryAuthTemplateRequest `json:"auth"`
}
