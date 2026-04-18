<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div class="header-left">
          <div class="header-title-wrapper">
             <h2>DevOps 控制台</h2>
             <p class="subtitle">系统概览与核心指标</p>
          </div>
        </div>
        <div class="header-right">
          <el-button type="primary" @click="refreshData">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="mb-20">
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon bg-gradient-blue">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.totalInstances || 0 }}</div>
              <div class="stat-label">总实例数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon bg-gradient-green">
              <el-icon><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.activeInstances || 0 }}</div>
              <div class="stat-label">活跃实例</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon bg-gradient-red">
              <el-icon><CircleClose /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.errorInstances || 0 }}</div>
              <div class="stat-label">错误实例</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon bg-gradient-orange">
              <el-icon><Timer /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.todayTests || 0 }}</div>
              <div class="stat-label">今日测试</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <div class="charts-section">
      <el-row :gutter="20">
        <el-col :span="12">
          <el-card class="content-card">
            <template #header>
              <div class="flex-between">
                <span>实例类型分布</span>
              </div>
            </template>
            <div ref="typeChartRef" class="chart-container"></div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card class="content-card">
            <template #header>
              <div class="flex-between">
                <span>最近连接测试</span>
              </div>
            </template>
            <div class="test-history">
              <div v-if="recentTests.length === 0" class="empty-state">
                <el-empty description="暂无测试记录" />
              </div>
              <div v-else v-for="test in recentTests" :key="test.id" class="test-item">
                <div class="test-status">
                  <el-icon v-if="test.result === 'success'" color="#67C23A"><CircleCheck /></el-icon>
                  <el-icon v-else-if="test.result === 'warning'" color="#E6A23C"><Warning /></el-icon>
                  <el-icon v-else color="#F56C6C"><CircleClose /></el-icon>
                </div>
                <div class="test-info">
                  <div class="test-name">{{ test.instanceName || '未知实例' }}</div>
                  <div class="test-time">{{ formatTime(test.testedAt) }}</div>
                </div>
                <div class="test-response">
                  <span v-if="test.responseTime">{{ test.responseTime }}ms</span>
                  <span v-else class="error-msg">{{ test.errorMessage || '测试失败' }}</span>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 实例状态概览 -->
    <el-card class="content-card mt-20">
      <template #header>
        <div class="flex-between">
          <span>实例状态概览</span>
          <el-button link type="primary" @click="$router.push('/instances')">查看全部</el-button>
        </div>
      </template>
      <div class="instances-grid">
        <div v-if="instances.length === 0" class="empty-state">
          <el-empty description="暂无实例数据" />
        </div>
        <div v-else v-for="instance in instances" :key="instance.id" class="instance-item">
          <div class="instance-name">{{ instance.name || '未命名实例' }}</div>
          <div class="instance-status">
            <el-tag v-if="instance.status === 'active'" type="success" size="small">
              <el-icon><CircleCheck /></el-icon>
              活跃
            </el-tag>
            <el-tag v-else-if="instance.status === 'warning'" type="warning" size="small">
              <el-icon><Warning /></el-icon>
              警告
            </el-tag>
            <el-tag v-else type="danger" size="small">
              <el-icon><CircleClose /></el-icon>
              错误
            </el-tag>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import {nextTick, onMounted, ref} from 'vue'
import {ElMessage} from 'element-plus'
import {CircleCheck, CircleClose, Monitor, Refresh, Timer, Warning} from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import {getInstanceList, getTestHistory, getTodayTestStats} from '@/api/instance.js'
import {createRouter as $router} from "vue-router";

// 统计数据
const stats = ref({
  totalInstances: 0,
  activeInstances: 0,
  errorInstances: 0,
  todayTests: 0
})

// 最近测试记录
const recentTests = ref([])

// 实例列表
const instances = ref([])

// 图表引用
const typeChartRef = ref(null)
let typeChart = null

// 格式化时间
const formatTime = (timeStr) => {
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 初始化类型分布图表
const initTypeChart = () => {
  if (!typeChartRef.value) return

  typeChart = echarts.init(typeChartRef.value)
  
  // 初始化时显示空数据
  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 10,
      textStyle: {
        color: '#606266'
      },
      data: ['暂无数据']
    },
    series: [
      {
        name: '实例类型',
        type: 'pie',
        radius: ['50%', '70%'],
        avoidLabelOverlap: false,
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: '18',
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        data: [
          { value: 1, name: '暂无数据', itemStyle: { color: '#e0e0e0' } }
        ]
      }
    ]
  }

  typeChart.setOption(option)
}

// 刷新数据
const refreshData = async () => {
  ElMessage.info('正在刷新数据...')
  try {
    await fetchDashboardData()
    ElMessage.success('数据刷新成功')
  } catch (error) {
    ElMessage.error('数据刷新失败: ' + error.message)
  }
}

