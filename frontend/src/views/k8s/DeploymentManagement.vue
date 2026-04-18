<template>
  <div class="autoops-container">
    <!-- Filter Bar -->
    <div class="autoops-filter-bar">
      <label class="autoops-filter-label">工作负载名称</label>
      <el-input 
        v-model="searchKeyword" 
        placeholder="请输入名称" 
        class="autoops-input"
        style="width: 240px"
        clearable
        @keyup.enter="handleSearch"
      />
      <el-button type="primary" class="autoops-btn-primary" @click="handleSearch">
        <el-icon><Search /></el-icon>
        <span>搜索</span>
      </el-button>
      <el-button class="autoops-btn-secondary" @click="handleReset">
        <el-icon><RefreshRight /></el-icon>
        <span>重置</span>
      </el-button>
      <el-button type="success" class="autoops-btn-success">
        <el-icon><Monitor /></el-icon>
        <span>监控仪表板</span>
      </el-button>
      <el-button type="primary" class="autoops-btn-primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:deployment:showcreatedialogtrue')">
        <el-icon><Plus /></el-icon>
        <span>创建工作负载</span>
      </el-button>
    </div>

    <!-- Table -->
    <div class="autoops-table-wrapper">
      <el-table 
        :data="deploymentList" 
        class="autoops-table" style="width: 100%" 
        v-loading="loading"
      >
        <!-- Name Column -->
        <el-table-column label="名称" min-width="220">
          <template #default="{ row }">
            <div class="name-cell">
              <div class="name-content">
                <a class="name-link" @click="handleViewDetail(row)">{{ row.name }}</a>
                <span class="resource-type">Deployment</span>
              </div>
            </div>
          </template>
        </el-table-column>

        <!-- Labels Column -->
        <el-table-column label="标签" min-width="80">
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

        <!-- Pods Column -->
        <el-table-column label="容器组数量" min-width="110" align="center">
           <template #default="{ row }">
              <div class="pods-cell">
                 <div class="pods-count">
                    <el-icon class="monitor-icon"><Platform /></el-icon>
                    <span :class="{'text-green': row.ready === row.replicas, 'text-warning': row.ready < row.replicas}">
                        {{ row.ready || 0 }}/{{ row.replicas || 0 }}
                    </span>
                 </div>
                 <div class="pods-status" :class="{'status-ready': row.ready === row.replicas, 'status-warn': row.ready < row.replicas}">
                    {{ row.ready === row.replicas ? '全部就绪' : '部分就绪' }}
                 </div>
              </div>
           </template>
        </el-table-column>

        <!-- Request/Limits -->
        <el-table-column label="Request/Limits" min-width="160">
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
        <el-table-column label="镜像" min-width="180">
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
        <el-table-column label="创建时间" width="175">
            <template #default="{ row }">
               <span class="time-text">{{ formatDate(row.created) }}</span>
            </template>
        </el-table-column>

        <!-- Operations -->
        <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
               <div class="autoops-actions">
                  <el-button link type="primary" size="small" @click="handleScale(row)">扩缩容</el-button>
                  <el-button link type="primary" size="small" @click="handleViewDetail(row)" v-show="permStore.hasPerm('k8s:deployment:handleviewdetail')">详情</el-button>
                  <el-button link type="primary" size="small" @click="handleUpdate(row)">更新</el-button>
                  <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, row)">
                    <span class="el-dropdown-link">
                      更多<el-icon class="el-icon--right"><arrow-down /></el-icon>
                    </span>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item command="shell">Shell</el-dropdown-item>
                        <el-dropdown-item command="logs">日志</el-dropdown-item>
                        <el-dropdown-item command="delete" divided style="color: #F56C6C;">删除</el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
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

    <!-- Keep Existing Dialogs -->
    <el-dialog v-model="showCreateDialog" title="创建 Deployment" width="900px" :close-on-click-modal="false" append-to-body>
      <!-- ... (Keep existing Create Content but simplify or reuse logic) ... -->
      <div style="padding: 20px;">
          <el-tabs v-model="activeTab" type="border-card">
              <el-tab-pane label="表单创建" name="form">
                  <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="120px">
                     <el-form-item label="名称" prop="name"><el-input v-model="createForm.name" /></el-form-item>
                     <el-form-item label="命名空间" prop="namespace">
                        <el-select v-model="createForm.namespace" placeholder="请选择命名空间" filterable style="width: 100%">
                           <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
                        </el-select>
                     </el-form-item>
                     <el-form-item label="镜像" prop="image"><el-input v-model="createForm.image" /></el-form-item>
                     <el-form-item label="副本数" prop="replicas"><el-input-number v-model="createForm.replicas" /></el-form-item>
                  </el-form>
              </el-tab-pane>
              <el-tab-pane label="YAML" name="yaml">
                 <div class="mb-2">
                    <el-button type="success" size="small" icon="CircleCheck" @click="handleValidateYAML">YAML 语法校验</el-button>
                 </div>
                 <el-input type="textarea" v-model="yamlContent" :rows="15" placeholder="在此粘贴 YAML 配置..." style="font-family: monospace;" />
              </el-tab-pane>
          </el-tabs>
      </div>
      <template #footer>
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- Scale Dialog -->
    <el-dialog v-model="showScaleDialog" title="扩缩容" width="400px" append-to-body>
      <div class="p-4">
         <p class="mb-2">目标副本数</p>
         <el-input-number v-model="scaleForm.replicas" :min="0" :max="50" class="w-full" />
      </div>
      <template #footer>
        <el-button @click="showScaleDialog = false">取消</el-button>
        <el-button type="primary" @click="handleScaleConfirm">确定</el-button>
      </template>
    </el-dialog>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetailDialog" :title="detailTitle" width="800px" append-to-body>
        <el-input v-model="detailContent" type="textarea" :rows="20" readonly style="font-family: monospace;" />
    </el-dialog>

    <!-- Update Image Dialog -->
    <el-dialog v-model="showUpdateDialog" title="更新镜像" width="500px" append-to-body>
        <el-form :model="updateForm" label-width="100px">
            <el-form-item label="当前镜像">
                <el-input v-model="updateForm.currentImage" disabled />
            </el-form-item>
            <el-form-item label="新镜像">
                <el-input v-model="updateForm.newImage" placeholder="例如: nginx:1.19.0" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="showUpdateDialog = false">取消</el-button>
            <el-button type="primary" @click="confirmUpdate">更新</el-button>
        </template>
    </el-dialog>

    <!-- Logs Dialog -->
    <el-dialog v-model="showLogsDialog" title="日志查看" width="80%" top="5vh" append-to-body destroy-on-close>
        <div class="mb-4 flex items-center gap-4">
            <el-select v-model="selectedLogPod" placeholder="选择 Pod" @change="fetchLogs" style="width: 300px;">
                <el-option v-for="pod in logPodList" :key="pod.name" :label="pod.name" :value="pod.name" />
            </el-select>
            <el-button @click="fetchLogs" :loading="logLoading" icon="RefreshRight" circle />
        </div>
        <div class="log-container">
            <pre>{{ logContent }}</pre>
        </div>
    </el-dialog>

    <!-- 容器选择对话框 (For Shell) -->
    <el-dialog v-model="showContainerSelectDialog" title="选择容器连接 Shell" width="500px" append-to-body>
      <el-form label-width="80px">
        <el-form-item label="Pod">
           <el-select v-model="selectedShellPodName" placeholder="请选择 Pod" style="width: 100%;" @change="handleShellPodChange">
             <el-option v-for="pod in shellPodList" :key="pod.name" :label="pod.name" :value="pod.name" />
           </el-select>
        </el-form-item>
        <el-form-item label="容器">
           <el-select v-model="selectedContainerForTerminal" placeholder="请选择容器" style="width: 100%;">
             <el-option v-for="c in shellContainerList" :key="c" :label="c" :value="c" />
           </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showContainerSelectDialog = false">取消</el-button>
          <el-button type="primary" @click="confirmContainerSelect" :disabled="!selectedShellPodName || !selectedContainerForTerminal">连接</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 容器终端 -->
    <PodTerminal
      v-model="showTerminalDialog"
      :namespace="terminalPod.namespace"
      :pod-name="terminalPod.podName"
      :container-name="terminalPod.containerName"
      :instance-id="getSelectedInstanceId()"
      @close="handleTerminalClose"
    />

  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Search, RefreshRight, Monitor, Plus, 
  Aim, PriceTag, Platform, Box, Cloudy, 
  InfoFilled, Setting, Document, Delete, ArrowDown,
  CircleCheck
} from '@element-plus/icons-vue'
import { getDeploymentList, scaleDeployment, deleteDeployment, createDeployment, getDeploymentDetail, updateDeployment } from '@/api/k8s/deployment'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getPodList, getPodLogs } from '@/api/k8s/pod'
import dayjs from 'dayjs'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import yaml from 'js-yaml'
import PodTerminal from '@/components/PodTerminal.vue'

