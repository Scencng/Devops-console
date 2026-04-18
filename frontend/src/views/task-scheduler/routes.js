export default [
  {
    path: 'task-scheduler/workflows',
    name: 'WorkflowList',
    component: () => import('./WorkflowList.vue'),
    meta: { title: '工作流列表', icon: 'Clock' }
  },
  {
    path: 'task-scheduler/workflows/new',
    name: 'WorkflowCreate',
    component: () => import('./WorkflowEditor.vue'),
    meta: { title: '新建工作流', hidden: true }
  },
  {
    path: 'task-scheduler/workflows/:id/edit',
    name: 'WorkflowEdit',
    component: () => import('./WorkflowEditor.vue'),
    meta: { title: '编辑工作流', hidden: true }
  },
  {
    path: 'task-scheduler/executions',
    name: 'ExecutionHistory',
    component: () => import('./ExecutionHistory.vue'),
    meta: { title: '执行历史', icon: 'Document' }
  }
]
