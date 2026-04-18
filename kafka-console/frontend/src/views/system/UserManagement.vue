<template>
  <div class="page-container">
    <el-card class="page-header-card mb-20" shadow="never">
      <div class="page-header">
        <div class="header-left">
          <div class="header-icon-wrapper">
            <el-icon :size="24"><User /></el-icon>
          </div>
          <div class="header-title-wrapper">
            <h2>用户管理</h2>
            <p class="subtitle">管理系统用户及权限分配</p>
          </div>
        </div>
        <div class="header-right">
          <el-button v-if="permStore.hasPerm('sys:user:add')" :icon="Plus" type="primary" @click="openDialog()">新增用户</el-button>
          <el-button :icon="Refresh" @click="fetchList">刷新</el-button>
        </div>
      </div>
    </el-card>

    <!-- Filters -->
    <el-card class="content-card mb-20" shadow="never">
      <div class="filter-row flex items-center gap-4">
        <el-input
          v-model="searchForm.username"
          :prefix-icon="Search"
          class="filter-item w-64"
          clearable
          placeholder="搜索用户名"
          @keyup.enter="fetchList"
        />
        <el-input
          v-model="searchForm.nickname"
          :prefix-icon="Search"
          class="filter-item w-40"
          clearable
          placeholder="搜索昵称"
          @keyup.enter="fetchList"
        />
        <el-select v-model="searchForm.status" class="filter-item w-32" clearable placeholder="状态筛选">
          <el-option :value="1" label="启用" />
          <el-option :value="0" label="停用" />
        </el-select>
        <div class="flex-spacer flex-1"></div>
        <el-button-group>
          <el-button :icon="Search" :loading="loading" type="primary" @click="fetchList">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="resetSearch">重置</el-button>
        </el-button-group>
      </div>
    </el-card>

    <!-- Table View -->
    <el-card class="content-card" shadow="never">
      <el-table v-loading="loading" :data="tableData" stripe style="width: 100%">
        <el-table-column label="用户名" min-width="120" prop="username">
          <template #default="{ row }">
            <span class="font-medium text-main">{{ row.username }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="nickname" label="昵称" min-width="120" />
        <el-table-column prop="deptName" label="部门" min-width="120" />
        <el-table-column label="邮箱" min-width="160" prop="email" show-overflow-tooltip />
        <el-table-column prop="phone" label="手机号" min-width="130" />
        <el-table-column label="角色" min-width="150">
          <template #default="{ row }">
            <el-tag v-for="r in row.roles" :key="r.id" :type="r.id === 1 ? 'danger' : 'primary'" class="mr-1" effect="light" size="small">{{ r.name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" :disabled="!permStore.hasPerm('sys:user:status')" @change="(v) => handleStatusChange(row, v)" />
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="160" prop="createdAt">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button v-if="permStore.hasPerm('sys:user:edit')" link size="small" type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button v-if="permStore.hasPerm('sys:user:resetpwd')" link size="small" type="warning" @click="handleResetPwd(row)">重置密码</el-button>
            <el-button v-if="permStore.hasPerm('sys:user:del')" link size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <!-- 新增/编辑 Dialog -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="580px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="用户名" prop="username">
              <el-input v-model="form.username" :disabled="!!form.id" placeholder="登录用户名" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="昵称">
              <el-input v-model="form.nickname" placeholder="显示昵称" />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="!form.id">
            <el-form-item label="密码" prop="password">
              <el-input v-model="form.password" type="password" show-password placeholder="至少6位" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="部门">
              <el-tree-select
                v-model="form.deptId"
                :data="deptOptions"
                :props="{ label: 'name', value: 'id', children: 'children' }"
                placeholder="选择部门"
                clearable
                style="width:100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="岗位">
              <el-select v-model="form.positionIds" multiple placeholder="选择岗位" style="width:100%">
                <el-option v-for="p in positionOptions" :key="p.id" :label="p.name" :value="p.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="角色">
              <el-select v-model="form.roleIds" multiple placeholder="选择角色" style="width:100%">
                <el-option v-for="r in roleOptions" :key="r.id" :label="r.name" :value="r.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="手机号">
              <el-input v-model="form.phone" placeholder="手机号" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱">
              <el-input v-model="form.email" placeholder="邮箱" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-radio-group v-model="form.status">
                <el-radio :value="1">启用</el-radio>
                <el-radio :value="0">停用</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="备注" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>

    <!-- 重置密码 Dialog -->
    <el-dialog v-model="resetPwdVisible" title="重置密码" width="380px">
      <el-form :model="resetPwdForm" :rules="resetPwdRules" ref="resetPwdRef" label-width="80px">
        <el-form-item label="新密码" prop="password">
          <el-input v-model="resetPwdForm.password" type="password" show-password placeholder="至少6位" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetPwdVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="confirmResetPwd">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import {onMounted, reactive, ref} from 'vue'
import {Plus, Refresh, RefreshLeft, Search, User} from '@element-plus/icons-vue'
import {ElMessage, ElMessageBox} from 'element-plus'
import {SHA256} from 'crypto-js'
import dayjs from 'dayjs'
import {
  createUser,
  deleteUser,
  getUserList,
  resetUserPassword,
  updateUser,
  updateUserStatus
} from '@/api/system/user.js'
import {getDeptTree} from '@/api/system/dept.js'
import {getAllPositions} from '@/api/system/position.js'
import {getAllRoles} from '@/api/system/role.js'
import {usePermissionStore} from '@/stores/permissionStore.js'

const permStore = usePermissionStore()
const loading = ref(false)
const submitting = ref(false)
const tableData = ref([])
const deptOptions = ref([])
const positionOptions = ref([])
const roleOptions = ref([])

const formatTime = (time) => time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'

const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const searchForm = reactive({ username: '', nickname: '', status: null })

const dialogVisible = ref(false)
const dialogTitle = ref('新增用户')
const formRef = ref(null)
const form = reactive({ id: null, username: '', password: '', nickname: '', email: '', phone: '', deptId: null, positionIds: [], roleIds: [], status: 1, remark: '' })
const formRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }, { min: 3, max: 50, message: '3-50个字符', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '至少6位', trigger: 'blur' }]
}

