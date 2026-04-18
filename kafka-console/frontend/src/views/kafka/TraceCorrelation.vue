<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>链路追踪关联</h2>
          <p>维护 Trace 与 Topic / Message Key / Consumer Group 的关联，辅助跨系统消息链路排障</p>
        </div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="loadData">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-input v-model="keyword" style="width:220px" clearable placeholder="搜索 trace/topic/service" @keyup.enter="loadData" />
          <el-button @click="loadData" :loading="loading">刷新</el-button>
          <el-button type="primary" @click="openDialog()">新增关联</el-button>
        </div>
      </div>
    </el-card>
    <el-card class="content-card" v-loading="loading">
      <el-table :data="rows" empty-text="暂无链路关联">
        <el-table-column prop="traceId" label="Trace ID" min-width="220" />
        <el-table-column prop="serviceName" label="服务" min-width="160" />
        <el-table-column prop="topic" label="Topic" min-width="180" />
        <el-table-column prop="messageKey" label="Message Key" min-width="160" />
        <el-table-column prop="consumerGroup" label="Consumer Group" min-width="180" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }"><el-button link type="danger" @click="remove(row)">删除</el-button></template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="dialogVisible" title="新增链路关联" width="680px" destroy-on-close>
      <el-form label-position="top" :model="form">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="Trace ID"><el-input v-model="form.traceId" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="Span ID"><el-input v-model="form.spanId" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="服务"><el-input v-model="form.serviceName" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="Topic"><el-input v-model="form.topic" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="Message Key"><el-input v-model="form.messageKey" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="Consumer Group"><el-input v-model="form.consumerGroup" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="Headers(JSON)"><el-input v-model="form.headers" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createKafkaTraceLink, deleteKafkaTraceLink, getKafkaClusterOptions, getKafkaTraceLinks } from '@/api/kafka.js'

const clusters = ref([])
const selectedClusterId = ref(null)
const keyword = ref('')
const loading = ref(false)
const saving = ref(false)
const rows = ref([])
const dialogVisible = ref(false)
const form = reactive({ traceId: '', spanId: '', serviceName: '', topic: '', messageKey: '', consumerGroup: '', headers: '{}', description: '' })

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id
}
const loadData = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const res = await getKafkaTraceLinks({ clusterId: selectedClusterId.value, keyword: keyword.value })
    rows.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || '链路关联加载失败')
  } finally {
    loading.value = false
  }
}
const openDialog = () => {
  Object.assign(form, { traceId: '', spanId: '', serviceName: '', topic: '', messageKey: '', consumerGroup: '', headers: '{}', description: '' })
  dialogVisible.value = true
}
const save = async () => {
  if (!selectedClusterId.value || !form.traceId) {
    ElMessage.warning('请填写 Trace ID')
    return
  }
  saving.value = true
  try {
    await createKafkaTraceLink({ ...form, clusterId: selectedClusterId.value })
    ElMessage.success('链路关联已保存')
    dialogVisible.value = false
    await loadData()
  } catch (error) {
    ElMessage.error(error.message || '链路关联保存失败')
  } finally {
    saving.value = false
  }
}
const remove = async (row) => {
  await ElMessageBox.confirm(`确认删除 Trace ${row.traceId} 吗？`, '提示', { type: 'warning' })
  try {
    await deleteKafkaTraceLink(row.id)
    ElMessage.success('链路关联已删除')
    await loadData()
  } catch (error) {
    ElMessage.error(error.message || '链路关联删除失败')
  }
}
onMounted(async () => {
  await loadClusters()
  await loadData()
})
</script>
