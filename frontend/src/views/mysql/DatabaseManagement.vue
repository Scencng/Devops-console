<template>
  <div class="page-container mysql-page">
    <MySQLPageHeader
      title="数据库管理"
      description="仅处理数据库、表、视图和函数的结构管理，右侧展示结构详情，不混入表数据编辑"
      :metrics="headerMetrics"
    >
      <template #actions>
        <el-button class="soft-button" @click="refreshExplorer">刷新资源树</el-button>
        <el-button
          v-if="canCreateDatabase"
          type="primary"
          class="soft-button"
          @click="dialogs.createDatabaseVisible = true"
        >
          新建数据库
        </el-button>
        <el-button
          v-if="canCreateTable"
          class="soft-button"
          :disabled="!selectedDatabaseName"
          @click="openCreateTableDialog"
        >
          新建数据表
        </el-button>
        <el-button
          v-if="canImportData"
          class="soft-button"
          :disabled="!selectedDatabaseName"
          :loading="smartImporting"
          @click="openSmartImportDialog"
        >
          智能导入
        </el-button>
        <el-button
          v-if="canExportData"
          class="soft-button"
          :disabled="!canSmartExport"
          :loading="smartExporting"
          @click="handleSmartExport"
        >
          智能导出
        </el-button>
      </template>
    </MySQLPageHeader>

    <div class="mysql-page__body">
      <MySQLPageSkeleton v-show="pageLoading" layout="split" :cards="2" />

      <div v-show="!pageLoading" class="mysql-page__grid">
      <MySQLExplorerTree
        ref="explorerRef"
        title="数据库资源树"
        description="点击数据库、表、视图或函数后，在右侧查看对应结构详情"
        @select="handleNodeSelect"
        @loaded="handleExplorerLoaded"
      />

      <div class="mysql-page__main">
        <el-card class="content-card" shadow="never">
          <template #header>
            <div class="toolbar-row">
              <div class="toolbar-left">
                <strong>结构详情</strong>
                <el-tag v-if="selectedNode" effect="light">{{ nodeTypeLabel }}</el-tag>
              </div>
              <div class="toolbar-right structure-toolbar">
                <el-button
                  v-if="selectedNode?.type === 'table'"
                  class="soft-button"
                  @click="openDataPage"
                >
                  打开数据
                </el-button>
                <el-button
                  v-if="selectedNode && canCompareStructure"
                  class="soft-button"
                  :disabled="!canOpenCompare"
                  @click="openCompareDialog"
                >
                  结构对比
                </el-button>
                <el-button
                  v-if="selectedNode && canRenameSelected"
                  class="soft-button"
                  @click="renameSelected"
                >
                  重命名
                </el-button>
                <el-button
                  v-if="selectedNode && canDeleteSelected"
                  type="danger"
                  class="soft-button"
                  @click="deleteSelected"
                >
                  删除
                </el-button>
              </div>
            </div>
          </template>

          <el-empty v-if="!selectedNode" description="请从左侧资源树选择一个数据库对象" />

          <template v-else>
            <div class="structure-summary-grid">
              <div class="structure-summary-card">
                <span>对象名称</span>
                <strong>{{ selectedNode.label }}</strong>
              </div>
              <div class="structure-summary-card">
                <span>所属数据库</span>
                <strong>{{ selectedNode.databaseName || '-' }}</strong>
              </div>
              <div class="structure-summary-card">
                <span>对象类型</span>
                <strong>{{ nodeTypeLabel }}</strong>
              </div>
            </div>

            <div v-loading="detailLoading" class="structure-content">
              <template v-if="selectedNode.type === 'database'">
                <el-descriptions :column="2" border class="structure-descriptions">
                  <el-descriptions-item label="数据库">{{ selectedNode.databaseName }}</el-descriptions-item>
                  <el-descriptions-item label="对象总数">{{ databaseObjectCount }}</el-descriptions-item>
                  <el-descriptions-item label="数据表">{{ currentDatabaseStats.tables }}</el-descriptions-item>
                  <el-descriptions-item label="视图">{{ currentDatabaseStats.views }}</el-descriptions-item>
                  <el-descriptions-item label="函数">{{ currentDatabaseStats.functions }}</el-descriptions-item>
                </el-descriptions>

                <div class="structure-list-grid">
                  <section class="structure-list-card">
                    <div class="structure-list-card__header">
                      <strong>数据表</strong>
                      <span>{{ currentDatabaseObjects.tables.length }}</span>
                    </div>
                    <el-empty v-if="currentDatabaseObjects.tables.length === 0" description="暂无数据表" />
                    <div v-else class="structure-chip-list">
                      <button
                        v-for="item in currentDatabaseObjects.tables"
                        :key="`table-${item}`"
                        type="button"
                        class="structure-chip"
                        @click="selectObjectNode('table', item)"
                      >
                        {{ item }}
                      </button>
                    </div>
                  </section>

                  <section class="structure-list-card">
                    <div class="structure-list-card__header">
                      <strong>视图</strong>
                      <span>{{ currentDatabaseObjects.views.length }}</span>
                    </div>
                    <el-empty v-if="currentDatabaseObjects.views.length === 0" description="暂无视图" />
                    <div v-else class="structure-chip-list">
                      <button
                        v-for="item in currentDatabaseObjects.views"
                        :key="`view-${item}`"
                        type="button"
                        class="structure-chip"
                        @click="selectObjectNode('view', item)"
                      >
                        {{ item }}
                      </button>
                    </div>
                  </section>

                  <section class="structure-list-card">
                    <div class="structure-list-card__header">
                      <strong>函数</strong>
                      <span>{{ currentDatabaseObjects.functions.length }}</span>
                    </div>
                    <el-empty v-if="currentDatabaseObjects.functions.length === 0" description="暂无函数" />
                    <div v-else class="structure-chip-list">
                      <button
                        v-for="item in currentDatabaseObjects.functions"
                        :key="`function-${item}`"
                        type="button"
                        class="structure-chip"
                        @click="selectObjectNode('function', item)"
                      >
                        {{ item }}
                      </button>
                    </div>
                  </section>
                </div>
              </template>

              <template v-else-if="selectedNode.type === 'group'">
                <el-empty :description="`${groupLabel}分组下共有 ${selectedNode.children?.length ?? 0} 个对象，请继续选择具体对象查看结构详情`" />
              </template>

              <template v-else-if="selectedNode.type === 'table'">
                <section class="structure-section">
                  <div class="structure-section__header">
                    <strong>字段结构</strong>
                    <span>{{ tableColumns.length }} 个字段</span>
                  </div>
                  <div class="structure-table-wrap">
                    <el-table :data="tableColumns" border class="structure-table">
                      <el-table-column prop="field" label="字段名" min-width="180" />
                      <el-table-column prop="type" label="类型" min-width="260" />
                      <el-table-column prop="nullable" label="允许空值" min-width="120" />
                      <el-table-column prop="key" label="键" min-width="120" />
                      <el-table-column prop="defaultValue" label="默认值" min-width="160" />
                      <el-table-column prop="extra" label="额外属性" min-width="180" />
                      <el-table-column prop="comment" label="注释" min-width="220" />
                    </el-table>
                  </div>
                </section>

                <section class="structure-section">
                  <div class="structure-section__header">
                    <strong>索引信息</strong>
                    <span>{{ tableIndexes.length }} 条</span>
                  </div>
                  <div class="structure-table-wrap">
                    <el-table :data="tableIndexes" border class="structure-table">
                      <el-table-column prop="name" label="索引名" min-width="220" />
                      <el-table-column prop="column" label="字段" min-width="220" />
                      <el-table-column prop="type" label="索引类型" min-width="140" />
                      <el-table-column prop="nonUnique" label="唯一性" min-width="120" />
                      <el-table-column prop="sequence" label="顺序" min-width="100" />
                    </el-table>
                  </div>
                </section>

                <section class="structure-section">
                  <div class="structure-section__header">
                    <strong>建表语句</strong>
                    <div class="structure-section__actions">
                      <el-button class="soft-button" @click="copyDefinition" :disabled="!definitionText">
                        复制语句
                      </el-button>
                      <el-button class="soft-button" :loading="detailLoading" @click="reloadCurrentDetails">
                        刷新定义
                      </el-button>
                    </div>
                  </div>
                  <el-input
                    :model-value="definitionText || '暂无结构定义'"
                    type="textarea"
                    readonly
                    :autosize="{ minRows: 10, maxRows: 20 }"
                    class="structure-definition-input"
                  />
                </section>
              </template>

              <template v-else>
                <section class="structure-section">
                  <div class="structure-section__header">
                    <strong>{{ selectedNode.type === 'view' ? '视图定义' : '函数定义' }}</strong>
                    <div class="structure-section__actions">
                      <el-button class="soft-button" @click="copyDefinition" :disabled="!definitionText">
                        复制语句
                      </el-button>
                      <el-button class="soft-button" :loading="detailLoading" @click="reloadCurrentDetails">
                        刷新定义
                      </el-button>
                    </div>
                  </div>
                  <el-input
                    :model-value="definitionText || '暂无结构定义'"
                    type="textarea"
                    readonly
                    :autosize="{ minRows: 10, maxRows: 20 }"
                    class="structure-definition-input"
                  />
                </section>
              </template>
            </div>
          </template>
        </el-card>
      </div>
    </div>
    </div>

    <StructureCompareDialog
      v-model:visible="dialogs.compareVisible"
      :scope="compareScope"
      :source-database="selectedNode?.databaseName || ''"
      :source-table="selectedNode?.tableName || ''"
      @refresh-explorer="refreshExplorer"
    />

    <el-dialog v-model="dialogs.createDatabaseVisible" title="新建数据库" width="420px">
      <el-form label-position="top">
        <el-form-item label="数据库名称">
          <el-input v-model="forms.createDatabaseName" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogs.createDatabaseVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCreateDatabase">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="dialogs.createTableVisible" title="新建数据表" width="720px">
      <el-form label-position="top">
        <el-form-item label="数据库">
          <el-input v-model="forms.createTableDatabase" />
        </el-form-item>
        <el-form-item label="表名">
          <el-input v-model="forms.createTableName" />
        </el-form-item>
        <el-form-item label="字段定义">
          <el-input
            v-model="forms.createTableColumns"
            type="textarea"
            :autosize="{ minRows: 8, maxRows: 16 }"
            placeholder="id BIGINT PRIMARY KEY AUTO_INCREMENT,&#10;name VARCHAR(128) NOT NULL"
          />
        </el-form-item>
        <el-form-item label="表选项">
          <el-input v-model="forms.createTableOptions" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogs.createTableVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCreateTable">创建</el-button>
      </template>
    </el-dialog>

    <input
      ref="smartImportInputRef"
      class="smart-import-input"
      type="file"
      accept=".xlsx,.xls,.csv,text/csv,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
      @change="handleSmartImportFileChange"
    >
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'

