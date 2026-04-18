<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>Schema Registry</h2>
          <p>配置 Schema Registry，查看 Subject / Version / Schema，并做兼容性校验</p>
        </div>
        <div class="header-actions">
          <el-select v-model="selectedClusterId" style="width:260px" @change="refreshAll">
            <el-option v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
          <el-button type="primary" @click="openRegistryDialog()">新增 Registry</el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16">
      <el-col :span="9">
        <el-card class="content-card" v-loading="loading">
          <template #header><div class="flex-between"><div>Registry 配置</div><el-button @click="loadRegistries">刷新</el-button></div></template>
          <el-table :data="registries" empty-text="暂无 Registry 配置">
            <el-table-column prop="name" label="名称" min-width="140" />
            <el-table-column prop="endpoint" label="Endpoint" min-width="220" show-overflow-tooltip />
            <el-table-column label="操作" width="140" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="openRegistryDialog(row)">编辑</el-button>
                <el-button link type="danger" @click="deleteRegistry(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="7">
        <el-card class="content-card" v-loading="loading">
          <template #header><div class="flex-between"><div>Subjects</div><el-button @click="loadSubjects">刷新</el-button></div></template>
          <el-table :data="subjects" empty-text="暂无 Subject" @row-click="selectSubject">
            <el-table-column prop="subject" label="Subject" min-width="220" />
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card class="content-card" v-loading="loading">
          <template #header><div class="flex-between"><div>Schema 详情</div><el-button @click="openCompatibilityDialog" :disabled="!selectedSubject">兼容性校验</el-button></div></template>
          <div class="subject-meta">当前 Subject: {{ selectedSubject || '-' }}</div>
          <el-select v-model="selectedVersion" style="width:100%; margin-bottom:12px;" :disabled="!selectedSubject" @change="loadSchemaDetail">
            <el-option v-for="item in versions" :key="item.version" :label="String(item.version)" :value="String(item.version)" />
          </el-select>
          <pre class="schema-pre">{{ schemaDetail?.schema || '请选择 Subject 和 Version' }}</pre>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="registryDialogVisible" :title="editingRegistry ? '编辑 Registry' : '新增 Registry'" width="680px" destroy-on-close>
      <el-form label-position="top" :model="registryForm">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="名称"><el-input v-model="registryForm.name" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="Endpoint"><el-input v-model="registryForm.endpoint" placeholder="http://schema-registry:8081" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="认证方式"><el-select v-model="registryForm.authType" style="width:100%"><el-option label="无认证" value="none" /><el-option label="Basic" value="basic" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="用户名"><el-input v-model="registryForm.username" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="密码"><el-input v-model="registryForm.password" type="password" show-password /></el-form-item>
        <el-form-item label="描述"><el-input v-model="registryForm.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="registryDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveRegistry">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="compatibilityDialogVisible" title="Schema 兼容性校验" width="760px" destroy-on-close>
      <el-form label-position="top">
        <el-form-item label="Schema">
          <el-input v-model="compatibilitySchema" type="textarea" :rows="12" />
        </el-form-item>
      </el-form>
      <div class="compatibility-result" v-if="compatibilityResult">{{ compatibilityResult }}</div>
      <template #footer>
        <el-button @click="compatibilityDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="checkCompatibility">开始校验</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  checkKafkaSchemaCompatibility,
  createKafkaSchemaRegistry,
  deleteKafkaSchemaRegistry,
  getKafkaClusterOptions,
  getKafkaSchemaDetail,
  getKafkaSchemaRegistries,
  getKafkaSchemaSubjects,
  getKafkaSchemaVersions,
  updateKafkaSchemaRegistry,
} from '@/api/kafka.js'

const clusters = ref([])
const selectedClusterId = ref(null)
const loading = ref(false)
const saving = ref(false)
const registries = ref([])
const subjects = ref([])
const versions = ref([])
const selectedSubject = ref('')
const selectedVersion = ref('')
const schemaDetail = ref(null)
const registryDialogVisible = ref(false)
const editingRegistry = ref(null)
const compatibilityDialogVisible = ref(false)
const compatibilitySchema = ref('')
const compatibilityResult = ref('')
const registryForm = reactive({ name: '', endpoint: '', authType: 'none', username: '', password: '', description: '', verifySsl: true })

const loadClusters = async () => {
  const res = await getKafkaClusterOptions()
  clusters.value = res?.data?.data || []
  if (!selectedClusterId.value && clusters.value.length > 0) selectedClusterId.value = clusters.value[0].id
}