const resetPwdVisible = ref(false)
const resetPwdRef = ref(null)
const resetPwdForm = reactive({ userId: null, password: '' })
const resetPwdRules = { password: [{ required: true, message: '请输入新密码' }, { min: 6, message: '至少6位' }] }

onMounted(() => {
  fetchList()
  loadOptions()
})

async function fetchList() {
  loading.value = true
  try {
    const params = { page: pagination.page, pageSize: pagination.pageSize, ...searchForm }
    if (params.status === null) delete params.status
    const res = await getUserList(params)
    const d = res?.data?.data || res?.data || {}
    tableData.value = d.list || []
    pagination.total = d.total || 0
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || e?.message || '获取用户列表失败')
  } finally {
    loading.value = false
  }
}

async function loadOptions() {
  try {
    const [deptRes, posRes, roleRes] = await Promise.all([getDeptTree(), getAllPositions(), getAllRoles()])
    deptOptions.value = deptRes?.data?.data || []
    positionOptions.value = posRes?.data?.data || []
    roleOptions.value = roleRes?.data?.data || []
  } catch (e) {
    console.warn('获取选项数据失败:', e?.message)
  }
}

function resetSearch() {
  searchForm.username = ''
  searchForm.nickname = ''
  searchForm.status = null
  pagination.page = 1
  fetchList()
}

function openDialog(row = null) {
  Object.assign(form, { id: null, username: '', password: '', nickname: '', email: '', phone: '', deptId: null, positionIds: [], roleIds: [], status: 1, remark: '' })
  if (row) {
    Object.assign(form, {
      id: row.id, username: row.username, nickname: row.nickname || '',
      email: row.email || '', phone: row.phone || '', deptId: row.deptId || null,
      positionIds: (row.positions || []).map(p => p.id),
      roleIds: (row.roles || []).map(r => r.id),
      status: row.status, remark: row.remark || ''
    })
    dialogTitle.value = '编辑用户'
  } else {
    dialogTitle.value = '新增用户'
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    const payload = { ...form }
    if (!payload.id) {
      payload.password = SHA256(payload.password).toString()
    }
    if (payload.id) {
      await updateUser(payload.id, payload)
    } else {
      await createUser(payload)
    }
    ElMessage.success('操作成功')
    dialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function handleStatusChange(row, val) {
  try {
    await updateUserStatus(row.id, { status: val ? 1 : 0 })
    row.status = val ? 1 : 0
    ElMessage.success('状态更新成功')
  } catch {
    ElMessage.error('操作失败')
  }
}

function handleResetPwd(row) {
  resetPwdForm.userId = row.id
  resetPwdForm.password = ''
  resetPwdVisible.value = true
}

async function confirmResetPwd() {
  await resetPwdRef.value?.validate()
  submitting.value = true
  try {
    await resetUserPassword(resetPwdForm.userId, { password: SHA256(resetPwdForm.password).toString() })
    ElMessage.success('密码重置成功')
    resetPwdVisible.value = false
  } catch {
    ElMessage.error('重置失败')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(row) {
  await ElMessageBox.confirm(`确认删除用户「${row.username}」？`, '提示', { type: 'warning' })
  try {
    await deleteUser(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch {
    ElMessage.error('删除失败')
  }
}
</script>

<style scoped>
.page-container { padding: 16px; display: flex; flex-direction: column; gap: 0; }
.header-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: var(--primary-color, #409eff);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-left { display: flex; align-items: center; gap: 12px; }
.header-title-wrapper h2 { margin: 0; font-size: 18px; color: var(--text-main, #303133); }
.header-title-wrapper .subtitle { margin: 0; font-size: 12px; color: var(--text-sub, #909399); }
.header-right { display: flex; align-items: center; gap: 12px; }

.flex { display: flex; }
.items-center { align-items: center; }
.gap-4 { gap: 16px; }
.mb-20 { margin-bottom: 20px; }
.w-64 { width: 16rem; }
.w-40 { width: 10rem; }
.w-32 { width: 8rem; }
.flex-1 { flex: 1; }
.mr-1 { margin-right: 4px; }
.pagination { margin-top: 16px; display: flex; justify-content: flex-end; }
.text-main { color: var(--text-main, #303133); }
.font-medium { font-weight: 500; }
</style>
