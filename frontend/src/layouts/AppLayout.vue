<template>
  <div class="app-layout">
    <!-- 左侧边栏 -->
    <aside class="sidebar" :class="{ 'is-collapsed': isCollapsed }">
      <!-- Logo -->
      <div class="sidebar-header" @click="handleLogoClick">
        <div class="logo-icon">
          <el-icon size="24" color="#fff">
            <Monitor />
          </el-icon>
        </div>
        <transition name="fade">
          <span v-if="!isCollapsed" class="logo-text">DevOps Console</span>
        </transition>
      </div>

      <!-- 垂直菜单 -->
      <el-scrollbar class="sidebar-menu-container">
        <el-menu
          :default-active="$route.path"
          class="sidebar-menu"
          :collapse="isCollapsed"
          background-color="#ffffff"
          text-color="#606266"
          active-text-color="#409EFF"
          unique-opened
          router
        >
          <template v-for="item in sidebarMenus" :key="item.id">
            <!-- 有子菜单 -->
            <el-sub-menu v-if="item.children && item.children.length > 0" :index="String(item.id)">
              <template #title>
                <el-icon><component :is="item.icon" /></el-icon>
                <span>{{ item.name }}</span>
              </template>
              <template v-for="sub in item.children" :key="sub.id">
                <!-- 二级有子菜单（三级菜单） -->
                <el-sub-menu v-if="sub.children && sub.children.length > 0" :index="String(sub.id)">
                  <template #title>
                    <el-icon><component :is="sub.icon" /></el-icon>
                    <span>{{ sub.name }}</span>
                  </template>
                  <el-menu-item v-for="leaf in sub.children" :key="leaf.id" :index="leaf.path">
                    <el-icon v-if="leaf.icon"><component :is="leaf.icon" /></el-icon>
                    <template #title>{{ leaf.name }}</template>
                  </el-menu-item>
                </el-sub-menu>
                <!-- 二级无子菜单 -->
                <el-menu-item v-else :index="sub.path">
                  <el-icon v-if="sub.icon"><component :is="sub.icon" /></el-icon>
                  <template #title>{{ sub.name }}</template>
                </el-menu-item>
              </template>
            </el-sub-menu>
            <!-- 无子菜单（一级直达） -->
            <el-menu-item v-else :index="item.path">
              <el-icon><component :is="item.icon" /></el-icon>
              <template #title><span>{{ item.name }}</span></template>
            </el-menu-item>
          </template>
        </el-menu>
      </el-scrollbar>

      <!-- 底部折叠按钮 -->
      <div class="sidebar-footer" @click="toggleCollapse">
        <el-icon :size="20" color="#909399">
          <Expand v-if="isCollapsed" />
          <Fold v-else />
        </el-icon>
      </div>
    </aside>

    <!-- 右侧主体内容 -->
    <main class="main-content">
