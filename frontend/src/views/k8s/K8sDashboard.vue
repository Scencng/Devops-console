<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div class="header-left">
          <div class="header-icon-wrapper">
            <el-icon :size="24"><Monitor /></el-icon>
          </div>
          <div class="header-title-wrapper">
             <h2>Kubernetes 控制台</h2>
             <p class="subtitle">系统概览与集群指标</p>
          </div>
        </div>
        <div class="header-right">
          <el-button type="primary" :loading="refreshing" @click="refreshData">
            <el-icon><Refresh /></el-icon>
            刷新数据
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- Metrics Grid -->
    <el-row :gutter="20" class="mb-20">
      <el-col :span="6" v-for="(metric, index) in [
        { label: '总节点数', value: clusterMetrics.totalNodes, sub: `${clusterMetrics.readyNodes} 就绪`, icon: 'Monitor', color: 'bg-gradient-blue' },
        { label: '活跃工作负载', value: clusterMetrics.totalPods, sub: `${clusterMetrics.totalPods} Pods 运行中`, icon: 'Box', color: 'bg-gradient-green' },
        { label: 'CPU 分配', value: clusterMetrics.cpuUsage + '%', sub: `${clusterMetrics.cpuAvailable} / ${clusterMetrics.cpuTotal} 核`, icon: 'Cpu', color: 'bg-gradient-orange' },
        { label: '内存分配', value: clusterMetrics.memoryUsage + '%', sub: `${clusterMetrics.memoryAvailable} / ${clusterMetrics.memoryTotal} Mi`, icon: 'Memo', color: 'bg-gradient-purple' }
      ]" :key="index">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" :class="metric.color">
              <component :is="metric.icon" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ metric.value }}</div>
              <div class="stat-label">{{ metric.label }}</div>
              <div class="stat-sub">{{ metric.sub }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="main-content-row">
      <!-- Left Column -->
      <el-col :span="16">
        <!-- Cluster Overview -->
        <el-card class="content-card mb-20">
          <template #header>
            <div class="flex-between">
              <span>集群概览</span>
              <el-button link type="primary" @click="navigateTo('/k8s/cluster')">管理集群</el-button>
            </div>
          </template>
          <div class="cluster-list-container">
            <el-empty v-if="clusterList.length === 0" description="未连接集群">
              <el-button type="primary" @click="navigateTo('/k8s/cluster')">初始化集群</el-button>
            </el-empty>
            <div v-else class="cluster-grid">
              <div v-for="cluster in clusterList" :key="cluster.id" class="cluster-item" @click="viewClusterDetail(cluster)">
                <div class="cluster-status" :class="cluster.status === 'active' ? 'active' : 'inactive'"></div>
                <div class="cluster-info">
                  <div class="cluster-name">{{ cluster.cluster_name }}</div>
                  <div class="cluster-meta">最近同步: {{ cluster.last_sync || '未知' }}</div>
                </div>
                <div class="cluster-actions">
                   <el-button link type="primary" @click.stop="handleTestConnection(cluster)">测试连接</el-button>
                   <el-button link type="primary" @click.stop="handleChangeCluster(cluster)">切换</el-button>
                </div>
              </div>
            </div>
          </div>
        </el-card>

        <!-- Workload Stats -->
        <el-card class="content-card">
          <template #header>
            <div class="flex-between">
              <span>工作负载统计</span>
            </div>
          </template>
          <div class="workload-grid">
            <div class="workload-item" v-for="(stat, key) in {
              'Deployments': workloadStats.deployments,
              'StatefulSets': workloadStats.statefulSets,
              'DaemonSets': workloadStats.daemonSets,
              'Jobs': workloadStats.jobs,
              'Total Pods': workloadStats.totalPods,
              'Running': workloadStats.runningPods
            }" :key="key">
              <div class="workload-value">{{ stat }}</div>
              <div class="workload-label">{{ key }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- Right Column -->
      <el-col :span="8">
        <!-- Quick Actions -->
        <el-card class="content-card mb-20">
          <template #header>
            <span>快捷操作</span>
          </template>
          <div class="quick-actions-grid">
             <div class="action-btn" @click="navigateTo('/k8s/cluster')">
               <el-icon class="action-btn-icon bg-blue-100 text-blue-500"><Setting /></el-icon>
               <span>Cluster</span>
             </div>
             <div class="action-btn" @click="navigateTo('/k8s/pod')">
               <el-icon class="action-btn-icon bg-green-100 text-green-500"><Box /></el-icon>
               <span>Pods</span>
             </div>
             <div class="action-btn" @click="navigateTo('/k8s/deployment')">
               <el-icon class="action-btn-icon bg-orange-100 text-orange-500"><Files /></el-icon>
               <span>Deploy</span>
             </div>
             <div class="action-btn" @click="navigateTo('/k8s/namespace')">
               <el-icon class="action-btn-icon bg-purple-100 text-purple-500"><Folder /></el-icon>
               <span>Namespace</span>
             </div>
             <div class="action-btn" @click="navigateTo('/k8s/service')">
               <el-icon class="action-btn-icon bg-cyan-100 text-cyan-500"><Connection /></el-icon>
               <span>Service</span>
             </div>
             <div class="action-btn" @click="navigateTo('/k8s/cronjob')">
               <el-icon class="action-btn-icon bg-pink-100 text-pink-500"><Clock /></el-icon>
               <span>CronJob</span>
             </div>
          </div>
        </el-card>

        <!-- Resource Health -->
        <el-card class="content-card mb-20">
          <template #header>
            <span>资源健康状况</span>
          </template>
          <div class="health-list">
            <div class="health-item">
               <div class="health-icon success"><el-icon><Check /></el-icon></div>
               <div class="health-text">健康节点</div>
               <div class="health-val">{{ resourceStatus.healthyNodes }} / {{ resourceStatus.totalNodes }}</div>
            </div>
            <div class="health-item" :class="{ 'warning': resourceStatus.warningPods > 0 }">
               <div class="health-icon warning"><el-icon><Warning /></el-icon></div>
               <div class="health-text">警告 Pods</div>
               <div class="health-val">{{ resourceStatus.warningPods }}</div>
            </div>
            <div class="health-item" :class="{ 'error': resourceStatus.failedPods > 0 }">
               <div class="health-icon error"><el-icon><Close /></el-icon></div>
               <div class="health-text">失败 Pods</div>
               <div class="health-val">{{ resourceStatus.failedPods }}</div>
            </div>
          </div>
        </el-card>

        <!-- Storage Overview -->
        <el-card class="content-card">
           <template #header>
              <span>存储概览</span>
            </template>
            <div class="storage-list">
               <div class="storage-item" v-for="(item, index) in [
                 { label: 'PV 总数', value: storageInfo.totalPV, icon: 'Document' },
                 { label: 'PVC 总数', value: storageInfo.totalPVC, icon: 'Folder' },
                 { label: 'Storage Classes', value: storageInfo.storageClasses, icon: 'Coin' },
                 { label: '已用存储', value: storageInfo.usedStorage + ' Gi', icon: 'PieChart' }
               ]" :key="index">
                 <div class="storage-icon-wrapper">
                    <el-icon><component :is="item.icon" /></el-icon>
                 </div>
                 <div class="storage-info">
                   <div class="storage-val">{{ item.value }}</div>
                   <div class="storage-lbl">{{ item.label }}</div>
                 </div>
               </div>
            </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Monitor, Box, Files, Connection, Setting, Folder, Refresh, Cpu, Memo,
  DataAnalysis, Collection, Operation, Timer, VideoPlay, Menu, Clock,
  FolderOpened, Document, Coin, PieChart, View, RefreshRight, Check, Warning, Close
} from '@element-plus/icons-vue'
import { getClusterList, changeCluster, testClusterConnection } from '@/api/k8s/cluster'
import { getClusterMetrics, getClusterInfo } from '@/api/k8s/cluster-info'
import { getSelectedInstanceId } from '@/stores/instanceStore'
import eventBus from '@/utils/eventBus'

