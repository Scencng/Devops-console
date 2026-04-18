// Chaos Mesh 路由配置（静态路由方式）
// 前端路由配置文件: src/router/index.js
// 添加到 routes 数组的 children 中

// Chaos Mesh 路由配置
const chaosRoutes = [
  {
    path: 'chaos',
    name: 'ChaosMeshList',
    component: () => import('../views/chaos/ChaosMeshList'),
    meta: { title: '混沌实验列表', icon: 'DataAnalysis' }
  },
  {
    path: 'chaos/create',
    name: 'ChaosMeshCreate',
    component: () => import('../views/chaos/ChaosMeshCreate'),
    meta: { title: '创建混沌实验', hidden: true }
  },
  {
    path: 'chaos/detail',
    name: 'ChaosMeshDetail',
    component: () => import('../views/chaos/ChaosMeshDetail'),
    meta: { title: '混沌实验详情', hidden: true }
  }
];

// ============================================================
// 方式一：添加到静态路由（用于调试或无需权限控制）
// 在 src/router/index.js 的 routes[0].children 数组中添加
// ============================================================
/*
  {
    path: '/',
    name: 'AppLayout',
    component: () => import('../layouts/AppLayout.vue'),
    children: [
      // ... 现有路由
      ...chaosRoutes  // 添加Chaos Mesh路由
    ]
  }
*/

// ============================================================
// 方式二：使用动态路由（推荐，生产环境）
// 路由通过后端菜单表动态加载
// 执行 chaos_menu_simple.sql 后，登录时系统会自动注册路由
// ============================================================

// 导出供其他模块使用
export { chaosRoutes };
