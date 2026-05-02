<template>
  <div class="page-container">
    <el-card class="page-header-card" shadow="never">
      <div class="page-header">
        <div class="page-header-copy">
          <div class="page-eyebrow">Kafka</div>
          <h2>消息浏览</h2>
          <p>按 Topic / Partition 采样查看消息，发送测试消息前先看风险提示，再决定是否写入。</p>
        </div>

        <div class="page-header-side">
          <div class="page-header-meta">
            <div class="page-header-kpi">
              <span>当前集群</span>
              <strong>{{ currentClusterName }}</strong>
            </div>
            <div class="page-header-kpi">
              <span>风险信号</span>
              <strong>{{ currentTopicHints.length }}</strong>
            </div>
          </div>
          <div class="page-header-note">
            {{ currentTopicHints[0]?.description || '当前 Topic 没有明显的风险提示。' }}
          </div>
          <div class="page-header-actions">
            <el-button @click="loadMessages" :loading="loading">刷新</el-button>
            <el-button
              v-if="permStore.hasPerm('kafka:message:produce') || permStore.roles.includes('admin')"
              type="primary"
              @click="openProduceDialog"
            >
              发消息
            </el-button>
          </div>
        </div>
      </div>
    </el-card>

    <div class="page-metrics">
      <div class="page-metric-card">
        <span>当前 Topic</span>
        <strong>{{ form.topic || '-' }}</strong>
      </div>
      <div class="page-metric-card is-success">
        <span>当前分区</span>
        <strong>{{ form.partition }}</strong>
      </div>
      <div class="page-metric-card is-warning">
        <span>返回消息数</span>
        <strong>{{ result.count || 0 }}</strong>
      </div>
    </div>

    <el-card class="content-card">
      <template #header>
        <div class="card-header">
          <span>当前 Topic 提示</span>
          <span class="card-subtitle">浏览前先确认上下文</span>
        </div>
      </template>

      <div class="compact-list">
        <div v-for="item in currentTopicHints" :key="item.title" class="compact-item">
          <div>
            <strong>{{ item.title }}</strong>
            <span>{{ item.description }}</span>
          </div>
          <el-tag :type="item.level === 'high' ? 'danger' : item.level === 'success' ? 'success' : 'warning'">
            {{ item.level === 'high' ? '高风险' : item.level === 'success' ? '正常' : '关注' }}
          </el-tag>
        </div>
      </div>
    </el-card>

    <el-card class="content-card filter-card">
      <div class="toolbar-row">
        <div class="toolbar-left">
          <el-select v-model="form.topic" style="width: 240px" @change="handleTopicChange">
            <el-option v-for="topic in topics" :key="topic.name" :label="topic.name" :value="topic.name" />
          </el-select>
          <el-select v-model="form.partition" style="width: 140px">
            <el-option v-for="p in partitionOptions" :key="p" :label="String(p)" :value="p" />
          </el-select>
          <el-select v-model="form.mode" style="width: 160px">
            <el-option label="最新消息" value="latest" />
            <el-option label="最早消息" value="earliest" />
            <el-option label="指定 Offset" value="offset" />
          </el-select>
          <el-input-number v-model="form.limit" :min="1" :max="500" style="width: 140px" />
        </div>
        <div class="toolbar-right">
          <el-button type="primary" @click="loadMessages" :loading="loading">查询</el-button>
        </div>
      </div>

      <div class="toolbar-row secondary-toolbar">
        <div class="toolbar-left">
          <el-input-number v-if="form.mode === 'offset'" v-model="form.offset" :min="0" placeholder="指定 Offset" style="width: 180px" />
          <el-input v-model="form.keyword" placeholder="按 key/value 过滤" style="width: 260px" />
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>消息列表</span>
          <span class="card-subtitle">
            {{ result.count || 0 }} 条
            <template v-if="result.partial">{{ result.timedOut ? ' / 已截断' : ' / 部分结果' }}</template>
          </span>
        </div>
      </template>

      <el-alert
        v-if="result.warningMessage"
        class="table-alert"
        :type="result.timedOut || result.partial ? 'warning' : 'info'"
        :closable="false"
        show-icon
        :title="result.timedOut ? '本次消息浏览已提前结束' : '消息浏览提醒'"
        :description="result.warningMessage"
      />

      <el-table :data="result.messages || []" empty-text="暂无消息数据">
        <el-table-column prop="offset" label="Offset" width="110" />
        <el-table-column prop="partition" label="Partition" width="100" />
        <el-table-column prop="timestamp" label="时间" width="180">
          <template #default="{ row }">{{ formatTime(row.timestamp) }}</template>
        </el-table-column>
        <el-table-column prop="keyPreview" label="Key" min-width="220" show-overflow-tooltip />
        <el-table-column prop="valuePreview" label="Value" min-width="360" show-overflow-tooltip />
        <el-table-column label="Headers" width="100">
          <template #default="{ row }">{{ row.headers?.length || 0 }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openMessageDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="produceDialogVisible" title="发送测试消息" width="760px" destroy-on-close>
      <el-form label-position="top">
        <div class="produce-helper-grid">
          <div class="surface-muted">
            <div class="editor-title">快捷模板</div>
            <div class="produce-helper-list">
              <button
                v-for="template in produceTemplates"
                :key="template.key"
                type="button"
                class="produce-helper-card"
                @click="applyProduceTemplate(template)"
              >
                <strong>{{ template.label }}</strong>
                <span>{{ template.description }}</span>
              </button>
            </div>
          </div>

          <div v-if="reusableProduceLogs.length" class="surface-muted">
            <div class="editor-title">最近发送</div>
            <div class="produce-helper-list">
              <button
                v-for="log in reusableProduceLogs"
                :key="log.id"
                type="button"
                class="produce-helper-card"
                @click="reuseProduceLog(log)"
              >
                <strong>{{ log.topic || '未识别 Topic' }}</strong>
                <span>{{ log.summary }}</span>
              </button>
            </div>
          </div>
        </div>

        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="集群">
              <el-input :model-value="currentClusterName" disabled />
            </el-form-item>
          </el-col>
          <el-col :span="10">
            <el-form-item label="Topic">
              <el-select v-model="produceForm.topic">
                <el-option v-for="topic in produceTopics" :key="topic.name" :label="topic.name" :value="topic.name" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="发送到">
              <el-select v-model="producePartitionMode">
                <el-option label="自动分区" value="auto" />
                <el-option label="指定分区" value="manual" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col v-if="producePartitionMode === 'manual'" :span="8">
            <el-form-item label="Partition">
              <el-select v-model="produceForm.partition" style="width: 100%">
                <el-option v-for="partition in producePartitionOptions" :key="partition" :label="String(partition)" :value="partition" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Key 编码">
              <el-select v-model="produceForm.keyEncoding">
                <el-option label="普通文本" value="plain" />
                <el-option label="Base64" value="base64" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Value 编码">
              <el-select v-model="produceForm.valueEncoding">
                <el-option label="普通文本 / JSON" value="plain" />
                <el-option label="Base64" value="base64" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="Key">
          <el-input v-model="produceForm.key" placeholder="可为空" />
        </el-form-item>
        <el-form-item label="Value">
          <el-input
            v-model="produceForm.value"
            type="textarea"
            :rows="8"
            placeholder="输入要发送的消息体；如果是 JSON，会在消息浏览中自动格式化展示"
          />
        </el-form-item>

        <div class="editor-title">Headers</div>
        <div class="header-editor">
          <div v-for="(header, index) in produceHeaders" :key="index" class="header-row">
            <el-row :gutter="12">
              <el-col :span="7"><el-input v-model="header.key" placeholder="Header Key" /></el-col>
              <el-col :span="11"><el-input v-model="header.value" placeholder="Header Value" /></el-col>
              <el-col :span="4">
                <el-select v-model="header.valueEncoding">
                  <el-option label="文本" value="plain" />
                  <el-option label="Base64" value="base64" />
                </el-select>
              </el-col>
              <el-col :span="2" class="row-actions">
                <el-button link type="danger" @click="removeProduceHeader(index)">删除</el-button>
              </el-col>
            </el-row>
          </div>
          <el-button text type="primary" @click="addProduceHeader">新增 Header</el-button>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="produceDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="sending" @click="handleProduceMessage">发送</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailDrawerVisible" title="消息详情" size="55%">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="Offset">{{ activeMessage?.offset ?? '-' }}</el-descriptions-item>
        <el-descriptions-item label="Partition">{{ activeMessage?.partition ?? '-' }}</el-descriptions-item>
        <el-descriptions-item label="时间">{{ formatTime(activeMessage?.timestamp) }}</el-descriptions-item>
        <el-descriptions-item label="Headers">{{ activeMessage?.headers?.length || 0 }}</el-descriptions-item>
      </el-descriptions>

      <div class="detail-section">
        <div class="section-title">Key 预览</div>
        <pre class="detail-pre">{{ activeMessage?.keyPreview || '(empty)' }}</pre>
      </div>
      <div class="detail-section">
        <div class="section-title">Value 预览</div>
        <pre class="detail-pre">{{ activeMessage?.valuePreview || '(empty)' }}</pre>
      </div>
      <div class="detail-section">
        <div class="section-title">Headers</div>
        <el-table :data="activeMessage?.headers || []" empty-text="暂无 Header">
          <el-table-column prop="key" label="Key" min-width="180" />
          <el-table-column prop="value" label="Value" min-width="300" show-overflow-tooltip />
        </el-table>
      </div>
      <div class="detail-section">
        <div class="section-title">Key Base64</div>
        <pre class="detail-pre">{{ activeMessage?.keyBase64 || '(empty)' }}</pre>
      </div>
      <div class="detail-section">
        <div class="section-title">Value Base64</div>
        <pre class="detail-pre">{{ activeMessage?.valueBase64 || '(empty)' }}</pre>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { ElMessage } from 'element-plus'
