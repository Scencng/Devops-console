<template>
  <div class="page-container">
    <el-card class="page-header-card">
      <div class="page-header">
        <div>
          <h2>自动发现</h2>
          <p>按 CIDR 网段和端口批量扫描 Kafka，识别候选节点并一键导入为集群</p>
        </div>
      </div>
    </el-card>

    <el-card class="content-card">
      <template #header><div class="card-header">扫描条件</div></template>
      <el-form label-position="top">
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="CIDR 网段">
              <el-input v-model="scanForm.cidr" placeholder="例如 192.168.1.0/24" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="端口列表">
              <el-input v-model="portsInput" placeholder="例如 9092,9093,29092" />
            </el-form-item>
          </el-col>
          <el-col :span="4">
            <el-form-item label="超时(ms)">
              <el-input-number v-model="scanForm.timeoutMs" :min="200" :max="30000" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="4">
            <el-form-item label="并发">
              <el-input-number v-model="scanForm.concurrency" :min="1" :max="1024" style="width:100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :span="4">
            <el-form-item label="认证模板">
              <el-select v-model="authMode" style="width:100%" @change="applyAuthTemplate">
                <el-option label="无认证" value="none" />
                <el-option label="SASL/PLAIN" value="plain" />
                <el-option label="SCRAM-SHA-256" value="scram_sha256" />
                <el-option label="SCRAM-SHA-512" value="scram_sha512" />
                <el-option label="TLS" value="tls" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="4">
            <el-form-item label="Kafka 版本">
              <el-input v-model="scanForm.auth.version" placeholder="留空自动探测，例如 3.9.0" />
            </el-form-item>
          </el-col>
          <el-col :span="5">
            <el-form-item label="用户名">
              <el-input v-model="scanForm.auth.username" :disabled="authMode === 'none' || authMode === 'tls'" />
            </el-form-item>
          </el-col>
          <el-col :span="5">
            <el-form-item label="密码">
              <el-input v-model="scanForm.auth.password" type="password" show-password :disabled="authMode === 'none' || authMode === 'tls'" />
            </el-form-item>
          </el-col>
          <el-col :span="3">
            <el-form-item label="TLS">
              <el-switch v-model="scanForm.auth.tlsEnabled" />
            </el-form-item>
          </el-col>
          <el-col :span="3">
            <el-form-item label="跳过校验">
              <el-switch v-model="scanForm.auth.insecureSkipVerify" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item v-if="scanForm.auth.tlsEnabled" label="CA 证书">
          <el-input v-model="scanForm.auth.caCert" type="textarea" :rows="4" />
        </el-form-item>
        <el-form-item v-if="scanForm.auth.tlsEnabled" label="客户端证书">
          <el-input v-model="scanForm.auth.clientCert" type="textarea" :rows="4" />
        </el-form-item>
        <el-form-item v-if="scanForm.auth.tlsEnabled" label="客户端私钥">
          <el-input v-model="scanForm.auth.clientKey" type="textarea" :rows="4" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="loading" @click="runScan">开始扫描</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="content-card" v-loading="loading">
      <template #header><div class="card-header">扫描结果</div></template>
      <el-table :data="results" empty-text="暂无扫描结果">
        <el-table-column prop="ip" label="IP" width="150" />
        <el-table-column prop="port" label="端口" width="100" />
        <el-table-column label="是否像 Kafka" width="130">
          <template #default="{ row }">
            <el-tag :type="row.looksLikeKafka ? 'success' : 'info'">{{ row.looksLikeKafka ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Kafka 版本" width="140">
          <template #default="{ row }">
            <el-tag v-if="row.kafkaVersion" type="success">{{ row.kafkaVersion }}</el-tag>
            <el-tag v-else-if="row.versionDetectError" type="warning">探测失败</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="brokerId" label="Broker ID" width="110" />
        <el-table-column prop="clusterId" label="Cluster ID" min-width="220" show-overflow-tooltip />
        <el-table-column prop="controllerId" label="Controller" width="110" />
        <el-table-column label="listeners" min-width="260" show-overflow-tooltip>
          <template #default="{ row }">{{ (row.listeners || []).join(', ') || '-' }}</template>
        </el-table-column>
        <el-table-column label="错误信息" min-width="260" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.versionDetectError || row.errorMessage || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" :disabled="!row.looksLikeKafka" @click="openImportDialog(row)">
              {{ row.versionDetectError ? '手动确认版本后导入' : '导入为集群' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="importVisible" title="导入为 Kafka 集群" width="680px" destroy-on-close>
      <el-form label-position="top" :model="importForm">
        <el-alert
          v-if="importSource?.versionDetectError"
          type="warning"
          :closable="false"
          show-icon
          :title="`Kafka 版本自动探测失败：${importSource.versionDetectError}`"
          style="margin-bottom: 16px;"
        />
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="集群名称">
              <el-input v-model="importForm.name" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="地址">
              <el-input v-model="importForm.address" disabled />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="环境"><el-input v-model="importForm.environment" placeholder="dev/test/prod" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="租户"><el-input v-model="importForm.tenant" placeholder="例如 core-team" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="Kafka 版本">
              <el-input
                v-model="importForm.auth.version"
                :placeholder="importSource?.versionDetectError ? '请手动填写，例如 3.9.0' : '自动探测成功时会自动回填'"
                :disabled="!!importSource?.kafkaVersion && !importSource?.versionDetectError"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12" v-if="importSource?.kafkaVersion && !importSource?.versionDetectError">
            <el-form-item label="版本来源">
              <el-input :model-value="`自动探测: ${importSource.kafkaVersion}`" disabled />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述">
          <el-input v-model="importForm.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button type="primary" :loading="importing" @click="importResult">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { importKafkaDiscoveryResult, scanKafkaNetwork } from '@/api/kafka.js'

const LAST_DISCOVERY_VERSION_KEY = 'kafka-console:last-discovery-version'
const DEFAULT_FALLBACK_VERSION = '3.9.0'

const loading = ref(false)
const importing = ref(false)
const authMode = ref('none')
const portsInput = ref('9092,9093,29092')
const results = ref([])
const importVisible = ref(false)
const importSource = ref(null)

const scanForm = reactive({
  cidr: '',
  timeoutMs: 2500,
  concurrency: 64,
  auth: {
    version: '',
    authType: 'none',
    username: '',
    password: '',
    tlsEnabled: false,
    insecureSkipVerify: false,
    caCert: '',
    clientCert: '',
    clientKey: '',
  },
})

const importForm = reactive({
  name: '',
  address: '',
  environment: '',
  tenant: '',
  description: '',
  auth: {},
})

const applyAuthTemplate = () => {
  scanForm.auth.authType = authMode.value === 'tls' ? 'none' : authMode.value
  scanForm.auth.tlsEnabled = authMode.value === 'tls'
  if (authMode.value === 'none') {
    scanForm.auth.username = ''
    scanForm.auth.password = ''
    scanForm.auth.tlsEnabled = false
  }
}

const getRememberedKafkaVersion = () => {
  const remembered = localStorage.getItem(LAST_DISCOVERY_VERSION_KEY)?.trim()
  if (remembered) return remembered
  return DEFAULT_FALLBACK_VERSION
}

const rememberKafkaVersion = (version) => {
  const normalized = String(version || '').trim()
  if (!normalized) return
  localStorage.setItem(LAST_DISCOVERY_VERSION_KEY, normalized)
}

const parsePorts = () => portsInput.value.split(',').map((item) => Number(item.trim())).filter((item) => Number.isInteger(item) && item > 0 && item <= 65535)

const runScan = async () => {
  const ports = parsePorts()
  if (!scanForm.cidr || ports.length === 0) {
    ElMessage.warning('请填写 CIDR 和至少一个有效端口')
    return
  }
  loading.value = true
  try {
    const res = await scanKafkaNetwork({
      cidr: scanForm.cidr.trim(),
      ports,
      timeoutMs: Number(scanForm.timeoutMs),
      concurrency: Number(scanForm.concurrency),
      auth: { ...scanForm.auth },
    })
    results.value = res?.data?.data || []
    ElMessage.success(`扫描完成，共返回 ${results.value.length} 条结果`)
  } catch (error) {
    ElMessage.error(error.message || '扫描失败')
  } finally {
    loading.value = false
  }
}

const openImportDialog = (row) => {
  importSource.value = row
  importForm.name = `kafka-${row.ip.replaceAll('.', '-')}-${row.port}`
  importForm.address = row.address
  importForm.environment = ''
  importForm.tenant = ''
  importForm.description = `自动发现导入，ClusterID=${row.clusterId || '-'}`
  importForm.auth = {
    ...scanForm.auth,
    version: row.kafkaVersion || getRememberedKafkaVersion(),
  }
  importVisible.value = true
}

const importResult = async () => {
  if (!importForm.name || !importForm.address) {
    ElMessage.warning('请填写集群名称')
    return
  }
  if (importSource.value?.versionDetectError && !String(importForm.auth.version || '').trim()) {
    ElMessage.warning('自动探测失败时，请在弹窗中手动填写 Kafka 版本后再导入')
    return
  }
  importing.value = true
  try {
    await importKafkaDiscoveryResult({
      name: importForm.name.trim(),
      address: importForm.address,
      environment: importForm.environment.trim(),
      tenant: importForm.tenant.trim(),
      description: importForm.description,
      auth: importForm.auth,
    })
    rememberKafkaVersion(importForm.auth.version)
    ElMessage.success('导入成功')
    importVisible.value = false
  } catch (error) {
    ElMessage.error(error.message || '导入失败')
  } finally {
    importing.value = false
  }
}
</script>
