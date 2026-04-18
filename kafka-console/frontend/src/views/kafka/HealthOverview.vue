<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>健康总览</h2>
          <p>统一查看 Kafka 集群健康、异常 Broker、积压风险和未恢复告警</p>
        </div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="loadData">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-button :loading="loading" @click="loadData">刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16" v-loading="loading">
      <el-col v-for="card in overview.cards || []" :key="card.name" :span="6" class="mb-16">
        <el-card class="stat-card">
          <div class="stat-item">
            <span>{{ card.name }}</span>
            <strong>{{ card.value }}</strong>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="content-card" v-loading="loading">
      <template #header><div class="card-header">健康问题清单</div></template>
      <el-table :data="overview.issues || []" empty-text="当前无明显健康问题">
        <el-table-column prop="severity" label="级别" width="120">
          <template #default="{ row }">
            <el-tag :type="row.severity === 'critical' ? 'danger' : row.severity === 'warning' ? 'warning' : 'info'">
              {{ row.severity }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="问题" min-width="240" />
        <el-table-column prop="detail" label="详情" min-width="360" show-overflow-tooltip />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getKafkaClusterOptions, getKafkaHealthOverview } from '@/api/kafka.js'

const loading = ref(false)
const clusters = ref([])
const selectedClusterId = ref(null)
const overview = ref({ cards: [], issues: [] })

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id
}

const loadData = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const res = await getKafkaHealthOverview(selectedClusterId.value)
    overview.value = res?.data?.data || overview.value
  } catch (error) {
    ElMessage.error(error.message || '健康总览加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    await loadClusters()
    await loadData()
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 健康总览初始化失败')
  }
})
</script>

<style scoped>
.mb-16 { margin-bottom: 16px; }
.stat-item { display:flex; flex-direction:column; gap:10px; }
.stat-item span { color:#909399; }
.stat-item strong { font-size:28px; font-weight:700; }
</style>
