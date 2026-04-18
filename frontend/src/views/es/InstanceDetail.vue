<template>
  <div class="instance-detail">
    <div class="page-header">
      <div class="header-left">
        <el-button link @click="handleBack">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h2>实例详情: {{ instance?.name }}</h2>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleEdit">
          <el-icon><Edit /></el-icon>
          编辑
        </el-button>
        <el-button type="danger" @click="handleDelete">
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
      </div>
    </div>

    <div v-loading="loading" class="detail-content">
      <!-- 基本信息 -->
      <el-card class="info-card">
        <template #header>
          <div class="card-header">
            <span>基本信息</span>
            <el-tag :type="getStatusType(instance?.status)">
              <el-icon><component :is="getStatusIcon(instance?.status)" /></el-icon>
              {{ getStatusLabel(instance?.status) }}
            </el-tag>
          </div>
        </template>
        <el-row :gutter="20">
          <el-col :span="12">
            <div class="info-grid">
              <div class="info-item">
                <span class="label">实例类型:</span>
                <span class="value">{{ getTypeLabel(instance?.type) }}</span>
              </div>
              <div class="info-item">
                <span class="label">实例名称:</span>
                <span class="value">{{ instance?.name }}</span>
              </div>
              <div class="info-item">
                <span class="label">实例地址:</span>
                <span class="value">{{ instance?.address }}</span>
              </div>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="info-grid">
              <div class="info-item">
                <span class="label">状态:</span>
                <el-tag :type="getStatusType(instance?.status)" size="small">
                  <el-icon><component :is="getStatusIcon(instance?.status)" /></el-icon>
                  {{ getStatusLabel(instance?.status) }}
                </el-tag>
              </div>
              <div class="info-item">
                <span class="label">创建时间:</span>
                <span class="value">{{ formatDateTime(instance?.createdAt) }}</span>
              </div>
              <div class="info-item">
                <span class="label">更新时间:</span>
                <span class="value">{{ formatDateTime(instance?.updatedAt) }}</span>
              </div>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <!-- 认证配置 -->
      <el-card class="info-card">
        <template #header>
          <div class="card-header">
            <span>认证配置</span>
            <el-button link @click="handleEditAuth">
              <el-icon><Edit /></el-icon>
              编辑认证
            </el-button>
          </div>
        </template>
        <div v-if="authConfigs.length > 0">
          <el-descriptions :column="2" border>
            <el-descriptions-item
              v-for="config in authConfigs"
              :key="config.id"
              :label="config.configKey"
            >
              <div class="auth-value">
                <span v-if="config.isEncrypted" class="encrypted-value">
                  ••••••••
                </span>
                <span v-else>{{ config.configValue }}</span>
                <el-button
                  v-if="config.isEncrypted"
                  type="text"
                  size="small"
                  @click="togglePasswordVisibility(config.id)"
                >
                  <el-icon>
                    <component :is="passwordVisible[config.id] ? Hide : View" />
                  </el-icon>
                </el-button>
              </div>
            </el-descriptions-item>
          </el-descriptions>
        </div>
        <div v-else class="no-auth">
          <el-empty description="暂无认证配置" :image-size="100">
            <el-button type="primary" @click="handleEditAuth">添加认证配置</el-button>
          </el-empty>
        </div>
        <div class="auth-actions">
          <el-button @click="handleTestConnection" :loading="testing">
            <el-icon><Connection /></el-icon>
            测试连接
          </el-button>
        </div>
      </el-card>

      <!-- 连接测试历史 -->
      <el-card class="info-card">
        <template #header>
          <div class="card-header">
            <span>连接测试历史</span>
            <el-button link @click="refreshTestHistory">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </template>
        <el-table :data="testHistory" stripe>
          <el-table-column prop="testedAt" label="测试时间" width="180">
            <template #default="{ row }">
              {{ formatDateTime(row.testedAt) }}
            </template>
          </el-table-column>
          <el-table-column prop="testResult" label="结果" width="100">
            <template #default="{ row }">
              <el-tag :type="getTestResultType(row.testResult)" size="small">
                <el-icon><component :is="getTestResultIcon(row.testResult)" /></el-icon>
                {{ getTestResultLabel(row.testResult) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="responseTime" label="响应时间" width="120">
            <template #default="{ row }">
              <span v-if="row.responseTime">{{ row.responseTime }}ms</span>
              <span v-else class="text-muted">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="errorMessage" label="错误信息">
            <template #default="{ row }">
              <span v-if="row.errorMessage" class="error-message">{{ row.errorMessage }}</span>
              <span v-else class="text-muted">-</span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 操作日志 -->
      <el-card class="info-card">
        <template #header>
          <div class="card-header">
            <span>操作日志</span>
            <el-button link @click="refreshOperationLogs">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </template>
        <el-timeline>
          <el-timeline-item
            v-for="log in operationLogs"
            :key="log.id"
            :timestamp="formatDateTime(log.timestamp)"
            :type="getLogType(log.action)"
          >
            <div class="log-content">
              <span class="log-action">{{ log.action }}</span>
              <span class="log-user">用户 {{ log.user }}</span>
            </div>
          </el-timeline-item>
        </el-timeline>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import {computed, onMounted, ref} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {ElMessage, ElMessageBox} from 'element-plus'
import {
  ArrowLeft,
  CircleCheck,
  CircleClose,
  Connection,
  Delete,
  Edit,
  Refresh,
  WarningFilled
} from '@element-plus/icons-vue'
import {getTestHistory} from '@/api/instance.js'

const route = useRoute()
const router = useRouter()

// 数据状态
const loading = ref(false)
const testing = ref(false)
const passwordVisible = ref({})

// 实例数据
const instance = ref(null)

// 认证配置
const authConfigsList = ref([])

// 测试历史
const testHistory = ref([])

// 操作日志
const operationLogs = ref([])

// 计算属性
const instanceId = computed(() => parseInt(route.params.id))

// 类型相关方法
const getTypeLabel = (type) => {
  const labels = {
    elasticsearch: 'Elasticsearch',
    kibana: 'Kibana',
    logstash: 'Logstash',
    filebeat: 'Filebeat',
    metricbeat: 'Metricbeat',
    apm: 'APM'
  }
  return labels[type] || type
}

// 状态相关方法
const getStatusIcon = (status) => {
  const icons = {
    active: CircleCheck,
    online: CircleCheck,
    offline: CircleClose,
    warning: WarningFilled,
    error: CircleClose
  }
  return icons[status] || CircleClose
}

const getStatusType = (status) => {
  const types = {
    active: 'success',
    online: 'success',
    offline: 'danger',
    warning: 'warning',
    error: 'danger'
  }
  return types[status] || 'info'
}

const getStatusLabel = (status) => {
  const labels = {
    active: '活跃(旧)',
    online: '在线',
    offline: '离线',
    warning: '警告',
    error: '错误'
  }
  return labels[status] || status
}

// 测试结果相关方法
const getTestResultIcon = (result) => {
  const icons = {
    success: CircleCheck,
    warning: WarningFilled,
    error: CircleClose
  }
  return icons[result] || CircleClose
}

const getTestResultType = (result) => {
  const types = {
    success: 'success',
    warning: 'warning',
    error: 'danger'
  }
  return types[result] || 'info'
}

const getTestResultLabel = (result) => {
  const labels = {
    success: '成功',
    warning: '警告',
    error: '失败'
  }
  return labels[result] || result
}

// 日志类型
const getLogType = (action) => {
  const types = {
    '创建实例': 'primary',
    '更新实例': 'primary',
    '删除实例': 'danger',
    '测试连接': 'success',
    '更新认证配置': 'warning'
  }
  return types[action] || 'info'
}

// 格式化时间
const formatDateTime = (timeStr) => {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN')
}

// 事件处理
const handleBack = () => {
  router.push('/es/instances')
}

const handleEdit = () => {
  router.push(`/es/instances/edit/${instanceId.value}`)
}

const handleDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除实例 "${instance.value.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 模拟删除操作
    await new Promise(resolve => setTimeout(resolve, 500))

    ElMessage.success('删除成功')
    router.push('/es/instances')
  } catch (error) {
    if (error.message !== 'cancel') {
      ElMessage.error('删除失败: ' + error.message)
    }
  }
}