import MySQLExplorerTree from '@/mysql/components/shared/MySQLExplorerTree.vue'
import MySQLPageHeader from '@/mysql/components/shared/MySQLPageHeader.vue'
import MySQLPageSkeleton from '@/mysql/components/shared/MySQLPageSkeleton.vue'
import StructureCompareDialog from '@/mysql/components/workspace/StructureCompareDialog.vue'
import { useConnectionStore } from '@/mysql/stores/connection'
import { useWorkspaceStore } from '@/mysql/stores/workspace'
import {
  ensureDatabaseChildrenLoaded,
  type ExplorerNodeType,
  findNodeByKey,
  type TreeNodeData
} from '@/mysql/utils/explorer'
import {
  normalizeImportedCellValue,
  normalizeImportedColumnName,
  readImportMatrix,
  sanitizeImportedIdentifier
} from '@/mysql/utils/import-file'
import request, { ensureConnectionReady } from '@/mysql/utils/request'
import { downloadExcel } from '@/mysql/utils/table-export'
import { usePermissionStore } from '@/stores/permissionStore.js'

interface QueryRowsResponse {
  rows?: Array<Record<string, unknown>>
}

interface TableColumnInfo {
  field: string
  type: string
  nullable: string
  key: string
  defaultValue: string
  extra: string
  comment: string
}

