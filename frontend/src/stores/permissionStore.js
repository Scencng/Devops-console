/**
 * Permission store for user info, menu tree, and dynamic route registration.
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getAuthInfo } from '@/api/system/user.js'

const viewModules = import.meta.glob('../views/**/*.vue')

function resolveComponent(component) {
    if (!component) return null

    const fullPath = `../views/${component}.vue`
    const matched = viewModules[fullPath]
    console.log(`[Router] Resolve component: ${component} -> ${fullPath}, matched:`, !!matched)
    if (!matched) {
        console.warn('Available component keys:', Object.keys(viewModules).filter(k => k.includes(component.split('/')[0])))
    }
    if (matched) return matched

    const withExt = `../views/${component}`
    return viewModules[withExt] || null
}

function flattenMenuToRoutes(menus) {
    let routes = []
    for (const menu of menus) {
        if (menu.type === 2 && menu.path && menu.component) {
            const compLoader = resolveComponent(menu.component)
            const route = {
                path: menu.path,
                name: menu.path || (menu.name + '_' + menu.id),
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

        if (menu.children && menu.children.length > 0) {
            routes.push(...flattenMenuToRoutes(menu.children))
        }
    }
    return routes
}

export const usePermissionStore = defineStore('permission', () => {
    const userInfo = ref(null)
    const menuTree = ref([])
    const perms = ref([])
    const roles = ref([])
    const isLoaded = ref(false)

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

        const dynamicRoutes = flattenMenuToRoutes(menuTree.value)
        for (const route of dynamicRoutes) {
            router.addRoute('AppLayout', route)
        }

        isLoaded.value = true
    }

    function reset() {
        userInfo.value = null
        menuTree.value = []
        perms.value = []
        roles.value = []
        isLoaded.value = false
    }

    function hasPerm(perm) {
        return perms.value.includes(perm) || roles.value.includes('admin')
    }

    return { userInfo, menuTree, perms, roles, isLoaded, loadUserAndRoutes, reset, hasPerm }
})
