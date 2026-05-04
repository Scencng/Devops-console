import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { getRuntimeConfig } from '@/mysql/runtime'

export interface ConnectionFormState {
  host: string
  port: number
  username: string
  password: string
  database: string
}

interface PersistedConnectionState {
  token: string
  connected: boolean
  profile: ConnectionFormState
}

const STORAGE_KEY = 'mysql-visual-platform:connection'

function getDefaultProfile(): ConnectionFormState {
  const runtimeConfig = getRuntimeConfig()
  return {
    host: runtimeConfig.mysql.host,
    port: runtimeConfig.mysql.port,
    username: runtimeConfig.mysql.username,
    password: runtimeConfig.mysql.password,
    database: runtimeConfig.mysql.database
  }
}

function buildProfile(nextProfile?: Partial<ConnectionFormState>): ConnectionFormState {
  return {
    ...getDefaultProfile(),
    ...(nextProfile ?? {})
  }
}

function loadPersistedState(): PersistedConnectionState {
  const fallback: PersistedConnectionState = {
    token: '',
    connected: false,
    profile: getDefaultProfile()
  }

  if (typeof window === 'undefined') {
    return fallback
  }

  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return fallback
    }

    const parsed = JSON.parse(raw) as Partial<PersistedConnectionState>
    return {
      token: typeof parsed.token === 'string' ? parsed.token : '',
      connected: Boolean(parsed.connected && parsed.token),
      profile: buildProfile(parsed.profile)
    }
  } catch {
    return fallback
  }
}

export const useConnectionStore = defineStore('mysqlConnection', () => {
  const persisted = loadPersistedState()
  const token = ref<string>(persisted.token)
  const connected = ref(persisted.connected)
  const profile = ref<ConnectionFormState>(persisted.profile)

  const hasConnection = computed(() => connected.value && token.value.length > 0)

  function persistState() {
    if (typeof window === 'undefined') {
      return
    }

    const payload: PersistedConnectionState = {
      token: token.value,
      connected: connected.value,
      profile: profile.value
    }

    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(payload))
  }

  function setConnection(nextToken: string, nextProfile?: Partial<ConnectionFormState>) {
    token.value = nextToken
    connected.value = true

    if (nextProfile) {
      profile.value = buildProfile({
        ...profile.value,
        ...nextProfile
      })
    }

    persistState()
  }

  function updateProfile(nextProfile: Partial<ConnectionFormState>) {
    profile.value = buildProfile({
      ...profile.value,
      ...nextProfile
    })

    persistState()
  }

  function clearConnection() {
    token.value = ''
    connected.value = false
    profile.value = {
      ...profile.value,
      database: getDefaultProfile().database
    }
    persistState()
  }

  return {
    token,
    connected,
    profile,
    hasConnection,
    setConnection,
    updateProfile,
    clearConnection
  }
})

