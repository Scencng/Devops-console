<template>
  <div class="namespace-management">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>命名空间管理</h2>
          <p>管理 Kubernetes 命名空间</p>
        </div>
        <el-button type="primary" @click="showCreateDialog = true" v-show="permStore.hasPerm('k8s:namespace:showcreatedialogtrue')" >
          <el-icon><Plus /></el-icon>
          创建命名空间
        </el-button>
      </div>
    </el-card>

    <el-card class="content-card">
      <div class="autoops-table-wrapper">
        <el-table 
          :data="namespaceList" 
          style="width: 100%" class="autoops-table" 
          v-loading="loading"
          height="calc(100vh - 280px)"
          :row-class-name="tableRowClassName"
          :empty-text="emptyText"
        >
          <el-table-column prop="name" label="名称" min-width="180">
            <template #default="scope">
              <div class="namespace-name">
                <el-icon class="namespace-icon"><Folder /></el-icon>
                <span>{{ scope.row.name }}</span>
                <el-tag v-if="isSystemNamespace(scope.row.name)" type="info" size="small">系统</el-tag>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="120" align="center">
            <template #default="scope">
              <el-tag :type="getStatusType(scope.row.status)" effect="light">
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="180" align="center">
            <template #default="scope">
              <div class="time-info">
                <div>{{ formatDate(scope.row.creationTimestamp) }}</div>
                <div class="age">{{ calculateAge(scope.row.age) }}</div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="标签" min-width="200">
            <template #default="scope">
              <div class="labels-container">
                <el-tag 
                  v-for="(value, key) in scope.row.labels" 
                  :key="key"
                  size="small"
                  class="label-tag"
                  effect="light"
                >
                  {{ key }}: {{ value }}
                </el-tag>
                <span v-if="!scope.row.labels || Object.keys(scope.row.labels).length === 0" class="no-labels">
                  无标签
                </span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" align="center" fixed="right">
            <template #default="scope">
              <el-button 
                size="small" 
                type="danger" 
                @click="handleDelete(scope.row)"
                :disabled="isSystemNamespace(scope.row.name)"
                link
               v-show="permStore.hasPerm('k8s:namespace:handledelete')" >
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <el-dialog v-model="showCreateDialog" title="创建命名空间" width="500px">
      <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入命名空间名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="submitting">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Folder } from '@element-plus/icons-vue'
import { getNamespaceList, createNamespace, deleteNamespace } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'

const loading = ref(false)
const submitting = ref(false)
const showCreateDialog = ref(false)
const namespaceList = ref([])
const createFormRef = ref(null)

const createForm = ref({
  name: ''
})

const createRules = {
  name: [
    { required: true, message: '请输入命名空间名称', trigger: 'blur' },
    { pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/, message: '名称只能包含小写字母、数字和连字符', trigger: 'blur' }
  ]
}

const systemNamespaces = ['default', 'kube-system', 'kube-public', 'kube-node-lease']

const emptyText = computed(() => {
  return loading.value ? '加载中...' : '暂无命名空间数据'
})

const isSystemNamespace = (name) => {
  return systemNamespaces.includes(name)
}

const getStatusType = (status) => {
  switch (status) {
    case 'Active':
      return 'success'
    case 'Terminating':
      return 'warning'
    default:
      return 'info'
  }
}

const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const calculateAge = (age) => {
  if (!age) return '-'
  const now = Date.now()
  const created = age * 1000 // 转换为毫秒
  const diff = now - created
  
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
  
  if (days > 0) {
    return `${days}天前`
  } else if (hours > 0) {
    return `${hours}小时前`
  } else {
    return '刚刚创建'
  }
}

const tableRowClassName = ({ row }) => {
  if (isSystemNamespace(row.name)) {
    return 'system-namespace-row'
  }
  return ''
}

