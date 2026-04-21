<template>
  <div class="app-layout">
    <aside class="sidebar" :class="{ 'is-collapsed': isCollapsed }">
      <div class="sidebar-header" @click="router.push('/')">
        <div class="logo-mark">
          <el-icon size="20"><Connection /></el-icon>
        </div>
        <div v-if="!isCollapsed" class="logo-copy">
          <strong>Kafka Console</strong>
          <span>Operate with clarity</span>
        </div>
      </div>

      <template v-if="!isCollapsed">
        <div class="sidebar-section">
          <span class="sidebar-section-label">Workspace</span>
        </div>

        <el-scrollbar class="sidebar-menu-container">
          <el-menu
            :default-active="route.path"
            class="sidebar-menu"
            router
            unique-opened
          >
            <template v-for="item in sidebarMenus" :key="item.id">
              <el-sub-menu v-if="hasChildren(item)" :index="String(item.id)">
                <template #title>
                  <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
                  <span>{{ item.name }}</span>
                </template>
                <template v-for="sub in item.children" :key="sub.id">
                  <el-sub-menu v-if="hasChildren(sub)" :index="String(sub.id)">
                    <template #title>
                      <el-icon v-if="sub.icon"><component :is="sub.icon" /></el-icon>
                      <span>{{ sub.name }}</span>
                    </template>
                    <el-menu-item v-for="leaf in sub.children" :key="leaf.id" :index="leaf.path">
                      <el-icon v-if="leaf.icon"><component :is="leaf.icon" /></el-icon>
                      <template #title>{{ leaf.name }}</template>
                    </el-menu-item>
                  </el-sub-menu>
                  <el-menu-item v-else :index="sub.path">
                    <el-icon v-if="sub.icon"><component :is="sub.icon" /></el-icon>
                    <template #title>{{ sub.name }}</template>
                  </el-menu-item>
                </template>
              </el-sub-menu>
              <el-menu-item v-else :index="item.path">
                <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
                <template #title>{{ item.name }}</template>
              </el-menu-item>
            </template>
          </el-menu>
        </el-scrollbar>
      </template>

      <template v-else>
        <div class="collapsed-rail">
          <template v-for="item in sidebarMenus" :key="item.id">
            <el-popover
              v-if="hasChildren(item)"
              placement="right-start"
              trigger="hover"
              :show-arrow="false"
              :width="280"
              popper-class="sidebar-flyout"
            >
              <template #reference>
                <button
                  type="button"
                  class="collapsed-item"
                  :class="{ 'is-active': isMenuActive(item) }"
                >
                  <el-icon v-if="item.icon" size="20"><component :is="item.icon" /></el-icon>
                </button>
              </template>

              <div class="flyout-panel">
                <div class="flyout-title">{{ item.name }}</div>
                <div class="flyout-list">
                  <template v-for="sub in item.children" :key="sub.id">
                    <template v-if="hasChildren(sub)">
                      <div class="flyout-group-title">{{ sub.name }}</div>
                      <button
                        v-for="leaf in sub.children"
                        :key="leaf.id"
                        type="button"
                        class="flyout-item"
                        :class="{ 'is-active': isMenuActive(leaf) }"
                        @click="router.push(leaf.path)"
                      >
                        <el-icon v-if="leaf.icon"><component :is="leaf.icon" /></el-icon>
                        <span>{{ leaf.name }}</span>
                      </button>
                    </template>
                    <button
                      v-else
                      type="button"
                      class="flyout-item"
                      :class="{ 'is-active': isMenuActive(sub) }"
                      @click="router.push(sub.path)"
                    >
                      <el-icon v-if="sub.icon"><component :is="sub.icon" /></el-icon>
                      <span>{{ sub.name }}</span>
                    </button>
                  </template>
                </div>
              </div>
            </el-popover>

            <el-tooltip
              v-else
              :content="item.name"
              placement="right"
              :show-arrow="false"
              popper-class="sidebar-tooltip"
            >
              <button
                type="button"
                class="collapsed-item"
                :class="{ 'is-active': isMenuActive(item) }"
                @click="router.push(item.path)"
              >
                <el-icon v-if="item.icon" size="20"><component :is="item.icon" /></el-icon>
              </button>
            </el-tooltip>
          </template>
        </div>
      </template>

      <div class="sidebar-footer">
        <button class="collapse-toggle" type="button" @click="isCollapsed = !isCollapsed">
          <el-icon><Expand v-if="isCollapsed" /><Fold v-else /></el-icon>
          <span v-if="!isCollapsed">收起侧栏</span>
        </button>
      </div>
    </aside>

    <main class="main-content">
      <header class="top-header">
        <div class="top-header-left">
          <div class="top-title">{{ route.meta?.title || 'Kafka Console' }}</div>
          <div class="top-subtitle">{{ routeDescription }}</div>
        </div>
        <div class="top-header-right">
          <div class="status-chip">Local Workspace</div>
          <el-dropdown @command="handleUserCommand" trigger="click">
            <div class="user-profile">
              <el-avatar :size="34">{{ userInitial }}</el-avatar>
              <div class="user-copy">
                <strong>{{ permStore.userInfo?.nickname || permStore.userInfo?.username || 'Admin' }}</strong>
                <span>管理员会话</span>
              </div>
              <el-icon><CaretBottom /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人资料</el-dropdown-item>
                <el-dropdown-item command="logout" divided style="color: #f56c6c">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <div class="content-wrapper">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { CaretBottom, Connection, Expand, Fold } from '@element-plus/icons-vue'
