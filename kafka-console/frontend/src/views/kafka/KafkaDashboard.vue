<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>Kafka 总览</h2>
          <p>查看 Broker、Topic、Consumer Group 的运行概况</p>
        </div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" placeholder="选择 Kafka 集群" style="width: 280px" @change="loadDashboard">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-button @click="loadDashboard" :loading="loading">刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-empty v-if="!selectedClusterId && !loading" description="请先创建并选择一个 Kafka 集群" />

    <template v-else>
      <el-row :gutter="20" class="stats-row" v-loading="loading">
        <el-col :span="6"><el-card><div class="stat"><span>Broker</span><strong>{{ dashboard.brokerCount }}</strong></div></el-card></el-col>
        <el-col :span="6"><el-card><div class="stat"><span>Topic</span><strong>{{ dashboard.topicCount }}</strong></div></el-card></el-col>
        <el-col :span="6"><el-card><div class="stat"><span>消费组</span><strong>{{ dashboard.consumerGroupCount }}</strong></div></el-card></el-col>
        <el-col :span="6"><el-card><div class="stat"><span>总 Lag</span><strong>{{ dashboard.totalLag }}</strong></div></el-card></el-col>
      </el-row>

      <el-card class="content-card" v-loading="loading">
        <template #header><div class="card-header"><span>Lag Top 5 消费组</span></div></template>
        <el-table :data="dashboard.topLagGroups || []" empty-text="暂无消费组数据">
          <el-table-column prop="groupId" label="消费组" min-width="220" />
          <el-table-column prop="state" label="状态" width="120">
            <template #default="scope"><el-tag :type="scope.row.state === 'Stable' ? 'success' : 'warning'">{{ scope.row.state || 'Unknown' }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="memberCount" label="成员数" width="120" />
          <el-table-column prop="partitionCount" label="分区数" width="120" />
          <el-table-column prop="committedLag" label="Lag" width="160" />
          <el-table-column label="Topics" min-width="220"><template #default="scope">{{ (scope.row.topics || []).join(', ') || '-' }}</template></el-table-column>
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
const dashboard = ref({ brokerCount: 0, topicCount: 0, consumerGroupCount: 0, totalPartitions: 0, totalLag: 0, topLagGroups: [] })
const currentClusterName = computed(() => clusters.value.find(item => item.id === selectedClusterId.value)?.name || '-')

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id
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

<style scoped>
.stats-row { margin: 20px 0; }
.stat { display: flex; flex-direction: column; gap: 10px; }
.stat span { color: #909399; }
.stat strong { font-size: 28px; font-weight: 700; }
</style>