import {
  getKafkaAuditLogs,
  getKafkaMessages,
  getKafkaTopics,
  produceKafkaMessage,
} from '@/api/kafka.js'
import { useKafkaStore } from '@/stores/kafkaStore.js'
import { usePermissionStore } from '@/stores/permissionStore.js'
import { formatDateTime } from '@/utils/dateTime.js'
import { confirmKafkaRiskAction } from '@/utils/kafkaRiskConfirm.js'

const permStore = usePermissionStore()
const kafkaStore = useKafkaStore()

const loading = ref(false)
const sending = ref(false)
const topics = ref([])
const produceTopics = ref([])
const result = ref({ count: 0, startOffset: 0, messages: [], partial: false, timedOut: false, warningMessage: '' })
const hasQueried = ref(false)
const produceDialogVisible = ref(false)
const detailDrawerVisible = ref(false)
const activeMessage = ref(null)
const producePartitionMode = ref('auto')
const recentProduceLogs = ref([])
const produceTemplates = [
  {
    key: 'json-event',
    label: 'JSON 事件模板',
    description: '适合验证消费链路、日志和 JSON 解析展示。',
    payload: {
      keyEncoding: 'plain',
      valueEncoding: 'plain',
      key: 'demo.event',
      value: JSON.stringify({ event: 'demo.created', source: 'kafka-console', ts: Date.now() }, null, 2),
      headers: [{ key: 'content-type', value: 'application/json', valueEncoding: 'plain' }],
    },
  },
  {
    key: 'plain-text',
    label: '纯文本模板',
    description: '适合快速验证 Topic 连通性和消费者是否能收到消息。',
    payload: {
      keyEncoding: 'plain',
      valueEncoding: 'plain',
      key: '',
      value: 'hello from kafka-console',
      headers: [],
    },
  },
  {
    key: 'base64',
    label: 'Base64 模板',
    description: '适合验证二进制内容或 Base64 编码场景。',
    payload: {
      keyEncoding: 'plain',
      valueEncoding: 'base64',
      key: 'binary-demo',
      value: 'aGVsbG8ga2Fma2E=',
      headers: [],
    },
  },
]
const { clusterOptions: clusters, selectedClusterId: sharedClusterId } = storeToRefs(kafkaStore)

