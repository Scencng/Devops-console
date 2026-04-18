<template>
  <div class="node-monitor-chart">
    <div v-if="loading" class="chart-loading">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span class="loading-text">加载监控数据中...</span>
    </div>
    <div v-else-if="emptyData" class="chart-empty">
      <el-empty description="未获取到相关的监控数据，请检查 Prometheus 采集配置或查询条件是否匹配当前节点/Pod" :image-size="60" />
    </div>
    <div v-else-if="error" class="chart-error">
      <el-icon><Warning /></el-icon>
      <span class="error-text">{{ error }}</span>
    </div>
    <div v-else class="chart-container" ref="chartRef"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'
import { Warning, Loading } from '@element-plus/icons-vue'
import axios from 'axios'
import request from '@/api/index.js'
import dayjs from 'dayjs'

const props = defineProps({
  nodeName: {
    type: String,
    required: true
  },
  prometheusInstanceId: {
    type: String,
    required: true
  },
  title: {
    type: String,
    default: '监控图表'
  },
  metricType: {
    type: String,
    required: true,
    validator: (value) => ['cpu', 'memory', 'network', 'disk'].includes(value)
  },
  // 查询时间范围(小时)
  hours: {
    type: Number,
    default: 1
  }
})

const chartRef = ref(null)
const loading = ref(false)
const error = ref('')
const emptyData = ref(false)
let chartInstance = null
let refreshTimer = null

// 生成 PromQL 查询语句
const getPromQL = () => {
  const node = props.nodeName
  switch (props.metricType) {
    case 'cpu':
      // CPU 使用率 (%) = 100 - (空闲时间 rate * 100)
      return `100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle", instance=~"${node}.*"}[5m])) * 100)`
    case 'memory':
      // 内存使用率 (%) = (总内存 - 可用内存) / 总内存 * 100
      return `(node_memory_MemTotal_bytes{instance=~"${node}.*"} - node_memory_MemAvailable_bytes{instance=~"${node}.*"}) / node_memory_MemTotal_bytes{instance=~"${node}.*"} * 100`
    case 'network':
      // 网络接收速率 (bytes/sec) - 这里只统计 eth0/ens 相关的网卡
      return `rate(node_network_receive_bytes_total{instance=~"${node}.*", device=~"eth.*|ens.*"}[5m])`
    case 'disk':
      // 根目录使用率 (%)
      return `100 - (node_filesystem_avail_bytes{instance=~"${node}.*", mountpoint="/"} / node_filesystem_size_bytes{instance=~"${node}.*", mountpoint="/"} * 100)`
    default:
      return ''
  }
}

// 格式化 Y 轴单位
const formatYAxis = (value) => {
  if (props.metricType === 'network') {
    // 转换为KB/MB
    if (value > 1024 * 1024) return (value / (1024 * 1024)).toFixed(1) + ' MB/s'
    if (value > 1024) return (value / 1024).toFixed(1) + ' KB/s'
    return value.toFixed(1) + ' B/s'
  }
  return value.toFixed(1) + ' %'
}

// 格式化 Tooltip 值
const formatTooltipValue = (value) => {
  if (props.metricType === 'network') {
    if (value > 1024 * 1024) return (value / (1024 * 1024)).toFixed(2) + ' MB/s'
    if (value > 1024) return (value / 1024).toFixed(2) + ' KB/s'
    return value.toFixed(2) + ' B/s'
  }
  return value.toFixed(2) + ' %'
}

// 获取 ECharts 颜色主题 (赛博朋克深色/科技感)
const getChartColors = () => {
  const colors = {
    cpu: ['#409eff', 'rgba(64, 158, 255, 0.2)'],     // Blue
    memory: ['#67c23a', 'rgba(103, 194, 58, 0.2)'],  // Green
    network: ['#e6a23c', 'rgba(230, 162, 60, 0.2)'], // Warning
    disk: ['#f56c6c', 'rgba(245, 108, 108, 0.2)']    // Danger
  }
  return colors[props.metricType] || colors.cpu
}

