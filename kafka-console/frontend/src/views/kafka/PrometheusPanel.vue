<template>
  <div class="page-container">
    <el-card class="page-header-card"><div class="page-header"><div><h2>Prometheus 面板</h2><p>查看 Kafka Exporter 指标，并支持自定义 PromQL 查询</p></div><div class="header-actions"><el-button @click="loadPanels" :loading="loading">刷新</el-button></div></div></el-card>
    <el-row :gutter="20" class="stats-row" v-loading="loading"><el-col :span="6" v-for="card in panel.cards || []" :key="card.name"><el-card><div class="stat"><span>{{ card.name }}</span><strong>{{ card.value }}</strong></div></el-card></el-col></el-row>
    <el-card class="content-card" v-loading="loading"><template #header><div class="card-header">消费组 Lag 趋势</div></template><div ref="chartRef" style="height:360px"></div></el-card>
    <el-card class="content-card">
      <template #header><div class="card-header">自定义 PromQL</div></template>
      <el-form label-position="top">
        <el-form-item label="PromQL"><el-input v-model="queryForm.query" type="textarea" :rows="3" /></el-form-item>
        <el-row :gutter="20">
          <el-col :span="6"><el-form-item label="开始时间"><el-date-picker v-model="queryForm.start" type="datetime" value-format="X" placeholder="选择开始时间" style="width:100%" /></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="结束时间"><el-date-picker v-model="queryForm.end" type="datetime" value-format="X" placeholder="选择结束时间" style="width:100%" /></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="步长"><el-input v-model="queryForm.step" placeholder="例如 1m" /></el-form-item></el-col>
          <el-col :span="6"><el-form-item label="快捷范围"><div class="range-actions"><el-button @click="setRelativeRange(900)">最近 15 分钟</el-button><el-button @click="setRelativeRange(3600)">最近 1 小时</el-button><el-button @click="setRelativeRange(21600)">最近 6 小时</el-button></div></el-form-item></el-col>
        </el-row>
        <el-form-item><el-button type="primary" @click="runQuery">执行查询</el-button></el-form-item>
      </el-form>
      <pre class="query-result">{{ queryResult }}</pre>
    </el-card>
  </div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { getKafkaPrometheusPanels, queryKafkaPrometheusRange } from '@/api/kafka.js'
const loading = ref(false)
const panel = ref({ cards: [], lagSeries: [] })
const queryResult = ref('点击“执行查询”后显示 Prometheus 返回结果')
const chartRef = ref(null)
let chart
const queryForm = reactive({ query: 'sum(kafka_consumergroup_lag)', start: '', end: '', step: '1m' })
const setRelativeRange = (seconds) => {
  const end = Math.floor(Date.now() / 1000)
  queryForm.end = String(end)
  queryForm.start = String(end - seconds)
}
const initChart = () => { if (!chartRef.value) return; chart = echarts.init(chartRef.value) }
const renderChart = () => { if (!chart) return; const first = panel.value.lagSeries?.[0]; const xAxis = (first?.points || []).map(item => new Date(item.timestamp * 1000).toLocaleTimeString()); const series = (panel.value.lagSeries || []).map(item => ({ name: item.metric, type: 'line', smooth: true, data: item.points.map(point => point.value) })); chart.setOption({ tooltip: { trigger: 'axis' }, legend: { type: 'scroll' }, xAxis: { type: 'category', data: xAxis }, yAxis: { type: 'value' }, series }) }
const resizeChart = () => { if (chart) chart.resize() }
const loadPanels = async () => { loading.value = true; try { const res = await getKafkaPrometheusPanels(); panel.value = res?.data?.data || panel.value; await nextTick(); renderChart() } catch (error) { ElMessage.error(error.message || 'Prometheus 面板加载失败') } finally { loading.value = false } }
const runQuery = async () => {
  if (!queryForm.query.trim()) {
    ElMessage.warning('请输入 PromQL')
    return
  }
  if (!queryForm.start || !queryForm.end) {
    setRelativeRange(3600)
  }
  try {
    const res = await queryKafkaPrometheusRange(queryForm)
    queryResult.value = JSON.stringify(res?.data?.data || {}, null, 2)
  } catch (error) {
    ElMessage.error(error.message || 'PromQL 查询失败')
  }
}
onMounted(async () => {
  setRelativeRange(3600)
  initChart()
  window.addEventListener('resize', resizeChart)
  await loadPanels()
})
onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeChart)
  if (chart) {
    chart.dispose()
    chart = null
  }
})
</script>

<style scoped>
.stats-row { margin: 20px 0; }
.stat { display:flex; flex-direction:column; gap:10px; }
.stat span { color:#909399; }
.stat strong { font-size:28px; font-weight:700; }
.range-actions { display:flex; gap:8px; flex-wrap:wrap; }
.query-result { white-space: pre-wrap; word-break: break-word; }
</style>
