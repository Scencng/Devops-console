import request from './index.js'

export const getChaosExperiments = (namespace, instanceId) => {
  return request({
    url: `/k8s/chaos/list/${namespace}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

export const createChaosExperiment = (namespace, data, instanceId) => {
  return request({
    url: `/k8s/chaos/create/${namespace}`,
    method: 'post',
    data,
    params: {
      instance_id: instanceId
    }
  })
}

export const getChaosExperiment = (namespace, name, instanceId) => {
  return request({
    url: `/k8s/chaos/get/${namespace}/${name}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

export const deleteChaosExperiment = (namespace, name, instanceId) => {
  return request({
    url: `/k8s/chaos/delete/${namespace}/${name}`,
    method: 'delete',
    params: {
      instance_id: instanceId
    }
  })
}

export const pauseChaosExperiment = (namespace, name, instanceId) => {
  return request({
    url: `/k8s/chaos/pause/${namespace}/${name}`,
    method: 'put',
    params: {
      instance_id: instanceId
    }
  })
}

export const resumeChaosExperiment = (namespace, name, instanceId) => {
  return request({
    url: `/k8s/chaos/resume/${namespace}/${name}`,
    method: 'put',
    params: {
      instance_id: instanceId
    }
  })
}

// ——— 演练节点驱逐 ———

export const getChaosTestingNodes = (instanceId) => {
  return request({
    url: '/k8s/chaos/nodes',
    method: 'get',
    params: { instance_id: instanceId }
  })
}

export const prepareEviction = (data, instanceId) => {
  return request({
    url: '/k8s/chaos/evict/prepare',
    method: 'post',
    data,
    params: { instance_id: instanceId }
  })
}

export const cleanupEviction = (data, instanceId) => {
  return request({
    url: '/k8s/chaos/evict/cleanup',
    method: 'post',
    data,
    params: { instance_id: instanceId }
  })
}
