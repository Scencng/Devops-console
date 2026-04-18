<template>
  <div class="settings">
    <div class="page-header">
      <h2>系统设置</h2>
    </div>

    <el-row :gutter="20">
      <!-- 左侧菜单 -->
      <el-col :span="6">
        <el-card class="menu-card">
          <el-menu
            :default-active="activeMenu"
            mode="vertical"
            @select="handleMenuSelect"
          >
            <el-menu-item index="general">
              <el-icon><Setting /></el-icon>
              <span>基本设置</span>
            </el-menu-item>
            <el-menu-item index="notification">
              <el-icon><Bell /></el-icon>
              <span>通知设置</span>
            </el-menu-item>
            <el-menu-item index="security">
              <el-icon><Lock /></el-icon>
              <span>安全设置</span>
            </el-menu-item>
            <el-menu-item index="monitoring">
              <el-icon><Monitor /></el-icon>
              <span>监控配置</span>
            </el-menu-item>
            <el-menu-item index="backup">
              <el-icon><FolderOpened /></el-icon>
              <span>备份设置</span>
            </el-menu-item>
            <el-menu-item index="log">
              <el-icon><Document /></el-icon>
              <span>日志配置</span>
            </el-menu-item>
            <el-menu-item index="system">
              <el-icon><Tools /></el-icon>
              <span>系统管理</span>
            </el-menu-item>
          </el-menu>
        </el-card>
      </el-col>

      <!-- 右侧内容 -->
      <el-col :span="18">
        <!-- 基本设置 -->
        <el-card v-show="activeMenu === 'general'">
          <template #header>
            <span>基本设置</span>
          </template>
          <el-form :model="generalSettings" label-width="120px">
            <el-form-item label="系统名称">
              <el-input v-model="generalSettings.systemName" />
            </el-form-item>
            <el-form-item label="系统描述">
              <el-input v-model="generalSettings.systemDescription" type="textarea" :rows="3" />
            </el-form-item>
            <el-form-item label="默认语言">
              <el-select v-model="generalSettings.language" style="width: 100%">
                <el-option label="中文" value="zh-CN" />
                <el-option label="English" value="en-US" />
              </el-select>
            </el-form-item>
            <el-form-item label="时区">
              <el-select v-model="generalSettings.timezone" style="width: 100%">
                <el-option label="UTC+8 北京时间" value="Asia/Shanghai" />
                <el-option label="UTC+0 格林威治时间" value="UTC" />
                <el-option label="UTC-5 纽约时间" value="America/New_York" />
              </el-select>
            </el-form-item>
            <el-form-item label="数据刷新间隔">
              <el-input-number v-model="generalSettings.refreshInterval" :min="5" :max="300" />
              <span style="margin-left: 10px">秒</span>
            </el-form-item>
            <el-form-item label="主题颜色">
              <el-radio-group v-model="generalSettings.theme">
                <el-radio label="light">浅色主题</el-radio>
                <el-radio label="dark">深色主题</el-radio>
                <el-radio label="auto">跟随系统</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveGeneralSettings">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 通知设置 -->
        <el-card v-show="activeMenu === 'notification'">
          <template #header>
            <span>通知设置</span>
          </template>
          <el-form :model="notificationSettings" label-width="120px">
            <el-form-item label="邮件通知">
              <el-switch v-model="notificationSettings.emailEnabled" />
            </el-form-item>
            <el-form-item v-if="notificationSettings.emailEnabled" label="SMTP服务器">
              <el-input v-model="notificationSettings.smtpServer" />
            </el-form-item>
            <el-form-item v-if="notificationSettings.emailEnabled" label="SMTP端口">
              <el-input-number v-model="notificationSettings.smtpPort" :min="1" :max="65535" />
            </el-form-item>
            <el-form-item v-if="notificationSettings.emailEnabled" label="发件人邮箱">
              <el-input v-model="notificationSettings.senderEmail" />
            </el-form-item>
            <el-form-item v-if="notificationSettings.emailEnabled" label="邮箱密码">
              <el-input v-model="notificationSettings.emailPassword" type="password" show-password />
            </el-form-item>
            <el-form-item label="短信通知">
              <el-switch v-model="notificationSettings.smsEnabled" />
            </el-form-item>
            <el-form-item v-if="notificationSettings.smsEnabled" label="短信API地址">
              <el-input v-model="notificationSettings.smsApiUrl" />
            </el-form-item>
            <el-form-item v-if="notificationSettings.smsEnabled" label="API密钥">
              <el-input v-model="notificationSettings.smsApiKey" type="password" show-password />
            </el-form-item>
            <el-form-item label="通知级别">
              <el-checkbox-group v-model="notificationSettings.notificationLevels">
                <el-checkbox label="CRITICAL">严重</el-checkbox>
                <el-checkbox label="WARNING">警告</el-checkbox>
                <el-checkbox label="INFO">信息</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveNotificationSettings">保存设置</el-button>
              <el-button @click="testNotification">测试通知</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 安全设置 -->
        <el-card v-show="activeMenu === 'security'">
          <template #header>
            <span>安全设置</span>
          </template>
          <el-form :model="securitySettings" label-width="120px">
            <el-form-item label="会话超时">
              <el-input-number v-model="securitySettings.sessionTimeout" :min="5" :max="1440" />
              <span style="margin-left: 10px">分钟</span>
            </el-form-item>
            <el-form-item label="密码策略">
              <el-checkbox-group v-model="securitySettings.passwordPolicy">
                <el-checkbox label="minLength">最小长度8位</el-checkbox>
                <el-checkbox label="uppercase">包含大写字母</el-checkbox>
                <el-checkbox label="lowercase">包含小写字母</el-checkbox>
                <el-checkbox label="numbers">包含数字</el-checkbox>
                <el-checkbox label="symbols">包含特殊字符</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item label="登录失败锁定">
              <el-switch v-model="securitySettings.loginLockEnabled" />
            </el-form-item>
            <el-form-item v-if="securitySettings.loginLockEnabled" label="失败次数阈值">
              <el-input-number v-model="securitySettings.maxFailedAttempts" :min="3" :max="10" />
            </el-form-item>
            <el-form-item v-if="securitySettings.loginLockEnabled" label="锁定时间">
              <el-input-number v-model="securitySettings.lockDuration" :min="1" :max="60" />
              <span style="margin-left: 10px">分钟</span>
            </el-form-item>
            <el-form-item label="双因素认证">
              <el-switch v-model="securitySettings.twoFactorEnabled" />
            </el-form-item>
            <el-form-item label="IP白名单">
              <el-switch v-model="securitySettings.ipWhitelistEnabled" />
            </el-form-item>
            <el-form-item v-if="securitySettings.ipWhitelistEnabled" label="允许的IP">
              <el-input v-model="securitySettings.allowedIps" type="textarea" :rows="3" placeholder="每行一个IP地址或CIDR" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveSecuritySettings">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 监控配置 -->
        <el-card v-show="activeMenu === 'monitoring'">
          <template #header>
            <span>监控配置</span>
          </template>
          <el-form :model="monitoringSettings" label-width="120px">
            <el-form-item label="监控开关">
              <el-switch v-model="monitoringSettings.enabled" />
            </el-form-item>
            <el-form-item v-if="monitoringSettings.enabled" label="数据收集间隔">
              <el-input-number v-model="monitoringSettings.collectionInterval" :min="10" :max="300" />
              <span style="margin-left: 10px">秒</span>
            </el-form-item>
            <el-form-item v-if="monitoringSettings.enabled" label="数据保留天数">
              <el-input-number v-model="monitoringSettings.retentionDays" :min="7" :max="365" />
            </el-form-item>
            <el-form-item v-if="monitoringSettings.enabled" label="性能阈值">
              <div class="threshold-item">
                <span>CPU使用率: </span>
                <el-input-number v-model="monitoringSettings.cpuThreshold" :min="50" :max="100" />
                <span>%</span>
              </div>
              <div class="threshold-item">
                <span>内存使用率: </span>
                <el-input-number v-model="monitoringSettings.memoryThreshold" :min="50" :max="100" />
                <span>%</span>
              </div>
              <div class="threshold-item">
                <span>磁盘使用率: </span>
                <el-input-number v-model="monitoringSettings.diskThreshold" :min="50" :max="100" />
                <span>%</span>
              </div>
            </el-form-item>
            <el-form-item v-if="monitoringSettings.enabled" label="监控指标">
              <el-checkbox-group v-model="monitoringSettings.metrics">
                <el-checkbox label="cpu">CPU使用率</el-checkbox>
                <el-checkbox label="memory">内存使用率</el-checkbox>
                <el-checkbox label="disk">磁盘使用率</el-checkbox>
                <el-checkbox label="network">网络流量</el-checkbox>
                <el-checkbox label="qps">查询QPS</el-checkbox>
                <el-checkbox label="latency">响应延迟</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveMonitoringSettings">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 备份设置 -->
        <el-card v-show="activeMenu === 'backup'">
          <template #header>
            <span>备份设置</span>
          </template>
          <el-form :model="backupSettings" label-width="120px">
            <el-form-item label="自动备份">
              <el-switch v-model="backupSettings.autoBackup" />
            </el-form-item>
            <el-form-item v-if="backupSettings.autoBackup" label="备份周期">
              <el-select v-model="backupSettings.schedule" style="width: 100%">
                <el-option label="每天" value="daily" />
                <el-option label="每周" value="weekly" />
                <el-option label="每月" value="monthly" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="backupSettings.autoBackup" label="备份时间">
              <el-time-picker v-model="backupSettings.backupTime" format="HH:mm" />
            </el-form-item>
            <el-form-item label="备份存储位置">
              <el-input v-model="backupSettings.storagePath" />
            </el-form-item>
            <el-form-item label="备份保留数量">
              <el-input-number v-model="backupSettings.retentionCount" :min="1" :max="100" />
            </el-form-item>
            <el-form-item label="压缩备份">
              <el-switch v-model="backupSettings.compression" />
            </el-form-item>
            <el-form-item label="备份内容">
              <el-checkbox-group v-model="backupSettings.content">
                <el-checkbox label="indices">索引数据</el-checkbox>
                <el-checkbox label="config">配置文件</el-checkbox>
                <el-checkbox label="logs">日志文件</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveBackupSettings">保存设置</el-button>
              <el-button @click="createBackup">立即备份</el-button>
              <el-button @click="restoreBackup">恢复备份</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 日志配置 -->
        <el-card v-show="activeMenu === 'log'">
          <template #header>
            <span>日志配置</span>
          </template>
          <el-form :model="logSettings" label-width="120px">
            <el-form-item label="日志级别">
              <el-select v-model="logSettings.level" style="width: 100%">
                <el-option label="DEBUG" value="DEBUG" />
                <el-option label="INFO" value="INFO" />
                <el-option label="WARNING" value="WARNING" />
                <el-option label="ERROR" value="ERROR" />
                <el-option label="CRITICAL" value="CRITICAL" />
              </el-select>
            </el-form-item>
            <el-form-item label="日志文件大小">
              <el-input-number v-model="logSettings.maxFileSize" :min="1" :max="1000" />
              <span style="margin-left: 10px">MB</span>
            </el-form-item>
            <el-form-item label="日志保留天数">
              <el-input-number v-model="logSettings.retentionDays" :min="1" :max="365" />
            </el-form-item>
            <el-form-item label="日志格式">
              <el-radio-group v-model="logSettings.format">
                <el-radio label="text">文本格式</el-radio>
                <el-radio label="json">JSON格式</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="日志组件">
              <el-checkbox-group v-model="logSettings.components">
                <el-checkbox label="system">系统日志</el-checkbox>
                <el-checkbox label="access">访问日志</el-checkbox>
                <el-checkbox label="error">错误日志</el-checkbox>
                <el-checkbox label="audit">审计日志</el-checkbox>
                <el-checkbox label="performance">性能日志</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveLogSettings">保存设置</el-button>
              <el-button @click="downloadLogs">下载日志</el-button>
              <el-button @click="clearLogs">清理日志</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 系统管理 -->
        <el-card v-show="activeMenu === 'system'">
          <template #header>
            <span>系统管理</span>
          </template>
          <div class="system-actions">
            <el-row :gutter="20">
              <el-col :span="12">
                <div class="action-card">
                  <h4>系统维护</h4>
                  <div class="action-buttons">
                    <el-button type="primary" @click="restartSystem">重启系统</el-button>
                    <el-button type="warning" @click="clearCache">清理缓存</el-button>
                    <el-button type="danger" @click="resetSystem">重置系统</el-button>
                  </div>
                </div>
              </el-col>
              <el-col :span="12">
                <div class="action-card">
                  <h4>数据管理</h4>
                  <div class="action-buttons">
                    <el-button type="primary" @click="exportData">导出数据</el-button>
                    <el-button type="success" @click="importData">导入数据</el-button>
                    <el-button type="warning" @click="optimizeData">优化数据</el-button>
                  </div>
                </div>
              </el-col>
            </el-row>
            <el-row :gutter="20" style="margin-top: 20px">
              <el-col :span="12">
                <div class="action-card">
                  <h4>许可证管理</h4>
                  <div class="license-info">
                    <p>当前许可证: {{ licenseInfo.type }}</p>
                    <p>到期时间: {{ formatTime(licenseInfo.expiry) }}</p>
                    <p>授权用户: {{ licenseInfo.users }}</p>
                  </div>
                  <div class="action-buttons">
                    <el-button type="primary" @click="updateLicense">更新许可证</el-button>
                    <el-button @click="viewLicense">查看详情</el-button>
                  </div>
                </div>
              </el-col>
              <el-col :span="12">
                <div class="action-card">
                  <h4>系统信息</h4>
                  <div class="system-info">
                    <p>版本: {{ systemInfo.version }}</p>
                    <p>构建时间: {{ formatTime(systemInfo.buildTime) }}</p>
                    <p>运行时间: {{ systemInfo.uptime }}</p>
                    <p>Java版本: {{ systemInfo.javaVersion }}</p>
                  </div>
                  <div class="action-buttons">
                    <el-button type="primary" @click="checkUpdate">检查更新</el-button>
                    <el-button @click="viewSystemInfo">查看详情</el-button>
                  </div>
                </div>
              </el-col>
            </el-row>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Setting, Bell, Lock, Monitor, FolderOpened, Document, Tools } from '@element-plus/icons-vue'

