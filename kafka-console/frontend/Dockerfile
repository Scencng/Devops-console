FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/node:20.19-alpine AS builder

WORKDIR /app

COPY . .

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
	npm config set registry https://registry.npmjs.org && \
	npm config set fetch-retries 5 && \
	npm config set fetch-retry-mintimeout 20000 && \
	npm config set fetch-retry-maxtimeout 120000 && \
	npm ci --registry https://registry.npmjs.org && \
    chmod -R +x node_modules/.bin && \
    ./node_modules/.bin/vite build

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/nginx:1.27.4-alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
	apk add --no-cache wget

COPY ./nginx.conf /etc/nginx/nginx.conf

COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 8081
EXPOSE 9200

CMD ["nginx","-g","daemon off;"]
