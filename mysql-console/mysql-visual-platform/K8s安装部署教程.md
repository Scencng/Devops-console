# 1. 编写Dockerfile文件

```dockerfile
# 阶段 1: 编译 Vue 前端
FROM node:20-alpine AS frontend-builder

WORKDIR /build/frontend

# 先复制依赖清单，尽量提高缓存命中
COPY frontend/package*.json ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

# 再复制源码并构建
COPY frontend/ ./
RUN npm run build:lean


# 阶段 2: 编译 Go 后端
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


# 阶段 3: 构建最终运行镜像
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
```

# 2. 构建并推送镜像

```bash
docker build -t harbor250.oldboyedu.com/oldboyedu-devops/mysql-visual-platform:v1.0 .

docker push harbor250.oldboyedu.com/oldboyedu-devops/mysql-visual-platform:v1.0
```

# 3. K8s 资源清单

- [mysql-visual-platform-k8s.yaml](C:/Users/15202/project/mysql-project/mysql-visual-platform/mysql-visual-platform-k8s.yaml)

内容如下：

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: mysql-visual-platform
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: mysql-visual-platform-config
  namespace: mysql-visual-platform
data:
  app.yaml: |
    server:
      address: 0.0.0.0:8080

    frontend:
      dist_dir: ../frontend/dist
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-visual-platform-backups-pvc
  namespace: mysql-visual-platform
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: nfs-csi
  resources:
    requests:
      storage: 5Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql-visual-platform
  namespace: mysql-visual-platform
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mysql-visual-platform
  template:
    metadata:
      labels:
        app: mysql-visual-platform
    spec:
      terminationGracePeriodSeconds: 30
      imagePullSecrets:
        - name: harbor-secret
      containers:
        - name: mysql-visual-platform
          image: harbor250.oldboyedu.com/oldboyedu-devops/mysql-visual-platform:v1.0
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
          env:
            - name: TZ
              value: Asia/Shanghai
          startupProbe:
            tcpSocket:
              port: 8080
            failureThreshold: 30
            periodSeconds: 5
          readinessProbe:
            tcpSocket:
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            tcpSocket:
              port: 8080
            initialDelaySeconds: 20
            periodSeconds: 20
          volumeMounts:
            - name: app-config
              mountPath: /app/backend/config/app.yaml
              subPath: app.yaml
            - name: backups-storage
              mountPath: /app/backend/storage/backups
          resources:
            requests:
              cpu: 100m
              memory: 256Mi
            limits:
              cpu: 1000m
              memory: 1Gi
      volumes:
        - name: app-config
          configMap:
            name: mysql-visual-platform-config
        - name: backups-storage
          persistentVolumeClaim:
            claimName: mysql-visual-platform-backups-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: mysql-visual-platform
  namespace: mysql-visual-platform
spec:
  type: NodePort
  selector:
    app: mysql-visual-platform
  ports:
    - name: http
      port: 80
      targetPort: 8080
      nodePort: 30080
```

# 4. 部署命令

```bash
kubectl apply -f mysql-visual-platform-k8s.yaml
```

# 5. 查看部署状态

## 5.1 查看 Pod

```bash
kubectl get pods -n mysql-visual-platform
```

## 5.2 查看 Service

```bash
kubectl get svc -n mysql-visual-platform
```

## 5.3 查看日志

```bash
kubectl logs -f deploy/mysql-visual-platform -n mysql-visual-platform
```

## 5.4 查看详细事件

```bash
kubectl describe pod -n mysql-visual-platform
```

# 6. 浏览器访问

```text
http://10.0.0.231:30080
```
