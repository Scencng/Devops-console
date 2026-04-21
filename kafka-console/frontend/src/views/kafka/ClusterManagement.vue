<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <div class="page-eyebrow">Kafka Clusters</div>
          <h2>Kafka 集群管理</h2>
          <p>维护连接地址、认证方式和环境归属，并在变更后立即完成连通性校验。</p>
        </div>
      </div>
    </el-card>

    <div class="page-metrics">
      <div class="page-metric-card is-accent">
        <span>集群总数</span>
        <strong>{{ clusterStats.total }}</strong>
        <p>当前筛选范围内可管理的 Kafka 集群。</p>
      </div>
      <div class="page-metric-card is-success">
        <span>状态正常</span>
        <strong>{{ clusterStats.active }}</strong>
        <p>最近一次测试成功、状态为正常的集群。</p>
      </div>
      <div class="page-metric-card is-warning">
        <span>状态异常</span>
        <strong>{{ clusterStats.error }}</strong>
        <p>需要优先检查网络、认证或版本配置。</p>
      </div>
      <div class="page-metric-card">
        <span>覆盖环境</span>
        <strong>{{ clusterStats.environments }}</strong>
        <p>当前列表里已经接入的环境数量。</p>
      </div>
      <div class="page-metric-card is-warning">
        <span>最近测试失败</span>
        <strong>{{ clusterStats.failedRecently }}</strong>
        <p>最近一次连接测试失败的集群数量，建议优先核查网络和认证信息。</p>
      </div>
    </div>

    <el-card v-if="failingClusters.length > 0" class="content-card">
      <template #header>
        <div class="card-header card-header-wrap">
          <span>最近测试失败原因</span>
          <span class="card-subtitle">把常见故障线索直接放在当前页，减少来回排查的切换成本</span>
        </div>
      </template>

      <div class="failure-grid">
        <div class="workspace-panel">
          <h3>失败集群</h3>
          <p>优先查看最近测试失败的集群、错误信息和对应环境归属。</p>
          <div class="compact-list">
            <div v-for="item in failingClusters" :key="item.id" class="compact-item">
              <div>
                <strong>{{ item.name }}</strong>
                <span>{{ item.environment || '未标记环境' }} / {{ item.bootstrapServers }}</span>
              </div>
              <div class="failure-actions">
                <el-tag type="danger">异常</el-tag>
                <el-button link type="primary" :loading="testingId === item.id" @click="handleTest(item)">重新测试</el-button>
              </div>
            </div>
          </div>
        </div>

        <div class="workspace-panel">
          <h3>快速诊断提示</h3>
          <p>根据当前失败信息优先排查最常见的三类问题。</p>
          <div class="compact-list">
            <div class="compact-item">
              <div>
                <strong>网络与地址</strong>
                <span>确认 Bootstrap Servers 可达，容器网络、端口映射、防火墙和 DNS 是否正确。</span>
              </div>
            </div>
            <div class="compact-item">
              <div>
                <strong>认证与 TLS</strong>
                <span>检查 SASL 类型、用户名密码、TLS 开关、证书链和跳过校验选项是否匹配真实集群。</span>
              </div>
            </div>
            <div class="compact-item">
              <div>
                <strong>版本与协议</strong>
                <span>当 Kafka 版本、认证机制或 Broker 配置不匹配时，客户端握手也会直接失败。</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <el-card class="content-card">
      <template #header>
        <div class="card-header card-header-wrap">
          <span>认证 / TLS 配置风险摘要</span>
          <span class="card-subtitle">优先识别未启用认证、TLS 配置较弱或证书校验被跳过的集群入口</span>
        </div>
      </template>

      <div class="workbench-grid">
        <div class="workspace-panel">
          <h3>配置风险概览</h3>
          <p>基于当前集群列表，快速判断哪些连接配置安全边界较弱。</p>
          <div class="compact-list">
            <div class="compact-item">
              <div>
                <strong>无认证集群</strong>
                <span>当前共有 {{ authRiskSummary.noAuthCount }} 个集群使用无认证方式。</span>
              </div>
            </div>
            <div class="compact-item">
              <div>
                <strong>TLS 未启用</strong>
                <span>当前共有 {{ authRiskSummary.tlsDisabledCount }} 个集群未启用 TLS。</span>
              </div>
            </div>
            <div class="compact-item">
              <div>
                <strong>跳过证书校验</strong>
                <span>当前共有 {{ authRiskSummary.insecureSkipVerifyCount }} 个集群启用了跳过证书校验。</span>
              </div>
            </div>
          </div>
        </div>

        <div class="workspace-panel">
          <h3>重点关注集群</h3>
          <p>以下集群的认证/TLS 配置更值得优先排查和加固。</p>
          <div class="compact-list">
            <div v-for="item in authRiskClusters" :key="item.id" class="compact-item">
              <div>
                <strong>{{ item.name }}</strong>
                <span>{{ item.authRiskReason }}</span>
              </div>
              <el-tag :type="item.authRiskLevel === 'high' ? 'danger' : 'warning'">
                {{ item.authRiskLevel === 'high' ? '高风险' : '关注' }}
              </el-tag>
            </div>
          </div>
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
            @keyup.enter="loadClusters"
          />
          <el-input
            v-model="environment"
            placeholder="环境"
            style="width: 140px"
            clearable
            @keyup.enter="loadClusters"
          />
          <el-input
            v-model="tenant"
            placeholder="租户"
            style="width: 140px"
            clearable
            @keyup.enter="loadClusters"
          />
          <el-select v-model="status" placeholder="状态" clearable style="width: 150px" @change="loadClusters">
            <el-option label="全部" value="" />
            <el-option label="正常" value="active" />
            <el-option label="异常" value="error" />
            <el-option label="未知" value="unknown" />
          </el-select>
        </div>
        <div class="toolbar-right">
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
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>集群列表</span>
          <span class="card-subtitle">按名称、环境、租户和状态筛选当前连接信息</span>
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
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑 Kafka 集群' : '新增 Kafka 集群'" width="760px" destroy-on-close>
      <el-form ref="formRef" :model="formData" :rules="rules" label-position="top">
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
            <el-form-item><el-checkbox v-model="formData.insecureSkipVerify">跳过证书校验</el-checkbox></el-form-item>
          </el-col>
        </el-row>

        <template v-if="formData.tlsEnabled">
          <el-form-item label="CA 证书">
            <el-input v-model="formData.caCert" type="textarea" :rows="4" placeholder="PEM 内容，可选" />
          </el-form-item>
          <el-form-item label="客户端证书">
            <el-input v-model="formData.clientCert" type="textarea" :rows="4" placeholder="PEM 内容，可选" />
          </el-form-item>
          <el-form-item label="客户端私钥">
            <el-input
              v-model="formData.clientKey"
              type="textarea"
              :rows="4"
              placeholder="PEM 内容，可选，编辑时留空表示保留原值"
            />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button @click="handleSubmit(false)" :loading="saving">保存</el-button>
        <el-button type="primary" @click="handleSubmit(true)" :loading="saving">保存并测试</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createKafkaCluster, deleteKafkaCluster, getKafkaClusters, testKafkaCluster, updateKafkaCluster } from '@/api/kafka.js'
