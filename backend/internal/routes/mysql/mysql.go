package mysql

import (
	"devops-console-backend/internal/mysql/api/handler"
	mysqlrouter "devops-console-backend/internal/mysql/api/router"
	"devops-console-backend/internal/mysql/service"

	"github.com/gin-gonic/gin"
)

func RegisterMySQLRouters(apiGroup *gin.RouterGroup) {
	connectionManager := service.NewConnectionManager()
	connectionHandler := handler.NewConnectionHandler(connectionManager)
	metadataHandler := handler.NewMetadataHandler(service.NewMetadataService())
	dataHandler := handler.NewDataHandler(service.NewDataService())
	queryHandler := handler.NewQueryHandler(service.NewQueryService())
	securityHandler := handler.NewSecurityHandler(service.NewSecurityService())
	backupHandler := handler.NewBackupHandler(service.NewBackupService("."))
	schemaCompareHandler := handler.NewSchemaCompareHandler(service.NewSchemaCompareService())

	mysqlrouter.Register(
		apiGroup,
		connectionHandler,
		connectionManager,
		metadataHandler,
		dataHandler,
		queryHandler,
		securityHandler,
		backupHandler,
		schemaCompareHandler,
	)
}
