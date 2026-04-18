<template>
  <div class="page-container">
    <el-card class="page-header-card"><div class="page-header"><div><h2>Broker 管理</h2><p>查看 Broker 分布、控制器和分区承载情况</p></div><div class="header-actions"><el-select v-model="selectedClusterId" placeholder="选择 Kafka 集群" style="width: 260px" @change="loadBrokers"><el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" /></el-select><el-button @click="loadBrokers" :loading="loading">刷新</el-button></div></div></el-card>
    <el-card class="content-card" v-loading="loading"><el-table :data="brokers" empty-text="暂无 Broker 数据"><el-table-column prop="id" label="Broker ID" width="120" /><el-table-column prop="address" label="地址" min-width="220" /><el-table-column label="控制器" width="120"><template #default="scope"><el-tag :type="scope.row.isController ? 'danger' : 'info'">{{ scope.row.isController ? '是' : '否' }}</el-tag></template></el-table-column><el-table-column label="连接状态" width="120"><template #default="scope"><el-tag :type="scope.row.connected ? 'success' : 'danger'">{{ scope.row.connected ? '已连接' : '断开' }}</el-tag></template></el-table-column><el-table-column prop="leaderPartitionCount" label="Leader 分区" width="130" /><el-table-column prop="replicaPartitionCount" label="Replica 分区" width="130" /><el-table-column label="Topics" min-width="260"><template #default="scope">{{ (scope.row.topics || []).join(', ') || '-' }}</template></el-table-column></el-table></el-card>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getKafkaBrokers, getKafkaClusterOptions } from '@/api/kafka.js'
const loading = ref(false)
const clusters = ref([])
const brokers = ref([])
const selectedClusterId = ref(null)
const loadClusters = async () => { const res = await getKafkaClusterOptions(); clusters.value = res?.data?.data || []; if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id }
const loadBrokers = async () => { if (!selectedClusterId.value) return; loading.value = true; try { const res = await getKafkaBrokers(selectedClusterId.value); brokers.value = res?.data?.data || [] } catch (error) { ElMessage.error(error.message || 'Broker 数据加载失败') } finally { loading.value = false } }
onMounted(async () => { try { await loadClusters(); await loadBrokers() } catch (error) { ElMessage.error(error.message || 'Kafka 集群加载失败') } })
</script>
