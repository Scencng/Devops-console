<template>
  <section class="data-tab">
    <div class="panel-toolbar">
      <div class="toolbar-heading">
        <span class="header-kicker">{{ t('tableData.title') }}</span>
        <h3>{{ db }} / {{ table }}</h3>
      </div>
    </div>

    <div class="filter-row">
      <div class="filter-field">
        <el-input
          v-model="keywordInput"
          class="filter-input"
          clearable
          :placeholder="t('tableData.filterPlaceholder')"
          @keyup.enter="applyKeywordFilter"
          @clear="applyKeywordFilter"
        />
      </div>
      <div class="filter-row__controls">
        <el-button class="soft-button toolbar-button" @click="applyKeywordFilter">{{ t('tableData.filter') }}</el-button>
        <el-button class="soft-button toolbar-button" @click="clearKeywordFilter">{{ t('tableData.clearFilter') }}</el-button>
        <el-button class="soft-button toolbar-button" @click="safeRefresh">{{ t('common.refresh') }}</el-button>
        <el-button class="soft-button toolbar-button" @click="addRow">{{ t('tableData.addRow') }}</el-button>
        <el-button
          type="primary"
          class="soft-button primary-button toolbar-button toolbar-button--primary filter-row__save"
          :disabled="!hasPendingChanges || saving"
          :loading="saving"
          @click="saveChanges"
        >
          {{ t('tableData.saveChanges') }}
        </el-button>
      </div>
    </div>

    <div class="change-summary glass-subpanel">
      <span class="summary-pill">{{ t('tableData.total') }}: {{ totalRows }}</span>
      <span class="summary-pill">{{ t('tableData.inserted') }}: {{ insertedRows.length }}</span>
      <span class="summary-pill">{{ t('tableData.updated') }}: {{ updatedRows.length }}</span>
      <span class="summary-pill">{{ t('tableData.deleted') }}: {{ deletedRows.length }}</span>
      <span class="summary-pill">{{ t('tableData.sort') }}: {{ sortLabel }}</span>
    </div>

    <div v-loading="loading" class="glass-subpanel table-panel">
      <template v-if="columns.length > 0">
        <el-table
          ref="tableRef"
          :data="rows"
          row-key="id"
          height="100%"
          class="anime-data-table"
          table-layout="auto"
          highlight-current-row
          :row-class-name="getRowClassName"
          @current-change="handleCurrentChange"
          @selection-change="handleSelectionChange"
          @sort-change="handleSortChange"
        >
          <el-table-column type="selection" width="48" fixed="left" align="center" header-align="center" />

          <el-table-column
            v-for="column in columns"
            :key="column"
            :prop="column"
            :label="column"
            :width="getColumnWidth(column)"
            :min-width="getColumnMinWidth(column)"
            :class-name="getColumnClassName(column)"
            :show-overflow-tooltip="false"
            align="center"
            header-align="center"
            sortable="custom"
          >
            <template #default="{ row }">
              <div
                class="editable-cell"
                :class="{
                  'editable-cell--editing': isEditingCell(row.id, column),
                  'editable-cell--changed': isChangedCell(row, column)
                }"
                @dblclick="activateEdit(row.id, column)"
              >
                <el-input
                  v-if="isEditingCell(row.id, column) && row.status !== 'deleted'"
                  :model-value="toInputValue(row.data[column])"
                  size="small"
                  @update:model-value="updateCellValue(row.id, column, $event)"
                  @blur="stopEditing"
                  @keyup.enter="stopEditing"
                  @keyup.esc="cancelEditing(row.id, column)"
                />
                <span
                  v-else
                  class="cell-text"
                  :class="{
                    'cell-text--datetime': isDateTimeColumn(column, row.data[column]),
                    'cell-text--wide': isWideTextColumn(column)
                  }"
                  :title="getCellTooltip(row.data[column], column)"
                >
                  {{ formatCellValue(row.data[column], column) }}
                </span>
              </div>
            </template>
          </el-table-column>

          <el-table-column :label="t('tableData.status')" width="132" fixed="right" align="center" header-align="center">
            <template #default="{ row }">
              <el-tag :type="getStatusTagType(row.status)" effect="light">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column :label="t('tableData.actions')" width="128" fixed="right" align="center">
            <template #default="{ row }">
              <el-button
                v-if="row.status !== 'deleted'"
                type="danger"
                link
                @click="handleDeleteAction(row)"
              >
                {{ t('tableData.delete') }}
              </el-button>
              <el-button
                v-else
                type="warning"
                link
                @click="undoDeleteRow(row)"
              >
                {{ t('tableData.undo') }}
              </el-button>
            </template>
          </el-table-column>

          <template #empty>
            <div class="table-empty-state">
              <el-empty :description="t('tableData.noData')" />
            </div>
          </template>
        </el-table>
      </template>

      <el-empty
        v-else
        :description="t('tableData.noData')"
      />
    </div>

    <div class="pagination-wrap">
      <el-pagination
        background
        layout="total, sizes, prev, pager, next"
        :current-page="currentPage"
        :page-size="pageSize"
        :page-sizes="[20, 50, 100, 200]"
        :total="totalRows"
        @current-change="handlePageChange"
        @size-change="handlePageSizeChange"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { Sort } from 'element-plus'