const form = reactive({
  clusterId: null,
  topic: '',
  partition: 0,
  mode: 'latest',
  offset: null,
  limit: 50,
  keyword: '',
})

const produceForm = reactive({
  clusterId: null,
  topic: '',
  partition: 0,
  key: '',
  keyEncoding: 'plain',
  value: '',
  valueEncoding: 'plain',
})

const produceHeaders = ref([])

const handledErrors = new WeakSet()
const markErrorHandled = (error) => {
  if (error && typeof error === 'object') {
    handledErrors.add(error)
  }
  return error
}

const currentClusterName = computed(
  () => clusters.value.find((item) => item.id === form.clusterId)?.name || '-',
)

const partitionOptions = computed(() => {
  const item = topics.value.find((topic) => topic.name === form.topic)
  const count = item?.partitions || 0
  return Array.from({ length: count }, (_, index) => index)
})

const currentTopicHints = computed(() => {
  const hints = []
  const topicMeta = topics.value.find((topic) => topic.name === form.topic)
  const selectedTopicName = form.topic

  if (currentClusterName.value && /(^|[-_\s])(prod|生产)([-_\s]|$)/i.test(currentClusterName.value)) {
    hints.push({
      title: '生产环境',
      description: '当前集群名称看起来像生产环境，浏览和发送测试消息前都应确认不会影响真实业务。',
      level: 'high',
    })
  }

  if (topicMeta?.internal) {
    hints.push({
      title: '内部 Topic',
      description: `${selectedTopicName} 是内部 Topic，不建议在没有明确目的时直接发送消息。`,
      level: 'high',
    })
  }

  if (result.value.warningMessage) {
    hints.push({
      title: result.value.timedOut ? '结果已截断' : '浏览提醒',
      description: result.value.warningMessage,
      level: 'warning',
    })
  }

  if (hasQueried.value && Number(result.value.count || 0) === 0 && selectedTopicName) {
    hints.push({
      title: '暂无采样结果',
      description: '当前条件下没有取到消息，建议确认分区、起始 Offset 或筛选关键词。',
      level: 'warning',
    })
  }

  if (result.value.count >= Number(form.limit || 0) && Number(form.limit || 0) > 0) {
    hints.push({
      title: '结果已触顶',
      description: `当前返回 ${result.value.count} 条消息，可能还有更多结果未展示，可调整采样条数继续查看。`,
      level: 'warning',
    })
  }

  if (hints.length === 0) {
    hints.push({
      title: '当前状态',
      description: '当前 Topic 没有识别到明显风险信号，可以继续查看消息明细。',
      level: 'success',
    })
  }

  return hints
})

