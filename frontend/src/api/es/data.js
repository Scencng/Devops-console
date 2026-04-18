import api from '../index.js'

// 通用ES请求函数，使用dev_console路由
const esRequest = (method, path, data, instanceId) => {
  const requestConfig = {
    method,
    url: `/elasticsearch/dev${path}`,
    params: {
      instance_id: instanceId
    }
  }

  if (data) {
    requestConfig.data = data
  }

  return api(requestConfig)
}

// 执行开发者控制台请求（JSON方式）
export const executeDevConsoleRequest = (requestData) => {
  return api({
    method: 'POST',
    url: '/elasticsearch/dev-console/execute',
    data: requestData
  })
}

// 获取索引列表
export const getIndices = (instanceId) => {
  return esRequest('GET', '/_cat/indices?format=json', null, instanceId)
}

// 获取索引详情
export const getIndexDetail = (instanceId, index) => {
  return esRequest('GET', `/${index}`, null, instanceId)
}

// 创建索引
export const createIndex = (instanceId, indexData) => {
  const { name, settings } = indexData
  return esRequest('PUT', `/${name}`, { settings }, instanceId)
}

// 搜索文档
export const searchDocuments = (instanceId, index, searchQuery) => {
  return esRequest('POST', `/${index}/_search`, searchQuery, instanceId)
}

// 获取单个文档
export const getDocument = (instanceId, index, docId) => {
  return esRequest('GET', `/${index}/_doc/${docId}`, null, instanceId)
}

// 兼容函数名
export const searchDocs = searchDocuments
export const getDoc = getDocument

// 创建文档
export const createDocument = (instanceId, index, document) => {
  return esRequest('POST', `/${index}/_doc`, document, instanceId)
}

// 更新文档
export const updateDocument = (instanceId, index, docId, document) => {
  return esRequest('POST', `/${index}/_doc/${docId}/_update`, document, instanceId)
}

// 删除文档
export const deleteDocument = (instanceId, index, docId) => {
  return esRequest('DELETE', `/${index}/_doc/${docId}`, null, instanceId)
}

// 批量操作
export const bulkOperation = (instanceId, operations) => {
  return esRequest('POST', '/_bulk', operations, instanceId)
}