import { useI18n } from '@/utils/i18n'
import request from '@/utils/request'

type RowStatus = 'clean' | 'inserted' | 'updated' | 'deleted'
type CellValue = unknown
type TableRowData = Record<string, CellValue>

interface TableDataResponse {
  columns?: string[]
  rows?: TableRowData[]
  limit?: number
  offset?: number
  total?: number
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
  keyword?: string
}

interface EditableRow {
  id: string
  data: TableRowData
  status: RowStatus
  originalData: TableRowData | null
  previousStatus: Exclude<RowStatus, 'deleted'> | null
}

interface DeletedRow {
  id: string
  data: TableRowData
}

interface UpdatedRowChange {
  id: string
  original: TableRowData
  current: TableRowData
}

interface SQLBatch {
  insertStatements: string[]
  updateStatements: string[]
  deleteStatements: string[]
  allStatements: string[]
}

interface ExecuteBatchResponse {
  success: boolean
  message: string
  affected_rows: number
}

interface QueryTableResponse {
  rows?: Array<Record<string, unknown>>
}

const props = defineProps<{
  db: string
  table: string
}>()

const { t } = useI18n()
const tableRef = ref()
const loading = ref(false)
const saving = ref(false)
const rows = ref<EditableRow[]>([])
const columns = ref<string[]>([])
const primaryKeyColumns = ref<string[]>([])
const deletedRows = ref<DeletedRow[]>([])
const selectedRows = ref<EditableRow[]>([])
const pageSize = ref(50)
const currentPage = ref(1)
const totalRows = ref(0)
const editingCell = ref<{ rowId: string; column: string } | null>(null)
const keywordInput = ref('')
const activeKeyword = ref('')
const sortBy = ref('')
const sortOrder = ref<'asc' | 'desc'>('asc')

const statusTagType: Record<RowStatus, '' | 'success' | 'warning' | 'danger'> = {
  clean: '',
  inserted: 'success',
  updated: 'warning',
  deleted: 'danger'
}

const currentOffset = computed(() => (currentPage.value - 1) * pageSize.value)
const insertedRows = computed(() => rows.value.filter((row) => row.status === 'inserted'))
const updatedRows = computed<UpdatedRowChange[]>(() =>
  rows.value
    .filter((row) => row.status === 'updated' && row.originalData)
    .map((row) => ({
      id: row.id,
      original: cloneRowData(row.originalData!),
      current: cloneRowData(row.data)
    }))
)
const hasPendingChanges = computed(() =>
  insertedRows.value.length > 0 || updatedRows.value.length > 0 || deletedRows.value.length > 0
)
const sortLabel = computed(() => {
  if (!sortBy.value) {
    return 'Default'
  }

  return `${sortBy.value} ${sortOrder.value.toUpperCase()}`
})

async function fetchTableData() {
  loading.value = true

  try {
    await fetchPrimaryKeyColumns()

    const data = await request.get<TableDataResponse>('/api/data/table', {
      params: {
        db: props.db,
        table: props.table,
        limit: pageSize.value,
        offset: currentOffset.value,
        keyword: activeKeyword.value,
        sortBy: sortBy.value,
        sortOrder: sortOrder.value
      }
    })

    const nextRows = data.rows ?? []
    const nextColumns = data.columns?.length ? data.columns : Object.keys(nextRows[0] ?? {})

    columns.value = nextColumns
    rows.value = nextRows.map((row, index) => ({
      id: createRowId(index),
      data: normalizeRowData(row, nextColumns),
      status: 'clean',
      originalData: null,
      previousStatus: null
    }))
    totalRows.value = data.total ?? nextRows.length
    selectedRows.value = []
    deletedRows.value = []
    editingCell.value = null
  } finally {
    loading.value = false
  }
}

