<template>
  <div class="page-container">
    <el-card class="page-header-card" shadow="never">
      <div class="page-header">
        <div class="page-header-copy">
          <div class="page-eyebrow">Kafka</div>
          <h2>消费组管理</h2>
          <p>集中看状态、Lag 和 Offset 干预入口，优先把异常消费组与热点 Topic 拉出来排查。</p>
        </div>

        <div class="page-header-side">
          <div class="page-header-meta">
            <div class="page-header-kpi">
              <span>当前集群</span>
              <strong>{{ currentClusterName }}</strong>
            </div>
            <div class="page-header-kpi">
              <span>优先处理</span>
              <strong>{{ prioritizedGroups.length }}</strong>
            </div>
          </div>
          <div class="page-header-note">
            {{ lagHotspotSummary.hotTopics }}
          </div>
          <div class="page-header-actions">
            <el-button @click="loadGroups" :loading="loading">刷新</el-button>
          </div>
        </div>
      </div>
    </el-card>

    <div class="page-metrics">
      <div class="page-metric-card">
        <span>消费组数量</span>
        <strong>{{ groupStats.total }}</strong>
      </div>
      <div class="page-metric-card is-accent">
        <span>在线成员</span>
        <strong>{{ groupStats.members }}</strong>
      </div>
      <div class="page-metric-card is-success">
        <span>稳定状态</span>
        <strong>{{ groupStats.stable }}</strong>
      </div>
      <div class="page-metric-card is-warning">
        <span>总 Lag</span>
        <strong>{{ groupStats.totalLag }}</strong>
      </div>
    </div>

    <el-card class="content-card">
      <template #header>
        <div class="card-header">
          <span>风险摘要</span>
          <span class="card-subtitle">消费状态</span>
        </div>
      </template>

      <div class="compact-list">
        <div v-for="item in prioritizedGroups" :key="item.groupId" class="compact-item">
          <div>
            <strong>{{ item.groupId }}</strong>
            <span>{{ item.priorityReason }}</span>
          </div>
          <el-tag :type="item.priorityLevel === 'high' ? 'danger' : 'warning'">
            {{ item.priorityLevel === 'high' ? '优先' : '关注' }}
          </el-tag>
        </div>
        <div v-if="prioritizedGroups.length === 0" class="compact-item">
          <div>
            <strong>当前状态</strong>
            <span>{{ lagHotspotSummary.highLagCount > 0 ? `高 Lag ${lagHotspotSummary.highLagCount} 个` : '整体稳定' }}</span>
          </div>
          <el-tag :type="lagHotspotSummary.highLagCount > 0 ? 'warning' : 'success'">{{ lagHotspotSummary.highLagCount > 0 ? '关注' : '正常' }}</el-tag>
        </div>
      </div>
    </el-card>

    <el-card class="content-card filter-card">
      <div class="toolbar-row">
        <div class="toolbar-left">
          <el-input
            v-model="keyword"
            placeholder="搜索消费组"
            style="width: 340px"
            clearable
            @keyup.enter="loadGroups"
          />
        </div>
        <div class="toolbar-right">
          <el-button type="primary" @click="loadGroups" :loading="loading">查询</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>消费组列表</span>
          <span class="card-subtitle">状态列表</span>
        </div>
      </template>

      <el-alert
        v-if="lagWarningGroups.length"
        class="table-alert"
        type="warning"
        :closable="false"
        show-icon
        :title="`有 ${lagWarningGroups.length} 个消费组的 Lag 结果不是完整快照`"
        :description="lagWarningSummary"
      />

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
        <el-table-column label="Lag" width="180">
          <template #default="{ row }">
            <div class="lag-cell">
              <span>{{ row.committedLag ?? 0 }}</span>
              <el-tooltip v-if="!row.lagAvailable || row.lagPartial" :content="row.lagWarningMessage || '当前 Lag 不是完整快照'" placement="top">
                <el-tag :type="row.lagAvailable ? 'warning' : 'danger'" size="small">
                  {{ row.lagAvailable ? '部分' : '不可用' }}
                </el-tag>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Topics" min-width="240">
          <template #default="{ row }">{{ (row.topics || []).join(', ') || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" min-width="280" fixed="right">
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
            <el-popconfirm
              v-if="permStore.hasPerm('kafka:group:delete') || permStore.roles.includes('admin')"
              :title="buildDeleteConfirmText(row)"
              confirm-button-text="确认删除"
              cancel-button-text="取消"
              @confirm="deleteGroup(row)"
            >
              <template #reference>
                <el-button link type="danger" :disabled="!canDeleteGroup(row)">删除消费组</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-drawer v-model="detailDrawerVisible" :title="`消费组详情: ${detailData.groupId || ''}`" size="75%">
      <el-skeleton :loading="detailLoading" animated :rows="8">
        <template #default>
          <div class="detail-toolbar">
            <div class="detail-toolbar-actions">
              <el-button
                v-if="permStore.hasPerm('kafka:group:offset:reset') || permStore.roles.includes('admin')"
                type="danger"
                plain
                :disabled="!detailData.topics || detailData.topics.length === 0"
                @click="openResetDialog(detailData)"
              >
                重置 Offset
              </el-button>
              <el-popconfirm
                v-if="permStore.hasPerm('kafka:group:delete') || permStore.roles.includes('admin')"
                :title="buildDeleteConfirmText(detailData)"
                confirm-button-text="确认删除"
                cancel-button-text="取消"
                @confirm="deleteGroup(detailData)"
              >
                <template #reference>
                  <el-button type="danger" :loading="saving" :disabled="!canDeleteGroup(detailData)">删除消费组</el-button>
                </template>
              </el-popconfirm>
            </div>
          </div>

          <div class="detail-summary-grid detail-summary">
            <div class="detail-summary-card">
              <span>成员数</span>
              <strong>{{ detailData.memberCount || 0 }}</strong>
            </div>
            <div class="detail-summary-card">
              <span>分区数</span>
              <strong>{{ detailData.partitionCount || 0 }}</strong>
            </div>
            <div class="detail-summary-card">
              <span>总 Lag</span>
              <strong>{{ detailData.totalLag || 0 }}</strong>
            </div>
            <div class="detail-summary-card">
              <span>状态</span>
              <strong>{{ detailData.state || '-' }}</strong>
            </div>
          </div>

          <el-card class="detail-card">
            <template #header>
              <div class="card-header">
                <span>消费者成员</span>
                <span class="card-subtitle">{{ detailData.members?.length || 0 }} 个成员</span>
              </div>
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

    <el-dialog v-model="resetDialogVisible" title="重置 Consumer Group Offset" width="min(640px, calc(100vw - 32px))" destroy-on-close>
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        title="Offset 重置会直接影响消费位置，请确保相关消费者已暂停或你明确知道后果。"
      />
      <el-alert
        v-if="resetRequiresForce"
        class="table-alert"
        type="error"
        :closable="false"
        show-icon
        :title="`当前消费组仍有 ${resetActiveMemberCount} 个活跃成员`"
        description="在线消费者可能立即触发 rebalance、重复消费或跳过消息。只有明确知道后果时才应强制重置。"
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
        <el-form-item v-if="resetRequiresForce">
          <el-checkbox v-model="resetForm.force">我确认当前消费者仍在线，继续强制重置</el-checkbox>
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
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { ElMessage } from 'element-plus'
import {
  deleteKafkaConsumerGroup,
  getKafkaConsumerGroupDetail,
  getKafkaConsumerGroups,
  resetKafkaGroupOffset,
} from '@/api/kafka.js'
import { useKafkaStore } from '@/stores/kafkaStore.js'
import { usePermissionStore } from '@/stores/permissionStore.js'
import { confirmKafkaRiskAction } from '@/utils/kafkaRiskConfirm.js'

