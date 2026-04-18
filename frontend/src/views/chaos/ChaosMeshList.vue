<template>
  <div class="autoops-container chaos-list-page">
    <!-- Header Stats Cards -->
    <div class="chaos-stats-row">
      <div class="chaos-stat-card" v-for="stat in statsCards" :key="stat.key">
        <div class="stat-icon-wrap" :class="stat.theme">
          <el-icon :size="22"><component :is="stat.icon" /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ stat.value }}</span>
          <span class="stat-label">{{ stat.label }}</span>
        </div>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="autoops-filter-bar">
      <label class="autoops-filter-label">命名空间</label>
      <el-select
        v-model="selectedNamespace"
        placeholder="选择命名空间"
        class="autoops-select"
        style="width: 180px"
        filterable
        @change="handleFilterChange"
      >
        <el-option label="全部命名空间" value="" />
        <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
      </el-select>

      <label class="autoops-filter-label">故障类型</label>
      <el-select
        v-model="selectedFaultType"
        placeholder="全部类型"
        class="autoops-select"
        style="width: 160px"
        clearable
        @change="handleFilterChange"
      >
        <el-option label="全部类型" value="" />
        <el-option v-for="ft in faultTypes" :key="ft.value" :label="ft.label" :value="ft.value" />
      </el-select>

      <label class="autoops-filter-label">状态</label>
      <el-select
        v-model="selectedStatus"
        placeholder="全部状态"
        class="autoops-select"
        style="width: 140px"
        clearable
        @change="handleFilterChange"
      >
        <el-option label="全部状态" value="" />
        <el-option label="运行中" value="Running" />
        <el-option label="已暂停" value="Paused" />
        <el-option label="已完成" value="Finished" />
        <el-option label="失败" value="Failed" />
      </el-select>

      <el-input
        v-model="searchKeyword"
        placeholder="搜索实验名称"
        class="autoops-input"
        style="width: 220px"
        clearable
        @keyup.enter="handleFilterChange"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>

      <el-button type="primary" class="autoops-btn-primary" @click="handleFilterChange">
        <el-icon><Search /></el-icon>
        <span>搜索</span>
      </el-button>
      <el-button class="autoops-btn-secondary" @click="handleReset">
        <el-icon><RefreshRight /></el-icon>
        <span>重置</span>
      </el-button>

      <div style="flex: 1" />

      <el-button type="primary" class="autoops-btn-primary chaos-create-btn" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        <span>创建实验</span>
      </el-button>
    </div>

    <!-- Data Table -->
    <div class="autoops-table-wrapper">
      <el-table
        :data="filteredList"
        class="autoops-table"
        style="width: 100%"
        v-loading="loading"
        :empty-text="loading ? '加载中...' : '暂无混沌实验'"
        row-key="name"
      >
        <!-- Name -->
        <el-table-column label="实验名称" min-width="200">
          <template #default="{ row }">
            <div class="name-cell">
              <div :class="getTypeTheme(row.faultType)" class="chaos-type-icon">
                <el-icon :size="16"><component :is="getTypeIcon(row.faultType)" /></el-icon>
              </div>
              <div class="name-content">
                <a class="name-link" @click="handleViewDetail(row)">{{ row.name }}</a>
                <span class="resource-type">{{ row.faultType || 'Chaos' }}</span>
              </div>
            </div>
          </template>
        </el-table-column>

        <!-- Namespace -->
        <el-table-column label="命名空间" min-width="120">
          <template #default="{ row }">
            <el-tag size="small" effect="plain" round>{{ row.namespace }}</el-tag>
          </template>
        </el-table-column>

        <!-- Fault Type -->
        <el-table-column label="故障类型" min-width="140">
          <template #default="{ row }">
            <div class="fault-type-cell">
              <span :class="getTypeTheme(row.faultType)" class="fault-badge">
                {{ getFaultLabel(row.faultType) }}
              </span>
            </div>
          </template>
        </el-table-column>

        <!-- Status -->
        <el-table-column label="状态" min-width="110" align="center">
          <template #default="{ row }">
            <div class="status-cell">
              <span class="status-dot" :class="getStatusTheme(row.status)"></span>
              <span class="status-text" :class="getStatusTheme(row.status)">{{ getStatusLabel(row.status) }}</span>
            </div>
          </template>
        </el-table-column>

        <!-- Duration -->
        <el-table-column label="持续时间" min-width="100" align="center">
          <template #default="{ row }">
            <span class="duration-text">{{ row.duration || '持续' }}</span>
          </template>
        </el-table-column>

        <!-- Created Time -->
        <el-table-column label="创建时间" width="175">
          <template #default="{ row }">
            <span class="time-text">{{ formatDate(row.createdAt) }}</span>
          </template>
        </el-table-column>

        <!-- Actions -->
        <el-table-column fixed="right" label="操作" width="240">
          <template #default="{ row }">
            <div class="autoops-actions">
              <el-tooltip content="查看详情" placement="top" :show-after="300">
                <button class="autoops-action-btn view" @click="handleViewDetail(row)">
                  <el-icon><View /></el-icon>
                </button>
              </el-tooltip>
              <el-tooltip :content="(row.status === 'Paused' || row.status === 'paused') ? '恢复实验' : '暂停实验'" :show-after="300" placement="top">
                <button
                  class="autoops-action-btn"
                  :class="(row.status === 'Paused' || row.status === 'paused') ? 'update' : 'scale'"
                  @click="handleTogglePause(row)"
                  :disabled="row.status === 'Finished' || row.status === 'Failed'"
                >
                  <el-icon>
                    <VideoPlay v-if="row.status === 'Paused' || row.status === 'paused'" />
                    <VideoPause v-else />
                  </el-icon>
                </button>
              </el-tooltip>
              <el-tooltip v-if="row.labels && row.labels['chaos-eviction'] === 'true'" :show-after="300" content="清理演练环境" placement="top">
                <button class="autoops-action-btn cleanup" @click="handleCleanupEviction(row)">
                  <el-icon><RefreshRight /></el-icon>
                </button>
              </el-tooltip>
              <el-tooltip content="删除实验" placement="top" :show-after="300">
                <button class="autoops-action-btn delete" @click="handleDelete(row)">
                  <el-icon><Delete /></el-icon>
                </button>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination-container">
        <div class="pagination-meta">
          <span v-if="lastUpdated" class="last-updated">最后更新：{{ lastUpdated }}</span>
          <span v-if="hasRunningExperiments" class="polling-badge">
            <span class="polling-dot"></span> 实时更新中
          </span>
        </div>
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <!-- Detail Dialog (Quick View) -->
    <el-dialog
      v-model="showDetailDialog"
      :title="'实验详情: ' + detailData.name"
      width="760px"
      :close-on-click-modal="false"
      append-to-body
      destroy-on-close
    >
      <ChaosDetail
        v-if="showDetailDialog"
        :experiment="detailData"
        :instance-id="getSelectedInstanceId()"
        @pause="handleTogglePauseFromDetail"
        @delete="handleDeleteFromDetail"
      />
    </el-dialog>
  </div>