async function fetchPrimaryKeyColumns() {
  const result = await request.post<QueryTableResponse>('/api/query/execute', {
    sql: `SHOW KEYS FROM ${quoteIdentifier(props.table)} WHERE Key_name = 'PRIMARY'`,
    database: props.db
  }).catch(() => ({ rows: [] } as QueryTableResponse))

  primaryKeyColumns.value = (result.rows ?? [])
    .map((row: Record<string, unknown>) => String(row.Column_name ?? row.COLUMN_NAME ?? row.column_name ?? '').trim())
    .filter((column: string) => column.length > 0)
}

async function safeRefresh() {
  if (!ensureNoPendingChanges('Refresh')) {
    return
  }

  await fetchTableData()
}

function createRowId(seed: number) {
  return `${Date.now()}-${seed}-${Math.random().toString(36).slice(2, 10)}`
}

function normalizeRowData(row: TableRowData, currentColumns: string[]) {
  const normalized: TableRowData = {}
  for (const column of currentColumns) {
    normalized[column] = normalizeDisplayValue(row[column], column)
  }

  return normalized
}

function cloneRowData(row: TableRowData) {
  return Object.fromEntries(Object.entries(row)) as TableRowData
}

function ensureNoPendingChanges(actionLabel: string) {
  if (!hasPendingChanges.value) {
    return true
  }

  ElMessage.warning(t('tableData.actionBlocked', { action: actionLabel }))
  return false
}

function handlePageChange(page: number) {
  if (!ensureNoPendingChanges('Page switching')) {
    return
  }

  currentPage.value = page
  void fetchTableData()
}

function handlePageSizeChange(size: number) {
  if (!ensureNoPendingChanges('Page size change')) {
    return
  }

  pageSize.value = size
  currentPage.value = 1
  void fetchTableData()
}

function handleCurrentChange() {}

function handleSelectionChange(nextRows: EditableRow[]) {
  selectedRows.value = nextRows
}

function isEditingCell(rowId: string, column: string) {
  return editingCell.value?.rowId === rowId && editingCell.value?.column === column
}

function activateEdit(rowId: string, column: string) {
  const row = rows.value.find((item) => item.id === rowId)
  if (!row || row.status === 'deleted') {
    return
  }

  editingCell.value = { rowId, column }
}

function stopEditing() {
  editingCell.value = null
}

function cancelEditing(rowId: string, column: string) {
  const row = rows.value.find((item) => item.id === rowId)
  if (!row) {
    stopEditing()
    return
  }

  if (row.status === 'updated' && row.originalData) {
    row.data[column] = row.originalData[column]
    syncUpdatedState(row)
  }

  stopEditing()
}

function updateCellValue(rowId: string, column: string, nextValue: string) {
  const row = rows.value.find((item) => item.id === rowId)
  if (!row || row.status === 'deleted') {
    return
  }

  if (row.status === 'clean' && !row.originalData) {
    row.originalData = cloneRowData(row.data)
  }

  row.data[column] = coerceEditedValue(nextValue, row.originalData?.[column] ?? row.data[column])
  syncUpdatedState(row)
}

function coerceEditedValue(rawValue: string, referenceValue: CellValue) {
  if (rawValue === '') {
    return null
  }

  if (typeof referenceValue === 'number') {
    const parsed = Number(rawValue)
    return Number.isNaN(parsed) ? rawValue : parsed
  }

  if (typeof referenceValue === 'boolean') {
    return rawValue === 'true'
  }

  return rawValue
}

function syncUpdatedState(row: EditableRow) {
  if (row.status === 'inserted') {
    return
  }

  if (!row.originalData) {
    row.originalData = cloneRowData(row.data)
    row.status = 'clean'
    return
  }

  const changed = columns.value.some((column) => !isSameValue(row.originalData?.[column], row.data[column]))
  row.status = changed ? 'updated' : 'clean'
}

function isSameValue(left: CellValue, right: CellValue) {
  return left === right
}

function isChangedCell(row: EditableRow, column: string) {
  return row.status === 'updated' && !!row.originalData && !isSameValue(row.originalData[column], row.data[column])
}

function addRow() {
  const data = Object.fromEntries(columns.value.map((column) => [column, null])) as TableRowData
  const row: EditableRow = {
    id: createRowId(rows.value.length),
    data,
    status: 'inserted',
    originalData: null,
    previousStatus: null
  }

  rows.value.unshift(row)
  totalRows.value += 1
  if (columns.value[0]) {
    activateEdit(row.id, columns.value[0])
  }
}

