<template>
  <div class="page-container">
    <el-card class="page-header-card">
       <div class="page-header">
         <div class="header-left">
           <div class="header-icon-wrapper">
             <el-icon :size="24"><Platform /></el-icon>
           </div>
           <div class="header-title-wrapper">
             <h2>实例管理</h2>
             <p class="subtitle">管理 Elasticsearch 及相关组件实例</p>
           </div>
         </div>
         <div class="header-right">
           <el-button type="primary" @click="handleAdd" :icon="Plus">添加实例</el-button>
           <el-button @click="handleRefresh" :icon="Refresh">刷新</el-button>
         </div>
       </div>
    </el-card>

    <!-- Filters -->
    <el-card class="content-card mb-20 p-4">
      <div class="filter-row flex items-center gap-4">
        <el-input
          v-model="searchQuery"
          placeholder="搜索实例名称或地址"
          class="filter-item w-64"
          clearable
          :prefix-icon="Search"
          @keyup.enter="handleSearch"
        />
        <el-select v-model="typeFilter" placeholder="类型筛选" clearable class="filter-item w-40">
            <el-option label="Elasticsearch" value="elasticsearch" />
            <el-option label="Kibana" value="kibana" />
            <el-option label="Logstash" value="logstash" />
            <el-option label="Filebeat" value="filebeat" />
            <el-option label="Metricbeat" value="metricbeat" />
            <el-option label="APM" value="apm" />
            <el-option label="Prometheus" value="prometheus" />
        </el-select>
        <el-select v-model="statusFilter" placeholder="状态筛选" clearable class="filter-item w-32">
            <el-option label="在线" value="active" />
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="非活跃" value="inactive" />
            <el-option label="错误" value="error" />
        </el-select>
        <div class="flex-spacer flex-1"></div>
        <el-button-group>
            <el-button type="primary" @click="handleSearch" :loading="searching" :icon="Search">搜索</el-button>
            <el-button @click="handleResetFilter" :icon="RefreshLeft">重置</el-button>
        </el-button-group>
        <el-divider direction="vertical" class="mx-4" />
        <el-radio-group v-model="viewMode" size="small">
            <el-radio-button value="table"><el-icon><List /></el-icon></el-radio-button>
            <el-radio-button value="card"><el-icon><Grid /></el-icon></el-radio-button>
        </el-radio-group>
      </div>
    </el-card>

    <!-- Batch Actions -->
    <div v-if="selectedInstances.length > 0" class="mb-20">
       <el-alert type="info" :closable="false" show-icon>
          <template #title>
             <div class="flex items-center gap-4">
                <span>已选择 {{ selectedInstances.length }} 个实例</span>
                <el-button size="small" type="primary" link @click="handleBatchTest">批量测试</el-button>
                <el-button size="small" type="danger" link @click="handleBatchDelete">批量删除</el-button>
                <el-button size="small" link @click="clearSelection">取消选择</el-button>
             </div>
          </template>
       </el-alert>
    </div>

    <!-- Table View -->
    <el-card v-if="viewMode === 'table'" class="content-card">
       <el-table
          v-loading="loading"
          :data="filteredInstances"
          @selection-change="handleSelectionChange"
          @sort-change="handleSortChange"
          style="width: 100%"
       >
          <el-table-column type="selection" width="55" />
          <el-table-column prop="name" label="名称" min-width="150" sortable="custom">
             <template #default="{ row }">
                <div class="flex items-center gap-2">
                   <el-icon :color="getTypeColor(row.instance_type)" size="18">
                      <component :is="getTypeIcon(row.instance_type)" />
                   </el-icon>
                   <span class="font-medium">{{ row.name }}</span>
                </div>
             </template>
          </el-table-column>
          <el-table-column prop="instance_type" label="类型" width="140" sortable="custom">
              <template #default="{ row }">
                 <el-tag :type="getTypeTagType(row.instance_type)" size="small" effect="plain">{{ getTypeLabel(row.instance_type) }}</el-tag>
              </template>
          </el-table-column>
          <el-table-column prop="address" label="地址" min-width="180" show-overflow-tooltip />
          <el-table-column label="SSL" width="100">
             <template #default="{ row }">
                <el-icon v-if="getHttpsStatus(row.https_enabled, row.skip_ssl_verify)" color="#67C23A"><Lock /></el-icon>
                <el-icon v-else color="#909399"><Unlock /></el-icon>
             </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="120" sortable="custom">
             <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)" size="small">
                   {{ getStatusLabel(row.status) }}
                </el-tag>
             </template>
          </el-table-column>
          <el-table-column prop="updated_at" label="更新时间" width="180" sortable="custom">
             <template #default="{ row }">
                {{ formatTime(row.updated_at) }}
             </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
             <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleTest(row)" :loading="row.testing">测试</el-button>
                <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
             </template>
          </el-table-column>
       </el-table>
        <div class="pagination-wrapper mt-4 flex justify-end">
            <el-pagination
              v-model:current-page="pagination.currentPage"
              v-model:page-size="pagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="pagination.total"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
            />
        </div>
    </el-card>

    <!-- Card View -->
    <div v-else class="grid-view">
       <el-row :gutter="20">
          <el-col :span="6" v-for="instance in filteredInstances" :key="instance.id" class="mb-20">
             <el-card class="content-card instance-card-view h-full" shadow="hover" :body-style="{ padding: '16px', height: '100%', display: 'flex', flexDirection: 'column' }">
                <div class="card-header flex justify-between items-start mb-4">
                   <div class="flex items-center gap-3">
                      <div class="icon-box" :style="{ backgroundColor: getTypeColor(instance.instance_type) + '20', color: getTypeColor(instance.instance_type) }">
                         <el-icon size="20"><component :is="getTypeIcon(instance.instance_type)" /></el-icon>
                      </div>
                      <div>
                         <div class="font-bold text-main">{{ instance.name }}</div>
                         <div class="text-xs text-sub">{{ getTypeLabel(instance.instance_type) }}</div>
                      </div>
                   </div>
                    <el-tag :type="getStatusType(instance.status)" size="small" effect="dark" class="status-tag">
                        {{ getStatusLabel(instance.status) }}
                    </el-tag>
                </div>
                <div class="card-body text-sm text-sub mb-4 flex-1">
                   <div class="mb-2 flex items-center gap-2"><el-icon><Location /></el-icon> {{ instance.address }}</div>
                   <div class="flex items-center gap-2">
                       <el-icon><Clock /></el-icon> {{ formatTime(instance.updated_at) }}
                   </div>
                </div>
                <div class="card-footer flex justify-end gap-2 pt-3 border-t border-gray-100">
                    <el-button circle size="small" @click="handleTest(instance)" :loading="instance.testing" title="测试连接"><el-icon><Connection /></el-icon></el-button>
                    <el-button circle size="small" @click="handleEdit(instance)" title="编辑"><el-icon><Edit /></el-icon></el-button>
                    <el-button circle size="small" type="danger" @click="handleDelete(instance)" title="删除"><el-icon><Delete /></el-icon></el-button>
                </div>
             </el-card>
          </el-col>
       </el-row>
    </div>

  </div>
