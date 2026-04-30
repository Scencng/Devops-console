<template>
  <div class="page-container">
    <el-card class="page-header-card" shadow="never">
      <div class="page-header">
        <div class="page-header-copy">
          <div class="page-eyebrow">Kafka</div>
          <h2>集群管理</h2>
          <p>维护连接、认证和环境归属，先保证可连通，再处理 Topic 和消费组问题。</p>
        </div>

        <div class="page-header-side">
          <div class="page-header-meta">
            <div class="page-header-kpi">
              <span>集群总数</span>
              <strong>{{ clusterStats.total }}</strong>
            </div>
            <div class="page-header-kpi">
              <span>认证 / TLS 关注项</span>
              <strong>{{ authRiskClusterCount }}</strong>
            </div>
          </div>
          <div class="page-header-actions">
            <el-button @click="loadClusters" :loading="loading">刷新</el-button>
            <el-button
              v-if="permStore.hasPerm('kafka:cluster:create') || permStore.roles.includes('admin')"
              type="primary"
              @click="openCreateDialog"
            >
              新增集群
            </el-button>
          </div>
        </div>
      </div>
    </el-card>

      <div class="page-metrics">
        <div class="page-metric-card is-success">
        <span>当前页正常</span>
          <strong>{{ clusterStats.active }}</strong>
        </div>
        <div class="page-metric-card is-warning">
        <span>当前页异常</span>
          <strong>{{ clusterStats.error }}</strong>
        </div>
        <div class="page-metric-card">
        <span>当前页测试失败</span>
          <strong>{{ clusterStats.failedRecently }}</strong>
        </div>
      </div>

    <el-card class="content-card">
      <template #header>
        <div class="card-header">
          <span>风险摘要</span>
          <span class="card-subtitle">连接与认证（当前页）</span>
        </div>
      </template>

      <div class="compact-list">
        <div v-if="failingCluster" class="compact-item">
          <div>
            <strong>最近失败</strong>
            <span>{{ failingCluster.name }} / {{ failingCluster.lastErrorMessage || '连接测试失败' }}</span>
          </div>
          <el-tag type="danger">异常</el-tag>
        </div>
        <div class="compact-item">
          <div>
            <strong>无认证集群</strong>
            <span>{{ authRiskSummary.noAuthCount }} 个</span>
          </div>
          <el-tag :type="authRiskSummary.noAuthCount > 0 ? 'warning' : 'success'">{{ authRiskSummary.noAuthCount > 0 ? '关注' : '正常' }}</el-tag>
        </div>
        <div class="compact-item">
          <div>
            <strong>TLS 未启用</strong>
            <span>{{ authRiskSummary.tlsDisabledCount }} 个</span>
          </div>
          <el-tag :type="authRiskSummary.tlsDisabledCount > 0 ? 'warning' : 'success'">{{ authRiskSummary.tlsDisabledCount > 0 ? '关注' : '正常' }}</el-tag>
        </div>
        <div class="compact-item">
          <div>
            <strong>跳过证书校验</strong>
            <span>{{ authRiskSummary.insecureSkipVerifyCount }} 个</span>
          </div>
          <el-tag :type="authRiskSummary.insecureSkipVerifyCount > 0 ? 'warning' : 'success'">{{ authRiskSummary.insecureSkipVerifyCount > 0 ? '关注' : '正常' }}</el-tag>
        </div>
      </div>
    </el-card>

    <el-card class="content-card filter-card">
      <div class="toolbar-row">
        <div class="toolbar-left">
          <el-input
            v-model="keyword"
            placeholder="搜索集群名称或地址"
            style="width: 280px"
            clearable
            @keyup.enter="handleQuery"
          />
          <el-input
            v-model="environment"
            placeholder="环境"
            style="width: 140px"
            clearable
            @keyup.enter="handleQuery"
          />
          <el-input
            v-model="tenant"
            placeholder="租户"
            style="width: 140px"
            clearable
            @keyup.enter="handleQuery"
          />
          <el-select v-model="status" placeholder="状态" clearable style="width: 150px" @change="handleQuery">
            <el-option label="全部" value="" />
            <el-option label="正常" value="active" />
            <el-option label="异常" value="error" />
            <el-option label="未知" value="unknown" />
          </el-select>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" @click="handleQuery" :loading="loading">查询</el-button>
        </div>
      </div>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>集群列表</span>
          <span class="card-subtitle">连接列表</span>
        </div>
      </template>

      <el-table :data="clusters" empty-text="暂无 Kafka 集群">
        <el-table-column prop="name" label="名称" min-width="180" />
        <el-table-column prop="bootstrapServers" label="Bootstrap Servers" min-width="260" show-overflow-tooltip />
        <el-table-column prop="version" label="版本" width="120" />
        <el-table-column prop="environment" label="环境" width="100" />
        <el-table-column prop="tenant" label="租户" width="120" />
        <el-table-column prop="authType" label="认证" width="140" />
        <el-table-column prop="tlsEnabled" label="TLS" width="100">
          <template #default="{ row }">
            <el-tag :type="row.tlsEnabled ? 'success' : 'info'">{{ row.tlsEnabled ? '开启' : '关闭' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastTestedAt" label="最近测试" width="180">
          <template #default="{ row }">{{ formatTime(row.lastTestedAt) }}</template>
        </el-table-column>
        <el-table-column prop="lastErrorMessage" label="最近失败原因" min-width="260" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.status === 'error'">{{ row.lastErrorMessage || '最近一次测试失败，但没有返回详细错误信息' }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="permStore.hasPerm('kafka:cluster:test') || permStore.roles.includes('admin')"
              link
              type="primary"
              :loading="testingId === row.id"
              @click="handleTest(row)"
            >
              测试
            </el-button>
            <el-button
              v-if="permStore.hasPerm('kafka:cluster:edit') || permStore.roles.includes('admin')"
              link
              type="primary"
              :loading="editingId === row.id"
              @click="openEditDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="permStore.hasPerm('kafka:cluster:delete') || permStore.roles.includes('admin')"
              link
              type="danger"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <div v-show="paginationTotal > paginationPageSize" class="table-pagination">
        <el-pagination
          background
          layout="prev, pager, next, total"
          :total="paginationTotal"
          :page-size="paginationPageSize"
          :current-page="paginationPage"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑 Kafka 集群' : '新增 Kafka 集群'" width="760px" destroy-on-close>
      <el-form ref="formRef" v-loading="dialogLoading" :model="formData" :rules="rules" label-position="top">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="集群名称" prop="name">
              <el-input v-model="formData.name" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Kafka 版本" prop="version">
              <el-input v-model="formData.version" placeholder="例如 3.6.0" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="环境">
              <el-input v-model="formData.environment" placeholder="dev/test/prod" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="租户">
              <el-input v-model="formData.tenant" placeholder="例如 core-team" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="Bootstrap Servers" prop="bootstrapServers">
          <el-input v-model="formData.bootstrapServers" placeholder="10.0.0.1:9092,10.0.0.2:9092" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="formData.description" type="textarea" :rows="2" />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="认证方式" prop="authType">
              <el-select v-model="formData.authType">
                <el-option label="无认证" value="none" />
                <el-option label="SASL/PLAIN" value="plain" />
                <el-option label="SCRAM-SHA-256" value="scram_sha256" />
                <el-option label="SCRAM-SHA-512" value="scram_sha512" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="用户名">
              <el-input v-model="formData.username" :disabled="formData.authType === 'none'" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="密码">
          <el-input
            v-model="formData.password"
            type="password"
            show-password
            placeholder="编辑时留空表示保留原密码"
            :disabled="formData.authType === 'none'"
          />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item><el-checkbox v-model="formData.tlsEnabled">启用 TLS</el-checkbox></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item><el-checkbox v-model="formData.insecureSkipVerify" :disabled="!formData.tlsEnabled">跳过证书校验</el-checkbox></el-form-item>
          </el-col>
        </el-row>

        <template v-if="formData.tlsEnabled">
          <el-form-item label="CA 证书">
            <el-input v-model="formData.caCert" type="textarea" :rows="4" placeholder="PEM 内容，可选" />
            <div class="form-help">{{ formData.hasCACert ? '当前已配置 CA 证书，保留内容可直接保存。' : '当前未配置 CA 证书。' }}</div>
          </el-form-item>
          <el-form-item label="客户端证书">
            <el-input v-model="formData.clientCert" type="textarea" :rows="4" placeholder="PEM 内容，可选" />
            <div class="form-help">{{ formData.hasClientCert ? '当前已配置客户端证书。' : '当前未配置客户端证书。' }}</div>
          </el-form-item>
          <el-form-item label="客户端私钥">
            <el-input
              v-model="formData.clientKey"
              type="textarea"
              :rows="4"
              placeholder="PEM 内容，可选，编辑时留空表示保留原值"
            />
            <div class="form-help">{{ formData.hasClientKey ? '当前已配置客户端私钥，留空表示保留原值。' : '当前未配置客户端私钥。' }}</div>
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false" :disabled="dialogLoading">取消</el-button>
        <el-button @click="handleSubmit(false)" :loading="saving" :disabled="dialogLoading">保存</el-button>
        <el-button type="primary" @click="handleSubmit(true)" :loading="saving" :disabled="dialogLoading">保存并测试</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { createKafkaCluster, deleteKafkaCluster, getKafkaCluster, getKafkaClusters, testKafkaCluster, updateKafkaCluster } from '@/api/kafka.js'
import { usePermissionStore } from '@/stores/permissionStore.js'
import { formatDateTime } from '@/utils/dateTime.js'
import { confirmKafkaRiskAction } from '@/utils/kafkaRiskConfirm.js'

const permStore = usePermissionStore()
const loading = ref(false)
const saving = ref(false)
const dialogLoading = ref(false)
const testingId = ref(null)
const editingId = ref(null)
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref()
const clusters = ref([])
const paginationPage = ref(1)
const paginationPageSize = ref(20)
const paginationTotal = ref(0)
const keyword = ref('')
const status = ref('')
const environment = ref('')
const tenant = ref('')

const emptyForm = () => ({
  id: null,
  name: '',
  bootstrapServers: '',
  version: '3.6.0',
  environment: '',
  tenant: '',
  authType: 'none',
  username: '',
  password: '',
  tlsEnabled: false,
  insecureSkipVerify: false,
  caCert: '',
  clientCert: '',
  clientKey: '',
  hasCACert: false,
  hasClientCert: false,
  hasClientKey: false,
  description: '',
})

const formData = reactive(emptyForm())
const bootstrapServerPattern = /^(?:\[[0-9A-Fa-f:.]+\]|[A-Za-z0-9._-]+):\d{1,5}$/
const kafkaVersionPattern = /^\d+\.\d+\.\d+(?:\.\d+)?$/
const rules = {
  name: [{ required: true, message: '请输入集群名称', trigger: 'blur' }],
  bootstrapServers: [
    { required: true, message: '请输入 Bootstrap Servers', trigger: 'blur' },
    {
      trigger: 'blur',
      validator: (_rule, value, callback) => {
        const items = String(value || '').split(',').map((item) => item.trim()).filter(Boolean)
        if (items.length === 0) {
          callback(new Error('请输入至少一个 host:port'))
          return
        }
        if (items.some((item) => !bootstrapServerPattern.test(item))) {
          callback(new Error('Bootstrap Servers 格式应为 host:port，多个地址用逗号分隔'))
          return
        }
        callback()
      },
    },
  ],
  version: [{
    trigger: 'blur',
    validator: (_rule, value, callback) => {
      const version = String(value || '').trim()
      if (!version || kafkaVersionPattern.test(version)) {
        callback()
        return
      }
      callback(new Error('Kafka 版本格式应为 3.6.0 或 0.10.2.0'))
    },
  }],
}

const clusterStats = computed(() => {
  const now = Date.now()
  return {
    total: paginationTotal.value,
    active: clusters.value.filter((item) => item.status === 'active').length,
    error: clusters.value.filter((item) => item.status === 'error').length,
    failedRecently: clusters.value.filter((item) => {
      if (item.status !== 'error' || !item.lastTestedAt) return false
      const testedAt = new Date(item.lastTestedAt).getTime()
      return !Number.isNaN(testedAt) && now - testedAt <= 24 * 60 * 60 * 1000
    }).length,
  }
})

const failingCluster = computed(() =>
  clusters.value
    .filter((item) => item.status === 'error')
    .sort((a, b) => new Date(b.lastTestedAt || 0) - new Date(a.lastTestedAt || 0))
    .at(0) || null,
)

const authRiskSummary = computed(() => ({
  noAuthCount: clusters.value.filter((item) => item.authType === 'none').length,
  tlsDisabledCount: clusters.value.filter((item) => !item.tlsEnabled).length,
  insecureSkipVerifyCount: clusters.value.filter((item) => item.insecureSkipVerify).length,
}))

const authRiskClusterCount = computed(() =>
  clusters.value
    .filter((item) => item.authType === 'none' || !item.tlsEnabled || item.insecureSkipVerify)
    .length,
)

const resetForm = () => Object.assign(formData, emptyForm())
const statusLabel = (value) => ({ active: '正常', error: '异常', unknown: '未知' }[value] || value || '未知')
const statusType = (value) => ({ active: 'success', error: 'danger', unknown: 'info' }[value] || 'info')
const formatTime = formatDateTime

const buildClusterPayload = () => ({
  name: formData.name,
  bootstrapServers: formData.bootstrapServers,
  version: formData.version,
  environment: formData.environment,
  tenant: formData.tenant,
  authType: formData.authType,
  username: formData.username,
  password: formData.password,
  tlsEnabled: formData.tlsEnabled,
  insecureSkipVerify: formData.tlsEnabled ? formData.insecureSkipVerify : false,
  caCert: formData.caCert,
  clientCert: formData.clientCert,
  clientKey: formData.clientKey,
  description: formData.description,
})

const loadClusters = async () => {
  if (loading.value) return
  loading.value = true
  try {
    const res = await getKafkaClusters({
      page: paginationPage.value,
      pageSize: paginationPageSize.value,
      keyword: keyword.value,
      status: status.value,
      environment: environment.value,
      tenant: tenant.value,
    })
    clusters.value = res?.data?.data?.list || []
    paginationTotal.value = Number(res?.data?.data?.total || 0)
    if (paginationPage.value > 1 && clusters.value.length === 0 && paginationTotal.value > 0) {
      paginationPage.value = 1
      await loadClusters()
      return
    }
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 集群列表加载失败')
  } finally {
    loading.value = false
  }
}

const handleQuery = async () => {
  paginationPage.value = 1
  await loadClusters()
}

const handlePageChange = async (page) => {
  paginationPage.value = page
  await loadClusters()
}

const openCreateDialog = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = async (row) => {
  isEdit.value = true
  resetForm()
  dialogVisible.value = true
  dialogLoading.value = true
  editingId.value = row.id
  try {
    const res = await getKafkaCluster(row.id)
    const detail = res?.data?.data || {}
    Object.assign(formData, { ...row, ...detail, password: '', clientKey: '' })
  } catch (error) {
    dialogVisible.value = false
    ElMessage.error(error.message || 'Kafka 集群详情加载失败')
  } finally {
    dialogLoading.value = false
    editingId.value = null
  }
}

const handleSubmit = async (testAfterSave) => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  saving.value = true
  try {
    let saved
    const payload = buildClusterPayload()
    if (isEdit.value) saved = await updateKafkaCluster(formData.id, payload)
    else saved = await createKafkaCluster(payload)
    const clusterId = saved?.data?.data?.id || formData.id
    ElMessage.success(isEdit.value ? 'Kafka 集群已更新' : 'Kafka 集群已创建')
    dialogVisible.value = false
    await loadClusters()
    if (testAfterSave && clusterId) await handleTest({ id: clusterId })
  } catch (error) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    saving.value = false
  }
}

