<template>
  <div class="shard-management">
    <div class="page-header">
      <h2>分片管理 - {{ getSelectedInstanceName() }}</h2>
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
        <el-select
          v-model="selectedNode"
          placeholder="选择节点"
          style="width: 200px; margin-right: 10px"
          @change="handleNodeChange"
        >
          <el-option
            v-for="node in nodeOptions"
            :key="node.value"
            :label="node.label"
            :value="node.value"
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
                <el-icon color="#409EFF"><Grid /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ stats.totalShards }}</div>
                <div class="stat-label">总分片数</div>
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
                <div class="stat-value">{{ stats.primaryShards }}</div>
                <div class="stat-label">主分片</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon">
                <el-icon color="#E6A23C"><CopyDocument /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ stats.replicaShards }}</div>
                <div class="stat-label">副本分片</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon">
                <el-icon color="#F56C6C"><Warning /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-value">{{ stats.unassignedShards }}</div>
                <div class="stat-label">未分配分片</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <el-card class="table-card">
      <el-table
        :data="filteredShards"
        v-loading="loading"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="index" label="索引名称" min-width="150" />
        <el-table-column prop="shard" label="分片ID" width="100" />
        <el-table-column prop="type" label="分片类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'primary' ? 'primary' : 'success'" size="small">
              {{ row.type === 'primary' ? '主分片' : '副本分片' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="state" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStateType(row.state)" size="small">
              {{ getStateText(row.state) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="node" label="所在节点" width="150" />
        <el-table-column prop="ip" label="节点IP" width="120" />
        <el-table-column prop="size" label="分片大小" width="120">
          <template #default="{ row }">
            {{ formatBytes(row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="docsCount" label="文档数" width="120">
          <template #default="{ row }">
            {{ row.docsCount.toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="relocating" label="是否迁移中" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.relocating" type="warning" size="small">迁移中</el-tag>
            <span v-else>-</span>
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
              type="success"
              size="small"
              link
              @click="migrateToInstance(row)"
            >
              跨实例迁移
            </el-button>
            <el-button
              type="danger"
              size="small"
              link
              @click="forceAllocate(row)"
              :disabled="row.state !== 'UNASSIGNED'"
            >
              强制分配
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="batch-actions" v-if="selectedShards.length > 0">
        <el-button type="warning" @click="batchRelocate">
          批量迁移 ({{ selectedShards.length }})
        </el-button>
        <el-button type="danger" @click="batchForceAllocate">
          批量强制分配 ({{ getUnassignedCount() }})
        </el-button>
      </div>
    </el-card>

    <!-- 分片详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="分片详情"
      width="60%"
    >
      <div v-if="currentShard" class="shard-details">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="索引名称">{{ currentShard.index }}</el-descriptions-item>
          <el-descriptions-item label="分片ID">{{ currentShard.shard }}</el-descriptions-item>
          <el-descriptions-item label="分片类型">
            <el-tag :type="currentShard.type === 'primary' ? 'primary' : 'success'">
              {{ currentShard.type === 'primary' ? '主分片' : '副本分片' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStateType(currentShard.state)">
              {{ getStateText(currentShard.state) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="所在节点">{{ currentShard.node }}</el-descriptions-item>
          <el-descriptions-item label="节点IP">{{ currentShard.ip }}</el-descriptions-item>
          <el-descriptions-item label="分片大小">{{ formatBytes(currentShard.size) }}</el-descriptions-item>
          <el-descriptions-item label="文档数">{{ currentShard.docsCount.toLocaleString() }}</el-descriptions-item>
          <el-descriptions-item label="是否迁移中">
            {{ currentShard.relocating ? '是' : '否' }}
          </el-descriptions-item>
          <el-descriptions-item label="分配原因">{{ currentShard.allocationReason }}</el-descriptions-item>
        </el-descriptions>

        <div class="shard-segments" style="margin-top: 20px">
          <h4>段信息</h4>
          <el-table :data="currentShard.segments" stripe>
            <el-table-column prop="name" label="段名称" />
            <el-table-column prop="size" label="段大小">
              <template #default="{ row }">
                {{ formatBytes(row.size) }}
              </template>
            </el-table-column>
            <el-table-column prop="docsCount" label="文档数">
              <template #default="{ row }">
                {{ row.docsCount.toLocaleString() }}
              </template>
            </el-table-column>
            <el-table-column prop="deletedDocs" label="已删除文档">
              <template #default="{ row }">
                {{ row.deletedDocs.toLocaleString() }}
              </template>
            </el-table-column>
            <el-table-column prop="committed" label="是否提交">
              <template #default="{ row }">
                <el-tag :type="row.committed ? 'success' : 'warning'" size="small">
                  {{ row.committed ? '已提交' : '未提交' }}
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

    <!-- 迁移分片对话框 -->
    <el-dialog
      v-model="relocateDialogVisible"
      title="迁移分片"
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

    <!-- 跨实例迁移对话框 -->
    <el-dialog
      v-model="migrateInstanceDialogVisible"
      title="跨实例迁移索引"
      width="50%"
    >
      <el-form :model="migrateInstanceForm" label-width="120px">
        <el-form-item label="当前实例">
          <el-input :value="getSelectedInstanceName()" disabled />
        </el-form-item>
        <el-form-item label="索引名称">
          <el-input :value="currentShard?.index" disabled />
        </el-form-item>
        <el-form-item label="目标实例" required>
          <el-select 
            v-model="migrateInstanceForm.targetInstanceId" 
            placeholder="选择目标实例" 
            style="width: 100%"
            @change="handleTargetInstanceChange"
          >
            <el-option
              v-for="instance in availableInstances"
              :key="instance.id"
              :label="`${instance.name} (${instance.address})`"
              :value="instance.id"
              :disabled="instance.id === getSelectedInstanceId()"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="目标实例状态">
          <el-tag v-if="targetInstanceStatus === 'checking'" type="warning">检查中...</el-tag>
          <el-tag v-else-if="targetInstanceStatus === 'available'" type="success">可用</el-tag>
          <el-tag v-else-if="targetInstanceStatus === 'unavailable'" type="danger">不可用</el-tag>
          <el-tag v-else type="info">未选择</el-tag>
        </el-form-item>
        <el-alert
          title="注意"
          type="warning"
          :closable="false"
          style="margin-top: 10px"
        >
          跨实例迁移会将整个索引（包括所有分片）迁移到目标实例，此操作可能需要较长时间，请确保网络连接稳定。
        </el-alert>
      </el-form>
      <template #footer>
        <el-button @click="migrateInstanceDialogVisible = false">取消</el-button>
        <el-button 
          type="primary" 
          @click="confirmMigrateToInstance" 
          :loading="migrateInstanceLoading"
          :disabled="!migrateInstanceForm.targetInstanceId || targetInstanceStatus !== 'available'"
        >
          确认迁移
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Grid, CircleCheck, CopyDocument, Warning } from '@element-plus/icons-vue'
import { getShardInfo, getShardStats, allocateShard, moveShard, migrateShardToInstance } from '@/api/es/shard.js'
import { getSelectedInstanceId, getSelectedInstanceName } from '@/stores/instanceStore'
import { getInstanceList } from '@/api/instance.js'

// 响应式数据
const loading = ref(false)
const relocateLoading = ref(false)
const migrateInstanceLoading = ref(false)
const searchQuery = ref('')
const selectedIndex = ref('')
const selectedNode = ref('')
const detailDialogVisible = ref(false)
const relocateDialogVisible = ref(false)
const migrateInstanceDialogVisible = ref(false)
const currentShard = ref(null)
const selectedShards = ref([])

// 迁移表单
const relocateForm = reactive({
  targetNode: ''
})

// 跨实例迁移表单
const migrateInstanceForm = reactive({
  targetInstanceId: null
})

// 可用实例列表
const availableInstances = ref([])

// 目标实例状态
const targetInstanceStatus = ref('') // 'checking', 'available', 'unavailable', ''

// 统计数据
const stats = reactive({
  totalShards: 0,
  primaryShards: 0,
  replicaShards: 0,
  unassignedShards: 0
})

// 索引选项
const indexOptions = ref([
  { label: '全部索引', value: '' }
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

// 分片列表数据
const shards = ref([])

// 过滤后的分片列表
const filteredShards = computed(() => {
  let filtered = shards.value

  // 按索引过滤
  if (selectedIndex.value) {
    filtered = filtered.filter(shard => shard.index === selectedIndex.value)
  }

  // 按节点过滤
  if (selectedNode.value) {
    filtered = filtered.filter(shard => shard.node === selectedNode.value)
  }

  // 按搜索关键词过滤
  if (searchQuery.value) {
    filtered = filtered.filter(shard => 
      shard.shard.toString().includes(searchQuery.value)
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

// 格式化字节
const formatBytes = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 获取未分配分片数量
const getUnassignedCount = () => {
  return selectedShards.value.filter(shard => shard.state === 'UNASSIGNED').length
}

// 转换分片数据格式
const transformShardData = (shardData) => {
  return shardData.map((shard, index) => ({
    id: index + 1,
    index: shard.index,
    shard: parseInt(shard.shard),
    type: shard.prirep === 'p' ? 'primary' : 'replica',
    state: shard.state,
    node: shard.node || '',
    ip: shard.ip || '',
    size: parseSize(shard.store),
    docsCount: parseInt(shard.docs) || 0,
    relocating: shard.state === 'RELOCATING',
    allocationReason: 'existing_allocation',
    segments: [] // 段信息需要额外的API调用
  }))
}

// 解析大小字符串为字节数
const parseSize = (sizeStr) => {
  if (!sizeStr) return 0
  const units = { 'b': 1, 'kb': 1024, 'mb': 1024 * 1024, 'gb': 1024 * 1024 * 1024, 'tb': 1024 * 1024 * 1024 * 1024 }
  const match = sizeStr.toLowerCase().match(/^(\d+(?:\.\d+)?)\s*(b|kb|mb|gb|tb)$/)
  if (!match) return 0
  return parseFloat(match[1]) * (units[match[2]] || 1)
}

// 更新统计数据
const updateStats = (statsData) => {
  stats.totalShards = statsData.total_shards || 0
  stats.activeShards = statsData.active_shards || 0
  
  // 计算主分片和副本分片数量
  let primaryCount = 0
  let replicaCount = 0
  let unassignedCount = 0
  
  shards.value.forEach(shard => {
    if (shard.type === 'primary') primaryCount++
    else if (shard.type === 'replica') replicaCount++
    if (shard.state === 'UNASSIGNED') unassignedCount++
  })
  
  stats.primaryShards = primaryCount
  stats.replicaShards = replicaCount
  stats.unassignedShards = unassignedCount
}

// 更新索引选项
const updateIndexOptions = () => {
  const indexSet = new Set([''])
  shards.value.forEach(shard => {
    if (shard.index) indexSet.add(shard.index)
  })
  
  indexOptions.value = Array.from(indexSet).map(index => ({
    label: index || '全部索引',
    value: index
  }))
}

// 刷新数据
const refreshData = async () => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) {
    ElMessage.warning('请先选择实例')
    return
  }

  loading.value = true
  try {
    // 获取分片信息
    const shardResponse = await getShardInfo(instanceId)
    const shardData = shardResponse.data?.shards || []

    // 获取分片统计信息
    const statsResponse = await getShardStats(instanceId)
    const statsData = statsResponse.data?.stats || {}

    // 转换数据格式
    shards.value = transformShardData(shardData)
    updateStats(statsData)
    
    // 更新索引选项
    updateIndexOptions()
    
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

// 节点变化处理
const handleNodeChange = () => {
  // 过滤逻辑已在computed中实现
}

// 搜索处理
const handleSearch = () => {
  // 搜索逻辑已在computed中实现
}

// 选择变化处理
const handleSelectionChange = (selection) => {
  selectedShards.value = selection
}

// 查看详情
const viewDetails = (shard) => {
  currentShard.value = shard
  detailDialogVisible.value = true
}

// 迁移分片
const relocateShard = (shard) => {
  currentShard.value = shard
  relocateForm.targetNode = ''
  relocateDialogVisible.value = true
}

// 确认迁移
const confirmRelocate = async () => {
  if (!relocateForm.targetNode) {
    ElMessage.warning('请选择目标节点')
    return
  }

  if (relocateForm.targetNode === currentShard.value.node) {
    ElMessage.warning('目标节点不能与当前节点相同')
    return
  }

  try {
    relocateLoading.value = true
    
    // 调用移动分片API
    await moveShard({
      instance_id: getSelectedInstanceId(),
      index: currentShard.value.index,
      shard: currentShard.value.shard,
      from_node: currentShard.value.node,
      to_node: relocateForm.targetNode
    })
    
    ElMessage.success(`分片迁移成功`)
    relocateDialogVisible.value = false
    refreshData()
  } catch (error) {
    ElMessage.error('迁移失败：' + error.message)
  } finally {
    relocateLoading.value = false
  }
}

// 强制分配
const forceAllocate = (shard) => {
  ElMessageBox.confirm(
    `确定要强制分配分片 ${shard.index}[${shard.shard}] 吗？`,
    '确认强制分配',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      // 调用分配分片API
      await allocateShard({
        instance_id: getSelectedInstanceId(),
        index: shard.index,
        shard: shard.shard,
        node: availableNodes.value[0].value // 默认分配到第一个可用节点
      })
      
      ElMessage.success('强制分配成功')
      refreshData()
    } catch (error) {
      ElMessage.error('强制分配失败：' + error.message)
    }
  }).catch(() => {
    ElMessage.info('已取消强制分配操作')
  })
}

// 批量迁移
const batchRelocate = () => {
  ElMessageBox.confirm(
    `确定要批量迁移选中的 ${selectedShards.value.length} 个分片吗？`,
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
      selectedShards.value = []
      refreshData()
    } catch (error) {
      ElMessage.error('批量迁移失败：' + error.message)
    }
  }).catch(() => {
    ElMessage.info('已取消批量迁移操作')
  })
}

// 批量强制分配
const batchForceAllocate = () => {
  const unassignedCount = getUnassignedCount()
  if (unassignedCount === 0) {
    ElMessage.warning('选中的分片中没有未分配的分片')
    return
  }

  ElMessageBox.confirm(
    `确定要批量强制分配选中的 ${unassignedCount} 个未分配分片吗？`,
    '确认批量强制分配',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 3000))
      ElMessage.success('批量强制分配成功')
      selectedShards.value = []
      refreshData()
    } catch (error) {
      ElMessage.error('批量强制分配失败：' + error.message)
    }
  }).catch(() => {
    ElMessage.info('已取消批量强制分配操作')
  })
}

// 跨实例迁移
const migrateToInstance = async (shard) => {
  currentShard.value = shard
  migrateInstanceForm.targetInstanceId = null
  targetInstanceStatus.value = ''
  
  // 加载可用实例列表
  await loadAvailableInstances()
  
  migrateInstanceDialogVisible.value = true
}

// 加载可用实例列表
const loadAvailableInstances = async () => {
  try {
    const response = await getInstanceList({ page: 1, page_size: 100 })
    availableInstances.value = response.data?.list?.data || []
  } catch (error) {
    ElMessage.error('获取实例列表失败：' + error.message)
    availableInstances.value = []
  }
}

// 处理目标实例变化
const handleTargetInstanceChange = async (instanceId) => {
  if (!instanceId) {
    targetInstanceStatus.value = ''
    return
  }
  
  targetInstanceStatus.value = 'checking'
  
  try {
    // 这里可以添加检查目标实例可用性的逻辑
    // 暂时假设所有实例都可用
    await new Promise(resolve => setTimeout(resolve, 1000))
    targetInstanceStatus.value = 'available'
  } catch (error) {
    targetInstanceStatus.value = 'unavailable'
    ElMessage.error('检查目标实例失败：' + error.message)
  }
}

// 确认跨实例迁移
const confirmMigrateToInstance = async () => {
  if (!migrateInstanceForm.targetInstanceId) {
    ElMessage.warning('请选择目标实例')
    return
  }

  if (migrateInstanceForm.targetInstanceId === getSelectedInstanceId()) {
    ElMessage.warning('目标实例不能与当前实例相同')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要将索引 "${currentShard.value.index}" 迁移到目标实例吗？此操作可能需要较长时间，请确认网络连接稳定。`,
      '确认跨实例迁移',
      {
        confirmButtonText: '确定迁移',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    migrateInstanceLoading.value = true
    
    // 调用跨实例迁移API
    await migrateShardToInstance({
      source_instance_id: getSelectedInstanceId(),
      target_instance_id: migrateInstanceForm.targetInstanceId,
      index: currentShard.value.index
    })
    
    ElMessage.success('索引迁移成功，正在后台执行...')
    migrateInstanceDialogVisible.value = false
    
    // 延迟刷新数据，给迁移操作一些时间
    setTimeout(() => {
      refreshData()
    }, 2000)
    
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('迁移失败：' + error.message)
    }
  } finally {
    migrateInstanceLoading.value = false
  }
}

// 组件挂载时加载数据
onMounted(() => {
  // 如果有选中的实例，则加载数据
  if (getSelectedInstanceId()) {
    refreshData()
  }
})
</script>

<style scoped>
.shard-management {
  padding: 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-shrink: 0;
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
  flex-shrink: 0;
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
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.table-card :deep(.el-card__body) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 0;
}

.table-card :deep(.el-table) {
  flex: 1;
}

.table-card :deep(.el-table__body-wrapper) {
  overflow-y: auto;
  max-height: calc(100vh - 400px);
}

.table-card :deep(.el-table) {
  --el-table-border-color: var(--el-border-color-lighter);
  --el-table-header-bg-color: var(--el-fill-color-light);
}

.table-card :deep(.el-table__header) {
  position: sticky;
  top: 0;
  z-index: 10;
}

.table-card :deep(.el-table__row:hover) {
  background-color: var(--el-fill-color-extra-light);
}

.batch-actions {
  margin-top: 20px;
  padding: 15px;
  background-color: #f5f7fa;
  border-radius: 4px;
  flex-shrink: 0;
}

.shard-details {
  padding: 20px 0;
}

.shard-segments h4 {
  margin-bottom: 10px;
  color: #303133;
}

/* 响应式优化 */
@media (max-width: 1200px) {
  .header-actions {
    flex-wrap: wrap;
    gap: 10px;
  }
  
  .header-actions > * {
    margin-right: 0;
  }
}

@media (max-width: 768px) {
  .shard-management {
    padding: 10px;
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }
  
  .header-actions {
    width: 100%;
    justify-content: flex-start;
  }
  
  .stats-cards :deep(.el-col) {
    margin-bottom: 10px;
  }
}
</style>