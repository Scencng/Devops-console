<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div><h2>成本治理</h2><p>生成 Kafka 存储、入流量、出流量和分区规模的估算成本记录</p></div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="loadData"><el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" /></el-select>
          <el-button type="primary" :loading="generating" @click="generate">生成记录</el-button>
        </div>
      </div>
    </el-card>
    <el-card class="content-card" v-loading="loading">
      <el-table :data="rows" empty-text="暂无成本记录">
        <el-table-column prop="metricDate" label="日期" min-width="180" />
        <el-table-column prop="storageBytes" label="存储(B)" min-width="160" />
        <el-table-column prop="ingressBytes" label="入流量(B)" min-width="160" />
        <el-table-column prop="egressBytes" label="出流量(B)" min-width="160" />
        <el-table-column prop="partitionCount" label="分区数" width="120" />
        <el-table-column label="估算成本" width="160"><template #default="{ row }">{{ row.estimatedCost }} {{ row.currency }}</template></el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { generateKafkaCostRecord, getKafkaClusterOptions, getKafkaCostRecords } from '@/api/kafka.js'

const clusters = ref([]); const selectedClusterId = ref(null); const loading = ref(false); const generating = ref(false); const rows = ref([])
const loadClusters = async () => { const res = await getKafkaClusterOptions(); clusters.value = res?.data?.data || []; if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id }
const loadData = async () => { if (!selectedClusterId.value) return; loading.value = true; try { const res = await getKafkaCostRecords({ clusterId: selectedClusterId.value }); rows.value = res?.data?.data || [] } catch (error) { ElMessage.error(error.message || '成本记录加载失败') } finally { loading.value = false } }
const generate = async () => { if (!selectedClusterId.value) return; generating.value = true; try { await generateKafkaCostRecord({ clusterId: selectedClusterId.value }); ElMessage.success('成本记录已生成'); await loadData() } catch (error) { ElMessage.error(error.message || '成本记录生成失败') } finally { generating.value = false } }
onMounted(async () => { await loadClusters(); await loadData() })
</script>
