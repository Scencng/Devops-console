package kafka

import (
	kafkactrl "devops-console-backend/internal/controllers/kafka"

	"github.com/gin-gonic/gin"
)

func RegisterKafkaRouters(apiGroup *gin.RouterGroup) {
	controller := kafkactrl.NewController()
	kafkaGroup := apiGroup.Group("/kafka")
	{
		clusterGroup := kafkaGroup.Group("/clusters")
		{
			clusterGroup.GET("", controller.ListClusters)
			clusterGroup.GET("/options", controller.ListClusterOptions)
			clusterGroup.GET("/:id", controller.GetCluster)
			clusterGroup.POST("", controller.CreateCluster)
			clusterGroup.PUT("/:id", controller.UpdateCluster)
			clusterGroup.DELETE("/:id", controller.DeleteCluster)
			clusterGroup.POST("/:id/test", controller.TestClusterConnection)
		}

		kafkaGroup.GET("/dashboard", controller.GetDashboard)
		kafkaGroup.POST("/topics", controller.CreateTopic)
		kafkaGroup.GET("/topics", controller.ListTopics)
		kafkaGroup.DELETE("/topics/:topic", controller.DeleteTopic)
		kafkaGroup.PUT("/topics/:topic/config", controller.UpdateTopicConfig)
		kafkaGroup.GET("/topics/:topic/partitions", controller.GetTopicPartitions)
		kafkaGroup.POST("/topics/:topic/partitions", controller.IncreaseTopicPartitions)
		kafkaGroup.GET("/brokers", controller.ListBrokers)
		kafkaGroup.GET("/consumer-groups", controller.ListConsumerGroups)
		kafkaGroup.GET("/consumer-groups/:groupId", controller.GetConsumerGroupDetail)
		kafkaGroup.POST("/consumer-groups/:groupId/reset-offset", controller.ResetConsumerGroupOffset)
		kafkaGroup.GET("/messages", controller.BrowseMessages)
		kafkaGroup.POST("/messages/produce", controller.ProduceMessage)
		kafkaGroup.GET("/audit-logs", controller.ListAuditLogs)
		kafkaGroup.POST("/discovery/scan", controller.ScanKafkaNetwork)
		kafkaGroup.POST("/discovery/probe", controller.ProbeKafkaBootstrapServers)
		kafkaGroup.POST("/discovery/import", controller.ImportDiscoveredKafka)
	}
}
