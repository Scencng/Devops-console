<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div><h2>自愈中心</h2><p>配置 Kafka 自愈策略，并在异常时触发巡检、告警评估或标准动作</p></div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="refreshAll"><el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" /></el-select>
          <el-button type="primary" @click="openDialog()">新建策略</el-button>
        </div>
      </div>
    </el-card>
    <el-row :gutter="16">
      <el-col :span="12">
        <el-card class="content-card" v-loading="loading">
          <template #header><div class="flex-between"><div>策略列表</div><el-button @click="loadPolicies">刷新</el-button></div></template>
          <el-table :data="policies" empty-text="暂无自愈策略">
            <el-table-column prop="name" label="名称" min-width="180" />
            <el-table-column prop="triggerType" label="触发条件" min-width="140" />
            <el-table-column prop="actionType" label="动作" min-width="160" />
            <el-table-column prop="enabled" label="启用" width="100"><template #default="{ row }"><el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag></template></el-table-column>
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="executePolicy(row)">执行</el-button>
                <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
                <el-button link type="danger" @click="removePolicy(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card class="content-card" v-loading="loading">
          <template #header><div class="flex-between"><div>执行记录</div><el-button @click="loadExecutions">刷新</el-button></div></template>
          <el-table :data="executions" empty-text="暂无执行记录">
            <el-table-column prop="status" label="状态" width="120" />
            <el-table-column prop="summary" label="摘要" min-width="240" show-overflow-tooltip />
            <el-table-column prop="startedAt" label="开始时间" min-width="180" />
            <el-table-column prop="completedAt" label="结束时间" min-width="180" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑策略' : '新建策略'" width="680px" destroy-on-close>
      <el-form label-position="top" :model="form">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="名称"><el-input v-model="form.name" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="触发条件"><el-input v-model="form.triggerType" placeholder="例如 broker_down / high_lag" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="动作"><el-select v-model="form.actionType" style="width:100%"><el-option label="巡检" value="inspection.run" /><el-option label="告警评估" value="alerts.evaluate" /><el-option label="创建 Topic" value="topic.create" /><el-option label="扩容分区" value="topic.partitions.increase" /><el-option label="重置 Offset" value="group.offset.reset" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="启用"><el-switch v-model="form.enabled" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="动作配置(JSON)"><el-input v-model="form.config" type="textarea" :rows="8" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createKafkaSelfHealingPolicy, deleteKafkaSelfHealingPolicy, executeKafkaSelfHealingPolicy, getKafkaClusterOptions, getKafkaSelfHealingExecutions, getKafkaSelfHealingPolicies, updateKafkaSelfHealingPolicy } from '@/api/kafka.js'

const clusters = ref([])
const selectedClusterId = ref(null)
const loading = ref(false)
const saving = ref(false)
const policies = ref([])
const executions = ref([])
const dialogVisible = ref(false)
const editing = ref(null)
const form = reactive({ name: '', triggerType: '', actionType: 'inspection.run', config: '{}', enabled: true })

const loadClusters = async () => { const res = await getKafkaClusterOptions(); clusters.value = res?.data?.data || []; if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id }
const loadPolicies = async () => { if (!selectedClusterId.value) return; const res = await getKafkaSelfHealingPolicies({ clusterId: selectedClusterId.value }); policies.value = res?.data?.data || [] }
const loadExecutions = async () => { if (!selectedClusterId.value) return; const res = await getKafkaSelfHealingExecutions({ clusterId: selectedClusterId.value }); executions.value = res?.data?.data || [] }
const refreshAll = async () => { loading.value = true; try { await Promise.all([loadPolicies(), loadExecutions()]) } catch (error) { ElMessage.error(error.message || '自愈中心加载失败') } finally { loading.value = false } }
const openDialog = (row = null) => { editing.value = row; Object.assign(form, row || { name: '', triggerType: '', actionType: 'inspection.run', config: '{}', enabled: true }); dialogVisible.value = true }
const save = async () => {
  if (!selectedClusterId.value || !form.name || !form.triggerType) { ElMessage.warning('请填写名称和触发条件'); return }
  saving.value = true
  try {
    const payload = { ...form, clusterId: selectedClusterId.value }
    if (editing.value?.id) await updateKafkaSelfHealingPolicy(editing.value.id, payload)
    else await createKafkaSelfHealingPolicy(payload)
    ElMessage.success('策略已保存')
    dialogVisible.value = false
    await loadPolicies()
  } catch (error) { ElMessage.error(error.message || '策略保存失败') } finally { saving.value = false }
}
const executePolicy = async (row) => {
  try { await executeKafkaSelfHealingPolicy(row.id); ElMessage.success('策略执行完成'); await loadExecutions() } catch (error) { ElMessage.error(error.message || '策略执行失败') }
}
const removePolicy = async (row) => {
  await ElMessageBox.confirm(`确认删除策略 ${row.name} 吗？`, '提示', { type: 'warning' })
  try { await deleteKafkaSelfHealingPolicy(row.id); ElMessage.success('策略已删除'); await loadPolicies() } catch (error) { ElMessage.error(error.message || '策略删除失败') }
}
onMounted(async () => { await loadClusters(); await refreshAll() })
</script>

<style scoped>.flex-between{display:flex;align-items:center;justify-content:space-between;}</style>
