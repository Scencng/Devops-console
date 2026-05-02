package dal

import "time"

const (
	KafkaClusterStatusUnknown = "unknown"
	KafkaClusterStatusActive  = "active"
	KafkaClusterStatusError   = "error"

	KafkaAuthTypeNone        = "none"
	KafkaAuthTypePlain       = "plain"
	KafkaAuthTypeSCRAMSHA256 = "scram_sha256"
	KafkaAuthTypeSCRAMSHA512 = "scram_sha512"
	KafkaResourceTypeCluster = "kafka_cluster"
)

// KafkaCluster stores the connection profile for a managed Kafka cluster.
type KafkaCluster struct {
	ID                  uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name                string     `gorm:"uniqueIndex;not null;column:name;size:191" json:"name"`
	BootstrapServers    string     `gorm:"not null;column:bootstrap_servers;size:2000" json:"bootstrap_servers"`
	Version             string     `gorm:"not null;default:'3.6.0';column:version;size:50" json:"version"`
	AuthType            string     `gorm:"not null;default:'none';column:auth_type;size:50" json:"auth_type"`
	Username            string     `gorm:"column:username;size:255" json:"username"`
	PasswordCiphertext  string     `gorm:"column:password_ciphertext;type:text" json:"-"`
	TLSEnabled          bool       `gorm:"not null;default:false;column:tls_enabled" json:"tls_enabled"`
	InsecureSkipVerify  bool       `gorm:"not null;default:false;column:insecure_skip_verify" json:"insecure_skip_verify"`
	CACert              string     `gorm:"column:ca_cert;type:longtext" json:"ca_cert"`
	ClientCert          string     `gorm:"column:client_cert;type:longtext" json:"client_cert"`
	ClientKeyCiphertext string     `gorm:"column:client_key_ciphertext;type:longtext" json:"-"`
	Description         string     `gorm:"column:description;type:text" json:"description"`
	Environment         string     `gorm:"column:environment;size:64" json:"environment"`
	Tenant              string     `gorm:"column:tenant;size:64" json:"tenant"`
	Status              string     `gorm:"not null;default:'unknown';index;column:status;size:50" json:"status"`
	LastErrorMessage    string     `gorm:"column:last_error_message;type:text" json:"last_error_message"`
	LastTestedAt        *time.Time `gorm:"column:last_tested_at" json:"last_tested_at"`
	CreatedAt           time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (KafkaCluster) TableName() string {
	return "kafka_clusters"
}
