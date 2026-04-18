<template>
  <div class="page-container">
    <el-card class="page-header-card">
       <div class="page-header">
         <div class="header-left">
           <div class="header-icon-wrapper">
             <el-icon :size="24"><Coin /></el-icon>
           </div>
           <div class="header-title-wrapper">
             <h2>Helm 仓库管理</h2>
             <p class="subtitle">管理 Helm Chart 仓库</p>
           </div>
         </div>
         <div class="header-right">
            <el-button type="primary" @click="showCreateDialog = true" :icon="Plus">添加仓库</el-button>
         </div>
       </div>
    </el-card>

    <el-card class="content-card">
       <el-table
         :data="repoList"
         style="width: 100%"
         v-loading="loading"
         stripe
       >
         <el-table-column prop="name" label="仓库名称" min-width="180">
            <template #default="{ row }">
               <span class="font-medium text-main">{{ row.name }}</span>
            </template>
         </el-table-column>
         <el-table-column prop="url" label="URL" min-width="300" show-overflow-tooltip />
         <el-table-column prop="username" label="用户名" width="150">
           <template #default="scope">
             <span v-if="scope.row.username">{{ scope.row.username }}</span>
             <span v-else class="text-sub">-</span>
           </template>
         </el-table-column>
         <el-table-column prop="created_at" label="创建时间" width="180">
           <template #default="scope">
             {{ formatTime(scope.row.created_at) }}
           </template>
         </el-table-column>
         <el-table-column label="操作" width="200" fixed="right">
           <template #default="scope">
             <el-button link type="primary" size="small" @click="handleSync(scope.row)">
               <el-icon><Refresh /></el-icon> 同步
             </el-button>
             <el-button link type="primary" size="small" @click="handleEdit(scope.row)">编辑</el-button>
             <el-button link type="danger" size="small" @click="handleDelete(scope.row)">删除</el-button>
           </template>
         </el-table-column>
       </el-table>
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="isEditMode ? '编辑仓库' : '添加仓库'"
      width="600px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form :model="repoForm" ref="repoFormRef" label-position="top" @submit.prevent>
        <el-form-item label="仓库名称" prop="name" required>
          <el-input v-model="repoForm.name" placeholder="例如：bitnami" :disabled="isEditMode" />
        </el-form-item>
        <el-form-item label="仓库 URL" prop="url" required>
          <el-input v-model="repoForm.url" placeholder="https://charts.bitnami.com/bitnami" />
        </el-form-item>
        <el-row :gutter="20">
            <el-col :span="12">
                <el-form-item label="用户名">
                  <el-input v-model="repoForm.username" placeholder="私有仓库认证（可选）" />
                </el-form-item>
            </el-col>
            <el-col :span="12">
                <el-form-item label="密码">
                  <el-input v-model="repoForm.password" type="password" placeholder="私有仓库密码（可选）" show-password />
                </el-form-item>
            </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="handleCloseDialog">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">
            {{ isEditMode ? '更新' : '创建' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElLoading } from 'element-plus'
import { Plus, Refresh, Coin } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import {
  getHelmRepos,
  createHelmRepo,
  updateHelmRepo,
  deleteHelmRepo,
  syncHelmRepo
} from '@/api/helm'

const loading = ref(false)
const submitting = ref(false)
const repoList = ref([])
const showCreateDialog = ref(false)
const isEditMode = ref(false)
const editingRepo = ref(null)
const repoFormRef = ref(null)

const repoForm = ref({
  name: '',
  url: '',
  username: '',
  password: ''
})

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getHelmRepos()
    repoList.value = res.data.repoList || []
  } catch (error) {
    console.error(error)
    ElMessage.error('获取仓库列表失败')
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  repoForm.value = {
    name: '',
    url: '',
    username: '',
    password: ''
  }
  isEditMode.value = false
  editingRepo.value = null
}

const handleCloseDialog = () => {
  showCreateDialog.value = false
  resetForm()
}

const handleCreate = async () => {
  submitting.value = true
  try {
    await createHelmRepo(repoForm.value)
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
  submitting.value = true
  try {
    await updateHelmRepo(editingRepo.value.id, repoForm.value)
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

const handleEdit = (row) => {
  isEditMode.value = true
  editingRepo.value = row
  repoForm.value = {
    name: row.name,
    url: row.url,
    username: row.username || '',
    password: '' // 密码不回显
  }
  showCreateDialog.value = true
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定要删除仓库 "${row.name}" 吗？这将同时删除该仓库的所有 Chart 缓存。`, '删除确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deleteHelmRepo(row.id)
      ElMessage.success('删除成功')
      fetchData()
    } catch (error) {
      console.error(error)
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

const handleSync = (row) => {
  ElMessageBox.confirm(
    `确定要同步仓库 "${row.name}" 吗？这将从仓库获取最新的 Chart 列表。`,
    '同步确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'info'
    }
  )
  .then(() => {
    doSync(row)
  })
  .catch(() => {})
}

const doSync = async (row) => {
  const loadingInstance = ElLoading.service({
    lock: true,
    text: `正在同步仓库 "${row.name}"，请稍候...`,
    background: 'rgba(255, 255, 255, 0.7)',
  })

  try {
    await syncHelmRepo(row.id)
    loadingInstance.close()
    ElMessage.success('同步成功')
    fetchData()
  } catch (error) {
    loadingInstance.close()
    console.error('Sync error:', error)
    ElMessage.error(error.message || '同步失败')
  }
}

const formatTime = (timestamp) => {
  return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.header-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: var(--primary-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.text-main { color: var(--text-main); }
.text-sub { color: var(--text-sub); }
.font-medium { font-weight: 500; }

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-left { display: flex; align-items: center; gap: 12px; }
.header-title-wrapper h2 { margin: 0; font-size: 18px; color: var(--text-main); }
.header-title-wrapper .subtitle { margin: 0; font-size: 12px; color: var(--text-sub); }
</style>