<!-- ... (Existing Header and Content Wrapper code remains unchanged up to script) ... -->
<!-- However, since replace_file_content expects contiguous blocks, I have to be careful. -->
<!-- If I am just replacing the menu area and data, I should do it in two chunks or use multi_replace. -->
<!-- But wait, I need to replace the TEMPLATE part and the SCRIPT part. replace_file_content only does one block. -->
<!-- I will use multi_replace_file_content. -->
      <!-- 顶部 Header -->
      <header class="top-header">
        <!-- 左侧：面包屑或当前路径标题 (可选) -->
        <div class="header-left">
           <!-- 实例选择器 (从旧版迁移过来) -->
           <div class="instance-selector">
              <el-dropdown
                trigger="click"
                placement="bottom-start"
                @visible-change="handleDropdownVisible"
              >
                <div class="instance-dropdown-trigger">
                  <div class="trigger-icon-box" :style="{ background: getTypeColor(getSelectedInstanceType()) + '20' }">
                     <el-icon :color="getTypeColor(getSelectedInstanceType())">
                        <component :is="getTypeIcon(getSelectedInstanceType())" />
                      </el-icon>
                  </div>
                  <div class="trigger-info">
                    <span class="trigger-label">当前实例</span>
                    <span class="trigger-value">{{ selectedInstanceName || '选择实例' }}</span>
                  </div>
                  <el-icon class="dropdown-arrow"><ArrowDown /></el-icon>
                </div>
                <!-- 保持原有的 Dropdown Menu 内容不变 -->
                <template #dropdown>
                  <el-dropdown-menu class="instance-management-menu">
                    <!-- ... 实例管理下拉内容保持不变，复用原有逻辑 ... -->
                    <div class="instance-management-content">
                      <div class="page-header">
                        <h2>实例管理</h2>
                        <div class="header-actions">
                          <el-button size="small" @click="$router.push('/es/instances')">
                             <el-icon><List /></el-icon>自发现列表
                          </el-button>
                          <el-button type="primary" size="small" @click="handleAddInstance">
                            <el-icon><Plus /></el-icon>添加实例
                          </el-button>
                          <el-button size="small" @click="handleRefreshInstances">
                            <el-icon><Refresh /></el-icon>刷新
                          </el-button>
                        </div>
                      </div>

                      <div class="filter-section">
                        <el-input v-model="searchQuery" placeholder="搜索实例..." size="small" clearable>
                           <template #prefix><el-icon><Search /></el-icon></template>
                        </el-input>
                        <el-select v-model="typeFilter" placeholder="类型" size="small" clearable>
                           <el-option v-for="t in instanceTypes" :key="t.id" :label="t.type_name" :value="t.type_name" />
                        </el-select>
                         <el-select v-model="statusFilter" placeholder="状态" size="small" clearable>
                          <el-option label="活跃(旧)" value="active" />
                          <el-option label="在线" value="online" />
                          <el-option label="离线" value="offline" />
                          <el-option label="非活跃" value="inactive" />
                          <el-option label="错误" value="error" />
                        </el-select>
                      </div>

                      <div class="instance-list-container">
                        <div
                          v-for="instance in filteredInstanceList"
                          :key="instance.id"
                          class="instance-row"
                          :class="{ 'selected': selectedInstance === instance.id }"
                          @click="selectInstance(instance.id)"
                        >
                           <div class="instance-info">
                              <el-icon class="type-icon" :color="getTypeColor(instance.instance_type)">
                                <component :is="getTypeIcon(instance.instance_type)" />
                              </el-icon>
                              <div class="instance-details">
                                <div class="instance-name">{{ instance.name }}</div>
                                <div class="instance-address">{{ instance.address }}</div>
                              </div>
                           </div>
                           <div class="instance-actions">
                               <el-tag :type="getStatusType(instance.status)" size="small" effect="plain">{{ getStatusLabel(instance.status) }}</el-tag>
                           </div>
                        </div>
                      </div>
                    </div>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
           </div>
        </div>

        <!-- 右侧：全局操作 -->
        <div class="header-right">
           <!-- 通知 -->
           <el-tooltip content="通知" effect="dark" placement="bottom">
              <el-badge :value="3" :max="99" class="action-item">
                <el-icon size="20"><Bell /></el-icon>
              </el-badge>
           </el-tooltip>

           <!-- 主题切换 -->
           <el-tooltip :content="isDark ? '切换亮色' : '切换暗色'" effect="dark" placement="bottom">
             <div class="action-item" @click="toggleTheme">
                <el-icon size="20">
                  <Sunny v-if="isDark"/>
                  <Moon v-else/>
                </el-icon>
             </div>
           </el-tooltip>

           <!-- 用户菜单 -->
           <el-dropdown @command="handleUserCommand" trigger="click">
              <div class="user-profile">
                <el-avatar :size="32" :src="permStore.userInfo?.avatar || 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png'" />
                <span class="user-name">{{ permStore.userInfo?.nickname || permStore.userInfo?.username || 'Admin' }}</span>
                <el-icon><CaretBottom /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu class="user-dropdown">
                  <div class="user-dropdown-header">
                    <p class="name">{{ permStore.userInfo?.nickname || permStore.userInfo?.username }}</p>
                    <p class="role">{{ permStore.roles?.join(', ') || '' }}</p>
                  </div>
                  <el-dropdown-item v-if="permStore.hasPerm('sys:user:profile')" command="profile" divided><el-icon><User /></el-icon>个人中心</el-dropdown-item>
                  <el-dropdown-item command="logout" divided style="color: #f56c6c;"><el-icon><SwitchButton /></el-icon>退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
           </el-dropdown>
        </div>
      </header>

      <!-- 页面内容区 -->
      <div class="content-wrapper">
         <router-view v-slot="{ Component }">
            <component :is="Component" :key="route.fullPath" />
         </router-view>
      </div>
    </main>
  </div>
