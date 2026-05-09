# MongoDB 认证与首次连接

## 1. MongoDB 认证和 MySQL 的区别

如果你更熟悉 MySQL，最容易产生的误解是：

- 以为 MongoDB 也会默认带一个类似 `root@localhost` 的账号
- 以为连接方式也像 MySQL 那样先记住固定用户名，再输密码

MongoDB 不是这样。

### MySQL 的常见认知

- 常见默认端口：`3306`
- 常见管理员用户名：`root`
- 连接时通常明确区分：
  - 主机
  - 端口
  - 用户
  - 密码

### MongoDB 的常见认知

- 默认端口：`27017`
- 常见认证参数：
  - `host`
  - `port`
  - `database`
  - `username`
  - `password`
  - `authSource`
- `authSource` 表示“去哪个数据库验证这个用户”
- `admin` 是常见默认认证库，但**不是默认用户**

## 2. 为什么不预填 root/admin

本项目连接页默认：

- `host = 127.0.0.1`
- `port = 27017`
- `authSource = admin`

但不会默认预填：

- `root`
- `admin`

原因是：

1. MongoDB 并不保证默认存在 `root` 或 `admin` 用户
2. 很多本地 MongoDB 实例根本没开启认证
3. 教学场景中预填一个并不存在的用户名，反而会误导同学以为“MongoDB 自带超级管理员账号”

所以更合理的做法是：

- 默认把用户名留空
- 明确告诉用户“如果实例开启了认证，请填写你自己创建的 MongoDB 用户”

## 3. 无认证实例怎么连接

如果你的 MongoDB 没有开启认证：

- `username` 留空
- `password` 留空
- `authSource` 保持 `admin`

这通常适用于：

- 本地练手环境
- 临时测试环境
- 课程展示环境

## 4. 有认证实例怎么连接

如果 MongoDB 开启了认证，连接时需要知道三件事：

1. 用户名
2. 密码
3. 这个用户是在哪个数据库里创建的

连接页填写方式：

- `host`：MongoDB 所在主机名、IP 或 svc 名
- `port`：通常为 `27017`
- `database`：你要浏览的目标业务库
- `username`：MongoDB 用户名
- `password`：对应密码
- `authSource`：创建该用户时所在的数据库

很多管理员用户会创建在 `admin` 下，因此 `authSource=admin` 很常见，但不是强制规则。

## 5. 首次连接一个全新实例时会看到什么

如果是一个新的 MongoDB 实例，常见情况有两种：

### 情况 A：未开启认证

- 可以直接通过连接页连上
- 但如果实例里没有业务库，页面可能只剩空状态
- 这是因为系统库 `admin`、`config`、`local` 会被默认隐藏

### 情况 B：开启了认证

- 如果没有可用用户，就无法连接成功
- 这时需要先在命令行或初始化脚本里创建 MongoDB 用户

## 6. 没有业务 database 时怎么办

MongoDB 里通常不是“先有空库，再建表”，而是：

- 当你在某个 database 下创建第一个 `collection`
- 或者写入第一条数据

这个 database 才真正出现。

因此对于一个全新实例，推荐初始化方式是：

1. 先通过命令行创建首个 `database + collection`
2. 再回到可视化平台进行 collection / document 级管理

## 7. 推荐的首次初始化命令思路

你可以在 README 或教学环境里准备类似这样的初始化方式：

### 无认证实例

```bash
mongosh --host 127.0.0.1 --port 27017
use demo_db
db.createCollection("students")
```

### 有认证实例

```bash
mongosh --host 127.0.0.1 --port 27017 -u <username> -p <password> --authenticationDatabase admin
use demo_db
db.createCollection("students")
```

完成这一步后，再刷新可视化平台，`demo_db` 就会出现在左侧 database 列表里。

## 8. 这个项目当前对首次连接的建议

对于“展示 + 给同学练手”的场景，最合理的策略不是假设系统已经有默认管理员账号，而是：

- 连接页负责连接与认证
- README / docs 负责说明首次初始化
- 页面负责后续可视化浏览与操作

这样更贴近真实 MongoDB 使用方式，也更利于教学。
