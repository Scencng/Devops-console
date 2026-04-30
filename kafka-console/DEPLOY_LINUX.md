# Linux 手动部署步骤

## 前提

服务器已安装：

```bash
docker --version
docker compose version
```

## 1. 上传并解压

```bash
cd /opt
tar -xJf kafka-console-prebuilt-*.tar.xz
cd kafka-console
```

## 2. 准备环境变量

```bash
cp .env.example .env
vi .env
```

至少修改这 3 项：

```env
MYSQL_ROOT_PASSWORD=改成强密码
REDIS_PASSWORD=改成强密码
JWT_SECRET=改成长随机串
```

如果你不想使用默认端口，也可以修改：

```env
FRONTEND_PORT=80
BACKEND_PORT=8081
```

## 3. 启动

```bash
docker compose up -d --build
```

说明：

- 后端直接从环境变量读取配置
- MySQL 首次启动会自动导入 SQL
- 数据使用 Docker 命名卷持久化
- MySQL / Redis 默认不暴露宿主机端口，避免端口冲突
- 这里的 `--build` 只是把包里已经带好的前端静态资源和 Linux 后端二进制装进镜像，不会重新编译 Go 源码

## 4. 查看状态

```bash
docker compose ps
docker logs -f kafka-console-backend
docker logs -f kafka-console-frontend
```

## 5. 访问

默认访问地址：

- 前端：`http://<server-ip>:8088`
- 后端健康检查：`http://<server-ip>:8081/health`

默认账号：

- 用户名：`admin`
- 密码：`admin123`

首次登录后必须立即修改密码。

## 6. 升级

新包上传后：

```bash
cd /opt
mv kafka-console kafka-console.bak
tar -xJf 新包.tar.xz
cd kafka-console
cp ../kafka-console.bak/.env .env
docker compose up -d --build
```

旧目录确认不再需要后可以删除：

```bash
rm -rf /opt/kafka-console.bak
```

## 7. 卸载

停止服务：

```bash
docker compose down
```

如果连数据一起删：

```bash
docker compose down -v
```
