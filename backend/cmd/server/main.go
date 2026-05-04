// 椤圭洰鐨勬€诲叆鍙?
// @title DevOps Console API
// @version 1.0
// @description DevOps Console鍚庣API鏂囨。
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	_ "devops-console-backend/docs" // swagger docs
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/controllers/monitor"
	"devops-console-backend/internal/dal/model"
	mysqlbootstrap "devops-console-backend/internal/mysql/bootstrap"
	"devops-console-backend/internal/middlewares"
	"devops-console-backend/internal/routes"
	"devops-console-backend/internal/services/probe"
	"devops-console-backend/internal/services/scheduler"
	"devops-console-backend/internal/services/task_scheduler/executor"
	"devops-console-backend/internal/websocket"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/database"
	"devops-console-backend/pkg/utils/logs"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

func main() {
	// 1. 鍔犺浇绋嬪簭鐨勯厤缃?
	// 2. 閰嶇疆gin
	r := gin.Default()
	err := configs.LoadConfig()
	if err != nil {
		logs.Error(nil, "鍔犺浇閰嶇疆鏂囦欢澶辫触")
		panic(err)
	}
	globalConfig := common.GetGlobalConfig()
	setMiddleware(r, globalConfig)
	// 鍒濆鍖栨暟鎹簱
	database.InitRedis()
	defer database.CloseRedis()
	db := configs.NewDB()
	if db == nil {
		logs.Error(nil, "database initialization failed, server exiting")
		return
	}
	defer configs.CloseDB()
	if err := mysqlbootstrap.EnsureMySQLMenus(configs.GORMDB); err != nil {
		logs.Error(nil, fmt.Sprintf("mysql menu bootstrap failed: %v", err))
		return
	}
	// 璺ㄥ煙閰嶇疆 todo 寰呰縼绉?
	r.Use(cors.New(cors.Config{
		//AllowOrigins:     []string{"http://127.0.0.1:5174", "http://localhost:5174"}, // 鍓嶇鍦板潃
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-ES-Host", "X-ES-Username", "X-ES-Password", "X-Connection-Token", "X-Requested-With", "Accept", "X-HTTP-Method-Override"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// 鍒濆鍖?prometheus monitor
	monitor.InitPrometheus()
	configs.InitConfig()
	probe.StartInstanceStatusProbe()
	// 3. 鏃ュ織閰嶇疆
	logs.Info(nil, "绋嬪簭鍚姩鎴愬姛")

	// Swagger API鏂囨。 - 鍒濆鍖栧凡绉昏嚦 config 鍖?
	configs.InitSwagger(r)

	// 娣诲姞鍋ュ悍妫€鏌ョ鐐?
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	executor.InitExecutors()

	// 娉ㄥ唽璺敱
	routers.RegisterRouters(r, configs.GORMDB)
	// 娉ㄥ唽WebSocket璺敱
	websocket.RegisterWebSocketRoutes(r)

	go func() {
		if err := loadCronSchedules(configs.GORMDB); err != nil {
			logs.Error(nil, fmt.Sprintf("鍔犺浇瀹氭椂璋冨害澶辫触: %v", err))
		}
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(":9090", nil)
	}()
	err = r.Run(configs.Port)
	if err != nil {
		return
	}
}

// 璁剧疆涓棿浠?
func setMiddleware(router *gin.Engine, globalConfig *common.GlobalConfig) {
	excludePaths := append([]string{}, globalConfig.Jwt.ExcludePaths...)
	// 璁よ瘉
	router.Use(middlewares.Authenticate(excludePaths...))
	router.Use(middlewares.Metrics())
	router.Use(middlewares.IPRateLimit())
}

func loadCronSchedules(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.TaskWorkflow{}) {
		logs.Info(nil, "task_workflows 琛ㄤ笉瀛樺湪锛岃烦杩囧畾鏃跺伐浣滄祦鍔犺浇")
		return nil
	}

	var workflows []*model.TaskWorkflow
	if err := db.Where("status = ? AND cron_expression IS NOT NULL AND cron_expression != ?", 1, "").Find(&workflows).Error; err != nil {
		return err
	}

	cronScheduler := scheduler.GetScheduler(nil)
	count := 0
	for _, workflow := range workflows {
		if workflow.CronExpression != nil && *workflow.CronExpression == "" && workflow.Status == 1 {
			continue
		}
		var nodes []*model.TaskNode
		if err := db.Where("workflow_id = ?", workflow.ID).Find(&nodes).Error; err != nil {
			continue
		}
		var edges []*model.TaskEdge
		if err := db.Where("workflow_id = ?", workflow.ID).Find(&edges).Error; err != nil {
			continue
		}
		err := cronScheduler.AddWorkflow(workflow, nodes, edges)
		if err != nil {
			log.Printf("璋冨害澶辫触锛?v", err.Error())
			continue
		}
		count++
	}

	logs.Info(nil, fmt.Sprintf("宸插姞杞?%d 涓畾鏃跺伐浣滄祦", count))
	return nil
}