interface TableIndexInfo {
  name: string
  column: string
  type: string
  nonUnique: string
  sequence: string
}

interface SmartImportColumn {
  name: string
  type: string
}

interface SmartImportPayload {
  tableName: string
  columns: SmartImportColumn[]
  rows: Array<Record<string, unknown>>
}

interface AutoImportResponse {
  success: boolean
  tableName: string
  rowCount: number
}

const route = useRoute()
const router = useRouter()
const connectionStore = useConnectionStore()
const workspaceStore = useWorkspaceStore()
const permissionStore = usePermissionStore()
const explorerRef = ref<InstanceType<typeof MySQLExplorerTree> | null>(null)
const pageLoading = ref(true)
const selectedNode = ref<TreeNodeData | null>(null)
const treeNodes = ref<TreeNodeData[]>([])
const detailLoading = ref(false)
const definitionText = ref('')
const tableColumns = ref<TableColumnInfo[]>([])
const tableIndexes = ref<TableIndexInfo[]>([])
const syncingRoute = ref(false)
const smartImportInputRef = ref<HTMLInputElement | null>(null)
const smartImporting = ref(false)
const smartExporting = ref(false)

const dialogs = reactive({
  compareVisible: false,
  createDatabaseVisible: false,
  createTableVisible: false
})

const forms = reactive({
  createDatabaseName: '',
  createTableDatabase: '',
  createTableName: '',
  createTableColumns: 'id BIGINT PRIMARY KEY AUTO_INCREMENT',
  createTableOptions: 'ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci'
})

const headerMetrics = computed(() => [
  { label: '连接状态', value: connectionStore.hasConnection ? '已连接' : '未连接' },
  { label: '当前数据库', value: selectedDatabaseName.value || connectionStore.profile.database || '-' },
  { label: '当前对象', value: selectedNode.value?.label || '-' }
])

const selectedDatabaseName = computed(() => {
  if (selectedNode.value?.databaseName) {
    return selectedNode.value.databaseName
  }
  return workspaceStore.activeDatabase || connectionStore.profile.database || ''
})

const nodeTypeLabel = computed(() => {
  if (!selectedNode.value) return ''

  const nodeTypeMap: Record<string, string> = {
    database: '数据库',
    group: groupLabel.value,
    table: '数据表',
    view: '视图',
    function: '函数',
    backup: '备份'
  }
  return nodeTypeMap[selectedNode.value.type] || selectedNode.value.type
})

const groupLabel = computed(() => {
  const groupKind = selectedNode.value?.groupKind
  const labels: Record<string, string> = {
    tables: '数据表',
    views: '视图',
    functions: '函数',
    queries: '查询',
    backups: '备份'
  }
  return groupKind ? labels[groupKind] || '分组' : '分组'
})

const currentDatabaseNode = computed(() => {
  const databaseName = selectedDatabaseName.value
  if (!databaseName) {
    return null
  }

  return treeNodes.value.find((item) => item.databaseName === databaseName) || null
})

