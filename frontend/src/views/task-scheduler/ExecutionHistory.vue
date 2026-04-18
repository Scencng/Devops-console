<template>
  <div class="page-container">
    <el-card class="page-header-card mb-20" shadow="never">
      <div class="page-header">
        <div class="header-left">
          <div class="header-icon-wrapper">
            <el-icon :size="24"><Document /></el-icon>
          </div>
          <div class="header-title-wrapper">
            <h2>执行历史</h2>
            <p class="subtitle">查看和追踪工作流的历史执行记录</p>
          </div>
        </div>
        <div class="header-right">
          <el-button :icon="Refresh" @click="fetchList">刷新</el-button>
        </div>
      </div>
    </el-card>

    <!-- Filters -->
    <el-card class="content-card mb-20" shadow="never">
      <div class="filter-row flex items-center gap-4">
        <el-select
          v-model="searchForm.workflowId"
          class="filter-item w-64"
          clearable
          filterable
          placeholder="选择工作流"
          @change="fetchList"
        >
          <el-option
            v-for="item in workflowOptions"
            :key="item.id"
            :label="item.name"
            :value="item.id"
          />
        </el-select>
        <el-select
          v-model="searchForm.status"
          class="filter-item w-32"
          clearable
          placeholder="状态筛选"
          @change="fetchList"
        >
          <el-option label="等待中" value="PENDING" />
          <el-option label="运行中" value="RUNNING" />
          <el-option label="已完成" value="COMPLETED" />
          <el-option label="失败" value="FAILED" />
          <el-option label="已取消" value="CANCELLED" />
        </el-select>
        <div class="flex-spacer flex-1"></div>
        <el-button-group>
          <el-button :icon="Search" :loading="loading" type="primary" @click="fetchList">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="resetSearch">重置</el-button>
        </el-button-group>
      </div>
    </el-card>

    <!-- Table View -->
    <el-card class="content-card" shadow="never">
      <el-table v-loading="loading" :data="tableData" stripe style="width: 100%">
        <el-table-column label="执行编号" min-width="180" prop="executionNo">
          <template #default="{ row }">
            <span class="font-medium text-main">{{ row.executionNo }}</span>
          </template>
        </el-table-column>
        <el-table-column label="工作流名称" min-width="150" prop="workflowId">
          <template #default="{ row }">
            {{ getWorkflowName(row.workflowId) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="触发方式" width="120" prop="triggerType" />
        <el-table-column label="开始时间" min-width="160">
          <template #default="{ row }">
            {{ formatTime(row.startTime) }}
          </template>
        </el-table-column>
        <el-table-column label="结束时间" min-width="160">
          <template #default="{ row }">
            {{ formatTime(row.endTime) }}
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="100">
          <template #default="{ row }">
            {{ formatDuration(row.startTime, row.endTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link size="small" type="primary" @click="handleDetail(row)">查看详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Document, Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'
import { getExecutionList, getWorkflowList } from '@/api/workflow.js'

dayjs.extend(duration)

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const tableData = ref([])
const workflowOptions = ref([])

const formatTime = (time) => time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'

const formatDuration = (start, end) => {
  if (!start || !end) return '-'
  const diff = dayjs(end).diff(dayjs(start))
  const d = dayjs.duration(diff)
  if (d.asHours() >= 1) {
    return `${Math.floor(d.asHours())}h ${d.minutes()}m`
  }
  if (d.asMinutes() >= 1) {
    return `${d.minutes()}m ${d.seconds()}s`
  }
  return `${d.seconds()}s`
}

const getStatusType = (status) => {
  const map = {
    PENDING: 'info',
    RUNNING: 'primary',
    COMPLETED: 'success',
    FAILED: 'danger',
    CANCELLED: 'warning'
  }
  return map[status] || 'info'
}

const getWorkflowName = (id) => {
  const workflow = workflowOptions.value.find(w => w.id === id)
  return workflow ? workflow.name : `ID: ${id}`
}

const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const searchForm = reactive({
  workflowId: route.query.workflowId ? Number(route.query.workflowId) : null,
  status: null
})

onMounted(() => {
  fetchWorkflows()
  fetchList()
})

watch(() => route.query.workflowId, (newVal) => {
  if (newVal) {
    searchForm.workflowId = Number(newVal)
    fetchList()
  }
})

async function fetchWorkflows() {
  try {
    const res = await getWorkflowList({ page: 1, pageSize: 1000 })
    const d = res?.data?.data || res?.data || {}
    workflowOptions.value = d.list || []
  } catch (e) {
    console.error('获取工作流列表失败', e)
  }
}

async function fetchList() {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      workflowId: searchForm.workflowId || undefined,
      status: searchForm.status || undefined
    }
    const res = await getExecutionList(params)
    const d = res?.data?.data || res?.data || {}
    tableData.value = d.list || []
    pagination.total = d.total || 0
  } catch (e) {
    ElMessage.error(e?.message || '获取执行历史失败')
  } finally {
    loading.value = false
  }
}

function resetSearch() {
  searchForm.workflowId = null
  searchForm.status = null
  pagination.page = 1
  fetchList()
}

function handleDetail(row) {
  router.push(`/task-scheduler/executions/${row.id}/logs`)
}
</script>

<style scoped>
.page-container { padding: 16px; display: flex; flex-direction: column; gap: 0; }
.header-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: var(--primary-color, #409eff);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-left { display: flex; align-items: center; gap: 12px; }
.header-title-wrapper h2 { margin: 0; font-size: 18px; color: var(--text-main, #303133); }
.header-title-wrapper .subtitle { margin: 0; font-size: 12px; color: var(--text-sub, #909399); }
.header-right { display: flex; align-items: center; gap: 12px; }

.flex { display: flex; }
.items-center { align-items: center; }
.gap-4 { gap: 16px; }
.mb-20 { margin-bottom: 20px; }
.w-64 { width: 16rem; }
.w-32 { width: 8rem; }
.flex-1 { flex: 1; }
.pagination { margin-top: 16px; display: flex; justify-content: flex-end; }
.text-main { color: var(--text-main, #303133); }
.font-medium { font-weight: 500; }
</style>
