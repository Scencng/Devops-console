import api from '../index.js'

// 获取集群健康状态
export const getClusterHealth = (instanceId) => {
  return api.post('/node/ClusterHealth', { instance_id: instanceId })
}

// 获取集群状态
export const getClusterState = (instanceId) => {
  return api.post('/node/ClusterStat', { instance_id: instanceId })
}

// 获取集群统计信息
export const getClusterStats = (instanceId) => {
  return api.post('/node/ClusterStats', { instance_id: instanceId })
}

// 获取节点统计信息
export const getNodesStats = (instanceId) => {
  return api.post('/node/NodeStats', { instance_id: instanceId })
}