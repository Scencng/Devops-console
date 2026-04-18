# Kafka Console

这是一个在原始 DevOps 平台仓库中新增的 Kafka 子项目，聚焦 Kafka 运维控制台核心能力。

当前保留的核心页面：
- Kafka 总览
- 集群管理
- 自动发现
- Topic 管理
- Broker 管理
- Consumer Group 管理
- 消息浏览
- 审计日志
- Prometheus 面板

初始化 SQL：
- `backend/sql/kafka_console_init.sql`
- `backend/sql/kafka_module.sql`

部署方式：
- 源码部署：`docker-compose.yml` + `deploy.sh`
- 预构建包部署：`docker-compose.prebuilt.yml` + `release.sh`
