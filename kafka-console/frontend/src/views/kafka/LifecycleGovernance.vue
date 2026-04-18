<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div><h2>生命周期治理</h2><p>识别缺少元数据、临时 Topic 和保留策略不合理的资源，给出治理建议</p></div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="loadData"><el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" /></el-select>
          <el-button type="primary" :loading="loading" @click="loadData">生成报告</el-button>
        </div>
      </div>
    </el-card>
    <el-card class="content-card" v-loading="loading">
      <el-table :data="rows" empty-text="暂无治理建议">
        <el-table-column prop="topicName" label="Topic" min-width="180" />
        <el-table-column prop="owner" label="Owner" width="120" />
        <el-table-column prop="action" label="建议动作" width="160" />
        <el-table-column prop="status" label="状态" width="120" />
        <el-table-column prop="targetRetentionHours" label="目标保留(h)" width="140" />
        <el-table-column prop="recommendation" label="建议说明" min-width="360" show-overflow-tooltip />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getKafkaClusterOptions, getKafkaLifecyclePolicies } from '@/api/kafka.js'

const clusters = ref([]); const selectedClusterId = ref(null); const loading = ref(false); const rows = ref([])
const loadClusters = async () => { const res = await getKafkaClusterOptions(); clusters.value = res?.data?.data || []; if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id }
const loadData = async () => { if (!selectedClusterId.value) return; loading.value = true; try { const res = await getKafkaLifecyclePolicies({ clusterId: selectedClusterId.value }); rows.value = res?.data?.data || [] } catch (error) { ElMessage.error(error.message || '生命周期治理加载失败') } finally { loading.value = false } }
onMounted(async () => { await loadClusters(); await loadData() })
</script>
