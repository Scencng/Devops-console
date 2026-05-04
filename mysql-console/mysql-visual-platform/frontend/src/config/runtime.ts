export interface RuntimeMySQLConfig {
  host: string
  port: number
  username: string
  database: string
}

export interface RuntimeConfig {
  apiBase: string
  mysql: RuntimeMySQLConfig
}

declare const __APP_BUILD_ID__: string

const defaultRuntimeConfig: RuntimeConfig = {
  apiBase: '/',
  mysql: {
    host: '127.0.0.1',
    port: 3306,
    username: 'root',
    database: ''
  }
}

let runtimeConfig: RuntimeConfig = { ...defaultRuntimeConfig, mysql: { ...defaultRuntimeConfig.mysql } }

function mergeRuntimeConfig(payload: Partial<RuntimeConfig> | null | undefined) {
  runtimeConfig = {
    apiBase: normalizeApiBase(payload?.apiBase ?? defaultRuntimeConfig.apiBase),
    mysql: {
      host: payload?.mysql?.host || defaultRuntimeConfig.mysql.host,
      port: normalizePort(payload?.mysql?.port, defaultRuntimeConfig.mysql.port),
      username: payload?.mysql?.username || defaultRuntimeConfig.mysql.username,
      database: payload?.mysql?.database ?? defaultRuntimeConfig.mysql.database
    }
  }
}

export async function loadRuntimeConfig() {
  try {
    const response = await fetch(`./app-config.json?v=${encodeURIComponent(__APP_BUILD_ID__)}`, { cache: 'no-store' })
    if (!response.ok) {
      mergeRuntimeConfig(undefined)
      return runtimeConfig
    }

    const payload = (await response.json()) as Partial<RuntimeConfig>
    mergeRuntimeConfig(payload)
    return runtimeConfig
  } catch {
    mergeRuntimeConfig(undefined)
    return runtimeConfig
  }
}

export function getRuntimeConfig() {
  return runtimeConfig
}

function normalizeApiBase(value: string) {
  const trimmed = value.trim()
  if (!trimmed) {
    return '/'
  }

  if (trimmed === '/') {
    return '/'
  }

  return trimmed.endsWith('/') ? trimmed.slice(0, -1) : trimmed
}

function normalizePort(value: number | undefined, fallback: number) {
  if (typeof value !== 'number' || !Number.isFinite(value) || value <= 0) {
    return fallback
  }

  return Math.trunc(value)
}