// 获取仪表板数据
const fetchDashboardData = async () => {
  try {
    // 获取实例列表
    const response = await getInstanceList({ page: 1, page_size: 100 })
    const allInstances = response.data?.list?.data || []
    
    // 更新统计数据
    stats.value.totalInstances = allInstances.length
    stats.value.activeInstances = allInstances.filter(i => i.status === 'active').length
    stats.value.errorInstances = allInstances.filter(i => i.status === 'error').length
    
    // 更新实例列表 - 只显示前4个
    instances.value = allInstances.slice(0, 4)
    
    // 获取今日测试数据
    await fetchTodayTestsData()
    
    // 获取最近测试记录
    await fetchRecentTestsData()
    
    // 更新图表数据
    updateTypeChart(allInstances)
    
  } catch (error) {
    console.error('获取仪表板数据失败:', error)
    ElMessage.error('获取数据失败: ' + error.message)
    throw error
  }
}

// 获取今日测试数据
const fetchTodayTestsData = async () => {
  try {
    const response = await getTodayTestStats()
    stats.value.todayTests = response.data?.total_tests || 0
  } catch (error) {
    console.warn('获取今日测试统计失败:', error)
    stats.value.todayTests = 0
  }
}

// 获取最近测试记录
const fetchRecentTestsData = async () => {
  try {
    // 获取所有实例的最近测试记录，这里简化为获取前几个实例的测试记录
    const testPromises = instances.value.map(instance =>
      getTestHistory(instance.id, { page_size: 10 }).catch(() => ({ data: { test_list: [] } }))
    )

    const testResults = await Promise.all(testPromises)
    const allTests = []
    
    testResults.forEach((result, index) => {
      // 修正数据结构解析路径
      const testList = result.data?.tests || result.data?.test_list || []
      // 遍历所有测试记录，而不是只取第一条
      testList.forEach(test => {
        // 转换为前端需要的格式
        const formattedTest = {
          id: test.id,
          instanceName: test.instance?.name || instances.value[index]?.name || '未知实例',
          instanceId: test.instance_id,
          result: test.test_result === 'success' ? 'success' : 
                 test.test_result === 'timeout' ? 'warning' : 'error',
          responseTime: test.response_time || 0,
          errorMessage: test.error_message || '',
          testedAt: test.tested_at
        }
        allTests.push(formattedTest)
      })
    })
    
    // 按测试时间排序，最新的在前
    allTests.sort((a, b) => new Date(b.testedAt) - new Date(a.testedAt))
    recentTests.value = allTests.slice(0, 8) // 只显示前8条
    
  } catch (error) {
    console.warn('获取最近测试记录失败:', error)
    recentTests.value = []
  }
}

// 更新类型分布图表
const updateTypeChart = (allInstances) => {
  if (!typeChart) return
  
  // 统计各类型实例数量
  const typeCount = {}
  const typeColors = {
    'elasticsearch': '#1976d2',
    'kibana': '#4caf50',
    'logstash': '#ff9800',
    'filebeat': '#f44336',
    'metricbeat': '#9c27b0',
    'apm': '#795548',
    'prometheus': '#e6522c',
    'beats': '#607d8b',
    'other': '#9e9e9e'
  }
  
  allInstances.forEach(instance => {
    // 使用正确的字段名 instance_type
    const type = instance.instance_type || 'other'
    typeCount[type] = (typeCount[type] || 0) + 1
  })
  
  // 转换为图表数据格式
  const chartData = Object.keys(typeCount).map(type => ({
    value: typeCount[type],
    name: type.charAt(0).toUpperCase() + type.slice(1),
    itemStyle: { color: typeColors[type] || typeColors.other }
  }))
  
  // 如果没有数据，显示空数据提示
  if (chartData.length === 0) {
    chartData.push({
      value: 1,
      name: '暂无数据',
      itemStyle: { color: '#e0e0e0' }
    })
  }
  
  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 10,
      textStyle: {
        color: '#606266'
      },
      data: chartData.map(item => item.name)
    },
    series: [
      {
        name: '实例类型',
        type: 'pie',
        radius: ['50%', '70%'],
        avoidLabelOverlap: false,
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: '18',
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        data: chartData
      }
    ]
  }
  
  typeChart.setOption(option)
}

onMounted(async () => {
  await nextTick()
  initTypeChart()
  await fetchDashboardData()
  
  // 监听窗口大小变化
  window.addEventListener('resize', () => {
    if (typeChart) {
      typeChart.resize()
    }
  })
})
</script>

<style scoped>
/* Only specific overrides not covered by global styles */
.chart-container {
  height: 320px;
}

.test-history {
  height: 320px;
  overflow-y: auto;
}

.test-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border-bottom: 1px solid var(--border-color);
  transition: background-color 0.2s;
}

.test-item:last-child {
  border-bottom: none;
}

.test-item:hover {
  background-color: #f5f7fa;
}

.test-status {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  font-size: 18px;
}

.test-info {
  flex: 1;
}

.test-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-main);
}

.test-time {
  font-size: 12px;
  color: var(--text-sub);
}

.test-response {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main);
}

.error-msg {
  color: var(--danger-color);
}

.instances-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
}

.instance-item {
  padding: 20px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: #fff;
  transition: all 0.3s;
}

.instance-item:hover {
  box-shadow: var(--shadow-hover);
  transform: translateY(-2px);
}

.instance-name {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
  color: var(--text-main);
}

.instance-status {
  display: flex;
  justify-content: flex-start;
}

.empty-state {
  padding: 40px;
  text-align: center;
}
</style>