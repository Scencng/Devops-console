import axios from 'axios'
import { showError } from "@/utils/errorPopup.js";

// 创建axios实例
const api = axios.create({
  baseURL: '/api/v1', // 直接使用相对路径，由nginx代理
  timeout: 60000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('access_token')
    console.log(`[API Request] URL: ${config.url} | Token exists: ${!!token}`)

    if (token) {
    } else {
      console.log('No token found in localStorage for this request')
    }
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => {
    console.error('Request Interceptor Error:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    // 对响应数据做点什么
    const { data } = response

    if (data.status !== 200) {
      if (data.status === 401) {
        if (localStorage.getItem('access_token')) {
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
        }
        showError(data.message || '登录已过期，请重新登录', '登录已过期')
      }
      throw new Error(data.message || '请求失败')
    }
    return data
  },
  error => {
    // 对响应错误做点什么
    if (error.response && error.response.status === 401) {
      console.error('API 401 Log:', {
        url: error.config?.url,
        status: error.response.status,
        data: error.response.data,
        headers: error.response.headers
      })
      if (!window.location.pathname.includes('/login')) {
        if (localStorage.getItem('access_token')) {
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
        }
        showError(error.response.data.message || '登录已过期，请重新登录')
      }
    }
    const message = error.response?.data?.message || error.message || '网络错误'
    throw new Error(message)
  }
)

export default api
