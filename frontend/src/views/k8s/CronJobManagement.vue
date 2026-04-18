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
          style="width: 180px;"
          @change="fetchData"
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
        <el-button type="primary" class="btn-create" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:cronjob:showcreatedialogtrue')" >
          <el-icon class="mr-1"><Plus /></el-icon> 创建 CronJob
        </el-button>
      </div>
    </div>

    <!-- Table -->
    <div class="table-wrapper">
      <el-table 
        :data="cronJobList" 
        style="width: 100%" class="autoops-table" 
        v-loading="loading"
        :header-cell-style="{ background: '#f5f7fa', color: '#606266', fontWeight: 'bold' }"
        row-class-name="cronjob-row"
      >
        <!-- Name Column -->
        <el-table-column label="名称" min-width="220">
          <template #default="{ row }">
            <div class="name-cell">
              <el-icon class="resource-icon cronjob-icon"><Timer /></el-icon>
              <div class="name-content">
                <a class="name-link" @click="handleViewDetail(row)">{{ row.name }}</a>
                <span class="resource-type">CronJob</span>
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

        <!-- Schedule -->
        <el-table-column label="调度规则" min-width="140">
           <template #default="{ row }">
              <span class="schedule-text">{{ row.schedule }}</span>
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

        <!-- Status -->
        <el-table-column label="状态" min-width="160">
           <template #default="{ row }">
              <div class="status-cell">
                 <el-tag size="small" :type="row.suspend ? 'warning' : 'success'">
                    {{ row.suspend ? 'Suspended' : 'Running' }}
                 </el-tag>
                 <el-tag size="small" type="info" v-if="row.active > 0" class="ml-2">
                    Active: {{ row.active }}
                 </el-tag>
              </div>
           </template>
        </el-table-column>

        <!-- Last Schedule -->
        <el-table-column label="上次调度" width="160">
           <template #default="{ row }">
              <span class="time-text">{{ row.lastScheduleTime ? formatTime(row.lastScheduleTime) : 'Never' }}</span>
           </template>
        </el-table-column>

        <!-- Created Time -->
        <el-table-column label="创建时间" width="160">
            <template #default="{ row }">
               <span class="time-text">{{ formatTime(row.creationTimestamp || row.created) }}</span>
            </template>
        </el-table-column>

        <!-- Operations -->
        <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
               <div class="ops-cell">
                  <el-tooltip content="编辑" placement="top">
                    <el-button circle size="small" type="primary" class="op-btn blue" @click="handleEdit(row)">
                       <el-icon><Edit /></el-icon>
                    </el-button>
                  </el-tooltip>
                  <el-button circle size="small" class="op-btn yellow" @click="handleViewDetail(row)" v-show="permStore.hasPerm('k8s:cronjob:handleviewdetail')" >
                       <el-icon><InfoFilled /></el-icon>
                    </el-button>
                  <el-button circle size="small" type="danger" class="op-btn red" @click="handleDelete(row)" v-show="permStore.hasPerm('k8s:cronjob:handledelete')" >
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
    <el-dialog v-model="showCreateDialog" title="创建 CronJob" width="900px" :close-on-click-modal="false" append-to-body class="autoops-dialog" @open="handleDialogOpen">
        <el-form label-width="100px">
          <el-form-item label="命名空间" required>
            <el-select v-model="selectedNamespace" placeholder="选择命名空间" style="width: 100%;" filterable>
              <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
            </el-select>
          </el-form-item>
          <el-form-item label="YAML配置" required>
            <el-input v-model="yamlContent" type="textarea" :rows="16" placeholder="请输入CronJob的YAML配置" style="font-family: 'Consolas', 'Monaco', monospace;" @input="validateYaml" />
            <el-alert v-if="!isYamlValid" title="YAML语法错误" type="error" :description="yamlError" :closable="false" style="margin-top: 8px;" />
          </el-form-item>

        </el-form>
        <template #footer>
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="handleCreate" :disabled="!isYamlValid || !yamlContent.trim()">创建</el-button>
        </template>
    </el-dialog>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetailDialog" title="CronJob 详情" width="800px" append-to-body>
        <el-tabs v-model="activeDetailTab">
            <el-tab-pane label="基本信息" name="info">
                <el-descriptions :column="2" border>
                    <el-descriptions-item label="名称">{{ detailContent.name }}</el-descriptions-item>
                    <el-descriptions-item label="命名空间">{{ detailContent.namespace }}</el-descriptions-item>
                    <el-descriptions-item label="状态">
                        <el-tag size="small" :type="detailContent.status === 'suspended' ? 'warning' : 'success'">
                            {{ detailContent.status }}
                        </el-tag>
                    </el-descriptions-item>
                    <el-descriptions-item label="创建时间">{{ formatTime(detailContent.age * 1000) }}</el-descriptions-item>
                    <el-descriptions-item label="调度规则">{{ detailContent.schedule }}</el-descriptions-item>
                    <el-descriptions-item label="镜像">{{ detailContent.image }}</el-descriptions-item>
                    <el-descriptions-item label="命令" :span="2">
                        <template v-if="detailContent.command && detailContent.command.length">
                            <el-tag v-for="(cmd, idx) in detailContent.command" :key="idx" size="small" style="margin-right: 5px;">{{ cmd }}</el-tag>
                        </template>
                        <span v-else>-</span>
                    </el-descriptions-item>
                    <el-descriptions-item label="标签" :span="2">
                         <el-tag v-for="(val, key) in detailContent.labels" :key="key" size="small" style="margin-right: 5px;">
                            {{ key }}: {{ val }}
                         </el-tag>
                    </el-descriptions-item>
                </el-descriptions>
            </el-tab-pane>
            <el-tab-pane label="YAML" name="yaml">
                 <el-input v-model="detailYamlContent" type="textarea" :rows="20" readonly style="font-family: 'Consolas', 'Monaco', monospace;" />
            </el-tab-pane>
        </el-tabs>
    </el-dialog>

    <!-- Edit Dialog -->
    <el-dialog v-model="showEditDialog" title="编辑 CronJob YAML" width="900px" :close-on-click-modal="false" append-to-body>
        <el-input v-model="editYamlContent" type="textarea" :rows="20" placeholder="请输入YAML" style="font-family: 'Consolas', 'Monaco', monospace;" />
        <template #footer>
           <el-button @click="showEditDialog = false">取消</el-button>
           <el-button type="primary" @click="handleUpdate">更新</el-button>
        </template>
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
  Timer, PriceTag, InfoFilled, Edit, Delete 
} from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import { getCronJobList, deleteCronJob, createCronJob, getCronJobDetail, getCronJobYAML, updateCronJobYAML } from '@/api/k8s/cronjob'
import { getNamespaceList } from '@/api/k8s/namespace'
import dayjs from 'dayjs'
import { getSelectedInstanceId } from '@/stores/instanceStore'

