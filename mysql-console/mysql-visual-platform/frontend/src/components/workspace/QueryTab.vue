<template>
  <section class="query-tab">
    <div class="panel-toolbar">
      <div>
        <span class="header-kicker">{{ t('query.title') }}</span>
        <h3>{{ title }}</h3>
      </div>

      <div class="header-actions">
        <el-button class="soft-button" @click="resetSQL">{{ t('query.clear') }}</el-button>
        <el-button
          type="primary"
          class="soft-button primary-button"
          :loading="running"
          @click="executeQuery"
        >
          {{ t('query.run') }}
        </el-button>
      </div>
    </div>

    <div class="glass-subpanel query-editor-panel">
      <div class="query-context">
        <span>{{ t('query.database') }}: {{ databaseName || t('common.notSelected') }}</span>
        <span>{{ t('query.executionHint') }}</span>
      </div>

      <div class="query-editor-shell">
        <div class="statement-gutter" :aria-label="t('query.gutterAria')">
          <div class="statement-gutter__inner" :style="{ transform: `translateY(${-editorScrollTop}px)` }">
            <button
              v-for="(statement, index) in statementRanges"
              :key="`${statement.start}-${statement.end}`"
              type="button"
              class="statement-marker"
              :class="{ active: activeStatementIndex === index }"
              :style="{ top: `${statement.topOffset}px` }"
              :title="t('query.runLine', { line: statement.startLine })"
              @click="executeStatementAtIndex(index)"
            >
              <span class="statement-marker__line">{{ statement.startLine }}</span>
            </button>
          </div>
        </div>

        <el-input
          ref="editorRef"
          v-model="sql"
          class="query-editor-input"
          type="textarea"
          :autosize="{ minRows: 8, maxRows: 16 }"
          resize="none"
          :placeholder="t('query.placeholder')"
          @click="syncActiveStatement"
          @focus="syncActiveStatement"
          @keydown="handleEditorKeydown"
          @keyup="syncActiveStatement"
          @mouseup="syncActiveStatement"
        />
      </div>
    </div>

    <div v-loading="running" class="glass-subpanel query-result-panel">
      <template v-if="resultMode === 'table'">
        <template v-if="resultRows.length > 0">
          <el-table :data="resultRows" height="100%" class="anime-data-table">
            <el-table-column
              v-for="column in resultColumns"
              :key="column"
              :prop="column"
              :label="column"
              min-width="160"
              show-overflow-tooltip
            />
          </el-table>
        </template>
        <el-empty
          v-else
          :description="t('query.noRows')"
        />
      </template>

      <el-empty
        v-else
        :description="t('query.idle')"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

import { useI18n } from '@/utils/i18n'
import request from '@/utils/request'

interface QueryTableRow {
  [key: string]: unknown
}

interface QueryTableResponse {
  columns?: string[]
  rows?: QueryTableRow[]
}

interface ExecuteResultResponse {
  rowsAffected: number
}

interface StatementRange {
  text: string
  start: number
  end: number
  startLine: number
  endLine: number
  topOffset: number
}

const props = defineProps<{
  title: string
  runSignal: number
  active: boolean
  databaseName: string
  tableName: string
  initialSql?: string
}>()

const editorRef = ref()
const { isChinese, t } = useI18n()
const defaultSql = 'SHOW DATABASES;\nSELECT DATABASE();'
const sql = ref(props.initialSql?.trim() || defaultSql)
const running = ref(false)
const resultMode = ref<'idle' | 'table'>('idle')
const resultColumns = ref<string[]>([])
const resultRows = ref<QueryTableRow[]>([])
const rowsAffected = ref(0)
const activeStatementIndex = ref(0)
const editorScrollTop = ref(0)

const statementRanges = computed(() => splitSQLStatements(sql.value))

async function executeQuery() {
  const statement = resolveExecutableStatement()
  if (!statement) {
    ElMessage.warning(t('query.selectStatement'))
    return
  }

  await executeStatement(statement)
}

async function executeStatementAtIndex(index: number) {
  const statement = statementRanges.value[index]
  if (!statement) {
    return
  }

  focusStatement(statement, index)
  await executeStatement(statement.text)
}