const currentDatabaseStats = computed(() => {
  const groups = currentDatabaseNode.value?.children ?? []
  return {
    tables: groups.find((item) => item.groupKind === 'tables')?.children?.length ?? 0,
    views: groups.find((item) => item.groupKind === 'views')?.children?.length ?? 0,
    functions: groups.find((item) => item.groupKind === 'functions')?.children?.length ?? 0
  }
})

const currentDatabaseObjects = computed(() => {
  const groups = currentDatabaseNode.value?.children ?? []
  return {
    tables: (groups.find((item) => item.groupKind === 'tables')?.children ?? []).map((item) => item.label),
    views: (groups.find((item) => item.groupKind === 'views')?.children ?? []).map((item) => item.label),
    functions: (groups.find((item) => item.groupKind === 'functions')?.children ?? []).map((item) => item.label)
  }
})

const databaseObjectCount = computed(
  () => currentDatabaseStats.value.tables + currentDatabaseStats.value.views + currentDatabaseStats.value.functions
)

const compareScope = computed<'database' | 'table'>(() => (selectedNode.value?.type === 'database' ? 'database' : 'table'))
const canOpenCompare = computed(() => selectedNode.value?.type === 'database' || selectedNode.value?.type === 'table')
const canCreateDatabase = computed(() => permissionStore.hasPerm('mysql:database:create'))
const canCreateTable = computed(() => permissionStore.hasPerm('mysql:table:create'))
const canImportData = computed(() => permissionStore.hasPerm('mysql:data:import'))
const canExportData = computed(() => permissionStore.hasPerm('mysql:data:export'))
const canCompareStructure = computed(() => true)
const canSmartExport = computed(
  () => canExportData.value && Boolean(selectedDatabaseName.value && activeTableName.value)
)
const canRenameSelected = computed(() => {
  if (!selectedNode.value) return false
  if (selectedNode.value.type === 'database') return permissionStore.hasPerm('mysql:database:rename')
  if (selectedNode.value.type === 'table' || selectedNode.value.type === 'view') return permissionStore.hasPerm('mysql:table:rename')
  return false
})
const canDeleteSelected = computed(() => {
  if (!selectedNode.value) return false
  if (selectedNode.value.type === 'database') return permissionStore.hasPerm('mysql:database:delete')
  if (selectedNode.value.type === 'table' || selectedNode.value.type === 'view') return permissionStore.hasPerm('mysql:table:delete')
  return false
})
const activeTableName = computed(() => {
  if (selectedNode.value?.type === 'table') {
    return selectedNode.value.tableName || selectedNode.value.label
  }
  return workspaceStore.activeTable || ''
})

void ensurePageReady()

watch(
  () => [route.query.db, route.query.type, route.query.object],
  () => {
    if (!syncingRoute.value && treeNodes.value.length > 0) {
      void syncSelectionFromRoute()
    }
  }
)

async function ensurePageReady() {
  if (!connectionStore.hasConnection) {
    await router.push('/mysql/workbench')
    return
  }

  await ensureConnectionReady()
}

async function refreshExplorer() {
  await explorerRef.value?.refresh()
}

function handleExplorerLoaded(nodes: TreeNodeData[]) {
  treeNodes.value = nodes
  void syncSelectionFromRoute().finally(() => {
    pageLoading.value = false
  })
}

function openCreateTableDialog() {
  forms.createTableDatabase = selectedDatabaseName.value
  forms.createTableName = ''
  dialogs.createTableVisible = true
}

async function handleNodeSelect(node: TreeNodeData) {
  await applySelectedNode(node, true)
}

async function applySelectedNode(node: TreeNodeData, syncRouteToPage: boolean) {
  selectedNode.value = node
  explorerRef.value?.selectNode(node.key)
  workspaceStore.setActiveDatabase(node.databaseName)

  if (node.type === 'table') {
    workspaceStore.setActiveTable(node.databaseName, node.tableName || node.label)
  } else if (node.type === 'database' || node.type === 'group') {
    workspaceStore.clearActiveTable()
  }

  if (syncRouteToPage) {
    await updateRouteForNode(node)
  }

  await loadNodeDetails(node)
}

async function loadNodeDetails(node: TreeNodeData) {
  detailLoading.value = false
  definitionText.value = ''
  tableColumns.value = []
  tableIndexes.value = []

  if (node.type !== 'table' && node.type !== 'view' && node.type !== 'function') {
    return
  }

  detailLoading.value = true
  try {
    if (node.type === 'table') {
      const [definitionResult, columnsResult, indexesResult] = await Promise.all([
        fetchDefinition(node),
        fetchTableColumns(node),
        fetchTableIndexes(node)
      ])

      definitionText.value = definitionResult
      tableColumns.value = columnsResult
      tableIndexes.value = indexesResult
      return
    }

    definitionText.value = await fetchDefinition(node)
  } finally {
    detailLoading.value = false
  }
}

