<template>
  <div class="clusterrolebinding-management">
    <el-card class="page-header-card cyber-card">
      <div class="page-header cyber-header">
        <div>
          <h2>ClusterRoleBinding 管理</h2>
          <p>将 ClusterRole 绑定到集群范围内的用户、组或 ServiceAccount</p>
        </div>
        <div class="header-actions">
          <el-button type="primary" class="autoops-btn-primary" @click="openCreateDialog" v-show="permStore.hasPerm('k8s:clusterrolebinding:opencreatedialog')" >
            <el-icon><Plus /></el-icon>创建 ClusterRoleBinding
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card cyber-card">
      <div class="autoops-table-wrapper">
        <el-table
          :data="clusterRoleBindingList"
          style="width: 100%"
          class="autoops-table"
          v-loading="loading"
          element-loading-background="rgba(0, 0, 0, 0.5)"
          height="calc(100vh - 280px)"
          :empty-text="loading ? '加载中...' : '暂无数据'"
        >
          <el-table-column prop="name" label="名称" min-width="220" show-overflow-tooltip />
          <el-table-column label="角色引用" width="220" show-overflow-tooltip>
            <template #default="scope">
              <el-tag type="warning" size="small">ClusterRole</el-tag>
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
              <el-button link type="primary" size="small" @click="handleViewDetail(scope.row)" v-show="permStore.hasPerm('k8s:clusterrolebinding:handleviewdetail')" >详情</el-button>
              <el-button link type="info" size="small" @click="handleEdit(scope.row)" v-show="permStore.hasPerm('k8s:clusterrolebinding:handleedit')" >编辑</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:clusterrolebinding:handledelete')" >删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog v-model="showFormDialog" :title="isEditMode ? '编辑 ClusterRoleBinding' : '创建 ClusterRoleBinding'" width="900px" :close-on-click-modal="false" @close="handleCloseDialog">
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="表单创建" name="form">
          <el-form :model="form" label-width="100px" style="padding: 10px 0;">
            <el-form-item label="名称" required>
              <el-input v-model="form.name" :disabled="isEditMode" placeholder="ClusterRoleBinding 名称" />
            </el-form-item>

            <el-divider content-position="left">角色引用 (roleRef)</el-divider>
            <el-row :gutter="20">
              <el-col :span="10">
                <el-form-item label="角色类型" required>
                  <el-select v-model="form.roleRefKind" style="width: 100%;" :disabled="isEditMode">
                    <el-option label="ClusterRole（集群级）" value="ClusterRole" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="14">
                <el-form-item label="角色名称" required>
                  <el-input v-model="form.roleRefName" :disabled="isEditMode" placeholder="ClusterRole 名称" />
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
                      <el-select v-model="subj.kind" style="width: 100%;" @change="onSubjectKindChange(subj)">
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
                      <el-input v-model="subj.namespace" placeholder="仅 SA 需要" :disabled="subj.kind !== 'ServiceAccount'" />
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
    <el-drawer v-model="showDetailDrawer" title="ClusterRoleBinding 详情" size="640px" direction="rtl">
      <div v-if="currentDetail" class="detail-content">
        <el-descriptions :column="2" border class="detail-desc">
          <el-descriptions-item label="名称">{{ currentDetail.metadata?.name }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTimestamp(currentDetail.metadata?.creationTimestamp) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left" style="margin-top: 20px;">角色引用 (roleRef)</el-divider>
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="类型">
            <el-tag type="warning" size="small">{{ currentDetail.roleRef?.kind }}</el-tag>
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

        <el-divider content-position="left" style="margin-top: 20px;">标签</el-divider>
        <div class="labels-section">
          <el-tag v-for="(val, key) in (currentDetail.metadata?.labels || {})" :key="key" size="small" type="info" style="margin: 2px;">{{ key }}={{ val }}</el-tag>
          <span v-if="!currentDetail.metadata?.labels || Object.keys(currentDetail.metadata.labels).length === 0" style="color: #909399; font-size: 13px;">无标签</span>
        </div>
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
  getClusterRoleBindingList, getClusterRoleBindingDetail,
  createClusterRoleBinding, updateClusterRoleBinding, deleteClusterRoleBinding
} from '@/api/k8s/rbac'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const loading = ref(false)
const submitting = ref(false)
const showFormDialog = ref(false)
const showDetailDrawer = ref(false)
const isEditMode = ref(false)
const editingRow = ref(null)
const clusterRoleBindingList = ref([])
const currentDetail = ref(null)
const activeTab = ref('form')

const defaultForm = () => ({
  name: '',
  roleRefKind: 'ClusterRole',
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

const onSubjectKindChange = (subj) => {
  if (subj.kind !== 'ServiceAccount') subj.namespace = ''
}

const fetchData = async () => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) return
  loading.value = true
  try {
    const res = await getClusterRoleBindingList(instanceId)
    if (res.status === 200) clusterRoleBindingList.value = res.data.clusterRoleBindingList || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const addSubject = () => {
  form.value.subjects.push({ kind: 'ServiceAccount', name: '', namespace: 'default' })
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
  const res = await getClusterRoleBindingDetail(row.name, instanceId)
  if (res.status === 200) {
    const detail = res.data.clusterRoleBindingDetail
    editingRow.value = row
    isEditMode.value = true
    form.value = {
      name: detail.metadata.name,
      roleRefKind: detail.roleRef?.kind || 'ClusterRole',
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
  const res = await getClusterRoleBindingDetail(row.name, instanceId)
  if (res.status === 200) {
    currentDetail.value = res.data.clusterRoleBindingDetail
    showDetailDrawer.value = true
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除 ClusterRoleBinding "${row.name}" 吗？`, '警告', { type: 'warning' }).then(async () => {
    const instanceId = getSelectedInstanceId()
    await deleteClusterRoleBinding(row.name, instanceId)
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
        await updateClusterRoleBinding(editingRow.value.name, payload, instanceId)
      } else {
        await createClusterRoleBinding(payload, instanceId)
      }
    } else {
      const payload = {
        name: form.value.name,
        roleRef: { kind: form.value.roleRefKind, name: form.value.roleRefName, apiGroup: 'rbac.authorization.k8s.io' },
        subjects: form.value.subjects
      }
      if (isEditMode.value) {
        await updateClusterRoleBinding(editingRow.value.name, payload, instanceId)
      } else {
        await createClusterRoleBinding(payload, instanceId)
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
  fetchData()
})
</script>

<style scoped>
.clusterrolebinding-management { padding: 20px; }
.page-header-card { margin-bottom: 0px !important; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.header-actions { display: flex; gap: 10px; align-items: center; }
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
.subject-index { font-size: 13px; font-weight: 600; color: #e6a23c; }
.detail-content { padding: 4px 0; }
.labels-section { padding: 4px 0; }
</style>
