<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <div class="page-eyebrow">Audit Trail</div>
          <h2>审计日志</h2>
          <p>追踪 Kafka 集群、Topic 与消费组相关的高风险操作记录，帮助你还原是谁在什么时候做了什么。</p>
        </div>
      </div>
    </el-card>

    <div class="page-metrics">
      <div class="page-metric-card is-accent">
        <span>日志条数</span>
        <strong>{{ logs.length }}</strong>
        <p>当前筛选条件下返回的审计记录数量。</p>
      </div>
      <div class="page-metric-card is-success">
        <span>成功记录</span>
        <strong>{{ logStats.success }}</strong>
        <p>最近筛选结果中执行成功的操作数量。</p>
      </div>
      <div class="page-metric-card is-warning">
        <span>失败记录</span>
        <strong>{{ logStats.failed }}</strong>
        <p>建议优先查看失败动作的错误原因。</p>
      </div>
      <div class="page-metric-card">
        <span>关联集群</span>
        <strong>{{ logStats.clusterCount }}</strong>
        <p>当前筛选结果覆盖的集群数量。</p>
      </div>
    </div>

    <el-card class="content-card">
      <template #header>
        <div class="card-header card-header-wrap">
          <span>失败原因聚类与高风险动作入口</span>
          <span class="card-subtitle">把失败最多的原因和最敏感的动作放在当前页，方便直接切换筛选条件</span>
        </div>
      </template>

      <div class="workbench-grid">
        <div class="workspace-panel">
          <h3>失败原因聚类</h3>
          <p>基于当前筛选结果，对失败记录的错误原因做简单聚合，帮助快速定位最常见问题。</p>
          <div class="compact-list">
            <div v-for="item in errorClusters" :key="item.reason" class="compact-item">
              <div>
                <strong>{{ item.reason }}</strong>
                <span>出现 {{ item.count }} 次</span>
              </div>
              <el-tag type="danger">失败聚类</el-tag>
            </div>
          </div>
        </div>

        <div class="workspace-panel">
          <h3>高风险筛选</h3>
          <p>常见动作快捷筛选。</p>
          <div class="quick-filter-grid">
            <el-button
              v-for="item in quickRiskActions"
              :key="item.value"
              class="quick-filter-btn"
              :class="{ 'is-active': filters.action === item.value }"
              @click="applyQuickActionFilter(item.value)"
            >
              {{ item.label }}
            </el-button>
            <el-button
              class="quick-filter-btn"
              :class="{ 'is-active': !filters.action && !filters.result }"
              @click="clearQuickFilters"
            >
              清空
            </el-button>
          </div>
        </div>
      </div>
    </el-card>

    <el-card class="content-card filter-card">
      <div class="toolbar-row">
        <div class="toolbar-left">
          <el-select v-model="filters.clusterId" clearable style="width: 240px">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-input v-model="filters.action" placeholder="如 topic:delete" style="width: 220px" />
          <el-select v-model="filters.result" clearable style="width: 160px">
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
          </el-select>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" @click="loadLogs">查询</el-button>
          <el-button @click="loadLogs" :loading="loading">刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>审计记录</span>
          <span class="card-subtitle">优先查看失败动作、删除类操作和 Offset 干预记录</span>
        </div>
      </template>

      <el-table :data="logs" empty-text="暂无审计日志">
        <el-table-column prop="createdAt" label="时间" width="180">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column prop="operatorUsername" label="操作人" width="140" />
        <el-table-column prop="action" label="动作" width="200" />
        <el-table-column prop="resourceType" label="资源类型" width="120" />
        <el-table-column prop="resourceName" label="资源名称" min-width="200" />
        <el-table-column prop="result" label="结果" width="100">
          <template #default="{ row }">
            <el-tag :type="row.result === 'success' ? 'success' : 'danger'">{{ row.result }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="errorMessage" label="错误信息" min-width="220" show-overflow-tooltip />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getKafkaAuditLogs, getKafkaClusterOptions } from '@/api/kafka.js'

const loading = ref(false)
const clusters = ref([])
const logs = ref([])
const filters = reactive({ clusterId: null, action: '', result: '', page: 1, pageSize: 100 })
const quickRiskActions = [
  {
    label: '删 Topic',
    value: 'topic:delete',
  },
  {
    label: '改配置',
    value: 'topic:config:update',
  },
  {
    label: '重置 Offset',
    value: 'group:offset:reset',
  },
  {
    label: '发消息',
    value: 'message:produce',
  },
  {
    label: '删集群',
    value: 'cluster:delete',
  },
]

const logStats = computed(() => {
  const clusterIds = new Set(logs.value.map((item) => item.clusterId).filter(Boolean))
  return {
    success: logs.value.filter((item) => item.result === 'success').length,
    failed: logs.value.filter((item) => item.result === 'failed').length,
    clusterCount: clusterIds.size,
  }
})

const errorClusters = computed(() => {
  const counter = new Map()
  logs.value
    .filter((item) => item.result === 'failed')
    .forEach((item) => {
      const raw = (item.errorMessage || '未返回详细错误信息').trim()
      const reason = raw.length > 40 ? `${raw.slice(0, 40)}...` : raw
      counter.set(reason, (counter.get(reason) || 0) + 1)
    })

  return Array.from(counter.entries())
    .map(([reason, count]) => ({ reason, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 5)
})

const formatTime = (value) => (value ? new Date(value).toLocaleString() : '-')

const applyQuickActionFilter = async (action) => {
  filters.action = action
  await loadLogs()
}

const clearQuickFilters = async () => {
  filters.action = ''
  filters.result = ''
  await loadLogs()
}

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
}

const loadLogs = async () => {
  loading.value = true
  try {
    const res = await getKafkaAuditLogs(filters)
    logs.value = res?.data?.data?.list || []
  } catch (error) {
    ElMessage.error(error.message || '审计日志加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadClusters()
  await loadLogs()
})
</script>

<style scoped>
.quick-filter-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.quick-filter-btn {
  width: 100%;
  min-height: 44px;
  margin: 0;
  padding: 0 12px;
  justify-content: center;
  font-size: 13px;
}

.quick-filter-btn.is-active {
  color: #2563eb;
  border-color: rgba(37, 99, 235, 0.4);
  background: rgba(239, 246, 255, 0.9);
  box-shadow: inset 0 0 0 1px rgba(37, 99, 235, 0.12);
}

@media (max-width: 960px) {
  .quick-filter-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .quick-filter-grid {
    grid-template-columns: 1fr;
  }
}
</style>