function handleDeleteAction(row: EditableRow) {
  if (row.status === 'inserted') {
    rows.value = rows.value.filter((item) => item.id !== row.id)
    selectedRows.value = selectedRows.value.filter((item) => item.id !== row.id)
    if (editingCell.value?.rowId === row.id) {
      stopEditing()
    }
    totalRows.value = Math.max(0, totalRows.value - 1)
    return
  }

  if (row.status === 'deleted') {
    return
  }

  row.previousStatus = row.status === 'updated' ? 'updated' : 'clean'
  row.status = 'deleted'
  upsertDeletedRow(row)

  if (editingCell.value?.rowId === row.id) {
    stopEditing()
  }
}

function undoDeleteRow(row: EditableRow) {
  if (row.status !== 'deleted') {
    return
  }

  row.status = row.previousStatus ?? 'clean'
  row.previousStatus = null
  deletedRows.value = deletedRows.value.filter((item) => item.id !== row.id)
}

function upsertDeletedRow(row: EditableRow) {
  const snapshot = cloneRowData(row.originalData ?? row.data)
  const nextDeleted: DeletedRow = {
    id: row.id,
    data: snapshot
  }

  const index = deletedRows.value.findIndex((item) => item.id === row.id)
  if (index >= 0) {
    deletedRows.value.splice(index, 1, nextDeleted)
    return
  }

  deletedRows.value.push(nextDeleted)
}

async function saveChanges() {
  const batch = buildSQLBatch()
  if (batch.allStatements.length === 0) {
    ElMessage.info(t('tableData.noPendingChanges'))
    return batch
  }

  saving.value = true

  try {
    const result = await request.post<ExecuteBatchResponse>('/api/sql/execute-batch', {
      database: props.db,
      statements: batch.allStatements,
      insertStatements: batch.insertStatements,
      updateStatements: batch.updateStatements,
      deleteStatements: batch.deleteStatements
    })

    resetPersistedState()
    await fetchTableData()
    ElMessage.success(result.message || t('query.rowsAffected', { count: result.affected_rows }))
  } finally {
    saving.value = false
  }

  return batch
}

function buildSQLBatch(): SQLBatch {
  const insertStatements = insertedRows.value.map((row) => buildInsertSQL(row.data))
  const updateStatements = updatedRows.value.map((row) => buildUpdateSQL(row.original, row.current))
  const deleteStatements = deletedRows.value.map((row) => buildDeleteSQL(row.data))

  return {
    insertStatements,
    updateStatements,
    deleteStatements,
    allStatements: [...insertStatements, ...updateStatements, ...deleteStatements]
  }
}

function buildInsertSQL(row: TableRowData) {
  const quotedColumns = columns.value.map(quoteIdentifier).join(', ')
  const values = columns.value.map((column) => formatSQLValue(row[column])).join(', ')

  return `INSERT INTO ${qualifyTableName()} (${quotedColumns}) VALUES (${values});`
}

function buildUpdateSQL(original: TableRowData, current: TableRowData) {
  const changedColumns = columns.value.filter((column) => !isSameValue(original[column], current[column]))
  const assignments = changedColumns
    .map((column) => `${quoteIdentifier(column)} = ${formatSQLValue(current[column])}`)
    .join(', ')

  return `UPDATE ${qualifyTableName()} SET ${assignments} WHERE ${buildWhereClause(original, true)};`
}

function buildDeleteSQL(row: TableRowData) {
  return `DELETE FROM ${qualifyTableName()} WHERE ${buildWhereClause(row, true)};`
}

function qualifyTableName() {
  return `${quoteIdentifier(props.db)}.${quoteIdentifier(props.table)}`
}

function buildWhereClause(row: TableRowData, preferPrimaryKey = false) {
  const keyColumns =
    preferPrimaryKey && primaryKeyColumns.value.length > 0
      ? primaryKeyColumns.value.filter((column) => columns.value.includes(column))
      : columns.value

  const sourceColumns = keyColumns.length > 0 ? keyColumns : columns.value

  return sourceColumns
    .map((column) => {
      const value = row[column]
      if (value == null) {
        return `${quoteIdentifier(column)} IS NULL`
      }

      return `${quoteIdentifier(column)} = ${formatSQLValue(value)}`
    })
    .join(' AND ')
}

function quoteIdentifier(identifier: string) {
  return `\`${identifier.split('`').join('``')}\``
}

