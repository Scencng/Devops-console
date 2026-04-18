SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

CREATE TABLE IF NOT EXISTS `kafka_clusters` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(191) NOT NULL,
  `bootstrap_servers` varchar(2000) NOT NULL,
  `version` varchar(50) NOT NULL DEFAULT '3.6.0',
  `auth_type` varchar(50) NOT NULL DEFAULT 'none',
  `username` varchar(255) DEFAULT NULL,
  `password_ciphertext` longtext,
  `tls_enabled` tinyint(1) NOT NULL DEFAULT '0',
  `insecure_skip_verify` tinyint(1) NOT NULL DEFAULT '0',
  `ca_cert` longtext,
  `client_cert` longtext,
  `client_key_ciphertext` longtext,
  `description` text,
  `environment` varchar(64) DEFAULT NULL,
  `tenant` varchar(64) DEFAULT NULL,
  `status` varchar(50) NOT NULL DEFAULT 'unknown',
  `last_error_message` text,
  `last_tested_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_kafka_clusters_name` (`name`),
  KEY `idx_kafka_clusters_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_audit_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `action` varchar(64) NOT NULL,
  `resource_type` varchar(64) NOT NULL,
  `resource_name` varchar(255) NOT NULL,
  `operator_user_id` bigint unsigned NOT NULL,
  `operator_username` varchar(128) NOT NULL,
  `request_payload` longtext,
  `result` varchar(32) NOT NULL DEFAULT 'success',
  `error_message` text,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_audit_logs_cluster` (`cluster_id`),
  KEY `idx_kafka_audit_logs_action` (`action`),
  KEY `idx_kafka_audit_logs_result` (`result`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `perm`, `sort`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES
(7000, 0, 'Kafka', 1, '', '', 'Connection', NULL, 20, 1, 1, NULL, NULL, NULL),
(7001, 7000, 'Kafka 总览', 2, '/kafka', 'kafka/KafkaDashboard', 'DataBoard', NULL, 10, 1, 1, NULL, NULL, NULL),
(7002, 7000, '集群管理', 2, '/kafka/clusters', 'kafka/ClusterManagement', 'Connection', NULL, 20, 1, 1, NULL, NULL, NULL),
(7010, 7000, '自动发现', 2, '/kafka/discovery', 'kafka/DiscoveryCenter', 'Search', NULL, 25, 1, 1, NULL, NULL, NULL),
(7003, 7000, 'Topic 管理', 2, '/kafka/topics', 'kafka/TopicManagement', 'DocumentCopy', NULL, 30, 1, 1, NULL, NULL, NULL),
(7004, 7000, 'Broker 管理', 2, '/kafka/brokers', 'kafka/BrokerManagement', 'Monitor', NULL, 40, 1, 1, NULL, NULL, NULL),
(7005, 7000, '消费组管理', 2, '/kafka/groups', 'kafka/ConsumerGroupManagement', 'Histogram', NULL, 50, 1, 1, NULL, NULL, NULL),
(7006, 7000, '消息浏览', 2, '/kafka/messages', 'kafka/MessageBrowser', 'Search', NULL, 60, 1, 1, NULL, NULL, NULL),
(7007, 7000, 'Prometheus 面板', 2, '/kafka/prometheus', 'kafka/PrometheusPanel', 'TrendCharts', NULL, 70, 1, 1, NULL, NULL, NULL),
(7008, 7000, '审计日志', 2, '/kafka/audits', 'kafka/AuditLog', 'Document', NULL, 80, 1, 1, NULL, NULL, NULL),
(7101, 7002, '新增 Kafka 集群', 3, NULL, NULL, NULL, 'kafka:cluster:create', 10, 1, 1, NULL, NULL, NULL),
(7102, 7002, '编辑 Kafka 集群', 3, NULL, NULL, NULL, 'kafka:cluster:edit', 20, 1, 1, NULL, NULL, NULL),
(7103, 7002, '删除 Kafka 集群', 3, NULL, NULL, NULL, 'kafka:cluster:delete', 30, 1, 1, NULL, NULL, NULL),
(7104, 7002, '测试 Kafka 集群', 3, NULL, NULL, NULL, 'kafka:cluster:test', 40, 1, 1, NULL, NULL, NULL),
(7105, 7003, '修改 Topic 配置', 3, NULL, NULL, NULL, 'kafka:topic:config:update', 10, 1, 1, NULL, NULL, NULL),
(7106, 7003, '删除 Topic', 3, NULL, NULL, NULL, 'kafka:topic:delete', 20, 1, 1, NULL, NULL, NULL),
(7107, 7005, '重置消费组 Offset', 3, NULL, NULL, NULL, 'kafka:group:offset:reset', 10, 1, 1, NULL, NULL, NULL),
(7108, 7003, '创建 Topic', 3, NULL, NULL, NULL, 'kafka:topic:create', 30, 1, 1, NULL, NULL, NULL),
(7109, 7003, '扩容 Topic 分区', 3, NULL, NULL, NULL, 'kafka:topic:partitions:increase', 40, 1, 1, NULL, NULL, NULL),
(7110, 7006, '发送消息', 3, NULL, NULL, NULL, 'kafka:message:produce', 10, 1, 1, NULL, NULL, NULL);

INSERT IGNORE INTO `sys_role_menu` (`role_id`, `menu_id`) VALUES
(1,7000),(1,7001),(1,7002),(1,7003),(1,7004),(1,7005),(1,7006),(1,7007),(1,7008),(1,7010),
(1,7101),(1,7102),(1,7103),(1,7104),(1,7105),(1,7106),(1,7107),(1,7108),(1,7109),(1,7110);

SET FOREIGN_KEY_CHECKS = 1;
