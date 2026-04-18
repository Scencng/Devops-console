<template>
  <div class="sc-management">
    <el-card class="page-header-card cyber-card">
      <div class="page-header cyber-header">
        <div>
          <h2>StorageClass 管理</h2>
          <p>查看和管理 Kubernetes Storage Classes</p>
        </div>
        <div class="header-actions">
           <el-button type="primary" class="autoops-btn-primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:storageclass:showcreatedialogtrue')" >
            <el-icon><Plus /></el-icon>
            创建 StorageClass
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card cyber-card">
      <div class="autoops-table-wrapper">
        <el-table 
          :data="scList" 
          style="width: 100%" class="autoops-table" 
          v-loading="loading"
          element-loading-background="rgba(0, 0, 0, 0.5)"
          height="calc(100vh - 280px)"
          :empty-text="loading ? '加载中...' : '暂无数据'"
        >
          <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip />
          <el-table-column prop="provisioner" label="PROVISIONER" min-width="250" show-overflow-tooltip />
          <el-table-column prop="reclaimPolicy" label="回收策略" width="120" />
          <el-table-column prop="volumeBindingMode" label="绑定模式" width="180" />
          <el-table-column label="创建时间" width="170">
            <template #default="scope">
              {{ formatTimestamp(scope.row.age) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="scope">
              <el-button link type="primary" size="small" @click="handleViewDetail(scope.row)" v-show="permStore.hasPerm('k8s:storageclass:handleviewdetail')" >详情</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:storageclass:handledelete')" >删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" title="创建 StorageClass" width="800px" :close-on-click-modal="false">
      <el-form :model="createForm" ref="createFormRef" label-width="120px">
        <el-tabs v-model="activeTab" type="border-card">
          <el-tab-pane label="表单创建" name="form">
             <el-form-item label="名称" prop="name" required>
              <el-input v-model="createForm.name" />
            </el-form-item>
            <el-form-item label="Provisioner" prop="provisioner" required>
              <el-input v-model="createForm.provisioner" placeholder="e.g. kubernetes.io/aws-ebs" />
            </el-form-item>
             <el-form-item label="Reclaim Policy" prop="reclaimPolicy">
               <el-select v-model="createForm.reclaimPolicy">
                 <el-option label="Delete" value="Delete" />
                 <el-option label="Retain" value="Retain" />
               </el-select>
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
    <el-dialog v-model="showDetailDialog" title="StorageClass 详情" width="800px">
       <div v-if="currentSc">
         <el-descriptions :column="1" border>
            <el-descriptions-item label="名称">{{ currentSc.metadata.name }}</el-descriptions-item>
            <el-descriptions-item label="Provisioner">{{ currentSc.provisioner }}</el-descriptions-item>
            <el-descriptions-item label="Reclaim Policy">{{ currentSc.reclaimPolicy }}</el-descriptions-item>
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
import { getScList, getScDetail, createSc, deleteSc } from '@/api/k8s/storage'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const loading = ref(false)
const submitting = ref(false)
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const scList = ref([])
const currentSc = ref(null)
const activeTab = ref('form')

const createForm = ref({
  name: '',
  provisioner: '',
  reclaimPolicy: 'Delete',
  yaml: ''
})

const formatTimestamp = (timestamp) => {
  return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

const fetchData = async () => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) return
  console.log(11111111)
  loading.value = true
  try {
    const res = await getScList(instanceId)
    console.log(res)
    if (res.status === 200) {
      scList.value = res.data.scList
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
    await createSc(createForm.value, instanceId)
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
  ElMessageBox.confirm(`确定删除 StorageClass ${row.name} 吗?`, '提示', {
    type: 'warning'
  }).then(async () => {
    const instanceId = getSelectedInstanceId()
    await deleteSc(row.name, instanceId)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleViewDetail = async (row) => {
  const instanceId = getSelectedInstanceId()
  const res = await getScDetail(row.name, instanceId)
  if (res.status === 200) {
    currentSc.value = res.data.scDetail
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
.sc-management {
  padding: 20px;
}
</style>