</template>

<script setup>
import {computed, onMounted, ref} from 'vue'
import {useRouter} from 'vue-router'
import {ElMessage, ElMessageBox} from 'element-plus'
import {
  Clock,
  Connection,
  DataLine,
  Delete,
  DocumentCopy,
  Edit,
  Grid,
  List,
  Location,
  Lock,
  Monitor,
  Platform,
  Plus,
  Refresh,
  RefreshLeft,
  Search,
  Unlock,
  WarningFilled
} from '@element-plus/icons-vue'
import {deleteInstance, getInstanceList, testConnection} from '@/api/instance.js'

const router = useRouter()

// 数据状态
const loading = ref(false)
const viewMode = ref('table')
const searchQuery = ref('')
const typeFilter = ref('')
const statusFilter = ref('')
const searching = ref(false)
const selectedInstances = ref([])

// 分页
const pagination = ref({
  currentPage: 1,
  pageSize: 20,
  total: 0
})

// 排序
const sortConfig = ref({
  prop: '',
  order: ''
})

// 实例数据
const instances = ref([])

// 计算属性
const filteredInstances = computed(() => {
  return instances.value
})

// 类型相关方法
const getTypeIcon = (type) => {
  const icons = {
    elasticsearch: Monitor,
    kibana: DataLine,
    logstash: DocumentCopy,
    filebeat: DocumentCopy,
    metricbeat: DocumentCopy,
    apm: WarningFilled,
    prometheus: DataLine
  }
  return icons[type] || Monitor
}

