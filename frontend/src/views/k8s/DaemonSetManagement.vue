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
        <el-button type="primary" class="btn-create" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:daemonset:showcreatedialogtrue')" >
          <el-icon class="mr-1"><Plus /></el-icon> 创建 DaemonSet
        </el-button>
      </div>
    </div>

    <!-- Table -->
    <div class="table-wrapper">
      <el-table 
        :data="daemonSetList" 
        style="width: 100%" class="autoops-table" 
        v-loading="loading"
        :header-cell-style="{ background: '#f5f7fa', color: '#606266', fontWeight: 'bold' }"
        row-class-name="daemonset-row"
      >
        <!-- Name Column -->
        <el-table-column label="名称" min-width="220">
          <template #default="{ row }">
            <div class="name-cell">
              <el-icon class="resource-icon daemonset-icon"><Files /></el-icon>
              <div class="name-content">
                <a class="name-link" @click="handleViewDetail(row)">{{ row.name }}</a>
                <span class="resource-type">DaemonSet</span>
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

        <!-- Pods Column (Desired/Current/Ready) -->
        <el-table-column label="容器组状态" min-width="160" align="center">
           <template #default="{ row }">
              <div class="pods-cell">
                 <div class="pods-count">
                    <el-icon class="monitor-icon"><Platform /></el-icon>
                    <span :class="{'text-green': row.ready === row.desired, 'text-warning': row.ready < row.desired}">
                        {{ row.ready || 0 }} / {{ row.desired || 0 }} (期望)
                    </span>
                 </div>
                 <div class="pods-status">
                     <span style="font-size: 11px; color: #909399;">
                        当前: {{ row.current || 0 }} | 可用: {{ row.available || 0 }}
                     </span>
                 </div>
              </div>
           </template>
        </el-table-column>

        <!-- Request/Limits (Reuse style) -->
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

        <!-- Image -->
        <el-table-column label="镜像" min-width="200">
           <template #default="{ row }">
              <div class="image-cell">
                 <el-icon class="image-icon"><Box /></el-icon>
                 <span class="image-name" :title="row.image">{{ row.image }}</span>
                 <span v-if="row.containers && row.containers.length > 1" class="more-images">
                    +{{ row.containers.length - 1 }}个镜像
                 </span>
              </div>
           </template>
        </el-table-column>

        <!-- Created Time -->
        <el-table-column label="创建时间" width="160">
            <template #default="{ row }">
               <span class="time-text">{{ formatDate(row.creationTimestamp || row.created) }}</span>
            </template>
        </el-table-column>

        <!-- Operations -->
        <el-table-column label="操作" width="240" fixed="right">
            <template #default="{ row }">
               <div class="ops-cell">
                  <el-tooltip content="编辑" placement="top">
                    <el-button circle size="small" type="primary" class="op-btn blue" @click="handleEdit(row)">
                       <el-icon><Edit /></el-icon>
                    </el-button>
                  </el-tooltip>
                  <el-button circle size="small" class="op-btn yellow" @click="handleViewDetail(row)" v-show="permStore.hasPerm('k8s:daemonset:handleviewdetail')" >
                       <el-icon><InfoFilled /></el-icon>
                    </el-button>
                  <el-tooltip content="Shell" placement="top">
                    <el-button circle size="small" class="op-btn gray" @click="handleShell(row)">
                       <el-icon><Monitor /></el-icon>
                    </el-button>
                  </el-tooltip>
                  <el-button circle size="small" type="danger" class="op-btn red" @click="handleDelete(row)" v-show="permStore.hasPerm('k8s:daemonset:handledelete')" >
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
    <el-dialog v-model="showCreateDialog" title="创建 DaemonSet" width="900px" :close-on-click-modal="false" append-to-body class="autoops-dialog">
        <el-form label-width="100px">
          <el-form-item label="命名空间">
            <el-select v-model="selectedNamespace" placeholder="选择命名空间" filterable style="width: 100%;">
              <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
            </el-select>
          </el-form-item>
          <el-form-item label="YAML配置">
            <div class="mb-2">
                 <el-button type="success" size="small" icon="CircleCheck" @click="handleValidateYAML" style="margin-bottom: 5px;">YAML 语法校验</el-button>
            </div>
            <el-input
              v-model="yamlContent"
              type="textarea"
              :rows="16"
              placeholder="请输入DaemonSet的YAML配置"
              style="font-family: 'Consolas', 'Monaco', monospace;"
            />
          </el-form-item>

        </el-form>
        <template #footer>
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="handleCreate">创建</el-button>
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
  Files, PriceTag, Platform, Box, 
  InfoFilled, Edit, Delete, CircleCheck 
} from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import { getDaemonSetList, deleteDaemonSet, createDaemonSet } from '@/api/k8s/daemonset'
import { getNamespaceList } from '@/api/k8s/namespace'
import dayjs from 'dayjs'
import { getSelectedInstanceId } from '@/stores/instanceStore'

