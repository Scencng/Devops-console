import { createRouter, createWebHistory } from 'vue-router'

import { ensureConnectionReady } from '@/utils/request'
import ConnectionView from '@/views/ConnectionView.vue'
import WorkspaceView from '@/views/WorkspaceView.vue'
import pinia from '@/stores'
import { useConnectionStore } from '@/stores/connection'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/connect'
    },
    {
      path: '/connect',
      name: 'connect',
      component: ConnectionView
    },
    {
      path: '/workspace',
      name: 'workspace',
      component: WorkspaceView
    }
  ]
})

router.beforeEach(async (to) => {
  const connectionStore = useConnectionStore(pinia)

  if (to.name === 'workspace') {
    if (!connectionStore.hasConnection) {
      return { name: 'connect' }
    }

    try {
      await ensureConnectionReady(true)
    } catch {
      connectionStore.clearConnection()
      return { name: 'connect' }
    }
  }

  return true
})

export default router