const permStore = usePermissionStore()
const kafkaStore = useKafkaStore()

const loading = ref(false)
const detailLoading = ref(false)
const saving = ref(false)
const groups = ref([])
const keyword = ref('')
const detailDrawerVisible = ref(false)
const resetDialogVisible = ref(false)
const formRef = ref()
const activeGroup = ref(null)
const { clusterOptions: clusters, selectedClusterId } = storeToRefs(kafkaStore)
const handledErrors = new WeakSet()

const markErrorHandled = (error) => {
  if (error && typeof error === 'object') {
    handledErrors.add(error)
  }
  return error
}

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
  force: false,
  partition: 0,
  resetType: 'earliest',
  offset: 0,
  timestampMs: null,
})

const currentClusterName = computed(
  () => clusters.value.find((item) => item.id === selectedClusterId.value)?.name || '-',
)

const groupStats = computed(() => ({
  total: groups.value.length,
  members: groups.value.reduce((sum, item) => sum + Number(item.memberCount || 0), 0),
  stable: groups.value.filter((item) => item.state === 'Stable').length,
  totalLag: groups.value.reduce((sum, item) => sum + Number(item.committedLag || 0), 0),
}))

const lagWarningGroups = computed(() =>
  groups.value.filter((item) => item.lagAvailable === false || item.lagPartial),
)