async function executeStatement(statement: string) {
  const normalizedStatement = trimStatementDelimiter(statement)
  if (!normalizedStatement) {
    ElMessage.warning(t('query.invalidStatement'))
    return
  }

  const statementKind = getStatementKind(normalizedStatement)
  running.value = true

  try {
    const result = await request.post<QueryTableResponse | ExecuteResultResponse>('/api/query/execute', {
      sql: normalizedStatement,
      database: props.databaseName
    }, {
      silentError: true
    })

    if ('rowsAffected' in result) {
      resultMode.value = 'idle'
      rowsAffected.value = result.rowsAffected
      resultColumns.value = []
      resultRows.value = []
      ElMessage.success(formatOperationSuccessMessage(statementKind, result.rowsAffected))
      return
    }

    resultMode.value = 'table'
    resultRows.value = result.rows ?? []
    resultColumns.value = result.columns?.length ? result.columns : Object.keys(resultRows.value[0] ?? {})
    ElMessage.success(formatQuerySuccessMessage())
  } catch (error) {
    ElMessage.error(formatOperationFailureMessage(statementKind, error))
  } finally {
    running.value = false
  }
}

function getStatementKind(statement: string) {
  const normalized = statement.trim().replace(/^[;(]+/, '').trim().toUpperCase()
  const keyword = normalized.split(/\s+/)[0] || ''

  if (['SELECT', 'SHOW', 'DESC', 'DESCRIBE', 'EXPLAIN', 'WITH'].includes(keyword)) {
    return 'query'
  }

  if (keyword === 'INSERT') {
    return 'insert'
  }

  if (keyword === 'UPDATE') {
    return 'update'
  }

  if (keyword === 'DELETE') {
    return 'delete'
  }

  if (keyword === 'CREATE') {
    return 'create'
  }

  if (keyword === 'ALTER') {
    return 'alter'
  }

  if (keyword === 'DROP') {
    return 'drop'
  }

  return 'execute'
}

function getOperationLabel(statementKind: string) {
  if (isChinese.value) {
    switch (statementKind) {
      case 'query':
        return '查询'
      case 'insert':
        return '新增'
      case 'update':
        return '更新'
      case 'delete':
        return '删除'
      case 'create':
        return '创建'
      case 'alter':
        return '修改'
      case 'drop':
        return '删除'
      default:
        return '执行'
    }
  }

  switch (statementKind) {
    case 'query':
      return 'Query'
    case 'insert':
      return 'Insert'
    case 'update':
      return 'Update'
    case 'delete':
      return 'Delete'
    case 'create':
      return 'Create'
    case 'alter':
      return 'Alter'
    case 'drop':
      return 'Drop'
    default:
      return 'Execute'
  }
}

function formatQuerySuccessMessage() {
  return isChinese.value ? '查询成功' : 'Query succeeded'
}

function formatOperationSuccessMessage(statementKind: string, affectedRows: number) {
  const label = getOperationLabel(statementKind)
  return isChinese.value
    ? `${label}成功，影响了 ${affectedRows} 行`
    : `${label} succeeded, affected ${affectedRows} rows`
}

function formatOperationFailureMessage(statementKind: string, error: unknown) {
  const label = getOperationLabel(statementKind)
  const reason = extractErrorMessage(error)
  return isChinese.value
    ? `${label}失败：${reason}`
    : `${label} failed: ${reason}`
}

function extractErrorMessage(error: unknown) {
  const fallback = isChinese.value ? '未知错误' : 'Unknown error'

  if (!error) {
    return fallback
  }

  if (typeof error === 'string') {
    return error.trim() || fallback
  }

  if (typeof error === 'object') {
    const maybeAxiosError = error as {
      message?: string
      response?: {
        data?: {
          msg?: string
          message?: string
          error?: string
        }
        statusText?: string
      }
    }

    const responseMessage =
      maybeAxiosError.response?.data?.msg ||
      maybeAxiosError.response?.data?.message ||
      maybeAxiosError.response?.data?.error ||
      maybeAxiosError.response?.statusText

    if (typeof responseMessage === 'string' && responseMessage.trim()) {
      return responseMessage.trim()
    }

    if (typeof maybeAxiosError.message === 'string' && maybeAxiosError.message.trim()) {
      return maybeAxiosError.message.trim()
    }
  }

  if (error instanceof Error && error.message.trim()) {
    return error.message.trim()
  }

  return fallback
}

function resolveExecutableStatement() {
  const textarea = getNativeTextarea()
  if (!textarea) {
    return sql.value.trim()
  }

  const selectionStart = textarea.selectionStart ?? 0
  const selectionEnd = textarea.selectionEnd ?? 0
  const selectedText = sql.value.slice(selectionStart, selectionEnd).trim()
  if (selectedText) {
    return trimStatementDelimiter(selectedText)
  }

  const statement = findStatementAtCursor(sql.value, selectionStart)
  return statement ? trimStatementDelimiter(statement.text) : ''
}

function getNativeTextarea() {
  const instance = editorRef.value as { textarea?: HTMLTextAreaElement; $el?: HTMLElement } | undefined
  return instance?.textarea ?? instance?.$el?.querySelector('textarea') ?? null
}

function handleEditorScroll() {
  const textarea = getNativeTextarea()
  editorScrollTop.value = textarea?.scrollTop ?? 0
}

function findStatementAtCursor(source: string, cursor: number) {
  const statements = splitSQLStatements(source)
  for (const statement of statements) {
    if (cursor >= statement.start && cursor <= statement.end) {
      return statement
    }
  }

  return statements.length > 0 ? statements[statements.length - 1] : null
}

function splitSQLStatements(source: string) {
  const statements: StatementRange[] = []
  let start = 0
  let quote: "'" | '"' | '`' | null = null
  let lineComment = false
  let blockComment = false

  for (let index = 0; index < source.length; index += 1) {
    const current = source[index]
    const next = source[index + 1]

    if (lineComment) {
      if (current === '\n') {
        lineComment = false
      }
      continue
    }

    if (blockComment) {
      if (current === '*' && next === '/') {
        blockComment = false
        index += 1
      }
      continue
    }

    if (quote) {
      if (current === '\\' && quote !== '`') {
        index += 1
        continue
      }

      if (current === quote) {
        quote = null
      }
      continue
    }

    if (current === '-' && next === '-') {
      lineComment = true
      index += 1
      continue
    }

    if (current === '#') {
      lineComment = true
      continue
    }

    if (current === '/' && next === '*') {
      blockComment = true
      index += 1
      continue
    }

    if (current === '\'' || current === '"' || current === '`') {
      quote = current
      continue
    }

    if (current === ';') {
      pushStatementRange(statements, source, start, index + 1)
      start = index + 1
    }
  }

  pushStatementRange(statements, source, start, source.length)
  return statements
}

function pushStatementRange(statements: StatementRange[], source: string, start: number, end: number) {
  const rawSegment = source.slice(start, end)
  const text = rawSegment.trim()
  if (!text) {
    return
  }

  const leadingWhitespaceLength = rawSegment.match(/^\s*/)?.[0].length ?? 0
  const trailingWhitespaceLength = rawSegment.match(/\s*$/)?.[0].length ?? 0
  const contentStart = start + leadingWhitespaceLength
  const contentEnd = Math.max(contentStart, end - trailingWhitespaceLength)
  const startLine = getLineNumber(source, contentStart)
  const endLine = getLineNumber(source, contentEnd)

  statements.push({
    text,
    start: contentStart,
    end: contentEnd,
    startLine,
    endLine,
    topOffset: 12 + (startLine - 1) * 24
  })
}

function getLineNumber(source: string, position: number) {
  let line = 1

  for (let index = 0; index < position; index += 1) {
    if (source[index] === '\n') {
      line += 1
    }
  }

  return line
}

function trimStatementDelimiter(statement: string) {
  return statement.replace(/;\s*$/, '').trim()
}

function resetSQL() {
  sql.value = ''
  activeStatementIndex.value = 0
  editorScrollTop.value = 0
  resultMode.value = 'idle'
  resultColumns.value = []
  resultRows.value = []
  rowsAffected.value = 0
}

function syncActiveStatement() {
  const textarea = getNativeTextarea()
  if (!textarea) {
    activeStatementIndex.value = 0
    return
  }

  const selectionStart = textarea.selectionStart ?? 0
  const statements = statementRanges.value
  const statementIndex = statements.findIndex((statement) => selectionStart >= statement.start && selectionStart <= statement.end)
  activeStatementIndex.value = statementIndex >= 0 ? statementIndex : 0
}

function focusStatement(statement: StatementRange, index: number) {
  activeStatementIndex.value = index

  const textarea = getNativeTextarea()
  if (!textarea) {
    return
  }

  textarea.focus()
  textarea.setSelectionRange(statement.start, statement.end)
}

function handleKeydown(event: KeyboardEvent) {
  if (props.active && event.key === 'F8') {
    event.preventDefault()
    void executeQuery()
  }
}

function handleEditorKeydown(event: KeyboardEvent) {
  const isModifierPressed = event.ctrlKey || event.metaKey
  if (!isModifierPressed) {
    return
  }

  const key = event.key.toLowerCase()
  const isUndo = key === 'z' && !event.shiftKey
  const isRedo = key === 'y' || (key === 'z' && event.shiftKey)

  if (isUndo || isRedo) {
    // Preserve the browser/native textarea history behavior,
    // but stop the shortcut from bubbling into the host shell.
    event.stopPropagation()
  }
}

watch(
  () => props.runSignal,
  (value, oldValue) => {
    if (value !== oldValue && value > 0) {
      void executeQuery()
    }
  }
)

defineExpose({
  executeQuery
})

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
  void nextTick(() => {
    syncActiveStatement()
    handleEditorScroll()
    getNativeTextarea()?.addEventListener('scroll', handleEditorScroll, { passive: true })
  })
  if (props.runSignal > 0 && props.initialSql?.trim()) {
    void executeQuery()
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown)
  getNativeTextarea()?.removeEventListener('scroll', handleEditorScroll)
})

