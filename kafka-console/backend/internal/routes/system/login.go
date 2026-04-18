package system

import (
	sysCtrl "devops-console-backend/internal/controllers/system"
	"devops-console-backend/internal/dal/mapper"
	redisdao "devops-console-backend/internal/dal/redis"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/database"
	jwtutil "devops-console-backend/pkg/utils/jwt"

	"github.com/gin-gonic/gin"
)

func RegisterLoginRoutes(router *gin.RouterGroup) {
	userMapper := mapper.NewUserMapper(configs.GORMDB)
	redisClient := redisdao.NewClient(database.GetRedisClient())
	blackListManager := jwtutil.NewBlackListManager(redisClient)
	loginController := sysCtrl.NewLoginController(userMapper, redisClient, blackListManager)
	systemGroup := router.Group("/system")
	{
		systemGroup.POST("/login", loginController.Login)
	}
}
