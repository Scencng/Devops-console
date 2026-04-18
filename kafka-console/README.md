# Kafka Console

这是一个在当前 `devops` 项目骨架基础上复制并开始开发的独立项目。

目录说明：
- `backend/`：复用原有 Go + Gin + Gorm + RBAC + Docker 骨架，并新增 Kafka 模块
- `frontend/`：复用原有 Vue 3 + Vite + Element Plus + 动态菜单骨架，并新增 Kafka 页面

当前已完成：
- Kafka 集群管理
- Kafka Dashboard
- Topic 查看/删除/配置修改
- Broker 查看
- Consumer Group 查看/Offset 重置
- 消息浏览
- 操作审计日志
- Prometheus 监控面板与自定义 PromQL 查询
- Kafka 菜单与权限 SQL

初始化 SQL：
- `backend/sql/kafka_console_init.sql`
- `backend/sql/kafka_module.sql`
- 如果使用本文的 `./deploy.sh up`，默认会自动导入这两份 SQL
- 如果你不是走本文这套 Docker 脚本部署，而是自行部署数据库/后端，则需要手动执行

## Linux Docker 部署

当前仓库已经带好这几个部署文件：
- `docker-compose.yml`
- `.env.example`
- `deploy.sh`

推荐在 Linux 上直接用这套脚本部署，流程是：
1. 准备服务器和 Docker 运行环境
2. 上传项目代码到服务器
3. 按实际环境修改 `.env`
4. 执行 `./deploy.sh up`
5. 用 `./deploy.sh status`、`./deploy.sh logs` 做验收

### 1. 部署前确认

建议先确认下面几点：
- 服务器系统为 Linux，且建议使用 `x86_64 / amd64`
- 已安装 Docker Engine，并且能使用 `docker compose` 或 `docker-compose`
- Docker 服务已启动，`docker info` 能正常返回
- 服务器能访问镜像仓库、Go/NPM 依赖源
- 如果要从外部访问页面，放通前端端口；默认是 `80`

说明：
- 后端 Dockerfile 当前固定编译为 `GOARCH=amd64`，因此 ARM 服务器默认不在这份部署说明覆盖范围内
- `deploy.sh` 会自动生成 `backend/config/config.yaml`，这个文件每次重新部署都可能被覆盖，不建议手改
- MySQL 和 Redis 数据默认持久化到项目目录下的 `data/mysql`、`data/redis`

### 2. 上传代码到 Linux 服务器

下面是一个常见目录示例：

```bash
sudo mkdir -p /opt/kafka-console
sudo chown -R $USER:$USER /opt/kafka-console
cd /opt/kafka-console
```

然后把本项目完整上传到这个目录，最终结构至少应包含：

```text
/opt/kafka-console
├── backend
├── frontend
├── .env.example
├── deploy.sh
└── docker-compose.yml
```

如果你是在 Windows 上先打包再传 Linux，可以继续使用仓库里的 `upload-release.ps1`，见下文“Windows 一键打包上传”。

### 3. 配置 `.env`

先复制一份环境文件：

```bash
cd /opt/kafka-console
cp .env.example .env
```

再按实际环境修改：

```bash
vim .env
```

最关键的配置项如下：

