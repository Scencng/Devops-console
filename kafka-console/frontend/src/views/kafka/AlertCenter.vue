<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>告警中心</h2>
          <p>管理 Kafka 告警规则，集中查看告警事件并支持手动评估、确认和恢复</p>
        </div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="refreshAll">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-button :loading="loading" @click="evaluateRules">立即评估</el-button>
        </div>
      </div>
    </el-card>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="规则管理" name="rules">
        <el-card class="content-card" v-loading="loading">
          <template #header>
            <div class="flex-between">
              <div>规则列表</div>
              <el-button type="primary" @click="openRuleDialog()">新增规则</el-button>
            </div>
          </template>
          <el-table :data="rules" empty-text="暂无告警规则">
            <el-table-column prop="name" label="规则名" min-width="180" />
            <el-table-column prop="metricType" label="指标" min-width="160" />
            <el-table-column prop="severity" label="级别" width="120" />
            <el-table-column label="阈值" width="120">
              <template #default="{ row }">{{ row.operator }} {{ row.threshold }}</template>
            </el-table-column>
            <el-table-column prop="enabled" label="启用" width="100">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="runbook" label="Runbook" min-width="220" show-overflow-tooltip />
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openRuleDialog(row)">编辑</el-button>
                <el-button link type="danger" @click="handleDeleteRule(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="告警事件" name="events">
        <el-card class="content-card" v-loading="loading">
          <template #header><div class="flex-between"><div>事件列表</div><el-button @click="loadEvents">刷新</el-button></div></template>
          <el-table :data="events" empty-text="暂无告警事件">
            <el-table-column prop="title" label="标题" min-width="200" />
            <el-table-column prop="severity" label="级别" width="120" />
            <el-table-column prop="status" label="状态" width="120" />
            <el-table-column prop="metricType" label="指标" min-width="160" />
            <el-table-column label="当前值/阈值" width="140">
              <template #default="{ row }">{{ row.metricValue }} / {{ row.threshold }}</template>
            </el-table-column>
            <el-table-column prop="message" label="说明" min-width="260" show-overflow-tooltip />
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="{ row }">
                <el-button link type="warning" :disabled="row.status !== 'open'" @click="updateEventStatus(row, 'acked')">确认</el-button>
                <el-button link type="success" :disabled="row.status === 'resolved'" @click="updateEventStatus(row, 'resolved')">恢复</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="dialogVisible" :title="editingRule ? '编辑告警规则' : '新增告警规则'" width="620px" destroy-on-close>
      <el-form label-position="top" :model="ruleForm">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="规则名"><el-input v-model="ruleForm.name" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="指标"><el-select v-model="ruleForm.metricType" style="width:100%"><el-option label="总 Lag" value="consumer_lag" /><el-option label="未同步分区" value="under_replicated_partitions" /><el-option label="异常 Broker" value="broker_down" /></el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="级别"><el-select v-model="ruleForm.severity" style="width:100%"><el-option label="critical" value="critical" /><el-option label="warning" value="warning" /><el-option label="info" value="info" /></el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="运算符"><el-select v-model="ruleForm.operator" style="width:100%"><el-option label=">" value=">" /><el-option label=">=" value=">=" /><el-option label="<" value="<" /><el-option label="<=" value="<=" /></el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="阈值"><el-input-number v-model="ruleForm.threshold" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="Runbook"><el-input v-model="ruleForm.runbook" type="textarea" :rows="3" /></el-form-item>
        <el-form-item><el-switch v-model="ruleForm.enabled" /><span class="switch-label">启用规则</span></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createKafkaAlertRule,
  deleteKafkaAlertRule,
  evaluateKafkaAlertRules,
  getKafkaAlertEvents,
  getKafkaAlertRules,
  getKafkaClusterOptions,
  updateKafkaAlertEventStatus,
  updateKafkaAlertRule,
} from '@/api/kafka.js'

const activeTab = ref('rules')
const loading = ref(false)
const saving = ref(false)
const clusters = ref([])
const selectedClusterId = ref(null)
const rules = ref([])
const events = ref([])
const dialogVisible = ref(false)
const editingRule = ref(null)
const ruleForm = reactive({ name: '', metricType: 'consumer_lag', severity: 'warning', operator: '>', threshold: 0, enabled: true, runbook: '' })

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id
}

const loadRules = async () => {
  if (!selectedClusterId.value) return
  const res = await getKafkaAlertRules({ clusterId: selectedClusterId.value })
  rules.value = res?.data?.data || []
}

const loadEvents = async () => {
  if (!selectedClusterId.value) return
  const res = await getKafkaAlertEvents({ clusterId: selectedClusterId.value })
  events.value = res?.data?.data || []
}

const refreshAll = async () => {
  loading.value = true
  try {
    await Promise.all([loadRules(), loadEvents()])
  } catch (error) {
    ElMessage.error(error.message || '告警中心加载失败')
  } finally {
    loading.value = false
  }
}

const evaluateRules = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    await evaluateKafkaAlertRules({ clusterId: selectedClusterId.value })
    ElMessage.success('规则评估完成')
    await refreshAll()
  } catch (error) {
    ElMessage.error(error.message || '规则评估失败')
  } finally {
    loading.value = false
  }
}

const openRuleDialog = (row = null) => {
  editingRule.value = row
  Object.assign(ruleForm, row || { name: '', metricType: 'consumer_lag', severity: 'warning', operator: '>', threshold: 0, enabled: true, runbook: '' })
  dialogVisible.value = true
}

const saveRule = async () => {
  if (!selectedClusterId.value || !ruleForm.name) {
    ElMessage.warning('请填写规则名')
    return
  }
  saving.value = true
  try {
    const payload = { ...ruleForm, clusterId: selectedClusterId.value }
    if (editingRule.value?.id) await updateKafkaAlertRule(editingRule.value.id, payload)
    else await createKafkaAlertRule(payload)
    ElMessage.success('规则已保存')
    dialogVisible.value = false
    await loadRules()
  } catch (error) {
    ElMessage.error(error.message || '规则保存失败')
  } finally {
    saving.value = false
  }
}

const handleDeleteRule = async (row) => {
  await ElMessageBox.confirm(`确认删除告警规则 ${row.name} 吗？`, '提示', { type: 'warning' })
  try {
    await deleteKafkaAlertRule(row.id)
    ElMessage.success('规则已删除')
    await loadRules()
  } catch (error) {
    ElMessage.error(error.message || '规则删除失败')
  }
}

const updateEventStatus = async (row, status) => {
  try {
    await updateKafkaAlertEventStatus(row.id, { status })
    ElMessage.success('事件状态已更新')
    await loadEvents()
  } catch (error) {
    ElMessage.error(error.message || '事件状态更新失败')
  }
}

onMounted(async () => {
  try {
    await loadClusters()
    await refreshAll()
  } catch (error) {
    ElMessage.error(error.message || '告警中心初始化失败')
  }
})
</script>

<style scoped>
.flex-between { display:flex; align-items:center; justify-content:space-between; }
.switch-label { margin-left:10px; color:#606266; }
</style>
