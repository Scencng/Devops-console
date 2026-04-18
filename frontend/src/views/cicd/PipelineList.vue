<template>
  <div class="page-container">
    <el-card class="page-header-card">
       <div class="page-header">
         <div class="header-left">
           <div class="header-title-wrapper">
             <h2>流水线管理</h2>
             <p class="subtitle">CI/CD 构建与部署流程</p>
           </div>
         </div>
         <div class="header-right">
            <el-button type="primary" @click="handleCreate">
               <el-icon><Plus /></el-icon> 新建流水线
            </el-button>
            <el-button @click="refreshData" :loading="loading">
               <el-icon><Refresh /></el-icon>
            </el-button>
         </div>
       </div>
    </el-card>

    <!-- Stats Section -->
    <el-row :gutter="20" class="mb-20">
       <el-col :span="8">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-content">
               <div class="stat-icon bg-gradient-blue"><el-icon><Monitor /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ totalPipelines }}</div>
                  <div class="stat-label">流水线总数</div>
               </div>
            </div>
          </el-card>
       </el-col>
       <el-col :span="8">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-content">
               <div class="stat-icon bg-gradient-green"><el-icon><VideoPlay /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ activeRuns }}</div>
                  <div class="stat-label">运行中</div>
               </div>
            </div>
          </el-card>
       </el-col>
       <el-col :span="8">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-content">
               <div class="stat-icon bg-gradient-purple"><el-icon><TrendCharts /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ successRate }}%</div>
                  <div class="stat-label">成功率</div>
               </div>
            </div>
          </el-card>
       </el-col>
    </el-row>

    <!-- Main Content -->
    <el-card class="content-card">
       <template #header>
          <div class="flex-between">
             <div class="left-tools">
                <el-input
                  v-model="searchQuery"
                  placeholder="搜索流水线..."
                  style="width: 300px"
                  prefix-icon="Search"
                  clearable
                  @change="refreshData"
                />
             </div>
             <div class="right-tools">
                <el-radio-group v-model="viewMode" size="small">
                   <el-radio-button label="list"><el-icon><List /></el-icon></el-radio-button>
                   <el-radio-button label="grid"><el-icon><Grid /></el-icon></el-radio-button>
                </el-radio-group>
             </div>
          </div>
       </template>

       <!-- Grid View -->
       <div v-if="viewMode === 'grid'" class="pipeline-grid" v-loading="loading">
          <el-empty v-if="filteredPipelines.length === 0" description="暂无流水线" />
          <div v-for="item in filteredPipelines" :key="item.id" class="pipeline-card" @click="handleHistory(item)">
             <div class="card-header">
                <div class="pipeline-icon" :class="getStatusType(item.last_run_status)">
                   <el-icon><Cpu /></el-icon>
                </div>
                <div class="card-actions" @click.stop>
                   <el-dropdown trigger="click">
                      <el-icon class="more-btn"><MoreFilled /></el-icon>
                      <template #dropdown>
                       <el-dropdown-menu>
                           <el-dropdown-item @click="handleEdit(item)">编辑</el-dropdown-item>
                           <el-dropdown-item @click="handleDelete(item)" class="text-danger">删除</el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                   </el-dropdown>
                </div>
             </div>
             <div class="card-body">
                <h3 class="pipeline-title">{{ item.name }}</h3>
                <p class="pipeline-desc">{{ item.description || '暂无描述' }}</p>
                <div class="pipeline-meta">
                   <el-tag size="small" effect="plain">{{ item.branch || 'main' }}</el-tag>
                   <span class="meta-time">{{ formatTime(item.last_run_time) }}</span>
                </div>
             </div>
             <div class="card-footer">
                <div class="status-indicator">
                   <span class="status-dot" :class="getStatusClass(item.last_run_status)"></span>
                   <span class="status-text">{{ item.last_run_status || '未运行' }}</span>
                </div>
                <el-button type="primary" link size="small" @click.stop="handleRun(item)">
                   运行 <el-icon class="el-icon--right"><ArrowRight /></el-icon>
                </el-button>
             </div>
          </div>
       </div>

       <!-- List View -->
       <div v-else class="pipeline-list" v-loading="loading">
           <el-table :data="filteredPipelines" style="width: 100%">
             <el-table-column prop="name" label="流水线名称" min-width="200">
                <template #default="{ row }">
                   <div class="list-name-col">
                      <div class="list-icon" :class="getStatusType(row.last_run_status) + '-bg'">
                         <el-icon :class="getStatusType(row.last_run_status) + '-text'"><Cpu /></el-icon>
                      </div>
                      <div class="name-info">
                         <span class="name-text">{{ row.name }}</span>
                         <span class="desc-text">{{ row.description }}</span>
                      </div>
                   </div>
                </template>
             </el-table-column>
             <el-table-column prop="repository" label="代码仓库" min-width="200">
                <template #default="{ row }">
                   <div class="repo-info">
                      <el-icon><Connection /></el-icon>
                      <span>{{ row.gitUrl }}</span>
                   </div>
                </template>
             </el-table-column>
             <el-table-column prop="last_run" label="最近状态" width="150">
                 <template #default="{ row }">
                    <el-tag :type="getStatusType(row.last_run_status)" effect="light">
                       {{ row.last_run_status || '未运行' }}
                    </el-tag>
                 </template>
             </el-table-column>
             <el-table-column prop="updated_at" label="更新时间" width="180">
                <template #default="{ row }">
                   {{ formatTime(row.updated_at) }}
                </template>
             </el-table-column>
             <el-table-column label="操作" width="180" align="right">
                <template #default="{ row }">
                   <el-button-group>
                      <el-button size="small" @click="handleRun(row)"><el-icon><VideoPlay /></el-icon></el-button>
                      <el-button size="small" @click="handleHistory(row)"><el-icon><Timer /></el-icon></el-button>
                      <el-button size="small" @click="handleEdit(row)"><el-icon><Edit /></el-icon></el-button>
                      <el-button size="small" type="danger" plain @click="handleDelete(row)"><el-icon><Delete /></el-icon></el-button>
                   </el-button-group>
                </template>
             </el-table-column>
          </el-table>
       </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { getPipelines, triggerPipeline, deletePipeline } from '@/api/cicd.js'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import {
  Monitor, VideoPlay, TrendCharts, Search, List, Grid,
  Plus, Refresh, MoreFilled, Cpu, ArrowRight, Connection,
  Timer, Edit, Delete
} from '@element-plus/icons-vue'
import dayjs from 'dayjs'

