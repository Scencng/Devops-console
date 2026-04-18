<template>
  <el-card shadow="never" class="monitor-card" style="margin-bottom: 20px; position: relative;">
    <div class="card-header">
      <div class="card-title">
        <el-icon class="title-icon"><span v-html="getIconHtml"></span></el-icon>
        {{ title }}
      </div>
      <!-- Actions shown only if id exists (custom charts) -->
      <div class="card-actions" v-if="monitorId">
        <el-button type="primary" link size="small" @click="$emit('edit', monitorId)">
          <el-icon><Edit /></el-icon>
        </el-button>
        <el-button type="danger" link size="small" @click="$emit('delete', monitorId)">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
    </div>
    
    <div class="card-body">
      <div v-if="loading" class="loading-state">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>数据加载中...</span>
      </div>
      <div v-else-if="error" class="error-state">
        <el-icon><Warning /></el-icon>
        <span>{{ error }}</span>
      </div>
      <div v-else-if="!chartData || chartData.length === 0" class="empty-state">
        <el-icon><DataLine /></el-icon>
        <span>暂无数据</span>
      </div>
      <v-chart v-else class="chart" :option="chartOption" autoresize />
    </div>
  </el-card>
</template>

<script setup>
import { ref, watch, onMounted, computed, onUnmounted } from 'vue'
import { Loading, Warning, DataLine, Edit, Delete } from '@element-plus/icons-vue'
import axios from 'axios'
import VChart from 'vue-echarts'
import * as echarts from 'echarts'

const props = defineProps({
  monitorId: { type: Number, default: null },
  title: { type: String, required: true },
  promqlTemplate: { type: String, required: true },
  chartType: { type: String, default: 'line' },
  unit: { type: String, default: '' },
  color: { type: String, default: '#409EFF' },
  nodeName: { type: String, default: '' },
  podName: { type: String, default: '' },
  prometheusInstanceId: { type: Number, required: true },
  hours: { type: Number, default: 1 }
})

const emit = defineEmits(['edit', 'delete'])

const loading = ref(false)
const error = ref('')
const chartData = ref([])
const times = ref([])

let refreshInterval = null

// Compute actual PromQL by replacing template tags
const computedPromQL = computed(() => {
  let q = props.promqlTemplate
  if (props.nodeName) {
    q = q.replace(/\{\{nodeName\}\}/g, props.nodeName)
  }
  if (props.podName) {
    q = q.replace(/\{\{podName\}\}/g, props.podName)
  }
  return q
})

const getIconHtml = computed(() => {
  if (props.chartType === 'line') return '<svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg"><path fill="currentColor" d="M160 832h704a32 32 0 110 64H128a32 32 0 01-32-32V128a32 32 0 0164 0v704z"></path><path fill="currentColor" d="M840.448 317.056a32 32 0 0144.16 46.208l-384 368a32 32 0 01-44.544.576L320 598.528l-157.696 157.76a32 32 0 11-45.248-45.248l180.352-180.352a32 32 0 0144.608-.576l136 133.248 362.432-346.304z"></path></svg>'
  if (props.chartType === 'bar') return '<svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg"><path fill="currentColor" d="M160 832h704a32 32 0 110 64H128a32 32 0 01-32-32V128a32 32 0 0164 0v704z"></path><path fill="currentColor" d="M256 640a32 32 0 0132-32h64a32 32 0 0132 32v128a32 32 0 01-32 32h-64a32 32 0 01-32-32V640zm192-256a32 32 0 0132-32h64a32 32 0 0132 32v384a32 32 0 01-32 32h-64a32 32 0 01-32-32V384zm192-128a32 32 0 0132-32h64a32 32 0 0132 32v512a32 32 0 01-32 32h-64a32 32 0 01-32-32V256z"></path></svg>'
  return ''
})