function formatSQLValue(value: CellValue) {
  if (value == null) {
    return 'NULL'
  }

  if (typeof value === 'number') {
    return Number.isFinite(value) ? String(value) : 'NULL'
  }

  if (typeof value === 'boolean') {
    return value ? '1' : '0'
  }

  if (value instanceof Date) {
    return `'${value.toISOString().slice(0, 19).replace('T', ' ')}'`
  }

  return `'${String(value)
    .split('\\')
    .join('\\\\')
    .split("'")
    .join("\\'")}'`
}

function getStatusTagType(status: RowStatus) {
  return statusTagType[status]
}

function getRowClassName({ row }: { row: EditableRow }) {
  if (row.status === 'inserted') {
    return 'row--inserted'
  }

  if (row.status === 'updated') {
    return 'row--updated'
  }

  if (row.status === 'deleted') {
    return 'row--deleted'
  }

  return ''
}

function resetPersistedState() {
  deletedRows.value = []
  rows.value = rows.value.map((row) => ({
    ...row,
    status: 'clean',
    originalData: null,
    previousStatus: null
  }))
}

function toInputValue(value: CellValue) {
  if (value == null) {
    return ''
  }

  if (typeof value === 'string') {
    return value
  }

  return String(value)
}

function formatCellValue(value: CellValue, column?: string) {
  if (value == null) {
    return 'NULL'
  }

  if (isPhoneLikeColumn(column ?? '')) {
    return formatPhoneDisplayValue(value)
  }

  if (isDateTimeColumn(column ?? '', value)) {
    return formatDateTimeValue(value)
  }

  if (typeof value === 'string') {
    return normalizeRenderedTextValue(value, column)
  }

  return String(value)
}

function getCellTooltip(value: CellValue, column: string) {
  return formatCellValue(value, column)
}

function isDateTimeColumn(column: string, value: CellValue) {
  const normalizedColumn = column.trim().toLowerCase()
  if (
    normalizedColumn.includes('time') ||
    normalizedColumn.includes('date') ||
    normalizedColumn.includes('created') ||
    normalizedColumn.includes('updated')
  ) {
    return true
  }

  if (value instanceof Date) {
    return true
  }

  return typeof value === 'string' && isISODateTimeLike(value)
}

function getColumnMinWidth(column: string) {
  if (isProfessionLikeColumn(column)) {
    return getProfessionColumnMinWidth(column)
  }

  if (isWideTextColumn(column)) {
    return getWideColumnMinWidth(column)
  }

  return isDateTimeColumn(column, null) ? 220 : 180
}

function getColumnWidth(column: string) {
  return undefined
}

function getProfessionColumnMinWidth(column: string) {
  const estimated = estimateContentColumnWidth(column, 0.82)
  return Math.min(Math.max(estimated + 60, 220), 340)
}

function getWideColumnMinWidth(column: string) {
  const estimated = estimateContentColumnWidth(column, 0.88)
  return Math.min(Math.max(estimated + 72, 240), 460)
}

function estimateContentColumnWidth(column: string, quantile = 0.85) {
  const headerWidth = estimateDisplayWidth(column)
  const widths = rows.value
    .map((row) => estimateDisplayWidth(formatCellValue(row.data[column], column)))
    .filter((width) => width > 0)
    .sort((left, right) => left - right)

  if (widths.length === 0) {
    return headerWidth
  }

  const index = Math.max(0, Math.min(widths.length - 1, Math.floor((widths.length - 1) * quantile)))
  return Math.max(headerWidth, widths[index])
}

function estimateDisplayWidth(value: CellValue) {
  const text = value == null ? '' : String(value)
  let width = 0

  for (const char of text) {
    width += /[\u4e00-\u9fff\u3400-\u4dbf\u3000-\u303f\uff00-\uffef]/.test(char) ? 18 : 9
  }

  return width
}

function getColumnClassName(column: string) {
  if (isPhoneLikeColumn(column)) {
    return 'column--phone'
  }

  if (isProfessionLikeColumn(column)) {
    return 'column--profession'
  }

  if (isWideTextColumn(column)) {
    return 'column--wide-text'
  }

  if (isDateTimeColumn(column, null)) {
    return 'column--datetime'
  }

  return ''
}