</template>

<script setup>
import {computed, onMounted, ref, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {ElMessage} from 'element-plus'
import {getInstanceList, getInstanceTypes} from '@/api/instance.js'
import {getSelectedInstanceType, setSelectedInstance} from '@/stores/instanceStore.js'
import {usePermissionStore} from '@/stores/permissionStore.js'
import {
  ArrowDown,
  Bell,
  Box,
  CaretBottom,
  DataLine,
  DocumentCopy,
  Expand,
  Fold,
  List,
  Monitor,
  Moon,
  Plus,
  Refresh,
  Search,
  Sunny,
  SwitchButton,
  User,
  WarningFilled
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const permStore = usePermissionStore()

// 状态
const isDark = ref(false)
const isCollapsed = ref(false)
const selectedInstance = ref('')
const instanceList = ref([])
const instanceTypes = ref([])
const searchQuery = ref('')
const typeFilter = ref('')
const statusFilter = ref('')

// ======================================================
// 动态菜单：直接使用 permissionStore 中的 menuTree
// menuTree 已由路由守卫加载，这里只需读取
// ======================================================
const sidebarMenus = computed(() => permStore.menuTree)

// 保留，兼容其他地方的引用
const allMenuRoutes = computed(() => [])


// 计算选中实例名称
const selectedInstanceName = computed(() => {
  if (!selectedInstance.value) return ''
  const instance = instanceList.value.find(item => item.id === selectedInstance.value)
  return instance ? instance.name : ''
})

// 过滤实例列表
const filteredInstanceList = computed(() => {
  let result = [...instanceList.value]
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(instance =>
      instance.name.toLowerCase().includes(query) ||
      instance.address.toLowerCase().includes(query)
    )
  }
  if (typeFilter.value) {
    result = result.filter(instance => instance.instance_type === typeFilter.value)
  }
  if (statusFilter.value) {
    result = result.filter(instance => instance.status === statusFilter.value)
  }
  return result
})

// === 方法 ===

const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
}

