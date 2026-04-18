<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <div class="page-eyebrow">Consumer Groups</div>
          <h2>消费组管理</h2>
          <p>观察消费组状态、Lag 与分区明细，并在必要时执行 Offset 干预。</p>
        </div>
      </div>
    </el-card>

    <div class="page-metrics">
      <div class="page-metric-card is-accent">
        <span>当前集群</span>
        <strong>{{ currentClusterName }}</strong>
        <p>本页当前正在观测的 Kafka 集群。</p>
      </div>
      <div class="page-metric-card">
        <span>消费组数量</span>
        <strong>{{ groupStats.total }}</strong>
        <p>当前筛选条件下返回的消费组数。</p>
      </div>
      <div class="page-metric-card is-success">
        <span>稳定状态</span>
        <strong>{{ groupStats.stable }}</strong>
        <p>状态为 Stable 的消费组数量。</p>
      </div>
      <div class="page-metric-card is-warning">
        <span>总 Lag</span>
        <strong>{{ groupStats.totalLag }}</strong>
        <p>所有消费组累计的已提交 Lag。</p>
      </div>
    </div>

    <el-card class="content-card filter-card">
      <div class="toolbar-row">
        <div class="toolbar-left">
          <el-select
            v-model="selectedClusterId"
            placeholder="选择 Kafka 集群"
            style="width: 300px"
            @change="loadGroups"
          >
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-input
            v-model="keyword"
            placeholder="搜索消费组"
            style="width: 240px"
            clearable
            @keyup.enter="loadGroups"
          />
        </div>
        <div class="toolbar-right">
          <el-button @click="loadGroups" :loading="loading">刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>消费组列表</span>
          <span class="card-subtitle">先看状态与 Lag，再进入分区级明细处理问题</span>
        </div>
      </template>

      <el-table :data="groups" empty-text="暂无 Consumer Group 数据">
        <el-table-column prop="groupId" label="消费组" min-width="220" />
        <el-table-column prop="state" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.state === 'Stable' ? 'success' : 'warning'">{{ row.state || 'Unknown' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="protocolType" label="协议类型" width="140" />
        <el-table-column prop="memberCount" label="成员数" width="100" />
        <el-table-column prop="partitionCount" label="分区数" width="100" />
        <el-table-column prop="committedLag" label="Lag" width="160" />
        <el-table-column label="Topics" min-width="240">
          <template #default="{ row }">{{ (row.topics || []).join(', ') || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" min-width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetailDrawer(row)">查看明细</el-button>
            <el-button
              v-if="permStore.hasPerm('kafka:group:offset:reset') || permStore.roles.includes('admin')"
              link
              type="danger"
              :disabled="!row.topics || row.topics.length === 0"
              @click="openResetDialog(row)"
            >
              重置 Offset
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-drawer v-model="detailDrawerVisible" :title="`消费组详情: ${detailData.groupId || ''}`" size="75%">
      <el-skeleton :loading="detailLoading" animated :rows="8">
        <template #default>
          <el-row :gutter="16" class="detail-summary">
            <el-col :span="6">
              <el-card>
                <div class="summary-item">
                  <span>成员数</span>
                  <strong>{{ detailData.memberCount || 0 }}</strong>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card>
                <div class="summary-item">
                  <span>分区数</span>
                  <strong>{{ detailData.partitionCount || 0 }}</strong>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card>
                <div class="summary-item">
                  <span>总 Lag</span>
                  <strong>{{ detailData.totalLag || 0 }}</strong>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card>
                <div class="summary-item">
                  <span>状态</span>
                  <strong>{{ detailData.state || '-' }}</strong>
                </div>
              </el-card>
            </el-col>
          </el-row>

          <el-card class="detail-card">
            <template #header>
              <div class="card-header">成员分布</div>
            </template>
            <el-table :data="detailData.members || []" empty-text="暂无成员信息">
              <el-table-column prop="memberId" label="Member ID" min-width="220" />
              <el-table-column prop="clientId" label="Client ID" min-width="180" />
              <el-table-column prop="clientHost" label="Client Host" min-width="160" />
              <el-table-column label="Topics" min-width="220">
                <template #default="{ row }">{{ (row.topics || []).join(', ') || '-' }}</template>
              </el-table-column>
            </el-table>
          </el-card>

          <el-card class="detail-card">
            <template #header>
              <div class="card-header">分区级 Lag 明细</div>
            </template>
            <el-table :data="detailData.partitions || []" empty-text="暂无分区明细">
              <el-table-column prop="topic" label="Topic" min-width="200" />
              <el-table-column prop="partition" label="Partition" width="100" />
              <el-table-column prop="committedOffset" label="Committed" width="120" />
              <el-table-column prop="latestOffset" label="Latest" width="120" />
              <el-table-column prop="oldestOffset" label="Oldest" width="120" />
              <el-table-column prop="lag" label="Lag" width="120" />
              <el-table-column prop="memberId" label="Member ID" min-width="200" />
              <el-table-column prop="clientHost" label="Client Host" min-width="160" />
              <el-table-column label="操作" width="120" fixed="right">
                <template #default="{ row }">
                  <el-button
                    v-if="permStore.hasPerm('kafka:group:offset:reset') || permStore.roles.includes('admin')"
                    link
                    type="danger"
                    @click="openResetDialog(detailData, row)"
                  >
                    重置
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </template>
      </el-skeleton>
    </el-drawer>

    <el-dialog v-model="resetDialogVisible" title="重置 Consumer Group Offset" width="640px" destroy-on-close>
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        title="Offset 重置会直接影响消费位置，请确保相关消费者已暂停或你明确知道后果。"
      />
      <el-form ref="formRef" :model="resetForm" :rules="resetRules" label-position="top" class="offset-form">
        <el-form-item label="消费组">
          <el-input v-model="resetForm.groupId" disabled />
        </el-form-item>
        <el-form-item label="Topic" prop="topic">
          <el-select v-model="resetForm.topic" placeholder="请选择 Topic" style="width: 100%">
            <el-option v-for="topic in topicOptions" :key="topic" :label="topic" :value="topic" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-switch v-model="resetForm.allPartitions" />
          <span class="switch-label">应用到该 Topic 的全部分区</span>
        </el-form-item>
        <el-form-item v-if="!resetForm.allPartitions" label="Partition" prop="partition">
          <el-input-number v-model="resetForm.partition" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="重置方式" prop="resetType">
          <el-select v-model="resetForm.resetType" style="width: 100%">
            <el-option label="最早位置 (earliest)" value="earliest" />
            <el-option label="最新位置 (latest)" value="latest" />
            <el-option label="指定 Offset" value="offset" />
            <el-option label="按时间戳" value="timestamp" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="resetForm.resetType === 'offset'" label="指定 Offset" prop="offset">
          <el-input-number v-model="resetForm.offset" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item v-if="resetForm.resetType === 'timestamp'" label="按时间查找 Offset" prop="timestampMs">
          <el-date-picker
            v-model="resetForm.timestampMs"
            type="datetime"
            value-format="x"
            placeholder="选择时间"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="handleResetOffset" :loading="saving">确认重置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getKafkaClusterOptions,
  getKafkaConsumerGroupDetail,
  getKafkaConsumerGroups,
  resetKafkaGroupOffset,
} from '@/api/kafka.js'
import { usePermissionStore } from '@/stores/permissionStore.js'

