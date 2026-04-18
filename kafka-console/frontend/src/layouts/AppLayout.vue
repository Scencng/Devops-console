<template>
  <div class="app-layout">
    <aside class="sidebar" :class="{ 'is-collapsed': isCollapsed }">
      <div class="sidebar-header" @click="router.push('/')">
        <div class="logo-icon"><el-icon size="22"><Connection /></el-icon></div>
        <span v-if="!isCollapsed" class="logo-text">Kafka Console</span>
      </div>

      <el-scrollbar class="sidebar-menu-container">
        <el-menu :default-active="route.path" class="sidebar-menu" :collapse="isCollapsed" router unique-opened>
          <template v-for="item in sidebarMenus" :key="item.id">
            <el-sub-menu v-if="item.children && item.children.length > 0" :index="String(item.id)">
              <template #title>
                <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
                <span>{{ item.name }}</span>
              </template>
              <template v-for="sub in item.children" :key="sub.id">
                <el-sub-menu v-if="sub.children && sub.children.length > 0" :index="String(sub.id)">
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

      <div class="sidebar-footer" @click="isCollapsed = !isCollapsed">
        <el-icon><Expand v-if="isCollapsed" /><Fold v-else /></el-icon>
      </div>
    </aside>

    <main class="main-content">
      <header class="top-header">
        <div class="header-left">
          <div class="page-title">{{ route.meta?.title || 'Kafka Console' }}</div>
        </div>
        <div class="header-right">
          <el-dropdown @command="handleUserCommand" trigger="click">
            <div class="user-profile">
              <el-avatar :size="32">{{ userInitial }}</el-avatar>
              <span class="user-name">{{ permStore.userInfo?.nickname || permStore.userInfo?.username || 'Admin' }}</span>
              <el-icon><CaretBottom /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人资料</el-dropdown-item>
                <el-dropdown-item command="logout" divided style="color:#f56c6c">退出登录</el-dropdown-item>
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
const userInitial = computed(() => (permStore.userInfo?.nickname || permStore.userInfo?.username || 'A').slice(0, 1).toUpperCase())

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
  min-height: 100vh;
  background: var(--bg-app, #f5f7fa);
}

.sidebar {
  width: 240px;
  min-width: 240px;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--border-color, #e2e8f0);
  background: #ffffff;
  transition: width 0.2s ease, min-width 0.2s ease;
}

.sidebar.is-collapsed {
  width: 80px;
  min-width: 80px;
}

.sidebar-header {
  height: 64px;
  padding: 0 18px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid var(--border-color, #e2e8f0);
  cursor: pointer;
}

.logo-icon {
  width: 36px;
  height: 36px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  color: #ffffff;
  background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%);
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-main, #2d3748);
  white-space: nowrap;
}

.sidebar-menu-container {
  flex: 1;
  min-height: 0;
}

.sidebar-footer {
  height: 52px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-top: 1px solid var(--border-color, #e2e8f0);
  cursor: pointer;
  color: var(--text-sub, #718096);
}

.main-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.top-header {
  height: 64px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-color, #e2e8f0);
  background: #ffffff;
}

.page-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-main, #2d3748);
}

.content-wrapper {
  flex: 1;
  min-height: 0;
  overflow: auto;
  background: var(--bg-app, #f5f7fa);
}

.user-profile {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  user-select: none;
}

.user-name {
  color: var(--text-main, #2d3748);
  font-weight: 500;
}

:deep(.sidebar-menu) {
  border-right: none;
}

:deep(.sidebar-menu:not(.el-menu--collapse)) {
  width: 100%;
}

:deep(.sidebar-menu .el-menu-item),
:deep(.sidebar-menu .el-sub-menu__title) {
  height: 48px;
  line-height: 48px;
}

:deep(.sidebar-menu .el-menu-item.is-active) {
  background: rgba(64, 158, 255, 0.08);
}

@media (max-width: 960px) {
  .sidebar {
    width: 80px;
    min-width: 80px;
  }

  .top-header {
    padding: 0 16px;
  }

  .user-name {
    display: none;
  }
}
</style>