const toggleTheme = () => {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

const handleLogoClick = () => {
  router.push('/')
}

const handleUserCommand = (command) => {
  switch (command) {
    case 'profile':
      ElMessage.info('个人资料功能待实现')
      break
    case 'settings':
      router.push('/settings')
      break
    case 'logout':
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      permStore.reset()
      ElMessage.success('已退出登录')
      // 跳转登录页并强制刷新，彻底清除动态路由和内存状态
      window.location.href = '/login'
      break
  }
}

// 实例管理相关
const selectInstance = (instanceId) => {
  selectedInstance.value = instanceId
  const instance = instanceList.value.find(item => item.id === instanceId)
  if (instance) {
    setSelectedInstance(instance)
    ElMessage.success(`已切换到: ${instance.name}`)
    // 切换实例后保持在当前页面，不强制跳转
  }
}

const handleAddInstance = () => {
  router.push('/es/instances/add')
}

const handleRefreshInstances = async () => {
  await Promise.all([fetchInstances(), fetchInstanceTypes()])
  ElMessage.success('数据已刷新')
}

const handleDropdownVisible = (visible) => {
  if (visible) {
    Promise.all([fetchInstances(), fetchInstanceTypes()])
  }
}

// Data fetching
const fetchInstances = async () => {
  try {
    const response = await getInstanceList({ page: 1, page_size: 100 })
    const listData = response.data?.list || {}
    instanceList.value = listData.data || []

    if (instanceList.value.length > 0 && !selectedInstance.value) {
      // 优先选择 elasticsearch 或 kubernetes，避免默认选中 prometheus
      const preferType = route.path.startsWith('/es') ? 'elasticsearch' : (route.path.startsWith('/k8s') ? 'kubernetes' : null)
      let defaultInstance = instanceList.value.find(inst => inst.instance_type === preferType)
      if (!defaultInstance) {
          defaultInstance = instanceList.value.find(inst => inst.instance_type === 'elasticsearch' || inst.instance_type === 'kubernetes')
      }
      if (!defaultInstance) {
          defaultInstance = instanceList.value[0]
      }

      selectedInstance.value = defaultInstance.id
      setSelectedInstance(defaultInstance)
    }
  } catch (error) {
    console.error('获取实例列表失败:', error)
  }
}

const fetchInstanceTypes = async () => {
  try {
    const response = await getInstanceTypes()
    instanceTypes.value = response.data?.instance_types || []
  } catch (error) {
    console.error('获取类型失败:', error)
  }
}

// Utility
const getTypeIcon = (type) => {
  const icons = {
    elasticsearch: Monitor, kubernetes: Box, kibana: DataLine,
    logstash: DocumentCopy, filebeat: DocumentCopy, metricbeat: DocumentCopy, apm: WarningFilled
  }
  return icons[type] || Monitor
}

const getTypeColor = (type) => {
  const colors = {
    elasticsearch: '#005FD4', kubernetes: '#326CE5', kibana: '#00BFB3',
    logstash: '#FEC514', filebeat: '#00BFB3', metricbeat: '#00BFB3', apm: '#8A0A4A'
  }
  return colors[type] || '#666'
}

const getStatusType = (status) => ({ active: 'success', online: 'success', offline: 'danger', inactive: 'info', error: 'danger' }[status] || 'info')
const getStatusLabel = (status) => ({ active: '在线', online: '在线', offline: '离线', inactive: '非活跃', error: '异常' }[status] || status)

onMounted(() => {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark') {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
  fetchInstances()
  fetchInstanceTypes()
})

// Watchers
watch(() => getSelectedInstanceType(), () => {
  // Instance Type change might trigger menu refresh automatically via computed
})
</script>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  background-color: #f5f7fa; /* AutoOps 主题背景色 */
  overflow: hidden;
}

/* Sidebar Styling */
.sidebar {
  width: 260px;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 2px 0 8px rgba(0,0,0,0.04);
  z-index: 200;
  flex-shrink: 0;
  border-right: 1px solid #e4e7ed;
}

.sidebar.is-collapsed {
  width: 64px;
}

.sidebar-header {
  height: 70px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  border-bottom: 1px solid #e4e7ed;
  cursor: pointer;
  overflow: hidden;
}

.logo-icon {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, #005FD4, #326CE5);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  flex-shrink: 0;
}

.logo-text {
  color: #2c3e50;
  font-weight: 700;
  font-size: 18px;
  white-space: nowrap;
  letter-spacing: 0.5px;
}

.sidebar-menu-container {
  flex: 1;
}

.sidebar-menu {
  border-right: none;
  padding: 10px 0;
}

:deep(.el-menu-item) {
  height: 50px;
  line-height: 50px;
  margin: 4px 12px;
  border-radius: 8px;
  width: auto;
}

:deep(.el-menu-item.is-active) {
  background: #ecf5ff !important;
  color: #409EFF !important;
  font-weight: 500;
}

:deep(.el-menu-item:hover) {
  background-color: #f5f7fa !important;
  color: #303133 !important;
}

/* 子菜单样式 */
:deep(.el-sub-menu__title) {
  height: 50px;
  line-height: 50px;
  margin: 4px 12px;
  border-radius: 8px;
  color: #606266 !important;
}

:deep(.el-sub-menu__title:hover) {
  background-color: #f5f7fa !important;
  color: #303133 !important;
}

:deep(.el-sub-menu.is-active > .el-sub-menu__title) {
  color: #409EFF !important;
}

