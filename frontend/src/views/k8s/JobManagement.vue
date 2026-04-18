<template>
  <div class="autoops-container">
    <!-- Filter Bar -->
    <div class="filter-bar">
      <div class="filter-left">
        <label class="filter-label">命名空间</label>
        <el-select 
          v-model="selectedNamespace" 
          placeholder="选择命名空间" 
          class="filter-input"
          @change="fetchData"
          style="width: 180px;"
          filterable
        >
          <el-option 
            v-for="ns in namespaceList" 
            :key="ns.name" 
            :label="ns.name" 
            :value="ns.name" 
          />
          <el-option label="所有命名空间" value="all" />
        </el-select>
        <label class="filter-label">工作负载名称</label>
        <el-input 
          v-model="searchKeyword" 
          placeholder="请输入名称" 
          class="filter-input"
          clearable
          @keyup.enter="handleSearch"
        />
        <el-button type="primary" class="btn-search" @click="handleSearch">
          <el-icon class="mr-1"><Search /></el-icon> 搜索
        </el-button>
        <el-button class="btn-reset" @click="handleReset">
          <el-icon class="mr-1"><RefreshRight /></el-icon> 重置
        </el-button>
        <el-button type="success" class="btn-monitor">
          <el-icon class="mr-1"><Monitor /></el-icon> 监控仪表板
        </el-button>
        <el-button type="primary" class="btn-create" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:job:showcreatedialogtrue')" >
          <el-icon class="mr-1"><Plus /></el-icon> 创建 Job
        </el-button>
      </div>
    </div>

    <!-- Table -->
    <div class="table-wrapper">
      <el-table 
        :data="jobList" 
        style="width: 100%" class="autoops-table" 
        v-loading="loading"
        :header-cell-style="{ background: '#f5f7fa', color: '#606266', fontWeight: 'bold' }"
        row-class-name="job-row"
      >
        <!-- Name Column -->
        <el-table-column label="名称" min-width="220">
          <template #default="{ row }">
            <div class="name-cell">
              <el-icon class="resource-icon job-icon"><VideoPlay /></el-icon>
              <div class="name-content">
                <a class="name-link" @click="handleViewDetail(row)">{{ row.jobName || row.name }}</a>
                <span class="resource-type">Job</span>
              </div>
            </div>
          </template>
        </el-table-column>

        <!-- Labels Column -->
        <el-table-column label="标签" min-width="120">
          <template #default="{ row }">
             <div class="label-cell">
                <el-icon class="label-icon"><PriceTag /></el-icon>
                <el-popover placement="top" width="auto" trigger="hover">
                  <template #reference>
                    <el-tag type="danger" effect="plain" round size="small" class="label-count">
                       {{ Object.keys(row.labels || {}).length }}
                    </el-tag>
                  </template>
                  <div class="tags-popover">
                     <span v-for="(val, key) in row.labels" :key="key" class="tag-item">
                        {{ key }}: {{ val }}
                     </span>
                  </div>
                </el-popover>
             </div>
          </template>
        </el-table-column>

        <!-- Completions/Status -->
        <el-table-column label="完成情况" min-width="160" align="center">
           <template #default="{ row }">
              <div class="pods-cell">
                 <div class="pods-count">
                    <el-icon class="monitor-icon"><CircleCheck /></el-icon>
                    <span :class="{'text-green': parsePodsStatus(row.podsStatuses).succeeded > 0, 'text-warning': parsePodsStatus(row.podsStatuses).active > 0}">
                        {{ parsePodsStatus(row.podsStatuses).succeeded }} / {{ parsePodsStatus(row.podsStatuses).total }}
                    </span>
                 </div>
                 <div class="pods-status">
                    <el-tag size="small" :type="getJobStatusType(row)">
                       {{ getJobStatus(row) }}
                    </el-tag>
                 </div>
              </div>
           </template>
        </el-table-column>

        <!-- Duration -->
        <el-table-column label="持续时间" width="140">
           <template #default="{ row }">
              <span class="time-text">{{ formatDuration(row.startTime, row.endTime) }}</span>
           </template>
        </el-table-column>

         <!-- Image (Optional for Job, maybe less relevant but good for consistency) -->
         <el-table-column label="镜像" min-width="180">
            <template #default="{ row }">
               <div class="image-cell">
                  <el-icon class="image-icon"><Box /></el-icon>
                  <span class="image-name" :title="row.containerImage">{{ row.containerImage || '-' }}</span>
               </div>
            </template>
         </el-table-column>

        <!-- Request/Limits -->
        <el-table-column label="Request/Limits" min-width="180">
           <template #default="{ row }">
              <div class="resource-cell">
                 <div class="res-row">
                    <span class="res-label">CPU:</span>
                    <span class="res-val request">{{ formatResourceValue(row.resources?.cpuRequest) }}</span> / 
                    <span class="res-val limit">{{ formatResourceValue(row.resources?.cpuLimit) }}</span>
                 </div>
                 <div class="res-row">
                    <span class="res-label">Mem:</span>
                    <span class="res-val request">{{ formatResourceValue(row.resources?.memoryRequest) }}</span> / 
                    <span class="res-val limit">{{ formatResourceValue(row.resources?.memoryLimit) }}</span>
                 </div>
              </div>
           </template>
        </el-table-column>

        <!-- Created Time -->
        <el-table-column label="创建时间" width="160">
            <template #default="{ row }">
               <span class="time-text">{{ formatTime(row.startTime || row.creationTimestamp) }}</span>
            </template>
        </el-table-column>

        <!-- Operations -->
        <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
               <div class="ops-cell">

                  <el-tooltip content="详情" placement="top">
                    <el-button circle size="small" class="op-btn yellow" @click="handleViewDetail(row)">
                       <el-icon><InfoFilled /></el-icon>
                    </el-button>
                  </el-tooltip>
                  <el-button circle size="small" type="danger" class="op-btn red" @click="handleDelete(row)" v-show="permStore.hasPerm('k8s:job:handledelete')" >
                       <el-icon><Delete /></el-icon>
                    </el-button>
               </div>
            </template>
        </el-table-column>
      </el-table>
       <!-- Pagination -->
       <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="total"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @current-change="fetchData"
            @size-change="handleSizeChange"
          />
       </div>
    </div>

    <!-- Create Dialog -->
    <el-dialog 
      v-model="showCreateDialog" 
      title="创建 Job" 
      width="900px" 
      :close-on-click-modal="false" 
      append-to-body 
      class="autoops-dialog"
      @open="handleDialogOpen"
    >
        <el-form label-width="100px">
          <el-form-item label="命名空间" required>
            <el-select 
              v-model="selectedNamespace" 
              placeholder="选择命名空间" 
              style="width: 100%;"
              filterable
            >
              <el-option
                v-for="ns in namespaceList"
                :key="ns.name"
                :label="ns.name"
                :value="ns.name"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="YAML配置" required>
            <el-input
              v-model="yamlContent"
              type="textarea"
              :rows="18"
              placeholder="请输入Job的YAML配置"
              style="font-family: 'Consolas', 'Monaco', monospace;"
              @input="validateYaml"
            />
            <div style="margin-top: 8px; display: flex; gap: 8px;">
              <el-button size="small" @click="validateYaml">
                <el-icon><CircleCheck /></el-icon> 验证YAML
              </el-button>
              <el-tag v-if="yamlContent && isYamlValid" type="success" size="small">
                <el-icon><CircleCheck /></el-icon> 语法正确
              </el-tag>
            </div>
            <el-alert 
              v-if="!isYamlValid && yamlError" 
              title="YAML语法错误" 
              type="error" 
              :description="yamlError"
              :closable="false"
              style="margin-top: 8px;"
            />
          </el-form-item>
          <el-alert title="YAML示例" type="info" :closable="false">
            <div style="display: flex; justify-content: space-between; align-items: start;">
              <pre style="font-size: 12px; margin: 0; flex: 1;">apiVersion: batch/v1
