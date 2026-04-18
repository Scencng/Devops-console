<template>
  <div class="autoops-container">
    <el-card class="page-header-card cyber-card">
      <div class="page-header cyber-header">
        <div>
          <h2>ConfigMap 管理</h2>
          <p>查看和管理 Kubernetes ConfigMap资源</p>
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
          <el-button type="primary" class="autoops-btn-primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:configmap:showcreatedialogtrue')" >
            <el-icon><Plus /></el-icon>
            创建 ConfigMap
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card cyber-card">
      <div class="autoops-table-wrapper">
        <el-table 
          :data="configMapList" 
          style="width: 100%" class="autoops-table" 
          v-loading="loading"
          element-loading-background="rgba(0, 0, 0, 0.5)"
          height="calc(100vh - 280px)"
          :empty-text="loading ? '加载中...' : '暂无数据'"
        >
          <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip />
          <el-table-column prop="namespace" label="命名空间" width="150" />
          <el-table-column prop="dataCount" label="数据项" width="100" align="center" />
          <el-table-column label="创建时间" width="170">
            <template #default="scope">
              {{ formatTimestamp(scope.row.age) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="scope">
              <el-button link type="primary" size="small" @click="handleViewDetail(scope.row)" v-show="permStore.hasPerm('k8s:configmap:handleviewdetail')" >详情</el-button>
              <el-button link type="info" size="small" @click="handleEdit(scope.row)" v-show="permStore.hasPerm('k8s:configmap:handleedit')" >编辑</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:configmap:handledelete')" >删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog v-model="showCreateDialog" :title="isEditMode ? '编辑 ConfigMap' : '创建 ConfigMap'" width="900px" :close-on-click-modal="false">
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
              <el-input v-model="createForm.name" :disabled="isEditMode" />
            </el-form-item>

            <el-divider content-position="left">数据项</el-divider>
            <div class="data-editor">
              <div v-for="(keyName, index) in dataKeys" :key="index" class="data-item">
                <el-row :gutter="10">
                  <el-col :span="7">
                    <el-input v-model="dataKeys[index]" placeholder="键" @blur="updateDataKey(index, dataKeys[index])" />
                  </el-col>
                  <el-col :span="15">
                    <el-input v-model="createForm.data[keyName]" type="textarea" :rows="2" placeholder="值" />
                  </el-col>
                  <el-col :span="2">
                    <el-button link type="danger" @click="removeDataItem(keyName)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </el-col>
                </el-row>
              </div>
              <el-button type="primary" plain @click="addDataItem" style="margin-top: 10px;">
                <el-icon><Plus /></el-icon>
                添加数据项
              </el-button>
            </div>
          </el-tab-pane>
          
          <el-tab-pane label="YAML创建" name="yaml">
            <el-input
              v-model="createForm.yaml"
              type="textarea"
              :rows="20"
              placeholder="请输入YAML配置"
            />
          </el-tab-pane>
        </el-tabs>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="handleCloseDialog">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ isEditMode ? '更新' : '创建' }}</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetailDialog" title="ConfigMap 详情" width="800px">
      <div v-if="currentConfigMap">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ currentConfigMap.metadata.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ currentConfigMap.metadata.namespace }}</el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">{{ formatTimestamp(currentConfigMap.metadata.creationTimestamp) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">数据项</el-divider>
        <el-table :data="formatDataToTable(currentConfigMap.data)" style="margin-top: 10px;">
          <el-table-column prop="key" label="键" width="250" />
          <el-table-column prop="value" label="值" show-overflow-tooltip />
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { getConfigMapList, getConfigMapDetail, createConfigMap, updateConfigMap, deleteConfigMap } from '@/api/k8s/configmap'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const loading = ref(false)
const submitting = ref(false)
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const isEditMode = ref(false)
const editingConfigMap = ref(null)
const configMapList = ref([])
const namespaceList = ref([])
const selectedNamespace = ref('all')
const currentConfigMap = ref(null)
const activeTab = ref('form')
const dataKeys = ref([])