</template>

<script setup>
import {computed, onMounted, onUnmounted, reactive, ref} from 'vue'
import {useRouter} from 'vue-router'
import {ElMessage, ElMessageBox} from 'element-plus'
import {
  Aim,
  CircleCheck,
  Connection,
  Cpu,
  Delete,
  Document,
  Lightning,
  Plus,
  RefreshRight,
  Search,
  VideoPause,
  VideoPlay,
  View,
  Warning
} from '@element-plus/icons-vue'
import {
  cleanupEviction,
  deleteChaosExperiment,
  getChaosExperiment,
  getChaosExperiments,
  pauseChaosExperiment,
  resumeChaosExperiment
} from '@/api/chaos.js'
import {getNamespaceList} from '@/api/k8s/namespace'
import {getSelectedInstanceId} from '@/stores/instanceStore'
import dayjs from 'dayjs'
import ChaosDetail from './ChaosMeshDetail.vue'

const router = useRouter()
const loading = ref(false)
const experimentList = ref([])
const namespaceList = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const lastUpdated = ref('')
let pollingTimer = null

// Filters
const selectedNamespace = ref('')
const selectedFaultType = ref('')
const selectedStatus = ref('')
const searchKeyword = ref('')

// Detail
const showDetailDialog = ref(false)
const detailData = reactive({
  name: '',
  namespace: '',
  faultType: '',
  status: '',
  duration: '',
  createdAt: '',
  spec: {},
  events: []
})

// Fault type definitions
const faultTypes = [
  { value: 'PodChaos', label: 'Pod 故障', icon: Aim, theme: 'pod' },
  { value: 'NetworkChaos', label: '网络故障', icon: Connection, theme: 'network' },
  { value: 'IOChaos', label: 'IO 故障', icon: Document, theme: 'io' },
  { value: 'StressChaos', label: '压力故障', icon: Cpu, theme: 'stress' }
]