kind: Job
metadata:
  name: example-job
spec:
  template:
    spec:
      containers:
      - name: example
        image: busybox
        command: ["echo", "Hello World"]
      restartPolicy: Never
  backoffLimit: 4</pre>
              <el-button 
                size="small" 
                type="primary" 
                link
                @click="loadExample"
                style="margin-left: 16px;"
              >
                加载示例
              </el-button>
            </div>
          </el-alert>
        </el-form>
        <template #footer>
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button 
            type="primary" 
            @click="handleCreate"
            :disabled="!isYamlValid || !yamlContent.trim()"
          >
            创建
          </el-button>
        </template>
    </el-dialog>

    <!-- Detail Dialog -->
    <el-dialog
      v-model="showDetailDialog"
      title="Job 详情"
      width="800px"
      append-to-body
      class="autoops-dialog"
    >
        <el-input
          v-model="detailYamlContent"
          type="textarea"
          :rows="20"
          readonly
          style="font-family: 'Consolas', 'Monaco', monospace;"
        />
    </el-dialog>

  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Search, RefreshRight, Monitor, Plus, 
  VideoPlay, PriceTag, CircleCheck, Box, 
  InfoFilled, Delete 
} from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import { getJobList, deleteJob, createJob, getJobDetail } from '@/api/k8s/job'
import { getNamespaceList } from '@/api/k8s/namespace'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'
import relativeTime from 'dayjs/plugin/relativeTime'
dayjs.extend(duration)
dayjs.extend(relativeTime)
import { getSelectedInstanceId } from '@/stores/instanceStore'