const loading = ref(false)
const searchKeyword = ref('')
const deploymentList = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const selectedNamespace = ref('default') // or 'all'
const namespaceList = ref([])

const showCreateDialog = ref(false)
const showScaleDialog = ref(false)
const activeTab = ref('form')
const createForm = reactive({ name: '', namespace: 'default', image: 'nginx:latest', replicas: 1 })
const yamlContent = ref('')
const scaleForm = reactive({ replicas: 1, row: null })

// Detail Dialog
const showDetailDialog = ref(false)
const detailContent = ref('')
const detailTitle = ref('')

// Update Image Dialog
const showUpdateDialog = ref(false)
const updateForm = reactive({ currentImage: '', newImage: '', row: null, originalData: null })

// Logs Dialog
const showLogsDialog = ref(false)
const logContent = ref('')
const logPodList = ref([])
const selectedLogPod = ref('')
const logLoading = ref(false)

// Shell related
const showTerminalDialog = ref(false)
const terminalPod = ref({ namespace: '', podName: '', containerName: '' })
const showContainerSelectDialog = ref(false)
const shellPodList = ref([])
const selectedShellPodName = ref('')
const shellContainerList = ref([])
const selectedContainerForTerminal = ref('')

const fetchNamespaces = async () => {
    try {
        const instanceId = getSelectedInstanceId()
        const res = await getNamespaceList(instanceId)
        // Adjust based on actual API response structure
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
        const res = await getDeploymentList(selectedNamespace.value, instanceId)
        deploymentList.value = res.data?.deploymentList || []
        total.value = res.data?.total || deploymentList.value.length
    } catch (e) {
        ElMessage.error('获取 Deployment 列表失败')
        deploymentList.value = []
        total.value = 0
    } finally {
        loading.value = false
    }
}