const createForm = ref({
  name: '',
  namespace: 'default',
  data: {},
  yaml: ''
})

const formatTimestamp = (timestamp) => {
  if (typeof timestamp === 'number') {
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
  }
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

const formatDataToTable = (data) => {
  if (!data) return []
  return Object.keys(data).map(key => ({
    key,
    value: data[key]
  }))
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
    const res = await getConfigMapList(selectedNamespace.value, instanceId)
    if (res.status === 200) {
      configMapList.value = res.data.configMapList
    }
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const addDataItem = () => {
  const newKey = `key${Object.keys(createForm.value.data).length + 1}`
  createForm.value.data[newKey] = ''
  dataKeys.value.push(newKey)
}

const removeDataItem = (key) => {
  delete createForm.value.data[key]
  const index = dataKeys.value.indexOf(key)
  if (index > -1) {
    dataKeys.value.splice(index, 1)
  }
}

const updateDataKey = (index, newKey) => {
  const oldKey = dataKeys.value[index]
  if (oldKey !== newKey && createForm.value.data.hasOwnProperty(oldKey)) {
    createForm.value.data[newKey] = createForm.value.data[oldKey]
    delete createForm.value.data[oldKey]
    dataKeys.value[index] = newKey
  }
}

const resetForm = () => {
  createForm.value = {
    name: '',
    namespace: 'default',
    data: {},
    yaml: ''
  }
  dataKeys.value = []
  isEditMode.value = false
  editingConfigMap.value = null
  activeTab.value = 'form'
}

const handleCloseDialog = () => {
  showCreateDialog.value = false
  resetForm()
}

const handleCreate = async () => {
  const instanceId = getSelectedInstanceId()
  submitting.value = true
  try {
    await createConfigMap(createForm.value, instanceId)
    ElMessage.success('创建成功')
    handleCloseDialog()
    fetchData()
  } catch (error) {
    console.error(error)
    ElMessage.error('创建失败')
  } finally {
    submitting.value = false
  }
}

const handleUpdate = async () => {
  const instanceId = getSelectedInstanceId()
  submitting.value = true
  try {
    await updateConfigMap(
      editingConfigMap.value.namespace,
      editingConfigMap.value.name,
      createForm.value,
      instanceId
    )
    ElMessage.success('更新成功')
    handleCloseDialog()
    fetchData()
  } catch (error) {
    console.error(error)
    ElMessage.error('更新失败')
  } finally {
    submitting.value = false
  }
}

const handleSubmit = () => {
  if (isEditMode.value) {
    handleUpdate()
  } else {
    handleCreate()
  }
}

const handleEdit = async (row) => {
  const instanceId = getSelectedInstanceId()
  const res = await getConfigMapDetail(row.namespace, row.name, instanceId)
  if (res.status === 200) {
    const configMap = res.data.configMapDetail
    editingConfigMap.value = row
    isEditMode.value = true
    
    createForm.value = {
      name: configMap.metadata.name,
      namespace: configMap.metadata.namespace,
      data: configMap.data || {},
      yaml: ''
    }
    
    // Initialize dataKeys from existing data
    dataKeys.value = Object.keys(configMap.data || {})
    
    showCreateDialog.value = true
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除 ConfigMap ${row.name} 吗?`, '提示', {
    type: 'warning'
  }).then(async () => {
    const instanceId = getSelectedInstanceId()
    await deleteConfigMap(row.namespace, row.name, instanceId)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleViewDetail = async (row) => {
  const instanceId = getSelectedInstanceId()
  const res = await getConfigMapDetail(row.namespace, row.name, instanceId)
  if (res.status === 200) {
    currentConfigMap.value = res.data.configMapDetail
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
.configmap-management {
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

.data-editor {
  max-height: 400px;
  overflow-y: auto;
}

.data-item {
  margin-bottom: 10px;
  padding: 10px;
  background: #f9fafb;
  border-radius: 8px;
}

html.dark .data-item {
  background: rgba(255, 255, 255, 0.02);
}
</style>
