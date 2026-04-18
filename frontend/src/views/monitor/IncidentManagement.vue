<template>
  <div class="incident-management">
    <!-- 统计卡片行 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon-wrap" style="background:#5c6bc0;">
          <el-icon size="22" color="#fff"><Document /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">故障总数</div>
          <div class="stat-value">{{ incidentStats.total }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon-wrap" style="background:#e53935;">
          <el-icon size="22" color="#fff"><Warning /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">未处理</div>
          <div class="stat-value" style="color:#e53935">{{ incidentStats.pending }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon-wrap" style="background:#f57c00;">
          <el-icon size="22" color="#fff"><Loading /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">处理中</div>
          <div class="stat-value" style="color:#f57c00">{{ incidentStats.processing }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon-wrap" style="background:#43a047;">
          <el-icon size="22" color="#fff"><CircleCheck /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">已完成</div>
          <div class="stat-value" style="color:#43a047">{{ incidentStats.done }}</div>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <div class="content-card">
      <div class="section-title">故障管理</div>

      <!-- 搜索栏 -->
      <div class="bar">
        <el-select v-model="searchForm.businessLine" placeholder="业务线" clearable style="width:130px">
          <el-option v-for="b in businessLines" :key="b" :label="b" :value="b" />
        </el-select>
        <el-select v-model="searchForm.level" placeholder="故障级别" clearable style="width:120px">
          <el-option v-for="l in LEVELS" :key="l.value" :label="l.label" :value="l.value" />
        </el-select>
        <el-select v-model="searchForm.status" placeholder="处理状态" clearable style="width:120px">
          <el-option label="全部" value="" />
          <el-option label="未处理" value="pending" />
          <el-option label="处理中" value="processing" />
          <el-option label="已完成" value="done" />
        </el-select>
        <el-select v-model="searchForm.dept" placeholder="所属部门" clearable style="width:120px">
          <el-option v-for="d in DEPTS" :key="d" :label="d" :value="d" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="fetchIncidents">搜索</el-button>
        <el-button :icon="Refresh" @click="resetSearch">重置</el-button>
      </div>

      <!-- 操作按钮 -->
      <div class="bar">
        <el-button type="primary" :icon="Plus" @click="openDialog()">添加故障记录</el-button>
      </div>

      <!-- 故障表格 -->
      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="alertTime" label="告警时间" min-width="155" show-overflow-tooltip />
        <el-table-column prop="businessLine" label="业务线" min-width="100" show-overflow-tooltip />
        <el-table-column label="故障级别" width="90">
          <template #default="{ row }">
            <el-tag :type="getLevelTagType(row.level)" size="small" effect="dark">
              {{ row.level }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="频率" width="70">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.frequency || '偶发' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="alertDesc" label="告警描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="故障详情" width="80">
          <template #default="{ row }">
            <el-tooltip v-if="row.detail" :content="row.detail" placement="top">
              <el-button link type="primary" size="small">查看</el-button>
            </el-tooltip>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="dept" label="所属部门" min-width="90" show-overflow-tooltip />
        <el-table-column prop="handler" label="处理人" width="90" show-overflow-tooltip />
        <el-table-column label="处理状态" width="80">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)" size="small">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-tooltip content="开始处理" placement="top">
              <el-button
                link :icon="VideoPlay"
                type="warning" size="small"
                :disabled="row.status !== 'pending'"
                @click="handleStatusChange(row, 'processing')"
              />
            </el-tooltip>
            <el-tooltip content="标记完成" placement="top">
              <el-button
                link :icon="CircleCheckFilled"
                type="success" size="small"
                :disabled="row.status === 'done'"
                @click="handleStatusChange(row, 'done')"
              />
            </el-tooltip>
            <el-tooltip content="编辑" placement="top">
              <el-button link :icon="Edit" type="primary" size="small" @click="openDialog(row)" />
            </el-tooltip>
            <el-tooltip content="删除" placement="top">
              <el-button link :icon="Delete" type="danger" size="small" @click="handleDelete(row)" />
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        class="pagination"
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchIncidents"
        @current-change="fetchIncidents"
      />
    </div>

    <!-- ==================== 新增/编辑故障 Dialog ==================== -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="560px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="告警时间" prop="alertTime">
              <el-date-picker
                v-model="form.alertTime"
                type="datetime"
                placeholder="选择时间"
                format="YYYY/MM/DD HH:mm:ss"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width:100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="业务线" prop="businessLine">
              <el-select v-model="form.businessLine" allow-create filterable placeholder="业务线" style="width:100%">
                <el-option v-for="b in businessLines" :key="b" :label="b" :value="b" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="故障级别" prop="level">
              <el-select v-model="form.level" placeholder="故障级别" style="width:100%">
                <el-option v-for="l in LEVELS" :key="l.value" :label="l.label" :value="l.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="频率" prop="frequency">
              <el-select v-model="form.frequency" placeholder="频率" style="width:100%">
                <el-option label="偶发" value="偶发" />
                <el-option label="频繁" value="频繁" />
                <el-option label="持续" value="持续" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所属部门" prop="dept">
              <el-select v-model="form.dept" allow-create filterable placeholder="所属部门" style="width:100%">
                <el-option v-for="d in DEPTS" :key="d" :label="d" :value="d" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="处理人" prop="handler">
              <el-input v-model="form.handler" placeholder="处理人姓名" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="告警描述" prop="alertDesc">
              <el-input v-model="form.alertDesc" type="textarea" :rows="2" placeholder="告警描述" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="故障详情">
              <el-input v-model="form.detail" type="textarea" :rows="4" placeholder="故障详情、影响范围、处理过程..." />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="处理状态">
              <el-select v-model="form.status" style="width:100%">
                <el-option label="未处理" value="pending" />
                <el-option label="处理中" value="processing" />
                <el-option label="已完成" value="done" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="解决时间">
              <el-date-picker
                v-model="form.resolvedAt"
                type="datetime"
                placeholder="解决时间"
                format="YYYY/MM/DD HH:mm:ss"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width:100%"
              />
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
import {
  Document, Warning, CircleCheck, Search, Refresh, Plus,
  Edit, Delete, VideoPlay, CircleCheckFilled, Loading
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getIncidentStats, getIncidentList, createIncident,
  updateIncident, deleteIncident, updateIncidentStatus, getBusinessLines
} from '@/api/incident.js'

// ===================== 常量 =====================
const LEVELS = [
  { value: 'P1', label: 'P1 - 严重' },
  { value: 'P2', label: 'P2 - 高' },
  { value: 'P3', label: 'P3 - 中' },
  { value: 'P4', label: 'P4 - 低' },
]
const DEPTS = ['运维部', '研发部', '产品部', '测试部', '业务部']

// ===================== 状态 =====================
const loading = ref(false)
const submitting = ref(false)
const tableData = ref([])
const businessLines = ref([])
const incidentStats = reactive({ total: 0, pending: 0, processing: 0, done: 0 })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const searchForm = reactive({ businessLine: '', level: '', status: '', dept: '' })

const dialogVisible = ref(false)
const dialogTitle = ref('添加故障记录')
const formRef = ref(null)
const form = reactive({
  id: null, alertTime: '', businessLine: '', level: 'P4', frequency: '偶发',
  alertDesc: '', detail: '', dept: '', handler: 'admin', status: 'pending', resolvedAt: ''
})
const formRules = {
  alertTime: [{ required: true, message: '请选择告警时间', trigger: 'change' }],
  businessLine: [{ required: true, message: '请选择业务线', trigger: 'change' }],
  level: [{ required: true, message: '请选择故障级别' }],
  alertDesc: [{ required: true, message: '请填写告警描述', trigger: 'blur' }],
  dept: [{ required: true, message: '请选择所属部门' }],
}

// ===================== 生命周期 =====================
onMounted(() => {
  fetchStats()
  fetchIncidents()
  fetchBusinessLines()
})

// ===================== 数据获取 =====================
async function fetchStats() {
  try {
    const res = await getIncidentStats()
    Object.assign(incidentStats, res?.data?.data || {})
  } catch {}
}

async function fetchIncidents() {
  loading.value = true
  try {
    const params = { page: pagination.page, pageSize: pagination.pageSize, ...searchForm }
    Object.keys(params).forEach(k => !params[k] && delete params[k])
    const res = await getIncidentList(params)
    const d = res?.data?.data || {}
    tableData.value = d.list || []
    pagination.total = d.total || 0
  } catch { tableData.value = [] } finally { loading.value = false }
}

async function fetchBusinessLines() {
  try {
    const res = await getBusinessLines()
    businessLines.value = res?.data?.data || []
  } catch { businessLines.value = ['运维部门', '研发团队', '信息中心'] }
}

// ===================== 操作方法 =====================
function resetSearch() {
  Object.assign(searchForm, { businessLine: '', level: '', status: '', dept: '' })
  pagination.page = 1
  fetchIncidents()
}

function openDialog(row = null) {
  Object.assign(form, {
    id: null, alertTime: '', businessLine: '', level: 'P4', frequency: '偶发',
    alertDesc: '', detail: '', dept: '', handler: 'admin', status: 'pending', resolvedAt: ''
  })
  if (row) {
    Object.assign(form, {
      id: row.id, alertTime: row.alertTime, businessLine: row.businessLine,
      level: row.level, frequency: row.frequency || '偶发', alertDesc: row.alertDesc,
      detail: row.detail || '', dept: row.dept, handler: row.handler, status: row.status, resolvedAt: row.resolvedAt || ''
    })
    dialogTitle.value = '编辑故障记录'
  } else { dialogTitle.value = '添加故障记录' }
  dialogVisible.value = true
}

async function handleSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (form.id) { await updateIncident(form.id, form) } else { await createIncident(form) }
    ElMessage.success('操作成功'); dialogVisible.value = false; fetchIncidents(); fetchStats()
  } catch (e) { ElMessage.error(e?.message || '操作失败') } finally { submitting.value = false }
}

async function handleStatusChange(row, newStatus) {
  try {
    await updateIncidentStatus(row.id, { status: newStatus })
    row.status = newStatus
    ElMessage.success(`状态已更新为：${getStatusLabel(newStatus)}`)
    fetchStats()
  } catch (e) { ElMessage.error(e?.message || '操作失败') }
}

async function handleDelete(row) {
  await ElMessageBox.confirm(`确认删除故障记录「${row.alertDesc}」？`, '提示', { type: 'warning' })
  try { await deleteIncident(row.id); ElMessage.success('删除成功'); fetchIncidents(); fetchStats() }
  catch (e) { ElMessage.error(e?.message || '删除失败') }
}

// ===================== 工具函数 =====================
const getLevelTagType = (level) => {
  const types = { P1: 'danger', P2: 'warning', P3: '', P4: 'info' }
  return types[level] || 'info'
}
const getStatusLabel = (s) => ({ pending: '未处理', processing: '处理中', done: '已完成' }[s] || s)
const getStatusTagType = (s) => ({ pending: 'danger', processing: 'warning', done: 'success' }[s] || 'info')
</script>

<style scoped>
.incident-management {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 统计卡片 */
.stats-row {
  display: flex;
  gap: 12px;
}

.stat-card {
  flex: 1;
  background: #fff;
  border-radius: 10px;
  border: 1px solid #e4e7ed;
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 14px;
  transition: box-shadow 0.2s;
}

.stat-card:hover {
  box-shadow: 0 4px 16px rgba(0,0,0,0.08);
}

.stat-icon-wrap {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  line-height: 1;
}

/* 内容卡片 */
.content-card {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #e4e7ed;
  padding: 20px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
}

.bar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.pagination {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
</style>
