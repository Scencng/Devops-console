import request from '../index.js'

// 获取Job列表
export const getJobList = (namespace, instanceId) => {
  return request({
    url: `/k8s/job/list/${namespace}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

// 获取Job详情
export const getJobDetail = (namespace, name, instanceId) => {
  return request({
    url: `/k8s/job/detail/${namespace}/${name}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

// 创建Job
export const createJob = (namespace, yamlContent, instanceId) => {
  return request({
    url: `/k8s/job/create/${namespace}`,
    method: 'post',
    data: { yaml: yamlContent },
    params: {
      instance_id: instanceId
    }
  })
}

// 删除Job
export const deleteJob = (namespace, name, instanceId) => {
  return request({
    url: `/k8s/job/delete/${namespace}/${name}`,
    method: 'delete',
    params: {
      instance_id: instanceId
    }
  })
}