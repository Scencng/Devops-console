<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>自动扩容建议</h2>
          <p>基于 Lag、分区规模和当前并发，给出 Topic 分区与消费者扩容建议</p>
        </div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="loadData">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-button type="primary" :loading="loading" @click="loadData">重新生成</el-button>
        </div>
      </div>
    </el-card>
    <el-card class="content-card" v-loading="loading">
      <el-table :data="rows" empty-text="暂无扩容建议">
        <el-table-column prop="resourceType" label="资源类型" width="140" />
        <el-table-column prop="resourceName" label="资源名称" min-width="200" />
        <el-table-column label="当前值" width="120"><template #default="{ row }">{{ row.currentValue }}</template></el-table-column>
        <el-table-column label="建议值" width="120"><template #default="{ row }">{{ row.recommendedValue }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="120" />
        <el-table-column prop="reason" label="建议说明" min-width="360" show-overflow-tooltip />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getKafkaClusterOptions, getKafkaScalingRecommendations } from '@/api/kafka.js'

const clusters = ref([])
const selectedClusterId = ref(null)
const loading = ref(false)
const rows = ref([])

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id
}
const loadData = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const res = await getKafkaScalingRecommendations({ clusterId: selectedClusterId.value })
    rows.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || '扩容建议生成失败')
  } finally {
    loading.value = false
  }
}
onMounted(async () => {
  await loadClusters()
  await loadData()
})
</script>
