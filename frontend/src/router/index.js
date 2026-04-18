import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/system/login/index.vue'),
    meta: { title: 'Login', hidden: true }
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
        meta: { title: 'Home', icon: 'House' }
      },
      {
        path: 'cicd/pipelines/create',
        name: 'PipelineCreate',
        component: () => import('../views/cicd/PipelineEditor.vue'),
        meta: { title: 'Create Pipeline', hidden: true }
      },
      {
        path: 'cicd/pipelines/:id/edit',
        name: 'PipelineEdit',
        component: () => import('../views/cicd/PipelineEditor.vue'),
        meta: { title: 'Edit Pipeline', hidden: true }
      },
      {
        path: 'cicd/pipelines/:id/runs',
        name: 'PipelineRunList',
        component: () => import('../views/cicd/PipelineRunList.vue'),
        meta: { title: 'Run History', hidden: true }
      },
      {
        path: 'cicd/pipelines/:id/runs/:runId',
        name: 'PipelineRunDetail',
        component: () => import('../views/cicd/PipelineRunDetail.vue'),
        meta: { title: 'Run Detail', hidden: true }
      },
      {
        path: 'es/instances/add',
        name: 'InstanceAdd',
        component: () => import('../views/es/InstanceForm.vue'),
        meta: { title: 'Add Instance', hidden: true }
      },
      {
        path: 'es/instances/edit/:id',
        name: 'InstanceEdit',
        component: () => import('../views/es/InstanceForm.vue'),
        meta: { title: 'Edit Instance', hidden: true }
      },
      {
        path: 'es/instances/:id',
        name: 'InstanceDetail',
        component: () => import('../views/es/InstanceDetail.vue'),
        meta: { title: 'Instance Detail', hidden: true }
      },
      {
        path: 'task-scheduler/workflows/new',
        name: 'WorkflowCreate',
        component: () => import('../views/task-scheduler/WorkflowEditor.vue'),
        meta: { title: 'Create Workflow', hidden: true }
      },
      {
        path: 'task-scheduler/workflows/:id/edit',
        name: 'WorkflowEdit',
        component: () => import('../views/task-scheduler/WorkflowEditor.vue'),
        meta: { title: 'Edit Workflow', hidden: true }
      },
      {
        path: 'task-scheduler/executions',
        name: 'ExecutionHistory',
        component: () => import('../views/task-scheduler/ExecutionHistory.vue'),
        meta: { title: 'Execution History', icon: 'Document' }
      },
      {
        path: 'task-scheduler/executions/:id/logs',
        name: 'ExecutionLog',
        component: () => import('../views/task-scheduler/ExecutionLog.vue'),
        meta: { title: 'Execution Log', hidden: true }
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
    document.title = `${to.meta.title} - DevOps Console`
  }

  const token = localStorage.getItem('access_token')
  if (!token) {
    if (to.path === '/login') {
      next()
    } else {
      next(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
    }
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
      console.error('Failed to load dynamic routes', err)
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
