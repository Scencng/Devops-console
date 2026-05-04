# syntax=docker/dockerfile:1

# ==========================================
# 阶段 1: 编译 Vue 前端
# ==========================================
FROM node:20-alpine AS frontend-builder

WORKDIR /build/frontend

# 先复制依赖清单，尽量提高缓存命中
COPY frontend/package*.json ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

# 再复制源码并构建
COPY frontend/ ./
RUN npm run build:lean

# ==========================================
# 阶段 2: 编译 Go 后端
# ==========================================
FROM golang:1.25-alpine AS backend-builder

WORKDIR /build/backend

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# 先拉取依赖，保证 go build 阶段缓存更稳定
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 再复制源码并静态编译
COPY backend/ ./
RUN go build -trimpath -ldflags="-s -w" -o /build/mysql-visual-platform ./cmd/server

# ==========================================
# 阶段 3: 构建最终运行镜像
# ==========================================
FROM alpine:3.20

# 安装备份模块依赖和基础证书
RUN apk add --no-cache \
    ca-certificates \
    mysql-client \
    tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /opt/mysql-visual-platform

# 复制前端构建产物
COPY --from=frontend-builder /build/frontend/dist ./frontend/dist

# 复制后端二进制和配置
COPY --from=backend-builder /build/mysql-visual-platform ./backend/mysql-visual-platform
COPY --from=backend-builder /build/backend/config ./backend/config

# 预创建备份目录，便于宿主机或 K8s PVC 覆盖挂载
RUN mkdir -p ./backend/storage/backups

WORKDIR /opt/mysql-visual-platform/backend

EXPOSE 8080

# 启动命令与现有 Linux 文档保持一致
CMD ["./mysql-visual-platform", "--config", "./config/app.yaml", "--frontend-dist", "../frontend/dist"]