function isWideTextColumn(column: string) {
  const normalizedColumn = column.trim().toLowerCase()
  return (
    normalizedColumn.includes('profession') ||
    normalizedColumn.includes('occupation') ||
    normalizedColumn.includes('career') ||
    normalizedColumn.includes('job') ||
    normalizedColumn.includes('position') ||
    normalizedColumn.includes('description') ||
    normalizedColumn.includes('remark') ||
    normalizedColumn.includes('comment') ||
    normalizedColumn.includes('content') ||
    normalizedColumn.includes('address') ||
    normalizedColumn.includes('major') ||
    normalizedColumn.includes('specialty') ||
    normalizedColumn.includes('summary') ||
    normalizedColumn.includes('details') ||
    normalizedColumn.includes('notes')
  )
}

function legacyWideTextColumnMatcher(column: string) {
  const normalizedColumn = column.trim().toLowerCase()
  return (
    normalizedColumn.includes('description') ||
    normalizedColumn.includes('remark') ||
    normalizedColumn.includes('comment') ||
    normalizedColumn.includes('content') ||
    normalizedColumn.includes('address') ||
    normalizedColumn.includes('备注') ||
    normalizedColumn.includes('评论') ||
    normalizedColumn.includes('内容') ||
    normalizedColumn.includes('地址') ||
    normalizedColumn.includes('说明')
  )
}

function isPhoneLikeColumn(column: string) {
  const normalizedColumn = column.trim().toLowerCase()
  return (
    normalizedColumn === 'phone' ||
    normalizedColumn === 'mobile' ||
    normalizedColumn === 'tel' ||
    normalizedColumn === 'telephone' ||
    normalizedColumn.includes('phone_') ||
    normalizedColumn.includes('_phone') ||
    normalizedColumn.includes('mobile_') ||
    normalizedColumn.includes('_mobile') ||
    normalizedColumn.includes('telephone_') ||
    normalizedColumn.includes('_telephone')
  )
}

function formatPhoneDisplayValue(value: CellValue) {
  if (value == null) {
    return 'NULL'
  }

  const text = String(value).trim()
  if (!text) {
    return ''
  }

  if (/^-?\d+\.0+$/.test(text)) {
    return text.replace(/\.0+$/, '')
  }

  if (/^-?\d+(?:\.\d+)?$/.test(text)) {
    const [integerPart, decimalPart = ''] = text.split('.')
    if (/^0+$/.test(decimalPart)) {
      return integerPart
    }
  }

  return text
}

function isProfessionLikeColumn(column: string) {
  const normalizedColumn = column.trim().toLowerCase()
  return (
    normalizedColumn.includes('profession') ||
    normalizedColumn.includes('occupation') ||
    normalizedColumn.includes('career') ||
    normalizedColumn.includes('job') ||
    normalizedColumn.includes('position') ||
    normalizedColumn.includes('major') ||
    normalizedColumn.includes('specialty') ||
    normalizedColumn.includes('专业') ||
    normalizedColumn.includes('职业')
  )
}

function normalizeRenderedTextValue(value: string, column?: string) {
  const compact = value
    .replace(/\u00A0/g, ' ')
    .replace(/\u3000/g, ' ')
    .replace(/[\r\n\t]+/g, ' ')
    .trim()
    .replace(/\s+/g, ' ')

  if (!(column && (isProfessionLikeColumn(column) || isWideTextColumn(column)))) {
    return compact
  }

  let normalized = compact
  let previous = ''
  while (normalized !== previous) {
    previous = normalized
    normalized = normalized.replace(/([\u3400-\u9fff])\s+(?=[\u3400-\u9fff])/gu, '$1')
  }

  return normalized
}

function normalizeDisplayValue(value: CellValue, column?: string): CellValue {
  if (value == null) {
    return null
  }

  if (value instanceof Date) {
    return formatDateTimeValue(value)
  }

  if (typeof value === 'string' && isDateTimeColumn(column ?? '', value)) {
    return formatDateTimeValue(value)
  }

  return value
}

function formatDateTimeValue(value: CellValue) {
  if (value == null) {
    return 'NULL'
  }

  if (value instanceof Date) {
    return formatDateObject(value)
  }

  const text = String(value).trim()
  if (!text) {
    return ''
  }

  const parsed = parseDateValue(text)
  if (!parsed) {
    return text.replace('T', ' ').replace(/(\.\d+)?(Z|[+-]\d{2}:\d{2})$/, '')
  }

  return formatDateObject(parsed)
}

