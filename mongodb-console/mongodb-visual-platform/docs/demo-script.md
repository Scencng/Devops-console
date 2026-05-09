# 演示脚本

## 建议展示顺序

### 1. 项目介绍

先说明：

- 这是一个独立的 MongoDB 可视化平台
- 目标是降低 MongoDB 使用门槛
- 既能本地演示，也能部署到 k8s 给同学练手
- 支持连接、查询、导入导出、索引管理和备份恢复

### 2. 展示连接页与认证说明

- 先进入连接页
- 说明默认值：
  - `127.0.0.1`
  - `27017`
  - `authSource=admin`
- 说明为什么不预填 `root/admin`
- 说明 MongoDB 和 MySQL 认证方式不同
- 连接成功后进入工作区

### 3. 展示连接与结构浏览

- 展示当前连接状态
- 展示 database / collection 列表
- 说明系统 database 和系统 collection 默认隐藏
- 如果是空实例，说明为什么页面可能没有业务库

### 4. 展示 document 查询

- 进入某个 collection
- 先展示普通 document 列表
- 再展示条件查询：
  - 比如 `age > 25`
- 再展示 JSON 高级查询模式

### 5. 展示 document 编辑

- 新建 document
- 展示表单模式
- 展示 JSON 模式
- 编辑已有 document
- 删除 document

### 6. 展示导入导出与索引

- 导出当前查询结果为 JSON / CSV
- 导入一份 document 数据
- 展示当前 collection 的索引
- 展示索引创建或删除

### 7. 展示 collection 备份恢复

- 下载当前 collection 的 backup
- 从 backup 文件恢复 collection
- 说明 restore 会先清空再恢复

### 8. 展示语言切换

- 切换中文 / English
- 说明 MongoDB 术语保留英文

### 9. 如果老师问“空实例怎么初始化”

可以直接这样回答：

- 当前平台负责可视化连接和后续操作
- 对一个全新的 MongoDB 实例，通常先用命令行创建首个 `database + collection`
- 然后再回到页面继续练手
- 这样更贴近真实 MongoDB 使用方式

## 讲解重点

展示时建议强调：

- 不只是 CRUD
- 还支持连接不同 MongoDB 实例
- 还支持条件查询
- 还支持导入导出与索引管理
- 还支持 collection 级备份恢复
- 这是一个可独立运行、可部署到 k8s 的 MongoDB 可视化项目
