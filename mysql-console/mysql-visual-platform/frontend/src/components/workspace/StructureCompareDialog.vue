<template>
  <el-dialog
    :model-value="visible"
    :title="dialogTitle"
    width="1180px"
    class="workspace-dialog structure-compare-dialog"
    @update:model-value="emit('update:visible', $event)"
  >
    <div class="compare-shell">
      <section class="glass-subpanel compare-hero">
        <div class="compare-hero__copy">
          <span class="compare-hero__eyebrow">{{ text('eyebrow') }}</span>
          <h3>{{ text(scope === 'database' ? 'databaseModeTitle' : 'tableModeTitle') }}</h3>
          <p>{{ text(scope === 'database' ? 'databaseModeHint' : 'tableModeHint') }}</p>
        </div>
        <div class="compare-hero__actions">
          <el-button class="soft-button" :loading="comparing" @click="runCompare">
            {{ text('compareAction') }}
          </el-button>
          <el-button
            class="soft-button"
            :disabled="selectedStatements.length === 0 || syncing"
            :loading="syncing"
            @click="syncSelected"
          >
            {{ text('syncAction') }}
          </el-button>
        </div>
      </section>

      <section class="glass-subpanel compare-form">
        <div class="compare-grid">
          <el-form-item :label="text('sourceDatabase')">
            <el-input :model-value="sourceDatabase" readonly />
          </el-form-item>

          <el-form-item v-if="scope === 'table'" :label="text('sourceTable')">
            <el-input :model-value="sourceTable" readonly />
          </el-form-item>

          <el-form-item :label="text('targetDatabase')">
            <el-select v-model="targetDatabase" filterable @change="handleTargetDatabaseChange">
              <el-option v-for="database in databases" :key="database" :label="database" :value="database" />
            </el-select>
          </el-form-item>

          <el-form-item v-if="scope === 'table'" :label="text('targetTable')">
            <el-select v-model="targetTable" filterable>
              <el-option v-for="table in targetTables" :key="table" :label="table" :value="table" />
            </el-select>
          </el-form-item>
        </div>
      </section>

      <section v-if="items.length > 0" class="glass-subpanel compare-summary">
        <div class="compare-summary__cards">
          <article class="summary-card summary-card--add">
            <span>{{ text('addCount') }}</span>
            <strong>{{ addCount }}</strong>
          </article>
          <article class="summary-card summary-card--modify">
            <span>{{ text('modifyCount') }}</span>
            <strong>{{ modifyCount }}</strong>
          </article>
          <article class="summary-card summary-card--remove">
            <span>{{ text('removeCount') }}</span>
            <strong>{{ removeCount }}</strong>
          </article>
          <article class="summary-card summary-card--selected">
            <span>{{ text('selectedCount') }}</span>
            <strong>{{ selectedStatements.length }}</strong>
          </article>
        </div>
      </section>

      <section v-if="items.length > 0" class="glass-subpanel compare-result">
        <div class="compare-section__header">
          <h4>{{ text('diffList') }}</h4>
        </div>
        <el-table
          :data="items"
          class="compare-table"
          row-key="id"
          :row-class-name="resolveRowClassName"
        >
          <el-table-column width="64">
            <template #default="{ row }">
              <el-checkbox v-model="row.checked" :disabled="row.statements.length === 0" />
            </template>
          </el-table-column>
          <el-table-column :label="text('status')" width="110">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row.status)" effect="light">{{ statusLabel(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="category" :label="text('category')" width="130" />
          <el-table-column prop="objectName" :label="text('objectName')" min-width="180" />
          <el-table-column prop="title" :label="text('title')" min-width="220" />
          <el-table-column prop="detail" :label="text('detail')" min-width="260" show-overflow-tooltip />
          <el-table-column :label="text('safe')" width="100">
            <template #default="{ row }">
              <el-tag :type="row.safe ? 'success' : 'danger'" effect="light">
                {{ row.safe ? text('safeYes') : text('safeNo') }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section v-if="items.length > 0" class="glass-subpanel compare-sql">
        <div class="compare-section__header">
          <h4>{{ text('sqlPreview') }}</h4>
          <div class="compare-section__actions">
            <el-button class="soft-button" :disabled="selectedStatements.length === 0" @click="copySql">
              {{ text('copySql') }}
            </el-button>
          </div>
        </div>
        <el-input
          :model-value="sqlPreview"
          type="textarea"
          readonly
          :autosize="{ minRows: 10, maxRows: 18 }"
        />
      </section>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import request from '@/utils/request'
import { useI18n } from '@/utils/i18n'

type CompareScope = 'database' | 'table'
type DiffStatus = 'add' | 'remove' | 'modify'

interface SchemaDiffItem {
  id: string
  category: string
  objectName: string
  title: string
  detail: string
  status: DiffStatus
  sourceValue?: string
  targetValue?: string
  statements: string[]
  checked: boolean
  safe: boolean
}

interface SchemaCompareResponse {
  items: SchemaDiffItem[]
}

const props = defineProps<{
  visible: boolean
  scope: CompareScope
  sourceDatabase: string
  sourceTable?: string
}>()

const emit = defineEmits<{
  (event: 'update:visible', value: boolean): void
  (event: 'refresh-explorer'): void
}>()

const { isChinese } = useI18n()

const databases = ref<string[]>([])
const targetTables = ref<string[]>([])
const targetDatabase = ref('')
const targetTable = ref('')
const items = ref<SchemaDiffItem[]>([])
const comparing = ref(false)
const syncing = ref(false)

const dialogTitle = computed(() => text(props.scope === 'database' ? 'databaseDialogTitle' : 'tableDialogTitle'))
const scope = computed(() => props.scope)

const addCount = computed(() => items.value.filter((item) => item.status === 'add').length)
const modifyCount = computed(() => items.value.filter((item) => item.status === 'modify').length)
const removeCount = computed(() => items.value.filter((item) => item.status === 'remove').length)
const selectedStatements = computed(() => {
  const priority = (item: SchemaDiffItem) => {
    if (item.category === 'constraint' && item.status !== 'add') return 10
    if (item.category === 'index' && item.status !== 'add') return 20
    if (item.category === 'column' && item.status !== 'remove') return 30
    if (item.category === 'table') return 40
    if (item.category === 'index' && item.status === 'add') return 50
    if (item.category === 'constraint' && item.status === 'add') return 60
    if (item.category === 'column' && item.status === 'remove') return 70
    return 80
  }

  return items.value
    .filter((item) => item.checked)
    .slice()
    .sort((left, right) => priority(left) - priority(right))
    .flatMap((item) => item.statements)
    .filter(Boolean)
})
const sqlPreview = computed(() => selectedStatements.value.join(';\n\n') + (selectedStatements.value.length > 0 ? ';' : ''))

function text(key: string) {
  const zh: Record<string, string> = {
    eyebrow: '结构对比',
    databaseDialogTitle: '数据库结构对比',
    tableDialogTitle: '数据表结构对比',
    databaseModeTitle: '数据库 vs 数据库',
    tableModeTitle: '数据表 vs 数据表',
    databaseModeHint: '对比两个数据库中的表结构差异，并生成目标库同步 SQL。',
    tableModeHint: '对比两个数据表的字段、索引和约束差异，并生成目标表同步 SQL。',
    compareAction: '开始对比',
    syncAction: '执行同步',
    sourceDatabase: '源数据库',
    sourceTable: '源数据表',
    targetDatabase: '目标数据库',
    targetTable: '目标数据表',
    addCount: '新增',
    modifyCount: '修改',
    removeCount: '删除',
    selectedCount: '待执行 SQL',
    diffList: '差异列表',
    status: '状态',
    category: '分类',
    objectName: '对象',
    title: '标题',
    detail: '说明',
    safe: '安全级别',
    safeYes: '安全',
    safeNo: '谨慎',
    sqlPreview: 'SQL 预览',
    copySql: '复制 SQL',
    statusAdd: '新增',
    statusModify: '修改',
    statusRemove: '删除',
    compareSuccess: '结构对比完成。',
    syncConfirm: '确认执行已勾选的结构同步 SQL 吗？该操作只会修改目标库表结构。',
    syncSuccess: '结构同步已完成。'
  }
  const en: Record<string, string> = {
    eyebrow: 'Structure Compare',
    databaseDialogTitle: 'Database Structure Compare',
    tableDialogTitle: 'Table Structure Compare',
    databaseModeTitle: 'Database vs Database',
    tableModeTitle: 'Table vs Table',
    databaseModeHint: 'Compare table structures between two databases and generate sync SQL for the target database.',
    tableModeHint: 'Compare columns, indexes, and constraints between two tables and generate sync SQL for the target table.',
    compareAction: 'Compare',
    syncAction: 'Sync Selected',
    sourceDatabase: 'Source Database',
    sourceTable: 'Source Table',
    targetDatabase: 'Target Database',
    targetTable: 'Target Table',
    addCount: 'Add',
    modifyCount: 'Modify',
    removeCount: 'Remove',
    selectedCount: 'Selected SQL',
    diffList: 'Differences',
    status: 'Status',
    category: 'Category',
    objectName: 'Object',
    title: 'Title',
    detail: 'Detail',
    safe: 'Risk',
    safeYes: 'Safe',
    safeNo: 'Caution',
    sqlPreview: 'SQL Preview',
    copySql: 'Copy SQL',
    statusAdd: 'Add',
    statusModify: 'Modify',
    statusRemove: 'Remove',
    compareSuccess: 'Structure compare completed.',
    syncConfirm: 'Execute the selected structure sync SQL now? Only target structures will be changed.',
    syncSuccess: 'Structure sync completed.'
  }
  return (isChinese.value ? zh : en)[key] || key
}

async function loadDatabases() {
  databases.value = await request.get<string[]>('/api/metadata/databases')
  if (!targetDatabase.value) {
    targetDatabase.value = databases.value.find((item) => item !== props.sourceDatabase) || databases.value[0] || ''
  }
  if (props.scope === 'table' && targetDatabase.value) {
    await loadTargetTables(targetDatabase.value)
  }
}

async function loadTargetTables(database: string) {
  targetTables.value = await request.get<string[]>('/api/metadata/tables', {
    params: { db: database }
  })
  if (!targetTable.value) {
    targetTable.value = props.sourceTable && targetTables.value.includes(props.sourceTable)
      ? props.sourceTable
      : targetTables.value[0] || ''
  }
}

async function handleTargetDatabaseChange(value: string) {
  if (props.scope === 'table') {
    targetTable.value = ''
    await loadTargetTables(value)
  }
}

async function runCompare() {
  comparing.value = true
  try {
    const result = await request.post<SchemaCompareResponse>('/api/schema/compare', {
      scope: props.scope,
      sourceDatabase: props.sourceDatabase,
      sourceTable: props.sourceTable || '',
      targetDatabase: targetDatabase.value,
      targetTable: targetTable.value
    })
    items.value = (result.items || []).map((item) => ({ ...item }))
    ElMessage.success(text('compareSuccess'))
  } finally {
    comparing.value = false
  }
}

async function syncSelected() {
  if (selectedStatements.value.length === 0) {
    return
  }
  await ElMessageBox.confirm(text('syncConfirm'), dialogTitle.value)
  syncing.value = true
  try {
    await request.post('/api/sql/execute-batch', {
      database: targetDatabase.value,
      statements: selectedStatements.value
    })
    ElMessage.success(text('syncSuccess'))
    emit('refresh-explorer')
    await runCompare()
  } finally {
    syncing.value = false
  }
}

async function copySql() {
  await navigator.clipboard.writeText(sqlPreview.value)
  ElMessage.success(text('copySql'))
}

function statusLabel(status: DiffStatus) {
  if (status === 'add') return text('statusAdd')
  if (status === 'modify') return text('statusModify')
  return text('statusRemove')
}

function statusTagType(status: DiffStatus) {
  if (status === 'add') return 'success'
  if (status === 'modify') return 'warning'
  return 'danger'
}

function resolveRowClassName({ row }: { row: SchemaDiffItem }) {
  return `compare-row compare-row--${row.status}`
}

watch(
  () => props.visible,
  async (visible) => {
    if (!visible) {
      return
    }
    items.value = []
    await loadDatabases()
  }
)
</script>

<style scoped>
.compare-shell {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.compare-hero,
.compare-form,
.compare-summary,
.compare-result,
.compare-sql {
  padding: 20px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-lg);
  background: var(--devops-bg-panel);
  box-shadow: var(--devops-shadow-xs);
}

.compare-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
}

