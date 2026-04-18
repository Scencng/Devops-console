<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>趋势分析</h2>
          <p>回看 Lag、Broker、未同步分区和吞吐历史趋势，辅助排障与容量规划</p>
        </div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-select v-model="preset" style="width:160px">
            <el-option label="默认趋势" value="default" />
            <el-option label="吞吐趋势" value="throughput" />
            <el-option label="Lag 趋势" value="lag" />
          </el-select>
          <el-button :loading="loading" @click="loadData">刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-card v-for="item in trendSeries" :key="item.name" class="content-card" v-loading="loading">
      <template #header><div class="card-header">{{ item.name }}</div></template>
      <div :ref="setChartRef(item.name)" class="chart-box" />
      <div class="query-note">PromQL: {{ item.query }}</div>
    </el-card>
  </div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import * as echarts from 'echarts'
import { ElMessage } from 'element-plus'
import { getKafkaClusterOptions, getKafkaTrendSeries } from '@/api/kafka.js'

const clusters = ref([])
const selectedClusterId = ref(null)
const preset = ref('default')
const loading = ref(false)
const trendSeries = ref([])
const chartRefs = new Map()
const charts = new Map()

const setChartRef = (key) => (el) => {
  if (el) chartRefs.set(key, el)
}

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id
}

const renderCharts = async () => {
  await nextTick()
  trendSeries.value.forEach((item) => {
    const el = chartRefs.get(item.name)
    if (!el) return
    let chart = charts.get(item.name)
    if (!chart) {
      chart = echarts.init(el)
      charts.set(item.name, chart)
    }
    const first = item.series?.[0]
    const xAxis = (first?.points || []).map((point) => new Date(point.timestamp * 1000).toLocaleTimeString())
    const series = (item.series || []).map((seriesItem) => ({
      name: seriesItem.metric,
      type: 'line',
      smooth: true,
      data: (seriesItem.points || []).map((point) => point.value),
    }))
    chart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { type: 'scroll' },
      xAxis: { type: 'category', data: xAxis },
      yAxis: { type: 'value' },
      series,
    })
  })
}

const loadData = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const res = await getKafkaTrendSeries({ clusterId: selectedClusterId.value, preset: preset.value })
    trendSeries.value = res?.data?.data || []
    await renderCharts()
  } catch (error) {
    ElMessage.error(error.message || '趋势分析加载失败')
  } finally {
    loading.value = false
  }
}

const handleResize = () => {
  charts.forEach((chart) => chart.resize())
}

onMounted(async () => {
  try {
    await loadClusters()
    await loadData()
    window.addEventListener('resize', handleResize)
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 趋势分析初始化失败')
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  charts.forEach((chart) => chart.dispose())
  charts.clear()
})
</script>

<style scoped>
.chart-box { height: 340px; }
.query-note { margin-top: 12px; color:#909399; font-size:12px; word-break:break-all; }
</style>
