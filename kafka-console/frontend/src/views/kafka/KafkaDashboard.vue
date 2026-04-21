<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <div class="page-eyebrow">Overview</div>
          <h2>Kafka 总览</h2>
          <p>查看当前集群的 Broker、Topic、消费组与 Lag 概况，先判断哪里需要优先处理。</p>
        </div>
      </div>
    </el-card>

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
            <p>{{ item.description }}</p>
          </template>
        </div>
      </div>

      <el-card class="content-card">
        <template #header>
          <div class="card-header card-header-wrap">
            <span>平台运行建议卡片</span>
            <span class="card-subtitle">根据当前集群结构、消费状态和最近变更，给出优先处理建议</span>
          </div>
        </template>

        <div v-if="loading" class="workbench-grid">
          <div class="workspace-panel">
            <el-skeleton animated :rows="5" />
          </div>
          <div class="workspace-panel">
            <el-skeleton animated :rows="5" />
          </div>
        </div>

        <div v-else class="workbench-grid">
          <div class="workspace-panel">
            <h3>建议优先动作</h3>
            <p>按风险高低排序，建议先处理前面的事项。</p>
            <div class="compact-list">
              <div v-for="item in operationRecommendations" :key="item.title" class="compact-item">
                <div>
                  <strong>{{ item.title }}</strong>
                  <span>{{ item.description }}</span>
                </div>
                <el-tag :type="item.level === 'high' ? 'danger' : item.level === 'medium' ? 'warning' : 'success'">
                  {{ item.level === 'high' ? '优先处理' : item.level === 'medium' ? '建议关注' : '状态良好' }}
                </el-tag>
              </div>
            </div>
          </div>

          <div class="workspace-panel">
            <h3>当前运行态势</h3>
            <p>把核心运行风险收敛成几句话，便于快速判断当前平台是否处于稳定状态。</p>
            <div class="compact-list">
              <div class="compact-item">
                <div>
                  <strong>Broker 连通性</strong>
                  <span>
                    {{
                      riskSummary.connectedBrokers === dashboard.brokerCount
                        ? '当前所有 Broker 都处于连接状态。'
                        : `当前有 ${dashboard.brokerCount - riskSummary.connectedBrokers} 个 Broker 未连接，建议优先排查节点和网络。`
                    }}
                  </span>
                </div>
              </div>
              <div class="compact-item">
                <div>
                  <strong>消费组稳定性</strong>
                  <span>
                    {{
                      riskSummary.unstableGroups === 0
                        ? '当前消费组状态整体稳定。'
                        : `当前有 ${riskSummary.unstableGroups} 个消费组状态异常，可能存在重平衡或消费停滞。`
                    }}
                  </span>
                </div>
              </div>
              <div class="compact-item">
                <div>
                  <strong>最近变更活跃度</strong>
                  <span>
                    {{
                      recentAuditLogs.length === 0
                        ? '最近未检测到显著变更记录。'
                        : `最近记录到 ${recentAuditLogs.length} 条审计操作，建议关注失败或删除类动作。`
                    }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-card>

      <div class="dashboard-grid">
        <el-card class="content-card">
          <template #header>
            <div class="card-header">
              <span>集群结构分布</span>
              <span class="card-subtitle">把 Broker、Topic、消费组和分区规模放在一张图里快速看结构</span>
            </div>
          </template>
          <template v-if="loading">
            <el-skeleton animated>
              <template #template>
                <div class="chart-skeleton">
                  <el-skeleton-item variant="image" style="width: 100%; height: 320px; border-radius: 20px" />
                </div>
              </template>
            </el-skeleton>
          </template>
          <v-chart v-else class="dashboard-chart" :option="overviewChartOption" autoresize />
        </el-card>

        <el-card class="content-card">
          <template #header>
            <div class="card-header">
              <span>消费组状态</span>
              <span class="card-subtitle">优先识别 Stable 之外的状态分布</span>
            </div>
          </template>
          <template v-if="loading">
            <el-skeleton animated>
              <template #template>
                <div class="chart-skeleton">
                  <el-skeleton-item variant="image" style="width: 100%; height: 320px; border-radius: 20px" />
                </div>
              </template>
            </el-skeleton>
          </template>
          <v-chart v-else class="dashboard-chart" :option="consumerStateChartOption" autoresize />
        </el-card>
      </div>

      <el-card class="content-card">
        <template #header>
          <div class="card-header">
            <span>Lag Top 5 消费组</span>
            <span class="card-subtitle">优先查看高积压、状态不稳定的消费组</span>
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

      <el-card class="content-card">
        <template #header>
          <div class="card-header">
            <span>最近审计操作</span>
            <span class="card-subtitle">快速查看当前集群最近发生了哪些变更、谁执行了这些操作以及是否失败</span>
          </div>
        </template>

        <el-skeleton v-if="loading" animated :rows="5" />
        <template v-else>
        <el-table :data="recentAuditLogs" empty-text="暂无审计记录">
          <el-table-column prop="createdAt" label="时间" width="180">
            <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column prop="operatorUsername" label="操作人" width="140" />
          <el-table-column prop="action" label="动作" width="200" />
          <el-table-column prop="resourceType" label="资源类型" width="120" />
          <el-table-column prop="resourceName" label="资源名称" min-width="180" />
          <el-table-column prop="result" label="结果" width="100">
            <template #default="{ row }">
              <el-tag :type="row.result === 'success' ? 'success' : 'danger'">
                {{ row.result === 'success' ? '成功' : '失败' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="errorMessage" label="错误信息" min-width="240" show-overflow-tooltip />
        </el-table>
        </template>
      </el-card>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, PieChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { getKafkaAuditLogs, getKafkaBrokers, getKafkaClusterOptions, getKafkaConsumerGroups, getKafkaDashboard, getKafkaTopics } from '@/api/kafka.js'

use([CanvasRenderer, BarChart, PieChart, GridComponent, TooltipComponent, LegendComponent])

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

  if (dashboard.value.brokerCount > riskSummary.value.connectedBrokers) {
    recommendations.push({
      title: '优先检查 Broker 连通性',
      description: `当前有 ${dashboard.value.brokerCount - riskSummary.value.connectedBrokers} 个 Broker 未连接，建议先确认节点状态、网络和端口可达性。`,
      level: 'high',
    })
  }

  if (riskSummary.value.unstableGroups > 0) {
    recommendations.push({
      title: '处理异常消费组',
      description: `当前有 ${riskSummary.value.unstableGroups} 个消费组状态不是 Stable，建议优先查看高 Lag 与异常状态消费组。`,
      level: 'high',
    })
  }

  if (dashboard.value.totalLag > 0) {
    recommendations.push({
      title: '关注消费积压',
      description: `当前累计 Lag 为 ${dashboard.value.totalLag}，建议先处理 Lag Top 5 列表中的消费组。`,
      level: 'medium',
    })
  }

  if (riskSummary.value.internalTopics > 0) {
    recommendations.push({
      title: '谨慎处理内部 Topic',
      description: `当前共有 ${riskSummary.value.internalTopics} 个内部 Topic，涉及删除或配置调整时建议先确认用途。`,
      level: 'medium',
    })
  }

  const failedAuditCount = recentAuditLogs.value.filter((item) => item.result === 'failed').length
  if (failedAuditCount > 0) {
    recommendations.push({
      title: '复盘失败操作',
      description: `最近审计记录中有 ${failedAuditCount} 条失败动作，建议优先查看错误信息并确认是否需要回滚或重试。`,
      level: 'medium',
    })
  }

  if (recommendations.length === 0) {
    recommendations.push({
      title: '当前运行态势稳定',
      description: '暂未检测到明显的 Broker、消费组或审计风险信号，可以继续按日常节奏巡检。',
      level: 'low',
    })
  }

  return recommendations.slice(0, 5)
})

const overviewMetricCards = computed(() => [
  {
    label: '当前集群',
    value: currentClusterName.value,
    description: '本页当前正在观测的 Kafka 集群。',
    variant: 'is-accent',
  },
  {
    label: 'Broker',
    value: dashboard.value.brokerCount,
    description: '当前集群中返回的 Broker 节点数量。',
    variant: '',
  },
  {
    label: 'Topic',
    value: dashboard.value.topicCount,
    description: '当前可见 Topic 总数。',
    variant: 'is-success',
  },
  {
    label: '消费组',
    value: dashboard.value.consumerGroupCount,
    description: '当前集群下的消费组数量。',
    variant: '',
  },
  {
    label: '总分区数',
    value: dashboard.value.totalPartitions,
    description: '用来快速判断 Topic 承载规模。',
    variant: 'is-warning',
  },
  {
    label: '总 Lag',
    value: dashboard.value.totalLag,
    description: '用于粗略判断是否存在积压风险。',
    variant: '',
  },
  {
    label: '异常消费组',
    value: riskSummary.value.unstableGroups,
    description: '状态不是 Stable 的消费组，建议优先检查是否存在重平衡或消费异常。',
    variant: 'is-warning',
  },
  {
    label: '连接正常 Broker',
    value: riskSummary.value.connectedBrokers,
    description: '用于判断当前集群 Broker 连通性是否完整。',
    variant: '',
  },
  {
    label: '内部 Topic',
    value: riskSummary.value.internalTopics,
    description: '内部 Topic 不建议进行删除类操作，可作为治理时的保护边界。',
    variant: 'is-success',
  },
  {
    label: '高 Lag 消费组',
    value: riskSummary.value.highLagGroups,
    description: 'Lag 大于 0 的消费组数量，可快速判断积压是否已经扩散。',
    variant: '',
  },
])

const formatTime = (value) => (value ? new Date(value).toLocaleString() : '-')

const overviewChartOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  grid: { left: 16, right: 16, top: 32, bottom: 8, containLabel: true },
  xAxis: {
    type: 'category',
    data: ['Broker', 'Topic', '消费组', '分区'],
    axisTick: { show: false },
  },
  yAxis: {
    type: 'value',
    splitLine: { lineStyle: { color: 'rgba(148, 163, 184, 0.16)' } },
  },
  series: [
    {
      type: 'bar',
      barWidth: 34,
      itemStyle: {
        borderRadius: [10, 10, 0, 0],
        color: '#2f7df6',
      },
      data: [
        dashboard.value.brokerCount,
        dashboard.value.topicCount,
        dashboard.value.consumerGroupCount,
        dashboard.value.totalPartitions,
      ],
    },
  ],
}))

const consumerStateChartOption = computed(() => {
  const stateCountMap = consumerGroups.value.reduce((acc, item) => {
    const state = item.state || 'Unknown'
    acc[state] = (acc[state] || 0) + 1
    return acc
  }, {})
  const seriesData = Object.entries(stateCountMap).map(([name, value]) => ({ name, value }))

  return {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0 },
    series: [
      {
        type: 'pie',
        radius: ['48%', '72%'],
        center: ['50%', '44%'],
        label: { formatter: '{b}\n{c}' },
        data: seriesData.length > 0 ? seriesData : [{ name: '暂无数据', value: 1, itemStyle: { color: '#cbd5e1' } }],
      },
    ],
  }
})

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
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 22px;
  margin-bottom: 22px;
}

.dashboard-chart {
  width: 100%;
  height: 320px;
}

.metric-skeleton,
.chart-skeleton {
  width: 100%;
}

@media (max-width: 960px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }
}
</style>
