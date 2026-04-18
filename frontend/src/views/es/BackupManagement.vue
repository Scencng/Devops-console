<template>
  <div class="backup-management">
    <div class="page-header">
      <h2>备份管理</h2>
    </div>

    <!-- 仓库列表 -->
    <el-card class="repositories-card">
      <template #header>
        <div class="card-header">
          <span>快照仓库</span>
          <div class="card-header-actions">
            <el-button type="primary" @click="showCreateRepositoryDialog">
              创建仓库
            </el-button>
            <el-button @click="refreshData">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>
      <el-table :data="repositories" v-loading="loading" stripe>
        <el-table-column prop="name" label="仓库名称" width="200" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getRepositoryTypeColor(row.type)" size="small">
              {{ getRepositoryTypeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="settings.location" label="存储位置" min-width="300" />
        <el-table-column prop="settings.compress" label="压缩" width="100">
          <template #default="{ row }">
            <el-tag :type="row.settings.compress ? 'success' : 'info'" size="small">
              {{ row.settings.compress ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="showCreateSnapshotDialog(row)">
              创建快照
            </el-button>
            <el-button type="warning" size="small" link @click="viewSnapshots(row)">
              查看快照
            </el-button>
            <el-button type="danger" size="small" link @click="deleteRepository(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 快照列表对话框 -->
    <el-dialog
      v-model="snapshotDialogVisible"
      :title="`仓库 ${currentRepository?.name} 的快照`"
      width="80%"
    >
      <el-table :data="snapshots" v-loading="snapshotLoading" stripe>
        <el-table-column prop="snapshot" label="快照名称" width="200" />
        <el-table-column prop="state" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getSnapshotStateColor(row.state)" size="small">
              {{ getSnapshotStateLabel(row.state) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="startTime" label="开始时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.start_time) }}
          </template>
        </el-table-column>
        <el-table-column prop="endTime" label="结束时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.end_time) }}
          </template>
        </el-table-column>
        <el-table-column prop="duration" label="耗时" width="120">
          <template #default="{ row }">
            {{ formatDuration(row.duration) }}
          </template>
        </el-table-column>
        <el-table-column label="备份类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getBackupTypeColor(row.metadata?.backup_type)" size="small">
              {{ getBackupTypeLabel(row.metadata?.backup_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="索引数量" width="100">
          <template #default="{ row }">
            <el-tooltip :content="getIndicesTooltip(row.indices)" placement="top">
              <span>{{ row.indices?.length || 0 }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="分片状态" width="120">
          <template #default="{ row }">
            <el-tooltip :content="`${row.shards?.successful || 0}/${row.shards?.total || 0} 成功`" placement="top">
              <el-tag :type="getShardStateColor(row.shards)" size="small">
                {{ row.shards?.successful || 0 }}/{{ row.shards?.total || 0 }}
              </el-tag>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="showRestoreDialog(row)">
              恢复
            </el-button>
            <el-button type="warning" size="small" link @click="checkSnapshotStatus(row)">
              状态
            </el-button>
            <el-button type="danger" size="small" link @click="deleteSnapshot(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 创建仓库对话框 -->
    <el-dialog
      v-model="createRepositoryDialogVisible"
      title="创建快照仓库"
      width="600px"
    >
      <el-form :model="repositoryForm" :rules="repositoryRules" ref="repositoryFormRef" label-width="120px">
        <el-form-item label="仓库名称" prop="name">
          <el-input v-model="repositoryForm.name" placeholder="输入仓库名称" />
        </el-form-item>
        <el-form-item label="仓库类型" prop="type">
          <el-select v-model="repositoryForm.type" placeholder="选择仓库类型" style="width: 100%">
            <el-option label="文件系统" value="fs" />
            <el-option label="S3" value="s3" />
            <el-option label="HDFS" value="hdfs" />
            <el-option label="Azure" value="azure" />
            <el-option label="GCS" value="gcs" />
          </el-select>
        </el-form-item>
        <el-form-item label="存储位置" prop="location">
          <el-input 
            v-model="repositoryForm.location" 
            placeholder="如: /var/lib/elasticsearch/snapshots/backup"
          >
            <template #prepend>
              <el-select v-model="repositoryForm.location" placeholder="选择常用路径" style="width: 200px">
                <el-option label="/var/lib/elasticsearch/snapshots" value="/var/lib/elasticsearch/snapshots" />
                <el-option label="/tmp/es-data" value="/tmp/es-data" />
                <el-option label="/data/elasticsearch/backup" value="/data/elasticsearch/backup" />
                <el-option label="/mnt/elastic-backup" value="/mnt/elastic-backup" />
              </el-select>
            </template>
          </el-input>
          <div style="margin-top: 5px; font-size: 12px; color: #909399">
            注意：路径必须在Elasticsearch的path.repo配置中指定
          </div>
        </el-form-item>
        <el-form-item label="压缩">
          <el-switch v-model="repositoryForm.compress" />
        </el-form-item>
        <el-form-item label="最大快照数">
          <el-input-number v-model="repositoryForm.maxSnapshotCount" :min="1" />
        </el-form-item>
        <el-form-item label="验证仓库">
          <el-switch v-model="repositoryForm.verify" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createRepositoryDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createRepository">创建</el-button>
      </template>
    </el-dialog>

    <!-- 创建快照对话框 -->
    <el-dialog
      v-model="createSnapshotDialogVisible"
      title="创建快照"
      width="600px"
    >
      <el-form :model="snapshotForm" :rules="snapshotRules" ref="snapshotFormRef" label-width="120px">
        <el-form-item label="快照名称" prop="snapshot">
          <el-input v-model="snapshotForm.snapshot" placeholder="输入快照名称" />
        </el-form-item>
        <el-form-item label="备份类型">
          <el-radio-group v-model="snapshotForm.backupType" @change="handleBackupTypeChange">
            <el-radio label="all">全量备份</el-radio>
            <el-radio label="custom">自定义索引</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="索引" v-if="snapshotForm.backupType === 'custom'">
          <el-select 
            v-model="snapshotForm.indices" 
            multiple 
            placeholder="选择要备份的索引" 
            style="width: 100%"
            filterable
          >
            <el-option
              v-for="index in indexOptions"
              :key="index"
              :label="index"
              :value="index"
            />
          </el-select>
          <div style="margin-top: 5px; font-size: 12px; color: #909399">
            选择需要备份的索引，不选择则备份所有索引
          </div>
        </el-form-item>
        <el-form-item label="备份范围" v-if="snapshotForm.backupType === 'all'">
          <el-tag type="info">将备份所有索引和全局状态</el-tag>
        </el-form-item>
        <el-form-item label="忽略不可用索引">
          <el-switch v-model="snapshotForm.ignoreUnavailable" />
        </el-form-item>
        <el-form-item label="包含全局状态">
          <el-switch v-model="snapshotForm.includeGlobalState" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="snapshotForm.description" type="textarea" placeholder="输入快照描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createSnapshotDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createSnapshot">创建</el-button>
      </template>
    </el-dialog>

    <!-- 快照状态详情对话框 -->
    <el-dialog
      v-model="snapshotStatusDialogVisible"
      title="快照状态详情"
      width="80%"
    >
      <div v-if="snapshotStatusData">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="快照名称">{{ snapshotStatusData.snapshot }}</el-descriptions-item>
          <el-descriptions-item label="仓库">{{ snapshotStatusData.repository }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getSnapshotStateColor(snapshotStatusData.state)" size="small">
              {{ getSnapshotStateLabel(snapshotStatusData.state) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="UUID">{{ snapshotStatusData.uuid }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">{{ formatDateTime(snapshotStatusData.stats?.start_time_in_millis) }}</el-descriptions-item>
          <el-descriptions-item label="耗时">{{ formatDuration(snapshotStatusData.stats?.time_in_millis) }}</el-descriptions-item>
          <el-descriptions-item label="文件数量">{{ snapshotStatusData.stats?.total?.file_count || 0 }}</el-descriptions-item>
          <el-descriptions-item label="总大小">{{ formatBytes(snapshotStatusData.stats?.total?.size_in_bytes || 0) }}</el-descriptions-item>
          <el-descriptions-item label="增量文件">{{ snapshotStatusData.stats?.incremental?.file_count || 0 }}</el-descriptions-item>
          <el-descriptions-item label="增量大小">{{ formatBytes(snapshotStatusData.stats?.incremental?.size_in_bytes || 0) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider />

        <h4>分片统计</h4>
        <el-row :gutter="20">
          <el-col :span="6">
            <el-statistic title="总分片数" :value="snapshotStatusData.shards_stats?.total || 0" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="已完成" :value="snapshotStatusData.shards_stats?.done || 0" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="失败" :value="snapshotStatusData.shards_stats?.failed || 0" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="进行中" :value="snapshotStatusData.shards_stats?.started || 0" />
          </el-col>
        </el-row>

        <el-divider />

        <h4>索引详情</h4>
        <el-table :data="snapshotIndicesData" max-height="400" stripe>
          <el-table-column prop="name" label="索引名称" min-width="200" />
          <el-table-column label="分片状态" width="120">
            <template #default="{ row }">
              <el-tag :type="getIndexShardStateColor(row.shards_stats)" size="small">
                {{ row.shards_stats?.done || 0 }}/{{ row.shards_stats?.total || 0 }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="文件数量" width="100">
            <template #default="{ row }">
              {{ row.stats?.total?.file_count || 0 }}
            </template>
          </el-table-column>
          <el-table-column label="大小" width="120">
            <template #default="{ row }">
              {{ formatBytes(row.stats?.total?.size_in_bytes || 0) }}
            </template>
          </el-table-column>
          <el-table-column label="耗时" width="100">
            <template #default="{ row }">
              {{ formatDuration(row.stats?.time_in_millis) }}
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>

    <!-- 恢复快照对话框 -->
    <el-dialog
      v-model="restoreDialogVisible"
      title="恢复快照"
      width="700px"
    >
      <el-alert
        title="重要提示"
        type="warning"
        :closable="false"
        style="margin-bottom: 20px"
      >
        <template #default>
          <p>恢复快照时可能会遇到以下常见问题：</p>
          <ul style="margin: 8px 0 0 20px; font-size: 13px; line-height: 1.5;">
            <li><strong>索引已存在错误：</strong>如目标索引已存在且处于打开状态，恢复将失败</li>
            <li><strong>解决方案1：</strong>使用"重命名模式"和"重命名替换"自动重命名恢复的索引</li>
            <li><strong>解决方案2：</strong>先手动关闭或删除同名索引，再进行恢复</li>
            <li><strong>解决方案3：</strong>勾选"忽略不可用索引"跳过冲突的索引</li>
          </ul>
        </template>
      </el-alert>
      
      <el-form :model="restoreForm" :rules="restoreRules" ref="restoreFormRef" label-width="120px">
        <el-form-item label="目标索引">
          <el-select v-model="restoreForm.indices" multiple placeholder="选择要恢复的索引" style="width: 100%">
            <el-option
              v-for="index in currentSnapshot?.indices || []"
              :key="index"
              :label="index"
              :value="index"
            />
          </el-select>
          <div style="margin-top: 5px; font-size: 12px; color: #909399">
            选择要恢复的索引，如果索引已存在且未关闭，恢复将失败
          </div>
        </el-form-item>
        <el-form-item label="忽略不可用索引">
          <el-switch v-model="restoreForm.ignoreUnavailable" />
          <div style="margin-top: 5px; font-size: 12px; color: #909399">
            启用后将跳过不可用的索引，避免因部分索引问题导致整个恢复失败
          </div>
        </el-form-item>
        <el-form-item label="包含全局状态">
          <el-switch v-model="restoreForm.includeGlobalState" />
          <div style="margin-top: 5px; font-size: 12px; color: #909399">
            恢复集群全局状态（如模板、索引生命周期策略等）
          </div>
        </el-form-item>
        <el-form-item label="重命名模式">
          <div style="display: flex; gap: 10px; align-items: center;">
            <el-input v-model="restoreForm.renamePattern" placeholder="如: (.+)" style="flex: 1" />
            <el-button size="small" @click="setRenamePreset('prefix')">前缀预设</el-button>
            <el-button size="small" @click="setRenamePreset('suffix')">后缀预设</el-button>
            <el-button size="small" @click="clearRename()">清空</el-button>
          </div>
          <div style="margin-top: 5px; font-size: 12px; color: #909399">
            使用正则表达式匹配要重命名的索引，避免与现有索引冲突
          </div>
        </el-form-item>
        <el-form-item label="重命名替换">
          <el-input v-model="restoreForm.renameReplacement" placeholder="如: restored-$1" />
          <div style="margin-top: 5px; font-size: 12px; color: #909399">
            替换模式，$1 表示第一个捕获组的内容
          </div>
        </el-form-item>
        <el-form-item>
          <el-alert
            title="示例"
            type="info"
            :closable="false"
            style="margin-top: 10px"
          >
            <template #default>
              <p style="font-size: 13px; line-height: 1.5;">
                如果要恢复索引 <code>my-index-001</code> 到 <code>restored-my-index-001</code>：<br>
                重命名模式：<code>(.+)</code><br>
                重命名替换：<code>restored-$1</code>
              </p>
            </template>
          </el-alert>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="restoreDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="restoreSnapshot">恢复</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import {
  getRepositories,
  createRepository as apiCreateRepository,
  deleteRepository as apiDeleteRepository,
  getSnapshots,
  createSnapshot as apiCreateSnapshot,
  deleteSnapshot as apiDeleteSnapshot,
  getSnapshotStatus,
  restoreSnapshot as apiRestoreSnapshot,
  getRestoreStatus
} from '@/api/es/backup.js'
import { getInstanceList } from '@/api/instance.js'
import { getSelectedInstanceId } from '@/stores/instanceStore.js'

// 响应式数据
const loading = ref(false)
const snapshotLoading = ref(false)
const repositories = ref([])
const snapshots = ref([])
const instanceOptions = ref([])
const indexOptions = ref([])
const snapshotStatusData = ref(null)
const snapshotIndicesData = ref([])

// 对话框状态
const snapshotDialogVisible = ref(false)
const createRepositoryDialogVisible = ref(false)
const createSnapshotDialogVisible = ref(false)
const restoreDialogVisible = ref(false)
const snapshotStatusDialogVisible = ref(false)

// 当前操作对象
const currentRepository = ref(null)
const currentSnapshot = ref(null)

// 表单引用
const repositoryFormRef = ref(null)
const snapshotFormRef = ref(null)
const restoreFormRef = ref(null)

// 仓库表单
const repositoryForm = reactive({
  name: '',
  type: 'fs',
  location: '',
  compress: true,
  maxSnapshotCount: 10,
  verify: true
})

// 快照表单
const snapshotForm = reactive({
  snapshot: '',
  backupType: 'all',
  indices: [],
  ignoreUnavailable: true,
  includeGlobalState: true,
  description: ''
})

// 恢复表单
const restoreForm = reactive({
  indices: [],
  ignoreUnavailable: true,
  includeGlobalState: false,
  renamePattern: '',
  renameReplacement: ''
})

// 表单验证规则
const repositoryRules = {
  name: [{ required: true, message: '请输入仓库名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择仓库类型', trigger: 'change' }],
  location: [{ required: true, message: '请输入存储位置', trigger: 'blur' }]
}

const snapshotRules = {
  snapshot: [{ required: true, message: '请输入快照名称', trigger: 'blur' }]
}

const restoreRules = {
  indices: [{ required: true, message: '请选择要恢复的索引', trigger: 'change' }]
}

// 获取仓库类型颜色
const getRepositoryTypeColor = (type) => {
  const colorMap = {
    'fs': 'primary',
    's3': 'success',
    'hdfs': 'warning',
    'azure': 'info',
    'gcs': 'success'
  }
  return colorMap[type] || 'info'
}

// 获取仓库类型标签
const getRepositoryTypeLabel = (type) => {
  const labelMap = {
    'fs': '文件系统',
    's3': 'S3',
    'hdfs': 'HDFS',
    'azure': 'Azure',
    'gcs': 'GCS'
  }
  return labelMap[type] || type
}

// 获取快照状态颜色
const getSnapshotStateColor = (state) => {
  const colorMap = {
    'SUCCESS': 'success',
    'IN_PROGRESS': 'warning',
    'FAILED': 'danger',
    'PARTIAL': 'warning'
  }
  return colorMap[state] || 'info'
}

// 获取快照状态标签
const getSnapshotStateLabel = (state) => {
  const labelMap = {
    'SUCCESS': '成功',
    'IN_PROGRESS': '进行中',
    'FAILED': '失败',
    'PARTIAL': '部分'
  }
  return labelMap[state] || state
}

// 获取备份类型颜色
const getBackupTypeColor = (type) => {
  const colorMap = {
    'all': 'primary',
    'custom': 'success'
  }
  return colorMap[type] || 'info'
}

// 获取备份类型标签
const getBackupTypeLabel = (type) => {
  const labelMap = {
    'all': '全量备份',
    'custom': '自定义'
  }
  return labelMap[type] || '未知'
}

// 获取分片状态颜色
const getShardStateColor = (shards) => {
  if (!shards) return 'info'
  if (shards.failed > 0) return 'danger'
  if (shards.successful === shards.total) return 'success'
  return 'warning'
}

// 获取索引工具提示
const getIndicesTooltip = (indices) => {
  if (!indices || indices.length === 0) return '无索引'
  
  // 如果索引数量太多，只显示前10个
  const displayIndices = indices.slice(0, 10)
  const tooltip = displayIndices.join(', ')
  
  if (indices.length > 10) {
    return `${tooltip} ... 等${indices.length}个索引`
  }
  
  return tooltip
}

// 格式化字节数
const formatBytes = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 获取索引分片状态颜色
const getIndexShardStateColor = (shardsStats) => {
  if (!shardsStats) return 'info'
  if (shardsStats.failed > 0) return 'danger'
  if (shardsStats.done === shardsStats.total) return 'success'
  return 'warning'
}

// 格式化日期时间
const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

// 格式化持续时间
const formatDuration = (duration) => {
  if (!duration) return '-'
  if (typeof duration === 'string') {
    return duration
  }
  const seconds = Math.floor(duration / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  
  if (hours > 0) {
    return `${hours}h ${minutes % 60}m ${seconds % 60}s`
  } else if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`
  } else {
    return `${seconds}s`
  }
}

// 处理实例变更
const handleInstanceChange = () => {
  loadRepositories()
}

// 处理备份类型变更
const handleBackupTypeChange = (type) => {
  if (type === 'all') {
    snapshotForm.indices = []
    snapshotForm.includeGlobalState = true
  }
}

// 显示创建仓库对话框
const showCreateRepositoryDialog = () => {
  Object.assign(repositoryForm, {
    name: '',
    type: 'fs',
    location: '',
    compress: true,
    maxSnapshotCount: 10,
    verify: true
  })
  createRepositoryDialogVisible.value = true
}

// 创建仓库
const createRepository = async () => {
  if (!repositoryFormRef.value) return
  
  try {
    await repositoryFormRef.value.validate()
    
    const selectedInstanceId = getSelectedInstanceId()
    if (!selectedInstanceId) {
      ElMessage.error('请先选择一个实例')
      return
    }
    
    const data = {
      instance_id: selectedInstanceId,
      type: repositoryForm.type,
      settings: {
        location: repositoryForm.location,
        compress: repositoryForm.compress,
        max_snapshot_count: repositoryForm.maxSnapshotCount
      },
      verify: repositoryForm.verify
    }
    
    await apiCreateRepository(data)
    ElMessage.success('仓库创建成功')
    createRepositoryDialogVisible.value = false
    loadRepositories()
  } catch (error) {
    let errorMessage = '创建失败：' + error.message
    
    // 检查是否是path.repo配置错误
    if (error.message && error.message.includes('path.repo')) {
      errorMessage = `创建失败：Elasticsearch未配置仓库路径。
      
解决方案：
1. 在Elasticsearch的elasticsearch.yml中添加：path.repo: ["/var/lib/elasticsearch/snapshots","其他路径"]
2. 重启Elasticsearch服务
3. 确保指定的路径存在且有写入权限

当前尝试的路径：${repositoryForm.location}`
    }
    
    ElMessage.error(errorMessage)
  }
}

// 删除仓库
const deleteRepository = async (repository) => {
  try {
    await ElMessageBox.confirm(`确定要删除仓库 "${repository.name}" 吗？`, '确认删除', {
      type: 'warning'
    })
    
    const selectedInstanceId = getSelectedInstanceId()
    if (!selectedInstanceId) {
      ElMessage.error('请先选择一个实例')
      return
    }
    
    await apiDeleteRepository({
      instance_id: selectedInstanceId,
      repository: repository.name
    })
    
    ElMessage.success('仓库删除成功')
    loadRepositories()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败：' + error.message)
    }
  }
}

// 查看快照
const viewSnapshots = async (repository) => {
  currentRepository.value = repository
  snapshotDialogVisible.value = true
  await loadSnapshots(repository.name)
}

// 加载快照
const loadSnapshots = async (repository) => {
  snapshotLoading.value = true
  try {
    const selectedInstanceId = getSelectedInstanceId()
    if (!selectedInstanceId) {
      ElMessage.error('请先选择一个实例')
      return
    }
    
    const response = await getSnapshots(selectedInstanceId, repository)
    console.log('快照API响应数据:', response.data)
    snapshots.value = response.data.data?.snapshots || response.data.snapshots || []
  } catch (error) {
    console.log('加载快照错误:', error)
    // 不显示错误消息，因为可能是仓库为空
    snapshots.value = []
  } finally {
    snapshotLoading.value = false
  }
}

// 显示创建快照对话框
const showCreateSnapshotDialog = (repository) => {
  currentRepository.value = repository
  Object.assign(snapshotForm, {
    snapshot: `snapshot_${Date.now()}`,
    backupType: 'all',
    indices: [],
    ignoreUnavailable: true,
    includeGlobalState: true,
    description: ''
  })
  createSnapshotDialogVisible.value = true
}

// 创建快照
const createSnapshot = async () => {
  if (!snapshotFormRef.value) return
  
  try {
    await snapshotFormRef.value.validate()
    
    // 根据备份类型设置索引
    let indices = snapshotForm.indices
    if (snapshotForm.backupType === 'all') {
      indices = [] // 空数组表示所有索引
    }
    
    const selectedInstanceId = getSelectedInstanceId()
    if (!selectedInstanceId) {
      ElMessage.error('请先选择一个实例')
      return
    }
    
    const data = {
      instance_id: selectedInstanceId,
      repository: currentRepository.value.name,
      snapshot: snapshotForm.snapshot,
      indices: indices,
      ignore_unavailable: snapshotForm.ignoreUnavailable,
      include_global_state: snapshotForm.includeGlobalState,
      metadata: {
        description: snapshotForm.description,
        backup_type: snapshotForm.backupType,
        created_by: 'devops-console'
      }
    }
    
    await apiCreateSnapshot(data)
    ElMessage.success('快照创建任务已提交')
    createSnapshotDialogVisible.value = false
    if (snapshotDialogVisible.value) {
      loadSnapshots(currentRepository.value.name)
    }
  } catch (error) {
    if (error.message) {
      ElMessage.error('创建失败：' + error.message)
    }
  }
}

// 删除快照
const deleteSnapshot = async (snapshot) => {
  try {
    await ElMessageBox.confirm(`确定要删除快照 "${snapshot.snapshot}" 吗？`, '确认删除', {
      type: 'warning'
    })
    
    const selectedInstanceId = getSelectedInstanceId()
    if (!selectedInstanceId) {
      ElMessage.error('请先选择一个实例')
      return
    }
    
    await apiDeleteSnapshot({
      instance_id: selectedInstanceId,
      repository: currentRepository.value.name,
      snapshot: snapshot.snapshot
    })
    
    ElMessage.success('快照删除成功')
    loadSnapshots(currentRepository.value.name)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败：' + error.message)
    }
  }
}

// 检查快照状态
const checkSnapshotStatus = async (snapshot) => {
  try {
    const selectedInstanceId = getSelectedInstanceId()
    if (!selectedInstanceId) {
      ElMessage.error('请先选择一个实例')
      return
    }
    
    const response = await getSnapshotStatus(
      selectedInstanceId,
      currentRepository.value.name,
      snapshot.snapshot
    )
    
    console.log('快照状态完整响应:', JSON.stringify(response, null, 2))
    console.log('响应数据结构:', response.data)
    
    // 直接根据您提供的数据结构来解析
    let statusData
    if (response.data.data?.status?.snapshots && response.data.data.status.snapshots.length > 0) {
      statusData = response.data.data.status.snapshots[0]
      console.log('找到状态数据:', statusData)
    } else {
      console.error('数据结构不匹配，实际数据:', response.data)
      // 先显示基本信息，不解析复杂结构
      ElMessage.info(`快照状态：${snapshot.state || '未知'}`)
      return
    }
    
    snapshotStatusData.value = statusData
    
    // 处理索引数据
    const indices = []
    if (statusData.indices && typeof statusData.indices === 'object') {
      for (const [indexName, indexData] of Object.entries(statusData.indices)) {
        indices.push({
          name: indexName,
          ...indexData
        })
      }
    }
    snapshotIndicesData.value = indices
    
    snapshotStatusDialogVisible.value = true
  } catch (error) {
    console.error('快照状态错误:', error)
    ElMessage.error('获取状态失败：' + error.message)
  }
}

// 显示恢复对话框
const showRestoreDialog = (snapshot) => {
  currentSnapshot.value = snapshot
  Object.assign(restoreForm, {
    indices: snapshot.indices,
    ignoreUnavailable: true,
    includeGlobalState: false,
    renamePattern: '',
    renameReplacement: ''
  })
  restoreDialogVisible.value = true
}

// 恢复快照
const restoreSnapshot = async () => {
  if (!restoreFormRef.value) return
  
  try {
    await restoreFormRef.value.validate()
    
    const selectedInstanceId = getSelectedInstanceId()
    if (!selectedInstanceId) {
      ElMessage.error('请先选择一个实例')
      return
    }
    
    const data = {
      instance_id: selectedInstanceId,
      repository: currentRepository.value.name,
      snapshot: currentSnapshot.value.snapshot,
      indices: restoreForm.indices,
      ignore_unavailable: restoreForm.ignoreUnavailable,
      include_global_state: restoreForm.includeGlobalState,
      rename_pattern: restoreForm.renamePattern,
      rename_replacement: restoreForm.renameReplacement
    }
    
    console.log('恢复快照请求数据:', data)
    await apiRestoreSnapshot(data)
    ElMessage.success('快照恢复任务已提交')
    restoreDialogVisible.value = false
  } catch (error) {
    console.error('恢复快照错误:', error)
    let errorMessage = '恢复失败：' + error.message
    
    // 检查常见错误并提供详细解决方案
    if (error.message) {
      // 索引已存在错误
      if (error.message.includes('already exists') || error.message.includes('exists and is open')) {
        errorMessage = `恢复失败：目标索引已存在且处于打开状态
        
解决方案：
1. 使用重命名功能：在"重命名模式"中输入 (.+)，在"重命名替换"中输入 restored-$1
2. 手动关闭索引：先关闭同名索引，再进行恢复
3. 删除现有索引：删除同名索引后恢复（注意：这将丢失现有数据）
4. 选择不同索引：取消勾选冲突的索引，只恢复其他索引`
      }
      // 500内部服务器错误
      else if (error.message.includes('500')) {
        errorMessage = `恢复失败：服务器内部错误
        
可能的原因：
1. 索引已存在且未关闭
2. 快照数据损坏或不完整
3. 磁盘空间不足
4. 权限问题
5. 分片分配失败

建议解决方案：
1. 检查目标索引状态，关闭或删除冲突索引
2. 确认磁盘空间充足
3. 检查Elasticsearch日志获取详细错误信息
4. 尝试恢复部分索引而非全部`
      }
      // 仓库不存在错误
      else if (error.message.includes('repository') && error.message.includes('not found')) {
        errorMessage = `恢复失败：仓库不存在
        
解决方案：
1. 检查仓库名称是否正确
2. 确认仓库是否已创建并验证成功
3. 刷新仓库列表后重试`
      }
      // 快照不存在错误
      else if (error.message.includes('snapshot') && error.message.includes('not found')) {
        errorMessage = `恢复失败：快照不存在
        
解决方案：
1. 检查快照名称是否正确
2. 确认快照是否已完成创建
3. 刷新快照列表后重试`
      }
    }
    
    ElMessage.error(errorMessage)
  }
}

// 加载仓库列表
const loadRepositories = async () => {
  const selectedInstanceId = getSelectedInstanceId()
  if (!selectedInstanceId) return
  
  loading.value = true
  try {
    const response = await getRepositories(selectedInstanceId)
    repositories.value = response.data.data?.repositories || response.data.repositories || []
  } catch (error) {
    ElMessage.error('加载仓库失败：' + error.message)
    repositories.value = []
  } finally {
    loading.value = false
  }
}

// 加载实例列表
const loadInstances = async () => {
  try {
    const response = await getInstanceList({ page: 1, page_size: 100 })
    instanceOptions.value = response.data.list.data.map(instance => ({
      label: instance.name,
      value: instance.id
    }))
    
    // Remove this since we're now using global selected instance
    // The component should rely on the global state instead of setting its own
    if (getSelectedInstanceId()) {
      loadRepositories()
    }
  } catch (error) {
    ElMessage.error('加载实例失败：' + error.message)
  }
}

// 设置重命名预设
const setRenamePreset = (type) => {
  if (type === 'prefix') {
    restoreForm.renamePattern = '(.+)'
    restoreForm.renameReplacement = 'restored-$1'
  } else if (type === 'suffix') {
    restoreForm.renamePattern = '(.+)'
    restoreForm.renameReplacement = '$1-restored'
  }
}

// 清空重命名设置
const clearRename = () => {
  restoreForm.renamePattern = ''
  restoreForm.renameReplacement = ''
}

// 刷新数据
const refreshData = () => {
  loadRepositories()
  if (snapshotDialogVisible.value && currentRepository.value) {
    loadSnapshots(currentRepository.value.name)
  }
}

// 组件挂载
onMounted(() => {
  loadInstances()
})
</script>

<style scoped>
.backup-management {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  color: #303133;
}

.header-actions {
  display: flex;
  align-items: center;
}

.repositories-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header-actions {
  display: flex;
  gap: 10px;
}

.el-table {
  margin-top: 10px;
}
</style>