| 变量 | 说明 | 建议 |
| --- | --- | --- |
| `FRONTEND_BIND` | 前端监听地址 | 直接公网访问时用 `0.0.0.0`；宿主机 Nginx 反代时用 `127.0.0.1` |
| `FRONTEND_PORT` | 前端对外端口 | 默认 `80`；宿主机已有 Nginx 时可改成 `8088` |
| `BACKEND_BIND` | 后端监听地址 | 默认 `0.0.0.0`；若只通过前端转发访问，建议改成 `127.0.0.1` |
| `BACKEND_PORT` | 后端 HTTP 端口 | 默认 `8081`，主要用于健康检查和调试 |
| `MYSQL_BIND` | MySQL 暴露地址 | 默认 `127.0.0.1`，通常不要改成公网 |
| `REDIS_BIND` | Redis 暴露地址 | 默认 `127.0.0.1`，通常不要改成公网 |
| `MYSQL_ROOT_PASSWORD` | MySQL root 密码 | 必改 |
| `REDIS_PASSWORD` | Redis 密码 | 必改 |
| `JWT_SECRET` | JWT 签名密钥 | 必改，建议使用足够长的随机字符串 |
| `MYSQL_DATABASE` | 初始化数据库名 | 默认 `kafka_console` |
| `PROMETHEUS_BASE_URL` | Prometheus 地址 | 如果 Prometheus 部署在宿主机，默认 `http://host.docker.internal:9090` 可直接使用 |
| `INIT_DB` | 首次部署时是否自动导入 SQL | 默认 `true` |
| `FORCE_INIT_DB` | 是否强制重新导入 SQL | 默认 `false`，仅在你明确知道后果时再改 |

补充说明：
- `.env` 不存在时，第一次执行 `./deploy.sh up` 也会自动复制 `.env.example` 为 `.env`，但脚本会直接退出，提醒你先修改密码；所以更推荐手动先复制再编辑
- `docker-compose.yml` 已经为后端加了 `host.docker.internal:host-gateway`，因此 Linux 宿主机上的 Prometheus 默认可以通过 `host.docker.internal` 被容器访问
- 如果你的 Prometheus 不在当前宿主机，而是在其他机器，请把 `PROMETHEUS_BASE_URL` 改成实际可达地址

一个更适合“宿主机 Nginx + HTTPS 反代”的 `.env` 片段示例：

```env
FRONTEND_BIND=127.0.0.1
FRONTEND_PORT=8088
BACKEND_BIND=127.0.0.1
BACKEND_PORT=8081
MYSQL_BIND=127.0.0.1
REDIS_BIND=127.0.0.1
```

### 4. 首次部署

执行：

```bash
cd /opt/kafka-console
chmod +x deploy.sh
./deploy.sh up
```

这条命令会依次做这些事情：
1. 校验 `.env` 中的关键变量
2. 生成 `backend/config/config.yaml`
3. 创建数据目录 `data/mysql`、`data/redis`
4. 先启动 MySQL 和 Redis
5. 等待 MySQL 健康检查通过
6. 自动导入 `backend/sql/kafka_console_init.sql` 和 `backend/sql/kafka_module.sql`
7. 构建并启动后端和前端容器

部署完成后，默认访问地址是：
- 前端首页：`http://<服务器IP>:80`
- 前端健康检查：`http://<服务器IP>:80/health`
- 后端健康检查：`http://<服务器IP>:8081/health`

说明：
- 浏览器正常访问时，通常只需要打开前端地址
- 前端容器内部已经把 `/api/v1/` 和 `/ws/` 代理到后端，所以一般不需要把后端端口直接暴露给公网

### 5. 部署后检查

建议按下面顺序检查：

```bash
./deploy.sh status
./deploy.sh logs
./deploy.sh logs mysql
./deploy.sh logs redis
./deploy.sh logs devops-backend-svc
./deploy.sh logs frontend
```

也可以直接看容器状态：

```bash
docker compose ps
docker compose logs -f
```

如果只想做最基础的健康检查，可以执行：

```bash
curl http://127.0.0.1:${FRONTEND_PORT:-80}/health
curl http://127.0.0.1:${BACKEND_PORT:-8081}/health
```

### 6. 日常运维命令

```bash
./deploy.sh up         # 启动/更新全部服务
./deploy.sh down       # 停止并删除容器
./deploy.sh restart    # 重启服务
./deploy.sh status     # 查看状态
./deploy.sh logs       # 查看全部日志
./deploy.sh logs mysql
./deploy.sh logs redis
./deploy.sh logs devops-backend-svc
./deploy.sh logs frontend
./deploy.sh init-db    # 强制重新导入初始化 SQL
```