// 响应式数据
const activeMenu = ref('general')

// 基本设置
const generalSettings = reactive({
  systemName: 'DevOps Console',
  systemDescription: 'Elasticsearch集群运维管理平台',
  language: 'zh-CN',
  timezone: 'Asia/Shanghai',
  refreshInterval: 30,
  theme: 'light'
})

// 通知设置
const notificationSettings = reactive({
  emailEnabled: false,
  smtpServer: '',
  smtpPort: 587,
  senderEmail: '',
  emailPassword: '',
  smsEnabled: false,
  smsApiUrl: '',
  smsApiKey: '',
  notificationLevels: ['CRITICAL', 'WARNING']
})

// 安全设置
const securitySettings = reactive({
  sessionTimeout: 120,
  passwordPolicy: ['minLength', 'uppercase', 'numbers'],
  loginLockEnabled: true,
  maxFailedAttempts: 5,
  lockDuration: 30,
  twoFactorEnabled: false,
  ipWhitelistEnabled: false,
  allowedIps: ''
})

// 监控配置
const monitoringSettings = reactive({
  enabled: true,
  collectionInterval: 60,
  retentionDays: 30,
  cpuThreshold: 80,
  memoryThreshold: 85,
  diskThreshold: 90,
  metrics: ['cpu', 'memory', 'disk', 'qps', 'latency']
})

