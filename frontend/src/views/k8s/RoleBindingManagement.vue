<template>
  <div class="rolebinding-management">
    <el-card class="page-header-card cyber-card">
      <div class="page-header cyber-header">
        <div>
          <h2>RoleBinding 管理</h2>
          <p>将 Role 或 ClusterRole 绑定到命名空间内的用户、组或 ServiceAccount</p>
        </div>
        <div class="header-actions">
          <div class="namespace-selector">
            <span class="selector-label">命名空间：</span>
            <el-select v-model="selectedNamespace" placeholder="选择命名空间" @change="fetchData" style="width: 200px;" class="autoops-select">
              <el-option label="所有" value="all" />
              <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
            </el-select>
          </div>
          <el-button type="primary" class="autoops-btn-primary" @click="openCreateDialog" v-show="permStore.hasPerm('k8s:rolebinding:opencreatedialog')" >
            <el-icon><Plus /></el-icon>创建 RoleBinding
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card cyber-card">
      <div class="autoops-table-wrapper">
        <el-table
          :data="roleBindingList"
          style="width: 100%"
          class="autoops-table"
          v-loading="loading"
          element-loading-background="rgba(0, 0, 0, 0.5)"
          height="calc(100vh - 280px)"
          :empty-text="loading ? '加载中...' : '暂无数据'"
        >
          <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip />
          <el-table-column prop="namespace" label="命名空间" width="160" />
          <el-table-column label="角色引用" width="200" show-overflow-tooltip>
            <template #default="scope">
              <el-tag :type="scope.row.roleRefKind === 'ClusterRole' ? 'warning' : 'primary'" size="small">
                {{ scope.row.roleRefKind }}
              </el-tag>
              <span style="margin-left: 6px; font-size: 13px;">{{ scope.row.roleRefName }}</span>
            </template>
          </el-table-column>
          <el-table-column label="主体数" width="100" align="center">
            <template #default="scope">
              <el-tag type="info" size="small">{{ scope.row.subjectsCount ?? 0 }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="180">
            <template #default="scope">{{ formatTimestamp(scope.row.age) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="scope">
              <el-button link type="primary" size="small" @click="handleViewDetail(scope.row)" v-show="permStore.hasPerm('k8s:rolebinding:handleviewdetail')" >详情</el-button>
              <el-button link type="info" size="small" @click="handleEdit(scope.row)" v-show="permStore.hasPerm('k8s:rolebinding:handleedit')" >编辑</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:rolebinding:handledelete')" >删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog v-model="showFormDialog" :title="isEditMode ? '编辑 RoleBinding' : '创建 RoleBinding'" width="900px" :close-on-click-modal="false" @close="handleCloseDialog">
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="表单创建" name="form">
          <el-form :model="form" label-width="110px" style="padding: 10px 0;">
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="命名空间" required>
                  <el-select v-model="form.namespace" style="width: 100%;" :disabled="isEditMode">
                    <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="名称" required>
                  <el-input v-model="form.name" :disabled="isEditMode" placeholder="RoleBinding 名称" />
                </el-form-item>
              </el-col>
            </el-row>

            <el-divider content-position="left">角色引用 (roleRef)</el-divider>
            <el-row :gutter="20">
              <el-col :span="10">
                <el-form-item label="角色类型" label-width="90px" required>
                  <el-select v-model="form.roleRefKind" style="width: 100%;" :disabled="isEditMode">
                    <el-option label="Role（命名空间级）" value="Role" />
                    <el-option label="ClusterRole（集群级）" value="ClusterRole" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="14">
                <el-form-item label="角色名称" label-width="90px" required>
                  <el-input v-model="form.roleRefName" :disabled="isEditMode" placeholder="角色名称" />
                </el-form-item>
              </el-col>
            </el-row>

            <el-divider content-position="left">
              主体 (Subjects)
              <el-button type="primary" link @click="addSubject" style="margin-left: 10px;">
                <el-icon><Plus /></el-icon> 添加主体
              </el-button>
            </el-divider>
            <div class="subjects-editor">
              <div v-for="(subj, index) in form.subjects" :key="index" class="subject-item">
                <div class="subject-header">
                  <span class="subject-index">主体 {{ index + 1 }}</span>
                  <el-button link type="danger" size="small" @click="removeSubject(index)"><el-icon><Delete /></el-icon></el-button>
                </div>
                <el-row :gutter="12">
                  <el-col :span="7">
                    <el-form-item label="类型" label-width="60px">
                      <el-select v-model="subj.kind" style="width: 100%;">
                        <el-option label="User" value="User" />
                        <el-option label="Group" value="Group" />
                        <el-option label="ServiceAccount" value="ServiceAccount" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :span="9">
                    <el-form-item label="名称" label-width="60px">
                      <el-input v-model="subj.name" placeholder="用户/组/SA 名称" />
                    </el-form-item>
                  </el-col>
                  <el-col :span="8">
                    <el-form-item label="命名空间" label-width="80px">
                      <el-input v-model="subj.namespace" placeholder="仅 ServiceAccount 需要" :disabled="subj.kind !== 'ServiceAccount'" />
                    </el-form-item>
                  </el-col>
                </el-row>
              </div>
              <el-empty v-if="form.subjects.length === 0" description="暂无主体，请点击上方按钮添加" :image-size="60" />
            </div>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="YAML 创建" name="yaml">
          <el-input v-model="form.yaml" type="textarea" :rows="22" placeholder="请输入 YAML 配置"
            style="font-family: 'Courier New', monospace; font-size: 13px;" />
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="handleCloseDialog">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ isEditMode ? '更新' : '创建' }}</el-button>
      </template>
    </el-dialog>

    <!-- 详情抽屉 -->
    <el-drawer v-model="showDetailDrawer" title="RoleBinding 详情" size="640px" direction="rtl">
      <div v-if="currentDetail" class="detail-content">
        <el-descriptions :column="2" border class="detail-desc">
          <el-descriptions-item label="名称">{{ currentDetail.metadata?.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ currentDetail.metadata?.namespace }}</el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">{{ formatTimestamp(currentDetail.metadata?.creationTimestamp) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left" style="margin-top: 20px;">角色引用 (roleRef)</el-divider>
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="类型">
            <el-tag :type="currentDetail.roleRef?.kind === 'ClusterRole' ? 'warning' : 'primary'" size="small">
              {{ currentDetail.roleRef?.kind }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="名称">{{ currentDetail.roleRef?.name }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left" style="margin-top: 20px;">主体 (Subjects)</el-divider>
        <el-table :data="currentDetail.subjects || []" border size="small" style="width: 100%;">
          <el-table-column prop="kind" label="类型" width="140">
            <template #default="scope">
              <el-tag :type="subjectKindType(scope.row.kind)" size="small">{{ scope.row.kind }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="名称" />
          <el-table-column prop="namespace" label="命名空间" width="160">
            <template #default="scope">{{ scope.row.namespace || '—' }}</template>
          </el-table-column>
        </el-table>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { usePermissionStore } from '@/stores/permissionStore.js'
const permStore = usePermissionStore()

import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import {
  getRoleBindingList, getRoleBindingDetail, createRoleBinding, updateRoleBinding, deleteRoleBinding
} from '@/api/k8s/rbac'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const loading = ref(false)
const submitting = ref(false)
const showFormDialog = ref(false)
const showDetailDrawer = ref(false)
const isEditMode = ref(false)
const editingRow = ref(null)
const roleBindingList = ref([])
const namespaceList = ref([])
const selectedNamespace = ref('all')
const currentDetail = ref(null)
const activeTab = ref('form')

const defaultForm = () => ({
  name: '',
  namespace: 'default',
  roleRefKind: 'Role',
  roleRefName: '',
  subjects: [],
  yaml: ''
})
const form = ref(defaultForm())

const formatTimestamp = (ts) => {
  if (!ts) return '-'
  if (typeof ts === 'number') return dayjs.unix(ts).format('YYYY-MM-DD HH:mm:ss')
  return dayjs(ts).format('YYYY-MM-DD HH:mm:ss')
}

const subjectKindType = (kind) => ({ User: 'primary', Group: 'success', ServiceAccount: 'warning' }[kind] || 'info')

const fetchNamespaces = async () => {
  const instanceId = getSelectedInstanceId()
  const res = await getNamespaceList(instanceId)
  if (res.status === 200) namespaceList.value = res.data.namespaceList
}

const fetchData = async () => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) return
  loading.value = true
  try {
    const res = await getRoleBindingList(selectedNamespace.value, instanceId)
    if (res.status === 200) roleBindingList.value = res.data.roleBindingList || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const addSubject = () => {
  form.value.subjects.push({ kind: 'ServiceAccount', name: '', namespace: '' })
}

const removeSubject = (index) => {
  form.value.subjects.splice(index, 1)
}

const openCreateDialog = () => {
  isEditMode.value = false
  editingRow.value = null
  form.value = defaultForm()
  activeTab.value = 'form'
  showFormDialog.value = true
}

const handleEdit = async (row) => {
  const instanceId = getSelectedInstanceId()
  const res = await getRoleBindingDetail(row.namespace, row.name, instanceId)
  if (res.status === 200) {
    const detail = res.data.roleBindingDetail
    editingRow.value = row
    isEditMode.value = true
    form.value = {
      name: detail.metadata.name,
      namespace: detail.metadata.namespace,
      roleRefKind: detail.roleRef?.kind || 'Role',
      roleRefName: detail.roleRef?.name || '',
      subjects: (detail.subjects || []).map(s => ({ kind: s.kind, name: s.name, namespace: s.namespace || '' })),
      yaml: ''
    }
    activeTab.value = 'form'
    showFormDialog.value = true
  }
}

const handleViewDetail = async (row) => {
  const instanceId = getSelectedInstanceId()
  const res = await getRoleBindingDetail(row.namespace, row.name, instanceId)
  if (res.status === 200) {
    currentDetail.value = res.data.roleBindingDetail
    showDetailDrawer.value = true
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除 RoleBinding "${row.name}" 吗？`, '警告', { type: 'warning' }).then(async () => {
    const instanceId = getSelectedInstanceId()
    await deleteRoleBinding(row.namespace, row.name, instanceId)
    ElMessage.success('删除成功')
    fetchData()
  })
}

const handleCloseDialog = () => {
  showFormDialog.value = false
  form.value = defaultForm()
  isEditMode.value = false
  editingRow.value = null
}

const handleSubmit = async () => {
  const instanceId = getSelectedInstanceId()
  submitting.value = true
  try {
    if (activeTab.value === 'yaml') {
      const payload = { yaml: form.value.yaml }
      if (isEditMode.value) {
        await updateRoleBinding(editingRow.value.namespace, editingRow.value.name, payload, instanceId)
      } else {
        await createRoleBinding(payload, instanceId)
      }
    } else {
      const payload = {
        name: form.value.name,
        namespace: form.value.namespace,
        roleRef: { kind: form.value.roleRefKind, name: form.value.roleRefName, apiGroup: 'rbac.authorization.k8s.io' },
        subjects: form.value.subjects
      }
      if (isEditMode.value) {
        await updateRoleBinding(editingRow.value.namespace, editingRow.value.name, payload, instanceId)
      } else {
        await createRoleBinding(payload, instanceId)
      }
    }
    ElMessage.success(isEditMode.value ? '更新成功' : '创建成功')
    handleCloseDialog()
    fetchData()
  } catch (e) {
    console.error(e)
    ElMessage.error(isEditMode.value ? '更新失败' : '创建失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchNamespaces()
  fetchData()
})
</script>

<style scoped>
.rolebinding-management { padding: 20px; }
.page-header-card { margin-bottom: 0px !important; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.header-actions { display: flex; gap: 10px; align-items: center; }
.namespace-selector { display: flex; align-items: center; gap: 8px; }
.selector-label { font-size: 14px; color: #606266; white-space: nowrap; }
.subjects-editor { max-height: 400px; overflow-y: auto; padding-right: 4px; }
.subject-item {
  margin-bottom: 14px;
  padding: 14px 16px 4px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fafafa;
}
html.dark .subject-item {
  background: rgba(255,255,255,0.03);
  border-color: rgba(255,255,255,0.08);
}
.subject-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.subject-index { font-size: 13px; font-weight: 600; color: #67c23a; }
.detail-content { padding: 4px 0; }
</style>
