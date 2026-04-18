<template>
  <div class="page-container">
    <el-card shadow="never">
      <div class="toolbar">
        <span class="title">部门管理</span>
        <el-button type="primary" :icon="Plus" @click="openDialog()">新增部门</el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" row-key="id" border default-expand-all
        :tree-props="{ children: 'children' }">
        <el-table-column prop="name" label="部门名称" min-width="200" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column prop="leader" label="负责人" width="120" />
        <el-table-column prop="phone" label="联系电话" width="140" />
        <el-table-column prop="email" label="邮箱" min-width="160" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(null, row.id)">新增子级</el-button>
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="520px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="上级部门">
          <el-tree-select v-model="form.parentId" :data="treeData"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="顶级部门" clearable style="width:100%" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="部门名称" prop="name">
              <el-input v-model="form.name" placeholder="部门名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="排序">
              <el-input-number v-model="form.sort" :min="0" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="负责人">
              <el-input v-model="form.leader" placeholder="负责人" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="联系电话">
              <el-input v-model="form.phone" placeholder="电话" />
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
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDeptTree, createDept, updateDept, deleteDept } from '@/api/system/dept.js'

const loading = ref(false)
const submitting = ref(false)
const tableData = ref([])
const treeData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新增部门')
const formRef = ref(null)
const form = reactive({ id: null, parentId: null, name: '', sort: 0, leader: '', phone: '', email: '', status: 1 })
const formRules = { name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }] }

onMounted(fetchList)

async function fetchList() {
  loading.value = true
  try {
    const res = await getDeptTree()
    const data = res?.data?.data || []
    tableData.value = data
    treeData.value = data
  } finally {
    loading.value = false
  }
}

function openDialog(row = null, parentId = null) {
  Object.assign(form, { id: null, parentId: parentId || null, name: '', sort: 0, leader: '', phone: '', email: '', status: 1 })
  if (row) {
    Object.assign(form, { id: row.id, parentId: row.parentId || null, name: row.name, sort: row.sort, leader: row.leader || '', phone: row.phone || '', email: row.email || '', status: row.status })
    dialogTitle.value = '编辑部门'
  } else {
    dialogTitle.value = parentId ? '新增子部门' : '新增部门'
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    const payload = { ...form, parentId: form.parentId || 0 }
    if (payload.id) {
      await updateDept(payload.id, payload)
    } else {
      await createDept(payload)
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

async function handleDelete(row) {
  await ElMessageBox.confirm(`确认删除部门「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await deleteDept(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    ElMessage.error(e?.response?.data?.message || '删除失败')
  }
}
</script>

<style scoped>
.page-container { padding: 16px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.title { font-size: 15px; font-weight: 600; }
</style>