注意：
- `down` 不会删除项目目录下的 `data/mysql`、`data/redis` 持久化数据
- `init-db` 会强制重新导入初始化 SQL，执行前建议先备份数据库，避免覆盖或重复初始化带来风险

### 7. 更新部署

如果代码已经在服务器上，常见更新步骤如下：

```bash
cd /opt/kafka-console
# 更新代码，例如 git pull 或重新上传新版本文件
./deploy.sh up
```

如果新版本新增了环境变量，记得同步比对 `.env.example` 和 `.env`。

### 8. 常见问题

1. `./deploy.sh up` 一执行就退出，并提示已经创建 `.env`
   这是正常行为，说明脚本帮你生成了 `.env`，但要求你先修改密码和地址，再重新执行

2. 页面能打开，但监控面板没有数据
   先检查 `PROMETHEUS_BASE_URL` 是否可达；如果 Prometheus 不在宿主机，请改成实际地址

3. 后端配置改了又被覆盖
   `backend/config/config.yaml` 是 `deploy.sh` 自动生成的，请以 `.env` 为准，不要把它当作长期手工维护文件

4. 想把 MySQL 或 Redis 暴露到公网
   不建议这么做；默认绑定 `127.0.0.1` 更安全，外部如需访问建议走 SSH 隧道或内网

5. 已经部署过一次，不想重复导入初始化 SQL
   默认不会重复导入；脚本会检测关键表是否已存在，只有你显式使用 `./deploy.sh init-db` 或设置 `FORCE_INIT_DB=true` 时才会强制重新导入

## Windows 一键打包上传

- 使用 `upload-release.ps1`
- 仅上传并在服务器解压：
  `.\upload-release.ps1 -ServerHost 1.2.3.4 -ExtractOnServer`
- 上传后直接在服务器执行部署：
  `.\upload-release.ps1 -ServerHost 1.2.3.4 -ExtractOnServer -RunDeploy`

## Linux 预构建包部署

如果你在 Windows 上已经用 `package-release.ps1` 生成了预构建 tar 包，那么 Linux 服务器解压后，推荐直接使用：

```bash
chmod +x release.sh
./release.sh install
```

重新部署新包时：

```bash
# 先在旧包目录执行
./release.sh uninstall

# 再进入新包目录执行
./release.sh install
```

说明：
- 运行时 `.env`、MySQL、Redis 数据会保存在发布目录外的固定运行时目录
- 删除旧解压目录不会删掉运行时数据
- `release.sh install` 会先校验预构建资产是否完整，再调用 `deploy-prebuilt.sh`

## Nginx / 域名 HTTPS 反向代理

- 示例配置见 `deploy/nginx/kafka-console.https.conf.example`
- 如果宿主机上再套一层 Nginx，建议把 `.env` 改成：
  `FRONTEND_BIND=127.0.0.1`
  `FRONTEND_PORT=8088`
  `BACKEND_BIND=127.0.0.1`
- 然后让宿主机 Nginx 反代到 `127.0.0.1:8088`
- 前端容器内部已经处理了 `/api/v1/` 和 `/ws/` 到后端的转发，所以宿主机 Nginx 只需要代理前端即可
- 获取证书可用：
  `sudo certbot --nginx -d your-domain.example.com`

## Kubernetes / Helm 部署

- Helm Chart 目录：`deploy/k8s/helm/kafka-console`
- 主要文件：
  - `Chart.yaml`
  - `values.yaml`
  - `templates/deployment.yaml`
  - `templates/service.yaml`
  - `templates/ingress.yaml`
  - `templates/secret.yaml`
- 使用前请先修改 `values.yaml` 中的镜像地址、数据库、Redis、JWT、Prometheus 等配置
- 示例安装：
  `helm upgrade --install kafka-console deploy/k8s/helm/kafka-console -n kafka-console --create-namespace`
