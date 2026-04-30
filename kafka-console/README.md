# Kafka Console

这是一个精简后的 Kafka 运维控制台项目。

## 当前保留功能

- Kafka 总览
- 集群管理
- 自动发现
- Topic 管理
- Broker 管理
- Consumer Group 管理
- 消息浏览与测试消息发送
- 审计日志

## 部署方式

当前项目不再依赖部署脚本。

手动部署只需要：

1. 解压发布包
2. 复制并修改 `.env`
3. 执行 `docker compose up -d --build`

说明：

- `--build` 只会基于包内自带的前端静态资源和 Linux 后端二进制组装镜像
- 不会在服务器上重新编译 Go 源码

详细步骤见：

- [DEPLOY_LINUX.md](./DEPLOY_LINUX.md)

## 目录说明

```text
kafka-console/
├── backend/                   # Go 后端
├── frontend/                  # Vue 3 前端
├── deploy/                    # Nginx / Helm 示例
├── docker-compose.yml         # 源码模式 compose
├── docker-compose.prebuilt.yml# 预构建模式 compose
├── package-release.ps1        # 本地打包脚本
├── .env.example               # 环境变量示例
└── DEPLOY_LINUX.md            # Linux 手动部署说明
```
