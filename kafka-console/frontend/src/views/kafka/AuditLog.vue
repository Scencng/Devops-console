<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <div class="page-eyebrow">Audit Trail</div>
          <h2>审计日志</h2>
          <p>追踪 Kafka 集群、Topic 与消费组相关的高风险操作记录，帮助你还原是谁在什么时候做了什么。</p>
        </div>
      </div>
    </el-card>

    <div class="page-metrics">
      <div class="page-metric-card is-accent">
        <span>日志条数</span>
        <strong>{{ logs.length }}</strong>
        <p>当前筛选条件下返回的审计记录数量。</p>
      </div>
      <div class="page-metric-card is-success">
        <span>成功记录</span>
        <strong>{{ logStats.success }}</strong>
        <p>最近筛选结果中执行成功的操作数量。</p>
      </div>
      <div class="page-metric-card is-warning">
        <span>失败记录</span>
        <strong>{{ logStats.failed }}</strong>
        <p>建议优先查看失败动作的错误原因。</p>
      </div>
      <div class="page-metric-card">
        <span>关联集群</span>
        <strong>{{ logStats.clusterCount }}</strong>
        <p>当前筛选结果覆盖的集群数量。</p>
      </div>
    </div>

    <el-card class="content-card filter-card">
      <div class="toolbar-row">
        <div class="toolbar-left">
          <el-select v-model="filters.clusterId" clearable style="width: 240px">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-input v-model="filters.action" placeholder="如 topic:delete" style="width: 220px" />
          <el-select v-model="filters.result" clearable style="width: 160px">
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
          </el-select>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" @click="loadLogs">查询</el-button>
          <el-button @click="loadLogs" :loading="loading">刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>审计记录</span>
          <span class="card-subtitle">优先查看失败动作、删除类操作和 Offset 干预记录</span>
        </div>
      </template>

      <el-table :data="logs" empty-text="暂无审计日志">
        <el-table-column prop="createdAt" label="时间" width="180">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column prop="operatorUsername" label="操作人" width="140" />
        <el-table-column prop="action" label="动作" width="200" />
        <el-table-column prop="resourceType" label="资源类型" width="120" />
        <el-table-column prop="resourceName" label="资源名称" min-width="200" />
        <el-table-column prop="result" label="结果" width="100">
          <template #default="{ row }">
            <el-tag :type="row.result === 'success' ? 'success' : 'danger'">{{ row.result }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="errorMessage" label="错误信息" min-width="220" show-overflow-tooltip />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getKafkaAuditLogs, getKafkaClusterOptions } from '@/api/kafka.js'

const loading = ref(false)
const clusters = ref([])
const logs = ref([])
const filters = reactive({ clusterId: null, action: '', result: '', page: 1, pageSize: 100 })

const logStats = computed(() => {
  const clusterIds = new Set(logs.value.map((item) => item.clusterId).filter(Boolean))
  return {
    success: logs.value.filter((item) => item.result === 'success').length,
    failed: logs.value.filter((item) => item.result === 'failed').length,
    clusterCount: clusterIds.size,
  }
})

const formatTime = (value) => (value ? new Date(value).toLocaleString() : '-')

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
}

const loadLogs = async () => {
  loading.value = true
  try {
    const res = await getKafkaAuditLogs(filters)
    logs.value = res?.data?.data?.list || []
  } catch (error) {
    ElMessage.error(error.message || '审计日志加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadClusters()
  await loadLogs()
})
</script>
