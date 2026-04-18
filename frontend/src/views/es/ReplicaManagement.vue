<template>
  <div class="replica-management">
    <div class="page-header">
      <h2>副本管理</h2>
      <div class="header-actions">
        <el-select
          v-model="selectedIndex"
          placeholder="选择索引"
          style="width: 200px; margin-right: 10px"
          @change="handleIndexChange"
        >
          <el-option
            v-for="index in indexOptions"
            :key="index.value"
            :label="index.label"
            :value="index.value"
          />
        </el-select>
        <el-input
          v-model="searchQuery"
          placeholder="搜索分片ID"
          style="width: 200px; margin-right: 10px"
          @input="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" @click="showIncreaseDialog = true">
          <el-icon><Plus /></el-icon>
          增加副本
        </el-button>
        <el-button @click="refreshData">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div class="stats-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon">
                <el-icon color="#409EFF"><CopyDocument /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ stats.totalReplicas }}</div>
                <div class="stat-label">总副本数</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon">
                <el-icon color="#67C23A"><CircleCheck /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ stats.activeReplicas }}</div>
                <div class="stat-label">活跃副本</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon">
                <el-icon color="#E6A23C"><Warning /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ stats.initializingReplicas }}</div>
                <div class="stat-label">初始化中</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon">
                <el-icon color="#F56C6C"><CircleClose /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ stats.unassignedReplicas }}</div>
                <div class="stat-label">未分配副本</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <el-card class="table-card">
      <el-table
        :data="filteredReplicas"
        v-loading="loading"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="index" label="索引名称" min-width="150" />
        <el-table-column prop="shard" label="分片ID" width="100" />
        <el-table-column prop="replica" label="副本ID" width="100" />
        <el-table-column prop="state" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStateType(row.state)" size="small">
              {{ getStateText(row.state) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="node" label="所在节点" width="150" />
        <el-table-column prop="ip" label="节点IP" width="120" />
        <el-table-column prop="size" label="副本大小" width="120">
          <template #default="{ row }">
            {{ formatBytes(row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="docsCount" label="文档数" width="120">
          <template #default="{ row }">
            {{ row.docsCount.toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="syncLag" label="同步延迟" width="120">
          <template #default="{ row }">
            <span v-if="row.syncLag === 0" style="color: #67C23A">已同步</span>
            <span v-else style="color: #E6A23C">{{ row.syncLag }}s</span>
          </template>
        </el-table-column>
        <el-table-column prop="lastSyncTime" label="最后同步时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.lastSyncTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              link
              @click="viewDetails(row)"
            >
              详情
            </el-button>
            <el-button
              type="warning"
              size="small"
              link
              @click="relocateReplica(row)"
              :disabled="row.state === 'UNASSIGNED'"
            >
              迁移
            </el-button>
            <el-button
              type="danger"
              size="small"
              link
              @click="deleteReplica(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="batch-actions" v-if="selectedReplicas.length > 0">
        <el-button type="warning" @click="batchRelocate">
          批量迁移 ({{ selectedReplicas.length }})
        </el-button>
        <el-button type="danger" @click="batchDelete">
          批量删除 ({{ selectedReplicas.length }})
        </el-button>
      </div>
    </el-card>

    <!-- 增加副本对话框 -->
    <el-dialog
      v-model="showIncreaseDialog"
      title="增加副本"
      width="40%"
    >
      <el-form
        ref="increaseFormRef"
        :model="increaseForm"
        :rules="increaseRules"
        label-width="100px"
      >
        <el-form-item label="索引名称" prop="index">
          <el-select v-model="increaseForm.index" placeholder="选择索引" style="width: 100%">
            <el-option
              v-for="index in indexOptions.slice(1)"
              :key="index.value"
              :label="index.label"
              :value="index.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="副本数量" prop="count">
          <el-input-number v-model="increaseForm.count" :min="1" :max="10" />
        </el-form-item>
        <el-form-item label="分配策略">
          <el-radio-group v-model="increaseForm.allocation">
            <el-radio label="balanced">均衡分配</el-radio>
            <el-radio label="specific">指定节点</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="increaseForm.allocation === 'specific'" label="目标节点">
          <el-select v-model="increaseForm.targetNodes" multiple placeholder="选择节点" style="width: 100%">
            <el-option
              v-for="node in nodeOptions.slice(1)"
              :key="node.value"
              :label="node.label"
              :value="node.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showIncreaseDialog = false">取消</el-button>
        <el-button type="primary" @click="increaseReplicas" :loading="increaseLoading">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 副本详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="副本详情"
      width="60%"
    >
      <div v-if="currentReplica" class="replica-details">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="索引名称">{{ currentReplica.index }}</el-descriptions-item>
          <el-descriptions-item label="分片ID">{{ currentReplica.shard }}</el-descriptions-item>
          <el-descriptions-item label="副本ID">{{ currentReplica.replica }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStateType(currentReplica.state)">
              {{ getStateText(currentReplica.state) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="所在节点">{{ currentReplica.node }}</el-descriptions-item>
          <el-descriptions-item label="节点IP">{{ currentReplica.ip }}</el-descriptions-item>
          <el-descriptions-item label="副本大小">{{ formatBytes(currentReplica.size) }}</el-descriptions-item>
          <el-descriptions-item label="文档数">{{ currentReplica.docsCount.toLocaleString() }}</el-descriptions-item>
          <el-descriptions-item label="同步延迟">
            <span v-if="currentReplica.syncLag === 0" style="color: #67C23A">已同步</span>
            <span v-else style="color: #E6A23C">{{ currentReplica.syncLag }}s</span>
          </el-descriptions-item>
          <el-descriptions-item label="最后同步时间">{{ formatTime(currentReplica.lastSyncTime) }}</el-descriptions-item>
        </el-descriptions>

        <div class="sync-history" style="margin-top: 20px">
          <h4>同步历史</h4>
          <el-table :data="currentReplica.syncHistory" stripe>
            <el-table-column prop="time" label="同步时间" width="160">
              <template #default="{ row }">
                {{ formatTime(row.time) }}
              </template>
            </el-table-column>
            <el-table-column prop="type" label="同步类型" width="120">
              <template #default="{ row }">
                <el-tag :type="getSyncTypeTag(row.type)" size="small">
                  {{ getSyncTypeText(row.type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="docsCount" label="同步文档数">
              <template #default="{ row }">
                {{ row.docsCount.toLocaleString() }}
              </template>
            </el-table-column>
            <el-table-column prop="duration" label="耗时" width="100">
              <template #default="{ row }">
                {{ row.duration }}ms
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
                  {{ row.status === 'success' ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 迁移副本对话框 -->
    <el-dialog
      v-model="relocateDialogVisible"
      title="迁移副本"
      width="40%"
    >
      <el-form :model="relocateForm" label-width="100px">
        <el-form-item label="目标节点">
          <el-select v-model="relocateForm.targetNode" placeholder="选择目标节点" style="width: 100%">
            <el-option
              v-for="node in availableNodes"
              :key="node.value"
              :label="node.label"
              :value="node.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="relocateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmRelocate" :loading="relocateLoading">
          确认迁移
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus, CopyDocument, CircleCheck, Warning, CircleClose } from '@element-plus/icons-vue'

// 响应式数据
const loading = ref(false)
const increaseLoading = ref(false)
const relocateLoading = ref(false)
const searchQuery = ref('')
const selectedIndex = ref('')
const showIncreaseDialog = ref(false)
const detailDialogVisible = ref(false)
const relocateDialogVisible = ref(false)
const currentReplica = ref(null)
const selectedReplicas = ref([])
const increaseFormRef = ref(null)

// 增加副本表单
const increaseForm = reactive({
  index: '',
  count: 1,
  allocation: 'balanced',
  targetNodes: []
})

// 表单验证规则
const increaseRules = {
  index: [
    { required: true, message: '请选择索引', trigger: 'change' }
  ],
  count: [
    { required: true, message: '请输入副本数量', trigger: 'blur' }
  ]
}

// 迁移表单
const relocateForm = reactive({
  targetNode: ''
})

// 统计数据
const stats = reactive({
  totalReplicas: 0,
  activeReplicas: 0,
  initializingReplicas: 0,
  unassignedReplicas: 0
})

// 索引选项
const indexOptions = ref([
  { label: '全部索引', value: '' },
  { label: 'logs-2025-11-14', value: 'logs-2025-11-14' },
  { label: 'metrics-2025-11', value: 'metrics-2025-11' },
  { label: 'users', value: 'users' },
  { label: 'products', value: 'products' },
  { label: 'orders-2025', value: 'orders-2025' }
])

// 节点选项
const nodeOptions = ref([
  { label: '全部节点', value: '' },
  { label: 'node-01', value: 'node-01' },
  { label: 'node-02', value: 'node-02' },
  { label: 'node-03', value: 'node-03' },
  { label: 'node-04', value: 'node-04' },
  { label: 'node-05', value: 'node-05' }
])

// 可用节点列表（用于迁移）
const availableNodes = ref([
  { label: 'node-01 (192.168.1.101)', value: 'node-01' },
  { label: 'node-02 (192.168.1.102)', value: 'node-02' },
  { label: 'node-03 (192.168.1.103)', value: 'node-03' },
  { label: 'node-04 (192.168.1.104)', value: 'node-04' },
  { label: 'node-05 (192.168.1.105)', value: 'node-05' }
])

// 副本列表数据
const replicas = ref([])

// 过滤后的副本列表
const filteredReplicas = computed(() => {
  let filtered = replicas.value

  // 按索引过滤
  if (selectedIndex.value) {
    filtered = filtered.filter(replica => replica.index === selectedIndex.value)
  }

  // 按搜索关键词过滤
  if (searchQuery.value) {
    filtered = filtered.filter(replica => 
      replica.shard.toString().includes(searchQuery.value) ||
      replica.replica.toString().includes(searchQuery.value)
    )
  }

  return filtered
})

// 获取状态类型
const getStateType = (state) => {
  const stateMap = {
    'STARTED': 'success',
    'INITIALIZING': 'warning',
    'RELOCATING': 'warning',
    'UNASSIGNED': 'danger'
  }
  return stateMap[state] || 'info'
}

// 获取状态文本
const getStateText = (state) => {
  const stateMap = {
    'STARTED': '已启动',
    'INITIALIZING': '初始化中',
    'RELOCATING': '迁移中',
    'UNASSIGNED': '未分配'
  }
  return stateMap[state] || state
}

// 获取同步类型标签
const getSyncTypeTag = (type) => {
  const typeMap = {
    'full': 'primary',
    'incremental': 'success',
    'recovery': 'warning'
  }
  return typeMap[type] || 'info'
}

// 获取同步类型文本
const getSyncTypeText = (type) => {
  const typeMap = {
    'full': '全量同步',
    'incremental': '增量同步',
    'recovery': '恢复同步'
  }
  return typeMap[type] || type
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

// 生成模拟数据
const generateMockData = () => {
  const mockReplicas = [
    {
      id: 1,
      index: 'logs-2025-11-14',
      shard: 0,
      replica: 1,
      state: 'STARTED',
      node: 'node-02',
      ip: '192.168.1.102',
      size: 1024 * 1024 * 128, // 128MB
      docsCount: 5078,
      syncLag: 0,
      lastSyncTime: Date.now() - 1000 * 60 * 2,
      syncHistory: [
        {
          time: Date.now() - 1000 * 60 * 2,
          type: 'incremental',
          docsCount: 23,
          duration: 120,
          status: 'success'
        },
        {
          time: Date.now() - 1000 * 60 * 30,
          type: 'incremental',
          docsCount: 45,
          duration: 150,
          status: 'success'
        }
      ]
    },
    {
      id: 2,
      index: 'logs-2025-11-14',
      shard: 1,
      replica: 1,
      state: 'RELOCATING',
      node: 'node-03',
      ip: '192.168.1.103',
      size: 1024 * 1024 * 96,
      docsCount: 3845,
      syncLag: 15,
      lastSyncTime: Date.now() - 1000 * 60 * 5,
      syncHistory: [
        {
          time: Date.now() - 1000 * 60 * 5,
          type: 'recovery',
          docsCount: 3845,
          duration: 3000,
          status: 'success'
        }
      ]
    },
    {
      id: 3,
      index: 'logs-2025-11-14',
      shard: 2,
      replica: 1,
      state: 'UNASSIGNED',
      node: '',
      ip: '',
      size: 0,
      docsCount: 0,
      syncLag: 0,
      lastSyncTime: null,
      syncHistory: []
    },
    {
      id: 4,
      index: 'metrics-2025-11',
      shard: 0,
      replica: 1,
      state: 'STARTED',
      node: 'node-04',
      ip: '192.168.1.104',
      size: 1024 * 1024 * 256,
      docsCount: 9135,
      syncLag: 0,
      lastSyncTime: Date.now() - 1000 * 60 * 1,
      syncHistory: [
        {
          time: Date.now() - 1000 * 60 * 1,
          type: 'incremental',
          docsCount: 112,
          duration: 200,
          status: 'success'
        }
      ]
    },
    {
      id: 5,
      index: 'users',
      shard: 0,
      replica: 1,
      state: 'STARTED',
      node: 'node-05',
      ip: '192.168.1.105',
      size: 1024 * 1024 * 64,
      docsCount: 8901,
      syncLag: 0,
      lastSyncTime: Date.now() - 1000 * 60 * 3,
      syncHistory: [
        {
          time: Date.now() - 1000 * 60 * 3,
          type: 'incremental',
          docsCount: 45,
          duration: 180,
          status: 'success'
        }
      ]
    },
    {
      id: 6,
      index: 'products',
      shard: 0,
      replica: 1,
      state: 'INITIALIZING',
      node: 'node-01',
      ip: '192.168.1.101',
      size: 1024 * 1024 * 32,
      docsCount: 11728,
      syncLag: 30,
      lastSyncTime: Date.now() - 1000 * 60 * 10,
      syncHistory: [
        {
          time: Date.now() - 1000 * 60 * 10,
          type: 'full',
          docsCount: 11728,
          duration: 5000,
          status: 'success'
        }
      ]
    },
    {
      id: 7,
      index: 'orders-2025',
      shard: 0,
      replica: 1,
      state: 'STARTED',
      node: 'node-02',
      ip: '192.168.1.102',
      size: 1024 * 1024 * 384,
      docsCount: 34567,
      syncLag: 0,
      lastSyncTime: Date.now() - 1000 * 60 * 1,
      syncHistory: [
        {
          time: Date.now() - 1000 * 60 * 1,
          type: 'incremental',
          docsCount: 89,
          duration: 250,
          status: 'success'
        }
      ]
    },
    {
      id: 8,
      index: 'orders-2025',
      shard: 0,
      replica: 2,
      state: 'STARTED',
      node: 'node-03',
      ip: '192.168.1.103',
      size: 1024 * 1024 * 384,
      docsCount: 34567,
      syncLag: 0,
      lastSyncTime: Date.now() - 1000 * 60 * 2,
      syncHistory: [
        {
          time: Date.now() - 1000 * 60 * 2,
          type: 'incremental',
          docsCount: 89,
          duration: 230,
          status: 'success'
        }
      ]
    }
  ]

  replicas.value = mockReplicas
  
  // 更新统计数据
  stats.totalReplicas = mockReplicas.length
  stats.activeReplicas = mockReplicas.filter(replica => replica.state === 'STARTED').length
  stats.initializingReplicas = mockReplicas.filter(replica => replica.state === 'INITIALIZING').length
  stats.unassignedReplicas = mockReplicas.filter(replica => replica.state === 'UNASSIGNED').length
}

// 刷新数据
const refreshData = async () => {
  loading.value = true
  try {
    // 模拟API调用延迟
    await new Promise(resolve => setTimeout(resolve, 1000))
    generateMockData()
    ElMessage.success('数据刷新成功')
  } catch (error) {
    ElMessage.error('数据刷新失败：' + error.message)
  } finally {
    loading.value = false
  }
}

// 索引变化处理
const handleIndexChange = () => {
  // 过滤逻辑已在computed中实现
}

// 搜索处理
const handleSearch = () => {
  // 搜索逻辑已在computed中实现
}

// 选择变化处理
const handleSelectionChange = (selection) => {
  selectedReplicas.value = selection
}

// 查看详情
const viewDetails = (replica) => {
  currentReplica.value = replica
  detailDialogVisible.value = true
}

// 增加副本
const increaseReplicas = async () => {
  if (!increaseFormRef.value) return
  
  try {
    await increaseFormRef.value.validate()
    increaseLoading.value = true
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 3000))
    
    ElMessage.success('副本增加成功')
    
    // 重置表单和关闭对话框
    Object.assign(increaseForm, {
      index: '',
      count: 1,
      allocation: 'balanced',
      targetNodes: []
    })
    showIncreaseDialog.value = false
    refreshData()
  } catch (error) {
    if (error.message) {
      ElMessage.error('增加副本失败：' + error.message)
    }
  } finally {
    increaseLoading.value = false
  }
}

// 迁移副本
const relocateReplica = (replica) => {
  currentReplica.value = replica
  relocateForm.targetNode = ''
  relocateDialogVisible.value = true
}

// 确认迁移
const confirmRelocate = async () => {
  if (!relocateForm.targetNode) {
    ElMessage.warning('请选择目标节点')
    return
  }

  if (relocateForm.targetNode === currentReplica.value.node) {
    ElMessage.warning('目标节点不能与当前节点相同')
    return
  }

  try {
    relocateLoading.value = true
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 3000))
    
    ElMessage.success(`副本迁移成功`)
    relocateDialogVisible.value = false
    refreshData()
  } catch (error) {
    ElMessage.error('迁移失败：' + error.message)
  } finally {
    relocateLoading.value = false
  }
}

// 删除副本
const deleteReplica = (replica) => {
  ElMessageBox.confirm(
    `确定要删除副本 ${replica.index}[分片${replica.shard}-副本${replica.replica}] 吗？此操作可能会影响数据冗余。`,
    '确认删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'danger',
    }
  ).then(async () => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 2000))
      
      const indexToDelete = replicas.value.findIndex(r => r.id === replica.id)
      if (indexToDelete > -1) {
        replicas.value.splice(indexToDelete, 1)
      }
      
      ElMessage.success('副本删除成功')
      refreshData()
    } catch (error) {
      ElMessage.error('删除失败：' + error.message)
    }
  }).catch(() => {
    ElMessage.info('已取消删除操作')
  })
}

// 批量迁移
const batchRelocate = () => {
  ElMessageBox.confirm(
    `确定要批量迁移选中的 ${selectedReplicas.value.length} 个副本吗？`,
    '确认批量迁移',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 5000))
      ElMessage.success('批量迁移成功')
      selectedReplicas.value = []
      refreshData()
    } catch (error) {
      ElMessage.error('批量迁移失败：' + error.message)
    }
  }).catch(() => {
    ElMessage.info('已取消批量迁移操作')
  })
}

// 批量删除
const batchDelete = () => {
  ElMessageBox.confirm(
    `确定要批量删除选中的 ${selectedReplicas.value.length} 个副本吗？此操作可能会影响数据冗余。`,
    '确认批量删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'danger',
    }
  ).then(async () => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 3000))
      
      const idsToDelete = selectedReplicas.value.map(replica => replica.id)
      replicas.value = replicas.value.filter(replica => !idsToDelete.includes(replica.id))
      selectedReplicas.value = []
      
      ElMessage.success('批量删除成功')
      refreshData()
    } catch (error) {
      ElMessage.error('批量删除失败：' + error.message)
    }
  }).catch(() => {
    ElMessage.info('已取消批量删除操作')
  })
}

// 组件挂载时加载数据
onMounted(() => {
  refreshData()
})
</script>

<style scoped>
.replica-management {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  color: #303133;
}

.header-actions {
  display: flex;
  align-items: center;
}

.stats-cards {
  margin-bottom: 20px;
}

.stat-card {
  height: 100px;
}

.stat-content {
  display: flex;
  align-items: center;
  height: 100%;
}

.stat-icon {
  font-size: 40px;
  margin-right: 15px;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.table-card {
  margin-bottom: 20px;
}

.batch-actions {
  margin-top: 20px;
  padding: 15px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.replica-details {
  padding: 20px 0;
}

.sync-history h4 {
  margin-bottom: 10px;
  color: #303133;
}
</style>