const lagWarningSummary = computed(() => {
  const sample = lagWarningGroups.value
    .slice(0, 3)
    .map((item) => `${item.groupId}：${item.lagWarningMessage || 'Lag 结果不完整'}`)
    .join('；')
  return sample || '请优先查看消费组详情，确认 offset 与 broker 状态。'
})

const resetActiveMemberCount = computed(() => Number(activeGroup.value?.memberCount ?? detailData.value?.memberCount ?? 0))
const resetRequiresForce = computed(() => resetActiveMemberCount.value > 0)

const prioritizedGroups = computed(() =>
  groups.value
    .map((item) => {
      const lag = Number(item.committedLag || 0)
      const partitionCount = Number(item.partitionCount || 0)
      const reasons = []
      let score = 0

      if (item.state !== 'Stable') {
        reasons.push(`状态异常（${item.state || 'Unknown'}）`)
        score += 3
      }
      if (lag > 0) {
        reasons.push(`存在 Lag（${lag}）`)
        score += lag >= 1000 ? 3 : 2
      }
      if (item.lagAvailable === false) {
        reasons.push('Lag 暂不可用')
        score += 2
      } else if (item.lagPartial) {
        reasons.push('Lag 为部分结果')
        score += 1
      }
      if (partitionCount >= 20) {
        reasons.push(`分区数较高（${partitionCount}）`)
        score += 1
      }

      return {
        ...item,
        priorityScore: score,
        priorityLevel: score >= 5 ? 'high' : 'medium',
        priorityReason: reasons.join('；') || '当前没有明显异常信号',
      }
    })
    .filter((item) => item.priorityScore > 0)
    .sort((a, b) => b.priorityScore - a.priorityScore || Number(b.committedLag || 0) - Number(a.committedLag || 0))
    .slice(0, 5),
)

const lagHotspotSummary = computed(() => {
  const highLagGroups = groups.value.filter((item) => Number(item.committedLag || 0) > 0)
  const topicCounter = new Map()

  highLagGroups.forEach((group) => {
    ;(group.topics || []).forEach((topic) => {
      if (!topic) return
      topicCounter.set(topic, (topicCounter.get(topic) || 0) + 1)
    })
  })

  const hotTopics = Array.from(topicCounter.entries())
    .sort((a, b) => b[1] - a[1])
    .slice(0, 3)
    .map(([topic, count]) => `${topic}（${count} 个消费组）`)
    .join('，')

  return {
    highLagCount: highLagGroups.length,
    hotTopics: hotTopics || '当前没有明显的 Lag 热点 Topic',
  }
})

const topicOptions = computed(() => activeGroup.value?.topics || detailData.value?.topics || [])

const isGroupEmpty = (group) => String(group?.state || '').trim() === 'Empty'

const canDeleteGroup = (group) =>
  Boolean(selectedClusterId.value) &&
  Boolean(String(group?.groupId || '').trim()) &&
  isGroupEmpty(group) &&
  Number(group?.memberCount || 0) === 0

const buildDeleteConfirmText = (group) => {
  const groupId = String(group?.groupId || '').trim()
  return groupId ? `确认删除消费组 ${groupId} 吗？` : '确认删除当前消费组吗？'
}

const resetRules = {
  topic: [{ required: true, message: '请选择 Topic', trigger: 'change' }],
  partition: [{
    trigger: 'change',
    validator: (_rule, value, callback) => {
      if (resetForm.allPartitions) {
        callback()
        return
      }
      const numericValue = Number(value)
      if (value === null || value === undefined || value === '' || Number.isNaN(numericValue) || numericValue < 0 || !Number.isInteger(numericValue)) {
        callback(new Error('请输入 Partition'))
        return
      }
      callback()
    },
  }],
  resetType: [{ required: true, message: '请选择重置方式', trigger: 'change' }],
}

const loadClusters = async () => {
  try {
    await kafkaStore.loadClusterOptions()
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 集群列表加载失败')
    throw markErrorHandled(error)
  }
}

