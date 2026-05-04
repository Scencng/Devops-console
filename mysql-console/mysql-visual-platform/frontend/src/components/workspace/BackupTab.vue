<template>
  <section class="glass-subpanel backup-panel">
    <header class="backup-toolbar">
      <div class="backup-toolbar__meta">
        <span class="backup-toolbar__eyebrow">{{ t('common.backup') }}</span>
        <h3>{{ databaseName }}</h3>
        <p>{{ t('backup.subtitle', { database: databaseName }) }}</p>
      </div>

      <div class="backup-toolbar__actions">
        <el-button class="soft-button" :loading="creatingBackup" @click="createDatabaseBackup">
          {{ t('backup.createDatabase') }}
        </el-button>
        <el-button class="soft-button" :loading="creatingBackup" @click="promptCreateTableBackup">
          {{ t('backup.createTable') }}
        </el-button>
        <el-button class="soft-button" @click="scheduleDialog.visible = true">
          {{ t('backup.schedule') }}
        </el-button>
        <el-button class="soft-button" @click="loadAll">
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </header>

    <div v-if="activeTask" class="backup-task glass-subpanel">
      <div class="backup-task__head">
        <strong>{{ activeTaskLabel }}</strong>
        <span>{{ activeTask.progress }}%</span>
      </div>
      <el-progress :percentage="activeTask.progress" :status="activeTask.status === 'failed' ? 'exception' : activeTask.status === 'success' ? 'success' : undefined" />
      <p>{{ activeTask.message || t('backup.taskRunning') }}</p>
    </div>

    <section class="backup-schedule glass-subpanel">
      <div class="backup-section__header">
        <h4>{{ t('backup.scheduleList') }}</h4>
      </div>

      <el-empty
        v-if="schedules.length === 0"
        :description="t('backup.noSchedules')"
      />

      <div v-else class="schedule-list">
        <div v-for="schedule in schedules" :key="schedule.id" class="schedule-item">
          <div class="schedule-item__meta">
            <strong>{{ schedule.tableName || t('backup.fullDatabase') }}</strong>
            <span>{{ t('backup.everyMinutes', { minutes: schedule.intervalMinutes }) }}</span>
            <small>{{ t('backup.nextRunAt', { time: formatDateTime(schedule.nextRunAt) }) }}</small>
          </div>
          <el-button type="danger" link @click="deleteSchedule(schedule.id)">
            {{ t('tableData.delete') }}
          </el-button>
        </div>
      </div>
    </section>

    <section class="backup-list glass-subpanel">
      <div class="backup-section__header">
        <h4>{{ t('backup.records') }}</h4>
      </div>

      <el-empty
        v-if="!loading && records.length === 0"
        :description="t('backup.empty')"
      />

      <el-table
        v-else
        v-loading="loading"
        :data="records"
        row-key="fileName"
        class="backup-table"
      >
        <el-table-column prop="displayName" :label="t('backup.fileName')" min-width="220" />
        <el-table-column :label="t('backup.scope')" width="120">
          <template #default="{ row }">
            {{ row.scope === 'table' ? t('common.table') : t('common.database') }}
          </template>
        </el-table-column>
        <el-table-column prop="tableName" :label="t('common.table')" min-width="140" />
        <el-table-column :label="t('backup.size')" width="120">
          <template #default="{ row }">
            {{ formatBytes(row.size) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('backup.updatedAt')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.updatedAt) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('tableData.actions')" fixed="right" width="300">
          <template #default="{ row }">
            <div class="backup-actions">
              <el-button link @click="openRestoreDialog(row)">{{ t('backup.restore') }}</el-button>
              <el-button link @click="downloadRecord(row)">{{ t('backup.download') }}</el-button>
              <el-button link @click="openRenameDialog(row)">{{ t('workspace.renameObjectTitle') }}</el-button>
              <el-button type="danger" link @click="deleteRecord(row)">{{ t('tableData.delete') }}</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </section>

    <el-dialog v-model="restoreDialog.visible" :title="t('backup.restore')" width="460px" class="workspace-dialog backup-dialog">
      <el-form label-position="top">
        <el-form-item :label="t('backup.targetDatabase')">
          <el-input v-model="restoreDialog.targetDatabase" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="restoreDialog.visible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="restoreDialog.loading" @click="submitRestore">
          {{ t('backup.confirmRestore') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="renameDialog.visible" :title="t('workspace.renameObjectTitle')" width="420px" class="workspace-dialog backup-dialog">
      <el-form label-position="top">
        <el-form-item :label="t('backup.fileName')">
          <el-input v-model="renameDialog.name" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="renameDialog.visible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="renameDialog.loading" @click="submitRename">
          {{ t('common.create') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="scheduleDialog.visible" :title="t('backup.schedule')" width="460px" class="workspace-dialog backup-dialog">
      <el-form label-position="top">
        <el-form-item :label="t('common.table')">
          <el-input v-model="scheduleDialog.tableName" :placeholder="t('backup.fullDatabase')" />
        </el-form-item>
        <el-form-item :label="t('backup.intervalMinutes')">
          <el-input-number v-model="scheduleDialog.intervalMinutes" :min="1" :max="1440" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="scheduleDialog.visible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="scheduleDialog.loading" @click="submitSchedule">
          {{ t('common.create') }}
        </el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { getRuntimeConfig } from '@/config/runtime'
import { useConnectionStore } from '@/stores/connection'
import { useI18n } from '@/utils/i18n'
import request from '@/utils/request'

interface BackupRecord {
  id: string
  database: string
  tableName?: string
  scope: 'database' | 'table'
  fileName: string
  displayName: string
  size: number
  compressed: boolean
  createdAt: string
  updatedAt: string
}

interface BackupTask {
  id: string
  type: 'backup' | 'restore'
  database: string
  fileName?: string
  status: 'pending' | 'running' | 'success' | 'failed'
  progress: number
  message?: string
}

interface BackupSchedule {
  id: string
  database: string
  tableName?: string
  intervalMinutes: number
  nextRunAt: string
}

const props = defineProps<{
  databaseName: string
  selectedBackupName?: string
}>()

const emit = defineEmits<{
  (event: 'refresh-explorer'): void
}>()

const { t } = useI18n()
const connectionStore = useConnectionStore()

const loading = ref(false)
const creatingBackup = ref(false)
const records = ref<BackupRecord[]>([])
const schedules = ref<BackupSchedule[]>([])
const activeTask = ref<BackupTask | null>(null)
let taskTimer: number | null = null

const restoreDialog = reactive({
  visible: false,
  loading: false,
  fileName: '',
  targetDatabase: props.databaseName
})

const renameDialog = reactive({
  visible: false,
  loading: false,
  fileName: '',
  name: ''
})

const scheduleDialog = reactive({
  visible: false,
  loading: false,
  tableName: '',
  intervalMinutes: 60
})

const activeTaskLabel = computed(() => {
  if (!activeTask.value) {
    return ''
  }

  return activeTask.value.type === 'restore' ? t('backup.restore') : t('common.backup')
})

async function loadAll() {
  loading.value = true
  try {
    const [backupResponse, scheduleResponse] = await Promise.all([
      request.get<{ records: BackupRecord[] }>('/api/backup/list', {
        params: { database: props.databaseName }
      }),
      request.get<BackupSchedule[]>('/api/backup/schedules', {
        params: { database: props.databaseName }
      })
    ])

    records.value = backupResponse.records ?? []
    schedules.value = scheduleResponse ?? []
  } finally {
    loading.value = false
  }
}

async function createDatabaseBackup() {
  creatingBackup.value = true
  try {
    const response = await request.post<{ taskId: string }>('/api/backup/create', {
      database: props.databaseName,
      compress: true
    })
    await monitorTask(response.taskId)
  } finally {
    creatingBackup.value = false
  }
}

async function promptCreateTableBackup() {
  const { value } = await ElMessageBox.prompt(t('backup.tablePrompt'), t('backup.createTable'), {
    inputValue: props.selectedBackupName || ''
  })

  const tableName = value.trim()
  if (!tableName) {
    return
  }

  creatingBackup.value = true
  try {
    const response = await request.post<{ taskId: string }>('/api/backup/create', {
      database: props.databaseName,
      tableName,
      compress: true
    })
    await monitorTask(response.taskId)
  } finally {
    creatingBackup.value = false
  }
}

function openRestoreDialog(record: BackupRecord) {
  restoreDialog.visible = true
  restoreDialog.fileName = record.fileName
  restoreDialog.targetDatabase = props.databaseName
}

async function submitRestore() {
  restoreDialog.loading = true
  try {
    const response = await request.post<{ taskId: string }>('/api/backup/restore', {
      database: props.databaseName,
      fileName: restoreDialog.fileName,
      targetDatabase: restoreDialog.targetDatabase.trim()
    })
    restoreDialog.visible = false
    await monitorTask(response.taskId)
  } finally {
    restoreDialog.loading = false
  }
}

function openRenameDialog(record: BackupRecord) {
  renameDialog.visible = true
  renameDialog.fileName = record.fileName
  renameDialog.name = record.displayName
}

async function submitRename() {
  renameDialog.loading = true
  try {
    await request.post('/api/backup/rename', {
      database: props.databaseName,
      fileName: renameDialog.fileName,
      newName: renameDialog.name
    })
    renameDialog.visible = false
    await loadAll()
    emit('refresh-explorer')
    ElMessage.success(t('backup.renameSuccess'))
  } finally {
    renameDialog.loading = false
  }
}

async function deleteRecord(record: BackupRecord) {
  await ElMessageBox.confirm(t('backup.deleteConfirm', { name: record.displayName }), t('tableData.delete'))
  await request.post('/api/backup/delete', {
    database: props.databaseName,
    fileName: record.fileName
  })
  await loadAll()
  emit('refresh-explorer')
  ElMessage.success(t('workspace.objectDeleted'))
}

async function downloadRecord(record: BackupRecord) {
  try {
    const response = await fetch(
      `${getRuntimeConfig().apiBase}/api/backup/download?database=${encodeURIComponent(props.databaseName)}&fileName=${encodeURIComponent(record.fileName)}`,
      {
        headers: {
          'X-Connection-Token': connectionStore.token
        }
      }
    )

    if (!response.ok) {
      throw new Error(await response.text())
    }

    const blob = await response.blob()
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = record.fileName
    document.body.appendChild(anchor)
    anchor.click()
    anchor.remove()
    URL.revokeObjectURL(url)
  } catch (error) {
    const message = error instanceof Error && error.message.trim()
      ? error.message.trim()
      : t('backup.download')
    ElMessage.error(message)
  }
}

async function submitSchedule() {
  scheduleDialog.loading = true
  try {
    await request.post('/api/backup/schedule/create', {
      database: props.databaseName,
      tableName: scheduleDialog.tableName.trim(),
      intervalMinutes: scheduleDialog.intervalMinutes,
      compress: true
    })
    scheduleDialog.visible = false
    scheduleDialog.tableName = ''
    await loadAll()
    ElMessage.success(t('backup.scheduleCreated'))
  } finally {
    scheduleDialog.loading = false
  }
}

async function deleteSchedule(id: string) {
  await request.post('/api/backup/schedule/delete', { id })
  await loadAll()
  ElMessage.success(t('workspace.objectDeleted'))
}

async function monitorTask(taskId: string) {
  cleanupTaskTimer()

  const tick = async () => {
    const task = await request.get<BackupTask>('/api/backup/task', {
      params: { id: taskId }
    })
    activeTask.value = task

    if (task.status === 'success') {
      cleanupTaskTimer()
      await loadAll()
      emit('refresh-explorer')
      ElMessage.success(task.type === 'restore' ? t('backup.restoreSuccess') : t('backup.createSuccess'))
      return
    }

    if (task.status === 'failed') {
      cleanupTaskTimer()
      ElMessage.error(task.message || t('workspace.smartImportFailed'))
      return
    }

    taskTimer = window.setTimeout(tick, 1200)
  }

  await tick()
}

function cleanupTaskTimer() {
  if (taskTimer !== null) {
    window.clearTimeout(taskTimer)
    taskTimer = null
  }
}

function formatBytes(size: number) {
  if (size < 1024) {
    return `${size} B`
  }
  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`
  }
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}

function formatDateTime(value?: string) {
  if (!value) {
    return '-'
  }

  return value.replace('T', ' ').replace(/\.\d+Z?$/, '').replace(/Z$/, '')
}

watch(
  () => props.databaseName,
  () => {
    loadAll()
  }
)

onMounted(() => {
  loadAll()
})

onBeforeUnmount(() => {
  cleanupTaskTimer()
})
</script>

<style scoped>
.backup-panel {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.backup-toolbar,
.backup-task,
.backup-schedule,
.backup-list {
  padding: 20px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-lg);
  background: var(--devops-bg-panel);
  box-shadow: var(--devops-shadow-xs);
}

.backup-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
}

.backup-toolbar__meta {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.backup-toolbar__meta h3,
.backup-section__header h4 {
  margin: 0;
  color: var(--devops-text-primary);
}

.backup-toolbar__meta p {
  margin: 0;
  color: var(--devops-text-secondary);
}

.backup-toolbar__eyebrow {
  color: var(--devops-primary);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.backup-toolbar__actions,
.backup-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.backup-toolbar__actions :deep(.el-button) {
  min-width: 108px;
}

.backup-task__head,
.backup-section__header,
.schedule-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.backup-task p {
  margin: 10px 0 0;
  color: var(--devops-text-secondary);
}

.backup-task :deep(.el-progress-bar__outer) {
  background: var(--devops-bg-hover);
}

.schedule-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.schedule-item {
  padding: 14px 16px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-lg);
  background: var(--devops-bg-panel-soft);
}

.schedule-item__meta {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.schedule-item__meta strong {
  color: var(--devops-text-primary);
}

.schedule-item__meta span,
.schedule-item__meta small {
  color: var(--devops-text-secondary);
}

.backup-table {
  width: 100%;
}

.backup-actions {
  justify-content: flex-start;
}

.backup-actions :deep(.el-button) {
  padding: 0;
}

@media (max-width: 960px) {
  .backup-toolbar {
    flex-direction: column;
  }
}
</style>