const getTypeColor = (type) => {
  const colors = {
    elasticsearch: '#1976d2',
    kibana: '#4caf50',
    logstash: '#ff9800',
    filebeat: '#f44336',
    metricbeat: '#9c27b0',
    apm: '#795548',
    prometheus: '#e6522c'
  }
  return colors[type] || '#666'
}

const getTypeLabel = (type) => {
  const labels = {
    elasticsearch: 'Elasticsearch',
    kibana: 'Kibana',
    logstash: 'Logstash',
    filebeat: 'Filebeat',
    metricbeat: 'Metricbeat',
    apm: 'APM',
    prometheus: 'Prometheus'
  }
  return labels[type] || type
}

const getTypeTagType = (type) => {
  const types = {
    elasticsearch: 'primary',
    kibana: 'success',
    logstash: 'warning',
    filebeat: 'danger',
    metricbeat: 'info',
    apm: 'info',
    kubernetes: 'primary',
    prometheus: 'danger'
  }
  return types[type] || 'info'
}

// 状态相关方法
const getStatusLabel = (status) => {
  const labels = { active: '在线', online: '在线', offline: '离线', inactive: '非活跃', error: '错误' }
  return labels[status] || status
}

const getStatusType = (status) => {
  const types = { active: 'success', online: 'success', offline: 'danger', inactive: 'info', error: 'danger' }
  return types[status] || 'info'
}

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return '-'
  return new Date(timeStr).toLocaleString()
}

// 获取HTTPS状态
const getHttpsStatus = (httpsEnabled, skipSslVerify) => {
  if (skipSslVerify === true || skipSslVerify === 1 || skipSslVerify === '1' || skipSslVerify === 'true') {
    return true
  }
  return false
}

// 事件处理
const handleSearch = () => {
  searching.value = true
  pagination.value.currentPage = 1
  fetchInstances().finally(() => {
    searching.value = false
  })
}

const handleResetFilter = () => {
  searchQuery.value = ''
  typeFilter.value = ''
  statusFilter.value = ''
  pagination.value.currentPage = 1
  fetchInstances()
}

// 获取实例数据
const fetchInstances = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.currentPage,
      page_size: pagination.value.pageSize
    }
    if (searchQuery.value) params.keyword = searchQuery.value
    if (typeFilter.value) params.type_name = typeFilter.value
    if (statusFilter.value) params.status = statusFilter.value

    const response = await getInstanceList(params)
    const listData = response.data?.list || {}

    instances.value = listData.data || []
    pagination.value.total = listData.total || 0
  } catch (error) {
    ElMessage.error('数据加载失败: ' + error.message)
    instances.value = []
    pagination.value.total = 0
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  router.push('/es/instances/add')
}

const handleEdit = (instance) => {
  router.push(`/es/instances/edit/${instance.id}`)
}

const handleTest = async (instance) => {
  instance.testing = true
  try {
    const response = await testConnection(instance.id)
    const testData = response.data?.test_result

    if (testData?.test_result === 'success') {
      ElMessage.success(`${instance.name} 连接测试成功 (响应时间: ${testData.response_time}ms)`)
      instance.status = 'active'
    } else {
      const errorMsg = testData?.error_message || '测试失败'
      ElMessage.error(`${instance.name} 连接测试失败: ${errorMsg}`)
      instance.status = testData?.test_result === 'timeout' ? 'warning' : 'error'
    }
    instance.updated_at = new Date().toISOString()
  } catch (error) {
    const errorMsg = error.response?.data?.message || error.message || '连接测试失败'
    ElMessage.error(`${instance.name} 连接测试失败: ${errorMsg}`)
    instance.status = 'error'
  } finally {
    instance.testing = false
  }
}

