import axios from 'axios'
import { showError } from "@/utils/errorPopup.js";
import { useUiStore } from '@/stores/uiStore.js'

// 创建axios实例
const api = axios.create({
  baseURL: '/api/v1', // 直接使用相对路径，由nginx代理
  timeout: 300000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    const uiStore = useUiStore()
    uiStore.incrementRequests()
    const token = localStorage.getItem('access_token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => {
    const uiStore = useUiStore()
    uiStore.decrementRequests()
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    const uiStore = useUiStore()
    uiStore.decrementRequests()
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
    const uiStore = useUiStore()
    uiStore.decrementRequests()
    // 对响应错误做点什么
    if (error.response && error.response.status === 401) {
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
