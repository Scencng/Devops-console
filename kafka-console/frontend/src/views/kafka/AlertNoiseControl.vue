<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>告警降噪</h2>
          <p>通过静默窗口控制告警风暴，配合现有规则去重逻辑降低重复告警噪声</p>
        </div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="loadData">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-button type="primary" @click="openDialog()">新增静默</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <el-table :data="rows" empty-text="暂无静默规则">
        <el-table-column prop="name" label="名称" min-width="180" />
        <el-table-column prop="metricType" label="指标" min-width="160" />
        <el-table-column prop="severity" label="级别" width="120" />
        <el-table-column prop="startsAt" label="开始时间" min-width="180" />
        <el-table-column prop="endsAt" label="结束时间" min-width="180" />
        <el-table-column prop="enabled" label="启用" width="100">
          <template #default="{ row }"><el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑静默' : '新增静默'" width="620px" destroy-on-close>
      <el-form label-position="top" :model="form">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="名称"><el-input v-model="form.name" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="指标"><el-input v-model="form.metricType" placeholder="可为空，表示全部指标" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="级别"><el-input v-model="form.severity" placeholder="可为空，表示全部级别" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="启用"><el-switch v-model="form.enabled" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="开始时间"><el-date-picker v-model="form.startsAt" type="datetime" value-format="x" style="width:100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="结束时间"><el-date-picker v-model="form.endsAt" type="datetime" value-format="x" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="备注"><el-input v-model="form.comment" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createKafkaAlertSilence, deleteKafkaAlertSilence, getKafkaAlertSilences, getKafkaClusterOptions, updateKafkaAlertSilence } from '@/api/kafka.js'

const clusters = ref([])
const selectedClusterId = ref(null)
const loading = ref(false)
const saving = ref(false)
const rows = ref([])
const dialogVisible = ref(false)
const editing = ref(null)
const form = reactive({ name: '', metricType: '', severity: '', startsAt: '', endsAt: '', enabled: true, comment: '' })

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id
}
const loadData = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const res = await getKafkaAlertSilences({ clusterId: selectedClusterId.value })
    rows.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || '静默规则加载失败')
  } finally {
    loading.value = false
  }
}
const openDialog = (row = null) => {
  editing.value = row
  Object.assign(form, row || { name: '', metricType: '', severity: '', startsAt: '', endsAt: '', enabled: true, comment: '' })
  form.startsAt = row?.startsAt ? String(new Date(row.startsAt).getTime()) : ''
  form.endsAt = row?.endsAt ? String(new Date(row.endsAt).getTime()) : ''
  dialogVisible.value = true
}
const save = async () => {
  if (!selectedClusterId.value || !form.name || !form.startsAt || !form.endsAt) {
    ElMessage.warning('请填写名称和时间范围')
    return
  }
  saving.value = true
  try {
    const payload = { ...form, clusterId: selectedClusterId.value }
    if (editing.value?.id) await updateKafkaAlertSilence(editing.value.id, payload)
    else await createKafkaAlertSilence(payload)
    ElMessage.success('静默规则已保存')
    dialogVisible.value = false
    await loadData()
  } catch (error) {
    ElMessage.error(error.message || '静默规则保存失败')
  } finally {
    saving.value = false
  }
}
const remove = async (row) => {
  await ElMessageBox.confirm(`确认删除静默规则 ${row.name} 吗？`, '提示', { type: 'warning' })
  try {
    await deleteKafkaAlertSilence(row.id)
    ElMessage.success('静默规则已删除')
    await loadData()
  } catch (error) {
    ElMessage.error(error.message || '静默规则删除失败')
  }
}

onMounted(async () => {
  await loadClusters()
  await loadData()
})
</script>
