import request from '@/mysql/utils/request'

export type PrincipalKind = 'user' | 'role'

export interface PrincipalSummary {
  user: string
  host: string
  kind: PrincipalKind
  locked: boolean
  passwordExpired: boolean
  plugin: string
  privilegeSummary: string
  privilegeDetails: string
}

export interface SecurityOverviewResponse {
  capabilities: {
    version: string
    supportsRoles: boolean
  }
  users: PrincipalSummary[]
  roles: PrincipalSummary[]
}

const OVERVIEW_TIMEOUT_MS = 15000

let overviewCache: SecurityOverviewResponse | null = null
let overviewRequest: Promise<SecurityOverviewResponse> | null = null

function withTimeout<T>(promise: Promise<T>, timeoutMs: number) {
  return new Promise<T>((resolve, reject) => {
    const timer = window.setTimeout(() => reject(new Error('REQUEST_TIMEOUT')), timeoutMs)
    promise
      .then((value) => {
        window.clearTimeout(timer)
        resolve(value)
      })
      .catch((error) => {
        window.clearTimeout(timer)
        reject(error)
      })
  })
}

export function getSecurityOverviewCache() {
  return overviewCache
}

export function hasSecurityOverviewCache() {
  return overviewCache !== null
}

export function clearSecurityOverviewCache() {
  overviewCache = null
  overviewRequest = null
}

export function fetchSecurityOverview(force = false) {
  if (!force && overviewCache) {
    return Promise.resolve(overviewCache)
  }

  if (!force && overviewRequest) {
    return overviewRequest
  }

  overviewRequest = withTimeout(
    request.get<SecurityOverviewResponse>('/api/security/overview'),
    OVERVIEW_TIMEOUT_MS
  )
    .then((response) => {
      overviewCache = response
      return response
    })
    .finally(() => {
      overviewRequest = null
    })

  return overviewRequest
}

export function preloadSecurityOverview() {
  return fetchSecurityOverview(false)
}
