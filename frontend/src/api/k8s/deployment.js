import request from '../index.js'

export const getDeploymentList = (namespace, instanceId) => {
  return request({
    url: `/k8s/deployment/list/${namespace}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

export const getDeploymentDetail = (namespace, deploymentName, instanceId) => {
  return request({
    url: `/k8s/deployment/detail/${namespace}/${deploymentName}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

export const createDeployment = (namespace, data, instanceId) => {
  return request({
    url: `/k8s/deployment/create/${namespace}`,
    method: 'post',
    data,
    params: {
      instance_id: instanceId
    }
  })
}

export const updateDeployment = (namespace, deploymentName, data, instanceId) => {
  return request({
    url: `/k8s/deployment/update/${namespace}/${deploymentName}`,
    method: 'put',
    data,
    params: {
      instance_id: instanceId
    }
  })
}

export const scaleDeployment = (namespace, deploymentName, replicas, instanceId) => {
  return request({
    url: `/k8s/deployment/scale/${namespace}/${deploymentName}/${replicas}`,
    method: 'put',
    params: {
      instance_id: instanceId
    }
  })
}

export const deleteDeployment = (namespace, deploymentName, instanceId) => {
  return request({
    url: `/k8s/deployment/delete/${namespace}/${deploymentName}`,
    method: 'delete',
    params: {
      instance_id: instanceId
    }
  })
}