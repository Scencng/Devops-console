<template>
  <div class="page-container">
    <!-- Header -->
    <el-card class="page-header-card">
      <div class="page-header">
        <div class="header-left">
          <el-button link class="back-btn" @click="$router.push('/cicd/pipelines')">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <div class="header-content ml-4">
            <div class="flex items-center gap-3">
              <h2 class="m-0 page-title">{{ pipelineName || '流水线' }}</h2>
              <el-tag type="info" effect="plain" class="font-medium">运行历史</el-tag>
            </div>
            <div class="meta-row mt-2">
              <span class="meta-item">共 {{ total }} 次运行</span>
            </div>
          </div>
        </div>
        <div class="header-right">
          <el-button
            v-show="selectedRows.length > 0"
            type="danger"
            :icon="Delete"
            @click="handleBatchDelete"
          >批量删除（{{ selectedRows.length }}）</el-button>
          <el-button @click="fetchData" :loading="loading" :icon="Refresh">刷新</el-button>
          <el-button type="primary" @click="handleRun" :loading="triggering" :icon="VideoPlay">立即运行</el-button>
        </div>
      </div>
    </el-card>

    <!-- Run List -->
    <el-card class="content-card" v-loading="loading">
      <el-table
        :data="runs"
        style="width: 100%"
        row-class-name="run-row"
        @row-click="handleRowClick"
        @selection-change="selectedRows = $event"
      >
        <el-table-column type="selection" width="44" @click.stop />
        <el-table-column label="运行编号" width="100">
          <template #default="{ row }">
            <span class="run-id">#{{ row.id }}</span>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="130">
          <template #default="{ row }">
            <div class="status-cell">
              <span class="status-dot" :class="getStatusDotClass(row.status)"></span>
              <el-tag :type="getStatusType(row.status)" effect="light" size="small">
                {{ row.status || 'Pending' }}
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="触发人" width="120">
          <template #default="{ row }">
            <div class="user-cell">
              <el-icon><User /></el-icon>
              <span>{{ row.operator || '-' }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="分支" width="140">
          <template #default="{ row }">
            <el-tag size="small" effect="plain" v-if="row.branch">{{ row.branch }}</el-tag>
            <span v-else class="text-sub">-</span>
          </template>
        </el-table-column>

        <el-table-column label="耗时" width="120">
          <template #default="{ row }">
            <div class="duration-cell">
              <el-icon><Timer /></el-icon>
              <span>{{ formatDuration(row.duration) }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="开始时间" min-width="180">
          <template #default="{ row }">
            <span class="text-sub">{{ formatTime(row.startTime) }}</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="180" align="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              link
              size="small"
              @click.stop="handleViewDetail(row)"
            >
              查看详情 <el-icon class="el-icon--right"><ArrowRight /></el-icon>
            </el-button>
            <el-button
              type="danger"
              link
              size="small"
              @click.stop="handleDelete(row)"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination-wrapper" v-if="total > pageSize">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @change="fetchData"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getPipeline, getPipelineRuns, triggerPipeline, deletePipelineRun } from '@/api/cicd.js'
import { ArrowLeft, ArrowRight, Refresh, VideoPlay, User, Timer, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()
const pipelineId = route.params.id

const pipelineName = ref('')
const runs = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const triggering = ref(false)
const pollingTimer = ref(null)
const selectedRows = ref([])

const fetchData = async (silent = false) => {
  if (!silent) loading.value = true
  try {
    const res = await getPipelineRuns({
      pageNum: page.value,
      pageSize: pageSize.value,
      pipelineId: pipelineId
    })
    const data = res.data?.data || {}
    runs.value = data.data || []
    total.value = data.total || 0
  } catch(e) {
    if (!silent) ElMessage.error('加载运行记录失败')
  } finally {
    if (!silent) loading.value = false
  }
}

const fetchPipelineInfo = async () => {
  try {
    const res = await getPipeline(pipelineId)
    pipelineName.value = res.data?.data?.name || ''
  } catch(e) { /* 静默失败 */ }
}

const handleRun = async () => {
  triggering.value = true
  try {
    await triggerPipeline(pipelineId, {})
    ElNotification({
      title: '触发成功',
      message: '流水线已开始运行',
      type: 'success',
      duration: 3000
    })
    setTimeout(fetchData, 1500)
  } catch(e) {
    ElMessage.error(e.response?.data?.message || '触发失败')
  } finally {
    triggering.value = false
  }
}

const handleViewDetail = (row) => {
  router.push(`/cicd/pipelines/${pipelineId}/runs/${row.id}`)
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确认删除运行记录 #${row.id}？同时会删除 Argo 中对应的 Workflow。`,
      '删除确认',
      { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' }
    )
    await deletePipelineRun(row.id)
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const handleBatchDelete = async () => {
  if (selectedRows.value.length === 0) return
  try {
    await ElMessageBox.confirm(
      `确认删除选中的 ${selectedRows.value.length} 条运行记录？同时会删除 Argo 中对应的 Workflow。`,
      '批量删除确认',
      { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' }
    )
    await Promise.all(selectedRows.value.map(row => deletePipelineRun(row.id)))
    ElMessage.success(`已删除 ${selectedRows.value.length} 条记录`)
    selectedRows.value = []
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('批量删除失败')
  }
}

const handleRowClick = (row) => {
  handleViewDetail(row)
}

const getStatusType = (status) => {
  const map = { 'Succeeded': 'success', 'Failed': 'danger', 'Running': 'warning', 'Pending': 'info', 'Error': 'danger' }
  return map[status] || 'info'
}

const getStatusDotClass = (status) => {
  const map = { 'Succeeded': 'dot-success', 'Failed': 'dot-failed', 'Running': 'dot-running', 'Error': 'dot-failed' }
  return map[status] || 'dot-pending'
}

const formatTime = (time) => {
  if (!time) return '-'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const formatDuration = (seconds) => {
  if (!seconds && seconds !== 0) return '-'
  if (seconds === 0) return '< 1s'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  if (h > 0) return `${h}h ${m}m`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

const startPolling = () => {
  pollingTimer.value = setInterval(() => {
    // 有选中行时跳过刷新，避免丢失 checkbox 状态
    if (selectedRows.value.length > 0) return
    const hasRunning = runs.value.some(r => r.status === 'Running' || r.status === 'Pending')
    if (hasRunning) fetchData(true)
  }, 15000) // 15s 轮询，减少频率
}

onMounted(() => {
  fetchPipelineInfo()
  fetchData()
  startPolling()
})

onUnmounted(() => {
  if (pollingTimer.value) clearInterval(pollingTimer.value)
})
</script>

<style scoped>
.page-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header-card { flex-shrink: 0; }

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left { display: flex; align-items: center; }
.header-right { display: flex; gap: 10px; }
.back-btn { font-size: 18px; color: var(--text-sub); }
.ml-4 { margin-left: 16px; }
.m-0 { margin: 0; }
.mt-2 { margin-top: 8px; }

.page-title { font-size: 18px; font-weight: 600; color: var(--text-main); }
.meta-row { display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--text-sub); }
.meta-item { display: flex; align-items: center; gap: 4px; }

.content-card { flex: 1; }

/* Table row clickable */
:deep(.run-row) { cursor: pointer; }
:deep(.run-row:hover td) { background: var(--el-table-row-hover-bg-color) !important; }

.run-id {
  font-family: 'Consolas', monospace;
  font-weight: 600;
  color: var(--primary-color);
}

.status-cell { display: flex; align-items: center; gap: 8px; }
.status-dot {
  width: 7px; height: 7px;
  border-radius: 50%; flex-shrink: 0;
}
.dot-success { background: #67c23a; }
.dot-failed { background: #f56c6c; }
.dot-running { background: #409eff; animation: pulse 1.5s infinite; }
.dot-pending { background: #909399; }

@keyframes pulse {
  0% { box-shadow: 0 0 0 0 rgba(64, 158, 255, 0.5); }
  70% { box-shadow: 0 0 0 5px rgba(64, 158, 255, 0); }
  100% { box-shadow: 0 0 0 0 rgba(64, 158, 255, 0); }
}

.user-cell, .duration-cell {
  display: flex; align-items: center; gap: 6px;
  color: var(--text-sub); font-size: 13px;
}

.text-sub { color: var(--text-sub); }

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

/* Flex helpers */
.flex { display: flex; }
.items-center { align-items: center; }
.gap-3 { gap: 12px; }
.font-medium { font-weight: 500; }
</style>
