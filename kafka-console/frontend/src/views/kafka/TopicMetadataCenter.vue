<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>Topic 元数据中心</h2>
          <p>维护 Topic 所属系统、负责人、生命周期、敏感级别和租户环境信息，形成 Kafka 数据资产台账</p>
        </div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="loadData">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-input v-model="keyword" style="width:220px" clearable placeholder="搜索 Topic / 系统 / Owner" @keyup.enter="loadData" />
          <el-button @click="loadData" :loading="loading">刷新</el-button>
          <el-button type="primary" @click="openDialog()">新增元数据</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <el-table :data="rows" empty-text="暂无 Topic 元数据">
        <el-table-column prop="topicName" label="Topic" min-width="180" />
        <el-table-column prop="systemName" label="系统" min-width="140" />
        <el-table-column prop="owner" label="Owner" width="120" />
        <el-table-column prop="ownerEmail" label="Owner Email" min-width="180" />
        <el-table-column prop="environment" label="环境" width="100" />
        <el-table-column prop="tenant" label="租户" width="120" />
        <el-table-column prop="lifecycle" label="生命周期" width="120" />
        <el-table-column prop="sensitivity" label="敏感级别" width="120" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="deleteRow(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑 Topic 元数据' : '新增 Topic 元数据'" width="760px" destroy-on-close>
      <el-form label-position="top" :model="form">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="Topic"><el-input v-model="form.topicName" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="系统"><el-input v-model="form.systemName" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="Owner"><el-input v-model="form.owner" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="Owner Email"><el-input v-model="form.ownerEmail" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="环境"><el-input v-model="form.environment" placeholder="dev/test/prod" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="租户"><el-input v-model="form.tenant" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="生命周期"><el-input v-model="form.lifecycle" placeholder="长期/临时/待下线" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="敏感级别"><el-input v-model="form.sensitivity" placeholder="public/internal/confidential" /></el-form-item>
        <el-form-item label="标签(JSON)"><el-input v-model="form.labels" type="textarea" :rows="4" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="4" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createKafkaTopicMetadata, deleteKafkaTopicMetadata, getKafkaClusterOptions, getKafkaTopicMetadata, updateKafkaTopicMetadata } from '@/api/kafka.js'

const clusters = ref([])
const selectedClusterId = ref(null)
const keyword = ref('')
const loading = ref(false)
const saving = ref(false)
const rows = ref([])
const dialogVisible = ref(false)
const editing = ref(null)
const form = reactive({ topicName: '', systemName: '', owner: '', ownerEmail: '', environment: '', tenant: '', lifecycle: '', sensitivity: '', labels: '{}', description: '' })

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id
}

const loadData = async () => {
  if (!selectedClusterId.value) return
  loading.value = true
  try {
    const res = await getKafkaTopicMetadata({ clusterId: selectedClusterId.value, keyword: keyword.value })
    rows.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || 'Topic 元数据加载失败')
  } finally {
    loading.value = false
  }
}

const openDialog = (row = null) => {
  editing.value = row
  Object.assign(form, row || { topicName: '', systemName: '', owner: '', ownerEmail: '', environment: '', tenant: '', lifecycle: '', sensitivity: '', labels: '{}', description: '' })
  dialogVisible.value = true
}

const save = async () => {
  if (!selectedClusterId.value || !form.topicName) {
    ElMessage.warning('请填写 Topic 名称')
    return
  }
  saving.value = true
  try {
    const payload = { ...form, clusterId: selectedClusterId.value }
    if (editing.value?.id) await updateKafkaTopicMetadata(editing.value.id, payload)
    else await createKafkaTopicMetadata(payload)
    ElMessage.success('Topic 元数据已保存')
    dialogVisible.value = false
    await loadData()
  } catch (error) {
    ElMessage.error(error.message || 'Topic 元数据保存失败')
  } finally {
    saving.value = false
  }
}

const deleteRow = async (row) => {
  await ElMessageBox.confirm(`确认删除 Topic 元数据 ${row.topicName} 吗？`, '提示', { type: 'warning' })
  try {
    await deleteKafkaTopicMetadata(row.id)
    ElMessage.success('Topic 元数据已删除')
    await loadData()
  } catch (error) {
    ElMessage.error(error.message || 'Topic 元数据删除失败')
  }
}

onMounted(async () => {
  try {
    await loadClusters()
    await loadData()
  } catch (error) {
    ElMessage.error(error.message || 'Topic 元数据中心初始化失败')
  }
})
</script>
