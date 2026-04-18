<template>
  <div class="event-management">
    <el-card class="page-header-card cyber-card">
      <div class="page-header cyber-header">
        <div>
          <h2>Event 管理</h2>
          <p>实时监控 Kubernetes 集群事件</p>
        </div>
        <div class="header-actions">
          <div class="namespace-selector">
            <span class="selector-label">命名空间：</span>
            <el-select v-model="selectedNamespace" placeholder="选择命名空间" @change="fetchData" style="width: 200px;" class="autoops-select">
              <el-option label="所有" value="all" />
              <el-option
                v-for="ns in namespaceList"
                :key="ns.name"
                :label="ns.name"
                :value="ns.name"
              />
            </el-select>
          </div>
          <el-button type="primary" class="autoops-btn-primary" @click="fetchData" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card cyber-card">
      <div class="autoops-table-wrapper">
        <el-table 
          :data="eventList" 
          style="width: 100%" class="autoops-table" 
          v-loading="loading"
          element-loading-background="rgba(0, 0, 0, 0.5)"
          height="calc(100vh - 280px)"
          :empty-text="loading ? '加载中...' : '暂无数据'"
          :default-sort="{ prop: 'lastTimestamp', order: 'descending' }"
        >
          <el-table-column label="类型" width="100" align="center">
            <template #default="scope">
              <el-tag 
                :type="scope.row.type === 'Warning' ? 'danger' : 'success'" 
                size="small"
                effect="dark"
              >
                {{ scope.row.type }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="involvedKind" label="资源类型" width="120" />
          <el-table-column prop="involvedObject" label="资源名称" min-width="150" show-overflow-tooltip />
          <el-table-column prop="namespace" label="命名空间" width="120" />
          <el-table-column prop="reason" label="原因" width="150" show-overflow-tooltip />
          <el-table-column prop="message" label="消息" min-width="250" show-overflow-tooltip />
          <el-table-column prop="source" label="来源" width="150" show-overflow-tooltip />
          <el-table-column prop="count" label="次数" width="80" align="center" />
          <el-table-column label="最后发生" width="170" sortable prop="lastTimestamp">
            <template #default="scope">
              {{ formatTimestamp(scope.row.lastTimestamp) }}
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { getEventList } from '@/api/k8s/event'
import { getNamespaceList } from '@/api/k8s/namespace'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import dayjs from 'dayjs'

const loading = ref(false)
const eventList = ref([])
const namespaceList = ref([])
const selectedNamespace = ref('all')
let refreshInterval = null

const formatTimestamp = (timestamp) => {
  if (typeof timestamp === 'number') {
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
  }
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

const fetchNamespaces = async () => {
  const instanceId = getSelectedInstanceId()
  const res = await getNamespaceList(instanceId)
  if (res.status === 200) {
    namespaceList.value = res.data.namespaceList
  }
}

const fetchData = async () => {
  const instanceId = getSelectedInstanceId()
  if (!instanceId) return
  
  loading.value = true
  try {
    const res = await getEventList(selectedNamespace.value, instanceId)
    if (res.status === 200) {
      eventList.value = res.data.eventList
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('获取Event列表失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchNamespaces()
  fetchData()
  
  // Auto-refresh every 30 seconds
  refreshInterval = setInterval(() => {
    fetchData()
  }, 30000)
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<style scoped>
.page-header-card {
  margin-bottom: 0px !important;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.event-management {
  padding: 20px;
}
.header-actions {
  display: flex;
  gap: 10px;
}
.namespace-selector {
  display: flex;
  align-items: center;
}
</style>