const handleSearch = () => fetchData()
const handleReset = () => { searchKeyword.value = ''; fetchData() }
const handleSizeChange = () => fetchData()
const formatDate = (ts) => {
    if (!ts) return '-'
    // Backend returns ISO string often, or check if unix
    return dayjs(ts).isValid() ? dayjs(ts).format('YYYY-MM-DD HH:mm:ss') : '-'
}

const formatResourceValue = (val) => {
    if (!val || val === '0' || val === '0m' || val === '0Mi') return '-'
    return val
}

const handleViewDetail = async (row) => {
    try {
        const instanceId = getSelectedInstanceId()
        const res = await getDeploymentDetail(row.namespace, row.name, instanceId)
        if (res.data) {
            // Backend returns { deploymentDetail: { ... } }
            const detail = res.data.deploymentDetail || res.data
            detailContent.value = yaml.dump(detail)
            detailTitle.value = `详情: ${row.name}`
            showDetailDialog.value = true
        }
    } catch(e) { ElMessage.error('获取详情失败') }
}

const handleScale = (row) => { 
    scaleForm.row = row; 
    scaleForm.replicas = row.replicas; 
    showScaleDialog.value = true; 
}

const handleScaleConfirm = async () => {
    try {
        const instanceId = getSelectedInstanceId()
        await scaleDeployment(scaleForm.row.namespace, scaleForm.row.name, scaleForm.replicas, instanceId)
        ElMessage.success('扩缩容指令已发送')
        showScaleDialog.value = false
        fetchData()
        // K8s is async, fetch again after a short delay to show progress
        setTimeout(() => fetchData(), 2000)
    } catch(e) { 
        console.error(e)
        // Log detailed error
        ElMessage.error('操作失败: ' + (e.response?.data?.message || e.message || '未知错误')) 
    }
}

const handleUpdate = async (row) => {
    try {
        const instanceId = getSelectedInstanceId()
        const res = await getDeploymentDetail(row.namespace, row.name, instanceId)
        if (res.data) {
            const detail = res.data.deploymentDetail || res.data
            updateForm.originalData = detail
            updateForm.row = row
            // Extract image from containers returned by backend
            const containers = detail.containers || []
            updateForm.currentImage = containers.length > 0 ? containers[0].image : ''
            updateForm.newImage = updateForm.currentImage // Pre-fill new image with current
            showUpdateDialog.value = true
        }
    } catch(e) { ElMessage.error('获取详情失败') }
}

