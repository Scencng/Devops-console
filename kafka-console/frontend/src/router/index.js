import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/system/login/index.vue'),
    meta: { title: '登录', hidden: true }
  },
  {
    path: '/',
    name: 'AppLayout',
    component: () => import('../layouts/AppLayout.vue'),
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: { title: '首页', icon: 'House' }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('../views/404.vue'),
    meta: { hidden: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 })
})

router.beforeEach(async (to, from, next) => {
  if (to.meta?.title) {
    document.title = `${to.meta.title} - Kafka Console`
  }

  const token = localStorage.getItem('access_token')
  if (!token) {
    if (to.path === '/login') next()
    else next(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
    return
  }

  if (to.path === '/login') {
    next('/')
    return
  }

  const { usePermissionStore } = await import('../stores/permissionStore.js')
  const permStore = usePermissionStore()
  if (!permStore.isLoaded) {
    try {
      await permStore.loadUserAndRoutes(router)
      next({ path: to.path, query: to.query, hash: to.hash, replace: true })
    } catch (err) {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      next('/login')
    }
  } else {
    next()
  }
})

router.onError((error) => {
  console.error('Router error:', error)
})

export default router