function parseDateValue(text: string) {
  const normalizedText = text.replace(/\//g, '-')
  const parsed = new Date(normalizedText)
  if (!Number.isNaN(parsed.getTime())) {
    return parsed
  }

  return null
}

function padDateSegment(value: number) {
  return String(value).padStart(2, '0')
}

function formatDateObject(value: Date) {
  return [
    value.getFullYear(),
    padDateSegment(value.getMonth() + 1),
    padDateSegment(value.getDate())
  ].join('-') + ' ' + [
    padDateSegment(value.getHours()),
    padDateSegment(value.getMinutes()),
    padDateSegment(value.getSeconds())
  ].join(':')
}

function isISODateTimeLike(value: string) {
  return /^\d{4}-\d{1,2}-\d{1,2}(?:[ T]\d{1,2}:\d{2}(?::\d{2})?(?:\.\d+)?)?(?:Z|[+-]\d{2}:\d{2})?$/.test(value)
}

function applyKeywordFilter() {
  if (!ensureNoPendingChanges('Filtering')) {
    return
  }

  activeKeyword.value = keywordInput.value.trim()
  currentPage.value = 1
  void fetchTableData()
}

function clearKeywordFilter() {
  if (!ensureNoPendingChanges('Filtering')) {
    return
  }

  keywordInput.value = ''
  activeKeyword.value = ''
  currentPage.value = 1
  void fetchTableData()
}

function handleSortChange(sort: Sort) {
  if (!ensureNoPendingChanges('Sorting')) {
    if (tableRef.value) {
      tableRef.value.clearSort()
    }
    return
  }

  sortBy.value = sort.prop ?? ''
  sortOrder.value = sort.order === 'descending' ? 'desc' : 'asc'
  if (!sort.order) {
    sortBy.value = ''
  }
  currentPage.value = 1
  void fetchTableData()
}

watch(
  () => [props.db, props.table],
  () => {
    currentPage.value = 1
    keywordInput.value = ''
    activeKeyword.value = ''
    sortBy.value = ''
    sortOrder.value = 'asc'
    void fetchTableData()
  }
)

defineExpose({
  saveChanges,
  buildSQLBatch
})

onMounted(() => {
  void fetchTableData()
})
</script>

<style scoped>
.panel-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  margin-bottom: 14px;
  padding: 18px 20px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-lg);
  background: var(--devops-bg-panel);
  box-shadow: var(--devops-shadow-xs);
}

.toolbar-heading {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
}

.toolbar-heading h3 {
  max-width: min(100%, 420px);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 20px;
}

.header-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  align-items: center;
}

.toolbar-actions {
  flex: 0 1 auto;
  justify-content: flex-end;
  gap: 14px;
}

.filter-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin: 4px 0 18px;
  padding: 16px 18px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-lg);
  background: var(--devops-bg-panel);
  box-shadow: var(--devops-shadow-xs);
}

.filter-field {
  display: flex;
  min-width: min(100%, 280px);
  width: min(100%, 360px);
  max-width: 360px;
  flex: 0 1 360px;
  padding: 0;
}

.filter-row__save {
  flex-shrink: 0;
}

.filter-row__controls {
  display: flex;
  flex: 1 1 auto;
  justify-content: flex-end;
  align-items: center;
  gap: 10px;
  min-width: 0;
  overflow-x: auto;
  overflow-y: hidden;
  padding-bottom: 2px;
}

.filter-input {
  width: 100%;
}

.change-summary {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  padding: 14px 16px;
  font-size: 13px;
  flex-wrap: wrap;
  align-items: center;
  border: 1px solid var(--devops-border-light);
  background: var(--devops-bg-panel);
  box-shadow: var(--devops-shadow-xs);
}

.summary-pill {
  display: inline-flex;
  align-items: center;
  min-height: 36px;
  padding: 7px 14px;
  border-radius: 999px;
  border: 1px solid var(--devops-border-light);
  background: var(--devops-bg-panel-soft);
  color: var(--devops-text-regular);
  font-weight: 600;
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: "Microsoft YaHei", "PingFang SC", "Hiragino Sans GB", "Noto Sans SC", sans-serif;
  letter-spacing: 0;
  word-spacing: 0;
  font-kerning: none;
  font-synthesis: none;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.82);
}

.summary-pill--wide {
  max-width: min(100%, 460px);
}

.table-panel {
  min-height: 540px;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  width: 100%;
  overflow-x: auto;
}

.toolbar-button-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: flex-end;
}

.toolbar-button {
  min-height: 36px;
  padding: 0 18px;
  font-weight: 600;
  white-space: nowrap;
}

.toolbar-button--primary {
  min-width: 132px;
}

.editable-cell {
  min-height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2px 0;
  cursor: text;
  min-width: 0;
  text-align: center;
}

.editable-cell--editing {
  padding: 0;
}