const loading = ref(false)
const searchKeyword = ref('')
const cronJobList = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const selectedNamespace = ref('default') 

const showCreateDialog = ref(false)
const yamlContent = ref('')
const namespaceList = ref([])
const yamlError = ref('')
const isYamlValid = ref(true)

const fetchNamespaces = async () => {
    try {
        const instanceId = getSelectedInstanceId()
        const res = await getNamespaceList(instanceId)
        const list = res.data?.items || res.data?.namespaceList || res.data || []
        namespaceList.value = list.map(item => ({ name: typeof item === 'string' ? item : item.name }))
    } catch (e) {
        console.error('Fetch namespaces failed', e)
    }
}

// Detail Dialog
const showDetailDialog = ref(false)
const detailContent = ref({})
const detailYamlContent = ref('')
const activeDetailTab = ref('info')

// Edit Dialog
const showEditDialog = ref(false)
const editYamlContent = ref('')
const editingCronJob = ref(null)

const fetchData = async () => {
    loading.value = true
    try {
        const instanceId = getSelectedInstanceId()
        const res = await getCronJobList(selectedNamespace.value, instanceId)
        cronJobList.value = res.data?.cronJobList || []
        total.value = res.data?.total || cronJobList.value.length
    } catch (e) {
        ElMessage.error('获取 CronJob 列表失败')
        cronJobList.value = []
        total.value = 0
    } finally {
        loading.value = false
    }
}

const handleSearch = () => fetchData()
const handleReset = () => { searchKeyword.value = ''; fetchData() }
const handleSizeChange = () => fetchData()
const formatTime = (ts) => dayjs(ts).isValid() ? dayjs(ts).format('YYYY-MM-DD HH:mm:ss') : '-'

const formatResourceValue = (val) => {
    if (!val || val === '0' || val === '0m' || val === '0Mi') return '-'
    return val
}

const handleViewDetail = async (row) => {
    try {
        const instanceId = getSelectedInstanceId()
        const res = await getCronJobDetail(row.namespace, row.name, instanceId)
        if (res.data) {
             detailContent.value = res.data.cronJobDetail
             // Also fetch YAML for the YAML tab
             const yamlRes = await getCronJobYAML(row.namespace, row.name, instanceId)
             if (yamlRes.data) {
                 detailYamlContent.value = yamlRes.data.yaml
             }
             showDetailDialog.value = true
        }
    } catch (e) {
        ElMessage.error('获取详情失败: ' + e.message)
    }
}