const loading = ref(false)
const searchKeyword = ref('')
const jobList = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const selectedNamespace = ref('default') 

const showCreateDialog = ref(false)
const yamlContent = ref('')
const namespaceList = ref([])
const yamlError = ref('')
const isYamlValid = ref(true)

const showDetailDialog = ref(false)
const detailYamlContent = ref('')

const fetchData = async () => {
    loading.value = true
    try {
        const instanceId = getSelectedInstanceId()
        const res = await getJobList(selectedNamespace.value, instanceId)
        console.log('后端返回的数据:', res.data)
        console.log('jobList:', res.data?.jobList)
        jobList.value = res.data?.jobList || []
        total.value = res.data?.total || jobList.value.length
    } catch (e) {
        console.error('获取Job列表失败:', e)
        ElMessage.error('获取 Job 列表失败')
        jobList.value = []
        total.value = 0
    } finally {
        loading.value = false
    }
}

const handleSearch = () => fetchData()
const handleReset = () => { searchKeyword.value = ''; fetchData() }
const handleSizeChange = () => fetchData()
const formatTime = (ts) => dayjs(ts).isValid() ? dayjs(ts).format('YYYY-MM-DD HH:mm:ss') : '-'
const formatDuration = (startTime, endTime) => {
    if (!startTime || !endTime) return '-'
    const start = dayjs(startTime)
    const end = dayjs(endTime)
    const seconds = end.diff(start, 'second')
    
    if (seconds < 60) return `${seconds}s`
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    return `${hours}h ${minutes}m`
}

const formatResourceValue = (val) => {
    if (!val || val === '0' || val === '0m' || val === '0Mi') return '-'
    return val
}

// 解析后端返回的Pod状态字符串
const parsePodsStatus = (statusStr) => {
  if (!statusStr || typeof statusStr !== 'string') return { active: 0, succeeded: 0, failed: 0, total: 1 }
  
  // 后端返回格式: "活跃: 0, 成功: 1, 失败: 0"
  const match = statusStr.match(/活跃:\s*(\d+),\s*成功:\s*(\d+),\s*失败:\s*(\d+)/)
  if (match) {
    const active = parseInt(match[1])
    const succeeded = parseInt(match[2])
    const failed = parseInt(match[3])
    return { active, succeeded, failed, total: Math.max(succeeded + failed + active, 1) }
  }
  return { active: 0, succeeded: 0, failed: 0, total: 1 }
}

const getJobStatusType = (row) => {
    const parsed = parsePodsStatus(row.podsStatuses)
    if (parsed.failed > 0) return 'danger'
    if (parsed.succeeded > 0 && parsed.active === 0 && parsed.failed === 0) return 'success'
    if (parsed.active > 0) return 'warning'
    return 'info'
}

const getJobStatus = (row) => {
    const parsed = parsePodsStatus(row.podsStatuses)
    if (parsed.active > 0) return 'Running'
    if (parsed.failed > 0) return 'Failed'
    if (parsed.succeeded > 0) return 'Completed'
    return 'Pending'
}

