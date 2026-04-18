<template>
  <div class="page-container">
    <el-card class="page-header-card mb-20" shadow="never">
      <div class="page-header">
        <div class="header-left">
          <div class="header-icon-wrapper">
            <el-icon :size="24"><Clock /></el-icon>
          </div>
          <div class="header-title-wrapper">
            <h2>工作流编排</h2>
            <p class="subtitle">管理和调度自动化工作流任务</p>
          </div>
        </div>
        <div class="header-right">
          <el-button :icon="Plus" type="primary" @click="handleCreate">新建工作流</el-button>
          <el-button :icon="Refresh" @click="fetchList">刷新</el-button>
        </div>
      </div>
    </el-card>

    <!-- Filters -->
    <el-card class="content-card mb-20" shadow="never">
      <div class="filter-row flex items-center gap-4">
        <el-input
          v-model="searchForm.name"
          :prefix-icon="Search"
          class="filter-item w-64"
          clearable
          placeholder="搜索工作流名称"
          @keyup.enter="fetchList"
        />
        <el-select v-model="searchForm.status" class="filter-item w-32" clearable placeholder="状态筛选">
          <el-option :value="1" label="启用" />
          <el-option :value="0" label="禁用" />
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
        <el-table-column label="工作流名称" min-width="150" prop="name">
          <template #default="{ row }">
            <span class="font-medium text-main">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="cronExpression" label="调度周期" min-width="120" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="(v) => handleStatusChange(row, v)"
            />
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="160">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link size="small" type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link size="small" type="success" @click="handleExecute(row)">执行</el-button>
            <el-button link size="small" type="primary" @click="viewHistory(row)">历史</el-button>
            <el-button link size="small" type="danger" @click="handleDelete(row)">删除</el-button>
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
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Refresh, RefreshLeft, Search, Clock } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import {
  getWorkflowList,
  deleteWorkflow,
  updateWorkflowStatus,
  executeWorkflow
} from '@/api/workflow.js'

const router = useRouter()
const loading = ref(false)
const tableData = ref([])

const formatTime = (time) => time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'

const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const searchForm = reactive({ name: '', status: null })

onMounted(() => {
  fetchList()
})

async function fetchList() {
  loading.value = true
  try {
    const params = { page: pagination.page, pageSize: pagination.pageSize, ...searchForm }
    if (params.status === null) delete params.status
    const res = await getWorkflowList(params)
    const d = res?.data?.data || res?.data || {}
    tableData.value = d.list || []
    pagination.total = d.total || 0
  } catch (e) {
    ElMessage.error(e?.message || '获取工作流列表失败')
  } finally {
    loading.value = false
  }
}

function resetSearch() {
  searchForm.name = ''
  searchForm.status = null
  pagination.page = 1
  fetchList()
}

function handleCreate() {
  router.push('/task-scheduler/workflows/new')
}

function handleEdit(row) {
  router.push(`/task-scheduler/workflows/${row.id}/edit`)
}

const viewHistory = (row) => {
  router.push({
    path: '/task-scheduler/executions',
    query: { workflowId: row.id }
  })
}

async function handleExecute(row) {
  try {
    const res = await executeWorkflow(row.id)
    const execId = res.data?.execution_id
    if (execId) {
      ElMessage.success('执行已启动')
      router.push(`/task-scheduler/executions/${execId}/logs`)
    } else {
      ElMessage.success('执行已启动')
    }
  } catch (e) {
    ElMessage.error('执行失败')
  }
}

async function handleStatusChange(row, val) {
  try {
    await updateWorkflowStatus(row.id, val)
    ElMessage.success('状态更新成功')
  } catch {
    row.status = val === 1 ? 0 : 1
    ElMessage.error('操作失败')
  }
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除工作流「${row.name}」？`, '提示', { type: 'warning' })
    await deleteWorkflow(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
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