const router = useRouter()
const viewMode = ref('grid')
const loading = ref(false)
const searchQuery = ref('')
const pipelines = ref([])

// Stats (Mocked for now)
const totalPipelines = ref(0)
const activeRuns = ref(0)
const successRate = ref(98)
const page = ref(1)
const pageSize = ref(20)

const fetchData = async () => {
    loading.value = true
    try {
        const res = await getPipelines({
            pageNum: page.value,
            pageSize: pageSize.value,
            keyword: searchQuery.value
        })
        console.log(res.data.data.data)
        pipelines.value = res.data.data.data || []
        totalPipelines.value = res.data.data.total || pipelines.value.length

        const activeCount = pipelines.value.filter(p => p.last_run_status === 'Running').length
        if (activeCount > 0) activeRuns.value = activeCount
    } catch (e) {
        ElMessage.error('加载流水线失败')
    } finally {
        loading.value = false
    }
}

const refreshData = () => {
   fetchData()
   ElMessage.success('数据已刷新')
}

const filteredPipelines = computed(() => {
    return pipelines.value
})

const formatTime = (time) => {
    if (!time) return '-'
    return dayjs(time).format('YYYY-MM-DD HH:mm')
}

const getStatusClass = (status) => {
    const map = {
        'Succeeded': 'status-success',
        'Failed': 'status-failed',
        'Running': 'status-running',
        'Pending': 'status-pending'
    }
    return map[status] || 'status-pending'
}

const getStatusType = (status) => {
    const map = {
        'Succeeded': 'success',
        'Failed': 'danger',
        'Running': 'warning',
        'Pending': 'info'
    }
    return map[status] || 'info'
}

