<template>
  <div class="page-container">
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="角色名称">
          <el-input v-model="searchForm.name" placeholder="角色名称" clearable style="width:180px" />
        </el-form-item>
        <el-form-item label="角色编码">
          <el-input v-model="searchForm.code" placeholder="角色编码" clearable style="width:180px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width:120px">
            <el-option label="启用" :value="1" />
            <el-option label="停用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <div class="toolbar">
        <span class="title">角色列表</span>
        <el-button type="primary" :icon="Plus" @click="openDialog()">新增角色</el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column prop="name" label="角色名称" min-width="150" />
        <el-table-column prop="code" label="角色编码" min-width="150" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="180" />
        <el-table-column label="创建时间" min-width="160" prop="createdAt">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="190" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="warning" @click="openPermDialog(row)">权限分配</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination class="pagination" v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50]" :total="pagination.total" layout="total, sizes, prev, pager, next"
        @size-change="fetchList" @current-change="fetchList" />
    </el-card>

    <!-- 新增/编辑 Dialog -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="480px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="角色名称" />
        </el-form-item>
        <el-form-item label="角色编码" prop="code">
          <el-input v-model="form.code" placeholder="如：admin, operator" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" style="width:100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>

    <!-- 权限分配 Dialog -->
    <el-dialog v-model="permDialogVisible" title="权限分配" width="480px" destroy-on-close>
      <div class="perm-header">
        <span>为角色「{{ currentRole?.name }}」分配菜单权限</span>
      </div>
      <el-tree
        ref="menuTreeRef"
        :data="menuTree"
        show-checkbox
        node-key="id"
        :default-checked-keys="checkedMenuIds"
        :props="{ label: 'name', children: 'children' }"
        style="max-height: 400px; overflow-y: auto;"
      />
      <template #footer>
        <el-button @click="permDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleAssignMenus">保存权限</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import {onMounted, reactive, ref} from 'vue'
import {Plus} from '@element-plus/icons-vue'
import {ElMessage, ElMessageBox} from 'element-plus'
import dayjs from 'dayjs'
import {assignRoleMenus, createRole, deleteRole, getRoleList, getRoleMenuIds, updateRole} from '@/api/system/role.js'
import {getMenuTree} from '@/api/system/menu.js'

const loading = ref(false), submitting = ref(false)
const tableData = ref([])
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const searchForm = reactive({ name: '', code: '', status: null })
const dialogVisible = ref(false), dialogTitle = ref('新增角色')
const formRef = ref(null)
const form = reactive({ id: null, name: '', code: '', sort: 0, status: 1, remark: '' })
const formRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入角色编码', trigger: 'blur' }]
}

const permDialogVisible = ref(false)
const menuTreeRef = ref(null)
const menuTree = ref([])
const checkedMenuIds = ref([])
const currentRole = ref(null)

const formatTime = (time) => time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'

onMounted(fetchList)

async function fetchList() {
  loading.value = true
  try {
    const params = { page: pagination.page, pageSize: pagination.pageSize, ...searchForm }
    if (params.status === null) delete params.status
    const res = await getRoleList(params)
    const d = res?.data?.data || {}
    tableData.value = d.list || []
    pagination.total = d.total || 0
  } finally { loading.value = false }
}

function resetSearch() {
  Object.assign(searchForm, { name: '', code: '', status: null })
  pagination.page = 1
  fetchList()
}

function openDialog(row = null) {
  Object.assign(form, { id: null, name: '', code: '', sort: 0, status: 1, remark: '' })
  if (row) {
    Object.assign(form, { id: row.id, name: row.name, code: row.code, sort: row.sort, status: row.status, remark: row.remark || '' })
    dialogTitle.value = '编辑角色'
  } else { dialogTitle.value = '新增角色' }
  dialogVisible.value = true
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (form.id) { await updateRole(form.id, form) } else { await createRole(form) }
    ElMessage.success('操作成功')
    dialogVisible.value = false
    fetchList()
  } catch (e) { ElMessage.error(e?.message || '操作失败') } finally { submitting.value = false }
}

async function openPermDialog(row) {
  currentRole.value = row
  checkedMenuIds.value = []
  // 先加载菜单树
  const menuRes = await getMenuTree()
  menuTree.value = menuRes?.data?.data || []
  // 加载已分配菜单
  const idsRes = await getRoleMenuIds(row.id)
  const allAssignedIds = idsRes?.data?.data || []
  // 关键修复：el-tree 的 default-checked-keys 只能传叶子节点 ID
  // 否则传入父节点会导致其所有子节点自动全选
  // 找出所有叶子节点ID（没有children或children为空的节点）
  const leafIds = filterLeafIds(menuTree.value, allAssignedIds)
  checkedMenuIds.value = leafIds
  permDialogVisible.value = true
}

// 从菜单树中过滤出叶子节点ID（没有子节点的节点）
function filterLeafIds(tree, assignedIds) {
  const assignedSet = new Set(assignedIds)
  const leafIds = []
  function walk(nodes) {
    for (const node of nodes) {
      const hasChildren = node.children && node.children.length > 0
      if (!hasChildren && assignedSet.has(node.id)) {
        leafIds.push(node.id)
      }
      if (hasChildren) {
        walk(node.children)
      }
    }
  }
  walk(tree)
  return leafIds
}

async function handleAssignMenus() {
  submitting.value = true
  try {
    // 获取所有选中节点（含半选）
    const checked = menuTreeRef.value?.getCheckedKeys() || []
    const halfChecked = menuTreeRef.value?.getHalfCheckedKeys() || []
    const menuIds = [...new Set([...checked, ...halfChecked])]
    await assignRoleMenus(currentRole.value.id, { menuIds })
    ElMessage.success('权限分配成功')
    permDialogVisible.value = false
  } catch { ElMessage.error('分配失败') } finally { submitting.value = false }
}

async function handleDelete(row) {
  await ElMessageBox.confirm(`确认删除角色「${row.name}」？`, '提示', { type: 'warning' })
  try { await deleteRole(row.id); ElMessage.success('删除成功'); fetchList() }
  catch { ElMessage.error('删除失败') }
}
</script>

<style scoped>
.page-container { padding: 16px; display: flex; flex-direction: column; gap: 12px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.title { font-size: 15px; font-weight: 600; }
.pagination { margin-top: 12px; display: flex; justify-content: flex-end; }
.perm-header { margin-bottom: 12px; font-size: 14px; color: #666; }
</style>
