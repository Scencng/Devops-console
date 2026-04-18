<template>
  <div class="page-container" v-loading="loading">
    <!-- Page Header -->
    <el-card class="page-header-card">
       <div class="page-header">
         <div class="header-left">
           <div class="header-icon-wrapper">
             <el-icon :size="24"><Search /></el-icon>
           </div>
           <div class="header-title-wrapper">
             <h2>Elasticsearch 控制台</h2>
             <p class="subtitle">集群监控与数据管理</p>
           </div>
         </div>
         <div class="header-right">
           <el-button @click="refreshData">
              <el-icon><Refresh /></el-icon> 刷新
           </el-button>
         </div>
       </div>
    </el-card>

    <!-- Stats Section -->
    <el-row :gutter="20" class="mb-20">
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
               <div class="stat-icon bg-gradient-blue"><el-icon><DocumentCopy /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ stats.totalIndices }}</div>
                  <div class="stat-label">索引总数</div>
               </div>
            </div>
          </el-card>
       </el-col>
       <el-col :span="6">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-content">
               <div class="stat-icon bg-gradient-orange"><el-icon><Grid /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ stats.totalShards }}</div>
                  <div class="stat-label">分片总数</div>
               </div>
            </div>
          </el-card>
       </el-col>
       <el-col :span="6">
          <el-card class="stat-card" shadow="hover">
            <div class="stat-content">
               <div class="stat-icon bg-gradient-purple"><el-icon><Monitor /></el-icon></div>
               <div class="stat-info">
                  <div class="stat-value">{{ stats.totalDocs }}</div>
                  <div class="stat-label">文档总数</div>
               </div>
            </div>
          </el-card>
       </el-col>
    </el-row>

    <!-- Main Content -->
    <el-row :gutter="20" class="main-content-row">
       <el-col :span="12">
          <el-card class="content-card h-full">
             <template #header>
                <div class="flex-between">
                   <div class="items-center flex gap-2">
                      <el-icon><Tools /></el-icon> <span>主节点信息</span>
                   </div>
                   <el-tag type="success" size="small">在线</el-tag>
                </div>
             </template>
             <div class="info-list">
                <div class="info-item">
                  <span class="info-label">节点名称:</span>
                  <span class="info-value">{{ masterNode.name || '未知' }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">集群UUID:</span>
                  <span class="info-value text-xs">{{ clusterInfo.clusterUUID || '未知' }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">版本:</span>
                  <el-tooltip :content="getVersionDetails(clusterInfo.version)" placement="top" :disabled="!clusterInfo.version">
                    <span class="info-value cursor-pointer flex items-center gap-1">
                      {{ formatVersion(clusterInfo.version) || '未知' }}
                      <el-icon v-if="clusterInfo.version"><InfoFilled /></el-icon>
                    </span>
                  </el-tooltip>
                </div>
                <div class="info-item">
                  <span class="info-label">角色:</span>
                  <span class="info-value">master</span>
                </div>
                <div class="info-item">
                  <span class="info-label">状态:</span>
                  <el-tag :type="getClusterStatusType(clusterHealth.status)" size="small">
                    {{ clusterHealth.status?.toUpperCase() || '未知' }}
                  </el-tag>
                </div>
             </div>
          </el-card>
       </el-col>

       <el-col :span="12">
          <el-card class="content-card h-full">
             <template #header>
                <div class="flex-between">
                   <div class="items-center flex gap-2">
                       <el-icon><DataLine /></el-icon> <span>集群健康状态</span>
                   </div>
                   <el-tag :type="getClusterStatusType(clusterHealth.status)" size="small">
                     {{ clusterHealth.status?.toUpperCase() }}
                   </el-tag>
                </div>
             </template>
             <div class="info-list">
                <div class="info-item">
                  <span class="info-label">集群名称:</span>
                  <span class="info-value">{{ clusterInfo.name || '未知' }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">节点总数:</span>
                  <span class="info-value">{{ clusterHealth.number_of_nodes || 0 }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">活跃主分片:</span>
                  <span class="info-value">{{ clusterHealth.active_primary_shards || 0 }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">活跃分片:</span>
                  <span class="info-value">{{ clusterHealth.active_shards || 0 }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">数据节点:</span>
                  <span class="info-value">{{ stats.activeNodes || 0 }}</span>
                </div>
                 <div class="info-item">
                  <span class="info-label">总分片:</span>
                  <span class="info-value">{{ stats.totalShards || 0 }}</span>
                </div>
             </div>
          </el-card>
       </el-col>
    </el-row>
  </div>
</template>

<script setup>
import {nextTick, onMounted, ref } from 'vue'
import {ElMessage} from 'element-plus'
import {
    CircleCheck, Monitor, DocumentCopy, Grid, Warning, Connection, 
    InfoFilled, Search, Refresh, Tools, DataLine
} from '@element-plus/icons-vue'
import { getClusterStatus, getClusterHealth, getClusterInfo } from '../../api/es/cluster.js'
import { getInstanceList } from '../../api/instance.js'
import { getIndices } from '../../api/es/indices.js'

// 统计数据
const stats = ref({
  activeNodes: 0,
  totalIndices: 0,
  totalShards: 0,
  totalDocs: '0'
})

// 集群健康数据
const clusterHealth = ref({
  status: 'unknown',
  number_of_nodes: 0,
  active_primary_shards: 0,
  active_shards: 0
})

// 集群信息
const clusterInfo = ref({
  name: '',
  version: '',
  masterNode: ''
})

// 主节点信息
const masterNode = ref({
  name: '',
  id: '',
  ip: '',
  version: '',
  jvmHeap: ''
})

// 加载状态
const loading = ref(false)
const selectedInstance = ref(null)

// 获取第一个可用的elasticsearch实例
const fetchFirstInstance = async () => {
  try {
    const response = await getInstanceList({ 
      type_name: 'elasticsearch',
      page: 1,
      page_size: 100
    })
    if (response && response.data && response.data.list && response.data.list.data && response.data.list.data.length > 0) {
      selectedInstance.value = response.data.list.data[0].id
      await fetchData()
    } else {
      ElMessage.warning('未找到可用的Elasticsearch实例')
    }
  } catch (error) {
    console.error('获取实例失败:', error)
    ElMessage.error('获取实例失败: ' + (error.response?.data?.message || error.message))
  }
}

const getClusterStatusType = (status) => {
  switch (status) {
    case 'green': return 'success'
    case 'yellow': return 'warning'
    case 'red': return 'danger'
    default: return 'info'
  }
}

const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  } else if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

const formatVersion = (versionStr) => {
  if (!versionStr) return '未知'
  const versionMatch = versionStr.match(/(\d+\.\d+\.\d+)/)
  if (versionMatch) return versionMatch[1]
  const altVersionMatch = versionStr.match(/(\d+\.\d+)/)
  if (altVersionMatch) return altVersionMatch[1]
  return versionStr
}

const getVersionDetails = (versionStr) => {
  if (!versionStr) return ''
  const buildMatch = versionStr.match(/构建:\s*([^,)]+)/)
  const dateMatch = versionStr.match(/构建日期:\s*([^)]+)/)
  
  let details = versionStr
  if (buildMatch && dateMatch) {
    details = `版本: ${formatVersion(versionStr)}\n构建: ${buildMatch[1]}\n日期: ${dateMatch[1]}`
  } else if (buildMatch) {
    details = `版本: ${formatVersion(versionStr)}\n构建: ${buildMatch[1]}`
  }
  return details
}

const fetchData = async () => {
  if (!selectedInstance.value) {
    ElMessage.warning('请先选择一个Elasticsearch实例')
    return
  }

  loading.value = true
  try {
    const [statusResponse, healthResponse, infoResponse, indicesResponse] = await Promise.all([
      getClusterStatus(selectedInstance.value),
      getClusterHealth(selectedInstance.value),
      getClusterInfo(selectedInstance.value),
      getIndices(selectedInstance.value)
    ])

    if (statusResponse && statusResponse.data && statusResponse.data.cluster_status) {
      const statusData = statusResponse.data.cluster_status
      stats.value.activeNodes = statusData.number_of_nodes || 0
      stats.value.totalShards = statusData.active_shards || 0
      clusterHealth.value.status = statusData.status || 'unknown'
      clusterHealth.value.number_of_nodes = statusData.number_of_nodes || 0
      clusterHealth.value.active_primary_shards = statusData.active_primary_shards || 0
      clusterHealth.value.active_shards = statusData.active_shards || 0
      clusterInfo.value.name = statusData.cluster_name || ''
      clusterInfo.value.version = statusData.version_info || ''
    }

    if (healthResponse && healthResponse.data && healthResponse.data.health_info) {
      const healthData = healthResponse.data.health_info
      clusterHealth.value.status = healthData.status || 'unknown'
      clusterHealth.value.number_of_nodes = healthData.number_of_nodes || 0
      clusterHealth.value.active_primary_shards = healthData.active_primary_shards || 0
      clusterHealth.value.active_shards = healthData.active_shards || 0
      
      if (healthData.doc_count !== undefined && healthData.doc_count !== null) {
        stats.value.totalDocs = formatNumber(healthData.doc_count)
      }
    }

    if (infoResponse && infoResponse.data && infoResponse.data.cluster_info) {
      const infoData = infoResponse.data.cluster_info
      clusterInfo.value.name = infoData.cluster_name || ''
      clusterInfo.value.version = infoData.version || ''
      clusterInfo.value.clusterUUID = infoData.cluster_uuid || ''
      masterNode.value.name = infoData.name || ''
    }

    if (indicesResponse && indicesResponse.data) {
      let indicesData = indicesResponse.data
      if (indicesData.raw_response && Array.isArray(indicesData.raw_response)) {
        indicesData = indicesData.raw_response
      }
      
      if (Array.isArray(indicesData)) {
        stats.value.totalIndices = indicesData.length
        let totalDocs = 0
        indicesData.forEach(index => {
          const docCount = parseInt(index['docs.count']) || 0
          totalDocs += docCount
        })
        stats.value.totalDocs = formatNumber(totalDocs)
      } else if (indicesData.list && Array.isArray(indicesData.list)) {
        stats.value.totalIndices = indicesData.list.length
      }
    }

  } catch (error) {
    console.error('获取数据失败:', error)
    ElMessage.error('获取集群数据失败: ' + (error.response?.data?.message || error.message))
  } finally {
    loading.value = false
  }
}

const refreshData = async () => {
  ElMessage.info('正在刷新数据...')
  try {
    await fetchData()
    ElMessage.success('数据刷新成功')
  } catch (error) {
    ElMessage.error('数据刷新失败: ' + error.message)
  }
}

onMounted(async () => {
  await nextTick()
  await fetchFirstInstance()
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

.items-center { display: flex; align-items: center; }
.flex { display: flex; }
.gap-2 { gap: 8px; }
.gap-1 { gap: 4px; }
.h-full { height: 100%; }
.text-xs { font-size: 12px; }
.cursor-pointer { cursor: pointer; }

/* Info List */
.info-list {
  display: flex;
  flex-direction: column;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-color);
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-size: 14px;
  color: var(--text-sub);
  font-weight: 500;
}

.info-value {
  font-size: 14px;
  color: var(--text-main);
}
</style>