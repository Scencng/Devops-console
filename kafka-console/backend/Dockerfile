FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/golang:1.25-alpine AS builder

WORKDIR /build
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o devops ./cmd/server


FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/alpine:3.20.2
ENV TZ=Asia/Shanghai
WORKDIR /app

COPY --from=builder /build/devops .
COPY config/ ./config/

RUN chmod +x ./devops
EXPOSE 8081

CMD ["./devops"]