const handleDelete = async (instance) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除实例 "${instance.name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'warning' }
    )

    const response = await deleteInstance(instance.id)
    if (response.message === '删除成功') {
      ElMessage.success(`实例 "${instance.name}" 删除成功`)
      await fetchInstances()
    } else {
      ElMessage.error(response.data?.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') ElMessage.error('删除失败: ' + error.message)
  }
}

const handleRefresh = async () => {
  searchQuery.value = ''
  typeFilter.value = ''
  statusFilter.value = ''
  pagination.value.currentPage = 1
  await fetchInstances()
}

const handleSelectionChange = (selection) => {
  selectedInstances.value = selection
}

const handleSortChange = ({ prop, order }) => {
  sortConfig.value = { prop, order }
}

const handleSizeChange = (size) => {
  pagination.value.pageSize = size
  pagination.value.currentPage = 1
  fetchInstances()
}

const handleCurrentChange = (page) => {
  pagination.value.currentPage = page
  fetchInstances()
}

const handleBatchTest = async () => {
  ElMessage.info('开始批量测试...')
  let successCount = 0
  let failCount = 0

  for (const instance of selectedInstances.value) {
    try {
      await handleTest(instance)
      successCount++
    } catch (error) {
      failCount++
    }
  }
  ElMessage.success(`批量测试完成：成功 ${successCount} 个，失败 ${failCount} 个`)
}

const handleBatchDelete = async () => {
    try {
        await ElMessageBox.confirm(`确定要删除选中的 ${selectedInstances.value.length} 个实例吗？`, '警告', { type: 'warning' })
        let successCount = 0
        for (const instance of selectedInstances.value) {
             try {
                await deleteInstance(instance.id)
                successCount++
             } catch(e){}
        }
        ElMessage.success(`成功删除 ${successCount} 个实例`)
        fetchInstances()
        selectedInstances.value = []
    } catch(e) {}
}

const clearSelection = () => {
  selectedInstances.value = []
}

onMounted(() => {
  fetchInstances()
})
</script>

<style scoped>
.header-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: var(--primary-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.icon-box {
    width: 36px;
    height: 36px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.text-main { color: var(--text-main); }
.text-sub { color: var(--text-sub); }
.text-xs { font-size: 12px; }
.font-bold { font-weight: 600; }
.font-medium { font-weight: 500; }
.flex { display: flex; }
.items-center { align-items: center; }
.justify-between { justify-content: space-between; }
.justify-end { justify-content: flex-end; }
.items-start { align-items: flex-start; }
.gap-2 { gap: 8px; }
.gap-3 { gap: 12px; }
.gap-4 { gap: 16px; }
.mb-1 { margin-bottom: 4px; }
.mb-2 { margin-bottom: 8px; }
.mb-4 { margin-bottom: 16px; }
.mb-20 { margin-bottom: 20px; }
.mt-1 { margin-top: 4px; }
.mt-4 { margin-top: 16px; }
.mx-4 { margin-left: 16px; margin-right: 16px; }
.p-4 { padding: 16px; }
.pt-3 { padding-top: 12px; }
.border-t { border-top: 1px solid #f0f0f0; }
.w-64 { width: 16rem; }
.w-40 { width: 10rem; }
.w-32 { width: 8rem; }
.h-full { height: 100%; }
.flex-1 { flex: 1; }

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-left { display: flex; align-items: center; gap: 12px; }
.header-title-wrapper h2 { margin: 0; font-size: 18px; color: var(--text-main); }
.header-title-wrapper .subtitle { margin: 0; font-size: 12px; color: var(--text-sub); }
.header-right { display: flex; align-items: center; gap: 12px; }
</style>