async function fetchDefinition(node: TreeNodeData) {
  const sql = resolveDefinitionSQL(node)
  const result = await request.post<QueryRowsResponse>('/api/query/execute', {
    sql,
    database: node.databaseName
  }, {
    silentError: true
  })

  const firstRow = result.rows?.[0] ?? {}
  const createKey = Object.keys(firstRow).find((key) => /^create\s/i.test(key))
  if (createKey && firstRow[createKey]) {
    return String(firstRow[createKey])
  }

  const values = Object.values(firstRow).filter(Boolean)
  return values.length ? String(values[values.length - 1]) : '未返回结构定义'
}

async function fetchTableColumns(node: TreeNodeData) {
  const result = await request.post<QueryRowsResponse>('/api/query/execute', {
    sql: `SHOW FULL COLUMNS FROM ${quoteIdentifier(node.tableName || node.label)} FROM ${quoteIdentifier(node.databaseName)};`,
    database: node.databaseName
  }, {
    silentError: true
  })

  return (result.rows ?? []).map((row) => ({
    field: String(row.Field ?? row.field ?? ''),
    type: String(row.Type ?? row.type ?? ''),
    nullable: String(row.Null ?? row.null ?? ''),
    key: String(row.Key ?? row.key ?? ''),
    defaultValue: formatNullableValue(row.Default ?? row.default),
    extra: String(row.Extra ?? row.extra ?? '-'),
    comment: String(row.Comment ?? row.comment ?? '-')
  }))
}

async function fetchTableIndexes(node: TreeNodeData) {
  const result = await request.post<QueryRowsResponse>('/api/query/execute', {
    sql: `SHOW INDEX FROM ${quoteIdentifier(node.tableName || node.label)} FROM ${quoteIdentifier(node.databaseName)};`,
    database: node.databaseName
  }, {
    silentError: true
  })

  return (result.rows ?? []).map((row) => ({
    name: String(row.Key_name ?? row.key_name ?? '-'),
    column: String(row.Column_name ?? row.column_name ?? '-'),
    type: String(row.Index_type ?? row.index_type ?? '-'),
    nonUnique: Number(row.Non_unique ?? row.non_unique ?? 1) === 0 ? '唯一' : '普通',
    sequence: String(row.Seq_in_index ?? row.seq_in_index ?? '-')
  }))
}

function resolveDefinitionSQL(node: TreeNodeData) {
  if (node.type === 'function') {
    return `SHOW CREATE FUNCTION ${quoteIdentifier(node.databaseName)}.${quoteIdentifier(node.tableName || node.label)};`
  }

  if (node.type === 'view') {
    return `SHOW CREATE VIEW ${quoteIdentifier(node.databaseName)}.${quoteIdentifier(node.tableName || node.label)};`
  }

  return `SHOW CREATE TABLE ${quoteIdentifier(node.databaseName)}.${quoteIdentifier(node.tableName || node.label)};`
}

function quoteIdentifier(identifier: string) {
  return `\`${identifier.split('`').join('``')}\``
}

function formatNullableValue(value: unknown) {
  if (value === null || value === undefined || value === '') {
    return '-'
  }
  return String(value)
}

async function syncSelectionFromRoute() {
  const requestedDatabase = firstQueryValue(route.query.db)
  const requestedType = (firstQueryValue(route.query.type) as ExplorerNodeType | 'group' | '') || 'database'
  const requestedObject = firstQueryValue(route.query.object)

  const databaseNode =
    treeNodes.value.find((item) => item.databaseName === requestedDatabase) ||
    treeNodes.value.find((item) => item.databaseName === workspaceStore.activeDatabase) ||
    treeNodes.value[0]

  if (!databaseNode) {
    return
  }

  await ensureDatabaseChildrenLoaded(databaseNode, {
    includeBackups: requestedType === 'backup'
  })

  let targetNode: TreeNodeData = databaseNode
  if (requestedType && requestedType !== 'database') {
    targetNode =
      findRouteTargetNode(databaseNode, requestedType, requestedObject) ||
      databaseNode
  }

  syncingRoute.value = true
  try {
    await applySelectedNode(targetNode, false)
  } finally {
    syncingRoute.value = false
  }
}

function findRouteTargetNode(databaseNode: TreeNodeData, type: ExplorerNodeType | 'group', objectName: string) {
  if (type === 'group') {
    return databaseNode.children?.find((item) => item.type === 'group' && item.groupKind === objectName) || null
  }

  const groupMap: Record<string, string> = {
    table: 'tables',
    view: 'views',
    function: 'functions',
    backup: 'backups'
  }

  const groupKind = groupMap[type]
  const groupNode = databaseNode.children?.find((item) => item.groupKind === groupKind)
  if (!groupNode) {
    return null
  }

  return groupNode.children?.find((item) => (item.tableName || item.label) === objectName) || null
}

async function updateRouteForNode(node: TreeNodeData) {
  const nextQuery: Record<string, string> = {
    db: node.databaseName
  }

  if (node.type === 'database') {
    nextQuery.type = 'database'
  } else if (node.type === 'group' && node.groupKind) {
    nextQuery.type = 'group'
    nextQuery.object = node.groupKind
  } else if (node.type === 'table' || node.type === 'view' || node.type === 'function' || node.type === 'backup') {
    nextQuery.type = node.type
    nextQuery.object = node.tableName || node.label
  }

  if (JSON.stringify(nextQuery) === JSON.stringify(route.query)) {
    return
  }

  await router.replace({
    path: '/mysql/databases',
    query: nextQuery
  })
}

