# Kafka Console

这是一个在当前 `devops` 项目骨架基础上复制并持续演进的 Kafka 运维子项目，当前版本已经收敛为核心 Kafka 控制台。

## 最近更新

- 自动发现链路已支持按 `Cluster ID` 聚合显示
- 自动发现支持按集群导入与批量导入
- 自动发现结果已区分真实 `Broker` 节点与“访问入口”
- 已移除 `Prometheus 面板` 页面与对应菜单
- 页面布局已统一为更轻量的工作台风格
- 收起侧栏已切换为图标圆角按钮 + 自定义 tooltip / flyout 方案
- 后端编译链路已修复，打包流程改为“现场编译后端二进制再打包”

## 当前保留的核心功能

- Kafka 总览
- 集群管理
- 自动发现
- Topic 管理
- Broker 管理
- Consumer Group 管理
- 消息浏览与测试消息发送
- 审计日志

## 初始化 SQL

- `backend/sql/kafka_console_init.sql`
- `backend/sql/kafka_module.sql`

如果你使用仓库内置部署脚本：
- `./deploy.sh up`
- 或 `./release.sh install`

脚本会自动导入这两份 SQL。

## 目录说明

```text
kafka-console/
├── backend/         # Go 后端
├── frontend/        # Vue 3 前端
├── deploy/          # Nginx / Helm 示例
├── docker-compose.yml
├── docker-compose.prebuilt.yml
├── deploy.sh
├── deploy-prebuilt.sh
├── release.sh
└── README.md
```

## 三种使用模式

### 1. 本地开发模式

适用场景：
- 在 Windows / WSL 本地改代码、调页面、调接口
- 需要从源码直接构建前后端

主要文件：
- `docker-compose.yml`
- `deploy.sh`

使用方式：

```bash
cp .env.example .env
./deploy.sh up
```

说明：
- `deploy.sh` 会读取当前目录的 `.env`
- 会生成 `backend/config/config.yaml`
- 会自动初始化 MySQL / Redis 和核心 SQL
- 这是一套“源码构建并启动”的方式

适合：
- 日常开发
- WSL 本地验证
- 修改前后端代码后快速重启

### 2. 打包发布模式

适用场景：
- 在 Windows 开发机上生成一个可发到 Linux 服务器的发布包

主要文件：
- `package-release.ps1`

使用方式：

```powershell
.\package-release.ps1
```

说明：
- 会自动在本地构建前端 `dist`
- 会在 WSL 中现场编译 Linux 后端二进制
- 编译产物只进入临时打包目录，不会长期留在仓库里
- 最终生成一个 `kafka-console-prebuilt-*.tar.gz`

适合：
- 发版
- 给 Linux / WSL 服务器交付安装包

### 3. 服务器部署模式

适用场景：
- 你已经把预构建压缩包上传到 Linux / WSL 服务器
- 想直接安装、升级、卸载

主要文件：
- `release.sh`
- `deploy-prebuilt.sh`
- `docker-compose.prebuilt.yml`

推荐方式：

```bash
chmod +x release.sh
./release.sh install
```

升级新包：

```bash
# 旧包目录
./release.sh uninstall

# 新包目录
./release.sh install
```

说明：
- `release.sh` 是推荐入口
- 它会把 `.env`、MySQL 数据、Redis 数据放到发布目录外的固定运行时目录
- 删除旧解压目录时，不会把运行时数据一起删掉

`deploy-prebuilt.sh` 的定位：
- 它是一个“低层部署脚本”
- 默认要求当前目录已经具备预构建产物，例如：
  - `backend/devops`
  - `frontend/dist`
- 更适合被 `release.sh` 间接调用
- 如果你不想手工管理 `.env`、数据目录、运行时目录，优先用 `release.sh`

## 脚本职责

- `deploy.sh`
  - 源码模式入口
  - 负责读取 `.env`、生成后端配置、初始化数据库、启动源码构建版容器

- `package-release.ps1`
  - 打包入口
  - 负责构建前端、编译 Linux 后端，并生成预构建发布包

- `deploy-prebuilt.sh`
  - 预构建部署入口
  - 直接基于 `docker-compose.prebuilt.yml` 启动已经打包好的产物

- `release.sh`
  - 发布包安装/卸载入口
  - 在 `deploy-prebuilt.sh` 外面再包一层“运行时目录管理”
  - 适合服务器长期运维
