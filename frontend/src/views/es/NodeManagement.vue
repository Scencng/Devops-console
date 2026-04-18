<template>
  <div class="page-container">
    <!-- Header -->
    <el-card class="page-header-card">
       <div class="page-header">
         <div class="header-left">
           <div class="header-icon-wrapper">
             <el-icon :size="24"><Connection /></el-icon>
           </div>
           <div class="header-title-wrapper">
             <h2>节点管理</h2>
             <p class="subtitle">监控与管理 Elasticsearch 集群节点</p>
           </div>
         </div>
         <div class="header-right">
            <el-input
              v-model="searchQuery"
              placeholder="搜索节点名称"
              class="w-64 mr-2"
              @input="handleSearch"
              :prefix-icon="Search"
            />
            <el-button @click="refreshData" :icon="Refresh">刷新</el-button>
         </div>
       </div>
    </el-card>

    <!-- Stats Cards -->
    <el-row :gutter="20" class="mb-20">
       <el-col :span="6">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-content">
               <div class="stat-icon bg-gradient-blue"><el-icon><Connection /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ stats.totalNodes }}</div>
                  <div class="stat-label">总节点数</div>
               </div>
            </div>
          </el-card>
       </el-col>
       <el-col :span="6">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-content">
               <div class="stat-icon bg-gradient-green"><el-icon><CircleCheck /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ stats.activeNodes }}</div>
                  <div class="stat-label">活跃节点</div>
               </div>
            </div>
          </el-card>
       </el-col>
       <el-col :span="6">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-content">
               <div class="stat-icon bg-gradient-orange"><el-icon><Warning /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ stats.warningNodes }}</div>
                  <div class="stat-label">警告节点</div>
               </div>
            </div>
          </el-card>
       </el-col>
       <el-col :span="6">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-content">
               <div class="stat-icon bg-gradient-red"><el-icon><CircleClose /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ stats.failedNodes }}</div>
                  <div class="stat-label">故障节点</div>
               </div>
            </div>
          </el-card>
       </el-col>
    </el-row>

    <!-- Table Content -->
    <el-card class="content-card">
      <el-table
        :data="filteredNodes"
        v-loading="loading"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="name" label="节点名称" min-width="180">
          <template #default="{ row }">
            <div class="node-name">
              <div>
                <el-tag :type="getStatusType(row.status)" size="small" effect="plain" class="mr-2">
                  {{ row.status }}
                </el-tag>
                <span class="font-medium">{{ row.name }}</span>
              </div>
              <div class="status-reason" v-if="row.statusReason && row.statusReason !== '运行正常'">
                <el-tooltip :content="row.statusReason" placement="top">
                  <span class="text-xs text-red-500">{{ row.statusReason }}</span>
                </el-tooltip>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP地址" width="140" align="right" />
        <el-table-column prop="port" label="端口" width="100" align="right" />
        <el-table-column prop="version" label="版本" width="120" align="right" />
        <el-table-column prop="cpuUsage" label="CPU使用率" width="160">
          <template #default="{ row }">
            <div class="flex items-center gap-2">
              <el-progress
                :percentage="row.cpuUsage"
                :color="getProgressColor(row.cpuUsage)"
                :stroke-width="6"
                class="flex-1"
              />
              <span class="text-xs w-12 text-right" v-if="row.loadAverage">L: {{ row.loadAverage.toFixed(2) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="memoryUsage" label="内存使用" width="160">
          <template #default="{ row }">
            <div class="flex flex-col">
              <el-progress
                :percentage="row.memoryUsage"
                :color="getProgressColor(row.memoryUsage)"
                :stroke-width="6"
              />
              <span class="text-xs text-secondary mt-1" v-if="row.totalMemory">
                {{ formatBytes(row.totalMemory - row.availableMemory) }} / {{ formatBytes(row.totalMemory) }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="diskUsage" label="磁盘使用" width="160">
          <template #default="{ row }">
             <div class="flex flex-col">
              <el-progress
                :percentage="row.diskUsage"
                :color="getProgressColor(row.diskUsage)"
                :stroke-width="6"
              />
              <span class="text-xs text-secondary mt-1" v-if="row.totalDisk">
                {{ formatBytes(row.totalDisk - row.availableDisk) }} / {{ formatBytes(row.totalDisk) }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="lastHeartbeat" label="最后心跳" width="160" align="right">
          <template #default="{ row }">
            {{ formatTime(row.lastHeartbeat) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetails(row)">详情</el-button>
            <el-button
                type="warning"
                link
                size="small"
                @click="restartNodeHandler(row)"
                :disabled="row.status === '离线'"
            >
                重启
            </el-button>
            <el-button type="danger" link size="small" @click="removeNode(row)">移除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

     <!-- 节点详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="节点详情"
      width="80%"
      top="5vh"
    >
      <div v-if="currentNode" class="node-details">
        <el-tabs v-model="activeTab" class="detail-tabs">
          <!-- 基本信息 -->
          <el-tab-pane label="基本信息" name="basic">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="节点名称">{{ currentNode.name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <div class="flex items-center gap-2">
                  <el-tag :type="getStatusType(currentNode.status)">{{ currentNode.status || '未知' }}</el-tag>
                  <span v-if="currentNode.statusReason && currentNode.statusReason !== '运行正常'" class="text-red-500 text-xs">
                    {{ currentNode.statusReason }}
                  </span>
                </div>
              </el-descriptions-item>
              <el-descriptions-item label="节点ID">{{ currentNode.id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="主机名">{{ currentNode.host || currentNode.ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="传输地址">{{ currentNode.transportAddress || '-' }}</el-descriptions-item>
              <el-descriptions-item label="HTTP地址">{{ currentNode.httpAddress || '-' }}</el-descriptions-item>
              <el-descriptions-item label="版本">{{ currentNode.version || '-' }}</el-descriptions-item>
              <el-descriptions-item label="角色">
                   <el-tag v-for="role in currentNode.roles" :key="role" size="small" class="mr-1">{{ role }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="JVM堆内存" span="2">
                   <div class="w-full">
                       <el-progress
                         :percentage="currentNode.memoryUsage || 0"
                         :color="getProgressColor(currentNode.memoryUsage || 0)"
                         :stroke-width="15"
                         :text-inside="true"
                       />
                   </div>
              </el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>

          <!-- 系统信息 -->
          <el-tab-pane label="系统信息" name="system">
             <el-descriptions :column="2" border>
                <el-descriptions-item label="CPU核心">{{ currentNode.cpuCores || '-' }}</el-descriptions-item>
                <el-descriptions-item label="CPU使用率">
                    <el-progress :percentage="currentNode.cpuUsage || 0" :color="getProgressColor(currentNode.cpuUsage || 0)" />
                </el-descriptions-item>
                <el-descriptions-item label="操作系统" span="2">
                    {{ currentNode.osName || '-' }} {{ currentNode.osVersion || '' }} ({{ currentNode.osArch || '-' }})
                </el-descriptions-item>
                 <el-descriptions-item label="总内存">{{ formatBytes(currentNode.totalMemory) }}</el-descriptions-item>
                 <el-descriptions-item label="可用内存">{{ formatBytes(currentNode.availableMemory) }}</el-descriptions-item>
             </el-descriptions>
          </el-tab-pane>

          <!-- JVM信息 -->
          <el-tab-pane label="JVM信息" name="jvm">
             <el-descriptions :column="2" border>
                 <el-descriptions-item label="JVM版本">{{ currentNode.jvm?.version || '-' }}</el-descriptions-item>
                 <el-descriptions-item label="供应商">{{ currentNode.jvm?.vm_vendor || '-' }}</el-descriptions-item>
             </el-descriptions>

             <div class="mt-4">
                 <h4 class="mb-2">GC收集器详情</h4>
                 <el-table :data="getGCCollectors()" stripe border>
                   <el-table-column prop="name" label="收集器" />
                   <el-table-column prop="collectionCount" label="次" width="100" />
                   <el-table-column label="总耗时" width="150">
                       <template #default="{row}">{{ formatDuration(row.collection_time_in_millis) }}</template>
                   </el-table-column>
                 </el-table>
             </div>
          </el-tab-pane>

        </el-tabs>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import {computed, onMounted, reactive, ref} from 'vue'
import {ElMessage, ElMessageBox} from 'element-plus'
import {CircleCheck, CircleClose, Connection, Refresh, Search, Warning} from '@element-plus/icons-vue'
import {getNodes, getNodeStats, restartNode} from '@/api/es/node.js'
import {getSelectedInstanceId} from '@/stores/instanceStore'

// 状态和数据
const loading = ref(false)
const searchQuery = ref('')
const detailDialogVisible = ref(false)
const currentNode = ref(null)
const activeTab = ref('basic')
const nodes = ref([])

const stats = reactive({
  totalNodes: 0,
  activeNodes: 0,
  warningNodes: 0,
  failedNodes: 0
})

// 过滤节点
const filteredNodes = computed(() => {
  if (!searchQuery.value) return nodes.value
  const query = searchQuery.value.toLowerCase()
  return nodes.value.filter(node =>
    node.name.toLowerCase().includes(query) ||
    node.ip.includes(query)
  )
})

// 工具函数
const formatBytes = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatTime = (time) => {
    if (!time) return '-'
    return new Date(time).toLocaleString()
}

const formatDuration = (ms) => {
    if (!ms) return '0ms'
    if (ms < 1000) return ms + 'ms'
    const s = Math.floor(ms / 1000)
    if (s < 60) return s + 's'
    const m = Math.floor(s / 60)
    if (m < 60) return m + 'm ' + (s % 60) + 's'
    const h = Math.floor(m / 60)
    return h + 'h ' + (m % 60) + 'm'
}

const getProgressColor = (percentage) => {
  if (percentage < 60) return '#67C23A'
  if (percentage < 80) return '#E6A23C'
  return '#F56C6C'
}

const getStatusType = (status) => {
  const map = {
    '在线': 'success',
    '离线': 'danger',
    '警告': 'warning'
  }
  return map[status] || 'info'
}

// 数据获取
const refreshData = async () => {
  loading.value = true
  try {
    const instanceId = getSelectedInstanceId()
    if (!instanceId) {
        ElMessage.warning('请选择实例')
        return
    }

    const [nodesRes, statsRes] = await Promise.all([
        getNodes(instanceId),
        getNodeStats(instanceId)
    ])

    // Assume API returns a map of nodes inside raw_response
    const nodesData = nodesRes.data?.raw_response?.nodes || {}
    const statsData = statsRes.data?.raw_response?.nodes || {}

    // Transform to array
    const nodeList = Object.keys(nodesData).map(nodeId => {
        const nodeInfo = nodesData[nodeId]
        const nodeStats = statsData[nodeId] || {}

        // Calculate status
        let status = '在线'
        if (!nodeInfo) status = '离线'

        return {
            id: nodeId,
            name: nodeInfo.name,
            ip: nodeInfo.ip,
            host: nodeInfo.host,
            version: nodeInfo.version,
            transportAddress: nodeInfo.transport_address,
            httpAddress: nodeInfo.http?.publish_address,
            roles: nodeInfo.roles,
            attributes: nodeInfo.attributes,
            jvm: nodeInfo.jvm,
            os: nodeInfo.os,
            process: nodeInfo.process,

            // Stats
            cpuUsage: nodeStats.os?.cpu?.percent || 0,
            loadAverage: nodeStats.os?.cpu?.load_average?.['1m'] || 0,
            memoryUsage: nodeStats.os?.mem?.used_percent || 0,
            totalMemory: nodeStats.os?.mem?.total_in_bytes || 0,
            availableMemory: nodeStats.os?.mem?.free_in_bytes || 0,
            diskUsage: nodeStats.fs?.total?.total_in_bytes ? Math.round((nodeStats.fs.total.total_in_bytes - nodeStats.fs.total.available_in_bytes) * 100 / nodeStats.fs.total.total_in_bytes) : 0,
            totalDisk: nodeStats.fs?.total?.total_in_bytes || 0,
            availableDisk: nodeStats.fs?.total?.available_in_bytes || 0,

            lastHeartbeat: Date.now(), // Mock
            status: status
        }
    })

    nodes.value = nodeList

    // Update stats
    stats.totalNodes = nodeList.length
    stats.activeNodes = nodeList.filter(n => n.status === '在线').length
    stats.failedNodes = nodeList.filter(n => n.status === '离线').length
    stats.warningNodes = nodeList.filter(n => n.status === '警告').length

    ElMessage.success('刷新成功')
  } catch (e) {
      console.error(e)
      ElMessage.error('获取数据失败')
  } finally {
      loading.value = false
  }
}

// 操作
const viewDetails = (row) => {
    currentNode.value = row
    detailDialogVisible.value = true
}

const restartNodeHandler = async (row) => {
    try {
        await ElMessageBox.confirm(`确定要重启节点 ${row.name} 吗?`, '警告', {
            type: 'warning',
            confirmButtonText: '确定重启',
            cancelButtonText: '取消'
        })

        const instanceId = getSelectedInstanceId()
        await restartNode(instanceId, row.id)
        ElMessage.success('重启命令已发送')
    } catch (e) {
        if (e !== 'cancel') ElMessage.error('重启失败: ' + e.message)
    }
}

const removeNode = (row) => {
    ElMessage.warning('暂不支持移除节点功能')
}

// Helpers for detail view
const getGCCollectors = () => {
    if (!currentNode.value?.nodes_stats?.jvm?.gc?.collectors) return []
    return Object.entries(currentNode.value.nodes_stats.jvm.gc.collectors).map(([name, data]) => ({
        name,
        ...data
    }))
}

onMounted(() => {
    refreshData()
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

.mr-2 { margin-right: 8px; }
.mr-1 { margin-right: 4px; }
.mt-1 { margin-top: 4px; }
.mt-4 { margin-top: 16px; }
.mb-2 { margin-bottom: 8px; }
.w-64 { width: 16rem; }
.w-full { width: 100%; }
.text-xs { font-size: 12px; }
.text-right { text-align: right; }
.text-secondary { color: var(--text-sub); }
.text-red-500 { color: var(--danger-color); }
.flex { display: flex; }
.flex-col { flex-direction: column; }
.items-center { align-items: center; }
.gap-2 { gap: 8px; }
.flex-1 { flex: 1; }

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-left { display: flex; align-items: center; gap: 12px; }
.header-title-wrapper h2 { margin: 0; font-size: 18px; color: var(--text-main); }
.header-title-wrapper .subtitle { margin: 0; font-size: 12px; color: var(--text-sub); }
.header-right { display: flex; align-items: center; gap: 12px; }
</style>
