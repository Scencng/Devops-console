import api from '../index.js'

// 获取集群状态
export const getClusterStatus = (instanceId) => {
  return api.get('/elasticsearch/cluster/status', {
    params: {
      instance_id: instanceId
    }
  })
}

// 获取集群健康状态
export const getClusterHealth = (instanceId) => {
  return api.get('/elasticsearch/cluster/health', {
    params: {
      instance_id: instanceId
    }
  })
}

// 获取集群基本信息
export const getClusterInfo = (instanceId) => {
  return api.get('/elasticsearch/cluster/info', {
    params: {
      instance_id: instanceId
    }
  })
}

// 获取完整集群信息
export const getFullClusterInfo = (instanceId) => {
  return api.get('/elasticsearch/cluster/full-info', {
    params: {
      instance_id: instanceId
    }
  })
}

// 验证实例连接
export const validateInstance = (instanceId) => {
  return api.get('/elasticsearch/dev-console/validate', {
    params: {
      instance_id: instanceId
    }
  })
}

// 获取实例详细信息
export const getInstanceInfo = (instanceId) => {
  return api.get('/elasticsearch/dev-console/info', {
    params: {
      instance_id: instanceId
    }
  })
}

// 开发者控制台 - JSON请求体方式
export const executeDevConsoleRequest = (data) => {
  return api.post('/elasticsearch/dev-console/execute', data)
}

// 开发者控制台 - 查询参数方式
export const devConsoleRequest = (instanceId, method, path, data = null, params = {}) => {
  const requestConfig = {
    method,
    url: `/elasticsearch/dev${path}`,
    params: {
      instance_id: instanceId,
      ...params
    },
    ...(data && { data })
  }

  return api(requestConfig)
}