function firstQueryValue(value: unknown) {
  if (Array.isArray(value)) {
    return String(value[0] || '')
  }
  return value ? String(value) : ''
}

async function selectObjectNode(type: 'table' | 'view' | 'function', objectName: string) {
  const databaseNode = currentDatabaseNode.value
  if (!databaseNode) {
    return
  }

  await ensureDatabaseChildrenLoaded(databaseNode)
  const targetNode = findRouteTargetNode(databaseNode, type, objectName)
  if (targetNode) {
    await applySelectedNode(targetNode, true)
  }
}

async function openDataPage() {
  if (!selectedNode.value || selectedNode.value.type !== 'table') {
    return
  }

  await router.push({
    path: '/mysql/data',
    query: {
      db: selectedNode.value.databaseName,
      table: selectedNode.value.tableName || selectedNode.value.label
    }
  })
}

function openSmartImportDialog() {
  if (!canImportData.value) {
    ElMessage.warning('当前账号没有导入数据权限')
    return
  }
  if (!selectedDatabaseName.value) {
    ElMessage.warning('请先选择数据库')
    return
  }

  if (smartImportInputRef.value) {
    smartImportInputRef.value.value = ''
    smartImportInputRef.value.click()
  }
}

async function handleSmartImportFileChange(event: Event) {
  const input = event.target as HTMLInputElement | null
  const file = input?.files?.[0]
  if (!file) {
    return
  }

  try {
    await performSmartImport(file)
  } finally {
    if (input) {
      input.value = ''
    }
  }
}

async function performSmartImport(file: File) {
  if (!selectedDatabaseName.value) {
    ElMessage.warning('请先选择数据库')
    return
  }

  smartImporting.value = true
  try {
    const payload = await parseSmartImportFile(file)
    if (payload.columns.length === 0) {
      ElMessage.warning('导入文件缺少表头')
      return
    }

    const result = await request.post<AutoImportResponse>('/api/metadata/table/auto-import', {
      database: selectedDatabaseName.value,
      name: payload.tableName,
      columns: payload.columns,
      rows: payload.rows
    })

    workspaceStore.setActiveTable(selectedDatabaseName.value, result.tableName)
    await refreshExplorer()
    await router.push({
      path: '/mysql/data',
      query: {
        db: selectedDatabaseName.value,
        table: result.tableName
      }
    })

    ElMessage.success(`已导入到数据表 ${result.tableName}，共 ${result.rowCount} 行`)
  } catch (error) {
    const message = error instanceof Error ? error.message : '智能导入失败'
    ElMessage.error(message)
  } finally {
    smartImporting.value = false
  }
}

async function handleSmartExport() {
  if (!canExportData.value) {
    ElMessage.warning('当前账号没有导出数据权限')
    return
  }
  if (!selectedDatabaseName.value || !activeTableName.value) {
    ElMessage.warning('请先选择数据表')
    return
  }

  smartExporting.value = true
  try {
    const dataset = await fetchAllTableRows(selectedDatabaseName.value, activeTableName.value)
    if (dataset.length === 0) {
      ElMessage.info('当前数据表暂无可导出数据')
      return
    }

    await downloadExcel(dataset, buildSmartExportFilename(selectedDatabaseName.value, activeTableName.value))
    ElMessage.success(`已导出 ${dataset.length} 行数据`)
  } catch (error) {
    const message = error instanceof Error ? error.message : '智能导出失败'
    ElMessage.error(message)
  } finally {
    smartExporting.value = false
  }
}

async function reloadCurrentDetails() {
  if (!selectedNode.value) {
    return
  }
  await loadNodeDetails(selectedNode.value)
}

async function copyDefinition() {
  if (!definitionText.value) {
    return
  }

  await navigator.clipboard.writeText(definitionText.value)
  ElMessage.success('定义语句已复制')
}

function openCompareDialog() {
  dialogs.compareVisible = true
}

async function submitCreateDatabase() {
  const name = forms.createDatabaseName.trim()
  if (!name) {
    ElMessage.warning('请输入数据库名称')
    return
  }

  await request.post('/api/metadata/database/create', { name })
  forms.createDatabaseName = ''
  dialogs.createDatabaseVisible = false
  ElMessage.success('数据库已创建')
  await refreshExplorer()
}

async function submitCreateTable() {
  const database = forms.createTableDatabase.trim()
  const name = forms.createTableName.trim()
  const columns = forms.createTableColumns.trim()

  if (!database || !name || !columns) {
    ElMessage.warning('请填写数据库、表名和字段定义')
    return
  }

  await request.post('/api/metadata/table/create', {
    database,
    name,
    columns,
    options: forms.createTableOptions.trim()
  })

  dialogs.createTableVisible = false
  ElMessage.success('数据表已创建')
  await refreshExplorer()
}

