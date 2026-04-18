import request from '../index.js'

// 获取CronJob列表
export const getCronJobList = (namespace, instanceId) => {
  return request({
    url: `/k8s/cronjob/list/${namespace}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

// 创建CronJob
export const createCronJob = (namespace, yamlContent, instanceId) => {
  return request({
    url: `/k8s/cronjob/create/${namespace}`,
    method: 'post',
    data: { yaml: yamlContent },
    params: {
      instance_id: instanceId
    }
  })
}

// 更新CronJob
export const updateCronJob = (data, instanceId) => {
  return request({
    url: '/k8s/cronjob/update',
    method: 'put',
    data,
    params: {
      instance_id: instanceId
    }
  })
}

// 删除CronJob
export const deleteCronJob = (namespace, name, instanceId) => {
  return request({
    url: '/k8s/cronjob/delete',
    method: 'delete',
    data: { namespace, name },
    params: {
      instance_id: instanceId
    }
  })
}

// 获取CronJob详情
export const getCronJobDetail = (namespace, cronJobName, instanceId) => {
  return request({
    url: `/k8s/cronjob/detail/${namespace}/${cronJobName}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

// 获取CronJob YAML
export const getCronJobYAML = (namespace, cronJobName, instanceId) => {
  return request({
    url: `/k8s/cronjob/yaml/${namespace}/${cronJobName}`,
    method: 'get',
    params: {
      instance_id: instanceId
    }
  })
}

// 更新CronJob (YAML)
export const updateCronJobYAML = (namespace, yamlContent, instanceId) => {
  return request({
    url: `/k8s/cronjob/update/yaml/${namespace}`,
    method: 'put',
    data: { yaml: yamlContent },
    params: {
      instance_id: instanceId
    }
  })
}