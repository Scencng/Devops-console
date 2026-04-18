// 路由层 管理程序的路由信息
package routers

import (
	"devops-console-backend/internal/routes/kafka"
	"devops-console-backend/internal/routes/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRouters(r *gin.Engine, db *gorm.DB) {
	apiGroup := r.Group("/api/v1")
	{
		system.RegisterSystemRouters(apiGroup, db)
		kafka.RegisterKafkaRouters(apiGroup)
	}
}
