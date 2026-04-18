package main

import (
	_ "devops-console-backend/docs"
	"devops-console-backend/internal/common"
	monitorcontroller "devops-console-backend/internal/controllers/monitor"
	"devops-console-backend/internal/middlewares"
	routers "devops-console-backend/internal/routes"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/database"
	"devops-console-backend/pkg/utils/logs"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	r := gin.Default()
	err := configs.LoadConfig()
	if err != nil {
		logs.Error(nil, "加载配置文件失败")
		panic(err)
	}
	monitorcontroller.InitPrometheus()
	globalConfig := common.GetGlobalConfig()
	setMiddleware(r, globalConfig)
	database.InitRedis()
	defer database.CloseRedis()
	configs.NewDB()
	defer configs.CloseDB()
	if err = configs.AutoMigrateCustomTables(); err != nil {
		logs.Error(map[string]interface{}{"error": err.Error()}, "自定义表自动迁移失败")
		panic(err)
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With", "Accept", "X-HTTP-Method-Override"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	logs.Info(nil, "Kafka Console 启动成功")
	configs.InitSwagger(r)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now().Unix()}) })
	routers.RegisterRouters(r, configs.GORMDB)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(":9090", nil)
	}()
	_ = r.Run(configs.Port)
}

func setMiddleware(router *gin.Engine, globalConfig *common.GlobalConfig) {
	router.Use(middlewares.Authenticate(globalConfig.Jwt.ExcludePaths...))
	router.Use(middlewares.Metrics())
	router.Use(middlewares.IPRateLimit())
}
