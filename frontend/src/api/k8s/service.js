import request from '../index.js'

// 获取Service列表
export const getServiceList = (namespace, instanceId) => {
  return request({
    url: `/k8s/service/list/${namespace}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

// 获取Service详情
export const getServiceDetail = (namespace, name, instanceId) => {
  return request({
    url: `/k8s/service/detail/${namespace}/${name}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

// 创建Service
export const createService = (namespace, name, data, instanceId) => {
  return request({
    url: `/k8s/service/create/${namespace}/${name}`,
    method: 'post',
    data,
    params: {
      instance_id: instanceId
    }
  })
}

// 更新Service
export const updateService = (namespace, name, data, instanceId) => {
  return request({
    url: `/k8s/service/update/${namespace}/${name}`,
    method: 'put',
    data,
    params: {
      instance_id: instanceId
    }
  })
}

// 删除Service
export const deleteService = (namespace, name, instanceId) => {
  return request({
    url: `/k8s/service/delete/${namespace}/${name}`,
    method: 'delete',
    params: {
      instance_id: instanceId
    }
  })
}

// 批量删除Service
export const deleteMultipleServices = (namespace, names, instanceId) => {
  return request({
    url: `/k8s/service/delete-multiple/${namespace}`,
    method: 'delete',
    data: { names },
    params: {
      instance_id: instanceId
    }
  })
}

// 根据标签删除Service
export const deleteServiceByLabel = (namespace, labels) => {
  return request({
    url: `/k8s/service/delete-by-label/${namespace}`,
    method: 'delete',
    data: { labels }
  })
}