async function renameSelected() {
  if (!selectedNode.value) {
    return
  }

  if (selectedNode.value.type === 'database') {
    const { value } = await ElMessageBox.prompt('请输入新的数据库名称', '重命名数据库', {
      inputValue: selectedNode.value.databaseName
    })
    const nextName = value.trim()
    if (!nextName || nextName === selectedNode.value.databaseName) {
      return
    }

    await request.post('/api/metadata/database/rename', {
      oldName: selectedNode.value.databaseName,
      newName: nextName
    })
  } else if (selectedNode.value.type === 'table' || selectedNode.value.type === 'view') {
    const currentName = selectedNode.value.tableName || selectedNode.value.label
    const { value } = await ElMessageBox.prompt('请输入新的对象名称', '重命名对象', {
      inputValue: currentName
    })
    const nextName = value.trim()
    if (!nextName || nextName === currentName) {
      return
    }

    await request.post('/api/metadata/table/rename', {
      database: selectedNode.value.databaseName,
      oldName: currentName,
      newName: nextName
    })
  }

  ElMessage.success('重命名已完成')
  await refreshExplorer()
}

async function deleteSelected() {
  if (!selectedNode.value) {
    return
  }

  if (selectedNode.value.type === 'database') {
    await ElMessageBox.confirm(`确认删除数据库 ${selectedNode.value.databaseName} 吗？`, '删除数据库', {
      type: 'warning'
    })
    await request.post('/api/metadata/database/delete', {
      name: selectedNode.value.databaseName
    })
  } else if (selectedNode.value.type === 'table' || selectedNode.value.type === 'view') {
    const currentName = selectedNode.value.tableName || selectedNode.value.label
    await ElMessageBox.confirm(`确认删除对象 ${currentName} 吗？`, '删除对象', {
      type: 'warning'
    })
    await request.post('/api/metadata/table/delete', {
      database: selectedNode.value.databaseName,
      name: currentName
    })
  }

  selectedNode.value = null
  definitionText.value = ''
  tableColumns.value = []
  tableIndexes.value = []
  ElMessage.success('删除已完成')
  await refreshExplorer()
}

async function fetchAllTableRows(databaseName: string, tableName: string) {
  const pageLimit = 500
  const dataset: Array<Record<string, unknown>> = []
  let offset = 0
  let total = Number.POSITIVE_INFINITY

  while (offset < total) {
    const response = await request.get<QueryRowsResponse & { total?: number }>('/api/data/table', {
      params: {
        db: databaseName,
        table: tableName,
        limit: pageLimit,
        offset
      }
    })

    const rows = response.rows ?? []
    total = Number(response.total ?? rows.length)
    dataset.push(...rows)

    if (rows.length < pageLimit) {
      break
    }

    offset += rows.length
  }

  return dataset
}

function buildSmartExportFilename(databaseName: string, tableName: string) {
  const now = new Date()
  const timestamp = [
    now.getFullYear(),
    String(now.getMonth() + 1).padStart(2, '0'),
    String(now.getDate()).padStart(2, '0')
  ].join('') + '-' + [
    String(now.getHours()).padStart(2, '0'),
    String(now.getMinutes()).padStart(2, '0'),
    String(now.getSeconds()).padStart(2, '0')
  ].join('')

  return `${databaseName}_${tableName}_full_${timestamp}.xlsx`
}

async function parseSmartImportFile(file: File): Promise<SmartImportPayload> {
  const baseTableName = sanitizeImportedTableName(file.name.replace(/\.[^.]+$/, ''))
  const matrix = await readImportMatrix(file, '导入文件中没有可读取的工作表')
  return buildSmartImportPayloadFromMatrix(matrix, baseTableName)
}

function buildSmartImportPayloadFromMatrix(matrix: unknown[][], fallbackTableName: string): SmartImportPayload {
  if (matrix.length === 0) {
    throw new Error('导入文件内容为空')
  }

  const headerRow = (matrix[0] ?? []).map((cell) => String(cell ?? '').trim())
  const dataRows = matrix.slice(1)
  const usedNames = new Set<string>()
  const headers = headerRow.map((header, index) => normalizeImportedColumnName(header, index, usedNames))

  const rows = dataRows
    .map((cells) => {
      const row: Record<string, unknown> = {}
      headers.forEach((header, index) => {
        row[header] = normalizeImportedCellValue(cells[index])
      })
      return row
    })
    .filter((row) => Object.values(row).some((value) => value !== null && String(value).trim() !== ''))

  return {
    tableName: fallbackTableName || `import_${Date.now()}`,
    columns: inferImportColumns(headers, rows),
    rows
  }
}

function sanitizeImportedTableName(value: string) {
  const sanitized = sanitizeImportedIdentifier(value)
  return sanitized || `import_${Date.now()}`
}

function inferImportColumns(headers: string[], rows: Array<Record<string, unknown>>): SmartImportColumn[] {
  return headers.map((header) => ({
    name: header,
    type: inferImportedColumnType(
      header,
      rows
        .map((row) => row[header])
        .filter((value) => value !== null && value !== undefined && String(value).trim() !== '')
    )
  }))
}