const confirmUpdate = async () => {
    try {
        const instanceId = getSelectedInstanceId()
        const data = JSON.parse(JSON.stringify(updateForm.originalData))
        if (data.spec?.template?.spec?.containers?.length > 0) {
            data.spec.template.spec.containers[0].image = updateForm.newImage
            await updateDeployment(updateForm.row.namespace, updateForm.row.name, data, instanceId)
            ElMessage.success('更新镜像指令已发送')
            showUpdateDialog.value = false
            fetchData()
        } else {
            ElMessage.error('无法解析容器信息，请检查后端返回数据')
        }
    } catch(e) { ElMessage.error('更新失败') }
}

const getPodsBySelector = async (row) => {
    const instanceId = getSelectedInstanceId()
    const detailRes = await getDeploymentDetail(row.namespace, row.name, instanceId)
    const detail = detailRes.data?.deploymentDetail || detailRes.data
    const selector = detail?.selector
    if (!selector) { throw new Error('无法获取 Pod 选择器') }

    const podsRes = await getPodList(row.namespace, instanceId)
    const allPods = podsRes.data?.items || podsRes.data?.podList || []
    
    return allPods.filter(pod => {
        const podLabels = pod.labels || {}
        return Object.entries(selector).every(([k, v]) => podLabels[k] === v)
    })
}

const handleShell = async (row) => {
    try {
        const pods = await getPodsBySelector(row)
        if (pods.length === 0) {
            ElMessageBox.alert('未找到属于该 Deployment 的 Pod。可能副本数为 0，或者 Pod 尚未创建。', '提示', {
                confirmButtonText: '确定',
                type: 'warning'
            })
            return
        }
        
        // Populate shell selection list
        shellPodList.value = pods
        
        // If only 1 pod and 1 container, connect directly
        if (pods.length === 1 && pods[0].containers && pods[0].containers.length === 1) {
            terminalPod.value = {
                namespace: row.namespace,
                podName: pods[0].name,
                containerName: pods[0].containers[0].name
            }
            showTerminalDialog.value = true
            return
        }

        // Else, show selection dialog
        selectedShellPodName.value = pods[0].name
        handleShellPodChange(pods[0].name) // init containers
        showContainerSelectDialog.value = true
    } catch(e) {
        console.error(e)
        ElMessage.error('准备 Shell 环境失败: ' + e.message)
    }
}

const handleShellPodChange = (podName) => {
    const pod = shellPodList.value.find(p => p.name === podName)
    if (pod && pod.containers) {
        shellContainerList.value = pod.containers.map(c => c.name)
        selectedContainerForTerminal.value = shellContainerList.value[0] || ''
    } else {
        shellContainerList.value = []
        selectedContainerForTerminal.value = ''
    }
}

const confirmContainerSelect = () => {
    const pod = shellPodList.value.find(p => p.name === selectedShellPodName.value)
    terminalPod.value = {
        namespace: pod.namespace, // Use pod namespace to be safe
        podName: selectedShellPodName.value,
        containerName: selectedContainerForTerminal.value
    }
    showContainerSelectDialog.value = false
    showTerminalDialog.value = true
}

const handleTerminalClose = () => {
    showTerminalDialog.value = false
    // cleanup if needed
}

const handleCommand = (cmd, row) => {
    if (cmd === 'shell') handleShell(row)
    if (cmd === 'logs') handleLogs(row)
    if (cmd === 'delete') handleDelete(row)
}

const handleLogs = async (row) => {
    try {
        logLoading.value = true
        const pods = await getPodsBySelector(row)
        logPodList.value = pods

        if (logPodList.value.length === 0) {
            ElMessageBox.alert('未找到属于该 Deployment 的 Pod。可能副本数为 0，或者 Pod 尚未创建。', '提示', {
                confirmButtonText: '确定',
                type: 'warning'
            })
            return
        }

        selectedLogPod.value = logPodList.value[0].name
        showLogsDialog.value = true
        fetchLogs()
    } catch(e) { 
        console.error(e)
        ElMessage.error('获取 Pod 日志失败') 
    } finally {
        logLoading.value = false
    }
}