const router = useRouter()
const refreshing = ref(false)

// 集群指标数据
const clusterMetrics = ref({
  totalNodes: 0,
  readyNodes: 0,
  totalPods: 0,
  cpuUsage: 0,
  cpuAvailable: 0,
  cpuTotal: 0,
  memoryUsage: 0,
  memoryAvailable: 0,
  memoryTotal: 0
})

// 工作负载统计
const workloadStats = ref({
  deployments: 0,
  statefulSets: 0,
  daemonSets: 0,
  jobs: 0,
  totalPods: 0,
  runningPods: 0
})

// 资源状态
const resourceStatus = ref({
  totalNodes: 0,
  healthyNodes: 0,
  warningPods: 0,
  failedPods: 0,
  pendingPods: 0
})

// 存储信息
const storageInfo = ref({
  totalPV: 0,
  totalPVC: 0,
  storageClasses: 0,
  usedStorage: 0
})

const clusterList = ref([])

// 集群信息
const clusterInfo = ref({})

const navigateTo = (path) => {
  router.push(path)
}

const viewClusterDetail = (cluster) => {
  router.push(`/k8s/cluster/${cluster.cluster_name}`)
}

const refreshData = async () => {
  refreshing.value = true
  try {
    await fetchDashboardData()
    ElMessage.success('数据刷新成功')
  } catch (error) {
    ElMessage.error('刷新失败: ' + (error.message || '未知错误'))
  } finally {
    refreshing.value = false
  }
}