function inferImportedColumnType(header: string, values: unknown[]) {
  if (values.length === 0) {
    return 'VARCHAR(255)'
  }

  const normalizedHeader = header.trim().toLowerCase()
  const dateKinds = values.map((value) => getStrictDateKind(value))

  if (!shouldPreferDiscreteCodeType(normalizedHeader) && dateKinds.every((kind) => kind !== null)) {
    if (dateKinds.every((kind) => kind === 'date')) {
      return 'DATE'
    }
    return 'DATETIME'
  }

  if (values.every((value) => isIntegerLikeValue(value)) && !looksLikePhoneNumber(values)) {
    return 'BIGINT'
  }

  if (values.every((value) => isDecimalLikeValue(value))) {
    return 'DECIMAL(18,4)'
  }

  const maxLength = values.reduce<number>((max, value) => Math.max(max, String(value ?? '').length), 0)
  if (maxLength > 1000) {
    return 'LONGTEXT'
  }
  if (maxLength > 255) {
    return 'TEXT'
  }

  return `VARCHAR(${Math.max(64, Math.min(255, maxLength + 32))})`
}

function looksLikePhoneNumber(values: unknown[]) {
  return values.every((value) => /^\d{11,}$/.test(String(value ?? '').trim()))
}

function shouldPreferDiscreteCodeType(header: string) {
  return (
    header === 'gender' ||
    header === 'sex' ||
    header === 'status' ||
    header === 'state' ||
    header === 'type' ||
    header === 'level' ||
    header === 'grade' ||
    header === 'code' ||
    header.endsWith('_code') ||
    header.endsWith('_status') ||
    header.endsWith('_type') ||
    header.includes('gender') ||
    header.includes('status') ||
    header.includes('state') ||
    header.includes('编号') ||
    header.includes('性别') ||
    header.includes('状态') ||
    header.includes('类型')
  )
}

function isIntegerLikeValue(value: unknown) {
  return /^-?\d+$/.test(String(value ?? '').trim())
}

function isDecimalLikeValue(value: unknown) {
  return /^-?\d+(\.\d+)?$/.test(String(value ?? '').trim())
}

function getStrictDateKind(value: unknown): 'date' | 'datetime' | null {
  if (value instanceof Date) {
    return 'datetime'
  }

  const text = String(value ?? '').trim()
  if (!text) {
    return null
  }
  if (/^\d{4}-\d{2}-\d{2}$/.test(text)) {
    return 'date'
  }
  if (/^\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}(:\d{2})?$/.test(text)) {
    return 'datetime'
  }
  return null
}
</script>

<style scoped>
.mysql-page__body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.mysql-page__grid {
  display: grid;
  grid-template-columns: minmax(320px, 380px) minmax(0, 1fr);
  gap: 20px;
}

:deep(.page-header-card .page-header-side) {
  flex: 1 1 auto;
  width: auto;
  max-width: none;
}

:deep(.page-header-card .page-header-actions) {
  flex-wrap: nowrap;
  justify-content: flex-end;
}

:deep(.page-header-card .page-header-actions .el-button) {
  flex-shrink: 0;
}

.mysql-page__main {
  min-width: 0;
}

.structure-toolbar {
  flex-wrap: wrap;
}

.structure-summary-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(220px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.structure-summary-card {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 8px;
  padding: 16px 18px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-md);
  background: var(--devops-bg-panel-soft);
}

.structure-summary-card span {
  color: var(--devops-text-secondary);
  font-size: 12px;
  font-weight: 600;
}

.structure-summary-card strong {
  color: var(--devops-text-primary);
  font-size: 15px;
  line-height: 1.6;
  white-space: normal;
  word-break: break-word;
}

.structure-content,
.structure-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.structure-descriptions {
  margin-bottom: 8px;
}

.structure-list-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.structure-list-card {
  padding: 18px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-md);
  background: var(--devops-bg-panel-soft);
}

.structure-list-card__header,
.structure-section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.structure-section__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.structure-list-card__header {
  margin-bottom: 14px;
}

.structure-list-card__header span,
.structure-section__header span {
  color: var(--devops-text-secondary);
  font-size: 12px;
}

.structure-chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.structure-chip {
  padding: 8px 12px;
  border: 1px solid var(--devops-border-light);
  border-radius: 999px;
  background: var(--devops-bg-panel);
  color: var(--devops-text-regular);
  cursor: pointer;
}

.structure-chip:hover {
  color: var(--devops-primary);
  border-color: var(--devops-primary);
}

.structure-table {
  width: 100%;
}

.structure-table-wrap {
  width: 100%;
  overflow-x: auto;
}

:deep(.structure-table) {
  min-width: 1240px;
}

.structure-definition-input {
  width: 100%;
}

:deep(.structure-definition-input .el-textarea__inner) {
  font-family: Consolas, 'Courier New', monospace;
  line-height: 1.6;
}

.smart-import-input {
  display: none;
}

@media (max-width: 1200px) {
  :deep(.page-header-card .page-header-actions) {
    flex-wrap: wrap;
  }

  .mysql-page__grid,
  .structure-list-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 900px) {
  .structure-summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>