const handleCreate = () => {
    router.push('/cicd/pipelines/create')
}

const handleEdit = (row) => {
    router.push(`/cicd/pipelines/${row.id}/edit`)
}

const handleRun = async (row) => {
    try {
       await triggerPipeline(row.id, {})
       ElNotification({
          title: '触发流水线',
          message: `流水线 ${row.name} 已成功开始运行。`,
          type: 'success',
          duration: 3000
       })
       fetchData()
    } catch(e) {
       console.error(e)
       ElMessage.error(e.response?.data?.message || '触发流水线失败')
    }
}
const handleHistory = (row) => {
    router.push(`/cicd/pipelines/${row.id}/runs`)
}

const handleDelete = async (row) => {
    try {
        await ElMessageBox.confirm(
            `确定要删除 ${row.name} 吗?`,
            '警告',
            {
                confirmButtonText: '删除',
                cancelButtonText: '取消',
                type: 'warning',
            }
        )
        await deletePipeline(row.id)
        ElMessage.success('流水线已删除')
        fetchData()
    } catch (e) {
        // Cancelled or error
    }
}

onMounted(fetchData)
</script>

<style scoped>
.pipeline-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 20px;
}

.pipeline-card {
    background: #fff;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    transition: all 0.3s;
    cursor: pointer;
}

.pipeline-card:hover {
    border-color: var(--primary-color);
    box-shadow: var(--shadow-hover);
    transform: translateY(-2px);
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
}

.pipeline-icon {
    width: 40px;
    height: 40px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;
}

.pipeline-icon.success { background: #e8f5e9; color: #4caf50; }
.pipeline-icon.danger { background: #ffebee; color: #f44336; }
.pipeline-icon.warning { background: #fff3e0; color: #ff9800; }
.pipeline-icon.info { background: #f5f7fa; color: #909399; }

.card-body { flex: 1; }
.pipeline-title { font-size: 16px; font-weight: 600; color: var(--text-main); margin: 0 0 8px 0; }
.pipeline-desc { font-size: 13px; color: var(--text-sub); margin: 0 0 16px 0; line-height: 1.4; height: 36px; overflow: hidden; }

.pipeline-meta { display: flex; align-items: center; gap: 12px; }
.meta-time { font-size: 12px; color: var(--text-sub); }

.card-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-top: 1px solid #f0f0f0;
    padding-top: 16px;
}

.status-indicator { display: flex; align-items: center; gap: 6px; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; }
.status-text { font-size: 12px; color: var(--text-sub); }

.status-success { background: #67c23a; }
.status-failed { background: #f56c6c; }
.status-running { background: #409eff; animation: pulse 1.5s infinite; }
.status-pending { background: #909399; }

@keyframes pulse {
    0% { box-shadow: 0 0 0 0 rgba(64, 158, 255, 0.4); }
    70% { box-shadow: 0 0 0 6px rgba(64, 158, 255, 0); }
    100% { box-shadow: 0 0 0 0 rgba(64, 158, 255, 0); }
}

.list-name-col { display: flex; align-items: center; gap: 12px; }
.list-icon {
    width: 36px; height: 36px; border-radius: 8px;
    display: flex; align-items: center; justify-content: center; font-size: 18px;
}
.list-icon.success-bg { background: #e8f5e9; }
.list-icon.danger-bg { background: #ffebee; }
.list-icon.warning-bg { background: #fff3e0; }
.list-icon.info-bg { background: #f5f7fa; }

.list-icon .success-text { color: #4caf50; }
.list-icon .danger-text { color: #f56c6c; }
.list-icon .warning-text { color: #ff9800; }
.list-icon .info-text { color: #909399; }

.name-info { display: flex; flex-direction: column; }
.name-text { font-weight: 500; color: var(--text-main); }
.desc-text { font-size: 12px; color: var(--text-sub); }
.repo-info { display: flex; align-items: center; gap: 6px; color: var(--text-sub); font-size: 13px; }
.text-danger { color: var(--danger-color); }
</style>