const loading = ref(false)
const searchKeyword = ref('')
const daemonSetList = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const selectedNamespace = ref('default') 

const showCreateDialog = ref(false)
const yamlContent = ref(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: example-daemonset
spec:
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - name: example
        image: busybox
        command: ["sleep", "3600"]`)
const namespaceList = ref([])
const yamlError = ref('')
const isYamlValid = ref(true)

const formatResourceValue = (val) => {
    if (!val || val === '0' || val === '0m' || val === '0Mi') return '-'
    return val
}

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

const handleValidateYAML = () => {
    if (!yamlContent.value) {
        ElMessage.warning('请输入 YAML 内容')
        return
    }
    try {
        const docs = yaml.loadAll(yamlContent.value)
        ElMessage.success('YAML 语法校验通过 ✅')
    } catch (e) {
        console.error(e)
        ElMessageBox.alert(
            `<pre style="color: #F56C6C; max-height: 300px; overflow: auto;">${e.message}</pre>`, 
            '语法错误', 
            { dangerouslyUseHTMLString: true, type: 'error' }
        )
    }
}

const fetchData = async () => {
    loading.value = true
    try {
        const instanceId = getSelectedInstanceId()
        const res = await getDaemonSetList(selectedNamespace.value, instanceId)
        daemonSetList.value = res.data?.daemonSetList || []
        total.value = res.data?.total || daemonSetList.value.length
    } catch (e) {
        ElMessage.error('获取 DaemonSet 列表失败')
        daemonSetList.value = []
        total.value = 0
    } finally {
        loading.value = false
    }
}

const handleSearch = () => fetchData()
const handleReset = () => { searchKeyword.value = ''; fetchData() }
const handleSizeChange = () => fetchData()
const formatDate = (ts) => dayjs(ts).isValid() ? dayjs(ts).format('YYYY-MM-DD HH:mm:ss') : '-'

const handleViewDetail = (row) => ElMessage.info(`查看详情: ${row.name}`)
const handleEdit = (row) => ElMessage.info('编辑功能待完善')
const handleShell = (row) => ElMessage.info('Shell 功能待完善')
const handleCreate = async () => {
    if (!yamlContent.value.trim()) {
        ElMessage.warning('请输入YAML配置')
        return
    }
    const instanceId = getSelectedInstanceId()
    try {
        await createDaemonSet(selectedNamespace.value, { yaml: yamlContent.value }, instanceId)
        ElMessage.success('创建成功')
        showCreateDialog.value = false
        yamlContent.value = ''
        fetchData()
    } catch (error) {
        ElMessage.error(error.message || '创建失败')
    }
}
const handleDelete = (row) => {
    const instanceId = getSelectedInstanceId()
    ElMessageBox.confirm(`确定删除 ${row.name}?`, '警告', { type: 'warning' })
    .then(async () => {
         await deleteDaemonSet(row.namespace, row.name, instanceId)
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
.filter-label {
  font-size: 14px;
  color: #606266;
}
.filter-input {
  width: 240px;
}

.table-wrapper {
  background: white;
}

.name-cell {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}
.daemonset-icon {
  font-size: 24px;
  color: #7B1FA2; /* Purple/Violet */
  background: rgba(123, 31, 162, 0.1);
  padding: 6px;
  border-radius: 6px;
}
.name-content {
  display: flex;
  flex-direction: column;
}
.name-link {
  color: var(--primary-color);
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  font-size: 14px;
}
.name-link:hover { text-decoration: underline; }
.resource-type {
  font-size: 12px;
  color: #7B1FA2; 
  margin-top: 2px;
}

.label-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.label-icon { color: #909399; }
.label-count {
  cursor: pointer;
  background: #fef0f0;
  color: #f56c6c;
  border-color: #fde2e2;
}

.pods-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.pods-count {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
}
.monitor-icon { color: #67C23A; font-size: 16px; }
.text-green { color: #67C23A; }
.text-warning { color: #E6A23C; }

.resource-cell { font-size: 12px; font-family: monospace; }
.res-row { margin-bottom: 2px; }
.res-label { color: #909399; margin-right: 4px; }
.res-val.request { color: #67C23A; }
.res-val.limit { color: #E6A23C; }

.image-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.image-icon { color: #909399; }
.image-name {
  font-size: 13px;
  color: #606266;
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.more-images { font-size: 12px; color: var(--primary-color); }

.ops-cell {
  display: flex;
  gap: 8px;
  justify-content: flex-start;
}
.op-btn { border: none; font-size: 14px; }
.op-btn.blue { background: rgba(64, 158, 255, 0.1); color: #409EFF; }
/* ... reuse other matched op button styles from Deployment ... */
.op-btn.yellow { background: rgba(230, 162, 60, 0.1); color: #E6A23C; }
.op-btn.gray { background: rgba(144, 147, 153, 0.1); color: #909399; }
.op-btn.red { background: rgba(245, 108, 108, 0.1); color: #F56C6C; }

.pagination-container {
    padding-top: 20px;
    display: flex;
    justify-content: center;
}
</style>