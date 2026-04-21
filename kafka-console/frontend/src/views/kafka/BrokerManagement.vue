<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <div class="page-eyebrow">Brokers</div>
          <h2>Broker 管理</h2>
          <p>查看 Broker 节点、Controller 角色与分区承载情况，快速判断集群当前的节点分布是否健康。</p>
        </div>
      </div>
    </el-card>

    <div class="page-metrics">
      <div class="page-metric-card is-accent">
        <span>当前集群</span>
        <strong>{{ currentClusterName }}</strong>
        <p>当前正在查看的 Kafka 集群。</p>
      </div>
      <div class="page-metric-card">
        <span>Broker 节点</span>
        <strong>{{ brokerStats.total }}</strong>
        <p>当前集群返回的 Broker 数量。</p>
      </div>
      <div class="page-metric-card is-success">
        <span>已连接</span>
        <strong>{{ brokerStats.connected }}</strong>
        <p>当前能正常返回连接状态的 Broker。</p>
      </div>
      <div class="page-metric-card is-warning">
        <span>Controller</span>
        <strong>{{ brokerStats.controllers }}</strong>
        <p>通常应只有一个 Controller 节点。</p>
      </div>
    </div>

    <el-card class="content-card">
      <template #header>
        <div class="card-header card-header-wrap">
          <span>热点 Broker / Controller 风险提示</span>
          <span class="card-subtitle">优先识别连接异常、Controller 数异常和分区承载过热的 Broker</span>
        </div>
      </template>

      <div class="workbench-grid">
        <div class="workspace-panel">
          <h3>热点 Broker</h3>
          <p>按 Leader 分区承载、Replica 承载和连接状态综合识别最值得优先检查的 Broker。</p>
          <div class="compact-list">
            <div v-for="item in hotspotBrokers" :key="item.id" class="compact-item">
              <div>
                <strong>Broker {{ item.id }} / {{ item.address }}</strong>
                <span>{{ item.riskReason }}</span>
              </div>
              <el-tag :type="item.riskLevel === 'high' ? 'danger' : 'warning'">
                {{ item.riskLevel === 'high' ? '高风险' : '关注' }}
              </el-tag>
            </div>
          </div>
        </div>

        <div class="workspace-panel">
          <h3>Controller 风险提示</h3>
          <p>快速判断 Controller 角色是否稳定，以及当前集群是否出现明显异常信号。</p>
          <div class="compact-list">
            <div class="compact-item">
              <div>
                <strong>Controller 数量</strong>
                <span>
                  {{
                    brokerStats.controllers === 1
                      ? '当前只检测到 1 个 Controller，符合预期。'
                      : brokerStats.controllers === 0
                        ? '当前没有检测到 Controller，建议立即排查集群元数据状态。'
                        : `当前检测到 ${brokerStats.controllers} 个 Controller，可能存在异常切换或状态不一致。`
                  }}
                </span>
              </div>
            </div>
            <div class="compact-item">
              <div>
                <strong>连接状态</strong>
                <span>
                  {{
                    brokerStats.connected === brokerStats.total
                      ? '所有 Broker 当前都处于连接状态。'
                      : `当前有 ${brokerStats.total - brokerStats.connected} 个 Broker 未连接，建议先排查网络或节点状态。`
                  }}
                </span>
              </div>
            </div>
            <div class="compact-item">
              <div>
                <strong>承载偏斜</strong>
                <span>{{ brokerRiskSummary }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <el-card class="content-card filter-card">
      <div class="toolbar-row">
        <div class="toolbar-left">
          <el-select
            v-model="selectedClusterId"
            placeholder="选择 Kafka 集群"
            style="width: 300px"
            @change="loadBrokers"
          >
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
        </div>
        <div class="toolbar-right">
          <el-button @click="loadBrokers" :loading="loading">刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>Broker 列表</span>
          <span class="card-subtitle">聚焦 Controller、连接状态与分区承载</span>
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
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getKafkaBrokers, getKafkaClusterOptions } from '@/api/kafka.js'

const loading = ref(false)
const clusters = ref([])
const brokers = ref([])
const selectedClusterId = ref(null)

const currentClusterName = computed(
  () => clusters.value.find((item) => item.id === selectedClusterId.value)?.name || '-',
)

const brokerStats = computed(() => ({
  total: brokers.value.length,
  connected: brokers.value.filter((item) => item.connected).length,
  controllers: brokers.value.filter((item) => item.isController).length,
}))

const hotspotBrokers = computed(() => {
  const maxLeaderPartitions = Math.max(...brokers.value.map((item) => Number(item.leaderPartitionCount || 0)), 0)

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
  const maxLoad = Math.max(...leaderLoads, 0)
  const minLoad = Math.min(...leaderLoads, 0)
  return maxLoad - minLoad >= 10
    ? `Leader 分区负载存在明显偏斜，最高 ${maxLoad}、最低 ${minLoad}，建议检查分区分布。`
    : `Leader 分区负载差异可控，最高 ${maxLoad}、最低 ${minLoad}。`
})

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) {
    selectedClusterId.value = clusters.value[0].id
  }
}

const loadBrokers = async () => {
  if (!selectedClusterId.value) return
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

onMounted(async () => {
  try {
    await loadClusters()
    await loadBrokers()
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 集群加载失败')
  }
})
</script>
