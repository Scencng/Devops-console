// 璺敱灞?绠＄悊绋嬪簭鐨勮矾鐢变俊鎭?
package routers

import (
	"devops-console-backend/internal/routes/asset"
	"devops-console-backend/internal/routes/cicd"
	"devops-console-backend/internal/routes/es/backup"
	"devops-console-backend/internal/routes/es/elasticsearch"
	"devops-console-backend/internal/routes/es/indices"
	"devops-console-backend/internal/routes/es/instance"
	"devops-console-backend/internal/routes/es/node"
	"devops-console-backend/internal/routes/es/shard"
	"devops-console-backend/internal/routes/helm"
	"devops-console-backend/internal/routes/k8s"
	"devops-console-backend/internal/routes/kafka"
	"devops-console-backend/internal/routes/monitor"
	"devops-console-backend/internal/routes/mysql"
	"devops-console-backend/internal/routes/system"
	"devops-console-backend/internal/routes/task_scheduler"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRouters 娉ㄥ唽璺敱鐨勬柟娉?
func RegisterRouters(r *gin.Engine, db *gorm.DB) {
	apiGroup := r.Group("/api/v1")
	{
		elasticsearch.RegisterSubRouter(apiGroup)
		backup.RegisterSubRouter(apiGroup)
		instance.RegisterSubRouter(apiGroup)
		node.RegisterSubRouter(apiGroup)
		shard.RegisterSubRouter(apiGroup)
		indices.RegisterSubRouter(apiGroup)

		k8s.RegisterK8sRoutes(apiGroup, db)

		helmRoute := helm.NewHelmRoute(db)
		helmRoute.RegisterSubRouter(apiGroup)

		system.RegisterSystemRouters(apiGroup)
		kafka.RegisterKafkaRouters(apiGroup)
		mysql.RegisterMySQLRouters(apiGroup)
		monitor.RegisterMonitorRouters(apiGroup, db)
		asset.RegisterAssetRouters(apiGroup, db)
		cicd.RegisterCiCdRouters(apiGroup)
		task_scheduler.RegisterTaskSchedulerRouters(apiGroup, db)
	}
}