// 备份设置
const backupSettings = reactive({
  autoBackup: true,
  schedule: 'daily',
  backupTime: new Date(2025, 10, 14, 2, 0),
  storagePath: '/backup/elfk-console',
  retentionCount: 30,
  compression: true,
  content: ['indices', 'config']
})

// 日志配置
const logSettings = reactive({
  level: 'INFO',
  maxFileSize: 100,
  retentionDays: 30,
  format: 'text',
  components: ['system', 'access', 'error']
})

// 许可证信息
const licenseInfo = reactive({
  type: '企业版',
  expiry: Date.now() + 1000 * 60 * 60 * 24 * 365,
  users: 100
})

// 系统信息
const systemInfo = reactive({
  version: 'v1.0.0',
  buildTime: Date.now() - 1000 * 60 * 60 * 24 * 30,
  uptime: '30天12小时45分钟',
  javaVersion: 'OpenJDK 11.0.16'
})

// 菜单选择处理
const handleMenuSelect = (index) => {
  activeMenu.value = index
}

// 格式化时间
const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString()
}

// 保存基本设置
const saveGeneralSettings = () => {
  ElMessage.success('基本设置保存成功')
}

// 保存通知设置
const saveNotificationSettings = () => {
  ElMessage.success('通知设置保存成功')
}

