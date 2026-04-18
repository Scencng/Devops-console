<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>Topic 管理</h2>
          <p>创建 Topic、扩分区、查看 ISR/副本分配，并执行配置修改等高风险操作</p>
        </div>
        <div class="header-actions">
          <el-select
            v-model="selectedClusterId"
            placeholder="选择 Kafka 集群"
            style="width: 260px"
            @change="loadTopics"
          >
            <el-option
              v-for="cluster in clusters"
              :key="cluster.id"
              :label="cluster.name"
              :value="cluster.id"
            />
          </el-select>
          <el-input
            v-model="keyword"
            placeholder="搜索 Topic"
            style="width: 220px"
            clearable
            @keyup.enter="loadTopics"
          />
          <el-button @click="loadTopics" :loading="loading">刷新</el-button>
          <el-button
            v-if="permStore.hasPerm('kafka:topic:create') || permStore.roles.includes('admin')"
            type="primary"
            @click="openCreateDialog"
          >
            创建 Topic
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <el-table :data="topics" empty-text="暂无 Topic 数据">
        <el-table-column prop="name" label="Topic" min-width="220" />
        <el-table-column prop="partitions" label="分区数" width="110" />
        <el-table-column prop="replicationFactor" label="副本数" width="110" />
        <el-table-column prop="cleanupPolicy" label="清理策略" width="140" />
        <el-table-column prop="retentionMs" label="保留时间(ms)" width="160" />
        <el-table-column prop="minInSyncReplicas" label="Min ISR" width="120" />
        <el-table-column label="内部 Topic" width="120">
          <template #default="{ row }">
            <el-tag :type="row.internal ? 'warning' : 'success'">
              {{ row.internal ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="360" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openPartitionsDrawer(row)">ISR / 副本</el-button>
            <el-button
              link
              type="primary"
              @click="openExpandDialog(row)"
              v-if="permStore.hasPerm('kafka:topic:partitions:increase') || permStore.roles.includes('admin')"
            >
              扩分区
            </el-button>
            <el-button
              link
              type="primary"
              @click="openConfigDialog(row)"
              v-if="permStore.hasPerm('kafka:topic:config:update') || permStore.roles.includes('admin')"
            >
              修改配置
            </el-button>
            <el-button
              link
              type="danger"
              :disabled="row.internal"
              @click="handleDelete(row)"
              v-if="permStore.hasPerm('kafka:topic:delete') || permStore.roles.includes('admin')"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="创建 Topic" width="760px" destroy-on-close>
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-position="top"
      >
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Topic 名称" prop="name">
              <el-input v-model="createForm.name" placeholder="例如 orders.events" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="分区数" prop="numPartitions">
              <el-input-number v-model="createForm.numPartitions" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="副本数" prop="replicationFactor">
              <el-input-number v-model="createForm.replicationFactor" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <div class="editor-title">初始配置</div>
      <div class="config-editor">
        <div v-for="(entry, index) in createConfigRows" :key="index" class="config-row">
          <el-row :gutter="12">
            <el-col :span="10">
              <el-input v-model="entry.key" placeholder="配置项，例如 retention.ms" />
            </el-col>
            <el-col :span="10">
              <el-input v-model="entry.value" placeholder="配置值" />
            </el-col>
            <el-col :span="4" class="row-actions">
              <el-button link type="danger" @click="removeCreateConfigRow(index)">删除</el-button>
            </el-col>
          </el-row>
        </div>
        <el-button text type="primary" @click="addCreateConfigRow">新增配置项</el-button>
      </div>

      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleCreateTopic">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="configDialogVisible" :title="`修改 Topic 配置: ${activeTopic?.name || ''}`" width="760px" destroy-on-close>
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        title="这里的修改会直接写入 Kafka Topic 配置，请确认你了解变更影响。"
      />
      <div class="config-editor">
        <div v-for="(entry, index) in configRows" :key="index" class="config-row">
          <el-row :gutter="12">
            <el-col :span="9">
              <el-input v-model="entry.key" placeholder="配置项，例如 retention.ms" />
            </el-col>
            <el-col :span="5">
              <el-select v-model="entry.operation" style="width: 100%">
                <el-option label="设置" value="set" />
                <el-option label="删除" value="delete" />
              </el-select>
            </el-col>
            <el-col :span="8">
              <el-input v-model="entry.value" :disabled="entry.operation === 'delete'" placeholder="配置值" />
            </el-col>
            <el-col :span="2" class="row-actions">
              <el-button link type="danger" @click="removeConfigRow(index)">删除</el-button>
            </el-col>
          </el-row>
        </div>
        <el-button text type="primary" @click="addConfigRow">新增配置项</el-button>
      </div>
      <template #footer>
        <el-button @click="configDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleUpdateConfig" :loading="saving">保存配置</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="expandDialogVisible" :title="`扩容 Topic 分区: ${expandForm.topic}`" width="520px" destroy-on-close>
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        title="Kafka 只支持增加分区，不支持减少分区。扩分区后请检查生产者分区策略和消费者并行度。"
      />
      <el-form label-position="top" class="expand-form">
        <el-form-item label="当前分区数">
          <el-input :model-value="String(expandForm.currentPartitions)" disabled />
        </el-form-item>
        <el-form-item label="目标分区数">
          <el-input-number v-model="expandForm.count" :min="expandForm.currentPartitions + 1" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="expandDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleIncreasePartitions">确认扩容</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="partitionsDrawerVisible" :title="`ISR / 副本分配: ${partitionDetail.topic || ''}`" size="60%">
      <el-skeleton :loading="partitionsLoading" animated :rows="6">
        <template #default>
          <el-row :gutter="16" class="partition-summary">
            <el-col :span="8">
              <el-card>
                <div class="summary-item">
                  <span>分区总数</span>
                  <strong>{{ partitionDetail.partitionCount || 0 }}</strong>
                </div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card>
                <div class="summary-item">
                  <span>副本异常分区</span>
                  <strong>{{ partitionDetail.underReplicatedCount || 0 }}</strong>
                </div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card>
                <div class="summary-item">
                  <span>Topic</span>
                  <strong>{{ partitionDetail.topic || '-' }}</strong>
                </div>
              </el-card>
            </el-col>
          </el-row>

          <el-table :data="partitionDetail.partitions || []" empty-text="暂无分区明细">
            <el-table-column prop="partition" label="分区" width="90" />
            <el-table-column prop="leader" label="Leader" width="100" />
            <el-table-column label="Replicas" min-width="170">
              <template #default="{ row }">{{ formatIntList(row.replicas) }}</template>
            </el-table-column>
            <el-table-column label="ISR" min-width="150">
              <template #default="{ row }">{{ formatIntList(row.isr) }}</template>
            </el-table-column>
            <el-table-column label="掉队副本" min-width="170">
              <template #default="{ row }">
                <span>{{ formatIntList(row.outOfSyncReplicas) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="离线副本" min-width="170">
              <template #default="{ row }">{{ formatIntList(row.offlineReplicas) }}</template>
            </el-table-column>
            <el-table-column prop="latestOffset" label="最新 Offset" width="130" />
            <el-table-column prop="messageCountEstimate" label="消息量估算" width="130" />
            <el-table-column label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="row.underReplicated ? 'danger' : 'success'">
                  {{ row.underReplicated ? '异常' : '正常' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </template>
      </el-skeleton>
    </el-drawer>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createKafkaTopic,
  deleteKafkaTopic,
  getKafkaClusterOptions,
  getKafkaTopicPartitions,
  getKafkaTopics,
  increaseKafkaTopicPartitions,
  updateKafkaTopicConfig,
} from '@/api/kafka.js'
import { usePermissionStore } from '@/stores/permissionStore.js'

const permStore = usePermissionStore()

const loading = ref(false)
const saving = ref(false)
const partitionsLoading = ref(false)
const clusters = ref([])
const selectedClusterId = ref(null)
const keyword = ref('')
const topics = ref([])

const createDialogVisible = ref(false)
const configDialogVisible = ref(false)
const expandDialogVisible = ref(false)
const partitionsDrawerVisible = ref(false)

const createFormRef = ref()
const activeTopic = ref(null)

const createForm = reactive({
  name: '',
  numPartitions: 3,
  replicationFactor: 1,
})
const createConfigRows = ref([{ key: 'cleanup.policy', value: 'delete' }])

const configRows = ref([])

const expandForm = reactive({
  topic: '',
  currentPartitions: 0,
  count: 1,
})

const partitionDetail = ref({
  topic: '',
  partitionCount: 0,
  underReplicatedCount: 0,
  partitions: [],
})

const createRules = {
  name: [{ required: true, message: '请输入 Topic 名称', trigger: 'blur' }],
  numPartitions: [{ required: true, message: '请输入分区数', trigger: 'change' }],
  replicationFactor: [{ required: true, message: '请输入副本数', trigger: 'change' }],
}

const emptyConfigRow = () => ({ key: '', operation: 'set', value: '' })

const resetCreateForm = () => {
  createForm.name = ''
  createForm.numPartitions = 3
  createForm.replicationFactor = 1
  createConfigRows.value = [{ key: 'cleanup.policy', value: 'delete' }]
}

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) {
    selectedClusterId.value = clusters.value[0].id
  }
}

const loadTopics = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const res = await getKafkaTopics({ clusterId: selectedClusterId.value, keyword: keyword.value })
    topics.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || 'Topic 数据加载失败')
  } finally {
    loading.value = false
  }
}

