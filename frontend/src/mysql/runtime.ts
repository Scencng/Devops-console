export interface RuntimeMySQLConfig {
  host: string
  port: number
  username: string
  password: string
  database: string
}

export interface RuntimeConfig {
  apiBase: string
  mysql: RuntimeMySQLConfig
}

const runtimeConfig: RuntimeConfig = {
  apiBase: '/api/v1/mysql',
  mysql: {
    host: '47.104.247.159',
    port: 8002,
    username: 'root',
    password: 'peppa-pig',
    database: 'devops_console'
  }
}

export async function loadRuntimeConfig() {
  return runtimeConfig
}

export function getRuntimeConfig() {
  return runtimeConfig
}
