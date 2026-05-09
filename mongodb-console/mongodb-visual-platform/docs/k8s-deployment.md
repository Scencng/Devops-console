# k8s 部署配置说明

## 1. 目标

这个项目后续主要会部署到 k8s 中，用于：

- 小组展示
- 给同学们练手
- 演示 MongoDB 可视化连接、浏览和数据操作

因此配置方式必须从“只适合本地开发”调整为“可通过环境变量适配集群环境”。

## 2. 前端如何连接后端

前端通过环境变量 `VITE_API_BASE_URL` 指定后端地址。

### 本地开发

```env
VITE_API_BASE_URL=http://127.0.0.1:8080
```

### k8s 中常见写法

可以根据部署方式，配置成：

- Ingress 地址
- 网关地址
- 可直接访问的后端 svc 地址

例如：

```env
VITE_API_BASE_URL=http://mongodb-visual-backend.default.svc.cluster.local:8080
```

或者：

```env
VITE_API_BASE_URL=https://mongo.example.com/api
```

具体取决于你的前端是：

- 在集群内访问后端
- 还是通过 Ingress / 网关访问后端

## 3. 后端基础环境变量

后端仍通过环境变量启动：

- `MONGODB_URI`
- `MONGODB_DATABASE`
- `SERVER_PORT`
- `FRONTEND_ORIGINS`

示例：

```env
MONGODB_URI=mongodb://127.0.0.1:27017/admin
MONGODB_DATABASE=admin
SERVER_PORT=8080
FRONTEND_ORIGINS=http://localhost:5173,http://127.0.0.1:5173,https://mongo.example.com
```

说明：

- 这些默认值主要用于服务启动和兜底
- 真正的工作区连接，优先由前端连接页传入当前会话连接参数

## 4. 连接页的 host 为什么支持 svc 主机名

连接页中的 `host` 并不只支持 IP。

它可以填写：

- `127.0.0.1`
- 普通主机名
- 域名
- k8s Service 名
- 完整集群内 DNS

例如：

- `mongodb`
- `mongodb.default`
- `mongodb.default.svc.cluster.local`

这样做的好处是：

- 如果 MongoDB 部署在 k8s 中，就可以直接通过 CoreDNS 解析到对应 svc
- 同学们练手时，不需要先记住某个固定 IP
- 更符合 k8s 环境下服务访问的真实方式

## 5. CoreDNS 场景下的使用方式

如果前端运行环境能够解析集群内 DNS，那么连接页中可以直接填写：

```text
host: mongodb.default.svc.cluster.local
port: 27017
```

如果 MongoDB 开启认证，再补：

```text
username: <你的用户名>
password: <你的密码>
authSource: admin
```

## 6. 教学 / 展示环境的推荐部署思路

为了方便同学们练手，建议把环境分成三块：

1. 前端服务
2. 后端服务
3. MongoDB 服务

推荐思路：

- 前端通过 `VITE_API_BASE_URL` 指向后端
- 后端暴露 HTTP API
- MongoDB 通过 ClusterIP Service 暴露
- 同学们在连接页里直接填 MongoDB 的 svc 名

这样前端和后端仍然是统一入口，而 MongoDB 连接本身由可视化页面控制。

## 7. 一个新的 MongoDB 环境怎么准备

如果是一个新的练手环境，建议先做最基础初始化：

### 未开启认证

1. 部署 MongoDB svc
2. 直接通过连接页连接
3. 如无业务库，先创建首个 `database + collection`

### 已开启认证

1. 先通过 init 脚本或命令行创建 MongoDB 用户
2. 再通过连接页输入：
   - host
   - port
   - username
   - password
   - authSource
3. 如无业务库，再创建首个业务库

## 8. 文档和页面各自承担什么职责

在 k8s 展示环境里，建议职责划分如下：

- README / docs：
  - 说明怎么部署
  - 说明怎么初始化 MongoDB
  - 说明无认证和有认证的连接方式
- 页面：
  - 负责连接测试
  - 负责切换 MongoDB 连接
  - 负责后续 database / collection / document 可视化操作

这比在页面里假设某个默认管理员账号更可靠，也更适合教学。