const fetchMetrics = async () => {
  if (!props.prometheusInstanceId) return
  if (!props.nodeName && !props.podName) return
  if (!computedPromQL.value) return

  if (!chartData.value || chartData.value.length === 0) {
    loading.value = true
  }
  error.value = ''
  
  try {
    const end = Math.floor(Date.now() / 1000)
    const start = end - (props.hours * 3600)
    const step = Math.max(15, Math.floor((end - start) / 60))

    const queryUrl = `/api/v1/monitor/prometheus/${props.prometheusInstanceId}/api/v1/query_range`
    const token = localStorage.getItem('access_token')
    
    // Use raw axios to bypass the global interceptor's strict `data.status === 200` check
    const res = await axios.get(queryUrl, {
      params: { query: computedPromQL.value, start, end, step },
      headers: token ? { Authorization: `Bearer ${token}` } : {}
    })
    const data = res.data

    if (data && data.status === 'success' && data.data.result && data.data.result.length > 0) {
      // Use the first matrix result which holds an array in `.values`
      const result = data.data.result[0]

      if (result.values && result.values.length > 0) {
        times.value = result.values.map(v => {
          const d = new Date(v[0] * 1000)
          return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
        })
        chartData.value = result.values.map(v => parseFloat(v[1]).toFixed(2))
      } else {
        chartData.value = []
        times.value = []
      }
    } else {
      chartData.value = []
      times.value = []
    }
  } catch (err) {
    console.error('Failed to fetch custom metrics:', err)
    error.value = '数据获取失败'
    chartData.value = []
  } finally {
    loading.value = false
  }
}

// 工具函数：为任意颜色（hex/rgb/rgba）添加透明度
const addAlpha = (color, opacity) => {
  if (!color) return `rgba(64, 158, 255, ${opacity})`
  // 处理 rgb / rgba
  if (color.startsWith('rgb')) {
    const raw = color.substring(color.indexOf('(') + 1, color.lastIndexOf(')'))
    const parts = raw.split(',').map(s => s.trim())
    return `rgba(${parts[0]}, ${parts[1]}, ${parts[2]}, ${opacity})`
  }
  // 处理 hex
  let hex = color.replace('#', '')
  if (hex.length === 3) {
    hex = hex.split('').map(char => char + char).join('')
  }
  if (hex.length >= 6) {
    const r = parseInt(hex.substring(0, 2), 16)
    const g = parseInt(hex.substring(2, 4), 16)
    const b = parseInt(hex.substring(4, 6), 16)
    return `rgba(${r}, ${g}, ${b}, ${opacity})`
  }
  return `rgba(64, 158, 255, ${opacity})`
}

const chartOption = computed(() => {
  if (!chartData.value || chartData.value.length === 0) {
    return {}
  }

  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(10, 15, 30, 0.9)',
      borderColor: props.color,
      textStyle: { color: '#e0e0e0' },
      formatter: (params) => {
        if (!params || !params[0]) return ''
        return `${params[0].name}<br/>${params[0].marker} ${params[0].value} ${props.unit}`
      }
    },
    grid: { left: '3%', right: '4%', bottom: '3%', top: '15%', containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: props.chartType === 'bar',
      data: times.value,
      axisLine: { lineStyle: { color: '#2a3142' } },
      axisLabel: { color: '#8a92a6' }
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: '#8a92a6', formatter: `{value} ${props.unit}` },
      splitLine: { lineStyle: { color: '#1a2233', type: 'dashed' } }
    },
    series: [
      {
        name: props.title,
        type: props.chartType,
        smooth: props.chartType === 'line',
        itemStyle: { color: props.color },
        lineStyle: props.chartType === 'line' ? { color: props.color, width: 2 } : undefined,
        areaStyle: props.chartType === 'line' ? {
          color: {
            type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: addAlpha(props.color, 0.25) },
              { offset: 1, color: addAlpha(props.color, 0) }
            ]
          }
        } : undefined,
        data: chartData.value
      }
    ]
  }
})

watch(() => props.nodeName, fetchMetrics)
watch(() => props.podName, fetchMetrics)
watch(() => props.hours, fetchMetrics)
watch(() => props.prometheusInstanceId, fetchMetrics)
watch(computedPromQL, fetchMetrics)

onMounted(() => {
  fetchMetrics()
  refreshInterval = setInterval(fetchMetrics, 30000)
})

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
})
</script>

<style scoped>
.monitor-card {
  background: var(--devops-card-bg);
  border: 1px solid var(--devops-border-color);
  border-radius: 8px;
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 500;
  color: var(--devops-text-primary);
}

.title-icon {
  font-size: 16px;
  color: var(--devops-primary);
}

.card-body {
  height: 250px;
  position: relative;
}

.chart {
  width: 100%;
  height: 100%;
}

.loading-state, .error-state, .empty-state {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--devops-text-secondary);
}

.loading-state .el-icon { font-size: 24px; color: var(--devops-primary); }
.error-state .el-icon { font-size: 24px; color: var(--devops-error); }
.empty-state .el-icon { font-size: 32px; color: var(--devops-text-secondary); opacity: 0.5; }
</style>
