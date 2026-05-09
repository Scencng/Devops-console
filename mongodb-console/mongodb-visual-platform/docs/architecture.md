# 架构说明

## 1. 总体架构

项目采用前后端分离：

- 前端负责界面展示、连接页和会话级连接状态
- 后端负责 MongoDB 连接解析、查询、导入导出和备份恢复

## 2. 前端结构

- `frontend/src/views`
  - 页面视图（连接页、工作区、文档页）
- `frontend/src/api`
  - HTTP 请求封装
- `frontend/src/session`
  - 当前浏览器会话里的 MongoDB 连接配置
- `frontend/src/i18n`
  - 语言切换资源
- `frontend/src/styles`
  - 页面样式

## 3. 后端结构

- `backend/cmd/server`
  - 启动入口
- `backend/internal/config`
  - 配置加载
- `backend/internal/http`
  - 路由、请求头解析与处理器
- `backend/internal/service`
  - 业务逻辑
- `backend/internal/mongodb`
  - MongoDB 连接池、查询构建、导入导出与备份恢复工具
- `backend/internal/model`
  - 请求响应结构

## 4. 数据流

1. 用户先进入连接页，输入 MongoDB 连接参数
2. 前端调用连接测试接口验证是否可达
3. 连接成功后，前端把连接配置保存到 `sessionStorage`
4. 工作区请求会自动带上 `X-Mongo-*` 连接头
5. 后端根据请求头解析当前连接，并复用对应的 Mongo client
6. 用户展开 database 后，前端请求 collection 列表
7. 用户选中 collection 后，请求 document 列表
8. 用户通过表单或 JSON 模式创建 / 编辑 document
9. 后端把前端请求转换成 MongoDB 操作
10. 后端返回结果给前端展示

## 5. 连接认证方式

当前项目支持两种 MongoDB 场景：

- 无认证实例
- 用户名 / 密码 + `authSource` 的认证实例

前端连接参数包括：

- `host`
- `port`
- `database`
- `username`
- `password`
- `authSource`

后端会从请求头读取这些参数：

- `X-Mongo-Host`
- `X-Mongo-Port`
- `X-Mongo-Database`
- `X-Mongo-Username`
- `X-Mongo-Password`
- `X-Mongo-AuthSource`

## 6. 连接恢复

后端 MongoDB 客户端在每次关键请求前都会检查连接状态：

- 如果连接正常，继续使用当前 client
- 如果连接失效，自动重建连接

这样在 MongoDB 服务恢复后，不需要手动重启后端。

## 7. k8s 部署相关设计点

为了适配 k8s 展示和练手环境：

- 前端通过 `VITE_API_BASE_URL` 指向后端
- 连接页中的 `host` 支持：
  - IP
  - 普通主机名
  - k8s svc 名
  - CoreDNS 完整域名
- 后端默认环境变量 MongoDB 连接只作为兜底
- 真正工作区连接优先由当前浏览器会话指定