const fetchLogs = async () => {
    if (!selectedLogPod.value) return
    try {
        const instanceId = getSelectedInstanceId()
        const row = logPodList.value.find(p => p.name === selectedLogPod.value)
        const container = row?.containers ? row.containers[0].name : '' 
        const res = await getPodLogs(row.namespace, selectedLogPod.value, container, instanceId, { tailLines: 500 })
        logContent.value = typeof res.data === 'string' ? res.data : JSON.stringify(res.data, null, 2)
    } catch(e) {
        logContent.value = 'Fetch logs failed.'
    }
}
const handleDelete = (row) => {
    ElMessageBox.confirm(`确定删除 ${row.name}?`, '警告', { type: 'warning' })
    .then(async () => {
         await deleteDeployment(row.namespace, row.name)
         ElMessage.success('删除成功')
         fetchData()
    })
}

const createFormRef = ref(null)
const createRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  namespace: [{ required: true, message: '请输入命名空间', trigger: 'blur' }],
  image: [{ required: true, message: '请输入镜像', trigger: 'blur' }],
}

const handleCreate = async () => {
    if (!createFormRef.value) return
    
    await createFormRef.value.validate(async (valid) => {
        if (valid) {
            try {
                const instanceId = getSelectedInstanceId()
                // Construct payload matching backend 
                const data = {
                    name: createForm.name,
                    namespace: createForm.namespace,
                    image: createForm.image,
                    replicas: createForm.replicas,
                    // port and labels are optional in backend
                }
                
                // Note: namespace is passed in URL in API call usually, but let's check API def
                // api: createDeployment(namespace, data, instanceId)
                await createDeployment(createForm.namespace, data, instanceId)
                
                ElMessage.success('创建成功')
                showCreateDialog.value = false
                fetchData()
            } catch (e) {
                console.error(e)
                ElMessage.error('创建失败: ' + (e.response?.data?.message || e.message || '未知错误'))
            }
        }
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

/* Name Column */
.name-cell {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}
.deployment-icon {
  font-size: 24px;
  color: var(--primary-color);
  background: rgba(64, 158, 255, 0.1);
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
  color: #E6A23C;
  margin-top: 2px;
}

/* Labels */
.label-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.label-icon {
  color: #909399;
}
.label-count {
  cursor: pointer;
  background: #fef0f0;
  color: #f56c6c;
  border-color: #fde2e2;
}

/* Pods */
.pods-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.pods-count {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 15px;
  font-weight: 600;
}
.monitor-icon {
  color: #67C23A;
  font-size: 16px;
}
.pods-status {
  font-size: 11px;
  margin-top: 2px;
}
.status-ready { color: #67C23A; }
.status-warn { color: #E6A23C; }
.text-green { color: #67C23A; }
.text-warning { color: #E6A23C; }

/* Resource */
.resource-cell {
  font-size: 12px;
  font-family: monospace;
}
.res-row { margin-bottom: 2px; }
.res-label { color: #909399; margin-right: 4px; }
.res-val.request { color: #67C23A; }
.res-val.limit { color: #E6A23C; }

/* Actions */
.autoops-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: nowrap;
}
.el-dropdown-link {
  cursor: pointer;
  color: var(--primary-color);
  display: flex;
  align-items: center;
  font-size: 12px;
  margin-left: 4px;
}

/* Image */
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
.more-images {
  font-size: 12px;
  color: var(--primary-color);
}

.log-container {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 12px;
  border-radius: 4px;
  height: 500px;
  overflow-y: auto;
  font-family: monospace;
  white-space: pre-wrap;
  font-size: 12px;
}

/* Ops */
.ops-cell {
  display: flex;
  gap: 8px;
  justify-content: flex-start;
}
.op-btn {
  border: none;
  font-size: 14px;
}
.op-btn.blue { background: rgba(64, 158, 255, 0.1); color: #409EFF; }
.op-btn.blue:hover { background: #409EFF; color: white; }

.op-btn.yellow { background: rgba(230, 162, 60, 0.1); color: #E6A23C; }
.op-btn.yellow:hover { background: #E6A23C; color: white; }

.op-btn.green { background: rgba(103, 194, 58, 0.1); color: #67C23A; }
.op-btn.green:hover { background: #67C23A; color: white; }

.op-btn.gray { background: rgba(144, 147, 153, 0.1); color: #909399; }
.op-btn.gray:hover { background: #909399; color: white; }

.op-btn.blue-light { background: #ecf5ff; color: #409EFF; }
.op-btn.blue-light:hover { background: #409EFF; color: white; }

.op-btn.red { background: rgba(245, 108, 108, 0.1); color: #F56C6C; }
.op-btn.red:hover { background: #F56C6C; color: white; }

.pagination-container {
    padding-top: 20px;
    display: flex;
    justify-content: center;
}
</style>