// 测试通知
const testNotification = () => {
  ElMessage.info('正在发送测试通知...')
  setTimeout(() => {
    ElMessage.success('测试通知发送成功')
  }, 2000)
}

// 保存安全设置
const saveSecuritySettings = () => {
  ElMessage.success('安全设置保存成功')
}

// 保存监控配置
const saveMonitoringSettings = () => {
  ElMessage.success('监控配置保存成功')
}

// 保存备份设置
const saveBackupSettings = () => {
  ElMessage.success('备份设置保存成功')
}

// 创建备份
const createBackup = () => {
  ElMessageBox.confirm(
    '确定要立即创建备份吗？此操作可能会影响系统性能。',
    '确认备份',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(() => {
    ElMessage.info('正在创建备份，请稍候...')
    setTimeout(() => {
      ElMessage.success('备份创建成功')
    }, 3000)
  }).catch(() => {
    ElMessage.info('已取消备份操作')
  })
}

// 恢复备份
const restoreBackup = () => {
  ElMessageBox.confirm(
    '确定要恢复备份吗？此操作将覆盖当前数据，请确保已做好数据备份。',
    '确认恢复',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'danger',
    }
  ).then(() => {
    ElMessage.info('正在恢复备份，请稍候...')
    setTimeout(() => {
      ElMessage.success('备份恢复成功')
    }, 5000)
  }).catch(() => {
    ElMessage.info('已取消恢复操作')
  })
}

