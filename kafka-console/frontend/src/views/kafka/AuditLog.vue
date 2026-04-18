<template>
  <div class="page-container">
    <el-card class="page-header-card"><div class="page-header"><div><h2>审计日志</h2><p>查看 Kafka 危险操作的审计记录</p></div><div class="header-actions"><el-button @click="loadLogs" :loading="loading">刷新</el-button></div></div></el-card>
    <el-card class="content-card">
      <el-form inline>
        <el-form-item label="集群"><el-select v-model="filters.clusterId" clearable style="width:220px"><el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" /></el-select></el-form-item>
        <el-form-item label="操作"><el-input v-model="filters.action" placeholder="如 topic:delete" /></el-form-item>
        <el-form-item label="结果"><el-select v-model="filters.result" clearable style="width:160px"><el-option label="成功" value="success" /><el-option label="失败" value="failed" /></el-select></el-form-item>
        <el-form-item><el-button type="primary" @click="loadLogs">查询</el-button></el-form-item>
      </el-form>
    </el-card>
    <el-card class="content-card" v-loading="loading"><el-table :data="logs" empty-text="暂无审计日志"><el-table-column prop="createdAt" label="时间" width="180"><template #default="scope">{{ formatTime(scope.row.createdAt) }}</template></el-table-column><el-table-column prop="operatorUsername" label="操作人" width="140" /><el-table-column prop="action" label="动作" width="200" /><el-table-column prop="resourceType" label="资源类型" width="120" /><el-table-column prop="resourceName" label="资源名称" min-width="200" /><el-table-column prop="result" label="结果" width="100"><template #default="scope"><el-tag :type="scope.row.result === 'success' ? 'success' : 'danger'">{{ scope.row.result }}</el-tag></template></el-table-column><el-table-column prop="errorMessage" label="错误信息" min-width="220" show-overflow-tooltip /></el-table></el-card>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getKafkaAuditLogs, getKafkaClusterOptions } from '@/api/kafka.js'
const loading = ref(false)
const clusters = ref([])
const logs = ref([])
const filters = reactive({ clusterId: null, action: '', result: '', page: 1, pageSize: 100 })
const formatTime = (value) => value ? new Date(value).toLocaleString() : '-'
const loadClusters = async () => { const res = await getKafkaClusterOptions(); clusters.value = res?.data?.data || [] }
const loadLogs = async () => { loading.value = true; try { const res = await getKafkaAuditLogs(filters); logs.value = res?.data?.data?.list || [] } catch (error) { ElMessage.error(error.message || '审计日志加载失败') } finally { loading.value = false } }
onMounted(async () => { await loadClusters(); await loadLogs() })
</script>