const reusableProduceLogs = computed(() =>
  recentProduceLogs.value
    .map((item) => {
      let parsedPayload = null
      try {
        parsedPayload = item.requestPayload ? JSON.parse(item.requestPayload) : null
      } catch {
        parsedPayload = null
      }
      return {
        ...item,
        parsedPayload,
        topic: parsedPayload?.topic || '',
        summary: parsedPayload
          ? `Key 编码 ${parsedPayload.keyEncoding || 'plain'} / Value 编码 ${parsedPayload.valueEncoding || 'plain'} / Header ${parsedPayload.headerCount || 0} 个`
          : '未能解析请求参数',
      }
    })
    .filter((item) => item.parsedPayload)
    .slice(0, 5),
)

const producePartitionOptions = computed(() => {
  const item = produceTopics.value.find((topic) => topic.name === produceForm.topic)
  const count = item?.partitions || 0
  return Array.from({ length: count }, (_, index) => index)
})

const formatTime = formatDateTime

const loadClusters = async () => {
  try {
    await kafkaStore.loadClusterOptions()
    if (!form.clusterId || !clusters.value.some((item) => item.id === form.clusterId)) {
      form.clusterId = sharedClusterId.value || clusters.value[0]?.id || null
    }
    if (!produceForm.clusterId || !clusters.value.some((item) => item.id === produceForm.clusterId)) {
      produceForm.clusterId = form.clusterId
    }
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 集群列表加载失败')
    throw markErrorHandled(error)
  }
}

const loadTopics = async () => {
  if (!form.clusterId) return
  try {
    const res = await getKafkaTopics({ clusterId: form.clusterId })
    topics.value = res?.data?.data || []
    if (!form.topic && topics.value.length > 0) {
      form.topic = topics.value[0].name
    }
  } catch (error) {
    ElMessage.error(error.message || 'Topic 列表加载失败')
    topics.value = []
    form.topic = ''
    form.partition = 0
  }
}

const loadProduceTopics = async () => {
  if (!produceForm.clusterId) return
  try {
    const res = await getKafkaTopics({ clusterId: produceForm.clusterId })
    produceTopics.value = res?.data?.data || []
    if (!produceForm.topic && produceTopics.value.length > 0) {
      produceForm.topic = produceTopics.value[0].name
    }
  } catch (error) {
    ElMessage.error(error.message || '发送 Topic 列表加载失败')
    produceTopics.value = []
    produceForm.topic = ''
    produceForm.partition = 0
  }
}