const addCreateConfigRow = () => {
  createConfigRows.value.push({ key: '', value: '' })
}

const removeCreateConfigRow = (index) => {
  if (createConfigRows.value.length === 1) {
    createConfigRows.value[0] = { key: '', value: '' }
    return
  }
  createConfigRows.value.splice(index, 1)
}

const openCreateDialog = () => {
  resetCreateForm()
  createDialogVisible.value = true
}

const handleCreateTopic = async () => {
  if (!createFormRef.value || !selectedClusterId.value) return
  await createFormRef.value.validate()
  const configEntries = createConfigRows.value
    .filter((item) => item.key && item.key.trim())
    .map((item) => ({ key: item.key.trim(), value: String(item.value ?? '') }))
  saving.value = true
  try {
    await createKafkaTopic({
      clusterId: selectedClusterId.value,
      name: createForm.name.trim(),
      numPartitions: Number(createForm.numPartitions),
      replicationFactor: Number(createForm.replicationFactor),
      configEntries,
    })
    ElMessage.success('Topic 创建成功')
    createDialogVisible.value = false
    await loadTopics()
  } catch (error) {
    ElMessage.error(error.message || 'Topic 创建失败')
  } finally {
    saving.value = false
  }
}

const addConfigRow = () => {
  configRows.value.push(emptyConfigRow())
}