const handleTest = async (row) => {
  testingId.value = row.id
  try {
    const res = await testKafkaCluster(row.id)
    const result = res?.data?.data
    if (result?.status === 'error') {
      ElMessage.error(result?.errorMessage || '连接测试失败')
    } else {
      ElMessage.success(`连接测试成功，Broker 数: ${result?.brokerCount ?? '-'}`)
    }
    await loadClusters()
  } catch (error) {
    ElMessage.error(error.message || '连接测试失败')
    await loadClusters()
  } finally {
    testingId.value = null
  }
}

const handleDelete = async (row) => {
  const confirmed = await confirmKafkaRiskAction({
    title: '删除集群确认',
    resourceName: row.name,
    actionLabel: '删除 Kafka 集群连接',
    dangerPoints: [
      '会删除当前平台保存的连接配置和环境归属信息',
      '不会删除真实 Kafka 集群数据，但相关页面将无法继续访问该集群',
      row.status === 'active' ? '该集群当前状态正常，删除前请确认不是仍在使用的入口' : '该集群当前状态异常，删除前可先尝试连接测试',
    ],
    confirmButtonText: '确认删除',
  })
  if (!confirmed) return
  try {
    await deleteKafkaCluster(row.id)
    ElMessage.success('Kafka 集群已删除')
    await loadClusters()
  } catch (error) {
    ElMessage.error(error.message || '删除失败')
  }
}

watch(() => formData.tlsEnabled, (enabled) => {
  if (!enabled) {
    formData.insecureSkipVerify = false
    formData.caCert = ''
    formData.clientCert = ''
    formData.clientKey = ''
    formData.hasCACert = false
    formData.hasClientCert = false
    formData.hasClientKey = false
  }
})

watch(() => formData.authType, (authType) => {
  if (authType === 'none') {
    formData.username = ''
    formData.password = ''
  }
})

onMounted(loadClusters)
</script>

<style scoped>
.form-help {
  margin-top: 8px;
  color: var(--shell-text-soft);
  font-size: 12px;
  line-height: 1.5;
}

.table-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
