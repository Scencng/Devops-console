# MySQL Visual Platform

轻量化跨平台 MySQL 可视化管理项目

当前目录结构固定为：

```text
mysql-visual-platform/
├─ frontend/
├─ backend/
├─ README.md
└─ .gitignore
```

## 项目说明

- 支持的功能：数据库连接、库表管理、SQL 查询、表格增删改查、智能导入导出、备份还原、结构对比、用户权限管理、中英文切换、响应式 UI
- 支持两种标准运行方式：
  1. `Windows + WSL Ubuntu`
  2. `通用 Linux`，适用于 Ubuntu、Kylin、CentOS、VMware 虚拟机内 Linux
- 项目已按轻量化方式清理：
  - 前端依赖需要在部署或开发前手动执行 `npm install`

## 环境要求

- Go `1.25+`
- Node.js `20+`
- npm `10+`
- MySQL `5.7 / 8.0`
- Linux 环境需可用 `mysql`、`mysqldump`

## 配置文件

后端配置文件：
- [backend/config/app.yaml](C:/Users/15202/project/mysql-project/mysql-visual-platform/backend/config/app.yaml)

前端运行时配置文件：
- [frontend/public/app-config.json](C:/Users/15202/project/mysql-project/mysql-visual-platform/frontend/public/app-config.json)

默认核心配置：

```yaml
server:
  address: 0.0.0.0:8080

frontend:
  dist_dir: ../frontend/dist
```

说明：
- `server.address` 控制服务监听地址
- `frontend.dist_dir` 为相对 `backend/` 的前端静态资源目录
- 前后端统一通过 `8080` 端口访问

## Windows + WSL Ubuntu

### 1. 安装前端依赖

```powershell
wsl -d Ubuntu -- bash -lc "cd /mnt/c/path/to/mysql-visual-platform/frontend && npm install"
```

### 2. 构建前端

```powershell
wsl -d Ubuntu -- bash -lc "cd /mnt/c/path/to/mysql-visual-platform/frontend && npm run build:lean"
```

### 3. 编译后端

```powershell
wsl -d Ubuntu -- bash -lc "export PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin && cd /mnt/c/path/to/mysql-visual-platform/backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o ./mysql-visual-platform ./cmd/server"
```

### 4. 前台启动

```powershell
wsl -d Ubuntu -- bash -lc "export PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin && cd /mnt/c/path/to/mysql-visual-platform/backend && ./mysql-visual-platform --config ./config/app.yaml --frontend-dist ../frontend/dist"
```

### 5. 后台启动

```powershell
Start-Process -FilePath wsl.exe -ArgumentList '-d','Ubuntu','--','bash','-lc','export PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin && cd /mnt/c/path/to/mysql-visual-platform/backend && exec ./mysql-visual-platform --config ./config/app.yaml --frontend-dist ../frontend/dist'
```

### 6. 停止服务

```powershell
wsl -d Ubuntu -- bash -lc "export PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin && /usr/bin/pkill -f '/mnt/c/path/to/mysql-visual-platform/backend/mysql-visual-platform' || true"
```

### 7. 验证监听

```powershell
wsl -d Ubuntu -- bash -lc "export PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin && /usr/bin/ss -ltnp | /usr/bin/grep 8080"
```

### 8. 访问项目

- `http://127.0.0.1:8080`

如果部分机器的 WSL 本地转发异常，可创建：
- `C:\Users\你的用户名\.wslconfig`

内容如下：

```ini
[wsl2]
localhostForwarding=true
```

修改后执行：

```powershell
wsl --shutdown
```

## 通用 Linux

适用环境：
- Ubuntu
- Kylin
- CentOS
- VMware 虚拟机内 Linux

### 1. 安装前端依赖

```bash
cd /path/to/mysql-visual-platform/frontend
npm install
```

### 2. 构建前端

```bash
cd /path/to/mysql-visual-platform/frontend
npm run build:lean
```

### 3. 编译后端

```bash
cd /path/to/mysql-visual-platform/backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o ./mysql-visual-platform ./cmd/server
```

### 4. 前台启动

```bash
cd /path/to/mysql-visual-platform/backend
./mysql-visual-platform --config ./config/app.yaml --frontend-dist ../frontend/dist
```

### 5. 后台启动

```bash
cd /path/to/mysql-visual-platform/backend
nohup ./mysql-visual-platform --config ./config/app.yaml --frontend-dist ../frontend/dist > ./server.log 2>&1 &
```

### 6. 停止服务

```bash
pkill -f '/path/to/mysql-visual-platform/backend/mysql-visual-platform'
```

### 7. 验证监听

```bash
ss -ltnp | grep 8080
```

### 8. 访问项目

- 本机访问：`http://127.0.0.1:8080`
- 局域网访问：`http://服务器IP:8080`
- VMware 虚拟机访问：`http://虚拟机IP:8080`

### 9. VMware 内同机 MySQL 管理

Linux 版本已完整覆盖 VMware 场景：
- Web 服务部署在 VMware 虚拟机内 Linux
- MySQL 也部署在同一台虚拟机内
- 浏览器通过虚拟机内网 IP 访问 Web
- 登录时填写虚拟机内 MySQL 地址即可正常管理

推荐填写方式：
- MySQL 主机：`127.0.0.1`
- 或虚拟机实际内网 IP，例如 `192.168.x.x`
- 端口一般为 `3306`

## 标准构建命令

### Windows + WSL Ubuntu

```powershell
wsl -d Ubuntu -- bash -lc "cd /mnt/c/path/to/mysql-visual-platform/frontend && npm install && npm run build:lean"
wsl -d Ubuntu -- bash -lc "export PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin && cd /mnt/c/path/to/mysql-visual-platform/backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o ./mysql-visual-platform ./cmd/server"
```

### 通用 Linux

```bash
cd /path/to/mysql-visual-platform/frontend
npm install
npm run build:lean

cd /path/to/mysql-visual-platform/backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o ./mysql-visual-platform ./cmd/server
```
