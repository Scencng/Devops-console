<template>
  <div class="page-container">
    <el-card class="page-header-card">
       <div class="page-header">
         <div class="header-left">
           <div class="header-icon-wrapper">
             <el-icon :size="24"><Box /></el-icon>
           </div>
           <div class="header-title-wrapper">
             <h2>已安装应用</h2>
             <p class="subtitle">查看和管理 Helm Releases</p>
           </div>
         </div>
         <div class="header-right">
            <el-select v-model="selectedNamespace" placeholder="选择命名空间" @change="fetchReleases" class="w-48">
              <el-option label="所有命名空间" value="" />
              <el-option label="default" value="default" />
              <el-option label="kube-system" value="kube-system" />
            </el-select>
         </div>
       </div>
    </el-card>

    <el-card class="content-card">
       <el-table
         :data="releaseList"
         style="width: 100%"
         v-loading="loading"
         height="calc(100vh - 280px)"
         stripe
       >
         <el-table-column prop="release_name" label="Release 名称" min-width="200">
             <template #default="{ row }">
                 <span class="font-medium text-main">{{ row.release_name }}</span>
             </template>
         </el-table-column>
         <el-table-column prop="namespace" label="命名空间" width="150" />
         <el-table-column prop="chart_name" label="Chart" min-width="180" />
         <el-table-column prop="chart_version" label="版本" width="120" />
         <el-table-column prop="status" label="状态" width="120">
           <template #default="scope">
             <el-tag :type="getStatusType(scope.row.status)" size="small" effect="plain">
               {{ scope.row.status }}
             </el-tag>
           </template>
         </el-table-column>
         <el-table-column prop="updated_at" label="更新时间" width="180">
           <template #default="scope">
             {{ formatTime(scope.row.updated_at) }}
           </template>
         </el-table-column>
         <el-table-column label="操作" width="200" fixed="right">
           <template #default="scope">
             <el-button link type="primary" size="small" @click="handleViewDetail(scope.row)">详情</el-button>
             <el-button link type="warning" size="small" @click="handleUpgrade(scope.row)">升级</el-button>
             <el-button link type="danger" size="small" @click="handleUninstall(scope.row)">卸载</el-button>
           </template>
         </el-table-column>
       </el-table>
    </el-card>

    <!-- Upgrade Dialog -->
    <el-dialog
      v-model="showUpgradeDialog"
      title="升级 Release"
      width="800px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form :model="upgradeForm" label-position="top" @submit.prevent>
        <el-row :gutter="24">
            <el-col :span="12">
                <el-form-item label="Release 名称">
                  <el-input v-model="upgradeForm.release_name" disabled :prefix-icon="Box" />
                </el-form-item>
            </el-col>
            <el-col :span="12">
                <el-form-item label="当前版本">
                    <el-tag type="info" size="large" class="w-full text-center" effect="plain">{{ upgradeForm.current_version }}</el-tag>
                </el-form-item>
            </el-col>
        </el-row>

        <el-form-item label="目标版本" required>
          <el-select 
            v-model="upgradeForm.chart_version" 
            placeholder="选择目标版本" 
            class="w-full"
          >
            <el-option
              v-for="version in chartVersions"
              :key="version.version"
              :label="version.version"
              :value="version.version"
            >
                <div class="flex justify-between items-center w-full">
                    <span class="font-medium">{{ version.version }}</span>
                    <span class="text-xs text-sub" v-if="version.app_version">App: {{ version.app_version }}</span>
                </div>
            </el-option>
          </el-select>
        </el-form-item>

        <el-form-item>
            <template #label>
                <div class="flex justify-between items-center w-full">
                    <span>自定义 Values (YAML)</span>
                    <el-button 
                        type="primary" 
                        link 
                        size="small" 
                        @click="loadDefaultValues"
                    >
                        <el-icon><Document /></el-icon> 加载官方示例
                    </el-button>
                </div>
            </template>
          <el-input
            v-model="upgradeForm.values"
            type="textarea"
            :rows="12"
            placeholder="#在此输入 YAML 格式的配置覆盖..."
            class="font-mono text-sm"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showUpgradeDialog = false">取消</el-button>
          <el-button type="primary" @click="confirmUpgrade" :loading="upgrading">
            确认升级
            <el-icon class="el-icon--right"><Upload /></el-icon>
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- Detail Dialog -->
    <el-dialog
      v-model="showDetailDialog"
      title="Release 详情"
      width="800px"
    >
      <div v-if="currentRelease">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="Release 名称">{{ currentRelease.release_name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ currentRelease.namespace }}</el-descriptions-item>
          <el-descriptions-item label="Chart">{{ currentRelease.chart_name }}</el-descriptions-item>
          <el-descriptions-item label="版本">{{ currentRelease.chart_version }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentRelease.status)">{{ currentRelease.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatTime(currentRelease.updated_at) }}</el-descriptions-item>
        </el-descriptions>

        <div v-if="currentRelease.values" class="mt-6">
          <h4 class="mb-3 text-main font-medium">自定义 Values</h4>
          <div class="bg-gray-50 border border-gray-200 rounded p-4 font-mono text-sm overflow-auto max-h-96">
            <pre class="m-0 text-gray-700">{{ formatValues(currentRelease.values) }}</pre>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, Upload, Box } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { getHelmReleases, getHelmReleaseDetail, uninstallHelmRelease, upgradeHelmRelease, getChartVersions } from '@/api/helm'