.editable-cell--changed {
  color: var(--devops-warning);
  font-weight: 600;
  background: rgba(230, 162, 60, 0.12);
  border-radius: 8px;
  padding-inline: 8px;
}

.cell-text {
  display: inline-block;
  width: 100%;
  min-width: 0;
  white-space: nowrap;
  text-align: center;
}

.cell-text--datetime {
  font-variant-numeric: tabular-nums;
}

.cell-text--wide {
  width: auto;
  min-width: max-content;
  white-space: nowrap;
}

.table-empty-state {
  padding: 18px 0 26px;
}

:deep(.filter-input .el-input__wrapper) {
  min-height: 40px;
  padding: 0 14px;
}

:deep(.filter-input .el-input__inner) {
  font-size: 13px;
}

:deep(.filter-input .el-input__inner::placeholder) {
  color: var(--devops-text-placeholder);
}

:deep(.toolbar-button-group .el-dropdown) {
  display: inline-flex;
}

:deep(.anime-data-table .el-table__cell) {
  text-align: center;
}

:deep(.anime-data-table .cell) {
  text-align: center;
}

:deep(.anime-data-table .el-input__wrapper) {
  justify-content: center;
}

:deep(.anime-data-table .el-input__inner) {
  text-align: center;
}

:deep(.anime-data-table .column--profession .cell) {
  overflow: visible;
  white-space: nowrap;
  text-overflow: clip;
  padding-right: 18px;
  letter-spacing: 0 !important;
  word-spacing: 0 !important;
  font-kerning: none;
  text-rendering: optimizeLegibility;
  font-family: "Microsoft YaHei", "PingFang SC", "Hiragino Sans GB", "Noto Sans SC", sans-serif;
}

:deep(.anime-data-table .column--profession .cell .cell-text) {
  letter-spacing: 0 !important;
  word-spacing: 0 !important;
  white-space: nowrap;
  font-family: "Microsoft YaHei", "PingFang SC", "Hiragino Sans GB", "Noto Sans SC", sans-serif;
}

:deep(.anime-data-table .column--wide-text .cell) {
  overflow: visible;
  white-space: nowrap;
  text-overflow: clip;
  padding-right: 22px;
}

:deep(.anime-data-table .column--datetime .cell) {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: clip;
}

:deep(.anime-data-table .el-table__body td.el-table-fixed-column--right),
:deep(.anime-data-table .el-table__header th.el-table-fixed-column--right) {
  position: relative;
  z-index: 3;
  background: rgba(250, 248, 242, 0.98) !important;
}

:deep(.anime-data-table .el-table__fixed-right) {
  box-shadow: -18px 0 24px rgba(58, 67, 88, 0.12);
}

:deep(.anime-data-table .row--inserted > td.el-table__cell) {
  background: rgba(103, 194, 58, 0.08) !important;
}

:deep(.anime-data-table .row--updated > td.el-table__cell) {
  background: rgba(230, 162, 60, 0.08) !important;
}

:deep(.anime-data-table .row--deleted > td.el-table__cell) {
  background: rgba(245, 108, 108, 0.08) !important;
  color: var(--devops-text-secondary);
  text-decoration: line-through;
}

:deep(.anime-data-table .row--deleted .el-button) {
  text-decoration: none;
}

@media (max-width: 1200px) {
  .panel-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-actions,
  .toolbar-button-group {
    justify-content: flex-start;
  }

  .filter-row {
    flex-direction: column;
    align-items: stretch;
    padding: 16px;
  }

  .filter-field {
    width: 100%;
    max-width: none;
    flex-basis: 100%;
  }

  .filter-row__controls {
    width: 100%;
    justify-content: flex-start;
  }

  .filter-row__save {
    align-self: auto;
  }

  .change-summary {
    gap: 10px;
  }
}

@media (max-width: 960px) {
  .filter-row {
    gap: 12px;
  }

  .filter-row__controls {
    width: 100%;
  }

  .table-panel {
    min-height: 460px;
  }
}

@media (max-width: 720px) {
  .toolbar-heading h3 {
    max-width: 100%;
  }

  .summary-pill,
  .summary-pill--wide {
    max-width: 100%;
  }

  .filter-field {
    min-height: 48px;
  }

  .change-summary {
    padding: 12px;
  }

  .table-panel {
    min-height: 380px;
  }
}

@media (max-width: 480px) {
  .panel-toolbar,
  .query-tab {
    gap: 12px;
  }

  .summary-pill {
    width: 100%;
  }

  .table-panel {
    min-height: 320px;
    padding: 8px;
  }
}
</style>

