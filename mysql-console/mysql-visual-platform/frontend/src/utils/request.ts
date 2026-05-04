import axios, {
  AxiosError,
  type AxiosRequestConfig,
  type AxiosResponse,
  type InternalAxiosRequestConfig
} from 'axios'
import { ElMessage } from 'element-plus'

import { getRuntimeConfig } from '@/config/runtime'
import pinia from '@/stores'
import { useConnectionStore } from '@/stores/connection'

interface ApiResponse<T> {
  code: number
  msg: string
  data: T
}

interface ConnectionOpenResponse {
  connectionToken: string
}

interface RetriableRequestConfig extends AxiosRequestConfig {
  _connectionRetried?: boolean
  skipConnectionToken?: boolean
  silentError?: boolean
}

const instance = axios.create({
  baseURL: getRuntimeConfig().apiBase,
  timeout: 15000
})

const AUTH_ERROR_MESSAGES = new Set([
  'connection token not found',
  'missing X-Connection-Token header'
])

let reconnectPromise: Promise<string> | null = null

function getConnectionStore() {
  return useConnectionStore(pinia)
}

function shouldReconnect(error: AxiosError<ApiResponse<unknown>>, config: RetriableRequestConfig) {
  const message = error.response?.data?.msg ?? ''
  const status = error.response?.status ?? 0

  return (
    !config.skipConnectionToken &&
    !config._connectionRetried &&
    status === 401 &&
    AUTH_ERROR_MESSAGES.has(message)
  )
}

async function reopenConnection() {
  if (reconnectPromise) {
    return reconnectPromise
  }

  reconnectPromise = (async () => {
    const connectionStore = getConnectionStore()
    const nextProfile = { ...connectionStore.profile }

    if (!nextProfile.host || !nextProfile.username || !nextProfile.port) {
      connectionStore.clearConnection()
      throw new Error('Connection expired. Please reconnect to MySQL.')
    }

    const response = await instance.post<ApiResponse<ConnectionOpenResponse>>(
      '/api/connection/open',
      {
        host: nextProfile.host,
        port: nextProfile.port,
        username: nextProfile.username,
        password: nextProfile.password,
        database: nextProfile.database,
        params: {
          charset: 'utf8mb4'
        }
      },
      {
        skipConnectionToken: true
      } as RetriableRequestConfig
    )

    const data = unwrapResponse(response)
    connectionStore.setConnection(data.connectionToken, nextProfile)
    return data.connectionToken
  })()

  try {
    return await reconnectPromise
  } finally {
    reconnectPromise = null
  }
}

function redirectToConnect() {
  if (typeof window === 'undefined') {
    return
  }

  if (window.location.pathname !== '/connect') {
    window.location.replace('/connect')
  }
}

export async function ensureConnectionReady(forceReconnect = false) {
  const connectionStore = getConnectionStore()

  if (!connectionStore.hasConnection) {
    throw new Error('Connection expired. Please reconnect to MySQL.')
  }

  if (!forceReconnect && connectionStore.token) {
    return connectionStore.token
  }

  return reopenConnection()
}

instance.interceptors.request.use((config) => {
  const nextConfig = config as InternalAxiosRequestConfig & RetriableRequestConfig
  if (nextConfig.skipConnectionToken) {
    return nextConfig
  }

  const connectionStore = getConnectionStore()
  const token = connectionStore.token

  if (token) {
    nextConfig.headers['X-Connection-Token'] = token
  }

  return nextConfig
})

instance.interceptors.response.use(undefined, async (error: AxiosError<ApiResponse<unknown>>) => {
  const config = (error.config ?? {}) as RetriableRequestConfig

  if (shouldReconnect(error, config)) {
    try {
      config._connectionRetried = true
      await reopenConnection()
      return instance.request(config)
    } catch (reconnectError) {
      const reconnectMessage =
        reconnectError instanceof Error
          ? reconnectError.message
          : 'Connection expired. Please reconnect to MySQL.'

      getConnectionStore().clearConnection()
      ElMessage.error(reconnectMessage)
      redirectToConnect()
      return Promise.reject(reconnectError)
    }
  }

  const message =
    error.response?.data?.msg ||
    error.message ||
    'Network request failed'

  if (error.response?.status === 401) {
    getConnectionStore().clearConnection()
    if (!config.silentError) {
      ElMessage.error(message)
    }
    redirectToConnect()
    return Promise.reject(error)
  }

  if (!config.silentError) {
    ElMessage.error(message)
  }
  return Promise.reject(error)
})

function unwrapResponse<T>(response: AxiosResponse<ApiResponse<T>>) {
  const config = (response.config ?? {}) as RetriableRequestConfig

  if (response.status !== 200) {
    const message = response.data?.msg || `HTTP error: ${response.status}`
    if (!config.silentError) {
      ElMessage.error(message)
    }
    throw new Error(message)
  }

  const payload = response.data
  if (payload.code !== 200) {
    const message = payload.msg || 'Request failed'
    if (!config.silentError) {
      ElMessage.error(message)
    }
    throw new Error(message)
  }

  return payload.data
}

const request = {
  get<T>(url: string, config?: RetriableRequestConfig) {
    return instance.get<ApiResponse<T>>(url, config).then(unwrapResponse<T>)
  },

  post<T>(url: string, data?: unknown, config?: RetriableRequestConfig) {
    return instance.post<ApiResponse<T>>(url, data, config).then(unwrapResponse<T>)
  }
}

export default request