// 保存日志配置
const saveLogSettings = () => {
  ElMessage.success('日志配置保存成功')
}

// 下载日志
const downloadLogs = () => {
  ElMessage.info('正在准备日志文件，请稍候...')
  setTimeout(() => {
    ElMessage.success('日志文件下载成功')
  }, 2000)
}

// 清理日志
const clearLogs = () => {
  ElMessageBox.confirm(
    '确定要清理历史日志吗？此操作不可恢复。',
    '确认清理',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'danger',
    }
  ).then(() => {
    ElMessage.success('日志清理成功')
  }).catch(() => {
    ElMessage.info('已取消清理操作')
  })
}

// 重启系统
const restartSystem = () => {
  ElMessageBox.confirm(
    '确定要重启系统吗？此操作将中断所有正在运行的服务。',
    '确认重启',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'danger',
    }
  ).then(() => {
    ElMessage.info('系统正在重启，请稍候...')
  }).catch(() => {
    ElMessage.info('已取消重启操作')
  })
}

// 清理缓存
const clearCache = () => {
  ElMessage.info('正在清理缓存，请稍候...')
  setTimeout(() => {
    ElMessage.success('缓存清理成功')
  }, 2000)
}

// 重置系统
const resetSystem = () => {
  ElMessageBox.confirm(
    '确定要重置系统吗？此操作将清除所有配置和数据，请确保已做好数据备份。',
    '确认重置',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'danger',
    }
  ).then(() => {
    ElMessage.info('系统正在重置，请稍候...')
  }).catch(() => {
    ElMessage.info('已取消重置操作')
  })
}

// 导出数据
const exportData = () => {
  ElMessage.info('正在导出数据，请稍候...')
  setTimeout(() => {
    ElMessage.success('数据导出成功')
  }, 3000)
}

// 导入数据
const importData = () => {
  ElMessage.info('请选择要导入的数据文件')
}

// 优化数据
const optimizeData = () => {
  ElMessage.info('正在优化数据，请稍候...')
  setTimeout(() => {
    ElMessage.success('数据优化成功')
  }, 5000)
}

// 更新许可证
const updateLicense = () => {
  ElMessage.info('请选择新的许可证文件')
}

// 查看许可证
const viewLicense = () => {
  ElMessage.info('查看许可证详情')
}

// 检查更新
const checkUpdate = () => {
  ElMessage.info('正在检查更新，请稍候...')
  setTimeout(() => {
    ElMessage.success('当前已是最新版本')
  }, 2000)
}

// 查看系统信息
const viewSystemInfo = () => {
  ElMessage.info('查看系统详细信息')
}

// 组件挂载时加载设置
onMounted(() => {
  // 这里可以加载实际的设置数据
})
</script>

<style scoped>
.settings {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  color: #303133;
}

.menu-card {
  height: fit-content;
}

.threshold-item {
  margin-bottom: 10px;
  display: flex;
  align-items: center;
}

.threshold-item span {
  width: 100px;
}

.system-actions {
  padding: 20px 0;
}

.action-card {
  border: 1px solid #E4E7ED;
  border-radius: 4px;
  padding: 20px;
  height: 100%;
}

.action-card h4 {
  margin-top: 0;
  margin-bottom: 15px;
  color: #303133;
}

.action-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.license-info, .system-info {
  margin-bottom: 15px;
}

.license-info p, .system-info p {
  margin: 5px 0;
  color: #606266;
}
</style>