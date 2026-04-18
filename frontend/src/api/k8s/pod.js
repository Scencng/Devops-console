import request from '../index.js'

export const getPodList = (namespace, instanceId) => {
  const url = namespace === 'all' ? '/k8s/pod/list/all' : `/k8s/pod/list/${namespace}`
  return request({
    url: url,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

export const getPodDetail = (namespace, podName, instanceId) => {
  return request({
    url: `/k8s/pod/detail/${namespace}/${podName}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

export const getPodEvents = (namespace, podName, instanceId) => {
  return request({
    url: `/k8s/pod/events/${namespace}/${podName}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

export const getPodLogs = (namespace, podName, container, instanceId, options = {}) => {
  const params = {
    instance_id: instanceId,
    container: container,
    follow: options.follow || false,
    tail_lines: options.tailLines || 100,
    timestamps: options.timestamps || false
  }
  
  return request({
    url: `/k8s/pod/logs/${namespace}/${podName}`,
    method: 'get',
    params: params
  })
}

export const createPod = (data, instanceId) => {
  return request({
    url: '/k8s/pod/create',
    method: 'post',
    data,
    params: {
      instance_id: instanceId
    }
  })
}

export const updatePod = (data, instanceId) => {
  return request({
    url: '/k8s/pod/update',
    method: 'put',
    data,
    params: {
      instance_id: instanceId
    }
  })
}

export const deletePod = (namespace, podName, instanceId) => {
  return request({
    url: `/k8s/pod/delete/${namespace}/${podName}`,
    method: 'delete',
    params: {
      instance_id: instanceId
    }
  })
}