import { usePermissionStore } from '@/stores/permissionStore.js'

const route = useRoute()
const router = useRouter()
const permStore = usePermissionStore()
const isCollapsed = ref(false)

const sidebarMenus = computed(() => permStore.menuTree)
const userInitial = computed(() =>
  (permStore.userInfo?.nickname || permStore.userInfo?.username || 'A').slice(0, 1).toUpperCase(),
)

const hasChildren = (item) => Array.isArray(item?.children) && item.children.length > 0

const isPathActive = (path) => {
  if (!path) return false
  return route.path === path || (path !== '/' && route.path.startsWith(`${path}/`))
}

const isMenuActive = (item) => {
  if (!item) return false
  if (isPathActive(item.path)) return true
  return hasChildren(item) ? item.children.some((child) => isMenuActive(child)) : false
}

const routeDescription = computed(() => {
  const title = route.meta?.title || ''
  const map = {
    '首页': '概览今天最常用的 Kafka 工作入口',
    'Kafka总览': '查看集群总览、核心状态和最新风险点',
    'Kafka 总览': '查看集群总览、核心状态和最新风险点',
    '集群管理': '维护连接信息、环境归属与连通性',
    '自动发现': '扫描网段、识别集群入口并完成导入',
    'Topic 管理': '管理分区、副本、配置和高风险变更',
    'Broker 管理': '查看 Broker 承载、Controller 与节点分布',
    '消费组管理': '观察消费组、Lag 与 Offset 干预',
    '消息浏览': '按 Topic 与分区采样查看消息内容',
    '审计日志': '跟踪高风险操作与最近变更记录',
  }
  return map[title] || '统一管理 Kafka 集群、主题、消费组和消息流'
})

const handleUserCommand = (command) => {
  if (command === 'profile') {
    ElMessage.info('个人资料功能待完善')
    return
  }
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
  permStore.reset()
  router.push('/login')
}
</script>

<style scoped>
.app-layout {
  display: flex;
  width: 100%;
  min-width: 100%;
  flex: 1 0 100%;
  min-height: 100vh;
  background: #eaf0f6;
}

.sidebar {
  display: flex;
  flex-direction: column;
  width: 268px;
  min-width: 268px;
  border-right: 1px solid rgba(148, 163, 184, 0.14);
  background:
    radial-gradient(circle at top left, rgba(96, 165, 250, 0.12), transparent 28%),
    linear-gradient(180deg, #f7fbff 0%, #eef4fb 100%);
  transition: width 0.24s ease, min-width 0.24s ease;
}

.sidebar.is-collapsed {
  width: 64px;
  min-width: 64px;
  max-width: 64px;
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: 14px;
  height: 72px;
  padding: 0 20px;
  cursor: pointer;
}

.sidebar.is-collapsed .sidebar-header {
  justify-content: center;
  padding: 0;
}

.logo-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 14px;
  color: #ffffff;
  background: linear-gradient(135deg, #2f7df6 0%, #60a5fa 100%);
  box-shadow: 0 14px 24px rgba(47, 125, 246, 0.24);
}

.logo-copy {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.logo-copy strong {
  color: #0f172a;
  font-size: 20px;
  font-weight: 800;
  letter-spacing: -0.02em;
}

.logo-copy span {
  color: #64748b;
  font-size: 12px;
}

.sidebar-section {
  padding: 0 22px 14px;
}

.sidebar-section-label {
  color: #94a3b8;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.sidebar-menu-container {
  flex: 1;
  min-height: 0;
}

.collapsed-rail {
  display: flex;
  flex: 1;
  min-height: 0;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px 0 16px;
}

.collapsed-item {
  appearance: none;
  -webkit-appearance: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  padding: 0;
  margin: 0;
  border: 1px solid rgba(148, 163, 184, 0.14);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.74);
  color: #475569;
  cursor: pointer;
  transition: all 0.2s ease;
}

.collapsed-item:hover {
  border-color: rgba(47, 125, 246, 0.28);
  color: #2f7df6;
  box-shadow: 0 14px 24px rgba(15, 23, 42, 0.08);
}

.collapsed-item.is-active {
  background: rgba(47, 125, 246, 0.1);
  color: #2f7df6;
  border-color: rgba(47, 125, 246, 0.28);
}

.sidebar-footer {
  padding: 16px;
}

.sidebar.is-collapsed .sidebar-footer {
  display: flex;
  justify-content: center;
  padding: 14px 0;
}

.collapse-toggle {
  appearance: none;
  -webkit-appearance: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  width: 100%;
  height: 44px;
  padding: 0;
  margin: 0;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.7);
  color: #475569;
  cursor: pointer;
  transition: all 0.2s ease;
}

