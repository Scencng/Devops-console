import api from '../index.js'

// 获取分片信息
export const getShardInfo = (instanceId) => {
  return api.get('/shard/info', { params: { instance_id: instanceId } })
}

// 获取分片统计信息
export const getShardStats = (instanceId) => {
  return api.get('/shard/stats', { params: { instance_id: instanceId } })
}

// 分配分片
export const allocateShard = (data) => {
  return api.post('/shard/allocate', data)
}

// 移动分片
export const moveShard = (data) => {
  return api.post('/shard/move', data)
}

// 迁移分片到另一个实例
export const migrateShardToInstance = (data) => {
  return api.post('/shard/migrate-to-instance', data)
}