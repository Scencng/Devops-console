import request from '../index.js'

// 获取DaemonSet列表
export const getDaemonSetList = (namespace, instanceId) => {
  return request({
    url: `/k8s/daemonset/list/${namespace}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

// 获取DaemonSet详情
export const getDaemonSetDetail = (namespace, name, instanceId) => {
  return request({
    url: `/k8s/daemonset/detail/${namespace}/${name}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

// 创建DaemonSet
export const createDaemonSet = (namespace, data, instanceId) => {
  return request({
    url: `/k8s/daemonset/create/${namespace}`,
    method: 'post',
    data,
    params: {
      instance_id: instanceId
    }
  })
}

// 更新DaemonSet
export const updateDaemonSet = (namespace, name, data, instanceId) => {
  return request({
    url: `/k8s/daemonset/update/${namespace}/${name}`,
    method: 'put',
    data,
    params: {
      instance_id: instanceId
    }
  })
}

// 删除DaemonSet
export const deleteDaemonSet = (namespace, name, instanceId) => {
  return request({
    url: `/k8s/daemonset/delete/${namespace}/${name}`,
    method: 'delete',
    params: {
      instance_id: instanceId
    }
  })
}