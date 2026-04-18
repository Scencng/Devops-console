<template>
  <div class="autoops-container chaos-detail-page">
    <div v-if="!isComponentMode" class="page-header">
      <el-page-header @back="handleBack" title="返回">
        <template #content>
          <span class="page-title">混沌实验详情</span>
        </template>
      </el-page-header>
    </div>

    <div v-loading="loading" class="detail-content">
      <!-- Status Overview -->
      <el-card class="status-card">
        <template #header>
          <div class="card-header">
            <span class="card-title">{{ experiment.name }}</span>
            <el-tag :type="getStatusType(experiment.status)" size="large">
              {{ experiment.status }}
            </el-tag>
          </div>
        </template>

        <el-descriptions :column="2" border>
          <el-descriptions-item label="命名空间">
            {{ experiment.namespace }}
          </el-descriptions-item>
          <el-descriptions-item label="故障类型">
            {{ experiment.faultType }}
          </el-descriptions-item>
          <el-descriptions-item label="Phase">
            {{ experiment.phase || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="持续时间">
            {{ experiment.duration || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(experiment.createdAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="开始时间">
            {{ formatTime(experiment.startedAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="结束时间">
            {{ formatTime(experiment.finishedAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            {{ experiment.status }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="action-buttons">
          <el-button
            v-if="experiment.status === 'Running' || experiment.status === 'running'"
            type="warning"
            @click="handlePause"
          >
            暂停实验
          </el-button>
          <el-button
            v-if="experiment.status === 'paused' || experiment.status === 'Paused'"
            type="success"
            @click="handleResume"
          >
            恢复实验
          </el-button>
          <el-button
            type="danger"
            @click="handleDelete"
          >
            删除实验
          </el-button>
        </div>
      </el-card>

      <!-- Configuration -->
      <el-card class="config-card">
        <template #header>
          <span class="card-title">实验配置</span>
        </template>

        <el-collapse v-if="experiment.spec">
          <el-collapse-item title="基本配置" name="basic">
            <pre class="config-json">{{ formatJSON(experiment.spec) }}</pre>
          </el-collapse-item>
        </el-collapse>
      </el-card>

      <!-- Labels -->
      <el-card class="labels-card">
        <template #header>
          <span class="card-title">标签</span>
        </template>

        <div v-if="experiment.labels && Object.keys(experiment.labels).length > 0" class="labels-container">
          <el-tag
            v-for="(value, key) in experiment.labels"
            :key="key"
            class="label-tag"
          >
            {{ key }}: {{ value }}
          </el-tag>
        </div>
        <el-empty v-else description="暂无标签" :image-size="60" />
      </el-card>

      <!-- Events -->
      <el-card class="events-card">
        <template #header>
          <span class="card-title">事件日志</span>
        </template>

        <el-timeline v-if="experiment.events && experiment.events.length > 0">
          <el-timeline-item
            v-for="(event, index) in experiment.events"
            :key="index"
            :timestamp="formatTime(event.timestamp)"
            :type="getEventType(event.type)"
          >
            <div class="event-content">
              <div class="event-type">{{ event.type }}</div>
              <div class="event-reason">{{ event.reason }}</div>
              <div class="event-message">{{ event.message }}</div>
              <div v-if="event.count > 1" class="event-count">
                发生次数: {{ event.count }}
              </div>
            </div>
          </el-timeline-item>
        </el-timeline>
        <el-empty v-else description="暂无事件" :image-size="60" />
      </el-card>

      <!-- Targets -->
      <el-card class="targets-card">
        <template #header>
          <span class="card-title">影响目标</span>
        </template>

        <el-table
          v-if="experiment.targets && experiment.targets.length > 0"
          :data="experiment.targets"
          style="width: 100%"
        >
          <el-table-column label="Pod名称" min-width="200">
            <template #default="{ row }">
              {{ row.podName }}
            </template>
          </el-table-column>

          <el-table-column label="命名空间" min-width="150">
            <template #default="{ row }">
              {{ row.namespace }}
            </template>
          </el-table-column>

          <el-table-column label="节点名称" min-width="150">
            <template #default="{ row }">
              {{ row.nodeName || '-' }}
            </template>
          </el-table-column>

          <el-table-column label="状态" min-width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="getTargetStatusType(row.status)" size="small">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="消息" min-width="300">
            <template #default="{ row }">
              {{ row.message || '-' }}
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="暂无目标" :image-size="60" />
      </el-card>
    </div>
  </div>
</template>

<script setup>
import {computed, onMounted, reactive, ref} from 'vue'
import {ElMessage, ElMessageBox} from 'element-plus'
import {useRoute, useRouter} from 'vue-router'
import {deleteChaosExperiment, getChaosExperiment, pauseChaosExperiment, resumeChaosExperiment} from '@/api/chaos'
import {getSelectedInstanceId} from '@/stores/instanceStore'

const props = defineProps({
  experiment: {
    type: Object,
    default: null
  },
  instanceId: {
    type: [Number, String],
    default: null
  }
})

const emit = defineEmits(['pause', 'delete'])

const isComponentMode = computed(() => !!props.experiment)

const router = useRouter()
const route = useRoute()

const loading = ref(false)
const experiment = reactive({
  name: '',
  namespace: '',
  faultType: '',
  status: '',
  phase: '',
  createdAt: null,
  startedAt: null,
  finishedAt: null,
  duration: '',
  labels: {},
  annotations: {},
  spec: null,
  events: [],
  targets: []
})

const fetchDetail = async () => {
  loading.value = true
  try {
    const instanceId = props.instanceId || getSelectedInstanceId()
    const namespace = props.experiment?.namespace || route.query.namespace
    const name = props.experiment?.name || route.query.name

    if (!namespace || !name) {
       if (props.experiment) Object.assign(experiment, props.experiment)
       return
    }

    const res = await getChaosExperiment(namespace, name, instanceId)

    Object.assign(experiment, res.data || {})
  } catch (e) {
    ElMessage.error('获取实验详情失败')
  } finally {
    loading.value = false
  }
}

const handlePause = async () => {
  if (isComponentMode.value) {
    emit('pause', experiment)
    return
  }

  try {
    await ElMessageBox.confirm(`确认暂停实验 "${experiment.name}"?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const instanceId = getSelectedInstanceId()
    await pauseChaosExperiment(experiment.namespace, experiment.name, instanceId)
    ElMessage.success('暂停成功')
    fetchDetail()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('暂停失败')
    }
  }
}

const handleResume = async () => {
  if (isComponentMode.value) {
    emit('pause', experiment) // the list handles both via toggle pause
    return
  }

  try {
    await ElMessageBox.confirm(`确认恢复实验 "${experiment.name}"?`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const instanceId = getSelectedInstanceId()
    await resumeChaosExperiment(experiment.namespace, experiment.name, instanceId)
    ElMessage.success('恢复成功')
    fetchDetail()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('恢复失败')
    }
  }
}

const handleDelete = async () => {
  if (isComponentMode.value) {
    emit('delete', experiment)
    return
  }

  try {
    await ElMessageBox.confirm(
      `确认删除实验 "${experiment.name}"? 此操作不可恢复`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const instanceId = getSelectedInstanceId()
    await deleteChaosExperiment(experiment.namespace, experiment.name, instanceId)
    ElMessage.success('删除成功')
    router.push('/chaos')
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleBack = () => {
  router.push('/chaos')
}

const getStatusType = (status) => {
  if (!status) return 'info'
  const cap = typeof status === 'string' ? status.charAt(0).toUpperCase() + status.slice(1).toLowerCase() : status
  const typeMap = {
    'Running': 'success',
    'Paused': 'warning',
    'Failed': 'danger',
    'Finished': 'info'
  }
  return typeMap[cap] || 'info'
}

const getTargetStatusType = (status) => {
  const typeMap = {
    'affected': 'danger',
    'recovered': 'success',
    'pending': 'warning'
  }
  return typeMap[status] || 'info'
}

const getEventType = (type) => {
  const typeMap = {
    'Warning': 'warning',
    'Error': 'danger',
    'Info': 'primary',
    'Normal': 'success'
  }
  return typeMap[type] || 'primary'
}

const formatTime = (time) => {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleString('zh-CN')
}

const formatJSON = (obj) => {
  return JSON.stringify(obj, null, 2)
}

onMounted(() => {
  fetchDetail()
})
</script>

<style scoped>
.chaos-detail-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-title {
  font-size: 18px;
  font-weight: 500;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.status-card,
.config-card,
.labels-card,
.events-card,
.targets-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  font-size: 16px;
  font-weight: 500;
}

.action-buttons {
  margin-top: 20px;
  display: flex;
  gap: 12px;
}

.config-json {
  background: #f5f7fa;
  padding: 16px;
  border-radius: 4px;
  overflow-x: auto;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
}

.labels-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.label-tag {
  margin: 4px;
}

.event-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.event-type {
  font-weight: 500;
  color: #303133;
}

.event-reason {
  font-size: 14px;
  color: #606266;
}

.event-message {
  font-size: 13px;
  color: #909399;
}

.event-count {
  font-size: 12px;
  color: #c0c4cc;
}
</style>