import { usePermissionStore } from '@/stores/permissionStore.js'
import { openKafkaRiskConfirm } from '@/utils/kafkaRiskConfirm.js'

const permStore = usePermissionStore()
const loading = ref(false)
const saving = ref(false)
const testingId = ref(null)
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref()
const clusters = ref([])
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
  description: '',
})

const formData = reactive(emptyForm())
const rules = {
  name: [{ required: true, message: '请输入集群名称', trigger: 'blur' }],
  bootstrapServers: [{ required: true, message: '请输入 Bootstrap Servers', trigger: 'blur' }],
}

const clusterStats = computed(() => {
  const environments = new Set(clusters.value.map((item) => item.environment).filter(Boolean))
  return {
    total: clusters.value.length,
    active: clusters.value.filter((item) => item.status === 'active').length,
    error: clusters.value.filter((item) => item.status === 'error').length,
    environments: environments.size,
    failedRecently: clusters.value.filter((item) => item.status === 'error' && item.lastTestedAt).length,
  }
})

const failingClusters = computed(() =>
  clusters.value
    .filter((item) => item.status === 'error')
    .sort((a, b) => new Date(b.lastTestedAt || 0) - new Date(a.lastTestedAt || 0))
    .slice(0, 5),
)

const authRiskSummary = computed(() => ({
  noAuthCount: clusters.value.filter((item) => item.authType === 'none').length,
  tlsDisabledCount: clusters.value.filter((item) => !item.tlsEnabled).length,
  insecureSkipVerifyCount: clusters.value.filter((item) => item.insecureSkipVerify).length,
}))