:deep(.el-menu--inline .el-menu-item) {
  background-color: #fafafa !important;
}

.sidebar-footer {
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-top: 1px solid #e4e7ed;
  cursor: pointer;
  transition: background 0.2s;
}

.sidebar-footer:hover {
  background: #f5f7fa;
}

/* Main Content Styling */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background-color: #f5f7fa; /* AutoOps 主题背景色 */
}

/* Top Header */
.top-header {
  height: 70px;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.02);
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.instance-dropdown-trigger {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.02);
  min-width: 200px;
}

.instance-dropdown-trigger:hover {
  border-color: #409eff;
  background: #f9faff;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.1);
  transform: translateY(-1px);
}

.trigger-icon-box {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.3s ease;
}

.instance-dropdown-trigger:hover .trigger-icon-box {
  transform: scale(1.05);
}

.trigger-info {
  display: flex;
  flex-direction: column;
  flex: 1;
  gap: 2px;
}

.trigger-label {
  font-size: 11px;
  color: #909399;
  line-height: 1.2;
}

.trigger-value {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  line-height: 1.2;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 140px;
}

.dropdown-arrow {
  margin-left: 8px;
  color: #c0c4cc;
  transition: transform 0.3s;
}

.instance-dropdown-trigger:hover .dropdown-arrow {
  color: #409eff;
  transform: rotate(180deg);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.action-item {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #606266;
  cursor: pointer;
  transition: background 0.2s;
}

.action-item:hover {
  background: #f2f3f5;
  color: #303133;
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
}

.user-profile:hover {
  background: #f2f3f5;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

/* Dropdown Custom Styles */
.user-dropdown-header {
  padding: 16px 20px;
  border-bottom: 1px solid #EBEEF5;
  margin-bottom: 4px;
}

.user-dropdown-header .name {
  font-weight: 600;
  font-size: 16px;
  color: #303133;
  margin-bottom: 4px;
}

.user-dropdown-header .role {
  font-size: 12px;
  color: #909399;
}

/* Instance Menu Styles - Copied/Adpated from old layout */
.instance-management-content {
  width: 400px;
  padding: 16px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.page-header h2 { font-size: 16px; margin: 0; }
.filter-section { display: flex; gap: 8px; margin-bottom: 12px; }
.instance-list-container { max-height: 300px; overflow-y: auto; }
.instance-row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px; border-radius: 8px; cursor: pointer; transition: background 0.2s;
}
.instance-row:hover { background: #f5f7fa; }
.instance-row.selected { background: #ecf5ff; }
.instance-info { display: flex; align-items: center; gap: 10px; }
.instance-details .instance-name { font-weight: 500; font-size: 14px; }
.instance-details .instance-address { font-size: 12px; color: #909399; }

/* Content Wrapper */
.content-wrapper {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
  overflow-x: hidden;
}

/* Transitions */
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

.fade-transform-enter-active, .fade-transform-leave-active { transition: all 0.3s; }
.fade-transform-enter-from { opacity: 0; transform: translateX(10px); }
.fade-transform-leave-to { opacity: 0; transform: translateX(-10px); }

/* Dark Mode Overrides */
html.dark .sidebar { background: #18181c; } /* Darker sidebar */
html.dark .main-content { background: #121212; }
html.dark .top-header {
  background: #1e1e2d;
  border-bottom-color: #303030;
}
html.dark .instance-dropdown-trigger {
  background: #262734;
  border-color: #262734;
}
html.dark .instance-dropdown-trigger:hover {
  background: #30303d;
  border-color: #3e3e4a;
}
html.dark .trigger-label { color: #8d90a1; }
html.dark .trigger-value { color: #e1e1e1; }
html.dark .search-input :deep(.el-input__wrapper) { background-color: #262734; box-shadow: none; }
html.dark .action-item { color: #a0a0b0; }
html.dark .action-item:hover { background: #2a2a35; color: #fff; }
html.dark .user-name { color: #e1e1e1; }
html.dark .user-profile:hover { background: #2a2a35; }
</style>
