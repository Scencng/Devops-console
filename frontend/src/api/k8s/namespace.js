import request from '../index.js'

export const getNamespaceList = (instanceId) => {
  return request({
    url: '/k8s/namespace/list',
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

export const createNamespace = (namespace, data, instanceId) => {
  return request({
    url: `/k8s/namespace/create/${namespace}`,
    method: 'post',
    data,
    params: {
      instance_id: instanceId
    }
  })
}

export const deleteNamespace = (namespaceName, instanceId) => {
  return request({
    url: `/k8s/namespace/delete/${namespaceName}`,
    method: 'delete',
    params: {
      instance_id: instanceId
    }
  })
}