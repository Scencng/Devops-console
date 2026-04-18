<template>
  <div class="secret-management">
    <el-card class="page-header-card cyber-card">
      <div class="page-header cyber-header">
        <div>
          <h2>Secret 管理</h2>
          <p>查看和管理 Kubernetes Secret资源</p>
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
          <el-button type="primary" class="autoops-btn-primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:secret:showcreatedialogtrue')" >
            <el-icon><Plus /></el-icon>
            创建 Secret
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card cyber-card">
      <div class="autoops-table-wrapper">
        <el-table 
          :data="secretList" 
          style="width: 100%" class="autoops-table" 
          v-loading="loading"
          element-loading-background="rgba(0, 0, 0, 0.5)"
          height="calc(100vh - 280px)"
          :empty-text="loading ? '加载中...' : '暂无数据'"
        >
          <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip />
          <el-table-column prop="namespace" label="命名空间" width="150" />
          <el-table-column prop="type" label="类型" width="180" show-overflow-tooltip />
          <el-table-column prop="dataCount" label="数据项" width="100" align="center" />
          <el-table-column label="创建时间" width="170">
            <template #default="scope">
              {{ formatTimestamp(scope.row.age) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="scope">
              <el-button link type="primary" size="small" @click="handleViewDetail(scope.row)" v-show="permStore.hasPerm('k8s:secret:handleviewdetail')" >详情</el-button>
              <el-button link type="info" size="small" @click="handleEdit(scope.row)" v-show="permStore.hasPerm('k8s:secret:handleedit')" >编辑</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:secret:handledelete')" >删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog v-model="showCreateDialog" :title="isEditMode ? '编辑 Secret' : '创建 Secret'" width="900px" :close-on-click-modal="false">
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
            <el-form-item label="类型" prop="type">
              <el-select v-model="createForm.type" style="width: 100%;" :disabled="isEditMode">
                <el-option label="Opaque" value="Opaque" />
                <el-option label="kubernetes.io/tls" value="kubernetes.io/tls" />
                <el-option label="kubernetes.io/dockerconfigjson" value="kubernetes.io/dockerconfigjson" />
                <el-option label="kubernetes.io/basic-auth" value="kubernetes.io/basic-auth" />
                <el-option label="kubernetes.io/ssh-auth" value="kubernetes.io/ssh-auth" />
              </el-select>
            </el-form-item>

            <el-divider content-position="left">数据项</el-divider>
            <div class="data-editor">
              <div v-for="(keyName, index) in dataKeys" :key="index" class="data-item">
                <el-row :gutter="10">
                  <el-col :span="7">
                    <el-input v-model="dataKeys[index]" placeholder="键" @blur="updateDataKey(index, dataKeys[index])" />
                  </el-col>
                  <el-col :span="14">
                    <el-input 
                      v-model="createForm.data[keyName]" 
                      :type="dataVisibility[keyName] ? 'text' : 'password'" 
                      placeholder="值"
                      show-password
                    />
                  </el-col>
                  <el-col :span="1">
                    <el-button link @click="toggleVisibility(keyName)">
                      <el-icon><View v-if="!dataVisibility[keyName]" /><Hide v-else /></el-icon>
                    </el-button>
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
    <el-dialog v-model="showDetailDialog" title="Secret 详情" width="800px">
      <div v-if="currentSecret">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ currentSecret.metadata.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ currentSecret.metadata.namespace }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ currentSecret.type }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTimestamp(currentSecret.metadata.creationTimestamp) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">数据项</el-divider>
        <el-table :data="formatDataToTable(currentSecret.data)" style="margin-top: 10px;">
          <el-table-column prop="key" label="键" width="250" />
          <el-table-column label="值">
            <template #default="scope">
              <span v-if="detailVisibility[scope.row.key]">{{ decodeBase64(scope.row.value) }}</span>
              <span v-else>••••••••</span>
              <el-button link size="small" @click="toggleDetailVisibility(scope.row.key)" style="margin-left: 10px;">
                <el-icon><View v-if="!detailVisibility[scope.row.key]" /><Hide v-else /></el-icon>
              </el-button>
            </template>
          </el-table-column>
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
import { Plus, Delete, View, Hide } from '@element-plus/icons-vue'
import { getSecretList, getSecretDetail, createSecret, updateSecret, deleteSecret } from '@/api/k8s/secret'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const loading = ref(false)
const submitting = ref(false)
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const isEditMode = ref(false)
const editingSecret = ref(null)
const secretList = ref([])
const namespaceList = ref([])
const selectedNamespace = ref('all')
const currentSecret = ref(null)
const activeTab = ref('form')
const dataKeys = ref([])
const dataVisibility = ref({})
const detailVisibility = ref({})

