<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>变更审批</h2>
          <p>对高风险 Kafka 变更进行提交、审批和执行，适用于删 Topic、扩分区、重置 Offset 等操作</p>
        </div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="loadRequests">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-button type="primary" @click="openDialog()">新建变更单</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <el-table :data="requests" empty-text="暂无变更单">
        <el-table-column prop="changeType" label="类型" min-width="180" />
        <el-table-column prop="resourceType" label="资源类型" width="120" />
        <el-table-column prop="resourceName" label="资源名" min-width="180" />
        <el-table-column prop="status" label="状态" width="120" />
        <el-table-column prop="requesterUsername" label="申请人" width="140" />
        <el-table-column prop="approverUsername" label="审批人" width="140" />
        <el-table-column prop="createdAt" label="创建时间" min-width="180" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewPayload(row)">查看</el-button>
            <el-button link type="warning" :disabled="row.status !== 'pending'" @click="review(row, 'approved')">通过</el-button>
            <el-button link type="danger" :disabled="row.status !== 'pending'" @click="review(row, 'rejected')">拒绝</el-button>
            <el-button link type="success" :disabled="row.status !== 'approved'" @click="execute(row)">执行</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" title="新建变更单" width="700px" destroy-on-close>
      <el-form label-position="top" :model="form">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="变更类型"><el-select v-model="form.changeType" style="width:100%"><el-option label="扩容分区" value="topic.partitions.increase" /><el-option label="巡检任务" value="inspection.run" /><el-option label="告警评估" value="alerts.evaluate" /><el-option label="重置 Offset" value="group.offset.reset" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="资源类型"><el-input v-model="form.resourceType" placeholder="例如 topic / consumer_group" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="资源名"><el-input v-model="form.resourceName" /></el-form-item>
        <el-form-item label="变更原因"><el-input v-model="form.reason" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="Payload(JSON)"><el-input v-model="form.payload" type="textarea" :rows="10" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">提交审批</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="payloadVisible" title="变更详情" width="760px" destroy-on-close>
      <pre class="payload-pre">{{ activePayload }}</pre>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createKafkaChangeRequest, executeKafkaChangeRequest, getKafkaChangeRequests, getKafkaClusterOptions, reviewKafkaChangeRequest } from '@/api/kafka.js'

const clusters = ref([])
const selectedClusterId = ref(null)
const loading = ref(false)
const saving = ref(false)
const requests = ref([])
const dialogVisible = ref(false)
const payloadVisible = ref(false)
const activePayload = ref('')
const form = reactive({ changeType: 'topic.partitions.increase', resourceType: 'topic', resourceName: '', reason: '', payload: '{}' })

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id
}

const loadRequests = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const res = await getKafkaChangeRequests({ clusterId: selectedClusterId.value })
    requests.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || '变更单加载失败')
  } finally {
    loading.value = false
  }
}

const openDialog = () => {
  Object.assign(form, { changeType: 'topic.partitions.increase', resourceType: 'topic', resourceName: '', reason: '', payload: '{}' })
  dialogVisible.value = true
}

const save = async () => {
  if (!selectedClusterId.value || !form.resourceName) {
    ElMessage.warning('请填写资源名')
    return
  }
  saving.value = true
  try {
    await createKafkaChangeRequest({ ...form, clusterId: selectedClusterId.value })
    ElMessage.success('变更单已提交')
    dialogVisible.value = false
    await loadRequests()
  } catch (error) {
    ElMessage.error(error.message || '变更单创建失败')
  } finally {
    saving.value = false
  }
}

const review = async (row, status) => {
  try {
    await reviewKafkaChangeRequest(row.id, { status, comment: status === 'approved' ? '审批通过' : '审批拒绝' })
    ElMessage.success('审批完成')
    await loadRequests()
  } catch (error) {
    ElMessage.error(error.message || '审批失败')
  }
}

const execute = async (row) => {
  try {
    await executeKafkaChangeRequest(row.id)
    ElMessage.success('变更已执行')
    await loadRequests()
  } catch (error) {
    ElMessage.error(error.message || '变更执行失败')
  }
}

const viewPayload = (row) => {
  activePayload.value = row.payload || '{}'
  payloadVisible.value = true
}

onMounted(async () => {
  try {
    await loadClusters()
    await loadRequests()
  } catch (error) {
    ElMessage.error(error.message || '变更审批初始化失败')
  }
})
</script>

<style scoped>
.payload-pre {
  padding: 12px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  background: #0f172a;
  border-radius: 10px;
  color: #e2e8f0;
}
</style>
