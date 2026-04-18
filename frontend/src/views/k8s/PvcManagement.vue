<template>
  <div class="autoops-container">
    <el-card class="page-header-card cyber-card">
      <div class="page-header cyber-header">
        <div>
          <h2>PVC 管理</h2>
          <p>查看和管理 Kubernetes Persistent Volume Claims</p>
        </div>
        <div class="header-actions">
           <div class="namespace-selector">
            <span class="selector-label">命名空间：</span>
            <el-select v-model="selectedNamespace" placeholder="选择命名空间" @change="fetchData" style="width: 200px;" class="autoops-select">
              <el-option label="所有" value="all" />
              <el-option
                v-for="ns in namespaceList"
                :key="ns.name"
                :label="ns.name"
                :value="ns.name"
              />
            </el-select>
          </div>
          <el-button type="primary" class="autoops-btn-primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:pvc:showcreatedialogtrue')" >
            <el-icon><Plus /></el-icon>
            创建 PVC
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card cyber-card">
      <div class="autoops-table-wrapper">
        <el-table 
          :data="pvcList" 
          style="width: 100%" class="autoops-table" 
          v-loading="loading"
          element-loading-background="rgba(0, 0, 0, 0.5)"
          height="calc(100vh - 280px)"
          :empty-text="loading ? '加载中...' : '暂无数据'"
        >
          <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
          <el-table-column prop="namespace" label="命名空间" width="120" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="getStatusType(scope.row.status)" effect="dark" class="autoops-tag">
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="volume" label="VOLUME" min-width="180" show-overflow-tooltip />
          <el-table-column prop="capacity" label="容量" width="90" />
          <el-table-column prop="accessModes" label="访问模式" width="160" show-overflow-tooltip>
             <template #default="scope">
               <el-tag v-for="mode in scope.row.accessModes" :key="mode" size="small" effect="plain" style="margin-right: 4px;">
                 {{ mode }}
               </el-tag>
             </template>
          </el-table-column>
          <el-table-column prop="storageClass" label="STORAGE CLASS" min-width="160" show-overflow-tooltip />
          <el-table-column label="创建时间" width="170">
            <template #default="scope">
              {{ formatTimestamp(scope.row.age) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="scope">
              <el-button link type="primary" size="small" @click="handleViewDetail(scope.row)" v-show="permStore.hasPerm('k8s:pvc:handleviewdetail')" >详情</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:pvc:handledelete')" >删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" title="创建 PVC" width="800px" :close-on-click-modal="false">
      <el-form :model="createForm" ref="createFormRef" label-width="120px">
        <el-tabs v-model="activeTab" type="border-card">
          <el-tab-pane label="表单创建" name="form">
             <el-form-item label="命名空间" prop="namespace" required>
               <el-select v-model="createForm.namespace" style="width: 100%;">
                 <el-option
                  v-for="ns in namespaceList"
                  :key="ns.name"
                  :label="ns.name"
                  :value="ns.name"
                />
               </el-select>
            </el-form-item>
            <el-form-item label="名称" prop="name" required>
              <el-input v-model="createForm.name" />
            </el-form-item>
            <el-form-item label="容量" prop="capacity" required>
              <el-input v-model="createForm.capacity" placeholder="e.g. 1Gi" />
            </el-form-item>
            <el-form-item label="StorageClass" prop="storageClassName">
              <el-input v-model="createForm.storageClassName" />
            </el-form-item>
            <el-form-item label="访问模式" prop="accessModes" required>
              <el-checkbox-group v-model="createForm.accessModes">
                 <el-checkbox label="ReadWriteOnce">ReadWriteOnce</el-checkbox>
                 <el-checkbox label="ReadOnlyMany">ReadOnlyMany</el-checkbox>
                 <el-checkbox label="ReadWriteMany">ReadWriteMany</el-checkbox>
              </el-checkbox-group>
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
    <el-dialog v-model="showDetailDialog" title="PVC 详情" width="800px">
       <div v-if="currentPvc">
         <el-descriptions :column="1" border>
            <el-descriptions-item label="名称">{{ currentPvc.metadata.name }}</el-descriptions-item>
            <el-descriptions-item label="命名空间">{{ currentPvc.metadata.namespace }}</el-descriptions-item>
            <el-descriptions-item label="状态">{{ currentPvc.status.phase }}</el-descriptions-item>
            <el-descriptions-item label="Capacity" v-if="currentPvc.status.capacity">{{ currentPvc.status.capacity.storage }}</el-descriptions-item>
            <el-descriptions-item label="StorageClass">{{ currentPvc.spec.storageClassName }}</el-descriptions-item>
             <el-descriptions-item label="Volume" v-if="currentPvc.spec.volumeName">{{ currentPvc.spec.volumeName }}</el-descriptions-item>
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
import { getPvcList, getPvcDetail, createPvc, deletePvc } from '@/api/k8s/storage'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const loading = ref(false)
const submitting = ref(false)
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const pvcList = ref([])
const namespaceList = ref([])
const selectedNamespace = ref('all')
const currentPvc = ref(null)
const activeTab = ref('form')

const createForm = ref({
  name: '',
  namespace: 'default',
  capacity: '1Gi',
  storageClassName: '',
  accessModes: ['ReadWriteOnce'],
  yaml: ''
})

const getStatusType = (status) => {
  switch (status) {
    case 'Bound': return 'success'
    case 'Pending': return 'warning'
    case 'Lost': return 'danger'
    default: return ''
  }
}

const formatTimestamp = (timestamp) => {
  return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

const fetchNamespaces = async () => {
    const instanceId = getSelectedInstanceId()
    const res = await getNamespaceList(instanceId)
    if (res.status === 200) {
        namespaceList.value = res.data.namespaceList
    }
}

const fetchData = async () => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) return
  
  loading.value = true
  try {
    const res = await getPvcList(selectedNamespace.value, instanceId)
    if (res.status === 200) {
      pvcList.value = res.data.pvcList
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
    await createPvc(createForm.value, instanceId)
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
  ElMessageBox.confirm(`确定删除 PVC ${row.name} 吗?`, '提示', {
    type: 'warning'
  }).then(async () => {
    const instanceId = getSelectedInstanceId()
    await deletePvc(row.namespace, row.name, instanceId)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleViewDetail = async (row) => {
  const instanceId = getSelectedInstanceId()
  const res = await getPvcDetail(row.namespace, row.name, instanceId)
  if (res.status === 200) {
    currentPvc.value = res.data.pvcDetail
    showDetailDialog.value = true
  }
}

onMounted(() => {
  fetchNamespaces()
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
.pvc-management {
  padding: 20px;
}
.header-actions {
    display: flex;
    gap: 10px;
}
.namespace-selector {
    display: flex;
    align-items: center;
}
</style>