const loadRecentProduceLogs = async () => {
  if (!form.clusterId) {
    recentProduceLogs.value = []
    return
  }
  try {
    const res = await getKafkaAuditLogs({
      clusterId: form.clusterId,
      action: 'message:produce',
      page: 1,
      pageSize: 6,
    })
    recentProduceLogs.value = res?.data?.data?.list || []
  } catch {
    recentProduceLogs.value = []
  }
}

const handleClusterChange = async () => {
  kafkaStore.setSelectedClusterId(form.clusterId)
  form.topic = ''
  form.partition = 0
  form.offset = null
  hasQueried.value = false
  result.value = { count: 0, startOffset: 0, messages: [], partial: false, timedOut: false, warningMessage: '' }
  await loadTopics()
  await loadRecentProduceLogs()
}

const handleTopicChange = () => {
  form.partition = 0
  form.offset = null
  hasQueried.value = false
  result.value = { count: 0, startOffset: 0, messages: [], partial: false, timedOut: false, warningMessage: '' }
}

const handleProduceClusterChange = async () => {
  produceForm.topic = ''
  produceForm.partition = 0
  await loadProduceTopics()
}

const loadMessages = async () => {
  if (loading.value || !form.clusterId || !form.topic) return
  if (partitionOptions.value.length === 0) return
  if (!partitionOptions.value.includes(Number(form.partition))) {
    form.partition = partitionOptions.value[0]
  }
  const offsetValue = Number(form.offset)
  if (form.mode === 'offset' && (form.offset === null || form.offset === undefined || Number.isNaN(offsetValue))) {
    ElMessage.warning('请输入有效的 Offset')
    return
  }
  hasQueried.value = true
  loading.value = true
  try {
    const res = await getKafkaMessages({
      ...form,
      offset: form.mode === 'offset' ? offsetValue : 0,
    })
    const data = res?.data?.data
    result.value = data || result.value
    if (data?.warningMessage) {
      ElMessage.warning(data.warningMessage)
    }
  } catch (error) {
    ElMessage.error(error.message || '消息浏览失败')
  } finally {
    loading.value = false
  }
}

const openProduceDialog = async () => {
  produceForm.clusterId = form.clusterId
  produceForm.topic = form.topic
  produceForm.partition = form.partition
  produceForm.key = ''
  produceForm.keyEncoding = 'plain'
  produceForm.value = ''
  produceForm.valueEncoding = 'plain'
  produceHeaders.value = []
  producePartitionMode.value = 'auto'
  await loadProduceTopics()
  produceDialogVisible.value = true
}

const hydrateProduceForm = async ({
  clusterId = form.clusterId,
  topic = form.topic,
  partition = form.partition,
  key = '',
  keyEncoding = 'plain',
  value = '',
  valueEncoding = 'plain',
  headers = [],
  partitionMode = 'auto',
}) => {
  produceForm.clusterId = clusterId
  produceForm.topic = topic
  produceForm.partition = partition
  produceForm.key = key
  produceForm.keyEncoding = keyEncoding
  produceForm.value = value
  produceForm.valueEncoding = valueEncoding
  producePartitionMode.value = partitionMode
  produceHeaders.value = headers.length > 0 ? headers : []
  await loadProduceTopics()
  if (!produceTopics.value.find((item) => item.name === produceForm.topic) && produceTopics.value.length > 0) {
    produceForm.topic = produceTopics.value[0].name
  }
}

const applyProduceTemplate = async (template) => {
  await hydrateProduceForm({
    ...template.payload,
  })
  produceDialogVisible.value = true
}

const reuseProduceLog = async (log) => {
  const payload = log.parsedPayload
  if (!payload) {
    ElMessage.warning('该记录缺少可复用的发送参数')
    return
  }
  await hydrateProduceForm({
    clusterId: payload.clusterId || form.clusterId,
    topic: payload.topic || form.topic,
    partition: Number(payload.partition || 0),
    key: payload.key || '',
    keyEncoding: payload.keyEncoding || 'plain',
    value: '',
    valueEncoding: payload.valueEncoding || 'plain',
    headers: [],
    partitionMode: Number.isInteger(payload.partition) ? 'manual' : 'auto',
  })
  ElMessage.info('已复用最近发送的参数，请补充消息体后发送')
  produceDialogVisible.value = true
}

