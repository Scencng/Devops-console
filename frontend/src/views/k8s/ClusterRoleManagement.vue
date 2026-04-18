<template>
  <div class="clusterrole-management">
    <el-card class="page-header-card cyber-card">
      <div class="page-header cyber-header">
        <div>
          <h2>ClusterRole 管理</h2>
          <p>管理集群级别的 RBAC 角色权限规则（全局生效）</p>
        </div>
        <div class="header-actions">
          <el-button type="primary" class="autoops-btn-primary" @click="openCreateDialog" v-show="permStore.hasPerm('k8s:clusterrole:opencreatedialog')" >
            <el-icon><Plus /></el-icon>创建 ClusterRole
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card cyber-card">
      <div class="autoops-table-wrapper">
        <el-table
          :data="clusterRoleList"
          style="width: 100%"
          class="autoops-table"
          v-loading="loading"
          element-loading-background="rgba(0, 0, 0, 0.5)"
          height="calc(100vh - 280px)"
          :empty-text="loading ? '加载中...' : '暂无数据'"
        >
          <el-table-column prop="name" label="名称" min-width="250" show-overflow-tooltip />
          <el-table-column label="规则数" width="100" align="center">
            <template #default="scope">
              <el-tag type="info" size="small">{{ scope.row.rulesCount ?? 0 }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="是否聚合" width="100" align="center">
            <template #default="scope">
              <el-tag v-if="scope.row.aggregationRule" type="success" size="small">是</el-tag>
              <span v-else style="color: #c0c4cc;">—</span>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="180">
            <template #default="scope">{{ formatTimestamp(scope.row.age) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="scope">
              <el-button link type="primary" size="small" @click="handleViewDetail(scope.row)" v-show="permStore.hasPerm('k8s:clusterrole:handleviewdetail')" >详情</el-button>
              <el-button link type="info" size="small" @click="handleEdit(scope.row)" v-show="permStore.hasPerm('k8s:clusterrole:handleedit')" >编辑</el-button>
              <el-button link type="danger" size="small" @click="handleDelete(scope.row)" v-show="permStore.hasPerm('k8s:clusterrole:handledelete')" >删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog v-model="showFormDialog" :title="isEditMode ? '编辑 ClusterRole' : '创建 ClusterRole'" width="900px" :close-on-click-modal="false" @close="handleCloseDialog">
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="表单创建" name="form">
          <el-form :model="form" label-width="80px" style="padding: 10px 0;">
            <el-form-item label="名称" required>
              <el-input v-model="form.name" :disabled="isEditMode" placeholder="ClusterRole 名称" />
            </el-form-item>
            <el-divider content-position="left">
              权限规则 (Rules)
              <el-button type="primary" link @click="addRule" style="margin-left: 10px;">
                <el-icon><Plus /></el-icon> 添加规则
              </el-button>
            </el-divider>
            <div class="rules-editor">
              <div v-for="(rule, index) in form.rules" :key="index" class="rule-item">
                <div class="rule-header">
                  <span class="rule-index">规则 {{ index + 1 }}</span>
                  <el-button link type="danger" size="small" @click="removeRule(index)"><el-icon><Delete /></el-icon></el-button>
                </div>
                <el-row :gutter="12">
                  <el-col :span="8">
                    <el-form-item label="API 组" label-width="70px">
                      <el-select v-model="rule.apiGroups" multiple filterable allow-create placeholder="如: apps, *" style="width:100%;">
                        <el-option label='""（核心组）' value="" />
                        <el-option label="apps" value="apps" />
                        <el-option label="batch" value="batch" />
                        <el-option label="extensions" value="extensions" />
                        <el-option label="networking.k8s.io" value="networking.k8s.io" />
                        <el-option label="rbac.authorization.k8s.io" value="rbac.authorization.k8s.io" />
                        <el-option label="storage.k8s.io" value="storage.k8s.io" />
                        <el-option label="autoscaling" value="autoscaling" />
                        <el-option label="*" value="*" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :span="8">
                    <el-form-item label="资源" label-width="50px">
                      <el-select v-model="rule.resources" multiple filterable allow-create placeholder="如: pods, *" style="width:100%;">
                        <el-option label="pods" value="pods" />
                        <el-option label="services" value="services" />
                        <el-option label="deployments" value="deployments" />
                        <el-option label="replicasets" value="replicasets" />
                        <el-option label="daemonsets" value="daemonsets" />
                        <el-option label="statefulsets" value="statefulsets" />
                        <el-option label="configmaps" value="configmaps" />
                        <el-option label="secrets" value="secrets" />
                        <el-option label="namespaces" value="namespaces" />
                        <el-option label="nodes" value="nodes" />
                        <el-option label="persistentvolumes" value="persistentvolumes" />
                        <el-option label="persistentvolumeclaims" value="persistentvolumeclaims" />
                        <el-option label="ingresses" value="ingresses" />
                        <el-option label="roles" value="roles" />
                        <el-option label="rolebindings" value="rolebindings" />
                        <el-option label="clusterroles" value="clusterroles" />
                        <el-option label="clusterrolebindings" value="clusterrolebindings" />
                        <el-option label="serviceaccounts" value="serviceaccounts" />
                        <el-option label="jobs" value="jobs" />
                        <el-option label="cronjobs" value="cronjobs" />
                        <el-option label="*" value="*" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :span="8">
                    <el-form-item label="操作" label-width="50px">
                      <el-select v-model="rule.verbs" multiple filterable allow-create placeholder="选择操作" style="width:100%;">
                        <el-option label="get" value="get" />
                        <el-option label="list" value="list" />
                        <el-option label="watch" value="watch" />
                        <el-option label="create" value="create" />
                        <el-option label="update" value="update" />
                        <el-option label="patch" value="patch" />
                        <el-option label="delete" value="delete" />
                        <el-option label="deletecollection" value="deletecollection" />
                        <el-option label="*" value="*" />
                      </el-select>
                    </el-form-item>
                  </el-col>
                </el-row>
              </div>
              <el-empty v-if="form.rules.length === 0" description="暂无规则，请点击上方按钮添加" :image-size="60" />
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
    <el-drawer v-model="showDetailDrawer" title="ClusterRole 详情" size="640px" direction="rtl">
      <div v-if="currentDetail" class="detail-content">
        <el-descriptions :column="2" border class="detail-desc">
          <el-descriptions-item label="名称">{{ currentDetail.metadata?.name }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTimestamp(currentDetail.metadata?.creationTimestamp) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left" style="margin-top: 20px;">权限规则</el-divider>
        <div v-for="(rule, idx) in (currentDetail.rules || [])" :key="idx" class="rule-detail-item">
          <div class="rule-detail-header">规则 {{ idx + 1 }}</div>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="API 组">
              <el-tag v-for="g in rule.apiGroups" :key="g" size="small" style="margin-right: 4px;">{{ g || '""' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="资源">
              <el-tag v-for="r in rule.resources" :key="r" size="small" type="success" style="margin-right: 4px;">{{ r }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="操作">
              <el-tag v-for="v in rule.verbs" :key="v" size="small" type="warning" style="margin-right: 4px;">{{ v }}</el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </div>
        <el-empty v-if="!currentDetail.rules || currentDetail.rules.length === 0" description="无规则" :image-size="60" />

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
  getClusterRoleList, getClusterRoleDetail, createClusterRole, updateClusterRole, deleteClusterRole
} from '@/api/k8s/rbac'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const loading = ref(false)
const submitting = ref(false)
const showFormDialog = ref(false)
const showDetailDrawer = ref(false)
const isEditMode = ref(false)
const editingRow = ref(null)
const clusterRoleList = ref([])
const currentDetail = ref(null)
const activeTab = ref('form')

const defaultForm = () => ({ name: '', rules: [], yaml: '' })
const form = ref(defaultForm())

const formatTimestamp = (ts) => {
  if (!ts) return '-'
  if (typeof ts === 'number') return dayjs.unix(ts).format('YYYY-MM-DD HH:mm:ss')
  return dayjs(ts).format('YYYY-MM-DD HH:mm:ss')
}

const fetchData = async () => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) return
  loading.value = true
  try {
    const res = await getClusterRoleList(instanceId)
    if (res.status === 200) clusterRoleList.value = res.data.clusterRoleList || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const addRule = () => {
  form.value.rules.push({ apiGroups: [''], resources: [], verbs: [] })
}

const removeRule = (index) => {
  form.value.rules.splice(index, 1)
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
  const res = await getClusterRoleDetail(row.name, instanceId)
  if (res.status === 200) {
    const detail = res.data.clusterRoleDetail
    editingRow.value = row
    isEditMode.value = true
    form.value = {
      name: detail.metadata.name,
      rules: (detail.rules || []).map(r => ({
        apiGroups: r.apiGroups || [],
        resources: r.resources || [],
        verbs: r.verbs || []
      })),
      yaml: ''
    }
    activeTab.value = 'form'
    showFormDialog.value = true
  }
}

const handleViewDetail = async (row) => {
  const instanceId = getSelectedInstanceId()
  const res = await getClusterRoleDetail(row.name, instanceId)
  if (res.status === 200) {
    currentDetail.value = res.data.clusterRoleDetail
    showDetailDrawer.value = true
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除 ClusterRole "${row.name}" 吗？`, '警告', { type: 'warning' }).then(async () => {
    const instanceId = getSelectedInstanceId()
    await deleteClusterRole(row.name, instanceId)
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
        await updateClusterRole(editingRow.value.name, payload, instanceId)
      } else {
        await createClusterRole(payload, instanceId)
      }
    } else {
      const payload = { name: form.value.name, rules: form.value.rules }
      if (isEditMode.value) {
        await updateClusterRole(editingRow.value.name, payload, instanceId)
      } else {
        await createClusterRole(payload, instanceId)
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
.clusterrole-management {
  padding: 20px;
}
.page-header-card { margin-bottom: 0px !important; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.header-actions { display: flex; gap: 10px; align-items: center; }
.rules-editor { max-height: 420px; overflow-y: auto; padding-right: 4px; }
.rule-item {
  margin-bottom: 14px;
  padding: 14px 16px 4px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fafafa;
}
html.dark .rule-item {
  background: rgba(255,255,255,0.03);
  border-color: rgba(255,255,255,0.08);
}
.rule-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.rule-index { font-size: 13px; font-weight: 600; color: #409eff; }
.detail-content { padding: 4px 0; }
.rule-detail-item { margin-bottom: 16px; }
.rule-detail-header { font-size: 13px; font-weight: 600; color: #409eff; margin-bottom: 8px; }
.labels-section { padding: 4px 0; }
</style>
