/**
 * 权限 Store —— 动态路由核心
 * 负责：用户信息、菜单树、动态路由注册
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getAuthInfo } from '@/api/system/user.js'

// =============================================
// 组件路径映射表
// key = 后端返回的 component 字符串
// value = 懒加载的 Vue 组件
// =============================================
const viewModules = import.meta.glob('../views/**/*.vue')

function resolveComponent(component) {
    if (!component) return null
    // 尝试匹配 ../views/{component}.vue
    const fullPath = `../views/${component}.vue`
    const matched = viewModules[fullPath]
    console.log(`[Router] 解析组件: ${component} -> ${fullPath}, 结果:`, !!matched)
    if (!matched) {
        // print keys for matching
        console.warn('可用组件列表Keys:', Object.keys(viewModules).filter(k => k.includes(component.split('/')[0])))
    }
    if (matched) return matched
    // 兼容不带 .vue 的写法
    const withExt = `../views/${component}`
    return viewModules[withExt] || null
}

// 将后端菜单树打平，只提取 type=2 的真实页面为前端路由
function flattenMenuToRoutes(menus) {
    let routes = []
    for (const menu of menus) {
        // 如果是具体页面(type=2)，且配置了路径和组件，则注册为路由
        if (menu.type === 2 && menu.path && menu.component) {
            const compLoader = resolveComponent(menu.component)
            const route = {
                path: menu.path,
                name: menu.path || (menu.name + '_' + menu.id), // 使用 path 作为唯一路由 name 防止覆盖
                component: compLoader || (() => import('../views/404.vue')),
                meta: {
                    title: menu.name,
                    icon: menu.icon || '',
                    menuId: menu.id,
                    affix: false,
                    hidden: menu.visible === 0,
                },
            }
            routes.push(route)
        }

        // 无论何种类型，只要有子菜单，就递归拉平
        if (menu.children && menu.children.length > 0) {
            routes.push(...flattenMenuToRoutes(menu.children))
        }
    }
    return routes
}

export const usePermissionStore = defineStore('permission', () => {
    const userInfo = ref(null)        // 当前用户信息
    const menuTree = ref([])          // 菜单树（用于侧边栏渲染）
    const perms = ref([])             // 权限标识集合（按钮级权限）
    const roles = ref([])             // 角色列表
    const isLoaded = ref(false)       // 动态路由是否已加载

    /**
     * 从后端拉取用户信息和菜单，注册动态路由
     * @param {import('vue-router').Router} router
     */
    async function loadUserAndRoutes(router) {
        if (isLoaded.value) return

        const res = await getAuthInfo()
        const data = res?.data?.data || {}

        userInfo.value = {
            userId: data.userId,
            username: data.username,
            nickname: data.nickname,
            avatar: data.avatar,
        }
        menuTree.value = data.menus || []
        perms.value = data.perms || []
        roles.value = data.roles || []

        // 将后端菜单生成平级的叶子路由，注册到 '/' (AppLayout) 下
        const dynamicRoutes = flattenMenuToRoutes(menuTree.value)
        for (const route of dynamicRoutes) {
            router.addRoute('AppLayout', route)
        }

        isLoaded.value = true
    }

    /** 清除权限信息（退出登录时调用） */
    function reset() {
        userInfo.value = null
        menuTree.value = []
        perms.value = []
        roles.value = []
        isLoaded.value = false
    }

    /** 判断是否拥有某个权限标识 */
    function hasPerm(perm) {
        return perms.value.includes(perm) || roles.value.includes('admin')
    }

    return { userInfo, menuTree, perms, roles, isLoaded, loadUserAndRoutes, reset, hasPerm }
})