const handleEdit = async (row) => {
    try {
        const instanceId = getSelectedInstanceId()
        const res = await getCronJobYAML(row.namespace, row.name, instanceId)
        if (res.data) {
            editYamlContent.value = res.data.yaml
            editingCronJob.value = row
            showEditDialog.value = true
        }
    } catch (e) {
        ElMessage.error('获取YAML失败: ' + e.message)
    }
}

const handleUpdate = async () => {
    if (!editYamlContent.value.trim()) return
    try {
        const instanceId = getSelectedInstanceId()
        // Determine namespace from editingCronJob or parse it? 
        // Best to use the original namespace as param, but YAML might have changed it. 
        // The API uses URL param for namespace, so we use the original one.
        await updateCronJobYAML(editingCronJob.value.namespace, editYamlContent.value, instanceId)
        ElMessage.success('更新成功')
        showEditDialog.value = false
        fetchData()
    } catch (e) {
        ElMessage.error('更新失败: ' + e.message)
    }
}

const handleDialogOpen = async () => {
  // Use existing namespaceList if available, or fetch if empty (though onMounted should have fetched it)
  if (namespaceList.value.length === 0) {
      await fetchNamespaces()
  }
  
  if (namespaceList.value.length > 0 && !selectedNamespace.value) {
      // If "all" is selected, default to first or "default"
      selectedNamespace.value = selectedNamespace.value === 'all' ? namespaceList.value[0].name : selectedNamespace.value
  }

  // Auto-fill template if empty
  if (!yamlContent.value.trim()) {
      yamlContent.value = `apiVersion: batch/v1
kind: CronJob
metadata:
  name: example-cronjob
spec:
  schedule: "*/5 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: hello
            image: busybox
            command: ["echo", "Hello"]
          restartPolicy: Never`
      validateYaml()
  }
}

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



const handleCreate = async () => {
    if (!yamlContent.value.trim()) {
        ElMessage.warning('请输入YAML配置')
        return
    }
    try {
        const instanceId = getSelectedInstanceId()
        await createCronJob(selectedNamespace.value, yamlContent.value, instanceId)
        ElMessage.success('创建成功')
        showCreateDialog.value = false
        yamlContent.value = ''
        fetchData()
    } catch (error) {
        ElMessage.error(error.message || '创建失败')
    }
}
const handleDelete = (row) => {
    ElMessageBox.confirm(`确定删除 ${row.name}?`, '警告', { type: 'warning' })
    .then(async () => {
         const instanceId = getSelectedInstanceId()
         await deleteCronJob(row.namespace, row.name, instanceId)
         ElMessage.success('删除成功')
         fetchData()
    })
}

onMounted(() => {
    fetchData()
    fetchNamespaces()
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
.cronjob-icon {
  font-size: 24px;
  color: #F56C6C; /* Red/Orange for Cron */
  background: rgba(245, 108, 108, 0.1);
  padding: 6px;
  border-radius: 6px;
}
.name-content { display: flex; flex-direction: column; }
.name-link { color: var(--primary-color); font-weight: 600; cursor: pointer; text-decoration: none; font-size: 14px; }
.resource-type { font-size: 12px; color: #F56C6C; margin-top: 2px; }

.label-cell { display: flex; align-items: center; gap: 8px; }
.label-icon { color: #909399; }
.label-count { cursor: pointer; background: #fef0f0; color: #f56c6c; border-color: #fde2e2; }

.schedule-text { font-family: monospace; font-weight: 600; color: #606266; }

.resource-cell { font-size: 12px; font-family: monospace; }
.res-row { margin-bottom: 2px; }
.res-label { color: #909399; margin-right: 4px; }
.res-val.request { color: #67C23A; }
.res-val.limit { color: #E6A23C; }

.status-cell { display: flex; align-items: center; }

.ops-cell { display: flex; gap: 8px; justify-content: flex-start; }
.op-btn { border: none; font-size: 14px; }
.op-btn.blue { background: rgba(64, 158, 255, 0.1); color: #409EFF; }
.op-btn.yellow { background: rgba(230, 162, 60, 0.1); color: #E6A23C; }
.op-btn.red { background: rgba(245, 108, 108, 0.1); color: #F56C6C; }

.pagination-container { padding-top: 20px; display: flex; justify-content: center; }
</style>