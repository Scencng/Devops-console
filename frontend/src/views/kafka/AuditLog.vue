<template>
  <div class="page-container">
    <el-card class="page-header-card" shadow="never">
      <div class="page-header">
        <div class="page-header-copy">
          <div class="page-eyebrow">Kafka</div>
          <h2>审计日志</h2>
          <p>把高风险操作和失败记录收拢到同一视图里，优先确认删除、配置变更和 Offset 重置。</p>
        </div>

        <div class="page-header-side">
          <div class="page-header-meta">
            <div class="page-header-kpi">
              <span>筛选范围</span>
              <strong>{{ activeClusterLabel }}</strong>
            </div>
            <div class="page-header-kpi">
              <span>失败记录</span>
              <strong>{{ logStats.failed }}</strong>
            </div>
          </div>
          <div class="page-header-note">
            {{ errorClusters[0] ? `最常见失败原因：${errorClusters[0].reason}` : '最近没有明显的失败热点。' }}
          </div>
          <div class="page-header-actions">
            <el-button @click="loadLogs" :loading="loading">刷新</el-button>
          </div>
        </div>
      </div>
    </el-card>

    <div class="page-metrics">
      <div class="page-metric-card is-accent">
        <span>日志条数</span>
        <strong>{{ logStats.total }}</strong>
      </div>
      <div class="page-metric-card is-success">
        <span>当前页成功</span>
        <strong>{{ logStats.success }}</strong>
      </div>
      <div class="page-metric-card is-warning">
        <span>当前页失败</span>
        <strong>{{ logStats.failed }}</strong>
      </div>
    </div>

    <el-card class="content-card">
      <template #header>
        <div class="card-header">
          <span>高风险筛选</span>
          <span class="card-subtitle">快捷入口</span>
        </div>
      </template>

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
          <el-button type="primary" @click="handleSearch">查询</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>审计记录</span>
          <span class="card-subtitle">最近审计记录</span>
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

      <el-pagination
        v-model:current-page="filters.page"
        v-model:page-size="filters.pageSize"
        class="audit-pagination"
        :page-sizes="[20, 50, 100]"
        :total="paginationTotal"
        layout="total, sizes, prev, pager, next"
        @size-change="loadLogs"
        @current-change="loadLogs"
      />
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { ElMessage } from 'element-plus'
import { getKafkaAuditLogs } from '@/api/kafka.js'
import { useKafkaStore } from '@/stores/kafkaStore.js'
import { formatDateTime } from '@/utils/dateTime.js'

const loading = ref(false)
const logs = ref([])
const paginationTotal = ref(0)
const filters = reactive({ clusterId: null, action: '', result: '', page: 1, pageSize: 20 })
const kafkaStore = useKafkaStore()
const { clusterOptions: clusters } = storeToRefs(kafkaStore)
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

const activeClusterLabel = computed(
  () => clusters.value.find((item) => item.id === filters.clusterId)?.name || '全部集群',
)

const logStats = computed(() => {
  const clusterIds = new Set(logs.value.map((item) => item.clusterId).filter(Boolean))
  return {
    total: paginationTotal.value,
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

const formatTime = formatDateTime

const applyQuickActionFilter = async (action) => {
  filters.action = action
  filters.page = 1
  await loadLogs()
}

const handleSearch = async () => {
  filters.page = 1
  await loadLogs()
}

const clearQuickFilters = async () => {
  filters.action = ''
  filters.result = ''
  filters.page = 1
  await loadLogs()
}

const loadClusters = async () => {
  try {
    await kafkaStore.loadClusterOptions()
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 集群列表加载失败')
    throw error
  }
}

const loadLogs = async () => {
  if (loading.value) return
  loading.value = true
  try {
    const res = await getKafkaAuditLogs(filters)
    logs.value = res?.data?.data?.list || []
    paginationTotal.value = Number(res?.data?.data?.total || 0)
  } catch (error) {
    ElMessage.error(error.message || '审计日志加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    await loadClusters()
    await loadLogs()
  } catch (error) {
    if (!String(error?.message || '').includes('Kafka 集群列表加载失败')) {
      ElMessage.error(error.message || '审计日志初始化失败')
    }
  }
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
  color: var(--shell-accent);
  border-color: var(--shell-accent-border-strong);
  background: var(--shell-accent-faint);
  box-shadow: inset 0 0 0 1px var(--shell-accent-ring);
}

.audit-pagination {
  margin-top: 16px;
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