const authRiskClusters = computed(() =>
  clusters.value
    .map((item) => {
      const reasons = []
      let score = 0

      if (item.authType === 'none') {
        reasons.push('未启用认证')
        score += 3
      }
      if (!item.tlsEnabled) {
        reasons.push('未启用 TLS')
        score += 2
      }
      if (item.insecureSkipVerify) {
        reasons.push('跳过证书校验')
        score += 2
      }

      return {
        ...item,
        authRiskScore: score,
        authRiskLevel: score >= 4 ? 'high' : 'medium',
        authRiskReason: reasons.join('；') || '当前未识别到明显认证/TLS 风险',
      }
    })
    .filter((item) => item.authRiskScore > 0)
    .sort((a, b) => b.authRiskScore - a.authRiskScore)
    .slice(0, 5),
)

const resetForm = () => Object.assign(formData, emptyForm())
const statusLabel = (value) => ({ active: '正常', error: '异常', unknown: '未知' }[value] || value || '未知')
const statusType = (value) => ({ active: 'success', error: 'danger', unknown: 'info' }[value] || 'info')
const formatTime = (value) => (value ? new Date(value).toLocaleString() : '-')

const loadClusters = async () => {
  loading.value = true
  try {
    const res = await getKafkaClusters({
      page: 1,
      pageSize: 100,
      keyword: keyword.value,
      status: status.value,
      environment: environment.value,
      tenant: tenant.value,
    })
    clusters.value = res?.data?.data?.list || []
  } catch (error) {
    ElMessage.error(error.message || 'Kafka 集群列表加载失败')
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = (row) => {
  isEdit.value = true
  resetForm()
  Object.assign(formData, { ...row, password: '', clientKey: '' })
  dialogVisible.value = true
}

const handleSubmit = async (testAfterSave) => {
  if (!formRef.value) return
  await formRef.value.validate()
  saving.value = true
  try {
    let saved
    if (isEdit.value) saved = await updateKafkaCluster(formData.id, formData)
    else saved = await createKafkaCluster(formData)
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
    ElMessage.success(`连接测试成功，Broker 数: ${result?.brokerCount ?? '-'}`)
    await loadClusters()
  } catch (error) {
    ElMessage.error(error.message || '连接测试失败')
    await loadClusters()
  } finally {
    testingId.value = null
  }
}

const handleDelete = async (row) => {
  await openKafkaRiskConfirm({
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
  try {
    await deleteKafkaCluster(row.id)
    ElMessage.success('Kafka 集群已删除')
    await loadClusters()
  } catch (error) {
    ElMessage.error(error.message || '删除失败')
  }
}

onMounted(loadClusters)
</script>

<style scoped>
.failure-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(320px, 0.8fr);
  gap: 18px;
}

.failure-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

@media (max-width: 1120px) {
  .failure-grid {
    grid-template-columns: 1fr;
  }
}
</style>