.compare-hero__copy {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.compare-hero__copy h3,
.compare-section__header h4 {
  margin: 0;
  color: var(--devops-text-primary);
}

.compare-hero__copy p {
  margin: 0;
  color: var(--devops-text-secondary);
}

.compare-hero__eyebrow {
  color: var(--devops-primary);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.compare-hero__actions,
.compare-section__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.compare-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.compare-summary__cards {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.summary-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 14px 16px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-lg);
  background: var(--devops-bg-panel-soft);
}

.summary-card span {
  color: var(--devops-text-secondary);
  font-size: 12px;
}

.summary-card strong {
  font-size: 24px;
  color: var(--devops-text-primary);
}

.summary-card--add {
  box-shadow: inset 0 0 0 1px rgba(103, 194, 58, 0.2);
}

.summary-card--modify {
  box-shadow: inset 0 0 0 1px rgba(230, 162, 60, 0.2);
}

.summary-card--remove {
  box-shadow: inset 0 0 0 1px rgba(245, 108, 108, 0.2);
}

.summary-card--selected {
  box-shadow: inset 0 0 0 1px rgba(64, 158, 255, 0.2);
}

.compare-section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--devops-border-light);
}

.compare-table :deep(.compare-row--add) {
  --el-table-tr-bg-color: rgba(103, 194, 58, 0.08);
}

.compare-table :deep(.compare-row--modify) {
  --el-table-tr-bg-color: rgba(230, 162, 60, 0.08);
}

.compare-table :deep(.compare-row--remove) {
  --el-table-tr-bg-color: rgba(245, 108, 108, 0.08);
}

.compare-sql :deep(.el-textarea__inner) {
  min-height: 220px !important;
  font-family: "JetBrains Mono", Consolas, "Courier New", monospace;
  font-size: 12px;
  line-height: 1.6;
}

@media (max-width: 960px) {
  .compare-hero,
  .compare-section__header {
    flex-direction: column;
    align-items: stretch;
  }

  .compare-grid,
  .compare-summary__cards {
    grid-template-columns: 1fr;
  }
}
</style>
