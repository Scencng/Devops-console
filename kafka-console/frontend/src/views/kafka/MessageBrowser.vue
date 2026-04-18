<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <div class="page-eyebrow">Messages</div>
          <h2>消息浏览</h2>
          <p>按 Topic / Partition / Offset 抽样查看消息，并在需要时直接发送测试消息进行验证。</p>
        </div>
        <div class="header-actions">
          <el-button @click="loadMessages" :loading="loading">查询消息</el-button>
          <el-button
            v-if="permStore.hasPerm('kafka:message:produce') || permStore.roles.includes('admin')"
            type="primary"
            @click="openProduceDialog"
          >
            发送消息
          </el-button>
        </div>
      </div>
    </el-card>

    <div class="page-metrics">
      <div class="page-metric-card is-accent">
        <span>当前集群</span>
        <strong>{{ currentClusterName }}</strong>
        <p>本页当前正在浏览消息的 Kafka 集群。</p>
      </div>
      <div class="page-metric-card">
        <span>当前 Topic</span>
        <strong>{{ form.topic || '-' }}</strong>
        <p>当前正在查询的 Topic。</p>
      </div>
      <div class="page-metric-card is-success">
        <span>当前分区</span>
        <strong>{{ form.partition }}</strong>
        <p>消息浏览默认使用的目标分区。</p>
      </div>
      <div class="page-metric-card is-warning">
        <span>返回消息数</span>
        <strong>{{ result.count || 0 }}</strong>
        <p>最近一次查询返回的消息条数。</p>
      </div>
    </div>

    <el-card class="content-card filter-card">
      <div class="toolbar-row">
        <div class="toolbar-left">
          <el-select v-model="form.clusterId" style="width: 240px" @change="handleClusterChange">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
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
      </div>

      <div class="toolbar-row secondary-toolbar">
        <div class="toolbar-left">
          <el-input v-if="form.mode === 'offset'" v-model="form.offset" placeholder="指定 Offset" style="width: 180px" />
          <el-input v-model="form.keyword" placeholder="按 key/value 过滤" style="width: 260px" />
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>消息列表</span>
          <span class="card-subtitle">共 {{ result.count || 0 }} 条消息，起始 Offset {{ result.startOffset ?? '-' }}</span>
        </div>
      </template>

      <el-table :data="result.messages || []" empty-text="暂无消息数据" height="600">
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
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="集群">
              <el-select v-model="produceForm.clusterId" @change="handleProduceClusterChange">
                <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
              </el-select>
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
              <el-input-number v-model="produceForm.partition" :min="0" style="width: 100%" />
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
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  getKafkaClusterOptions,
  getKafkaMessages,
  getKafkaTopics,
  produceKafkaMessage,
} from '@/api/kafka.js'
import { usePermissionStore } from '@/stores/permissionStore.js'

const permStore = usePermissionStore()

const loading = ref(false)
const sending = ref(false)
const clusters = ref([])
const topics = ref([])
const produceTopics = ref([])
const result = ref({ count: 0, startOffset: 0, messages: [] })
const produceDialogVisible = ref(false)
const detailDrawerVisible = ref(false)
const activeMessage = ref(null)
const producePartitionMode = ref('auto')

const form = reactive({
  clusterId: null,
  topic: '',
  partition: 0,
  mode: 'latest',
  offset: 0,
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

const produceHeaders = ref([{ key: '', value: '', valueEncoding: 'plain' }])

const currentClusterName = computed(
  () => clusters.value.find((item) => item.id === form.clusterId)?.name || '-',
)

const partitionOptions = computed(() => {
  const item = topics.value.find((topic) => topic.name === form.topic)
  const count = item?.partitions || 0
  return Array.from({ length: count }, (_, index) => index)
})

const producePartitionOptions = computed(() => {
  const item = produceTopics.value.find((topic) => topic.name === produceForm.topic)
  const count = item?.partitions || 0
  return Array.from({ length: count }, (_, index) => index)
})

const formatTime = (value) => (value ? new Date(value).toLocaleString() : '-')

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!form.clusterId && clusters.value.length > 0) {
    form.clusterId = clusters.value[0].id
  }
  if (!produceForm.clusterId && clusters.value.length > 0) {
    produceForm.clusterId = clusters.value[0].id
  }
}

const loadTopics = async () => {
  if (!form.clusterId) return
  const res = await getKafkaTopics({ clusterId: form.clusterId })
  topics.value = res?.data?.data || []
  if (!form.topic && topics.value.length > 0) {
    form.topic = topics.value[0].name
  }
}

const loadProduceTopics = async () => {
  if (!produceForm.clusterId) return
  const res = await getKafkaTopics({ clusterId: produceForm.clusterId })
  produceTopics.value = res?.data?.data || []
  if (!produceForm.topic && produceTopics.value.length > 0) {
    produceForm.topic = produceTopics.value[0].name
  }
}

const handleClusterChange = async () => {
  form.topic = ''
  form.partition = 0
  await loadTopics()
}

const handleTopicChange = () => {
  form.partition = 0
}

const handleProduceClusterChange = async () => {
  produceForm.topic = ''
  produceForm.partition = 0
  await loadProduceTopics()
}

const loadMessages = async () => {
  if (!form.clusterId || !form.topic) return
  if (partitionOptions.value.length === 0) return
  if (!partitionOptions.value.includes(Number(form.partition))) {
    form.partition = partitionOptions.value[0]
  }
  loading.value = true
  try {
    const res = await getKafkaMessages(form)
    result.value = res?.data?.data || result.value
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
  produceHeaders.value = [{ key: '', value: '', valueEncoding: 'plain' }]
  producePartitionMode.value = 'auto'
  await loadProduceTopics()
  produceDialogVisible.value = true
}

const addProduceHeader = () => {
  produceHeaders.value.push({ key: '', value: '', valueEncoding: 'plain' })
}

const removeProduceHeader = (index) => {
  if (produceHeaders.value.length === 1) {
    produceHeaders.value[0] = { key: '', value: '', valueEncoding: 'plain' }
    return
  }
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
    await loadMessages()
  } catch (error) {
    ElMessage.error(error.message || '初始化消息浏览失败')
  }
})
</script>

<style scoped>
.secondary-toolbar {
  margin-top: 14px;
}

.editor-title {
  margin-top: 8px;
  margin-bottom: 12px;
  font-weight: 600;
}

.header-editor {
  margin-top: 12px;
}

.header-row {
  margin-bottom: 12px;
}

.row-actions {
  display: flex;
  align-items: center;
}

.detail-section {
  margin-top: 20px;
}

.section-title {
  margin-bottom: 8px;
  font-weight: 600;
}

.detail-pre {
  padding: 12px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  background: #0f172a;
  border-radius: 10px;
  color: #e2e8f0;
}
</style>
