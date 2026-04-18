import request from '../index.js'

// 获取集群基本信息
export const getClusterInfo = (instanceId) => {
  return request({
    url: '/cluster/info',
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

// 获取集群指标数据
export const getClusterMetrics = (instanceId) => {
  return request({
    url: '/cluster/metrics',
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

// 获取节点列表
export const getNodeList = (instanceId) => {
  return request({
    url: '/cluster/nodes',
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

// 节点隔离
export const cordonNode = (nodeName) => {
  return request({
    url: `/cluster/nodes/${nodeName}/cordon`,
    method: 'post'
  })
}

// 取消节点隔离
export const uncordonNode = (nodeName) => {
  return request({
    url: `/cluster/nodes/${nodeName}/uncordon`,
    method: 'post'
  })
}

// 节点排空
export const drainNode = (nodeName) => {
  return request({
    url: `/cluster/nodes/${nodeName}/drain`,
    method: 'post'
  })
}

// 添加节点标签
export const addNodeLabel = (nodeName, key, value) => {
  return request({
    url: `/cluster/nodes/${nodeName}/labels`,
    method: 'post',
    data: {
      key,
      value
    }
  })
}

// 删除节点标签
export const removeNodeLabel = (nodeName, key) => {
  return request({
    url: `/cluster/nodes/${nodeName}/labels/${key}`,
    method: 'delete'
  })
}