const loadGroups = async () => {
  if (loading.value || !selectedClusterId.value) return
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
  const targetGroup = group || activeGroup.value || detailData.value || null
  const fallbackTopic = targetGroup?.topics?.[0] || detailData.value?.topics?.[0] || ''
  const partitionNumber = Number(partitionRow?.partition)

  activeGroup.value = targetGroup
  resetForm.groupId = targetGroup?.groupId || detailData.value?.groupId || ''
  resetForm.topic = partitionRow?.topic || fallbackTopic
  resetForm.allPartitions = false
  resetForm.force = false
  resetForm.partition = Number.isInteger(partitionNumber) && partitionNumber >= 0 ? partitionNumber : 0
  resetForm.resetType = 'earliest'
  resetForm.offset = 0
  resetForm.timestampMs = null
  resetDialogVisible.value = true
}

const deleteGroup = async (group) => {
  const targetGroup = group || activeGroup.value || detailData.value || null
  const groupId = String(targetGroup?.groupId || '').trim()
  if (!selectedClusterId.value || !groupId) return
  if (!isGroupEmpty(targetGroup)) {
    ElMessage.warning(`当前消费组状态为 ${targetGroup?.state || 'Unknown'}，仅 Empty 状态允许删除`)
    return
  }
  const targetMemberCount = Number(targetGroup?.memberCount ?? detailData.value?.memberCount ?? 0)
  if (targetMemberCount > 0) {
    ElMessage.warning('当前消费组仍有在线消费者，请先停止消费者后再删除')
    return
  }

  saving.value = true
  try {
    await deleteKafkaConsumerGroup(groupId, selectedClusterId.value)
    ElMessage.success('消费组删除成功')
    resetDialogVisible.value = false
    if (detailDrawerVisible.value && detailData.value.groupId === groupId) {
      detailDrawerVisible.value = false
    }
    await loadGroups()
  } catch (error) {
    ElMessage.error(error.message || '消费组删除失败')
  } finally {
    saving.value = false
  }
}

const handleResetOffset = async () => {
  if (!formRef.value || !selectedClusterId.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  if (!resetForm.allPartitions && (resetForm.partition === null || resetForm.partition === undefined)) {
    ElMessage.warning('请输入 Partition')
    return
  }
  if (!resetForm.allPartitions && (!Number.isInteger(Number(resetForm.partition)) || Number(resetForm.partition) < 0)) {
    ElMessage.warning('请输入有效的 Partition')
    return
  }
  if (resetForm.resetType === 'offset' && (!Number.isInteger(Number(resetForm.offset)) || Number(resetForm.offset) < 0)) {
    ElMessage.warning('请输入有效的 Offset')
    return
  }
  if (resetForm.resetType === 'timestamp' && !resetForm.timestampMs) {
    ElMessage.warning('请选择时间')
    return
  }
  if (resetRequiresForce.value && !resetForm.force) {
    ElMessage.warning('当前消费组仍在线，请勾选强制重置后再继续')
    return
  }
  const confirmed = await confirmKafkaRiskAction({
    title: 'Offset 重置确认',
    resourceName: `${resetForm.groupId} / ${resetForm.topic}`,
    actionLabel: '重置消费组 Offset',
    dangerPoints: [
      resetForm.allPartitions
        ? '会修改该 Topic 下全部分区的消费位置'
        : `会修改分区 ${resetForm.partition} 的消费位置`,
      `重置方式为 ${resetForm.resetType}`,
      '如果消费者仍在运行，可能立刻触发重复消费或跳过消息',
    ],
    confirmButtonText: '确认重置',
  })
  if (!confirmed) return
  saving.value = true
  try {
    const payload = {
      clusterId: selectedClusterId.value,
      topic: resetForm.topic,
      allPartitions: resetForm.allPartitions,
      force: resetForm.force,
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
    if (!(error && typeof error === 'object' && handledErrors.has(error))) {
      ElMessage.error(error.message || 'Kafka 集群加载失败')
    }
  }
})

watch(selectedClusterId, async (nextValue, previousValue) => {
  if (!nextValue || nextValue === previousValue) return
  detailDrawerVisible.value = false
  detailData.groupId = ''
  detailData.members = []
  detailData.partitions = []
  await loadGroups()
})
</script>

<style scoped>
.detail-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: flex-end;
  gap: 16px;
  margin-bottom: 16px;
  padding: 14px 16px;
  border: 1px solid rgba(148, 163, 184, 0.14);
  border-radius: 18px;
  background: rgba(248, 250, 252, 0.82);
}

.detail-toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-start;
  gap: 10px;
}

.offset-form {
  margin-top: 16px;
}

.table-alert {
  margin-bottom: 16px;
}

.lag-cell {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

@media (max-width: 960px) {
  .detail-toolbar {
    flex-direction: column;
  }

  .detail-toolbar-actions {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>