const handleViewDetail = async (row) => {
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getJobDetail(row.nameSpace || row.namespace, row.jobName || row.name, instanceId)
    if (res.data) {
       const detail = res.data.jobDetail || res.data
       detailYamlContent.value = yaml.dump(detail)
       showDetailDialog.value = true
    }
  } catch (e) {
    console.error('获取Job详情失败:', e)
    ElMessage.error('获取详情失败')
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除 Job "${row.jobName || row.name}" 吗？删除后无法恢复！`,
    '删除确认',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
      center: true
    }
  ).then(async () => {
    try {
      const instanceId = getSelectedInstanceId()
      await deleteJob(row.nameSpace || row.namespace, row.jobName || row.name, instanceId)
      ElMessage.success('删除成功')
      fetchData()
    } catch (error) {
      console.error('删除失败:', error)
      ElMessage.error('删除失败: ' + (error.message || '未知错误'))
    }
  }).catch(() => {
    // 用户取消删除
  })
}

// Dialog打开时加载命名空间
const handleDialogOpen = async () => {
  try {
    const instanceId = getSelectedInstanceId()
    const res = await getNamespaceList(instanceId)
    namespaceList.value = res.data?.namespaceList || []
    if (namespaceList.value.length > 0 && !selectedNamespace.value) {
      selectedNamespace.value = namespaceList.value[0].name
    }
  } catch (error) {
    ElMessage.error('获取命名空间列表失败')
    namespaceList.value = []
  }
}

// YAML语法验证
const validateYaml = () => {
  if (!yamlContent.value.trim()) {
    isYamlValid.value = true
    yamlError.value = ''
    return
  }
  
  try {
    yaml.load(yamlContent.value)
    isYamlValid.value = true
    yamlError.value = ''
  } catch (e) {
    isYamlValid.value = false
    yamlError.value = e.message
  }
}

// 加载示例YAML
const loadExample = () => {
  yamlContent.value = `apiVersion: batch/v1
kind: Job
metadata:
  name: example-job
spec:
  template:
    spec:
      containers:
      - name: example
        image: busybox
        command: ["echo", "Hello World"]
      restartPolicy: Never
  backoffLimit: 4`
  validateYaml()
}

const handleCreate = async () => {
    if (!yamlContent.value.trim()) {
        ElMessage.warning('请输入YAML配置')
        return
    }
    try {
        const instanceId = getSelectedInstanceId()
        await createJob(selectedNamespace.value, yamlContent.value, instanceId)
        ElMessage.success('创建成功')
        showCreateDialog.value = false
        yamlContent.value = ''
        fetchData()
    } catch (error) {
        ElMessage.error(error.message || '创建失败')
    }
}

onMounted(async () => {
    // 先获取命名空间列表
    try {
        const instanceId = getSelectedInstanceId()
        const res = await getNamespaceList(instanceId)
        namespaceList.value = res.data?.namespaceList || []
        if (namespaceList.value.length > 0 && !selectedNamespace.value) {
            selectedNamespace.value = namespaceList.value[0].name
        }
    } catch (error) {
        console.error('获取命名空间列表失败:', error)
    }
    
    // 然后获取Job列表
    fetchData()
})
</script>

<style scoped>
.filter-bar {
  padding: 16px 0;
  border-bottom: 1px solid #f0f2f5;
  margin-bottom: 16px;
}
.filter-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.filter-label { font-size: 14px; color: #606266; }
.filter-input { width: 240px; }

.table-wrapper { background: white; }

.name-cell { display: flex; align-items: flex-start; gap: 12px; }
.job-icon {
  font-size: 24px;
  color: #67C23A; /* Green for Job */
  background: rgba(103, 194, 58, 0.1);
  padding: 6px;
  border-radius: 6px;
}
.name-content { display: flex; flex-direction: column; }
.name-link { color: var(--primary-color); font-weight: 600; cursor: pointer; text-decoration: none; font-size: 14px; }
.resource-type { font-size: 12px; color: #67C23A; margin-top: 2px; }

.label-cell { display: flex; align-items: center; gap: 8px; }
.label-icon { color: #909399; }
.label-count { cursor: pointer; background: #fef0f0; color: #f56c6c; border-color: #fde2e2; }

.pods-cell { display: flex; flex-direction: column; align-items: center; }
.pods-count { display: flex; align-items: center; gap: 6px; font-size: 14px; font-weight: 600; }
.monitor-icon { color: #67C23A; font-size: 16px; }
.text-green { color: #67C23A; }
.text-warning { color: #E6A23C; }

.resource-cell { font-size: 12px; font-family: monospace; }
.res-row { margin-bottom: 2px; }
.res-label { color: #909399; margin-right: 4px; }
.res-val.request { color: #67C23A; }
.res-val.limit { color: #E6A23C; }

.image-cell { display: flex; align-items: center; gap: 8px; }
.image-icon { color: #909399; }
.image-name { font-size: 13px; color: #606266; max-width: 140px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.ops-cell { display: flex; gap: 8px; justify-content: flex-start; }
.op-btn { border: none; font-size: 14px; }
.op-btn.blue { background: rgba(64, 158, 255, 0.1); color: #409EFF; }
.op-btn.yellow { background: rgba(230, 162, 60, 0.1); color: #E6A23C; }
.op-btn.red { background: rgba(245, 108, 108, 0.1); color: #F56C6C; }

.pagination-container { padding-top: 20px; display: flex; justify-content: center; }
</style>