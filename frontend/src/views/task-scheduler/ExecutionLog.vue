<template>
  <div class="execution-log-container">
    <div class="log-header">
      <div class="header-left">
        <el-button @click="$router.back()" :icon="ArrowLeft" circle />
        <span class="title">执行日志 #{{ executionId }}</span>
      </div>
      <div class="header-right">
        <el-tag :type="statusType" effect="dark">{{ executionInfo.status }}</el-tag>
        <span class="time-info" v-if="executionInfo.startTime">
          耗时: {{ duration }}
        </span>
      </div>
    </div>

    <div class="execution-info-bar">
      <div class="info-item">
        <span class="label">开始时间:</span>
        <span class="value">{{ formatTime(executionInfo.startTime) }}</span>
      </div>
      <div class="info-item" v-if="executionInfo.endTime">
        <span class="label">结束时间:</span>
        <span class="value">{{ formatTime(executionInfo.endTime) }}</span>
      </div>
    </div>

    <div class="terminal-container" ref="terminalRef">
      <div v-if="loading" class="loading-state">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>正在加载日志...</span>
      </div>
      <template v-else>
        <div v-for="node in nodeExecutions" :key="node.id" class="node-log-section">
          <div class="node-header" @click="toggleNode(node.id)">
            <el-icon :class="{ 'is-active': expandedNodes.includes(node.id) }">
              <CaretRight />
            </el-icon>
            <span class="node-name">节点: {{ node.nodeName || node.nodeId }}</span>
            <el-tag size="small" :type="nodeStatusType(node.status)" class="node-status">{{ node.status }}</el-tag>
            <span class="node-duration" v-if="node.duration">{{ node.duration }}ms</span>
          </div>
          <div v-if="expandedNodes.includes(node.id)" class="node-content">
            <pre class="log-text">{{ node.logs || '无日志内容' }}</pre>
            <div v-if="node.errorMessage" class="error-text">
              错误: {{ node.errorMessage }}
            </div>
          </div>
        </div>
        <div v-if="realtimeLogs.length > 0" class="realtime-section">
          <div class="section-title">实时输出:</div>
          <pre class="log-text">{{ realtimeLogs.join('\n') }}</pre>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft, CaretRight, Loading } from '@element-plus/icons-vue'
import { getExecutionLogs } from '@/api/workflow'
import dayjs from 'dayjs'

const route = useRoute()
const executionId = route.params.id
const executionInfo = ref({})
const nodeExecutions = ref([])
const expandedNodes = ref([])
const realtimeLogs = ref([])
const terminalRef = ref(null)
const loading = ref(false)
let ws = null

const statusType = computed(() => {
  const status = (executionInfo.value.status || '').toLowerCase()
  if (status === 'success' || status === 'completed') return 'success'
  if (status === 'running') return 'primary'
  if (status === 'failed') return 'danger'
  return 'info'
})

const duration = computed(() => {
  if (!executionInfo.value.startTime) return '-'
  const start = dayjs(executionInfo.value.startTime)
  const end = executionInfo.value.endTime ? dayjs(executionInfo.value.endTime) : dayjs()
  const diff = end.diff(start, 'second')
  if (diff < 60) return `${diff}s`
  return `${Math.floor(diff / 60)}m ${diff % 60}s`
})

const nodeStatusType = (status) => {
  const s = (status || '').toLowerCase()
  if (s === 'success' || s === 'completed') return 'success'
  if (s === 'running') return 'primary'
  if (s === 'failed') return 'danger'
  return 'info'
}

const formatTime = (time) => {
  if (!time) return '-'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const toggleNode = (id) => {
  const index = expandedNodes.value.indexOf(id)
  if (index > -1) {
    expandedNodes.value.splice(index, 1)
  } else {
    expandedNodes.value.push(id)
  }
}

const fetchLogs = async () => {
  loading.value = true
  try {
    const res = await getExecutionLogs(executionId)
    executionInfo.value = res.data.execution
    nodeExecutions.value = res.data.nodes
    // 默认展开所有节点
    expandedNodes.value = nodeExecutions.value.map(n => n.id)
  } catch (err) {
    console.error('Failed to fetch logs:', err)
  } finally {
    loading.value = false
  }
}

const initWebSocket = () => {
  const wsBaseUrl = import.meta.env.VITE_WS_BASE_URL || `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}`
  ws = new WebSocket(`${wsBaseUrl}/ws/executions/${executionId}/logs`)

  ws.onmessage = (event) => {
    realtimeLogs.value.push(event.data)
    nextTick(() => {
      if (terminalRef.value) {
        terminalRef.value.scrollTop = terminalRef.value.scrollHeight
      }
    })
  }

  ws.onclose = () => {
    console.log('WebSocket closed')
    fetchLogs()
  }

  ws.onerror = (error) => {
    console.error('WebSocket error:', error)
  }
}

onMounted(() => {
  fetchLogs()
  initWebSocket()
})

onUnmounted(() => {
  if (ws) ws.close()
})
</script>

<style scoped>
.execution-log-container {
  padding: 20px;
  height: calc(100vh - 120px);
  display: flex;
  flex-direction: column;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 15px;
}

.title {
  font-size: 18px;
  font-weight: bold;
}

.time-info {
  margin-left: 15px;
  font-size: 14px;
  color: #666;
}

.execution-info-bar {
  display: flex;
  gap: 30px;
  background-color: #f5f7fa;
  padding: 10px 15px;
  border-radius: 4px;
  margin-bottom: 20px;
  font-size: 14px;
}

.info-item .label {
  color: #909399;
  margin-right: 8px;
}

.terminal-container {
  flex: 1;
  background-color: #1e1e1e;
  color: #d4d4d4;
  padding: 15px;
  border-radius: 4px;
  overflow-y: auto;
  font-family: 'JetBrains Mono', 'Courier New', Courier, monospace;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 10px;
  color: #888;
}

.node-log-section {
  margin-bottom: 10px;
  border-bottom: 1px solid #333;
}

.node-header {
  display: flex;
  align-items: center;
  padding: 8px 0;
  cursor: pointer;
  user-select: none;
}

.node-header:hover {
  background-color: #2d2d2d;
}

.node-name {
  margin-left: 8px;
  font-weight: bold;
}

.node-status {
  margin-left: 10px;
}

.node-duration {
  margin-left: auto;
  color: #888;
  font-size: 12px;
}

.node-content {
  padding: 10px 20px;
}

.log-text {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 13px;
  color: #b5cea8;
}

.error-text {
  color: #f44336;
  margin-top: 5px;
  font-weight: bold;
}

.realtime-section {
  margin-top: 20px;
}

.section-title {
  color: #569cd6;
  margin-bottom: 5px;
}

.is-active {
  transform: rotate(90deg);
}

.el-icon {
  transition: transform 0.3s;
}
</style>
