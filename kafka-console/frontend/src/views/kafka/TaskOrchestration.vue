<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>任务编排</h2>
          <p>管理 Kafka 自动化任务，支持巡检、规则评估和高频运维动作的标准化执行</p>
        </div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="refreshAll">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-button type="primary" @click="openDialog()">新建任务</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <el-table :data="tasks" empty-text="暂无任务">
        <el-table-column prop="name" label="任务名" min-width="180" />
        <el-table-column prop="taskType" label="类型" min-width="180" />
        <el-table-column prop="cronExpr" label="Cron" min-width="140" />
        <el-table-column prop="lastRunStatus" label="最近执行" width="120" />
        <el-table-column prop="enabled" label="启用" width="100">
          <template #default="{ row }"><el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="runTask(row)">运行</el-button>
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="primary" @click="viewRuns(row)">记录</el-button>
            <el-button link type="danger" @click="deleteTask(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingTask ? '编辑任务' : '新建任务'" width="680px" destroy-on-close>
      <el-form label-position="top" :model="taskForm">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="任务名"><el-input v-model="taskForm.name" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="类型"><el-select v-model="taskForm.taskType" style="width:100%"><el-option label="巡检" value="inspection.run" /><el-option label="告警评估" value="alerts.evaluate" /><el-option label="创建 Topic" value="topic.create" /><el-option label="扩容分区" value="topic.partitions.increase" /><el-option label="重置 Offset" value="group.offset.reset" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="Cron"><el-input v-model="taskForm.cronExpr" placeholder="可为空，例如 0 */1 * * *" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="启用"><el-switch v-model="taskForm.enabled" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="Payload(JSON)">
          <el-input v-model="taskForm.payload" type="textarea" :rows="10" placeholder='例如 {"topic":"orders","count":12}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveTask">保存</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="runDrawerVisible" :title="`执行记录: ${activeTask?.name || ''}`" size="55%">
      <el-table :data="taskRuns" v-loading="runsLoading" empty-text="暂无执行记录">
        <el-table-column prop="status" label="状态" width="120" />
        <el-table-column prop="triggerMode" label="触发方式" width="120" />
        <el-table-column prop="resultSummary" label="结果摘要" min-width="240" show-overflow-tooltip />
        <el-table-column prop="startedAt" label="开始时间" min-width="180" />
        <el-table-column prop="finishedAt" label="结束时间" min-width="180" />
      </el-table>
    </el-drawer>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createKafkaTask, deleteKafkaTask, getKafkaClusterOptions, getKafkaTaskRuns, getKafkaTasks, runKafkaTask, updateKafkaTask } from '@/api/kafka.js'

const clusters = ref([])
const selectedClusterId = ref(null)
const loading = ref(false)
const saving = ref(false)
const tasks = ref([])
const dialogVisible = ref(false)
const editingTask = ref(null)
const runDrawerVisible = ref(false)
const activeTask = ref(null)
const taskRuns = ref([])
const runsLoading = ref(false)
const taskForm = reactive({ name: '', taskType: 'inspection.run', cronExpr: '', enabled: true, payload: '{}' })

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id
}

const loadTasks = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const res = await getKafkaTasks({ clusterId: selectedClusterId.value })
    tasks.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || '任务列表加载失败')
  } finally {
    loading.value = false
  }
}

const refreshAll = async () => { await loadTasks() }

const openDialog = (row = null) => {
  editingTask.value = row
  Object.assign(taskForm, row || { name: '', taskType: 'inspection.run', cronExpr: '', enabled: true, payload: '{}' })
  dialogVisible.value = true
}

const saveTask = async () => {
  if (!selectedClusterId.value || !taskForm.name) {
    ElMessage.warning('请填写任务名')
    return
  }
  saving.value = true
  try {
    const payload = { ...taskForm, clusterId: selectedClusterId.value }
    if (editingTask.value?.id) await updateKafkaTask(editingTask.value.id, payload)
    else await createKafkaTask(payload)
    ElMessage.success('任务已保存')
    dialogVisible.value = false
    await loadTasks()
  } catch (error) {
    ElMessage.error(error.message || '任务保存失败')
  } finally {
    saving.value = false
  }
}

const runTask = async (row) => {
  try {
    await runKafkaTask(row.id, { triggerMode: 'manual' })
    ElMessage.success('任务执行完成')
    await loadTasks()
  } catch (error) {
    ElMessage.error(error.message || '任务执行失败')
  }
}

const deleteTask = async (row) => {
  await ElMessageBox.confirm(`确认删除任务 ${row.name} 吗？`, '提示', { type: 'warning' })
  try {
    await deleteKafkaTask(row.id)
    ElMessage.success('任务已删除')
    await loadTasks()
  } catch (error) {
    ElMessage.error(error.message || '任务删除失败')
  }
}

const viewRuns = async (row) => {
  activeTask.value = row
  runDrawerVisible.value = true
  runsLoading.value = true
  try {
    const res = await getKafkaTaskRuns({ taskId: row.id })
    taskRuns.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || '执行记录加载失败')
  } finally {
    runsLoading.value = false
  }
}

onMounted(async () => {
  try {
    await loadClusters()
    await loadTasks()
  } catch (error) {
    ElMessage.error(error.message || '任务编排初始化失败')
  }
})
</script>