const handleChangeCluster = async (cluster) => {
  try {
    await changeCluster(cluster.cluster_name)
    ElMessage.success(`已切换到集群: ${cluster.cluster_name}`)
    await fetchDashboardData()
  } catch (error) {
    ElMessage.error('切换集群失败: ' + (error.response?.data?.message || error.message || '未知错误'))
  }
}

const handleTestConnection = async (cluster) => {
  try {
    await testClusterConnection(cluster.cluster_name)
    ElMessage.success(`集群 ${cluster.cluster_name} 连接正常`)
  } catch (error) {
    ElMessage.error('连接测试失败: ' + (error.response?.data?.message || error.message || '未知错误'))
  }
}

const fetchDashboardData = async () => {
  try {
    const instanceId = getSelectedInstanceId() || '1'
    const response = await getClusterList(instanceId)
    clusterList.value = response.data?.clasters || []

    const infoResponse = await getClusterInfo(instanceId)
    clusterInfo.value = infoResponse.data?.clusterInfo || {}

    const metricsResponse = await getClusterMetrics(instanceId)
    const metrics = metricsResponse.data?.metrics
    
    if (metrics) {
      clusterMetrics.value = {
        totalNodes: metrics.totalNodes,
        readyNodes: metrics.readyNodes,
        totalPods: metrics.totalPods,
        cpuUsage: metrics.cpuUsage,
        cpuAvailable: metrics.cpuAvailable,
        cpuTotal: metrics.cpuTotal,
        memoryUsage: metrics.memoryUsage,
        memoryAvailable: metrics.memoryAvailable,
        memoryTotal: metrics.memoryTotal
      }

      workloadStats.value = metrics.workloadStats
      storageInfo.value = metrics.storageInfo

      resourceStatus.value = {
        totalNodes: metrics.totalNodes,
        healthyNodes: metrics.readyNodes,
        warningPods: metrics.workloadStats.unknownPods || 0,
        failedPods: metrics.workloadStats.failedPods || 0,
        pendingPods: metrics.workloadStats.pendingPods || 0
      }
    }

  } catch (error) {
    console.error('获取仪表板数据失败:', error)
    ElMessage.error('获取数据失败: ' + (error.message || '未知错误'))
  }
}

