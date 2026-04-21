<template>
  <div class="page-container">
    <el-empty v-if="!selectedClusterId && !loading" description="请先创建并选择一个 Kafka 集群" />

    <template v-else>
      <el-card class="content-card filter-card">
        <div class="toolbar-row">
          <div class="toolbar-left">
            <el-select
              v-model="selectedClusterId"
              placeholder="选择 Kafka 集群"
              style="width: 300px"
              @change="loadDashboard"
            >
              <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
            </el-select>
          </div>
          <div class="toolbar-right">
            <el-button @click="loadDashboard" :loading="loading">刷新</el-button>
          </div>
        </div>
      </el-card>

      <div class="page-metrics">
        <div v-for="item in overviewMetricCards" :key="item.label" class="page-metric-card" :class="item.variant">
          <template v-if="loading">
            <el-skeleton animated>
              <template #template>
                <div class="metric-skeleton">
                  <el-skeleton-item variant="text" style="width: 42%" />
                  <el-skeleton-item variant="h3" style="width: 58%; margin-top: 14px" />
                  <el-skeleton-item variant="text" style="width: 86%; margin-top: 14px" />
                  <el-skeleton-item variant="text" style="width: 72%; margin-top: 8px" />
                </div>
              </template>
            </el-skeleton>
          </template>
          <template v-else>
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
            <p v-if="item.description">{{ item.description }}</p>
          </template>
        </div>
      </div>

      <el-card class="content-card">
        <template #header>
          <div class="card-header card-header-wrap">
            <span>风险摘要</span>
            <span class="card-subtitle">当前状态</span>
          </div>
        </template>

        <el-skeleton v-if="loading" animated :rows="4" />
        <div v-else class="compact-list">
          <div v-for="item in operationRecommendations" :key="item.title" class="compact-item">
            <div>
              <strong>{{ item.title }}</strong>
              <span>{{ item.description }}</span>
            </div>
            <el-tag :type="item.level === 'high' ? 'danger' : item.level === 'medium' ? 'warning' : 'success'">
              {{ item.level === 'high' ? '优先' : item.level === 'medium' ? '关注' : '正常' }}
            </el-tag>
          </div>
        </div>
      </el-card>

      <el-card class="content-card">
        <template #header>
          <div class="card-header">
            <span>Lag Top 5</span>
            <span class="card-subtitle">关键消费组</span>
          </div>
        </template>

        <el-skeleton v-if="loading" animated :rows="5" />
        <template v-else>
          <el-table :data="dashboard.topLagGroups || []" empty-text="暂无消费组数据">
            <el-table-column prop="groupId" label="消费组" min-width="220" />
          <el-table-column prop="state" label="状态" width="120">
            <template #default="{ row }">
              <el-tag :type="row.state === 'Stable' ? 'success' : 'warning'">
                {{ row.state || 'Unknown' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="memberCount" label="成员数" width="120" />
          <el-table-column prop="partitionCount" label="分区数" width="120" />
          <el-table-column prop="committedLag" label="Lag" width="160" />
          <el-table-column label="Topics" min-width="220">
            <template #default="{ row }">{{ (row.topics || []).join(', ') || '-' }}</template>
          </el-table-column>
          </el-table>
        </template>
        </el-card>

    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getKafkaAuditLogs, getKafkaBrokers, getKafkaClusterOptions, getKafkaConsumerGroups, getKafkaDashboard, getKafkaTopics } from '@/api/kafka.js'

const loading = ref(false)
const clusters = ref([])
const selectedClusterId = ref(null)
const dashboard = ref({
  brokerCount: 0,
  topicCount: 0,
  consumerGroupCount: 0,
  totalPartitions: 0,
  totalLag: 0,
  topLagGroups: [],
})
const brokers = ref([])
const topics = ref([])
const consumerGroups = ref([])
const recentAuditLogs = ref([])

const currentClusterName = computed(
  () => clusters.value.find((item) => item.id === selectedClusterId.value)?.name || '-',
)

const riskSummary = computed(() => ({
  unstableGroups: consumerGroups.value.filter((item) => item.state !== 'Stable').length,
  connectedBrokers: brokers.value.filter((item) => item.connected).length,
  internalTopics: topics.value.filter((item) => item.internal).length,
  highLagGroups: consumerGroups.value.filter((item) => Number(item.committedLag || 0) > 0).length,
}))

const operationRecommendations = computed(() => {
  const recommendations = []
  const disconnected = dashboard.value.brokerCount - riskSummary.value.connectedBrokers
  const failedAuditCount = recentAuditLogs.value.filter((item) => item.result === 'failed').length

  if (disconnected > 0) {
    recommendations.push({
      title: 'Broker 连通性',
      description: `未连接 ${disconnected} 个`,
      level: 'high',
    })
  }

  if (riskSummary.value.unstableGroups > 0) {
    recommendations.push({
      title: '消费组状态',
      description: `异常 ${riskSummary.value.unstableGroups} 个`,
      level: 'high',
    })
  }

  if (dashboard.value.totalLag > 0) {
    recommendations.push({
      title: '消费积压',
      description: `总 Lag ${dashboard.value.totalLag}`,
      level: 'medium',
    })
  }

  if (riskSummary.value.internalTopics > 0) {
    recommendations.push({
      title: '内部 Topic',
      description: `${riskSummary.value.internalTopics} 个`,
      level: 'medium',
    })
  }

  if (failedAuditCount > 0) {
    recommendations.push({
      title: '失败操作',
      description: `${failedAuditCount} 条`,
      level: 'medium',
    })
  }

  if (recommendations.length === 0) {
    recommendations.push({
      title: '运行状态',
      description: '暂无明显风险',
      level: 'low',
    })
  }

  return recommendations.slice(0, 5)
})

const overviewMetricCards = computed(() => [
  {
    label: 'Broker',
    value: dashboard.value.brokerCount,
    description: '',
    variant: 'is-accent',
  },
  {
    label: 'Topic',
    value: dashboard.value.topicCount,
    description: '',
    variant: '',
  },
  {
    label: '消费组',
    value: dashboard.value.consumerGroupCount,
    description: '',
    variant: 'is-success',
  },
  {
    label: '总分区',
    value: dashboard.value.totalPartitions,
    description: '',
    variant: '',
  },
  {
    label: '总 Lag',
    value: dashboard.value.totalLag,
    description: '',
    variant: 'is-warning',
  },
])

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) {
    selectedClusterId.value = clusters.value[0].id
  }
}

const loadDashboard = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const [dashboardRes, brokersRes, topicsRes, groupsRes] = await Promise.all([
      getKafkaDashboard(selectedClusterId.value),
      getKafkaBrokers(selectedClusterId.value),
      getKafkaTopics({ clusterId: selectedClusterId.value }),
      getKafkaConsumerGroups({ clusterId: selectedClusterId.value }),
    ])
    dashboard.value = dashboardRes?.data?.data || dashboard.value
    brokers.value = brokersRes?.data?.data || []
    topics.value = topicsRes?.data?.data || []
    consumerGroups.value = groupsRes?.data?.data || []
    const auditRes = await getKafkaAuditLogs({
      clusterId: selectedClusterId.value,
      page: 1,
      pageSize: 8,
    })
    recentAuditLogs.value = auditRes?.data?.data?.list || []
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 总览加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    await loadClusters()
    await loadDashboard()
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 集群列表加载失败')
  }
})
</script>

<style scoped>
.metric-skeleton {
  width: 100%;
}
</style>