const handleEditAuth = () => {
  ElMessage.info('编辑认证配置功能待实现')
}

const handleTestConnection = async () => {
  testing.value = true
  try {
    // 模拟测试连接
    await new Promise(resolve => setTimeout(resolve, 1000))

    // 随机生成测试结果
    const isSuccessful = Math.random() > 0.3
    const responseTime = Math.floor(Math.random() * 200) + 20

    if (isSuccessful) {
      ElMessage.success(`连接测试成功 (响应时间: ${responseTime}ms)`)
      instance.value.status = 'active'
    } else {
      const errorMsgs = ['连接超时', '连接被拒绝', '认证失败', '网络不可达']
      const errorMsg = errorMsgs[Math.floor(Math.random() * errorMsgs.length)]
      ElMessage.error(`连接测试失败: ${errorMsg}`)
      instance.value.status = 'error'
    }

    // 更新实例状态
    instance.value.updatedAt = new Date().toISOString()

    await refreshTestHistory()
  } catch (error) {
    const errorMsg = error.response?.data?.message || error.message || '连接测试失败'
    ElMessage.error(`连接测试失败: ${errorMsg}`)
  } finally {
    testing.value = false
  }
}

const togglePasswordVisibility = (configId) => {
  passwordVisible.value[configId] = !passwordVisible.value[configId]
}