const loadRegistries = async () => {
  if (!selectedClusterId.value) return
  const res = await getKafkaSchemaRegistries({ clusterId: selectedClusterId.value })
  registries.value = res?.data?.data || []
}

const loadSubjects = async () => {
  if (!selectedClusterId.value) return
  try {
    const res = await getKafkaSchemaSubjects({ clusterId: selectedClusterId.value })
    subjects.value = res?.data?.data || []
  } catch (error) {
    ElMessage.error(error.message || 'Schema Subjects 加载失败')
  }
}

const loadVersions = async () => {
  if (!selectedClusterId.value || !selectedSubject.value) return
  const res = await getKafkaSchemaVersions({ clusterId: selectedClusterId.value, subject: selectedSubject.value })
  versions.value = res?.data?.data || []
  if (!selectedVersion.value && versions.value.length > 0) selectedVersion.value = String(versions.value[0].version)
}

const loadSchemaDetail = async () => {
  if (!selectedClusterId.value || !selectedSubject.value || !selectedVersion.value) return
  try {
    const res = await getKafkaSchemaDetail({ clusterId: selectedClusterId.value, subject: selectedSubject.value, version: selectedVersion.value })
    schemaDetail.value = res?.data?.data || null
  } catch (error) {
    ElMessage.error(error.message || 'Schema 详情加载失败')
  }
}

const refreshAll = async () => {
  loading.value = true
  try {
    await loadRegistries()
    await loadSubjects()
    versions.value = []
    selectedSubject.value = ''
    selectedVersion.value = ''
    schemaDetail.value = null
  } finally {
    loading.value = false
  }
}

const selectSubject = async (row) => {
  selectedSubject.value = row.subject
  selectedVersion.value = ''
  await loadVersions()
  await loadSchemaDetail()
}

const openRegistryDialog = (row = null) => {
  editingRegistry.value = row
  Object.assign(registryForm, row || { name: '', endpoint: '', authType: 'none', username: '', password: '', description: '', verifySsl: true })
  registryForm.password = ''
  registryDialogVisible.value = true
}

const saveRegistry = async () => {
  if (!selectedClusterId.value || !registryForm.name || !registryForm.endpoint) {
    ElMessage.warning('请填写名称和 Endpoint')
    return
  }
  saving.value = true
  try {
    const payload = { ...registryForm, clusterId: selectedClusterId.value }
    if (editingRegistry.value?.id) await updateKafkaSchemaRegistry(editingRegistry.value.id, payload)
    else await createKafkaSchemaRegistry(payload)
    ElMessage.success('Registry 已保存')
    registryDialogVisible.value = false
    await loadRegistries()
  } catch (error) {
    ElMessage.error(error.message || 'Registry 保存失败')
  } finally {
    saving.value = false
  }
}

const deleteRegistry = async (row) => {
  await ElMessageBox.confirm(`确认删除 Registry ${row.name} 吗？`, '提示', { type: 'warning' })
  try {
    await deleteKafkaSchemaRegistry(row.id)
    ElMessage.success('Registry 已删除')
    await loadRegistries()
  } catch (error) {
    ElMessage.error(error.message || 'Registry 删除失败')
  }
}

const openCompatibilityDialog = () => {
  compatibilitySchema.value = schemaDetail.value?.schema || ''
  compatibilityResult.value = ''
  compatibilityDialogVisible.value = true
}

const checkCompatibility = async () => {
  if (!selectedClusterId.value || !selectedSubject.value || !selectedVersion.value || !compatibilitySchema.value) {
    ElMessage.warning('请先选择 Subject/Version，并填写待校验的 Schema')
    return
  }
  saving.value = true
  try {
    const res = await checkKafkaSchemaCompatibility({ clusterId: selectedClusterId.value, subject: selectedSubject.value, version: selectedVersion.value, schema: compatibilitySchema.value })
    const data = res?.data?.data
    compatibilityResult.value = `${data?.isCompatible ? '兼容' : '不兼容'}: ${data?.message || ''}`
  } catch (error) {
    ElMessage.error(error.message || '兼容性校验失败')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    await loadClusters()
    await refreshAll()
  } catch (error) {
    ElMessage.error(error.message || 'Schema Registry 初始化失败')
  }
})
</script>

<style scoped>
.flex-between { display:flex; align-items:center; justify-content:space-between; }
.subject-meta { margin-bottom:12px; color:#606266; }
.schema-pre {
  min-height: 380px;
  padding: 12px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  background: #0f172a;
  border-radius: 10px;
  color: #e2e8f0;
}
.compatibility-result { margin-top:12px; color:#409eff; }
</style>
