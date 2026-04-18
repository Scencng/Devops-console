<template>
  <div class="ingress-management">
    <el-card class="page-header-card cyber-card">
      <div class="page-header cyber-header">
        <div>
          <h2>Ingress 管理</h2>
          <p>查看和管理 Kubernetes Ingress资源</p>
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
          <el-button type="primary" class="autoops-btn-primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:ingress:showcreatedialogtrue')" >
            <el-icon><Plus /></el-icon>
            创建 Ingress
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card cyber-card">
      <div class="autoops-table-wrapper">
        <el-table 
          :data="ingressList" 
          style="width: 100%" class="autoops-table" 
          v-loading="loading"
          element-loading-background="rgba(0, 0, 0, 0.5)"
          height="calc(100vh - 280px)"
          :empty-text="loading ? '加载中...' : '暂无数据'"
        >
          <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
          <el-table-column prop="namespace" label="命名空间" width="120" />
          <el-table-column prop="className" label="INGRESS CLASS" width="150" show-overflow-tooltip />
          <el-table-column prop="hosts" label="HOSTS" min-width="200" show-overflow-tooltip>
            <template #default="scope">
              <el-tag v-for="host in scope.row.hosts" :key="host" size="small" effect="plain" style="margin-right: 4px;">
                {{ host }}
              </el-tag>
              <span v-if="scope.row.hosts.length === 0" style="color: #999;">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="address" label="ADDRESS" min-width="160" show-overflow-tooltip>
            <template #default="scope">
              <span v-if="scope.row.address">{{ scope.row.address }}</span>
              <span v-else style="color: #999;">-</span>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="170">
            <template #default="scope">
              {{ formatTimestamp(scope.row.age) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="scope">
              <el-button link type="primary" size="small" @click="handleViewDetail(scope.row)" v-show="permStore.hasPerm('k8s:ingress:handleviewdetail')" >详情</el-button>
              <el-button link type="info" size="small" @click="handleEdit(scope.row)" v-show="permStore.hasPerm('k8s:ingress:handleedit')" >编辑</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:ingress:handledelete')" >删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog v-model="showCreateDialog" :title="isEditMode ? '编辑 Ingress' : '创建 Ingress'" width="1100px" :close-on-click-modal="false">
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
            <el-form-item label="Ingress Class" prop="className">
              <el-input v-model="createForm.className" placeholder="nginx, traefik, etc." />
            </el-form-item>

            <el-divider content-position="left">Rules</el-divider>
            <div v-for="(rule, ruleIndex) in createForm.rules" :key="ruleIndex" class="rule-container">
              <div class="rule-header">
                <span class="rule-title">Rule {{ ruleIndex + 1 }}</span>
                <el-button link type="danger" size="small" @click="removeRule(ruleIndex)">删除规则</el-button>
              </div>
              <el-form-item label="Host" label-width="80px">
                <el-input v-model="rule.host" placeholder="example.com" />
              </el-form-item>
              
              <div v-for="(path, pathIndex) in rule.paths" :key="pathIndex" class="path-item">
                <div class="path-header">
                  <span class="path-label">Path {{ pathIndex + 1 }}</span>
                  <el-button link type="danger" size="small" @click="removePath(ruleIndex, pathIndex)">删除路径</el-button>
                </div>
                <el-row :gutter="8">
                  <el-col :span="5">
                    <el-form-item label="Path">
                      <el-input v-model="path.path" placeholder="/" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="5">
                    <el-form-item label="PathType">
                      <el-select v-model="path.pathType" style="width: 100%;">
                        <el-option label="Prefix" value="Prefix" />
                        <el-option label="Exact" value="Exact" />
                        <el-option label="ImplementationSpecific" value="ImplementationSpecific" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :span="8">
                    <el-form-item label="Service Name">
                      <el-input v-model="path.serviceName" placeholder="my-service" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="6">
                    <el-form-item label="Port">
                      <el-input-number v-model="path.servicePort" :min="1" :max="65535" style="width: 100%;" controls-position="right" />
                    </el-form-item>
                  </el-col>
                </el-row>
              </div>
              <el-button size="small" type="primary" plain @click="addPath(ruleIndex)">+ 添加 Path</el-button>
            </div>
            <el-button type="primary" plain @click="addRule" style="margin-top: 10px;">+ 添加 Rule</el-button>

            <el-divider content-position="left">TLS (可选)</el-divider>
            <div v-for="(tls, tlsIndex) in createForm.tls" :key="tlsIndex" style="margin-bottom: 15px; padding: 10px; border: 1px solid var(--border-color); border-radius: 8px;">
              <el-form-item label="Hosts">
                <el-select v-model="tls.hosts" multiple placeholder="选择或输入hosts" allow-create filterable style="width: calc(100% - 50px);">
                </el-select>
                <el-button link type="danger" size="small" @click="removeTLS(tlsIndex)" style="margin-left: 10px;">删除</el-button>
              </el-form-item>
              <el-form-item label="Secret Name">
                <el-input v-model="tls.secretName" placeholder="tls-secret" />
              </el-form-item>
            </div>
            <el-button @click="addTLS">+ 添加 TLS</el-button>
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
    <el-dialog v-model="showDetailDialog" title="Ingress 详情" width="900px">
      <div v-if="currentIngress">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ currentIngress.metadata.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ currentIngress.metadata.namespace }}</el-descriptions-item>
          <el-descriptions-item label="Ingress Class">{{ currentIngress.spec.ingressClassName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTimestamp(currentIngress.metadata.creationTimestamp) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">Rules</el-divider>
        <div v-for="(rule, index) in currentIngress.spec.rules" :key="index" style="margin-bottom: 15px;">
          <strong>Host:</strong> {{ rule.host || '*' }}
          <el-table :data="rule.http?.paths || []" style="margin-top: 10px;">
            <el-table-column prop="path" label="Path" width="150" />
            <el-table-column prop="pathType" label="PathType" width="150" />
            <el-table-column label="Backend Service">
              <template #default="scope">
                {{ scope.row.backend.service?.name }}:{{ scope.row.backend.service?.port.number }}
              </template>
            </el-table-column>
          </el-table>
        </div>

        <el-divider content-position="left" v-if="currentIngress.spec.tls && currentIngress.spec.tls.length > 0">TLS</el-divider>
        <el-table v-if="currentIngress.spec.tls && currentIngress.spec.tls.length > 0" :data="currentIngress.spec.tls">
          <el-table-column label="Hosts">
            <template #default="scope">
              {{ scope.row.hosts.join(', ') }}
            </template>
          </el-table-column>
          <el-table-column prop="secretName" label="Secret Name" />
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
import { Plus } from '@element-plus/icons-vue'
import { getIngressList, getIngressDetail, createIngress, updateIngress, deleteIngress } from '@/api/k8s/network'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const loading = ref(false)
const submitting = ref(false)
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const isEditMode = ref(false)
const editingIngress = ref(null)
const ingressList = ref([])
const namespaceList = ref([])
const selectedNamespace = ref('all')
const currentIngress = ref(null)
const activeTab = ref('form')

const createForm = ref({
  name: '',
  namespace: 'default',
  className: '',
  rules: [
    {
      host: '',
      paths: [
        {
          path: '/',
          pathType: 'Prefix',
          serviceName: '',
          servicePort: 80
        }
      ]
    }
  ],
  tls: [],
  yaml: ''
})

const formatTimestamp = (timestamp) => {
  if (typeof timestamp === 'number') {
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
  }
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
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
    const res = await getIngressList(selectedNamespace.value, instanceId)
    if (res.status === 200) {
      ingressList.value = res.data.ingressList
    }
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const addRule = () => {
  createForm.value.rules.push({
    host: '',
    paths: [
      {
        path: '/',
        pathType: 'Prefix',
        serviceName: '',
        servicePort: 80
      }
    ]
  })
}

const removeRule = (index) => {
  createForm.value.rules.splice(index, 1)
}

const addPath = (ruleIndex) => {
  createForm.value.rules[ruleIndex].paths.push({
    path: '/',
    pathType: 'Prefix',
    serviceName: '',
    servicePort: 80
  })
}

const removePath = (ruleIndex, pathIndex) => {
  createForm.value.rules[ruleIndex].paths.splice(pathIndex, 1)
}

const addTLS = () => {
  createForm.value.tls.push({
    hosts: [],
    secretName: ''
  })
}

const removeTLS = (index) => {
  createForm.value.tls.splice(index, 1)
}

const resetForm = () => {
  createForm.value = {
    name: '',
    namespace: 'default',
    className: '',
    rules: [
      {
        host: '',
        paths: [
          {
            path: '/',
            pathType: 'Prefix',
            serviceName: '',
            servicePort: 80
          }
        ]
      }
    ],
    tls: [],
    yaml: ''
  }
  isEditMode.value = false
  editingIngress.value = null
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
    // 清理空值字段
    const payload = {
      ...createForm.value,
      className: createForm.value.className?.trim() || undefined
    }
    await createIngress(payload, instanceId)
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
    // 清理空值字段
    const payload = {
      ...createForm.value,
      className: createForm.value.className?.trim() || undefined
    }
    await updateIngress(
      editingIngress.value.namespace,
      editingIngress.value.name,
      payload,
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
  const res = await getIngressDetail(row.namespace, row.name, instanceId)
  if (res.status === 200) {
    const ingress = res.data.ingressDetail
    editingIngress.value = row
    isEditMode.value = true
    
    // 将 K8s Ingress 对象转换为表单数据
    createForm.value = {
      name: ingress.metadata.name,
      namespace: ingress.metadata.namespace,
      className: ingress.spec.ingressClassName || '',
      rules: ingress.spec.rules?.map(rule => ({
        host: rule.host || '',
        paths: rule.http?.paths?.map(p => ({
          path: p.path,
          pathType: p.pathType,
          serviceName: p.backend.service?.name || '',
          servicePort: p.backend.service?.port?.number || 80
        })) || []
      })) || [],
      tls: ingress.spec.tls?.map(t => ({
        hosts: t.hosts || [],
        secretName: t.secretName || ''
      })) || [],
      annotations: ingress.metadata.annotations || {},
      yaml: ''
    }
    
    showCreateDialog.value = true
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除 Ingress ${row.name} 吗?`, '提示', {
    type: 'warning'
  }).then(async () => {
    const instanceId = getSelectedInstanceId()
    await deleteIngress(row.namespace, row.name, instanceId)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleViewDetail = async (row) => {
  const instanceId = getSelectedInstanceId()
  const res = await getIngressDetail(row.namespace, row.name, instanceId)
  if (res.status === 200) {
    currentIngress.value = res.data.ingressDetail
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
.ingress-management {
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

/* Dialog styles */
.rule-container {
  margin-bottom: 20px;
  padding: 20px;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  background: rgba(var(--bg-card), 0.3);
  transition: all 0.3s;
}

.rule-container:hover {
  border-color: var(--primary-color);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-color);
}

.rule-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-primary);
}

.path-item {
  margin-left: 0;
  margin-bottom: 16px;
  padding: 16px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

html.dark .path-item {
  background: rgba(255, 255, 255, 0.02);
  border-color: rgba(255, 255, 255, 0.1);
}

.path-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.path-label {
  font-size: 13px;
  font-weight: 500;
  color: #6b7280;
}

</style>