watch(
  () => props.initialSql,
  (value, oldValue) => {
    if (value && value !== oldValue && value !== sql.value) {
      sql.value = value
      void nextTick(() => {
        syncActiveStatement()
      })
    }
  }
)

watch(sql, () => {
  void nextTick(() => {
    syncActiveStatement()
    handleEditorScroll()
  })
})
</script>

<style scoped>
.header-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.panel-toolbar {
  padding: 18px 20px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-lg);
  background: var(--devops-bg-panel);
  box-shadow: var(--devops-shadow-xs);
}

.panel-toolbar h3 {
  font-size: 20px;
}

.query-context {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  font-size: 13px;
  color: var(--devops-text-secondary);
}

.query-editor-shell {
  display: flex;
  align-items: stretch;
  gap: 0;
  border: 1px solid var(--devops-border);
  border-radius: var(--devops-radius-lg);
  overflow: hidden;
  background: var(--devops-bg-panel);
}

.statement-gutter {
  position: relative;
  width: 72px;
  flex-shrink: 0;
  overflow: hidden;
  background: #1f2937;
  border-right: 1px solid #334155;
}

.statement-gutter__inner {
  position: relative;
  min-height: 100%;
  will-change: transform;
}

.statement-marker {
  position: absolute;
  left: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 46px;
  height: 22px;
  padding: 0 8px;
  border: none;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.22);
  color: #e2e8f0;
  cursor: pointer;
  transition: transform 0.15s ease, background-color 0.15s ease, color 0.15s ease, box-shadow 0.15s ease;
}

