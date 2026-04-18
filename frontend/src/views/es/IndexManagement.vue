<template>
  <div class="page-container">
    <!-- Header -->
    <el-card class="page-header-card">
       <div class="page-header">
         <div class="header-left">
           <div class="header-icon-wrapper">
             <el-icon :size="24"><Collection /></el-icon>
           </div>
           <div class="header-title-wrapper">
             <h2>索引管理</h2>
             <p class="subtitle">创建、监控与维护 Elasticsearch 索引</p>
           </div>
         </div>
         <div class="header-right">
            <el-input
              v-model="searchQuery"
              placeholder="搜索索引名称"
              class="w-64 mr-2"
              @input="handleSearch"
              :prefix-icon="Search"
            />
            <el-button type="primary" @click="showCreateDialog = true" :icon="Plus">创建索引</el-button>
            <el-button @click="refreshData" :icon="Refresh">刷新</el-button>
         </div>
       </div>
    </el-card>

    <!-- Stats Cards -->
    <el-row :gutter="20" class="mb-20">
       <el-col :span="6">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-content">
               <div class="stat-icon bg-gradient-blue"><el-icon><DocumentCopy /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ stats.totalIndices }}</div>
                  <div class="stat-label">总索引数</div>
               </div>
            </div>
          </el-card>
       </el-col>
       <el-col :span="6">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-content">
               <div class="stat-icon bg-gradient-green"><el-icon><Files /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ formatBytes(stats.totalSize) }}</div>
                  <div class="stat-label">总存储大小</div>
               </div>
            </div>
          </el-card>
       </el-col>
       <el-col :span="6">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-content">
               <div class="stat-icon bg-gradient-orange"><el-icon><Tickets /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ stats.totalDocs.toLocaleString() }}</div>
                  <div class="stat-label">总文档数</div>
               </div>
            </div>
          </el-card>
       </el-col>
       <el-col :span="6">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-content">
               <div class="stat-icon bg-gradient-red"><el-icon><Delete /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ stats.deletedDocs.toLocaleString() }}</div>
                  <div class="stat-label">已删除文档</div>
               </div>
            </div>
          </el-card>
       </el-col>
    </el-row>

    <!-- Table Content -->
    <el-card class="content-card">
       <el-table
        :data="filteredIndices"
        v-loading="loading"
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="index" label="索引名称" min-width="200">
          <template #default="{ row }">
            <div class="index-name">
              <el-tag :type="getStatusType(row.status)" size="small" effect="plain" class="mr-2">
                {{ row.status }}
              </el-tag>
              <span class="font-medium">{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="health" label="健康状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getHealthType(row.health)" size="small" effect="dark">
              {{ row.health }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="文档数" width="120">
          <template #default="{ row }">
            {{ row.docsCount.toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column label="已删除" width="120">
          <template #default="{ row }">
            {{ row.deletedDocs.toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column label="存储大小" width="120">
          <template #default="{ row }">
            {{ formatBytes(row.storeSize) }}
          </template>
        </el-table-column>
        <el-table-column label="主分片" width="100">
          <template #default="{ row }">
            {{ row.primaryShards }}
          </template>
        </el-table-column>
        <el-table-column label="副本分片" width="100">
          <template #default="{ row }">
            {{ row.replicaShards }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetails(row)">详情</el-button>
            <el-button type="warning" link size="small" @click="optimizeIndex(row)">优化</el-button>
            <el-button type="danger" link size="small" @click="deleteIndex(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="batch-actions mt-4 p-4 bg-gray-50 rounded" v-if="selectedIndices.length > 0">
        <el-button type="danger" @click="batchDelete">
           <el-icon class="mr-1"><Delete /></el-icon> 批量删除 ({{ selectedIndices.length }})
        </el-button>
        <el-button type="warning" @click="batchOptimize">
           <el-icon class="mr-1"><MagicStick /></el-icon> 批量优化 ({{ selectedIndices.length }})
        </el-button>
      </div>
    </el-card>

    <!-- Dialogs -->
    <el-dialog v-model="showCreateDialog" title="创建索引" width="500px">
       <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-position="top">
          <el-form-item label="索引名称" prop="name">
             <el-input v-model="createForm.name" placeholder="请输入索引名称" />
          </el-form-item>
          <el-row :gutter="20">
             <el-col :span="12">
                <el-form-item label="主分片数" prop="primaryShards">
                   <el-input-number v-model="createForm.primaryShards" :min="1" :max="100" class="w-full" />
                </el-form-item>
             </el-col>
             <el-col :span="12">
                <el-form-item label="副本分片数" prop="replicaShards">
                   <el-input-number v-model="createForm.replicaShards" :min="0" :max="10" class="w-full" />
                </el-form-item>
             </el-col>
          </el-row>
          <el-form-item label="刷新间隔">
             <el-input v-model="createForm.refreshInterval" placeholder="如: 1s" />
          </el-form-item>
          <el-form-item label="别名">
             <el-input v-model="createForm.alias" placeholder="可选的索引别名" />
          </el-form-item>
       </el-form>
       <template #footer>
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="createIndex" :loading="createLoading">创建</el-button>
       </template>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" title="索引详情" width="800px">
       <div v-if="currentIndex" class="index-details">
          <el-descriptions :column="2" border>
             <el-descriptions-item label="索引名称">{{ currentIndex.name }}</el-descriptions-item>
             <el-descriptions-item label="状态">
                <el-tag :type="getStatusType(currentIndex.status)">{{ currentIndex.status }}</el-tag>
             </el-descriptions-item>
             <el-descriptions-item label="健康状态">
                <el-tag :type="getHealthType(currentIndex.health)">{{ currentIndex.health }}</el-tag>
             </el-descriptions-item>
             <el-descriptions-item label="UUID">{{ currentIndex.uuid }}</el-descriptions-item>
             <el-descriptions-item label="文档数">{{ currentIndex.docsCount.toLocaleString() }}</el-descriptions-item>
             <el-descriptions-item label="已删除文档">{{ currentIndex.deletedDocs.toLocaleString() }}</el-descriptions-item>
             <el-descriptions-item label="存储大小">{{ formatBytes(currentIndex.storeSize) }}</el-descriptions-item>
             <el-descriptions-item label="主/副分片">{{ currentIndex.primaryShards }} / {{ currentIndex.replicaShards }}</el-descriptions-item>
          </el-descriptions>

          <el-tabs v-model="activeDetailTab" class="mt-4">
             <el-tab-pane label="映射配置" name="mapping">
                <el-input
                  v-model="currentIndex.mapping"
                  type="textarea"
                  :rows="12"
                  readonly
                  class="code-editor"
                />
             </el-tab-pane>
             <el-tab-pane label="索引设置" name="settings" v-if="currentIndex.settings">
                <el-input
                  v-model="currentIndex.settings"
                  type="textarea"
                  :rows="12"
                  readonly
                  class="code-editor"
                />
             </el-tab-pane>
          </el-tabs>
       </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Search, Refresh, Plus, DocumentCopy, Files, Tickets, Delete, 
  Collection, MagicStick 
} from '@element-plus/icons-vue'
import { 
  getIndices, 
  createIndex as createIndexAPI, 
  deleteIndex as deleteIndexAPI,
  updateIndex,
  getIndexDetail,
  getIndexMapping,
  getIndexSettings,
  optimizeIndex as optimizeIndexAPI
} from '@/api/es/indices.js'
import { getSelectedInstanceId } from '@/stores/instanceStore'

// 响应式数据
const loading = ref(false)
const createLoading = ref(false)
const searchQuery = ref('')
const showCreateDialog = ref(false)
const detailDialogVisible = ref(false)
const currentIndex = ref(null)
const selectedIndices = ref([])
const createFormRef = ref(null)
const activeDetailTab = ref('mapping')

// 统计数据
const stats = reactive({
  totalIndices: 0,
  totalSize: 0,
  totalDocs: 0,
  deletedDocs: 0
})

// 创建表单
const createForm = reactive({
  name: '',
  primaryShards: 1,
  replicaShards: 1,
  mapping: '',
  alias: '',
  refreshInterval: '1s',
  numberOfShards: 1,
  numberOfReplicas: 1
})

// 表单验证规则
const createRules = {
  name: [
    { required: true, message: '请输入索引名称', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9_]*$/, message: '索引名称只能包含小写字母、数字和下划线，且必须以字母开头', trigger: 'blur' }
  ],
  primaryShards: [
    { required: true, message: '请输入主分片数', trigger: 'blur' }
  ],
  replicaShards: [
    { required: true, message: '请输入副本分片数', trigger: 'blur' }
  ]
}

// 索引列表数据
const indices = ref([])

// 过滤后的索引列表
const filteredIndices = computed(() => {
  if (!searchQuery.value) return indices.value
  return indices.value.filter(index => 
    index.index.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

// 获取状态类型
const getStatusType = (status) => {
  const statusMap = {
    'open': 'success',
    'close': 'info'
  }
  return statusMap[status] || 'info'
}

// 获取健康状态类型
const getHealthType = (health) => {
  const healthMap = {
    'green': 'success',
    'yellow': 'warning',
    'red': 'danger'
  }
  return healthMap[health] || 'info'
}

// 格式化时间
const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString()
}

// 格式化字节
const formatBytes = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 解析索引数据
const parseIndexData = (rawData) => {
  if (!rawData || !Array.isArray(rawData)) {
    return []
  }
  
  return rawData.map((item, index) => ({
    id: index + 1,
    name: item.index,
    status: item.status || 'open',
    health: item.health || 'green',
    uuid: item.uuid || '',
    docsCount: parseInt(item['docs.count']) || 0,
    deletedDocs: parseInt(item['docs.deleted']) || 0,
    storeSize: parseSize(item['store.size']),
    primaryShards: parseInt(item['pri']) || 1,
    replicaShards: parseInt(item['rep']) || 0,
    createdAt: item['creation.date'] || Date.now(),
    mapping: ''
  }))
}

// 解析存储大小
const parseSize = (sizeStr) => {
  if (!sizeStr) return 0
  const match = sizeStr.match(/^([\d.]+)([bkmgtp]b?)$/i)
  if (!match) return 0
  
  const value = parseFloat(match[1])
  const unit = match[2].toLowerCase()
  const units = { b: 1, kb: 1024, mb: 1024 * 1024, gb: 1024 * 1024 * 1024, tb: 1024 * 1024 * 1024 * 1024 }
  
  return value * (units[unit] || 1)
}

// 更新统计数据
const updateStats = () => {
  stats.totalIndices = indices.value.length
  stats.totalSize = indices.value.reduce((sum, index) => sum + index.storeSize, 0)
  stats.totalDocs = indices.value.reduce((sum, index) => sum + index.docsCount, 0)
  stats.deletedDocs = indices.value.reduce((sum, index) => sum + index.deletedDocs, 0)
}

// 刷新数据
const refreshData = async () => {
  loading.value = true
  try {
    const instanceId = getSelectedInstanceId()
    if (!instanceId) {
      ElMessage.warning('请先选择实例')
      return
    }
    
    const response = await getIndices(instanceId)
    const parsedData = parseIndexData(response.data.raw_response)
    indices.value = parsedData
    updateStats()
    ElMessage.success('数据刷新成功')
  } catch (error) {
    ElMessage.error('数据刷新失败：' + error.message)
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  // 搜索逻辑已在computed中实现
}

// 选择变化处理
const handleSelectionChange = (selection) => {
  selectedIndices.value = selection
}

// 查看详情
const viewDetails = async (index) => {
  try {
    const instanceId = getSelectedInstanceId()
    if (!instanceId) {
      ElMessage.warning('请先选择实例')
      return
    }
    
    // 设置加载状态
    currentIndex.value = { ...index, mapping: '加载中...' }
    detailDialogVisible.value = true
    
    // 并行获取索引映射和设置信息
    const [mappingResponse, settingsResponse] = await Promise.all([
      getIndexMapping(index.name, instanceId).catch(err => {
        console.warn('获取索引映射失败:', err)
        return { data: { raw_response: { properties: {} } } }
      }),
      getIndexSettings(index.name, instanceId).catch(err => {
        console.warn('获取索引设置失败:', err)
        return { data: { raw_response: {} } }
      })
    ])
    
    // 提取映射信息
    const mappings = mappingResponse.data.raw_response
    const indexMapping = mappings[index.name]?.mappings || mappings.mappings || {}
    
    // 提取设置信息
    const settings = settingsResponse.data.raw_response
    const indexSettings = settings[index.name]?.settings?.index || settings.settings?.index || {}
    
    // 更新索引详情
    currentIndex.value = {
      ...index,
      mapping: JSON.stringify(indexMapping, null, 2),
      settings: JSON.stringify(indexSettings, null, 2),
      numberOfShards: indexSettings.number_of_shards || index.primaryShards,
      numberOfReplicas: indexSettings.number_of_replicas || index.replicaShards,
      refreshInterval: indexSettings.refresh_interval || '1s'
    }
  } catch (error) {
    ElMessage.error('获取索引详情失败：' + error.message)
    // 如果出错，显示基本信息
    if (currentIndex.value) {
      currentIndex.value.mapping = '获取映射信息失败'
      currentIndex.value.settings = '获取设置信息失败'
    }
  }
}

// 创建索引
const createIndex = async () => {
  if (!createFormRef.value) return
  
  try {
    await createFormRef.value.validate()
    createLoading.value = true
    
    const instanceId = getSelectedInstanceId()
    if (!instanceId) {
      ElMessage.warning('请先选择实例')
      return
    }
    
    const indexData = {
      instance_id: parseInt(instanceId),
      index_name: createForm.name,
      number_of_shards: createForm.primaryShards,
      number_of_replicas: createForm.replicaShards,
      refresh_interval: createForm.refreshInterval,
      alias: createForm.alias,
      analysis: {}
    }
    
    await createIndexAPI(indexData)
    ElMessage.success('索引创建成功')
    
    // 重置表单和关闭对话框
    Object.assign(createForm, {
      name: '',
      primaryShards: 1,
      replicaShards: 1,
      mapping: '',
      alias: '',
      refreshInterval: '1s'
    })
    showCreateDialog.value = false
    refreshData()
  } catch (error) {
    ElMessage.error('创建失败：' + error.message)
  } finally {
    createLoading.value = false
  }
}

// 优化索引
const optimizeIndex = (index) => {
  ElMessageBox.confirm(
    `确定要优化索引 "${index.name}" 吗？此操作可能会消耗较多资源。`,
    '确认优化',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      const instanceId = getSelectedInstanceId()
      if (!instanceId) {
        ElMessage.warning('请先选择实例')
        return
      }
      
      await optimizeIndexAPI(index.name, instanceId)
      ElMessage.success(`索引 "${index.name}" 优化成功`)
      refreshData()
    } catch (error) {
      ElMessage.error('优化失败：' + error.message)
    }
  }).catch(() => {
    ElMessage.info('已取消优化操作')
  })
}

// 删除索引
const deleteIndex = (index) => {
  ElMessageBox.confirm(
    `确定要删除索引 "${index.name}" 吗？此操作不可恢复。`,
    '确认删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'danger',
    }
  ).then(async () => {
    try {
      const instanceId = getSelectedInstanceId()
      if (!instanceId) {
        ElMessage.warning('请先选择实例')
        return
      }
      
      await deleteIndexAPI({
        instance_id: parseInt(instanceId),
        index_names: [index.name]
      })
      ElMessage.success(`索引 "${index.name}" 删除成功`)
      refreshData()
    } catch (error) {
      ElMessage.error('删除失败：' + error.message)
    }
  }).catch(() => {
    ElMessage.info('已取消删除操作')
  })
}

// 批量删除
const batchDelete = () => {
  const names = selectedIndices.value.map(index => index.name).join(', ')
  ElMessageBox.confirm(
    `确定要删除以下索引吗？此操作不可恢复：\n${names}`,
    '确认批量删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'danger',
    }
  ).then(async () => {
    try {
      const instanceId = getSelectedInstanceId()
      if (!instanceId) {
        ElMessage.warning('请先选择实例')
        return
      }
      
      const indexNames = selectedIndices.value.map(index => index.name)
      await deleteIndexAPI({
        instance_id: parseInt(instanceId),
        index_names: indexNames
      })
      
      selectedIndices.value = []
      ElMessage.success('批量删除成功')
      refreshData()
    } catch (error) {
      ElMessage.error('批量删除失败：' + error.message)
    }
  }).catch(() => {
    ElMessage.info('已取消批量删除操作')
  })
}

// 批量优化
const batchOptimize = () => {
  const names = selectedIndices.value.map(index => index.name).join(', ')
  ElMessageBox.confirm(
    `确定要优化以下索引吗？此操作可能会消耗较多资源：\n${names}`,
    '确认批量优化',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      const instanceId = getSelectedInstanceId()
      if (!instanceId) {
        ElMessage.warning('请先选择实例')
        return
      }
      
      // 逐个优化索引
      for (const index of selectedIndices.value) {
        await optimizeIndexAPI(index.name, instanceId)
      }
      
      ElMessage.success('批量优化成功')
      refreshData()
    } catch (error) {
      ElMessage.error('批量优化失败：' + error.message)
    }
  }).catch(() => {
    ElMessage.info('已取消批量优化操作')
  })
}

// 组件挂载时加载数据
onMounted(() => {
  refreshData()
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

.mr-2 { margin-right: 8px; }
.mr-1 { margin-right: 4px; }
.mt-4 { margin-top: 16px; }
.w-64 { width: 16rem; }
.w-full { width: 100%; }

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-left { display: flex; align-items: center; gap: 12px; }
.header-title-wrapper h2 { margin: 0; font-size: 18px; color: var(--text-main); }
.header-title-wrapper .subtitle { margin: 0; font-size: 12px; color: var(--text-sub); }
.header-right { display: flex; align-items: center; gap: 12px; }

.index-name {
  display: flex;
  align-items: center;
}

.batch-actions {
  display: flex;
  gap: 12px;
}

.code-editor :deep(.el-textarea__inner) {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  background-color: #f8f9fa;
  color: #333;
}
</style>