const refreshTestHistory = async () => {
  try {
    const response = await getTestHistory(instanceId.value, { page_size: 10 })

    // 使用后端返回的数据，并转换为前端需要的格式
    const rawTests = response.data.data.tests || response.data.data.test_list || []
    testHistory.value = rawTests.map(test => ({
      id: test.id,
      testedAt: test.tested_at,
      testResult: test.test_result,
      responseTime: test.response_time,
      errorMessage: test.error_message
    }))

    if (testHistory.value.length === 0) {
      ElMessage.info('暂无测试历史记录')
    } else {
      ElMessage.success(`已加载 ${testHistory.value.length} 条测试历史记录`)
    }
  } catch (error) {
    ElMessage.error('加载测试历史失败: ' + error.message)
    testHistory.value = []
  }
}

const refreshOperationLogs = async () => {
  ElMessage.info('刷新操作日志...')

  // 使用死数据
  const mockOperationLogs = [
    {
      id: 1,
      action: '创建实例',
      user: 'admin',
      timestamp: '2025-11-10T08:00:00Z'
    },
    {
      id: 2,
      action: '更新实例',
      user: 'admin',
      timestamp: '2025-11-12T14:30:00Z'
    },
    {
      id: 3,
      action: '测试连接',
      user: 'admin',
      timestamp: '2025-11-14T10:30:00Z'
    },
    {
      id: 4,
      action: '更新认证配置',
      user: 'admin',
      timestamp: '2025-11-13T16:45:00Z'
    },
    {
      id: 5,
      action: '测试连接',
      user: 'admin',
      timestamp: '2025-11-14T11:45:00Z'
    }
  ]

  operationLogs.value = mockOperationLogs

  await new Promise(resolve => setTimeout(resolve, 500))
  ElMessage.success('操作日志已更新')
}