const addProduceHeader = () => {
  produceHeaders.value.push({ key: '', value: '', valueEncoding: 'plain' })
}

const removeProduceHeader = (index) => {
  produceHeaders.value.splice(index, 1)
}

const handleProduceMessage = async () => {
  if (!produceForm.clusterId || !produceForm.topic || !produceForm.value) {
    ElMessage.warning('请填写集群、Topic 和消息体')
    return
  }
  if (producePartitionMode.value === 'manual' && !producePartitionOptions.value.includes(Number(produceForm.partition))) {
    ElMessage.warning('请选择有效的 Partition')
    return
  }
  const headers = produceHeaders.value
    .filter((item) => item.key && item.key.trim())
    .map((item) => ({
      key: item.key.trim(),
      value: item.value ?? '',
      valueEncoding: item.valueEncoding || 'plain',
    }))
  const payload = {
    clusterId: produceForm.clusterId,
    topic: produceForm.topic,
    key: produceForm.key,
    keyEncoding: produceForm.keyEncoding,
    value: produceForm.value,
    valueEncoding: produceForm.valueEncoding,
    headers,
  }
  if (producePartitionMode.value === 'manual') {
    payload.partition = Number(produceForm.partition)
  }
  const confirmed = await confirmKafkaRiskAction({
    title: '发送消息确认',
    resourceName: `${currentClusterName.value} / ${produceForm.topic}`,
    actionLabel: '发送测试消息',
    dangerPoints: [
      producePartitionMode.value === 'manual'
        ? `消息会发送到指定分区 ${produceForm.partition}`
        : '消息会由 Kafka 自动选择分区',
      `消息体长度约 ${produceForm.value.length} 个字符`,
      '如果当前是生产集群，请确认这条消息不会触发真实业务副作用',
    ],
    confirmButtonText: '确认发送',
  })
  if (!confirmed) return
  sending.value = true
  try {
    const res = await produceKafkaMessage(payload)
    const resultData = res?.data?.data
    ElMessage.success(`消息已发送到分区 ${resultData?.partition ?? '-'}，Offset ${resultData?.offset ?? '-'}`)
    produceDialogVisible.value = false
    if (
      form.clusterId === produceForm.clusterId &&
      form.topic === produceForm.topic &&
      (producePartitionMode.value === 'auto' || form.partition === Number(produceForm.partition))
    ) {
      await loadMessages()
    }
    if (form.clusterId === produceForm.clusterId) {
      await loadRecentProduceLogs()
    }
  } catch (error) {
    ElMessage.error(error.message || '消息发送失败')
  } finally {
    sending.value = false
  }
}

const openMessageDetail = (row) => {
  activeMessage.value = row
  detailDrawerVisible.value = true
}

onMounted(async () => {
  try {
    await loadClusters()
    await loadTopics()
    await loadRecentProduceLogs()
    await loadMessages()
  } catch (error) {
    if (!(error && typeof error === 'object' && handledErrors.has(error))) {
      ElMessage.error(error.message || '初始化消息浏览失败')
    }
  }
})

watch(sharedClusterId, async (nextValue, previousValue) => {
  if (!nextValue || nextValue === previousValue) return
  form.clusterId = nextValue
  produceForm.clusterId = nextValue
  await handleClusterChange()
  await handleProduceClusterChange()
})
</script>

<style scoped>
.secondary-toolbar {
  margin-top: 14px;
}

.produce-helper-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.produce-helper-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.produce-helper-card {
  appearance: none;
  -webkit-appearance: none;
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
  padding: 12px 14px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.92);
  color: inherit;
  text-align: left;
  cursor: pointer;
  transition: border-color 0.18s ease, box-shadow 0.18s ease, transform 0.18s ease;
}

.produce-helper-card:hover {
  transform: translateY(-1px);
  border-color: rgba(var(--shell-accent-rgb, 47, 125, 246), 0.26);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.produce-helper-card strong {
  color: var(--shell-text);
  font-size: 13px;
  font-weight: 700;
}

.produce-helper-card span {
  color: var(--shell-text-soft);
  font-size: 12px;
  line-height: 1.5;
}

.table-alert {
  margin-bottom: 16px;
}

@media (max-width: 960px) {
  .produce-helper-grid {
    grid-template-columns: 1fr;
  }
}
</style>
