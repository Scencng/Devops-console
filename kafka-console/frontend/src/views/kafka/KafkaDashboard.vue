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

      <div class="page-metrics" v-loading="loading">
        <div class="page-metric-card is-accent">
          <span>当前集群</span>
          <strong>{{ currentClusterName }}</strong>
          <p>本页当前正在观测的 Kafka 集群。</p>
        </div>
        <div class="page-metric-card">
          <span>Broker</span>
          <strong>{{ dashboard.brokerCount }}</strong>
          <p>当前集群中返回的 Broker 节点数量。</p>
        </div>
        <div class="page-metric-card is-success">
          <span>Topic</span>
          <strong>{{ dashboard.topicCount }}</strong>
          <p>当前可见 Topic 总数。</p>
        </div>
        <div class="page-metric-card">
          <span>消费组</span>
          <strong>{{ dashboard.consumerGroupCount }}</strong>
          <p>当前集群下的消费组数量。</p>
        </div>
        <div class="page-metric-card is-warning">
          <span>总分区数</span>
          <strong>{{ dashboard.totalPartitions }}</strong>
          <p>用来快速判断 Topic 承载规模。</p>
        </div>
        <div class="page-metric-card">
          <span>总 Lag</span>
          <strong>{{ dashboard.totalLag }}</strong>
          <p>用于粗略判断是否存在积压风险。</p>
        </div>
      </div>

      <el-card class="content-card" v-loading="loading">
        <template #header>
          <div class="card-header">
            <span>Lag Top 5 消费组</span>
            <span class="card-subtitle">优先查看高积压、状态不稳定的消费组</span>
          </div>
        </template>

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
      </el-card>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getKafkaClusterOptions, getKafkaDashboard } from '@/api/kafka.js'

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

const currentClusterName = computed(
  () => clusters.value.find((item) => item.id === selectedClusterId.value)?.name || '-',
)

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
    const res = await getKafkaDashboard(selectedClusterId.value)
    dashboard.value = res?.data?.data || dashboard.value
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