// Computed stats
const hasRunningExperiments = computed(() =>
  experimentList.value.some(e => {
    const s = (e.status || '').toLowerCase()
    // 只要不是明确的结束/暂停状态（如 Unknown、Running等），就保持静默轮询更新
    return !['finished', 'failed', 'paused'].includes(s)
  })
)

const statsCards = computed(() => {
  const list = experimentList.value
  return [
    {
      key: 'total',
      label: '实验总数',
      value: list.length,
      icon: Lightning,
      theme: 'primary'
    },
    {
      key: 'running',
      label: '运行中',
      value: list.filter(e => e.status === 'Running' || e.status === 'running').length,
      icon: VideoPlay,
      theme: 'success'
    },
    {
      key: 'paused',
      label: '已暂停',
      value: list.filter(e => e.status === 'Paused' || e.status === 'paused').length,
      icon: VideoPause,
      theme: 'warning'
    },
    {
      key: 'finished',
      label: '已完成',
      value: list.filter(e => ['Finished', 'finished', 'Failed', 'failed'].includes(e.status)).length,
      icon: CircleCheck,
      theme: 'info'
    }
  ]
})

// Filtered + paginated list
const filteredList = computed(() => {
  let list = experimentList.value

  if (selectedFaultType.value) {
    list = list.filter(e => e.faultType === selectedFaultType.value)
  }
  if (selectedStatus.value) {
    list = list.filter(e => {
        if (!e.status) return false
        return e.status.toLowerCase() === selectedStatus.value.toLowerCase()
    })
  }
  if (searchKeyword.value) {
    const kw = searchKeyword.value.toLowerCase()
    list = list.filter(e => e.name?.toLowerCase().includes(kw))
  }

  total.value = list.length
  const start = (currentPage.value - 1) * pageSize.value
  return list.slice(start, start + pageSize.value)
})

// Helpers
const getFaultLabel = (kind) => {
  const ft = faultTypes.find(f => f.value === kind)
  return ft ? ft.label : kind || '未知'
}

const getTypeIcon = (kind) => {
  const map = {
    PodChaos: Aim,
    NetworkChaos: Connection,
    IOChaos: Document,
    StressChaos: Cpu
  }
  return map[kind] || Warning
}

const getTypeTheme = (kind) => {
  const map = {
    PodChaos: 'pod',
    NetworkChaos: 'network',
    IOChaos: 'io',
    StressChaos: 'stress'
  }
  return map[kind] || 'default'
}

const getStatusTheme = (status) => {
  if (!status) return 'unknown'
  const cap = typeof status === 'string' ? status.charAt(0).toUpperCase() + status.slice(1).toLowerCase() : status
  const map = {
    Running: 'running',
    Paused: 'paused',
    Finished: 'finished',
    Failed: 'failed'
  }
  return map[cap] || 'unknown'
}

const getStatusLabel = (status) => {
  if (!status) return '未知'
  const cap = typeof status === 'string' ? status.charAt(0).toUpperCase() + status.slice(1).toLowerCase() : status
  const map = {
    Running: '运行中',
    Paused: '已暂停',
    Finished: '已完成',
    Failed: '失败'
  }
  return map[cap] || status || '未知'
}

const formatDate = (ts) => {
  if (!ts) return '-'
  return dayjs(ts).isValid() ? dayjs(ts).format('YYYY-MM-DD HH:mm:ss') : '-'
}

// Data fetching
const fetchNamespaces = async () => {
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getNamespaceList(instanceId)
    const list = res.data?.items || res.data?.namespaceList || res.data || []
    namespaceList.value = list.map(item => ({ name: typeof item === 'string' ? item : item.name }))
  } catch (e) {
    console.error('获取命名空间失败', e)
  }
}

