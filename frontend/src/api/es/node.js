import api from '../index.js'

// 获取节点信息
export const getNodeInfo = (instanceId) => {
  return api.get('/node/ClusterNodeInfo', {
    params: { instance_id: instanceId }
  })
}

// 获取节点统计信息
export const getNodeStats = (instanceId) => {
  return api.get('/node/ClusterNodeStats', {
    params: { instance_id: instanceId }
  })
}

// 获取所有节点 (mock 或其他正确实现的 API)
export const getNodes = (instanceId) => {
  return api.get('/node/ClusterNodeInfo', {
    params: { instance_id: instanceId }
  })
}

// 重启节点 (mock 或其他正确实现的 API)
export const restartNode = (instanceId, nodeId) => {
  return api.post('/node/RestartNode', {
    instance_id: instanceId,
    node_id: nodeId
  })
}