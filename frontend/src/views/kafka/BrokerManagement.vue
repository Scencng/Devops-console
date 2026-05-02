<template>
  <div class="page-container">
    <el-card class="page-header-card" shadow="never">
      <div class="page-header">
        <div class="page-header-copy">
          <div class="page-eyebrow">Kafka</div>
          <h2>Broker 管理</h2>
          <p>聚焦节点连通性、Controller 角色和分区承载，让热点 Broker 更快暴露出来。</p>
        </div>

        <div class="page-header-side">
          <div class="page-header-meta">
            <div class="page-header-kpi">
              <span>当前集群</span>
              <strong>{{ currentClusterName }}</strong>
            </div>
            <div class="page-header-kpi">
              <span>关注节点</span>
              <strong>{{ hotspotBrokers.length }}</strong>
            </div>
          </div>
          <div class="page-header-note">
            {{ brokerRiskSummary }}
          </div>
          <div class="page-header-actions">
            <el-button @click="loadBrokers" :loading="loading">刷新</el-button>
          </div>
        </div>
      </div>
    </el-card>

    <div class="page-metrics">
      <div class="page-metric-card">
        <span>Broker 节点</span>
        <strong>{{ brokerStats.total }}</strong>
      </div>
      <div class="page-metric-card is-success">
        <span>已连接</span>
        <strong>{{ brokerStats.connected }}</strong>
      </div>
      <div class="page-metric-card is-warning">
        <span>Controller</span>
        <strong>{{ brokerStats.controllers }}</strong>
      </div>
    </div>

    <el-card class="content-card">
      <template #header>
        <div class="card-header">
          <span>节点摘要</span>
          <span class="card-subtitle">热点 Broker</span>
        </div>
      </template>

      <div class="compact-list">
        <div v-for="item in hotspotBrokers" :key="item.id" class="compact-item">
          <div>
            <strong>Broker {{ item.id }}</strong>
            <span>{{ item.riskReason }}</span>
          </div>
          <el-tag :type="item.riskLevel === 'high' ? 'danger' : 'warning'">
            {{ item.riskLevel === 'high' ? '高风险' : '关注' }}
          </el-tag>
        </div>
        <div v-if="hotspotBrokers.length === 0" class="compact-item">
          <div>
            <strong>当前状态</strong>
            <span>{{ brokerRiskSummary }}</span>
          </div>
          <el-tag type="success">正常</el-tag>
        </div>
      </div>
    </el-card>

    <el-card class="content-card filter-card">
      <div class="toolbar-row">
        <div class="toolbar-left">
          <span class="toolbar-summary">顶部“当前 Kafka 集群”会决定当前 Broker 页面上下文。</span>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" @click="loadBrokers" :loading="loading">查询</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>Broker 列表</span>
          <span class="card-subtitle">节点列表</span>
        </div>
      </template>

      <el-table :data="brokers" empty-text="暂无 Broker 数据">
        <el-table-column prop="id" label="Broker ID" width="120" />
        <el-table-column prop="address" label="地址" min-width="220" />
        <el-table-column label="控制器" width="120">
          <template #default="{ row }">
            <el-tag :type="row.isController ? 'danger' : 'info'">{{ row.isController ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="连接状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.connected ? 'success' : 'danger'">{{ row.connected ? '已连接' : '断开' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="leaderPartitionCount" label="Leader 分区" width="130" />
        <el-table-column prop="replicaPartitionCount" label="Replica 分区" width="130" />
        <el-table-column label="Topics" min-width="260">
          <template #default="{ row }">{{ (row.topics || []).join(', ') || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="permStore.hasPerm('kafka:broker:config:update') || permStore.roles.includes('admin')"
              link
              type="primary"
              :disabled="!row.connected"
              @click="openBrokerConfigDialog(row)"
            >
              动态配置
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <BrokerConfigDialog ref="brokerConfigDialogRef" @success="loadBrokers" />
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { ElMessage } from 'element-plus'
import { getKafkaBrokers } from '@/api/kafka.js'
import BrokerConfigDialog from '@/components/BrokerConfigDialog.vue'
import { useKafkaStore } from '@/stores/kafkaStore.js'
import { usePermissionStore } from '@/stores/permissionStore.js'

const loading = ref(false)
const brokers = ref([])
const brokerConfigDialogRef = ref()
const kafkaStore = useKafkaStore()
const permStore = usePermissionStore()
const { clusterOptions: clusters, selectedClusterId } = storeToRefs(kafkaStore)

const currentClusterName = computed(
  () => clusters.value.find((item) => item.id === selectedClusterId.value)?.name || '-',
)

const brokerStats = computed(() => ({
  total: brokers.value.length,
  connected: brokers.value.filter((item) => item.connected).length,
  controllers: brokers.value.filter((item) => item.isController).length,
}))

const hotspotBrokers = computed(() => {
  const maxLeaderPartitions = brokers.value.reduce(
    (maxValue, item) => Math.max(maxValue, Number(item.leaderPartitionCount || 0)),
    0,
  )

  return brokers.value
    .map((item) => {
      const leaderPartitionCount = Number(item.leaderPartitionCount || 0)
      const replicaPartitionCount = Number(item.replicaPartitionCount || 0)
      const reasons = []
      let score = 0

      if (!item.connected) {
        reasons.push('当前连接状态异常')
        score += 3
      }
      if (item.isController && !item.connected) {
        reasons.push('Controller 节点未连接')
        score += 3
      }
      if (leaderPartitionCount > 0 && leaderPartitionCount === maxLeaderPartitions && leaderPartitionCount >= 10) {
        reasons.push(`Leader 分区承载偏高（${leaderPartitionCount}）`)
        score += 2
      }
      if (replicaPartitionCount >= 20) {
        reasons.push(`Replica 承载较高（${replicaPartitionCount}）`)
        score += 1
      }

      return {
        ...item,
        riskScore: score,
        riskLevel: score >= 4 ? 'high' : 'medium',
        riskReason: reasons.join('；') || '当前未识别到明显风险信号',
      }
    })
    .filter((item) => item.riskScore > 0)
    .sort((a, b) => b.riskScore - a.riskScore || Number(b.leaderPartitionCount || 0) - Number(a.leaderPartitionCount || 0))
    .slice(0, 5)
})

const brokerRiskSummary = computed(() => {
  if (brokers.value.length === 0) return '当前没有 Broker 数据。'
  const leaderLoads = brokers.value.map((item) => Number(item.leaderPartitionCount || 0))
  const maxLoad = leaderLoads.reduce((maxValue, value) => Math.max(maxValue, value), 0)
  const minLoad = leaderLoads.length > 0 ? Math.min(...leaderLoads) : 0
  return maxLoad - minLoad >= 10
    ? `Leader 分区负载存在明显偏斜，最高 ${maxLoad}、最低 ${minLoad}，建议检查分区分布。`
    : `Leader 分区负载差异可控，最高 ${maxLoad}、最低 ${minLoad}。`
})

const loadClusters = async ({ force = false } = {}) => {
  try {
    await kafkaStore.loadClusterOptions({ force })
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 集群列表加载失败')
    throw error
  }
}

const loadBrokers = async () => {
  if (loading.value || !selectedClusterId.value) return
  loading.value = true
  try {
    const res = await getKafkaBrokers(selectedClusterId.value)
    brokers.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || 'Broker 数据加载失败')
  } finally {
    loading.value = false
  }
}

const openBrokerConfigDialog = (broker) => {
  if (!broker?.id) return
  brokerConfigDialogRef.value?.openDialog(broker.id)
}

onMounted(async () => {
  try {
    await loadClusters()
    await loadBrokers()
  } catch (error) {
    if (!String(error?.message || '').includes('Kafka 集群列表加载失败')) {
      ElMessage.error(error.message || 'Kafka 集群加载失败')
    }
  }
})

watch(selectedClusterId, async (nextValue, previousValue) => {
  if (!nextValue || nextValue === previousValue) return
  await loadBrokers()
})
</script>

<style scoped>
.toolbar-summary {
  color: var(--shell-text-soft, var(--text-sub));
  font-size: 13px;
  line-height: 1.6;
}
</style>
