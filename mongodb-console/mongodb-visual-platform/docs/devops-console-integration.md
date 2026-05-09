# MongoDB 模块并入 devops-console 的迁移说明

## 目标

将当前独立原型迁入 `devops-console` 仓库的 `mongoDB` 分支，作为运维平台中的 MongoDB 运维台模块。

## 当前现状

- 当前目录 `/workspace/mongoDB_visual` 不是 Git 仓库
- 当前实现已经具备单连接 MongoDB 的集合与文档 CRUD
- 前端基于 Vue 3，后端基于 Gin
- 当前环境无法通过 `git@github.com:xbh-ux/devops-console.git` 的 SSH 地址访问 GitHub 22 端口

## 推荐迁入方式

采用“迁入目标仓库子目录”的方式，而不是直接把当前目录当成最终仓库：

1. 获取 `devops-console` 本地工作副本
2. 切换到 `mongoDB` 分支
3. 对齐目标仓库的前端入口、后端边界、菜单注册和鉴权方式
4. 将当前 MongoDB 模块迁入目标目录
5. 使用目标仓库统一的登录、权限、日志、配置规范

## 建议映射

- 当前 `frontend/src/views`、`frontend/src/components`
  - 迁入目标仓库的 MongoDB 页面模块
- 当前 `frontend/src/api`、`frontend/src/stores`
  - 迁入目标仓库的前端 API 和状态层
- 当前 `backend/internal/http`、`backend/internal/service`、`backend/internal/mongodb`
  - 迁入目标仓库后端的 MongoDB 服务模块

## 并入前需要补齐的能力

- 登录态接入
- RBAC 权限控制
- 危险操作审计
- 多连接管理
- 主平台菜单、路由、导航风格对齐

## 当前阻塞

当前无法直接完成推送，原因如下：

- 当前工作目录不是目标仓库
- SSH 到 GitHub 22 端口不可达
- HTTPS 方式在无凭据条件下也无法读取该仓库

## 后续真正迁移时需要满足的条件

- 本地可访问 `devops-console` 仓库
- 拥有 `mongoDB` 分支读写权限
- 已明确目标仓库的目录结构和接入规范