const removeConfigRow = (index) => {
  if (configRows.value.length === 1) {
    configRows.value[0] = emptyConfigRow()
    return
  }
  configRows.value.splice(index, 1)
}

const openConfigDialog = (row) => {
  activeTopic.value = row
  const rows = Object.entries(row.configEntries || {}).map(([key, value]) => ({
    key,
    operation: 'set',
    value: value ?? '',
  }))
  configRows.value = rows.length > 0 ? rows : [emptyConfigRow()]
  configDialogVisible.value = true
}

const handleUpdateConfig = async () => {
  if (!activeTopic.value || !selectedClusterId.value) return
  const entries = configRows.value
    .filter((item) => item.key && item.key.trim())
    .map((item) =>
      item.operation === 'delete'
        ? { key: item.key.trim(), operation: 'delete' }
        : { key: item.key.trim(), operation: 'set', value: String(item.value ?? '') },
    )
  if (entries.length === 0) {
    ElMessage.warning('请至少填写一条配置项')
    return
  }
  saving.value = true
  try {
    await updateKafkaTopicConfig(activeTopic.value.name, {
      clusterId: selectedClusterId.value,
      entries,
    })
    ElMessage.success('Topic 配置已更新')
    configDialogVisible.value = false
    await loadTopics()
  } catch (error) {
    ElMessage.error(error.message || 'Topic 配置更新失败')
  } finally {
    saving.value = false
  }
}

const openExpandDialog = (row) => {
  expandForm.topic = row.name
  expandForm.currentPartitions = Number(row.partitions || 0)
  expandForm.count = expandForm.currentPartitions + 1
  expandDialogVisible.value = true
}

const handleIncreasePartitions = async () => {
  if (!selectedClusterId.value || !expandForm.topic) return
  saving.value = true
  try {
    await increaseKafkaTopicPartitions(expandForm.topic, {
      clusterId: selectedClusterId.value,
      count: Number(expandForm.count),
    })
    ElMessage.success('Topic 分区扩容成功')
    expandDialogVisible.value = false
    await loadTopics()
    await openPartitionsDrawer({ name: expandForm.topic })
  } catch (error) {
    ElMessage.error(error.message || 'Topic 分区扩容失败')
  } finally {
    saving.value = false
  }
}

const openPartitionsDrawer = async (row) => {
  if (!selectedClusterId.value) return
  partitionsDrawerVisible.value = true
  partitionsLoading.value = true
  try {
    const res = await getKafkaTopicPartitions(selectedClusterId.value, row.name)
    partitionDetail.value = res?.data?.data || partitionDetail.value
  } catch (error) {
    ElMessage.error(error.message || 'Topic 分区详情加载失败')
  } finally {
    partitionsLoading.value = false
  }
}

const handleDelete = async (row) => {
  if (row.internal) {
    ElMessage.warning('内部 Topic 不允许删除')
    return
  }
  await ElMessageBox.confirm(`确认删除 Topic ${row.name} 吗？该操作不可恢复。`, '危险操作确认', {
    type: 'warning',
  })
  try {
    await deleteKafkaTopic(selectedClusterId.value, row.name)
    ElMessage.success('Topic 已删除')
    await loadTopics()
  } catch (error) {
    ElMessage.error(error.message || 'Topic 删除失败')
  }
}

const formatIntList = (value) => {
  if (!Array.isArray(value) || value.length === 0) return '-'
  return value.join(', ')
}

onMounted(async () => {
  try {
    await loadClusters()
    await loadTopics()
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 集群加载失败')
  }
})
</script>

<style scoped>
.config-editor {
  margin-top: 16px;
}

.config-row {
  margin-bottom: 12px;
}

.row-actions {
  display: flex;
  align-items: center;
}

.editor-title {
  margin-top: 8px;
  margin-bottom: 12px;
  font-weight: 600;
}

.expand-form {
  margin-top: 16px;
}

.partition-summary {
  margin-bottom: 16px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.summary-item span {
  color: #909399;
}

.summary-item strong {
  font-size: 26px;
  font-weight: 700;
}
</style>