.sidebar.is-collapsed .collapse-toggle {
  width: 40px;
  min-width: 40px;
  height: 40px;
  padding: 0;
  border-radius: 14px;
}

.collapse-toggle:hover {
  border-color: rgba(47, 125, 246, 0.28);
  color: #2f7df6;
}

.main-content {
  display: flex;
  flex: 1;
  min-width: 0;
  flex-direction: column;
}

.top-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 76px;
  padding: 0 28px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.12);
  background: rgba(255, 255, 255, 0.76);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

.top-header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.top-title {
  color: #0f172a;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: -0.03em;
}

.top-subtitle {
  color: #64748b;
  font-size: 13px;
}

.top-header-right {
  display: flex;
  align-items: center;
  gap: 14px;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(47, 125, 246, 0.08);
  color: #2f7df6;
  font-size: 12px;
  font-weight: 700;
}

.content-wrapper {
  flex: 1;
  width: 100%;
  min-height: 0;
  overflow: auto;
}

.user-profile {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  user-select: none;
}

.user-copy {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.user-copy strong {
  color: #0f172a;
  font-size: 14px;
  font-weight: 700;
}

.user-copy span {
  color: #64748b;
  font-size: 12px;
}

.flyout-panel {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.flyout-title {
  color: #94a3b8;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.flyout-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.flyout-group-title {
  margin-top: 6px;
  color: #94a3b8;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.flyout-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 14px;
  border: none;
  border-radius: 14px;
  background: transparent;
  color: #334155;
  cursor: pointer;
  transition: all 0.2s ease;
}

.flyout-item:hover {
  background: rgba(47, 125, 246, 0.08);
}

.flyout-item.is-active {
  background: rgba(47, 125, 246, 0.1);
  color: #2f7df6;
}

:deep(.sidebar-menu) {
  border-right: none;
  background: transparent;
}

:deep(.sidebar-menu:not(.el-menu--collapse)) {
  width: 100%;
}

:deep(.sidebar-menu .el-menu-item),
:deep(.sidebar-menu .el-sub-menu__title) {
  height: 48px;
  margin: 4px 12px;
  border-radius: 14px;
  line-height: 48px;
  color: #334155;
}

:deep(.sidebar-menu .el-menu-item:hover),
:deep(.sidebar-menu .el-sub-menu__title:hover) {
  background: rgba(255, 255, 255, 0.7);
}

:deep(.sidebar-menu .el-menu-item.is-active) {
  background: rgba(47, 125, 246, 0.1);
  color: #2f7df6;
}

:deep(.sidebar-menu .el-menu-item.is-active::before) {
  content: '';
  position: absolute;
  left: 0;
  top: 10px;
  width: 4px;
  height: 28px;
  border-radius: 999px;
  background: #2f7df6;
}

:deep(.sidebar-tooltip) {
  border: 1px solid rgba(148, 163, 184, 0.14);
  border-radius: 14px;
  background: rgba(15, 23, 42, 0.92);
  box-shadow: 0 18px 30px rgba(15, 23, 42, 0.18);
}

:deep(.sidebar-flyout) {
  padding: 12px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 20px 36px rgba(15, 23, 42, 0.12);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

@media (max-width: 1080px) {
  .sidebar {
    width: 64px;
    min-width: 64px;
  }

  .logo-copy,
  .sidebar-section,
  .collapse-toggle span,
  .user-copy,
  .status-chip {
    display: none;
  }
}

@media (max-width: 768px) {
  .top-header {
    padding: 14px 18px;
    flex-direction: column;
    align-items: flex-start;
  }

  .top-header-right {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