const fetchMetrics = async () => {
  if (!props.nodeName || !props.prometheusInstanceId) return

  loading.value = true
  error.value = ''
  emptyData.value = false
  
  try {
    const end = Math.floor(Date.now() / 1000)
    const start = end - (props.hours * 60 * 60)
    // 根据查询跨度动态调整步长
    const step = props.hours <= 1 ? 60 : (props.hours <= 6 ? 300 : 1800)
    const query = getPromQL()

    // 发请求到 Go 后端的代理总线
    const res = await request({
      url: `/prometheus/${props.prometheusInstanceId}/api/v1/query_range`,
      method: 'get',
      params: { query, start, end, step }
    })

    // 根据 Prometheus 的标准响应解析
    if (res.data?.status === 'success') {
       if (res.data?.data?.result?.length > 0) {
         renderChart(res.data.data.result[0].values)
       } else {
         emptyData.value = true
       }
    } else {
      error.value = '暂无监控数据'
    }
  } catch (err) {
    error.value = '获取监控数据失败'
  } finally {
    loading.value = false
  }
}

const renderChart = (values) => {
  if (!chartRef.value) return
  if (!chartInstance) {
    chartInstance = echarts.init(chartRef.value)
  }

  // values 是 [ [timestamp, "valueStr"], ... ]
  const xData = values.map(v => dayjs(v[0] * 1000).format('HH:mm'))
  const yData = values.map(v => parseFloat(v[1]))
  const [lineColor, areaColor] = getChartColors()

  const option = {
    title: {
      text: props.title,
      textStyle: { fontSize: 13, color: '#606266', fontWeight: '500' },
      left: 0,
      top: 0
    },
    tooltip: {
      trigger: 'axis',
      formatter: function (params) {
        const p = params[0]
        return `${p.name}<br/><span style="display:inline-block;margin-right:4px;border-radius:10px;width:10px;height:10px;background-color:${lineColor};"></span>${formatTooltipValue(p.value)}`
      }
    },
    grid: { left: '3%', right: '4%', bottom: '3%', top: '40px', containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: xData,
      axisLine: { lineStyle: { color: '#E4E7ED' } },
      axisLabel: { color: '#909399', fontSize: 11 }
    },
    yAxis: {
      type: 'value',
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: { lineStyle: { type: 'dashed', color: '#EBEEF5' } },
      axisLabel: { color: '#909399', fontSize: 11, formatter: formatYAxis },
      min: (value) => props.metricType === 'network' ? 0 : Math.max(0, Math.floor(value.min - 5)),
      max: (value) => props.metricType === 'network' ? null : Math.min(100, Math.ceil(value.max + 5))
    },
    series: [
      {
        name: props.title,
        type: 'line',
        smooth: true,
        symbol: 'none',
        sampling: 'average',
        itemStyle: { color: lineColor },
        lineStyle: { width: 2 },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: areaColor },
            { offset: 1, color: 'rgba(255, 255, 255, 0)' }
          ])
        },
        data: yData
      }
    ]
  }

  chartInstance.setOption(option)
}

const resizeChart = () => {
  if (chartInstance) {
    chartInstance.resize()
  }
}

watch(() => props.nodeName, fetchMetrics)
watch(() => props.podName, fetchMetrics)
watch(() => props.hours, fetchMetrics)
watch(() => props.prometheusInstanceId, fetchMetrics)

onMounted(() => {
  fetchMetrics()
  window.addEventListener('resize', resizeChart)
  
  // 30秒自动刷新一次
  refreshTimer = setInterval(fetchMetrics, 30000)
})

onUnmounted(() => {
  window.removeEventListener('resize', resizeChart)
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  if (chartInstance) {
    chartInstance.dispose()
  }
})
</script>

<style scoped>
.node-monitor-chart {
  width: 100%;
  height: 100%;
  min-height: 250px;
  position: relative;
  border: 1px solid var(--border-color);
  border-radius: 4px;
}

.chart-container {
  width: 100%;
  height: 100%;
  min-height: 250px;
}

.chart-loading, .chart-error, .chart-empty {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background-color: #fcfcfc;
  z-index: 10;
  border-radius: 4px;
}

.loading-text, .error-text {
  margin-top: 10px;
  font-size: 13px;
  color: #909399;
}

.chart-error .el-icon {
  font-size: 24px;
  color: #f56c6c;
}
</style>