const createForm = ref({
  name: '',
  namespace: 'default',
  type: 'Opaque',
  data: {},
  yaml: ''
})

const formatTimestamp = (timestamp) => {
  if (typeof timestamp === 'number') {
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
  }
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

const decodeBase64 = (str) => {
  try {
    return atob(str)
  } catch (e) {
    return str
  }
}

const formatDataToTable = (data) => {
  if (!data) return []
  return Object.keys(data).map(key => ({
    key,
    value: data[key]
  }))
}

const toggleVisibility = (key) => {
  dataVisibility.value[key] = !dataVisibility.value[key]
}

const toggleDetailVisibility = (key) => {
  detailVisibility.value[key] = !detailVisibility.value[key]
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
    const res = await getSecretList(selectedNamespace.value, instanceId)
    if (res.status === 200) {
      secretList.value = res.data.secretList
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
  dataVisibility.value[newKey] = false
}

const removeDataItem = (key) => {
  delete createForm.value.data[key]
  delete dataVisibility.value[key]
  const index = dataKeys.value.indexOf(key)
  if (index > -1) {
    dataKeys.value.splice(index, 1)
  }
}

const updateDataKey = (index, newKey) => {
  const oldKey = dataKeys.value[index]
  if (oldKey !== newKey && createForm.value.data.hasOwnProperty(oldKey)) {
    createForm.value.data[newKey] = createForm.value.data[oldKey]
    dataVisibility.value[newKey] = dataVisibility.value[oldKey]
    delete createForm.value.data[oldKey]
    delete dataVisibility.value[oldKey]
    dataKeys.value[index] = newKey
  }
}

const resetForm = () => {
  createForm.value = {
    name: '',
    namespace: 'default',
    type: 'Opaque',
    data: {},
    yaml: ''
  }
  dataKeys.value = []
  dataVisibility.value = {}
  isEditMode.value = false
  editingSecret.value = null
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
    await createSecret(createForm.value, instanceId)
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
    await updateSecret(
      editingSecret.value.namespace,
      editingSecret.value.name,
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
  const res = await getSecretDetail(row.namespace, row.name, instanceId)
  if (res.status === 200) {
    const secret = res.data.secretDetail
    editingSecret.value = row
    isEditMode.value = true
    
    // Decode base64 data for editing
    const decodedData = {}
    if (secret.data) {
      Object.keys(secret.data).forEach(key => {
        decodedData[key] = decodeBase64(secret.data[key])
      })
    }
    
    createForm.value = {
      name: secret.metadata.name,
      namespace: secret.metadata.namespace,
      type: secret.type,
      data: decodedData,
      yaml: ''
    }
    
    dataKeys.value = Object.keys(decodedData)
    dataKeys.value.forEach(key => {
      dataVisibility.value[key] = false
    })
    
    showCreateDialog.value = true
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除 Secret ${row.name} 吗?`, '提示', {
    type: 'warning'
  }).then(async () => {
    const instanceId = getSelectedInstanceId()
    await deleteSecret(row.namespace, row.name, instanceId)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleViewDetail = async (row) => {
  const instanceId = getSelectedInstanceId()
  const res = await getSecretDetail(row.namespace, row.name, instanceId)
  if (res.status === 200) {
    currentSecret.value = res.data.secretDetail
    detailVisibility.value = {}
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
.secret-management {
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
