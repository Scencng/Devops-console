<template>
  <div class="page-container">
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="岗位名称">
          <el-input v-model="searchForm.name" placeholder="岗位名称" clearable style="width:180px" />
        </el-form-item>
        <el-form-item label="岗位编码">
          <el-input v-model="searchForm.code" placeholder="岗位编码" clearable style="width:180px" />
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
        <span class="title">岗位列表</span>
        <el-button type="primary" :icon="Plus" @click="openDialog()">新增岗位</el-button>
      </div>
      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column prop="name" label="岗位名称" min-width="150" />
        <el-table-column prop="code" label="岗位编码" min-width="150" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="160" />
        <el-table-column label="创建时间" min-width="160" prop="createdAt">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination class="pagination" v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50]" :total="pagination.total" layout="total, sizes, prev, pager, next"
        @size-change="fetchList" @current-change="fetchList" />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="480px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="岗位名称" prop="name">
          <el-input v-model="form.name" placeholder="岗位名称" />
        </el-form-item>
        <el-form-item label="岗位编码" prop="code">
          <el-input v-model="form.code" placeholder="如：leader, engineer" />
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
  </div>
</template>

<script setup>
import {onMounted, reactive, ref} from 'vue'
import {Plus} from '@element-plus/icons-vue'
import {ElMessage, ElMessageBox} from 'element-plus'
import dayjs from 'dayjs'
import {createPosition, deletePosition, getPositionList, updatePosition} from '@/api/system/position.js'

const loading = ref(false), submitting = ref(false)
const tableData = ref([])
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const searchForm = reactive({ name: '', code: '', status: null })
const dialogVisible = ref(false), dialogTitle = ref('新增岗位')
const formRef = ref(null)
const form = reactive({ id: null, name: '', code: '', sort: 0, status: 1, remark: '' })
const formRules = {
  name: [{ required: true, message: '请输入岗位名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入岗位编码', trigger: 'blur' }]
}

const formatTime = (time) => time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'

onMounted(fetchList)

async function fetchList() {
  loading.value = true
  try {
    const params = { page: pagination.page, pageSize: pagination.pageSize, ...searchForm }
    if (params.status === null) delete params.status
    const res = await getPositionList(params)
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
    dialogTitle.value = '编辑岗位'
  } else { dialogTitle.value = '新增岗位' }
  dialogVisible.value = true
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (form.id) { await updatePosition(form.id, form) } else { await createPosition(form) }
    ElMessage.success('操作成功')
    dialogVisible.value = false
    fetchList()
  } catch (e) { ElMessage.error(e?.message || '操作失败') } finally { submitting.value = false }
}

async function handleDelete(row) {
  await ElMessageBox.confirm(`确认删除岗位「${row.name}」？`, '提示', { type: 'warning' })
  try { await deletePosition(row.id); ElMessage.success('删除成功'); fetchList() }
  catch { ElMessage.error('删除失败') }
}
</script>

<style scoped>
.page-container { padding: 16px; display: flex; flex-direction: column; gap: 12px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.title { font-size: 15px; font-weight: 600; }
.pagination { margin-top: 12px; display: flex; justify-content: flex-end; }
</style>
