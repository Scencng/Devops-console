package router

import (
	"github.com/gin-gonic/gin"

	"devops-console-backend/internal/mysql/api/handler"
	apimiddleware "devops-console-backend/internal/mysql/api/middleware"
	"devops-console-backend/internal/mysql/service"
)

func Register(api *gin.RouterGroup, connectionHandler *handler.ConnectionHandler, manager *service.ConnectionManager, metadataHandler *handler.MetadataHandler, dataHandler *handler.DataHandler, queryHandler *handler.QueryHandler, securityHandler *handler.SecurityHandler, backupHandler *handler.BackupHandler, schemaCompareHandler *handler.SchemaCompareHandler) {
	{
		mysqlGroup := api.Group("/mysql")
		connectionGroup := mysqlGroup.Group("/connection")
		{
			connectionGroup.POST("/open", connectionHandler.Open)
			connectionGroup.POST("/close", connectionHandler.Close)
		}

		authenticated := mysqlGroup.Group("")
		authenticated.Use(apimiddleware.ConnectionToken(manager))
		{
			metadataGroup := authenticated.Group("/metadata")
			{
				metadataGroup.GET("/databases", metadataHandler.ListDatabases)
				metadataGroup.GET("/tables", metadataHandler.ListTables)
				metadataGroup.POST("/database/create", metadataHandler.CreateDatabase)
				metadataGroup.POST("/database/rename", metadataHandler.RenameDatabase)
				metadataGroup.POST("/database/delete", metadataHandler.DeleteDatabase)
				metadataGroup.POST("/table/create", metadataHandler.CreateTable)
				metadataGroup.POST("/table/auto-import", metadataHandler.AutoImportTable)
				metadataGroup.POST("/table/rename", metadataHandler.RenameTable)
				metadataGroup.POST("/table/delete", metadataHandler.DeleteTable)
			}

			dataGroup := authenticated.Group("/data")
			{
				dataGroup.GET("/table", dataHandler.GetTableData)
			}

			queryGroup := authenticated.Group("/query")
			{
				queryGroup.POST("/execute", queryHandler.Execute)
			}

			sqlGroup := authenticated.Group("/sql")
			{
				sqlGroup.POST("/execute-batch", queryHandler.ExecuteBatch)
			}

			dbGroup := authenticated.Group("/db")
			{
				dbGroup.POST("/execute-batch", queryHandler.ExecuteBatch)
			}

			securityGroup := authenticated.Group("/security")
			{
				securityGroup.GET("/overview", securityHandler.Overview)
				securityGroup.GET("/principal", securityHandler.GetPrincipal)
				securityGroup.POST("/principal/create", securityHandler.CreatePrincipal)
				securityGroup.POST("/principal/update", securityHandler.UpdatePrincipal)
				securityGroup.POST("/principal/delete", securityHandler.DeletePrincipal)
				securityGroup.POST("/principal/clone", securityHandler.ClonePrincipal)
				securityGroup.POST("/principal/revoke-all", securityHandler.RevokeAll)
			}

			backupGroup := authenticated.Group("/backup")
			{
				backupGroup.GET("/list", backupHandler.List)
				backupGroup.POST("/create", backupHandler.Create)
				backupGroup.POST("/restore", backupHandler.Restore)
				backupGroup.POST("/rename", backupHandler.Rename)
				backupGroup.POST("/delete", backupHandler.Delete)
				backupGroup.GET("/download", backupHandler.Download)
				backupGroup.GET("/task", backupHandler.Task)
				backupGroup.GET("/schedules", backupHandler.ListSchedules)
				backupGroup.POST("/schedule/create", backupHandler.CreateSchedule)
				backupGroup.POST("/schedule/delete", backupHandler.DeleteSchedule)
			}

			schemaGroup := authenticated.Group("/schema")
			{
				schemaGroup.POST("/compare", schemaCompareHandler.Compare)
			}
		}
	}
}

