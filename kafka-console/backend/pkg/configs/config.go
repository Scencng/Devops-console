package configs

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/pkg/utils/logs"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	Port   string
	Config *AppConfig
)

type MySQLConfig struct {
	Host         string `mapstructure:"host" yaml:"host"`
	Port         int    `mapstructure:"port" yaml:"port"`
	Username     string `mapstructure:"username" yaml:"username"`
	Password     string `mapstructure:"password" yaml:"password"`
	Database     string `mapstructure:"database" yaml:"database"`
	Charset      string `mapstructure:"charset" yaml:"charset"`
	ParseTime    bool   `mapstructure:"parse_time" yaml:"parse_time"`
	MaxOpenConns int    `mapstructure:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
}

type DatabaseConfig struct {
	Type        string      `mapstructure:"type" yaml:"type"`
	AutoMigrate bool        `mapstructure:"auto_migrate" yaml:"auto_migrate"`
	MySQL       MySQLConfig `mapstructure:"mysql" yaml:"mysql"`
}

type ServerConfig struct {
	Port     string `mapstructure:"port" yaml:"port"`
	LogLevel string `mapstructure:"log_level" yaml:"log_level"`
}

type LoggingConfig struct {
	Format       string `mapstructure:"format" yaml:"format"`
	TimeFormat   string `mapstructure:"time_format" yaml:"time_format"`
	ReportCaller bool   `mapstructure:"report_caller" yaml:"report_caller"`
}

type ElasticsearchConfig struct {
	Timeout             int `mapstructure:"timeout" yaml:"timeout"`
	Retry               int `mapstructure:"retry" yaml:"retry"`
	HealthCheckInterval int `mapstructure:"health_check_interval" yaml:"health_check_interval"`
}

type KubernetesConfig struct {
	ConfigPath string `mapstructure:"config_path" yaml:"config_path"`
	Timeout    int    `mapstructure:"timeout" yaml:"timeout"`
	Retry      int    `mapstructure:"retry" yaml:"retry"`
}

type PrometheusConfig struct {
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`
	Timeout int    `mapstructure:"timeout" yaml:"timeout"`
}

type SwaggerConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
	Host     string `mapstructure:"host" yaml:"host"`
	BasePath string `mapstructure:"base_path" yaml:"base_path"`
}

type HealthConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
	Endpoint string `mapstructure:"endpoint" yaml:"endpoint"`
	Interval int    `mapstructure:"interval" yaml:"interval"`
}

type AppConfig struct {
	Server        ServerConfig        `mapstructure:"server" yaml:"server"`
	Database      DatabaseConfig      `mapstructure:"database" yaml:"database"`
	Logging       LoggingConfig       `mapstructure:"logging" yaml:"logging"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch" yaml:"elasticsearch"`
	Kubernetes    KubernetesConfig    `mapstructure:"kubernetes" yaml:"kubernetes"`
	Prometheus    PrometheusConfig    `mapstructure:"prometheus" yaml:"prometheus"`
	Swagger       SwaggerConfig       `mapstructure:"swagger" yaml:"swagger"`
	Health        HealthConfig        `mapstructure:"health" yaml:"health"`
}

func initLogConfig(config *AppConfig) {
	switch strings.ToLower(config.Server.LogLevel) {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.SetReportCaller(config.Logging.ReportCaller)
	if config.Logging.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: config.Logging.TimeFormat, CallerPrettyfier: func(f *runtime.Frame) (string, string) { return f.Function, filepath.Base(f.File) }})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: config.Logging.TimeFormat})
	}
}

func GetConfig() *AppConfig { return Config }
func GetServerConfig() ServerConfig { return Config.Server }
func GetDatabaseConfig() DatabaseConfig { return Config.Database }
func GetElasticsearchConfig() ElasticsearchConfig { return Config.Elasticsearch }
func GetKubernetesConfig() KubernetesConfig { return Config.Kubernetes }
func GetPrometheusConfig() PrometheusConfig { return Config.Prometheus }
func GetSwaggerConfig() SwaggerConfig { return Config.Swagger }
func GetHealthConfig() HealthConfig { return Config.Health }
func IsDebugMode() bool { return strings.ToLower(Config.Server.LogLevel) == "debug" }
func IsProductionMode() bool { return strings.ToLower(Config.Server.LogLevel) == "error" || strings.ToLower(Config.Server.LogLevel) == "warn" }

func InitSwagger(r *gin.Engine) {
	if !Config.Swagger.Enabled {
		logs.Info(nil, "Swagger API文档已禁用")
		return
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logs.Info(map[string]interface{}{"host": Config.Swagger.Host, "base_path": Config.Swagger.BasePath}, "Swagger API文档初始化完成")
}

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.SetEnvPrefix("DEVOPS")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.BindEnv("database.mysql.host", "DB_HOST")
	viper.BindEnv("database.mysql.port", "DB_PORT")
	viper.BindEnv("database.mysql.username", "DB_USER")
	viper.BindEnv("database.mysql.password", "DB_PASSWORD")
	viper.BindEnv("database.mysql.database", "DB_NAME")
	viper.BindEnv("prometheus.base_url", "PROMETHEUS_BASE_URL")
	setDefaults()
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) { logs.Info(nil, "配置文件未找到，使用默认配置和环境变量") }
	}
	if err := common.LoadConfig(); err != nil { return err }
	Config = &AppConfig{}
	if err := viper.Unmarshal(Config); err != nil { return fmt.Errorf("解析配置文件失败: %v", err) }
	Port = Config.Server.Port
	initLogConfig(Config)
	logs.Info(map[string]interface{}{"config_file": viper.ConfigFileUsed(), "log_level": Config.Server.LogLevel}, "配置加载完成")
	return nil
}

func setDefaults() {
	viper.SetDefault("server.port", ":8081")
	viper.SetDefault("server.log_level", "info")
	viper.SetDefault("database.type", "mysql")
	viper.SetDefault("database.auto_migrate", true)
	viper.SetDefault("database.mysql.host", "mysql")
	viper.SetDefault("database.mysql.port", 3306)
	viper.SetDefault("database.mysql.username", "devops")
	viper.SetDefault("database.mysql.password", "devops123456")
	viper.SetDefault("database.mysql.database", "kafka_console")
	viper.SetDefault("database.mysql.charset", "utf8mb4")
	viper.SetDefault("database.mysql.parse_time", true)
	viper.SetDefault("database.mysql.max_open_conns", 10)
	viper.SetDefault("database.mysql.max_idle_conns", 5)
	viper.SetDefault("swagger.enabled", true)
	viper.SetDefault("swagger.host", "localhost:8081")
	viper.SetDefault("swagger.base_path", "/")
	viper.SetDefault("health.enabled", true)
	viper.SetDefault("health.endpoint", "/health")
	viper.SetDefault("health.interval", 30)
	viper.SetDefault("prometheus.base_url", "")
	viper.SetDefault("prometheus.timeout", 10)
}