const fetchData = async (silent = false) => {
  if (!silent) {
    loading.value = true
  }
  try {
    const instanceId = getSelectedInstanceId()
    const ns = selectedNamespace.value || 'all'
    const res = await getChaosExperiments(ns, instanceId)

    // API might return { faults: null } or an array directly
    const d = res.data || {}
    let list = d.experiments || d.items || d.faults || d
    if (!Array.isArray(list)) {
      list = []
    }
    experimentList.value = list
    total.value = experimentList.value.length
    lastUpdated.value = dayjs().format('HH:mm:ss')
  } catch (e) {
    if (!silent) ElMessage.error('获取混沌实验列表失败')
    experimentList.value = []
    total.value = 0
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

const startPolling = () => {
  pollingTimer = setInterval(() => {
    if (hasRunningExperiments.value) {
      // 静默拉取，不触发 loading 动画，防止列表闪烁
      fetchData(true)
    }
  }, 5000)
}

// Event handlers
const handleFilterChange = () => {
  currentPage.value = 1
  fetchData()
}

const handleReset = () => {
  selectedNamespace.value = ''
  selectedFaultType.value = ''
  selectedStatus.value = ''
  searchKeyword.value = ''
  currentPage.value = 1
  fetchData()
}

const handlePageChange = () => {
  // Pagination is computed client-side, no need to re-fetch
}

const handleSizeChange = () => {
  currentPage.value = 1
}

const handleCreate = () => {
  router.push({ path: '/chaos/create' })
}

const handleViewDetail = async (row) => {
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getChaosExperiment(row.namespace, row.name, instanceId)
    const detail = res.data?.experiment || res.data || {}
    Object.assign(detailData, {
      name: detail.name || row.name,
      namespace: detail.namespace || row.namespace,
      faultType: detail.faultType || row.faultType,
      status: detail.status || row.status,
      duration: detail.duration || row.duration,
      createdAt: detail.createdAt || row.createdAt,
      spec: detail.spec || detail,
      events: detail.events || []
    })
    showDetailDialog.value = true
  } catch (e) {
    ElMessage.error('获取实验详情失败')
  }
}

const handleTogglePause = async (row) => {
  const isPaused = row.status === 'Paused' || row.status === 'paused'
  const action = isPaused ? '恢复' : '暂停'
  try {
    await ElMessageBox.confirm(`确定${action}实验 "${row.name}"？`, `${action}实验`, {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    })
    const instanceId = getSelectedInstanceId()
    if (isPaused) {
      await resumeChaosExperiment(row.namespace, row.name, instanceId)
    } else {
      await pauseChaosExperiment(row.namespace, row.name, instanceId)
    }
    ElMessage.success(`${action}实验成功`)
    fetchData()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error(`${action}实验失败`)
    }
  }
}

