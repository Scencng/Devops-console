import api from '../index.js'

// 仓库管理
// 创建快照仓库
export const createRepository = (data) => {
  return api.post('/backup/repository', data)
}

// 删除快照仓库
export const deleteRepository = (params) => {
  return api.delete('/backup/repository', { params })
}

// 获取仓库列表
export const getRepositories = (instanceId) => {
  return api.get('/backup/repositories', { params: { instance_id: instanceId } })
}

// 快照管理
// 创建快照
export const createSnapshot = (data) => {
  return api.post('/backup/snapshot', data)
}

// 删除快照
export const deleteSnapshot = (params) => {
  return api.delete('/backup/snapshot', { params })
}

// 获取快照列表
export const getSnapshots = (instanceId, repository) => {
  return api.get('/backup/snapshots', { 
    params: { 
      instance_id: instanceId,
      repository: repository
    } 
  })
}

// 获取快照状态
export const getSnapshotStatus = (instanceId, repository, snapshot) => {
  return api.get('/backup/snapshot/status', { 
    params: { 
      instance_id: instanceId,
      repository: repository,
      snapshot: snapshot
    } 
  })
}

// 恢复管理
// 恢复快照
export const restoreSnapshot = (data) => {
  return api.post('/backup/restore', data)
}

// 获取恢复状态
export const getRestoreStatus = (instanceId) => {
  return api.get('/backup/restore/status', { 
    params: { instance_id: instanceId }
  })
}