const handlePodEvent = () => {
  fetchDashboardData()
}

onMounted(() => {
  fetchDashboardData()
  eventBus.on('pod:created', handlePodEvent)
  eventBus.on('pod:deleted', handlePodEvent)
})

onUnmounted(() => {
  eventBus.off('pod:created', handlePodEvent)
  eventBus.off('pod:deleted', handlePodEvent)
})
</script>

<style scoped>
.header-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: var(--primary-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-sub {
  font-size: 12px;
  color: var(--text-sub);
  margin-top: 4px;
}

/* Cluster List */
.cluster-list-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.cluster-item {
  display: flex;
  align-items: center;
  padding: 16px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: #f8f9fa;
  cursor: pointer;
  transition: all 0.2s;
}

.cluster-item:hover {
  background: #fff;
  box-shadow: var(--shadow-sm);
  border-color: var(--primary-color);
}

.cluster-status {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  margin-right: 12px;
}

.cluster-status.active { background-color: var(--success-color); box-shadow: 0 0 4px var(--success-color); }
.cluster-status.inactive { background-color: var(--info-color); }

.cluster-info {
  flex: 1;
}

.cluster-name {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-main);
}

.cluster-meta {
  font-size: 12px;
  color: var(--text-sub);
  margin-top: 2px;
}

/* Workload Stats */
.workload-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.workload-item {
  background: #f8f9fa;
  padding: 16px;
  border-radius: var(--radius-md);
  text-align: center;
  border: 1px solid transparent;
  transition: all 0.2s;
}

.workload-item:hover {
  background: #fff;
  border-color: var(--border-color);
  box-shadow: var(--shadow-sm);
}

.workload-value {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-main);
  margin-bottom: 4px;
}

.workload-label {
  font-size: 12px;
  color: var(--text-sub);
}

/* Quick Actions */
.quick-actions-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: background-color 0.2s;
}

.action-btn:hover {
  background-color: #f8f9fa;
}

.action-btn-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.action-btn span {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-main);
}

/* Colors for actions */
.bg-blue-100 { background-color: #e3f2fd; }
.text-blue-500 { color: #2196f3; }
.bg-green-100 { background-color: #e8f5e9; }
.text-green-500 { color: #4caf50; }
.bg-orange-100 { background-color: #fff3e0; }
.text-orange-500 { color: #ff9800; }
.bg-purple-100 { background-color: #f3e5f5; }
.text-purple-500 { color: #9c27b0; }
.bg-cyan-100 { background-color: #e0f7fa; }
.text-cyan-500 { color: #00bcd4; }
.bg-pink-100 { background-color: #fce4ec; }
.text-pink-500 { color: #e91e63; }

/* Health List */
.health-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.health-item {
  display: flex;
  align-items: center;
  padding: 12px;
  background: #f8f9fa;
  border-radius: var(--radius-md);
}

.health-icon {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
}

.health-icon.success { background-color: #e8f5e9; color: #4caf50; }
.health-icon.warning { background-color: #fff3e0; color: #ff9800; }
.health-icon.error { background-color: #ffebee; color: #f44336; }

.health-text {
  flex: 1;
  font-size: 13px;
  color: var(--text-main);
  font-weight: 500;
}

.health-val {
  font-weight: 600;
  color: var(--text-main);
}

/* Storage List */
.storage-list {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.storage-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: var(--radius-md);
}

.storage-icon-wrapper {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary-color);
  box-shadow: var(--shadow-sm);
}

.storage-info {
  display: flex;
  flex-direction: column;
}

.storage-val {
  font-weight: 600;
  color: var(--text-main);
  font-size: 14px;
}

.storage-lbl {
  font-size: 11px;
  color: var(--text-sub);
}
</style>
