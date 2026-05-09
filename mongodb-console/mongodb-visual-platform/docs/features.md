# 功能说明

## 1. 连接页与 MongoDB 接入

- 支持独立连接页
- 支持测试 MongoDB 连接
- 支持按浏览器会话切换当前 MongoDB 目标
- 支持无认证实例
- 支持用户名 / 密码 + `authSource` 认证
- `host` 可以填写：
  - IP
  - 主机名
  - k8s svc 名
  - 完整集群 DNS

## 2. database 与 collection 浏览

- 左侧展示 database 列表
- 展开后可查看 collection
- 默认隐藏系统 database：
  - `admin`
  - `config`
  - `local`
- 默认隐藏系统 collection
- 展开后按需显示 document 数量

## 3. document 管理

- 查看 document 列表
- 按 `_id` 查看 document
- 新增 document
- 编辑 document
- 删除 document

## 4. 查询功能

支持两种查询模式：

- 条件构建器
  - 如 `age > 25`
  - 支持 `AND/OR`
  - 支持 `= != > >= < <= contains in`
- JSON 高级模式
  - 用于更复杂的 MongoDB 查询

## 5. 创建与编辑方式

支持两种编辑模式：

- 表单模式
  - 按字段名、类型、值逐行编辑
- JSON 模式
  - 用于复杂结构或高级 BSON 表达

## 6. 导入导出

document 级：

- 导出 JSON
- 导出 CSV
- 导入 JSON
- 导入 CSV

## 7. collection 级备份恢复

- 可对整个 collection 做 backup
- 可从备份文件 restore 到当前 collection
- restore 时会先清空当前 collection，再恢复备份内容

## 8. 索引管理

- 查看当前 collection 的索引
- 创建索引
- 删除非默认索引
- 显示索引名称、索引键和常见属性

## 9. 语言切换

- 支持中文 / English 切换
- MongoDB 专有术语保留英文，如：
  - `MongoDB`
  - `document`
  - `collection`

## 10. 当前关于 database 管理的边界

当前版本已经支持：

- 浏览所有可见业务 database

当前版本还没有把以下能力做成稳定的 UI 功能：

- 创建 database
- 删除 database

原因是 MongoDB 的 database 通常通过“创建首个 collection”自然出现，而不是像关系型数据库那样强依赖空库管理。
