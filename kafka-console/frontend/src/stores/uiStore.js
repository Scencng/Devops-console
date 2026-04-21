import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

export const useUiStore = defineStore('ui', () => {
  const pendingRequests = ref(0)
  const routeLoading = ref(false)

  const isPageLoading = computed(() => routeLoading.value || pendingRequests.value > 0)

  const startRouteLoading = () => {
    routeLoading.value = true
  }

  const finishRouteLoading = () => {
    routeLoading.value = false
  }

  const incrementRequests = () => {
    pendingRequests.value += 1
  }

  const decrementRequests = () => {
    pendingRequests.value = Math.max(0, pendingRequests.value - 1)
  }

  return {
    pendingRequests,
    routeLoading,
    isPageLoading,
    startRouteLoading,
    finishRouteLoading,
    incrementRequests,
    decrementRequests,
  }
})