import { getSelectedInstanceId } from '@/stores/instanceStore'

const loading = ref(false)
const releaseList = ref([])
const currentRelease = ref(null)
const selectedNamespace = ref('')
const showDetailDialog = ref(false)

// 升级相关
const showUpgradeDialog = ref(false)
const upgrading = ref(false)
const chartVersions = ref([])
const upgradeForm = ref({
  instance_id: 0,
  namespace: '',
  release_name: '',
  chart_name: '',
  current_version: '',
  chart_version: '',
  repo_name: '', 
  values: ''
})

const fetchReleases = async () => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) {
    ElMessage.warning('请先选择K8s实例')
    return
  }

  loading.value = true
  try {
    const res = await getHelmReleases(instanceId, selectedNamespace.value)
    releaseList.value = res.data.releaseList || []
  } catch (error) {
    console.error(error)
    ElMessage.error('获取Release列表失败')
  } finally {
    loading.value = false
  }
}

const handleViewDetail = async (row) => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) return

  try {
    const res = await getHelmReleaseDetail(row.namespace, row.release_name, instanceId)
    currentRelease.value = res.data.release
    showDetailDialog.value = true
  } catch (error) {
    console.error(error)
    ElMessage.error('获取Release详情失败')
  }
}

const handleUpgrade = async (row) => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) return

  try {
    const res = await getHelmReleaseDetail(row.namespace, row.release_name, instanceId)
    const detail = res.data.release
    
    try {
        const verRes = await getChartVersions(detail.chart_name) 
        chartVersions.value = verRes.data.versions || []
    } catch (e) {
        chartVersions.value = []
    }
    
    // 填充表单
    upgradeForm.value = {
      instance_id: instanceId,
      namespace: detail.namespace,
      release_name: detail.release_name,
      chart_name: detail.chart_name,
      current_version: detail.chart_version,
      chart_version: detail.chart_version, 
      values: detail.values || '',
      repo_name: '', 
      repo_url: '' 
    }
    
    showUpgradeDialog.value = true
  } catch (error) {
    console.error(error)
    ElMessage.error('准备升级失败')
  }
}

const loadDefaultValues = async () => {
    try {
        ElMessage.warning('升级模式下暂不支持加载默认 Values')
    } catch (e) {}
}

const confirmUpgrade = async () => {
    upgrading.value = true
    try { 
        const selectedVer = chartVersions.value.find(v => v.version === upgradeForm.value.chart_version)
        if (selectedVer && selectedVer.repo_url) {
            upgradeForm.value.repo_url = selectedVer.repo_url
        }

        await upgradeHelmRelease(upgradeForm.value, upgradeForm.value.instance_id)
        ElMessage.success('升级成功')
        showUpgradeDialog.value = false
        fetchReleases()
    } catch (error) {
        console.error(error)
        ElMessage.error('升级失败')
    } finally {
        upgrading.value = false
    }
}

const handleUninstall = (row) => {
  ElMessageBox.confirm(
    `确定要卸载 Release "${row.release_name}" 吗？此操作将删除所有相关的 Kubernetes 资源。`,
    '卸载确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    const instanceId = getSelectedInstanceId()
    if (!instanceId) return

    try {
      await uninstallHelmRelease(row.namespace, row.release_name, instanceId)
      ElMessage.success('卸载成功')
      fetchReleases()
    } catch (error) {
      console.error(error)
      ElMessage.error('卸载失败')
    }
  }).catch(() => {})
}

const getStatusType = (status) => {
  const statusMap = {
    deployed: 'success',
    failed: 'danger',
    pending: 'warning',
    uninstalled: 'info'
  }
  return statusMap[status?.toLowerCase()] || 'info'
}

const formatTime = (timestamp) => {
  return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

const formatValues = (values) => {
  try {
    if (typeof values === 'string') {
      return JSON.stringify(JSON.parse(values), null, 2)
    }
    return JSON.stringify(values, null, 2)
  } catch {
    return values
  }
}

onMounted(() => {
  fetchReleases()
})
</script>

<style scoped>
.header-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: var(--primary-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.text-main { color: var(--text-main); }
.text-sub { color: var(--text-sub); }
.text-gray-700 { color: #374151; }
.font-medium { font-weight: 500; }
.font-mono { font-family: 'Consolas', 'Monaco', monospace; }
.w-48 { width: 12rem; }
.w-full { width: 100%; }
.m-0 { margin: 0; }
.mt-6 { margin-top: 24px; }
.mb-3 { margin-bottom: 12px; }
.p-4 { padding: 16px; }
.max-h-96 { max-height: 24rem; }
.flex { display: flex; }
.items-center { align-items: center; }
.justify-between { justify-content: space-between; }
.overflow-auto { overflow: auto; }
.text-center { text-align: center; }
.text-xs { font-size: 12px; }

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-left { display: flex; align-items: center; gap: 12px; }
.header-title-wrapper h2 { margin: 0; font-size: 18px; color: var(--text-main); }
.header-title-wrapper .subtitle { margin: 0; font-size: 12px; color: var(--text-sub); }
</style>
