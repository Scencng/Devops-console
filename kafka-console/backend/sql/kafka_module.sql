SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- Kafka module bootstrap SQL
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

SET @ddl = IF (
  EXISTS (
    SELECT 1
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'kafka_clusters'
      AND COLUMN_NAME = 'environment'
  ),
  'SELECT 1',
  'ALTER TABLE `kafka_clusters` ADD COLUMN `environment` varchar(64) DEFAULT NULL'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl = IF (
  EXISTS (
    SELECT 1
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'kafka_clusters'
      AND COLUMN_NAME = 'tenant'
  ),
  'SELECT 1',
  'ALTER TABLE `kafka_clusters` ADD COLUMN `tenant` varchar(64) DEFAULT NULL'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS `kafka_alert_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(191) NOT NULL,
  `metric_type` varchar(64) NOT NULL,
  `severity` varchar(32) NOT NULL DEFAULT 'warning',
  `threshold` double NOT NULL DEFAULT '0',
  `operator` varchar(8) NOT NULL DEFAULT '>',
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `runbook` text,
  `environment` varchar(64) DEFAULT NULL,
  `tenant` varchar(64) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_alert_rules_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_alert_events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `rule_id` bigint unsigned DEFAULT NULL,
  `title` varchar(255) NOT NULL,
  `severity` varchar(32) NOT NULL DEFAULT 'warning',
  `status` varchar(32) NOT NULL DEFAULT 'open',
  `metric_type` varchar(64) NOT NULL,
  `metric_value` double NOT NULL DEFAULT '0',
  `threshold` double NOT NULL DEFAULT '0',
  `message` text,
  `runbook` text,
  `environment` varchar(64) DEFAULT NULL,
  `tenant` varchar(64) DEFAULT NULL,
  `acked_by` varchar(128) DEFAULT NULL,
  `resolved_by` varchar(128) DEFAULT NULL,
  `first_triggered_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `last_triggered_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `acked_at` datetime(3) DEFAULT NULL,
  `resolved_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_alert_events_cluster` (`cluster_id`),
  KEY `idx_kafka_alert_events_rule` (`rule_id`),
  KEY `idx_kafka_alert_events_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_inspection_reports` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(191) NOT NULL,
  `status` varchar(32) NOT NULL DEFAULT 'success',
  `summary` text,
  `issue_count` int NOT NULL DEFAULT '0',
  `environment` varchar(64) DEFAULT NULL,
  `tenant` varchar(64) DEFAULT NULL,
  `triggered_by` varchar(128) DEFAULT NULL,
  `executed_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_inspection_reports_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_inspection_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `report_id` bigint unsigned NOT NULL,
  `check_code` varchar(64) NOT NULL,
  `severity` varchar(32) NOT NULL DEFAULT 'info',
  `status` varchar(32) NOT NULL DEFAULT 'ok',
  `title` varchar(255) NOT NULL,
  `detail` text,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_inspection_items_report` (`report_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_tasks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(191) NOT NULL,
  `task_type` varchar(64) NOT NULL,
  `payload` longtext,
  `cron_expr` varchar(128) DEFAULT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `last_run_status` varchar(32) DEFAULT NULL,
  `environment` varchar(64) DEFAULT NULL,
  `tenant` varchar(64) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_tasks_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_task_runs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `task_id` bigint unsigned NOT NULL,
  `cluster_id` bigint unsigned NOT NULL,
  `status` varchar(32) NOT NULL DEFAULT 'running',
  `trigger_mode` varchar(32) NOT NULL DEFAULT 'manual',
  `result_summary` text,
  `result_payload` longtext,
  `started_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `finished_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_kafka_task_runs_task` (`task_id`),
  KEY `idx_kafka_task_runs_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_change_requests` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `change_type` varchar(64) NOT NULL,
  `resource_type` varchar(64) NOT NULL,
  `resource_name` varchar(255) NOT NULL,
  `payload` longtext,
  `reason` text,
  `status` varchar(32) NOT NULL DEFAULT 'pending',
  `requester_user_id` bigint unsigned NOT NULL DEFAULT '0',
  `requester_username` varchar(128) DEFAULT NULL,
  `approver_user_id` bigint unsigned NOT NULL DEFAULT '0',
  `approver_username` varchar(128) DEFAULT NULL,
  `approval_comment` text,
  `environment` varchar(64) DEFAULT NULL,
  `tenant` varchar(64) DEFAULT NULL,
  `executed_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_change_requests_cluster` (`cluster_id`),
  KEY `idx_kafka_change_requests_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_topic_metadata` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `topic_name` varchar(255) NOT NULL,
  `system_name` varchar(128) DEFAULT NULL,
  `owner` varchar(128) DEFAULT NULL,
  `owner_email` varchar(191) DEFAULT NULL,
  `environment` varchar(64) DEFAULT NULL,
  `tenant` varchar(64) DEFAULT NULL,
  `lifecycle` varchar(64) DEFAULT NULL,
  `sensitivity` varchar(64) DEFAULT NULL,
  `description` text,
  `labels` longtext,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_kafka_topic_metadata_cluster_topic` (`cluster_id`,`topic_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_schema_registries` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(191) NOT NULL,
  `endpoint` varchar(500) NOT NULL,
  `auth_type` varchar(32) NOT NULL DEFAULT 'none',
  `username` varchar(191) DEFAULT NULL,
  `password_ciphertext` text,
  `verify_ssl` tinyint(1) NOT NULL DEFAULT '1',
  `environment` varchar(64) DEFAULT NULL,
  `tenant` varchar(64) DEFAULT NULL,
  `description` text,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_schema_registries_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_alert_silences` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(191) NOT NULL,
  `metric_type` varchar(64) DEFAULT NULL,
  `severity` varchar(32) DEFAULT NULL,
  `starts_at` datetime(3) NOT NULL,
  `ends_at` datetime(3) NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `comment` text,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_alert_silences_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_trace_links` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `trace_id` varchar(191) NOT NULL,
  `span_id` varchar(191) DEFAULT NULL,
  `service_name` varchar(191) DEFAULT NULL,
  `topic` varchar(255) DEFAULT NULL,
  `message_key` varchar(255) DEFAULT NULL,
  `consumer_group` varchar(255) DEFAULT NULL,
  `headers` longtext,
  `description` text,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_trace_links_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_scaling_recommendations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `resource_type` varchar(64) NOT NULL,
  `resource_name` varchar(255) NOT NULL,
  `current_value` double NOT NULL DEFAULT '0',
  `recommended_value` double NOT NULL DEFAULT '0',
  `reason` text,
  `status` varchar(32) NOT NULL DEFAULT 'open',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_scaling_recommendations_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_self_healing_policies` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(191) NOT NULL,
  `trigger_type` varchar(64) NOT NULL,
  `action_type` varchar(64) NOT NULL,
  `config` longtext,
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_self_healing_policies_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_self_healing_executions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `policy_id` bigint unsigned NOT NULL,
  `cluster_id` bigint unsigned NOT NULL,
  `status` varchar(32) NOT NULL DEFAULT 'running',
  `summary` text,
  `result` longtext,
  `started_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `completed_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_kafka_self_healing_executions_policy` (`policy_id`),
  KEY `idx_kafka_self_healing_executions_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_gitops_profiles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(191) NOT NULL,
  `repo_url` varchar(500) NOT NULL,
  `branch` varchar(128) DEFAULT NULL,
  `base_path` varchar(255) DEFAULT NULL,
  `manifest_format` varchar(64) DEFAULT NULL,
  `auth_type` varchar(32) DEFAULT NULL,
  `token_ciphertext` text,
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `last_sync_status` varchar(32) DEFAULT NULL,
  `last_sync_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_gitops_profiles_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_gitops_sync_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `profile_id` bigint unsigned NOT NULL,
  `status` varchar(32) NOT NULL DEFAULT 'running',
  `summary` text,
  `commit_sha` varchar(128) DEFAULT NULL,
  `output` longtext,
  `started_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `finished_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_kafka_gitops_sync_records_profile` (`profile_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_cloud_adapters` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `provider` varchar(64) NOT NULL,
  `service_name` varchar(128) NOT NULL,
  `region` varchar(64) DEFAULT NULL,
  `cluster_identifier` varchar(191) DEFAULT NULL,
  `endpoint_mode` varchar(64) DEFAULT NULL,
  `notes` text,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_cloud_adapters_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_lineage_relations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `source_topic` varchar(255) NOT NULL,
  `target_topic` varchar(255) NOT NULL,
  `relation_type` varchar(64) NOT NULL,
  `producer_service` varchar(191) DEFAULT NULL,
  `consumer_service` varchar(191) DEFAULT NULL,
  `description` text,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_lineage_relations_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_lifecycle_policies` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `topic_name` varchar(255) NOT NULL,
  `action` varchar(64) DEFAULT NULL,
  `target_retention_hours` int DEFAULT NULL,
  `owner` varchar(128) DEFAULT NULL,
  `status` varchar(32) DEFAULT NULL,
  `recommendation` text,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_lifecycle_policies_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_mesh_gateway_configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `gateway_type` varchar(64) NOT NULL,
  `endpoint` varchar(500) DEFAULT NULL,
  `auth_mode` varchar(64) DEFAULT NULL,
  `config` longtext,
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_mesh_gateway_configs_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_cost_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `metric_date` datetime(3) NOT NULL,
  `storage_bytes` double DEFAULT NULL,
  `ingress_bytes` double DEFAULT NULL,
  `egress_bytes` double DEFAULT NULL,
  `partition_count` int DEFAULT NULL,
  `estimated_cost` double DEFAULT NULL,
  `currency` varchar(16) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_cost_records_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_sensitive_scan_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(191) NOT NULL,
  `pattern_type` varchar(64) NOT NULL,
  `pattern_value` varchar(500) NOT NULL,
  `severity` varchar(32) DEFAULT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_sensitive_scan_rules_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `kafka_sensitive_scan_results` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `topic` varchar(255) NOT NULL,
  `partition` int NOT NULL,
  `offset_value` bigint NOT NULL,
  `rule_name` varchar(191) DEFAULT NULL,
  `severity` varchar(32) DEFAULT NULL,
  `matched_text` varchar(500) DEFAULT NULL,
  `summary` text,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_kafka_sensitive_scan_results_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO `sys_menu` (`id`, `parent_id`, `name`, `type`, `path`, `component`, `icon`, `perm`, `sort`, `visible`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES
(7000, 0, 'Kafka', 1, '', '', 'Connection', NULL, 20, 1, 1, NULL, NULL, NULL),
(7001, 7000, 'Kafka 总览', 2, '/kafka', 'kafka/KafkaDashboard', 'DataBoard', NULL, 10, 1, 1, NULL, NULL, NULL),
(7010, 7000, '健康总览', 2, '/kafka/health', 'kafka/HealthOverview', 'DataBoard', NULL, 15, 1, 1, NULL, NULL, NULL),
(7002, 7000, '集群管理', 2, '/kafka/clusters', 'kafka/ClusterManagement', 'Connection', NULL, 20, 1, 1, NULL, NULL, NULL),
(7003, 7000, 'Topic 管理', 2, '/kafka/topics', 'kafka/TopicManagement', 'DocumentCopy', NULL, 30, 1, 1, NULL, NULL, NULL),
(7004, 7000, 'Broker 管理', 2, '/kafka/brokers', 'kafka/BrokerManagement', 'Monitor', NULL, 40, 1, 1, NULL, NULL, NULL),
(7005, 7000, '消费组管理', 2, '/kafka/groups', 'kafka/ConsumerGroupManagement', 'Histogram', NULL, 50, 1, 1, NULL, NULL, NULL),
(7006, 7000, '消息浏览', 2, '/kafka/messages', 'kafka/MessageBrowser', 'Search', NULL, 60, 1, 1, NULL, NULL, NULL),
(7007, 7000, 'Prometheus 面板', 2, '/kafka/prometheus', 'kafka/PrometheusPanel', 'TrendCharts', NULL, 70, 1, 1, NULL, NULL, NULL),
(7011, 7000, '趋势分析', 2, '/kafka/trends', 'kafka/TrendAnalysis', 'TrendCharts', NULL, 75, 1, 1, NULL, NULL, NULL),
(7008, 7000, '审计日志', 2, '/kafka/audits', 'kafka/AuditLog', 'Document', NULL, 80, 1, 1, NULL, NULL, NULL),
(7009, 7000, '安全管理', 2, '/kafka/security', 'kafka/SecurityManagement', 'Lock', NULL, 90, 1, 1, NULL, NULL, NULL),
(7012, 7000, '告警中心', 2, '/kafka/alerts', 'kafka/AlertCenter', 'Bell', NULL, 100, 1, 1, NULL, NULL, NULL),
(7013, 7000, '一键巡检', 2, '/kafka/inspection', 'kafka/InspectionCenter', 'Operation', NULL, 110, 1, 1, NULL, NULL, NULL),
(7014, 7000, '任务编排', 2, '/kafka/tasks', 'kafka/TaskOrchestration', 'List', NULL, 120, 1, 1, NULL, NULL, NULL),
(7015, 7000, '变更审批', 2, '/kafka/approvals', 'kafka/ChangeApproval', 'Stamp', NULL, 130, 1, 1, NULL, NULL, NULL),
(7016, 7000, 'Topic 元数据', 2, '/kafka/metadata', 'kafka/TopicMetadataCenter', 'Collection', NULL, 140, 1, 1, NULL, NULL, NULL),
(7017, 7000, 'Schema Registry', 2, '/kafka/schema-registry', 'kafka/SchemaRegistryCenter', 'Tickets', NULL, 150, 1, 1, NULL, NULL, NULL),
(7018, 7000, '告警降噪', 2, '/kafka/alert-noise', 'kafka/AlertNoiseControl', 'BellFilled', NULL, 160, 1, 1, NULL, NULL, NULL),
(7019, 7000, '链路追踪', 2, '/kafka/traces', 'kafka/TraceCorrelation', 'Share', NULL, 170, 1, 1, NULL, NULL, NULL),
(7020, 7000, '扩容建议', 2, '/kafka/scaling', 'kafka/ScalingAdvisor', 'TrendCharts', NULL, 180, 1, 1, NULL, NULL, NULL),
(7021, 7000, '自愈中心', 2, '/kafka/self-healing', 'kafka/SelfHealingCenter', 'Tools', NULL, 190, 1, 1, NULL, NULL, NULL),
(7022, 7000, 'GitOps 中心', 2, '/kafka/gitops', 'kafka/GitOpsCenter', 'Files', NULL, 200, 1, 1, NULL, NULL, NULL),
(7023, 7000, '云托管适配', 2, '/kafka/cloud-adapters', 'kafka/CloudAdapterCenter', 'Cloudy', NULL, 210, 1, 1, NULL, NULL, NULL),
(7024, 7000, '数据血缘', 2, '/kafka/lineage', 'kafka/DataLineageCenter', 'Connection', NULL, 220, 1, 1, NULL, NULL, NULL),
(7025, 7000, '生命周期治理', 2, '/kafka/lifecycle', 'kafka/LifecycleGovernance', 'Calendar', NULL, 230, 1, 1, NULL, NULL, NULL),
(7026, 7000, '网关/Mesh', 2, '/kafka/mesh-gateway', 'kafka/MeshGatewayCenter', 'Switch', NULL, 240, 1, 1, NULL, NULL, NULL),
(7027, 7000, '成本治理', 2, '/kafka/cost', 'kafka/CostGovernance', 'Money', NULL, 250, 1, 1, NULL, NULL, NULL),
(7028, 7000, '敏感识别', 2, '/kafka/sensitive-data', 'kafka/SensitiveDataCenter', 'Warning', NULL, 260, 1, 1, NULL, NULL, NULL),
(7029, 7000, '自动发现', 2, '/kafka/discovery', 'kafka/DiscoveryCenter', 'Search', NULL, 265, 1, 1, NULL, NULL, NULL),
(7101, 7002, '新增 Kafka 集群', 3, NULL, NULL, NULL, 'kafka:cluster:create', 10, 1, 1, NULL, NULL, NULL),
(7102, 7002, '编辑 Kafka 集群', 3, NULL, NULL, NULL, 'kafka:cluster:edit', 20, 1, 1, NULL, NULL, NULL),
(7103, 7002, '删除 Kafka 集群', 3, NULL, NULL, NULL, 'kafka:cluster:delete', 30, 1, 1, NULL, NULL, NULL),
(7104, 7002, '测试 Kafka 集群', 3, NULL, NULL, NULL, 'kafka:cluster:test', 40, 1, 1, NULL, NULL, NULL),
(7105, 7003, '创建 Topic', 3, NULL, NULL, NULL, 'kafka:topic:create', 10, 1, 1, NULL, NULL, NULL),
(7106, 7003, '修改 Topic 配置', 3, NULL, NULL, NULL, 'kafka:topic:config:update', 20, 1, 1, NULL, NULL, NULL),
(7107, 7003, '扩容 Topic 分区', 3, NULL, NULL, NULL, 'kafka:topic:partitions:increase', 30, 1, 1, NULL, NULL, NULL),
(7108, 7003, '删除 Topic', 3, NULL, NULL, NULL, 'kafka:topic:delete', 40, 1, 1, NULL, NULL, NULL),
(7109, 7005, '重置消费组 Offset', 3, NULL, NULL, NULL, 'kafka:group:offset:reset', 10, 1, 1, NULL, NULL, NULL),
(7110, 7006, '发送 Kafka 消息', 3, NULL, NULL, NULL, 'kafka:message:produce', 10, 1, 1, NULL, NULL, NULL),
(7111, 7009, '创建 ACL', 3, NULL, NULL, NULL, 'kafka:acl:create', 10, 1, 1, NULL, NULL, NULL),
(7112, 7009, '删除 ACL', 3, NULL, NULL, NULL, 'kafka:acl:delete', 20, 1, 1, NULL, NULL, NULL),
(7113, 7009, '保存 SCRAM 用户', 3, NULL, NULL, NULL, 'kafka:scram:user:upsert', 30, 1, 1, NULL, NULL, NULL),
(7114, 7009, '删除 SCRAM 用户', 3, NULL, NULL, NULL, 'kafka:scram:user:delete', 40, 1, 1, NULL, NULL, NULL),
(7115, 7012, '创建告警规则', 3, NULL, NULL, NULL, 'kafka:alert:rule:create', 10, 1, 1, NULL, NULL, NULL),
(7116, 7012, '编辑告警规则', 3, NULL, NULL, NULL, 'kafka:alert:rule:update', 20, 1, 1, NULL, NULL, NULL),
(7117, 7012, '删除告警规则', 3, NULL, NULL, NULL, 'kafka:alert:rule:delete', 30, 1, 1, NULL, NULL, NULL),
(7118, 7012, '评估告警规则', 3, NULL, NULL, NULL, 'kafka:alert:evaluate', 40, 1, 1, NULL, NULL, NULL),
(7119, 7012, '处理告警事件', 3, NULL, NULL, NULL, 'kafka:alert:event:update', 50, 1, 1, NULL, NULL, NULL),
(7120, 7013, '执行巡检', 3, NULL, NULL, NULL, 'kafka:inspection:run', 10, 1, 1, NULL, NULL, NULL),
(7121, 7014, '创建任务', 3, NULL, NULL, NULL, 'kafka:task:create', 10, 1, 1, NULL, NULL, NULL),
(7122, 7014, '编辑任务', 3, NULL, NULL, NULL, 'kafka:task:update', 20, 1, 1, NULL, NULL, NULL),
(7123, 7014, '删除任务', 3, NULL, NULL, NULL, 'kafka:task:delete', 30, 1, 1, NULL, NULL, NULL),
(7124, 7014, '执行任务', 3, NULL, NULL, NULL, 'kafka:task:run', 40, 1, 1, NULL, NULL, NULL),
(7125, 7015, '创建变更单', 3, NULL, NULL, NULL, 'kafka:change:create', 10, 1, 1, NULL, NULL, NULL),
(7126, 7015, '审批变更单', 3, NULL, NULL, NULL, 'kafka:change:review', 20, 1, 1, NULL, NULL, NULL),
(7127, 7015, '执行变更单', 3, NULL, NULL, NULL, 'kafka:change:execute', 30, 1, 1, NULL, NULL, NULL),
(7128, 7016, '创建 Topic 元数据', 3, NULL, NULL, NULL, 'kafka:metadata:create', 10, 1, 1, NULL, NULL, NULL),
(7129, 7016, '编辑 Topic 元数据', 3, NULL, NULL, NULL, 'kafka:metadata:update', 20, 1, 1, NULL, NULL, NULL),
(7130, 7016, '删除 Topic 元数据', 3, NULL, NULL, NULL, 'kafka:metadata:delete', 30, 1, 1, NULL, NULL, NULL),
(7131, 7017, '创建 Schema Registry', 3, NULL, NULL, NULL, 'kafka:schema_registry:create', 10, 1, 1, NULL, NULL, NULL),
(7132, 7017, '编辑 Schema Registry', 3, NULL, NULL, NULL, 'kafka:schema_registry:update', 20, 1, 1, NULL, NULL, NULL),
(7133, 7017, '删除 Schema Registry', 3, NULL, NULL, NULL, 'kafka:schema_registry:delete', 30, 1, 1, NULL, NULL, NULL),
(7134, 7017, '校验 Schema 兼容性', 3, NULL, NULL, NULL, 'kafka:schema_registry:compatibility', 40, 1, 1, NULL, NULL, NULL),
(7135, 7018, '创建静默规则', 3, NULL, NULL, NULL, 'kafka:alert:silence:create', 10, 1, 1, NULL, NULL, NULL),
(7136, 7018, '编辑静默规则', 3, NULL, NULL, NULL, 'kafka:alert:silence:update', 20, 1, 1, NULL, NULL, NULL),
(7137, 7018, '删除静默规则', 3, NULL, NULL, NULL, 'kafka:alert:silence:delete', 30, 1, 1, NULL, NULL, NULL),
(7138, 7019, '创建 Trace 关联', 3, NULL, NULL, NULL, 'kafka:trace:create', 10, 1, 1, NULL, NULL, NULL),
(7139, 7019, '删除 Trace 关联', 3, NULL, NULL, NULL, 'kafka:trace:delete', 20, 1, 1, NULL, NULL, NULL),
(7140, 7021, '创建自愈策略', 3, NULL, NULL, NULL, 'kafka:self_healing:create', 10, 1, 1, NULL, NULL, NULL),
(7141, 7021, '编辑自愈策略', 3, NULL, NULL, NULL, 'kafka:self_healing:update', 20, 1, 1, NULL, NULL, NULL),
(7142, 7021, '删除自愈策略', 3, NULL, NULL, NULL, 'kafka:self_healing:delete', 30, 1, 1, NULL, NULL, NULL),
(7143, 7021, '执行自愈策略', 3, NULL, NULL, NULL, 'kafka:self_healing:execute', 40, 1, 1, NULL, NULL, NULL),
(7144, 7022, '创建 GitOps 配置', 3, NULL, NULL, NULL, 'kafka:gitops:create', 10, 1, 1, NULL, NULL, NULL),
(7145, 7022, '编辑 GitOps 配置', 3, NULL, NULL, NULL, 'kafka:gitops:update', 20, 1, 1, NULL, NULL, NULL),
(7146, 7022, '删除 GitOps 配置', 3, NULL, NULL, NULL, 'kafka:gitops:delete', 30, 1, 1, NULL, NULL, NULL),
(7147, 7022, '执行 GitOps 同步', 3, NULL, NULL, NULL, 'kafka:gitops:sync', 40, 1, 1, NULL, NULL, NULL),
(7148, 7023, '创建云适配', 3, NULL, NULL, NULL, 'kafka:cloud_adapter:create', 10, 1, 1, NULL, NULL, NULL),
(7149, 7023, '编辑云适配', 3, NULL, NULL, NULL, 'kafka:cloud_adapter:update', 20, 1, 1, NULL, NULL, NULL),
(7150, 7023, '删除云适配', 3, NULL, NULL, NULL, 'kafka:cloud_adapter:delete', 30, 1, 1, NULL, NULL, NULL),
(7151, 7024, '创建血缘关系', 3, NULL, NULL, NULL, 'kafka:lineage:create', 10, 1, 1, NULL, NULL, NULL),
(7152, 7024, '编辑血缘关系', 3, NULL, NULL, NULL, 'kafka:lineage:update', 20, 1, 1, NULL, NULL, NULL),
(7153, 7024, '删除血缘关系', 3, NULL, NULL, NULL, 'kafka:lineage:delete', 30, 1, 1, NULL, NULL, NULL),
(7154, 7026, '创建网关配置', 3, NULL, NULL, NULL, 'kafka:mesh_gateway:create', 10, 1, 1, NULL, NULL, NULL),
(7155, 7026, '编辑网关配置', 3, NULL, NULL, NULL, 'kafka:mesh_gateway:update', 20, 1, 1, NULL, NULL, NULL),
(7156, 7026, '删除网关配置', 3, NULL, NULL, NULL, 'kafka:mesh_gateway:delete', 30, 1, 1, NULL, NULL, NULL),
(7157, 7028, '创建敏感规则', 3, NULL, NULL, NULL, 'kafka:sensitive_rule:create', 10, 1, 1, NULL, NULL, NULL),
(7158, 7028, '编辑敏感规则', 3, NULL, NULL, NULL, 'kafka:sensitive_rule:update', 20, 1, 1, NULL, NULL, NULL),
(7159, 7028, '删除敏感规则', 3, NULL, NULL, NULL, 'kafka:sensitive_rule:delete', 30, 1, 1, NULL, NULL, NULL),
(7160, 7028, '执行敏感扫描', 3, NULL, NULL, NULL, 'kafka:sensitive_scan:run', 40, 1, 1, NULL, NULL, NULL),
(7161, 7029, '扫描 Kafka 网络', 3, NULL, NULL, NULL, 'kafka:discovery:scan', 10, 1, 1, NULL, NULL, NULL),
(7162, 7029, '导入发现结果', 3, NULL, NULL, NULL, 'kafka:discovery:import', 20, 1, 1, NULL, NULL, NULL);

INSERT IGNORE INTO `sys_role_menu` (`role_id`, `menu_id`) VALUES
(1,7000),(1,7001),(1,7010),(1,7002),(1,7003),(1,7004),(1,7005),(1,7006),(1,7007),(1,7011),(1,7008),(1,7009),(1,7012),(1,7013),(1,7014),(1,7015),(1,7016),(1,7017),
(1,7018),(1,7019),(1,7020),(1,7021),(1,7022),(1,7023),(1,7024),(1,7025),(1,7026),(1,7027),(1,7028),(1,7029),
(1,7101),(1,7102),(1,7103),(1,7104),(1,7105),(1,7106),(1,7107),(1,7108),(1,7109),(1,7110),(1,7111),(1,7112),(1,7113),(1,7114),
(1,7115),(1,7116),(1,7117),(1,7118),(1,7119),(1,7120),(1,7121),(1,7122),(1,7123),(1,7124),(1,7125),(1,7126),(1,7127),(1,7128),(1,7129),(1,7130),(1,7131),(1,7132),(1,7133),(1,7134),
(1,7135),(1,7136),(1,7137),(1,7138),(1,7139),(1,7140),(1,7141),(1,7142),(1,7143),(1,7144),(1,7145),(1,7146),(1,7147),(1,7148),(1,7149),(1,7150),(1,7151),(1,7152),(1,7153),(1,7154),(1,7155),(1,7156),(1,7157),(1,7158),(1,7159),(1,7160),(1,7161),(1,7162);
