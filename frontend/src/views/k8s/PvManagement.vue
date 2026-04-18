<template>
  <div class="pv-management">
    <el-card class="page-header-card cyber-card">
      <div class="page-header cyber-header">
        <div>
          <h2>PV 管理</h2>
          <p>查看和管理 Kubernetes Persistent Volumes</p>
        </div>
        <div class="header-actions">
          <el-button type="primary" class="autoops-btn-primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:pv:showcreatedialogtrue')" >
            <el-icon><Plus /></el-icon>
            创建 PV
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card cyber-card">
      <div class="autoops-table-wrapper">
        <el-table 
          :data="pvList" 
          style="width: 100%" class="autoops-table" 
          v-loading="loading"
          element-loading-background="rgba(0, 0, 0, 0.5)"
          height="calc(100vh - 280px)"
          :empty-text="loading ? '加载中...' : '暂无数据'"
        >
          <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip />
          <el-table-column prop="capacity" label="容量" width="90" />
          <el-table-column prop="accessModes" label="访问模式" width="160">
            <template #default="scope">
              <el-tag v-for="mode in scope.row.accessModes" :key="mode" size="small" effect="plain" style="margin-right: 4px;">
                {{ mode }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="reclaimPolicy" label="回收策略" width="120" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="getStatusType(scope.row.status)" effect="dark" class="autoops-tag">
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="storageClass" label="STORAGE CLASS" width="150" show-overflow-tooltip />
          <el-table-column prop="claim" label="CLAIM" min-width="180" show-overflow-tooltip />
          <el-table-column label="创建时间" width="170">
            <template #default="scope">
              {{ formatTimestamp(scope.row.age) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="scope">
              <el-button link type="primary" size="small" @click="handleViewDetail(scope.row)" v-show="permStore.hasPerm('k8s:pv:handleviewdetail')" >详情</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:pv:handledelete')" >删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" title="创建 PV" width="800px" :close-on-click-modal="false">
      <el-form :model="createForm" ref="createFormRef" label-width="120px">
        <el-tabs v-model="activeTab" type="border-card">
          <el-tab-pane label="表单创建" name="form">
            <el-form-item label="名称" prop="name" required>
              <el-input v-model="createForm.name" />
            </el-form-item>
            <el-form-item label="容量" prop="capacity" required>
              <el-input v-model="createForm.capacity" placeholder="e.g. 5Gi" />
            </el-form-item>
            <el-form-item label="StorageClass" prop="storageClass">
              <el-input v-model="createForm.storageClass" />
            </el-form-item>
            <el-form-item label="访问模式" prop="accessModes" required>
              <el-checkbox-group v-model="createForm.accessModes">
                 <el-checkbox label="ReadWriteOnce">ReadWriteOnce</el-checkbox>
                 <el-checkbox label="ReadOnlyMany">ReadOnlyMany</el-checkbox>
                 <el-checkbox label="ReadWriteMany">ReadWriteMany</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
             <el-form-item label="HostPath" prop="hostPath">
              <el-input v-model="createForm.hostPath" placeholder="可选: /data/pv1" />
            </el-form-item>
            <el-form-item label="NFS Server">
               <el-input v-model="createForm.nfs.server" placeholder="NFS Server IP" />
            </el-form-item>
             <el-form-item label="NFS Path">
               <el-input v-model="createForm.nfs.path" placeholder="/nfs/share" />
            </el-form-item>
          </el-tab-pane>
          <el-tab-pane label="YAML创建" name="yaml">
            <el-input
              v-model="createForm.yaml"
              type="textarea"
              :rows="15"
              placeholder="请输入YAML配置"
            />
          </el-tab-pane>
        </el-tabs>
      </el-form>
       <template #footer>
        <span class="dialog-footer">
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="handleCreate" :loading="submitting">创建</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetailDialog" title="PV 详情" width="800px">
       <div v-if="currentPv">
         <el-descriptions :column="1" border>
            <el-descriptions-item label="名称">{{ currentPv.metadata.name }}</el-descriptions-item>
            <el-descriptions-item label="状态">{{ currentPv.status.phase }}</el-descriptions-item>
            <el-descriptions-item label="Capacity" v-if="currentPv.spec.capacity">{{ currentPv.spec.capacity.storage }}</el-descriptions-item>
            <el-descriptions-item label="StorageClass">{{ currentPv.spec.storageClass }}</el-descriptions-item>
            <el-descriptions-item label="Reclaim Policy">{{ currentPv.spec.persistentVolumeReclaimPolicy }}</el-descriptions-item>
         </el-descriptions>
       </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, View, Delete } from '@element-plus/icons-vue'
import { getPvList, getPvDetail, createPv, deletePv } from '@/api/k8s/storage'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const loading = ref(false)
const submitting = ref(false)
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const pvList = ref([])
const currentPv = ref(null)
const activeTab = ref('form')

const createForm = ref({
  name: '',
  capacity: '1Gi',
  storageClass: '',
  accessModes: ['ReadWriteOnce'],
  hostPath: '',
  nfs: {
    server: '',
    path: ''
  },
  yaml: ''
})

const getStatusType = (status) => {
  switch (status) {
    case 'Bound': return 'success'
    case 'Available': return 'info'
    case 'Released': return 'warning'
    case 'Failed': return 'danger'
    default: return ''
  }
}

const formatTimestamp = (timestamp) => {
  return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

const fetchData = async () => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) return
  
  loading.value = true
  try {
    const res = await getPvList(instanceId)
    if (res.status === 200) {
      pvList.value = res.data.pvList
    }
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const handleCreate = async () => {
  const instanceId = getSelectedInstanceId()
  submitting.value = true
  try {
    await createPv(createForm.value, instanceId)
    ElMessage.success('创建成功')
    showCreateDialog.value = false
    fetchData()
  } catch (error) {
    console.error(error)
  } finally {
    submitting.value = false
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除 PV ${row.name} 吗?`, '提示', {
    type: 'warning'
  }).then(async () => {
    const instanceId = getSelectedInstanceId()
    await deletePv(row.name, instanceId)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleViewDetail = async (row) => {
  const instanceId = getSelectedInstanceId()
  const res = await getPvDetail(row.name, instanceId)
  if (res.status === 200) {
    currentPv.value = res.data.pvDetail
    showDetailDialog.value = true
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.page-header-card {
  margin-bottom: 0px !important;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.pv-management {
  padding: 20px;
}
</style>