const permStore = usePermissionStore()

const loading = ref(false)
const detailLoading = ref(false)
const saving = ref(false)
const clusters = ref([])
const groups = ref([])
const selectedClusterId = ref(null)
const keyword = ref('')
const detailDrawerVisible = ref(false)
const resetDialogVisible = ref(false)
const formRef = ref()
const activeGroup = ref(null)

const detailData = ref({
  groupId: '',
  memberCount: 0,
  partitionCount: 0,
  totalLag: 0,
  state: '',
  members: [],
  partitions: [],
  topics: [],
})

const resetForm = reactive({
  groupId: '',
  topic: '',
  allPartitions: false,
  partition: 0,
  resetType: 'earliest',
  offset: 0,
  timestampMs: '',
})

const currentClusterName = computed(
  () => clusters.value.find((item) => item.id === selectedClusterId.value)?.name || '-',
)

const groupStats = computed(() => ({
  total: groups.value.length,
  stable: groups.value.filter((item) => item.state === 'Stable').length,
  totalLag: groups.value.reduce((sum, item) => sum + Number(item.committedLag || 0), 0),
}))

const topicOptions = computed(() => activeGroup.value?.topics || detailData.value?.topics || [])

const resetRules = {
  topic: [{ required: true, message: '请选择 Topic', trigger: 'change' }],
  partition: [{ required: true, message: '请输入 Partition', trigger: 'change' }],
  resetType: [{ required: true, message: '请选择重置方式', trigger: 'change' }],
}

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) {
    selectedClusterId.value = clusters.value[0].id
  }
}

