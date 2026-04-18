import request from '../index.js'

export const getNodeList = (instanceId) => {
  return request({
    url: '/k8s/node/list',
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

export const getNodeDetail = (nodeName, instanceId) => {
  return request({
    url: `/k8s/node/detail/${nodeName}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

export const cordonNode = (nodeName, instanceId) => {
  return request({
    url: `/k8s/node/${nodeName}/cordon`,
    method: 'post',
    params: {
      instance_id: instanceId
    }
  })
}

export const uncordonNode = (nodeName, instanceId) => {
  return request({
    url: `/k8s/node/${nodeName}/uncordon`,
    method: 'post',
    params: {
      instance_id: instanceId
    }
  })
}

export const drainNode = (nodeName, instanceId) => {
  return request({
    url: `/k8s/node/${nodeName}/drain`,
    method: 'post',
    params: {
      instance_id: instanceId
    }
  })
}

export const addNodeLabel = (nodeName, key, value, instanceId) => {
  return request({
    url: `/k8s/node/${nodeName}/labels`,
    method: 'post',
    data: {
      key,
      value
    },
    params: {
      instance_id: instanceId
    }
  })
}

export const removeNodeLabel = (nodeName, key, instanceId) => {
  return request({
    url: `/k8s/node/${nodeName}/labels/${key}`,
    method: 'delete',
    params: {
      instance_id: instanceId
    }
  })
}