const fetchNamespaceList = async () => {
  loading.value = true
  try {
    const instanceId = getSelectedInstanceId() || '1'
    const response = await getNamespaceList(instanceId)
    namespaceList.value = response.data?.namespaceList || []
  } catch (error) {
    ElMessage.error('获取命名空间列表失败: ' + (error.response?.data?.message || error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

const handleCreate = async () => {
  if (!createFormRef.value) return
  
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      const instanceId = getSelectedInstanceId() || '1'
      await createNamespace(createForm.value.name, createForm.value, instanceId)
      ElMessage.success('命名空间创建成功')
      showCreateDialog.value = false
      createForm.value = { name: '' }
      await fetchNamespaceList()
    } catch (error) {
      ElMessage.error('创建命名空间失败: ' + (error.response?.data?.message || error.message || '未知错误'))
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (namespace) => {
  if (isSystemNamespace(namespace.name)) {
    ElMessage.warning('系统命名空间不能删除')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要删除命名空间 "${namespace.name}" 吗？此操作将删除该命名空间下的所有资源。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const instanceId = getSelectedInstanceId() || '1'
    await deleteNamespace(namespace.name, instanceId)
    ElMessage.success('命名空间删除成功')
    await fetchNamespaceList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除命名空间失败: ' + (error.response?.data?.message || error.message || '未知错误'))
    }
  }
}

onMounted(() => {
  fetchNamespaceList()
})
</script>

<style scoped>
.namespace-management {
  padding: 20px;
  height: 100vh;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
}

.page-header-card {
  margin-bottom: 20px;
  flex-shrink: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-header h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.page-header p {
  margin: 0;
  color: var(--el-text-color-secondary);
  font-size: 14px;
}

.content-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.table-container {
  flex: 1;
  overflow: hidden;
}

/* 表格行样式优化 */
:deep(.el-table) {
  font-size: 14px;
}

:deep(.el-table__header) {
  background-color: #f5f7fa;
}

:deep(.el-table th) {
  background-color: #f5f7fa !important;
  color: #606266;
  font-weight: 600;
  padding: 12px 0;
}

:deep(.el-table td) {
  padding: 12px 0;
}

/* 系统命名空间行样式 */
:deep(.system-namespace-row) {
  background-color: #f8f9fa;
}

:deep(.system-namespace-row:hover) {
  background-color: #f0f2f5 !important;
}

/* 命名空间名称样式 */
.namespace-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.namespace-icon {
  color: #409eff;
  font-size: 16px;
}

/* 时间信息样式 */
.time-info {
  text-align: center;
}

.time-info .age {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

/* 标签容器样式 */
.labels-container {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.label-tag {
  font-size: 11px;
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.no-labels {
  color: #c0c4cc;
  font-size: 12px;
}

/* 操作按钮样式优化 */
:deep(.el-button--danger.is-link) {
  padding: 4px 8px;
}

:deep(.el-button--danger.is-link:hover) {
  background-color: #fef0f0;
  color: #f56c6c;
}

:deep(.el-button--danger.is-link:disabled) {
  color: #c0c4cc;
  background-color: transparent;
}

/* 状态标签样式优化 */
:deep(.el-tag--light) {
  font-weight: 500;
}

:deep(.el-tag--success) {
  background-color: #f0f9ff;
  border-color: #b3e0ff;
  color: #1890ff;
}

:deep(.el-tag--warning) {
  background-color: #fdf6ec;
  border-color: #f7ba2a;
  color: #e6a23c;
}

:deep(.el-tag--info) {
  background-color: #f4f4f5;
  border-color: #e9e9eb;
  color: #909399;
}

/* 滚动条样式优化 */
:deep(.el-table__body-wrapper) {
  scrollbar-width: thin;
  scrollbar-color: #c1c1c1 transparent;
}

:deep(.el-table__body-wrapper::-webkit-scrollbar) {
  width: 6px;
  height: 6px;
}

:deep(.el-table__body-wrapper::-webkit-scrollbar-track) {
  background: transparent;
}

:deep(.el-table__body-wrapper::-webkit-scrollbar-thumb) {
  background-color: #c1c1c1;
  border-radius: 3px;
}

:deep(.el-table__body-wrapper::-webkit-scrollbar-thumb:hover) {
  background-color: #a8a8a8;
}

/* 空状态样式 */
:deep(.el-table__empty-block) {
  min-height: 200px;
}

:deep(.el-table__empty-text) {
  color: #909399;
  font-size: 14px;
}
</style>
