import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getKafkaClusterOptions } from '@/api/kafka.js'

const SELECTED_CLUSTER_KEY = 'kafka:selected-cluster-id'

const normalizeClusterId = (value) => {
  if (value === null || value === undefined || value === '') return null
  const numericValue = Number(value)
  return Number.isNaN(numericValue) ? value : numericValue
}

export const useKafkaStore = defineStore('kafka', () => {
  const clusterOptions = ref([])
  const selectedClusterId = ref(normalizeClusterId(localStorage.getItem(SELECTED_CLUSTER_KEY)))

  const setSelectedClusterId = (value) => {
    selectedClusterId.value = normalizeClusterId(value)
    if (selectedClusterId.value === null) {
      localStorage.removeItem(SELECTED_CLUSTER_KEY)
      return
    }
    localStorage.setItem(SELECTED_CLUSTER_KEY, String(selectedClusterId.value))
  }

  const syncSelectedCluster = (preferredId = selectedClusterId.value) => {
    const normalizedPreferredId = normalizeClusterId(preferredId)
    const exists = clusterOptions.value.some((item) => item.id === normalizedPreferredId)
    if (exists) {
      setSelectedClusterId(normalizedPreferredId)
      return normalizedPreferredId
    }

    const fallbackId = clusterOptions.value[0]?.id ?? null
    setSelectedClusterId(fallbackId)
    return fallbackId
  }

  const loadClusterOptions = async ({ force = false } = {}) => {
    if (!force && clusterOptions.value.length > 0) {
      syncSelectedCluster()
      return clusterOptions.value
    }

    const res = await getKafkaClusterOptions()
    clusterOptions.value = res?.data?.data || []
    syncSelectedCluster()
    return clusterOptions.value
  }

  const reset = () => {
    clusterOptions.value = []
    setSelectedClusterId(null)
  }

  return {
    clusterOptions,
    selectedClusterId,
    setSelectedClusterId,
    syncSelectedCluster,
    loadClusterOptions,
    reset,
  }
})