.statement-marker:hover,
.statement-marker.active {
  background: #409eff;
  color: #ffffff;
  transform: translateX(4px);
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.18);
}

.statement-marker__line {
  font-size: 12px;
  font-weight: 800;
  line-height: 1;
}

.query-editor-input {
  flex: 1;
  min-width: 0;
}

:deep(.query-editor-input .el-textarea__inner) {
  border: none;
  border-radius: 0;
  box-shadow: none;
  background: transparent;
  min-height: 260px !important;
  padding: 12px 14px;
  font-family: "JetBrains Mono", Consolas, "Courier New", monospace;
  font-size: 14px;
  line-height: 24px;
}

@media (max-width: 960px) {
  .query-context {
    flex-direction: column;
    align-items: flex-start;
  }
}

@media (max-width: 720px) {
  .query-editor-shell {
    flex-direction: column;
  }

  .statement-gutter {
    width: 100%;
    min-height: 44px;
    border-right: none;
    border-bottom: 1px solid rgba(148, 163, 184, 0.24);
  }

  .statement-marker {
    position: static;
    margin: 10px 0 10px 12px;
    transform: none !important;
  }

  :deep(.query-editor-input .el-textarea__inner) {
    min-height: 220px !important;
    padding: 12px;
    font-size: 13px;
  }
}

@media (max-width: 480px) {
  .query-context {
    font-size: 12px;
  }

  .statement-gutter {
    min-height: 40px;
  }

  .statement-marker {
    min-width: 42px;
    height: 20px;
    margin-left: 10px;
  }
}
</style>