const loadGroups = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const res = await getKafkaConsumerGroups({
      clusterId: selectedClusterId.value,
      keyword: keyword.value,
    })
    groups.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || 'Consumer Group 数据加载失败')
  } finally {
    loading.value = false
  }
}

const loadGroupDetail = async (groupId) => {
  if (!selectedClusterId.value || !groupId) return
  detailLoading.value = true
  try {
    const res = await getKafkaConsumerGroupDetail(groupId, {
      clusterId: selectedClusterId.value,
    })
    detailData.value = res?.data?.data || detailData.value
  } catch (error) {
    ElMessage.error(error.message || '消费组明细加载失败')
  } finally {
    detailLoading.value = false
  }
}

const openDetailDrawer = async (row) => {
  activeGroup.value = row
  detailDrawerVisible.value = true
  await loadGroupDetail(row.groupId)
}

const openResetDialog = (group, partitionRow = null) => {
  activeGroup.value = group
  resetForm.groupId = group.groupId
  resetForm.topic = partitionRow?.topic || group.topics?.[0] || ''
  resetForm.allPartitions = false
  resetForm.partition = Number(partitionRow?.partition ?? 0)
  resetForm.resetType = 'earliest'
  resetForm.offset = 0
  resetForm.timestampMs = ''
  resetDialogVisible.value = true
}

const handleResetOffset = async () => {
  if (!formRef.value || !selectedClusterId.value) return
  await formRef.value.validate()
  if (!resetForm.allPartitions && (resetForm.partition === null || resetForm.partition === undefined)) {
    ElMessage.warning('请输入 Partition')
    return
  }
  if (resetForm.resetType === 'offset' && Number.isNaN(Number(resetForm.offset))) {
    ElMessage.warning('请输入有效的 Offset')
    return
  }
  if (resetForm.resetType === 'timestamp' && !resetForm.timestampMs) {
    ElMessage.warning('请选择时间')
    return
  }
  await ElMessageBox.confirm(
    `确认重置消费组 ${resetForm.groupId} 在 Topic ${resetForm.topic} 上的 Offset 吗？`,
    '危险操作确认',
    { type: 'warning' },
  )
  saving.value = true
  try {
    const payload = {
      clusterId: selectedClusterId.value,
      topic: resetForm.topic,
      allPartitions: resetForm.allPartitions,
      resetType: resetForm.resetType,
      offset: Number(resetForm.offset || 0),
      timestampMs: resetForm.resetType === 'timestamp' ? Number(resetForm.timestampMs) : 0,
    }
    if (!resetForm.allPartitions) {
      payload.partition = Number(resetForm.partition)
    }
    const res = await resetKafkaGroupOffset(resetForm.groupId, payload)
    const result = res?.data?.data
    ElMessage.success(`Offset 已重置，共更新 ${result?.partitions?.length || 0} 个分区`)
    resetDialogVisible.value = false
    await loadGroups()
    if (detailDrawerVisible.value && detailData.value.groupId === resetForm.groupId) {
      await loadGroupDetail(resetForm.groupId)
    }
  } catch (error) {
    ElMessage.error(error.message || 'Offset 重置失败')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    await loadClusters()
    await loadGroups()
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 集群加载失败')
  }
})
</script>

<style scoped>
.detail-summary {
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

.detail-card + .detail-card {
  margin-top: 16px;
}

.offset-form {
  margin-top: 16px;
}

.switch-label {
  margin-left: 10px;
  color: #606266;
}
</style>