// 加载实例数据
const loadInstanceData = async () => {
  loading.value = true
  try {
    // 使用死数据
    const mockInstances = {
      1: {
        id: 1,
        name: 'Elasticsearch-主节点',
        type: 'elasticsearch',
        address: '192.168.1.100:9200',
        https_enabled: true,
        skip_ssl_verify: false,
        status: 'active',
        createdAt: '2025-11-10T08:00:00Z',
        updatedAt: '2025-11-14T10:30:00Z',
        auth_configs: 'basic:{"username":"elastic","password":"changeme"}'
      },
      2: {
        id: 2,
        name: 'Kibana-可视化',
        type: 'kibana',
        address: '192.168.1.101:5601',
        https_enabled: true,
        skip_ssl_verify: false,
        status: 'active',
        createdAt: '2025-11-10T08:30:00Z',
        updatedAt: '2025-11-14T09:45:00Z',
        auth_configs: 'basic:{"username":"kibana","password":"changeme"}'
      },
      3: {
        id: 3,
        name: 'Logstash-数据处理',
        type: 'logstash',
        address: '192.168.1.102:5044',
        https_enabled: false,
        skip_ssl_verify: false,
        status: 'warning',
        createdAt: '2025-11-10T09:00:00Z',
        updatedAt: '2025-11-14T08:20:00Z',
        auth_configs: ''
      },
      4: {
        id: 4,
        name: 'Filebeat-日志收集',
        type: 'filebeat',
        address: '192.168.1.103:9000',
        https_enabled: false,
        skip_ssl_verify: false,
        status: 'error',
        createdAt: '2025-11-10T09:30:00Z',
        updatedAt: '2025-11-14T07:15:00Z',
        auth_configs: ''
      },
      5: {
        id: 5,
        name: 'Metricbeat-指标收集',
        type: 'metricbeat',
        address: '192.168.1.104:9200',
        https_enabled: true,
        skip_ssl_verify: false,
        status: 'active',
        createdAt: '2025-11-10T10:00:00Z',
        updatedAt: '2025-11-14T11:00:00Z',
        auth_configs: 'api_key:{"apiKey":"base64encodedkey","apiId":"test"}'
      },
      6: {
        id: 6,
        name: 'APM-性能监控',
        type: 'apm',
        address: '192.168.1.105:8200',
        https_enabled: true,
        skip_ssl_verify: false,
        status: 'active',
        createdAt: '2025-11-10T10:30:00Z',
        updatedAt: '2025-11-14T10:15:00Z',
        auth_configs: 'token:{"token":"secret_token","tokenExpiresAt":"2025-12-31T23:59:59Z"}'
      }
    }

    instance.value = mockInstances[instanceId.value]

    if (!instance.value) {
      ElMessage.error('实例不存在')
      router.push('/es/instances')
      return
    }

    // 解析认证配置
    if (instance.value.auth_configs) {
      try {
        // auth_configs格式: "basic:{\"username\":\"elastic\",\"password\":\"changeme\"}"
        const colonIndex = instance.value.auth_configs.indexOf(':')
        if (colonIndex > -1) {
          const authType = instance.value.auth_configs.substring(0, colonIndex)
          const configJson = instance.value.auth_configs.substring(colonIndex + 1)
          const config = JSON.parse(configJson)

          authConfigsList.value = [{
            id: 1,
            configKey: authType,
            configValue: JSON.stringify(config),
            isEncrypted: authType !== 'none'
          }]
        }
      } catch (e) {
        console.warn('解析认证配置失败', e)
        authConfigsList.value = []
      }
    }

    // 加载测试历史
    await refreshTestHistory()

  } catch (error) {
    ElMessage.error('加载实例数据失败: ' + error.message)
    router.push('/es/instances')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadInstanceData()
})
</script>

<style scoped>
.instance-detail {
  padding: 20px;
  background-color: #f5f5f5;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  min-height: 0;
  width: 100%;
  height: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-left h2 {
  margin: 0;
  color: #333;
  font-size: 24px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
  flex: 1;
  min-height: 0;
}

.info-card {
  flex-shrink: 0;
}

.info-card:last-child {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.info-card:last-child :deep(.el-card__body) {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.info-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 500;
}

.info-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.info-item .label {
  color: #666;
  font-size: 14px;
  min-width: 80px;
}

.info-item .value {
  color: #333;
  font-size: 14px;
  font-weight: 500;
}

.auth-value {
  display: flex;
  align-items: center;
  gap: 10px;
}

.encrypted-value {
  color: #999;
  font-family: monospace;
}

.no-auth {
  padding: 20px;
  text-align: center;
}

.auth-actions {
  margin-top: 20px;
  text-align: right;
}

.error-message {
  color: #f56c6c;
}

.text-muted {
  color: #999;
}

.log-content {
  display: flex;
  gap: 10px;
}

.log-action {
  font-weight: 500;
  color: #333;
}

.log-user {
  color: #666;
  font-size: 14px;
}


</style>
