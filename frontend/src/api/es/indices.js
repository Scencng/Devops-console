import api from '../index.js'
import axios from 'axios'

// Base64编码工具函数
const base64 = {
  encode: (str) => {
    return btoa(unescape(encodeURIComponent(str)))
  }
}

// 获取索引列表
export const getIndices = (instanceId) => {
  return api.get('/indices/catindices', {
    params: { instance_id: instanceId }
  })
}

// 创建索引
export const createIndex = (indexData) => {
  return api.post('/indices/indexcreate', indexData)
}

// 删除索引
export const deleteIndex = (indexData) => {
  return api.post('/indices/deleteindices', indexData)
}

// 更新索引设置
export const updateIndex = (indexData) => {
  return api.post('/indices/updateindices', indexData)
}

// 通用ES请求函数，使用dev_console路由
const esRequest = async (method, path, data, instanceId) => {
  // 获取实例详情以获得ES连接配置
  try {
    const instanceResponse = await api.get('/instance/get', { 
      params: { id: instanceId } 
    })
    
    const instance = instanceResponse.data.instance
    if (!instance) {
      throw new Error('实例不存在')
    }

    // 构建ES连接URL
    const protocol = instance.https_enabled ? 'https' : 'http'
    const host = `${protocol}://${instance.address}`
    
    // 解析认证配置
    let username = ''
    let password = ''
    if (instance.auth_configs) {
      try {
        const authMatch = instance.auth_configs.match(/^basic:\{(.+)\}$/)
        if (authMatch) {
          const authData = JSON.parse(`{${authMatch[1]}}`)
          username = authData.username || ''
          password = authData.password || ''
        }
      } catch (e) {
        console.warn('解析认证配置失败:', e)
      }
    }

    // 将用户名和密码编码为Base64
    const authString = username && password ? base64.encode(`${username}:${password}`) : ''
    
    const requestConfig = {
      method,
      url: `/elasticsearch/dev${path}`,
      headers: {
        'X-ES-Host': host,
        'X-ES-Password': authString,
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      }
    }

    if (data) {
      requestConfig.data = data
    }

    // 使用原生axios，绕过拦截器
    const axiosInstance = axios.create({
      baseURL: 'http://localhost:8080/api',
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json'
      }
    })
    
    const response = await axiosInstance(requestConfig)
    
    // 返回标准格式，但data字段包含ES原始响应
    return {
      status: 200,
      message: '请求成功',
      data: {
        raw_response: typeof response.data === 'string' ? JSON.parse(response.data) : response.data
      }
    }
  } catch (error) {
    throw new Error(`获取实例配置失败: ${error.message}`)
  }
}

// 获取索引详情
export const getIndexDetail = (indexName, instanceId) => {
  return esRequest('GET', `/${indexName}`, null, instanceId)
}

// 获取索引映射
export const getIndexMapping = (indexName, instanceId) => {
  return esRequest('GET', `/${indexName}/_mapping`, null, instanceId)
}

// 获取索引设置
export const getIndexSettings = (indexName, instanceId) => {
  return esRequest('GET', `/${indexName}/_settings`, null, instanceId)
}

// 优化索引
export const optimizeIndex = (indexName, instanceId) => {
  return esRequest('POST', `/${indexName}/_forcemerge`, {}, instanceId)
}

// 打开索引
export const openIndex = (indexName, instanceId) => {
  return esRequest('POST', `/${indexName}/_open`, {}, instanceId)
}

// 关闭索引
export const closeIndex = (indexName, instanceId) => {
  return esRequest('POST', `/${indexName}/_close`, {}, instanceId)
}