const handleTogglePauseFromDetail = (row) => {
  showDetailDialog.value = false
  handleTogglePause(row)
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定删除混沌实验 "${row.name}"？此操作不可恢复。`,
      '删除实验',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消', confirmButtonClass: 'el-button--danger' }
    )
    const instanceId = getSelectedInstanceId()
    await deleteChaosExperiment(row.namespace, row.name, instanceId)
    ElMessage.success('删除实验成功')
    fetchData()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('删除实验失败')
    }
  }
}

const handleDeleteFromDetail = (row) => {
  showDetailDialog.value = false
  handleDelete(row)
}

// 清理演练环境
const handleCleanupEviction = async (row) => {
  const nodeName = row.labels?.['chaos-node']
  const deployName = row.labels?.['chaos-deploy']
  if (!nodeName || !deployName) {
    ElMessage.warning('找不到演练节点或 Deployment 信息')
    return
  }
  try {
    await ElMessageBox.confirm(
      `确定清理实验 "${row.name}" 的演练环境？\n\n这将会回滚 Deployment、去掉污点并将 Pod 迁回正常节点。`,
      '清理演练环境',
      { type: 'warning', confirmButtonText: '确定清理', cancelButtonText: '取消' }
    )
    const instanceId = getSelectedInstanceId()
    const deployNamespace = row.labels?.['chaos-namespace'] || row.namespace
    await cleanupEviction({
      nodeName,
      namespace: row.namespace,
      deploymentNamespace: deployNamespace,
      deploymentName: deployName
    }, instanceId)
    ElMessage.success('演练环境已清理，Pod 正在恢复到正常节点')
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('清理失败: ' + (e.response?.data?.message || e.message || '未知错误'))
  }
}

onMounted(() => {
  fetchData()
  fetchNamespaces()
  startPolling()
})

onUnmounted(() => {
  if (pollingTimer) {
    clearInterval(pollingTimer)
    pollingTimer = null
  }
})
</script>

<style scoped>
/* ============================================
   Chaos Mesh List — Stats Row
   ============================================ */
.chaos-stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--autoops-spacing-md);
  margin-bottom: var(--autoops-spacing-lg);
}

.chaos-stat-card {
  background: var(--autoops-bg-card);
  border-radius: var(--autoops-radius-large);
  border: 1px solid var(--autoops-border-light);
  padding: var(--autoops-spacing-lg) var(--autoops-spacing-lg);
  display: flex;
  align-items: center;
  gap: var(--autoops-spacing-md);
  transition: all 0.3s ease;
  box-shadow: var(--autoops-shadow-light);
}

.chaos-stat-card:hover {
  box-shadow: var(--autoops-shadow-base);
  transform: translateY(-2px);
}

.stat-icon-wrap {
  width: 48px;
  height: 48px;
  border-radius: var(--autoops-radius-base);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon-wrap.primary {
  background: rgba(64, 158, 255, 0.1);
  color: var(--autoops-primary);
}
.stat-icon-wrap.success {
  background: rgba(103, 194, 58, 0.1);
  color: var(--autoops-success);
}
.stat-icon-wrap.warning {
  background: rgba(230, 162, 60, 0.1);
  color: var(--autoops-warning);
}
.stat-icon-wrap.info {
  background: rgba(144, 147, 153, 0.1);
  color: var(--autoops-info);
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--autoops-text-primary);
  line-height: 1.2;
}

.stat-label {
  font-size: 13px;
  color: var(--autoops-text-secondary);
  margin-top: 2px;
}

/* ============================================
   Type Icon in Table
   ============================================ */
.chaos-type-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--autoops-radius-base);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.chaos-type-icon.pod {
  background: rgba(245, 108, 108, 0.1);
  color: var(--autoops-danger);
}
.chaos-type-icon.network {
  background: rgba(64, 158, 255, 0.1);
  color: var(--autoops-primary);
}
.chaos-type-icon.io {
  background: rgba(230, 162, 60, 0.1);
  color: var(--autoops-warning);
}
.chaos-type-icon.stress {
  background: rgba(155, 89, 182, 0.1);
  color: #9B59B6;
}
.chaos-type-icon.default {
  background: rgba(144, 147, 153, 0.1);
  color: var(--autoops-info);
}

/* Name Cell */
.name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.name-content {
  display: flex;
  flex-direction: column;
}

.name-link {
  color: var(--autoops-primary);
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  font-size: 14px;
}
.name-link:hover {
  text-decoration: underline;
}

.resource-type {
  font-size: 12px;
  color: var(--autoops-warning);
  margin-top: 2px;
}

/* Fault Badge */
.fault-type-cell {
  display: flex;
  align-items: center;
}

.fault-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.fault-badge.pod {
  background: rgba(245, 108, 108, 0.1);
  color: var(--autoops-danger);
}
.fault-badge.network {
  background: rgba(64, 158, 255, 0.1);
  color: var(--autoops-primary);
}
.fault-badge.io {
  background: rgba(230, 162, 60, 0.1);
  color: var(--autoops-warning);
}
.fault-badge.stress {
  background: rgba(155, 89, 182, 0.1);
  color: #9B59B6;
}
.fault-badge.default {
  background: rgba(144, 147, 153, 0.1);
  color: var(--autoops-info);
}

/* Status */
.status-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot.running {
  background: var(--autoops-success);
  box-shadow: 0 0 6px rgba(103, 194, 58, 0.5);
  animation: pulse-dot 2s ease-in-out infinite;
}
.status-dot.paused {
  background: var(--autoops-warning);
}
.status-dot.finished {
  background: var(--autoops-info);
}
.status-dot.failed {
  background: var(--autoops-danger);
}
.status-dot.unknown {
  background: var(--autoops-text-placeholder);
}

.status-text {
  font-size: 13px;
  font-weight: 500;
}
.status-text.running { color: var(--autoops-success); }
.status-text.paused { color: var(--autoops-warning); }
.status-text.finished { color: var(--autoops-info); }
.status-text.failed { color: var(--autoops-danger); }
.status-text.unknown { color: var(--autoops-text-placeholder); }

@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

/* Duration */
.duration-text {
  font-size: 13px;
  color: var(--autoops-text-regular);
  font-family: monospace;
}

/* Time */
.time-text {
  font-size: 13px;
  color: var(--autoops-text-secondary);
}

/* Pagination */
.pagination-container {
  padding-top: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.pagination-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
}

.last-updated {
  color: var(--autoops-text-placeholder);
}

.polling-badge {
  display: flex;
  align-items: center;
  gap: 5px;
  color: var(--autoops-success);
  font-size: 12px;
  font-weight: 500;
}

.polling-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--autoops-success);
  animation: pulse-dot 1.5s ease-in-out infinite;
}

/* Create Button Enhancement */
.chaos-create-btn {
  background: linear-gradient(135deg, var(--autoops-primary) 0%, var(--autoops-primary-light) 100%) !important;
  border: none !important;
  font-weight: 600;
}

/* Responsive */
@media (max-width: 1024px) {
  .chaos-stats-row {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .chaos-stats-row {
    grid-template-columns: 1fr;
  }
}
</style>
