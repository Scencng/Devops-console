import api from './index.js'

// 获取实例列表
export const getInstanceList = (params) => {
  return api.get('/instance/list', { params })
}

// 获取实例详情
export const getInstanceDetail = (id) => {
  return api.get('/instance/get', { params: { id } })
}

// 添加实例
export const addInstance = (data) => {
  return api.post('/instance/add', data)
}

// 更新实例
export const updateInstance = (data) => {
  return api.post('/instance/update', data)
}

// 删除实例
export const deleteInstance = (id) => {
  return api.get('/instance/delete', { params: { id } })
}

// 测试连接
export const testConnection = (instanceId) => {
  return api.post('/instance/test-connection', { instance_id: instanceId })
}

// 获取测试历史记录
export const getTestHistory = (instanceId, params = {}) => {
  return api.get('/instance/test-history', { 
    params: { 
      instance_id: instanceId,
      ...params 
    } 
  })
}

// 获取今日测试统计
export const getTodayTestStats = () => {
  return api.get('/instance/today-test-stats')
}

// 获取实例类型列表
export const getInstanceTypes = () => {
  return api.get('/instance/instance-types')
}