<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>一键巡检</h2>
          <p>对 Kafka 集群执行标准化巡检，输出 Broker、复制、积压和告警维度的检查报告</p>
        </div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="loadReports">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-button type="primary" :loading="running" @click="runInspection">立即巡检</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <el-table :data="reports" empty-text="暂无巡检报告">
        <el-table-column prop="name" label="报告名称" min-width="220" />
        <el-table-column prop="status" label="状态" width="120" />
        <el-table-column prop="issueCount" label="问题数" width="120" />
        <el-table-column prop="triggeredBy" label="触发人" width="140" />
        <el-table-column prop="executedAt" label="执行时间" min-width="180" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-drawer v-model="detailVisible" :title="`巡检详情: ${activeReport?.name || ''}`" size="60%">
      <el-skeleton :loading="detailLoading" animated :rows="8">
        <template #default>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="状态">{{ activeReport?.status || '-' }}</el-descriptions-item>
            <el-descriptions-item label="问题数">{{ activeReport?.issueCount || 0 }}</el-descriptions-item>
            <el-descriptions-item label="执行时间">{{ activeReport?.executedAt || '-' }}</el-descriptions-item>
            <el-descriptions-item label="触发人">{{ activeReport?.triggeredBy || '-' }}</el-descriptions-item>
          </el-descriptions>
          <div class="summary">{{ activeReport?.summary }}</div>
          <el-table :data="activeReport?.items || []" empty-text="暂无检查项">
            <el-table-column prop="checkCode" label="检查项" width="180" />
            <el-table-column prop="severity" label="级别" width="120" />
            <el-table-column prop="status" label="状态" width="120" />
            <el-table-column prop="title" label="标题" min-width="220" />
            <el-table-column prop="detail" label="详情" min-width="320" show-overflow-tooltip />
          </el-table>
        </template>
      </el-skeleton>
    </el-drawer>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getKafkaClusterOptions, getKafkaInspectionReport, getKafkaInspectionReports, runKafkaInspection } from '@/api/kafka.js'

const clusters = ref([])
const selectedClusterId = ref(null)
const loading = ref(false)
const running = ref(false)
const reports = ref([])
const detailVisible = ref(false)
const detailLoading = ref(false)
const activeReport = ref(null)

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id
}

const loadReports = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const res = await getKafkaInspectionReports({ clusterId: selectedClusterId.value })
    reports.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || '巡检报告加载失败')
  } finally {
    loading.value = false
  }
}

const runInspection = async () => {
  if (!selectedClusterId.value) return
  running.value = true
  try {
    await runKafkaInspection({ clusterId: selectedClusterId.value })
    ElMessage.success('巡检完成')
    await loadReports()
  } catch (error) {
    ElMessage.error(error.message || '巡检失败')
  } finally {
    running.value = false
  }
}

const openDetail = async (row) => {
  detailVisible.value = true
  detailLoading.value = true
  try {
    const res = await getKafkaInspectionReport(row.id)
    activeReport.value = res?.data?.data || null
  } catch (error) {
    ElMessage.error(error.message || '巡检详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

onMounted(async () => {
  try {
    await loadClusters()
    await loadReports()
  } catch (error) {
    ElMessage.error(error.message || '巡检中心初始化失败')
  }
})
</script>

<style scoped>
.summary { margin: 16px 0; color:#606266; }
</style>
