<template>
  <div class="workspace-shell">
    <div class="anime-aura anime-aura-left" />
    <div class="anime-aura anime-aura-right" />

    <el-container class="glass-panel workspace-panel">
      <el-aside width="340px" class="workspace-aside">
        <section class="aside-section aside-card aside-card--connection">
          <div class="section-heading section-heading--card">
            <span class="section-title">{{ t('common.connection') }}</span>
            <el-button class="soft-button mini-button mini-button--ghost" @click="refreshConnectionMeta">{{ t('common.refresh') }}</el-button>
          </div>

          <div class="token-pill token-pill--interactive connection-token-card" @click="copyConnectionId">
            <span class="status-dot" />
            <div class="token-copy">
              <el-tooltip
                :disabled="!connectionStore.token"
                :content="connectionStore.token || t('common.noConnection')"
                placement="top"
              >
                <span class="connection-id">
                  {{ connectionIdPreviewText }}
                </span>
              </el-tooltip>
              <span class="connection-caption">{{ connectionCaptionText }}</span>
            </div>
          </div>
        </section>

        <section class="aside-section aside-card aside-card--explorer tree-panel">
          <div class="section-heading section-heading--card explorer-heading">
            <span class="section-title">{{ t('common.explorer') }}</span>
            <div class="explorer-toolbar explorer-toolbar--capsule">
              <el-tooltip :content="t('workspace.newDatabase')" placement="top">
                <el-button class="soft-button toolbar-button toolbar-button--capsule" @click="openCreateDatabaseDialog">
                  <el-icon><Plus /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip :content="t('workspace.newTable')" placement="top">
                <el-button
                  class="soft-button toolbar-button toolbar-button--capsule"
                  :disabled="!workspaceStore.activeDatabase"
                  @click="openCreateTableDialog(workspaceStore.activeDatabase)"
                >
                  <el-icon><DocumentAdd /></el-icon>
                </el-button>
              </el-tooltip>
              <el-tooltip :content="t('common.refresh')" placement="top">
                <el-button class="soft-button toolbar-button toolbar-button--capsule" :loading="treeLoading" @click="refreshExplorer">
                  <el-icon><RefreshRight /></el-icon>
                </el-button>
              </el-tooltip>
            </div>
          </div>

          <el-tree
            ref="treeRef"
            class="explorer-tree"
            node-key="key"
            :data="explorerTree"
            :props="treeProps"
            :default-expanded-keys="expandedKeysArray"
            :expand-on-click-node="false"
            :highlight-current="true"
            :current-node-key="currentNodeKey"
            @node-click="handleNodeClick"
            @node-contextmenu="handleNodeContextMenu"
            @node-expand="handleNodeExpand"
            @node-collapse="handleNodeCollapse"
          >
            <template #default="{ data }">
              <div class="tree-node" @dblclick.stop="handleNodeDoubleClick(data)">
                <span class="tree-node__icon">
                  <el-icon v-if="data.type === 'database'"><Coin /></el-icon>
                  <el-icon v-else-if="data.type === 'group' && data.groupKind === 'tables'"><FolderOpened /></el-icon>
                  <el-icon v-else-if="data.type === 'group' && data.groupKind === 'views'"><FolderOpened /></el-icon>
                  <el-icon v-else-if="data.type === 'group'"><FolderOpened /></el-icon>
                  <el-icon v-else-if="data.type === 'table'"><Grid /></el-icon>
                  <el-icon v-else-if="data.type === 'view'"><View /></el-icon>
                  <el-icon v-else-if="data.type === 'function'"><SetUp /></el-icon>
                  <el-icon v-else-if="data.type === 'query'"><Promotion /></el-icon>
                  <el-icon v-else-if="data.type === 'backup'"><Box /></el-icon>
                </span>

                <div class="tree-node__meta">
                  <span class="tree-node__title">{{ getNodePrimaryText(data) }}</span>
                  <span v-if="getNodeMetaText(data)" class="tree-node__subtitle">{{ getNodeMetaText(data) }}</span>
                </div>
              </div>
            </template>
          </el-tree>
        </section>

        <section class="aside-section aside-card aside-card--security">
          <div class="section-heading section-heading--card">
            <span class="section-title">{{ isChinese ? '用户管理' : 'User Security' }}</span>
          </div>
          <button type="button" class="security-entry" @click="openSecurityTab">
            <span class="security-entry__icon">
              <el-icon><Lock /></el-icon>
            </span>
            <span class="security-entry__copy">
              <strong>{{ isChinese ? '用户与权限' : 'Users & Privileges' }}</strong>
              <small>{{ isChinese ? '创建用户、角色并配置权限' : 'Create users, roles, and privileges' }}</small>
            </span>
          </button>
        </section>
      </el-aside>

      <el-container class="workspace-main">
        <el-header class="workspace-header">
          <div class="workspace-header__intro">
            <div class="workspace-breadcrumb">
              <span>DevOps Console</span>
              <span>/</span>
              <span>MySQL Module</span>
              <span>/</span>
              <span>{{ t('common.workspace') }}</span>
            </div>

            <div class="brand-block" :class="{ 'brand-block--hint': isContextHintState }">
              <span class="brand-kicker">MYSQL VISUAL PLATFORM</span>
              <h1>{{ t('common.workspace') }}</h1>
              <p>{{ isChinese ? '统一管理连接、库表、查询与备份任务' : 'Manage connections, schemas, queries, and backups in one workspace' }}</p>
              <div
                v-if="isContextHintState"
                class="workspace-context-pill workspace-context-pill--hint"
                :title="headerDescriptionText"
              >
                <span class="workspace-context-pill__label">{{ t('common.context') }}</span>
                <p class="workspace-context-pill__hint-text">{{ headerDescriptionText }}</p>
              </div>
              <div
                v-else
                class="workspace-context-pill workspace-context-pill--active"
                :title="headerDescriptionText"
              >
                <span class="workspace-context-pill__label">{{ t('common.context') }}</span>
                <p class="workspace-context-pill__active-text">{{ headerDescriptionText }}</p>
              </div>
            </div>

            <div class="workspace-overview">
              <div class="workspace-overview__card">
                <span class="workspace-overview__label">{{ isChinese ? '连接主机' : 'Connection Host' }}</span>
                <strong class="workspace-overview__value">{{ connectionStore.profile.host || '-' }}</strong>
              </div>
              <div class="workspace-overview__card">
                <span class="workspace-overview__label">{{ isChinese ? '当前数据库' : 'Active Database' }}</span>
                <strong class="workspace-overview__value">{{ workspaceStore.activeDatabase || t('common.notSelected') }}</strong>
              </div>
              <div class="workspace-overview__card">
                <span class="workspace-overview__label">{{ isChinese ? '打开标签' : 'Open Tabs' }}</span>
                <strong class="workspace-overview__value">{{ tabs.length }}</strong>
              </div>
            </div>
          </div>

          <div class="header-actions workspace-actions-card">
            <div class="workspace-actions-group">
              <el-button class="soft-button workspace-action-button" @click="refreshExplorer">{{ t('workspace.refreshExplorer') }}</el-button>
              <el-button class="soft-button workspace-action-button" @click="openCreateDatabaseDialog">{{ t('workspace.newDatabase') }}</el-button>
              <el-button
                class="soft-button workspace-action-button"
                :disabled="!workspaceStore.activeDatabase"
                @click="openCreateTableDialog(workspaceStore.activeDatabase)"
              >
                {{ t('workspace.newTable') }}
              </el-button>
              <el-button
                class="soft-button workspace-action-button"
                :disabled="!workspaceStore.activeDatabase"
                :loading="smartImporting"
                @click="openSmartImportDialog"
              >
                {{ t('workspace.smartImport') }}
              </el-button>
              <el-button
                class="soft-button workspace-action-button"
                :disabled="!canSmartExport"
                :loading="smartExporting"
                @click="handleSmartExport"
              >
                {{ t('workspace.smartExport') }}
              </el-button>
              <el-button type="primary" class="soft-button primary-button workspace-action-button workspace-action-button--primary" @click="createNewQueryTab">
                {{ t('workspace.newQuery') }}
              </el-button>
              <el-button class="soft-button workspace-action-button workspace-language-button" @click="toggleLocale">
                {{ localeToggleLabel }}
              </el-button>
            </div>
          </div>
        </el-header>

        <el-main class="workspace-content">
          <el-empty
            v-if="tabs.length === 0"
            class="glass-subpanel empty-stage"
            :description="t('workspace.selectHint')"
          />

          <el-tabs
            v-else
            v-model="activeTabId"
            class="workspace-tabs"
            type="border-card"
            closable
            @tab-remove="closeTab"
            @tab-change="handleTabChange"
          >
            <el-tab-pane
              v-for="tab in tabs"
              :key="tab.id"
              :name="tab.id"
            >
              <template #label>
                <div class="workspace-tab-label" :title="tab.kind === 'query' ? `${tab.title} / ${tab.databaseName || t('workspace.noDatabase')}` : `${tab.title} / ${tab.databaseName || ''}`">
                  <span class="workspace-tab-label__title">{{ tab.title }}</span>
                  <span class="workspace-tab-label__meta">{{ tab.databaseName || t('workspace.tabMetaFallback') }}</span>
                </div>
              </template>
              <template v-if="tab.kind === 'data'">
                <TableDataTab
                  :key="`${tab.id}:${tab.refreshToken}`"
                  :db="tab.databaseName"
                  :table="tab.tableName"
                />
              </template>

              <template v-else-if="tab.kind === 'query'">
                <QueryTab
                  :title="tab.title"
                  :run-signal="tab.runSignal"
                  :active="activeTabId === tab.id"
                  :database-name="tab.databaseName"
                  :table-name="tab.tableName"
                  :initial-sql="tab.initialSql"
                />
              </template>

              <template v-else-if="tab.kind === 'security'">
                <UserManagementTab />
              </template>

              <template v-else-if="tab.kind === 'backup'">
                <BackupTab
                  :database-name="tab.databaseName"
                  :selected-backup-name="tab.selectedBackupName"
                  @refresh-explorer="refreshExplorer"
                />
              </template>

              <template v-else>
                <section class="glass-subpanel empty-stage empty-group-stage">
                  <div class="empty-stage__content">
                    <span class="empty-stage__eyebrow">{{ getGroupLabel(tab.groupKind) }}</span>
                    <h3>{{ getGroupEmptyDescription(tab.databaseName, tab.groupKind) }}</h3>
                    <div class="empty-group-actions">
                      <el-button
                        class="soft-button"
                        @click="openQueryTab(tab.databaseName)"
                      >
                        {{ t('workspace.openQueryWindow') }}
                      </el-button>
                    </div>
                  </div>
                </section>
              </template>
            </el-tab-pane>
          </el-tabs>
        </el-main>
      </el-container>
    </el-container>

    <input
      ref="smartImportInputRef"
      class="smart-import-input"
      type="file"
      accept=".xlsx,.xls,.csv,text/csv,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
      @change="handleSmartImportFileChange"
    >

    <div
      v-if="contextMenu.visible"
      class="context-menu"
      :style="{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }"
      @click.stop
    >
      <button
        v-for="item in contextMenu.items"
        :key="item.key"
        type="button"
        class="context-menu__item"
        @click="handleContextAction(item.key)"
      >
        {{ item.label }}
      </button>
    </div>

    <el-dialog v-model="createDatabaseDialog.visible" :title="t('workspace.createDatabaseTitle')" width="480px" class="workspace-dialog">
      <el-form label-position="top">
        <el-form-item :label="t('workspace.databaseName')">
          <el-input v-model="createDatabaseDialog.name" :placeholder="t('workspace.databaseNamePlaceholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDatabaseDialog.visible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="createDatabaseDialog.loading" @click="submitCreateDatabase">
          {{ t('common.create') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="createTableDialog.visible" :title="t('workspace.createTableTitle')" width="1080px" class="workspace-dialog create-table-dialog">
      <el-form label-position="top">
        <div class="create-table-grid">
          <el-form-item :label="t('common.database')">
            <el-input :model-value="createTableDialog.database" readonly />
          </el-form-item>
          <el-form-item :label="t('workspace.tableName')">
            <el-input v-model="createTableDialog.name" :placeholder="t('workspace.tableNamePlaceholder')" />
          </el-form-item>
        </div>

        <el-form-item :label="t('workspace.sqlImportTitle')">
          <div class="create-table-sql-import">
            <el-input
              v-model="createTableDialog.sqlInput"
              type="textarea"
              :autosize="{ minRows: 7, maxRows: 12 }"
              :placeholder="t('workspace.sqlImportPlaceholder')"
              class="create-table-sql-import__input"
            />
            <div class="create-table-sql-import__actions">
              <el-button class="soft-button" @click="applyCreateTableSQL">
                {{ t('workspace.parseCreateTableSql') }}
              </el-button>
            </div>
          </div>
        </el-form-item>

        <div class="column-toolbar">
          <span class="section-title">{{ isChinese ? '\u5b57\u6bb5\u8bbe\u8ba1' : 'Column Design' }}</span>
          <el-button class="soft-button" @click="addColumnDraft">{{ isChinese ? '\u6dfb\u52a0\u5b57\u6bb5' : 'Add Column' }}</el-button>
        </div>

        <div class="column-list">
          <div v-for="(column, index) in createTableDialog.columns" :key="column.id" class="column-card glass-subpanel">
            <div class="column-card__header">
              <span>{{ isChinese ? `\u5b57\u6bb5 ${index + 1}` : `Column ${index + 1}` }}</span>
              <el-button
                type="danger"
                link
                :disabled="createTableDialog.columns.length === 1"
                @click="removeColumnDraft(column.id)"
              >
                {{ t('tableData.delete') }}
              </el-button>
            </div>

            <div class="column-card__grid">
              <el-form-item :label="t('workspace.columnName')">
                <el-input v-model="column.name" placeholder="phone" />
              </el-form-item>

              <el-form-item :label="t('workspace.type')">
                <el-select
                  v-model="column.type"
                  filterable
                  allow-create
                  default-first-option
                  placeholder="VARCHAR"
                  @change="handleColumnTypeChange(column)"
                >
                  <el-option
                    v-for="option in columnTypeOptions"
                    :key="option"
                    :label="option"
                    :value="option"
                  />
                </el-select>
              </el-form-item>

              <el-form-item :label="t('workspace.length')">
                <el-input
                  v-model="column.length"
                  :disabled="!supportsColumnLength(column.type)"
                  :placeholder="getLengthPlaceholder(column.type)"
                  @blur="handleColumnLengthBlur(column)"
                />
              </el-form-item>

              <el-form-item :label="t('workspace.defaultValue')">
                <el-input
                  v-model="column.defaultValue"
                  :placeholder="getDefaultValuePlaceholder(column.type)"
                  @blur="handleColumnDefaultBlur(column)"
                />
              </el-form-item>

              <el-form-item :label="t('workspace.comment')" class="column-card__comment">
                <el-input v-model="column.comment" :placeholder="t('workspace.comment')" />
              </el-form-item>
            </div>

            <div class="column-card__switches">
              <el-checkbox v-model="column.primaryKey" @change="handleColumnPrimaryKeyChange(column)">{{ isChinese ? '\u4e3b\u952e' : 'Primary Key' }}</el-checkbox>
              <el-checkbox v-model="column.notNull" @change="handleColumnNotNullChange(column)">{{ isChinese ? '\u975e\u7a7a' : 'Not Null' }}</el-checkbox>
              <el-checkbox
                v-model="column.autoIncrement"
                :disabled="!supportsAutoIncrement(column.type)"
                @change="handleColumnAutoIncrementChange(column)"
              >
                {{ isChinese ? '\u81ea\u589e' : 'Auto Increment' }}
              </el-checkbox>
            </div>
          </div>
        </div>

        <el-form-item :label="t('workspace.sqlPreview')">
          <el-input
            :model-value="createTablePreviewSQL"
            type="textarea"
            readonly
            :autosize="{ minRows: 6, maxRows: 10 }"
            class="create-table-sql-preview__input"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="createTableDialog.visible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="createTableDialog.loading" @click="submitCreateTable">
          {{ t('workspace.createTableAction') }}
        </el-button>
      </template>
    </el-dialog>

    <StructureCompareDialog
      v-model:visible="structureCompareDialog.visible"
      :scope="structureCompareDialog.scope"
      :source-database="structureCompareDialog.sourceDatabase"
      :source-table="structureCompareDialog.sourceTable"
      @refresh-explorer="refreshExplorer"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  Box,
  Coin,
  Delete,
  DocumentAdd,
  EditPen,
  FolderOpened,
  Grid,
  Lock,
  Plus,
  Promotion,
  RefreshRight,
  SetUp,
  View
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import QueryTab from '@/components/workspace/QueryTab.vue'
import BackupTab from '@/components/workspace/BackupTab.vue'
import StructureCompareDialog from '@/components/workspace/StructureCompareDialog.vue'
import TableDataTab from '@/components/workspace/TableDataTab.vue'
import UserManagementTab from '@/components/workspace/UserManagementTab.vue'
import { useConnectionStore } from '@/stores/connection'
import { useWorkspaceStore } from '@/stores/workspace'
import {
  normalizeImportedCellValue,
  normalizeImportedColumnName,
  normalizeImportedTextValue,
  readImportMatrix,
  sanitizeImportedIdentifier
} from '@/utils/import-file'
import { useI18n } from '@/utils/i18n'
import request from '@/utils/request'
import { downloadExcel } from '@/utils/table-export'

type ExplorerNodeType = 'database' | 'group' | 'table' | 'view' | 'function' | 'query' | 'backup'
type ExplorerGroupKind = 'tables' | 'views' | 'functions' | 'queries' | 'backups'
type ContextActionKey =
  | 'refresh'
  | 'create-database'
  | 'create-table'
  | 'backup-database'
  | 'backup-table'
  | 'compare-structure'
  | 'open-backups'
  | 'restore-backup'
  | 'download-backup'
  | 'rename-backup'
  | 'delete-backup'
  | 'open-query'
  | 'rename-database'
  | 'delete-database'
  | 'rename-table'
  | 'delete-table'

interface TreeNodeData {
  key: string
  type: ExplorerNodeType
  label: string
  databaseName: string
  tableName?: string
  groupKind?: ExplorerGroupKind
  children?: TreeNodeData[]
  isLeaf?: boolean
}

interface ContextMenuItem {
  key: ContextActionKey
  label: string
}

interface DataTabState {
  id: string
  kind: 'data'
  title: string
  databaseName: string
  tableName: string
  refreshToken: number
}

interface QueryTabState {
  id: string
  kind: 'query'
  title: string
  databaseName: string
  tableName: string
  runSignal: number
  initialSql?: string
}

interface EmptyGroupTabState {
  id: string
  kind: 'empty-group'
  title: string
  databaseName: string
  tableName: string
  groupKind: ExplorerGroupKind
  groupLabel: string
  description: string
}

interface SecurityTabState {
  id: string
  kind: 'security'
  title: string
  databaseName: string
  tableName: string
}

interface BackupTabState {
  id: string
  kind: 'backup'
  title: string
  databaseName: string
  tableName: string
  selectedBackupName?: string
}

type WorkspaceTab = DataTabState | QueryTabState | EmptyGroupTabState | SecurityTabState | BackupTabState

interface TableColumnDraft {
  id: string
  name: string
  type: string
  length: string
  defaultValue: string
  primaryKey: boolean
  notNull: boolean
  autoIncrement: boolean
  onUpdateCurrentTimestamp: boolean
  comment: string
}

interface QueryRowsResponse {
  columns?: string[]
  rows?: Array<Record<string, unknown>>
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

interface BackupRecordResponse {
  records?: Array<{
    fileName: string
    displayName?: string
    tableName?: string
  }>
}

interface TableDataResponse {
  columns?: string[]
  rows?: Array<Record<string, unknown>>
  total?: number
}

const router = useRouter()
const connectionStore = useConnectionStore()
const workspaceStore = useWorkspaceStore()
const { isChinese, locale, toggleLocale, t } = useI18n()

const treeRef = ref()
const smartImportInputRef = ref<HTMLInputElement>()
const treeLoading = ref(false)
const smartImporting = ref(false)
const smartExporting = ref(false)
const explorerTree = ref<TreeNodeData[]>([])
const expandedKeys = ref(new Set<string>())
const currentNodeKey = ref('')
const tabs = ref<WorkspaceTab[]>([])
const activeTabId = ref('')

const contextMenu = reactive<{
  visible: boolean
  x: number
  y: number
  node: TreeNodeData | null
  items: ContextMenuItem[]
}>({
  visible: false,
  x: 0,
  y: 0,
  node: null,
  items: []
})

const createDatabaseDialog = reactive({
  visible: false,
  loading: false,
  name: ''
})

const createTableDialog = reactive<{
  visible: boolean
  loading: boolean
  database: string
  name: string
  sqlInput: string
  tableOptions: string
  columns: TableColumnDraft[]
}>({
  visible: false,
  loading: false,
  database: '',
  name: '',
  sqlInput: '',
  tableOptions: 'ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci',
  columns: []
})

const structureCompareDialog = reactive({
  visible: false,
  scope: 'database' as 'database' | 'table',
  sourceDatabase: '',
  sourceTable: ''
})

const treeProps = {
  label: 'label',
  children: 'children'
}

const columnTypeOptions = [
  'BIGINT',
  'INT',
  'SMALLINT',
  'TINYINT',
  'DECIMAL',
  'FLOAT',
  'DOUBLE',
  'VARCHAR',
  'CHAR',
  'TEXT',
  'LONGTEXT',
  'DATE',
  'DATETIME',
  'TIMESTAMP',
  'TIME',
  'JSON',
  'BOOLEAN'
]

const expandedKeysArray = computed(() => Array.from(expandedKeys.value))
const localeToggleLabel = computed(() => (locale.value === 'zh-CN' ? 'EN' : 'ZH'))
const integerTypeDefaultLengths: Record<string, string> = {
  TINYINT: '4',
  SMALLINT: '6',
  MEDIUMINT: '9',
  INT: '11',
  INTEGER: '11',
  BIGINT: '20'
}
const connectionIdPreviewText = computed(() => {
  if (!connectionStore.token) {
    return t('common.noConnection')
  }

  return `${connectionStore.token.slice(0, 8)}...`
})
const connectionCaptionText = computed(() => {
  if (!connectionStore.token) {
    return t('common.connectionIdEmpty')
  }

  return t('common.copyConnectionId')
})
const headerDescriptionText = computed(() => {
  if (workspaceStore.activeDatabase && workspaceStore.activeTable) {
    return t('workspace.currentContext', {
      database: workspaceStore.activeDatabase,
      table: workspaceStore.activeTable
    })
  }

  if (workspaceStore.activeDatabase) {
    return t('workspace.currentDatabase', { database: workspaceStore.activeDatabase })
  }

  return t('workspace.contextHint')
})
const isContextHintState = computed(() => !workspaceStore.activeDatabase && !workspaceStore.activeTable)
const createTablePreviewSQL = computed(() => buildCreateTablePreviewSQL())
const canSmartExport = computed(() => Boolean(workspaceStore.activeDatabase && workspaceStore.activeTable))

function getGroupLabel(kind: ExplorerGroupKind) {
  return t(`workspace.group.${kind}`)
}

function getGroupEmptyDescription(databaseName: string, kind: ExplorerGroupKind) {
  if (isChinese.value) {
    return `\u5f53\u524d\u6570\u636e\u5e93${databaseName}\u6682\u65e0${getGroupLabel(kind)}`
  }

  return t('workspace.groupEmptyTemplate', {
    database: databaseName,
    group: getGroupLabel(kind).toLowerCase()
  })
}

function getNodePrimaryText(node: TreeNodeData) {
  if (node.type === 'group' && node.groupKind) {
    return getGroupLabel(node.groupKind)
  }

  return node.label
}

function getNodeMetaText(node: Pick<TreeNodeData, 'type' | 'groupKind' | 'databaseName' | 'label'>) {
  if (node.type === 'database') {
    return t('common.database')
  }

  if (node.type === 'group' && node.groupKind) {
    return ''
  }

  if (node.type === 'table') {
    return t('common.table')
  }

  if (node.type === 'view') {
    return t('common.view')
  }

  if (node.type === 'function') {
    return t('common.function')
  }

  if (node.type === 'query') {
    return t('common.query')
  }

  if (node.type === 'backup') {
    return t('common.backup')
  }

  return ''
}

function buildDefaultColumnDraft(partial?: Partial<TableColumnDraft>): TableColumnDraft {
  const draft = {
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    name: partial?.name ?? '',
    type: partial?.type ?? 'VARCHAR',
    length: partial?.length ?? '255',
    defaultValue: partial?.defaultValue ?? '',
    primaryKey: partial?.primaryKey ?? false,
    notNull: partial?.notNull ?? false,
    autoIncrement: partial?.autoIncrement ?? false,
    onUpdateCurrentTimestamp: partial?.onUpdateCurrentTimestamp ?? false,
    comment: partial?.comment ?? ''
  }

  return normalizeColumnDraft(draft)
}

function quoteIdentifier(identifier: string) {
  return `\`${identifier.replace(/`/g, '``')}\``
}

function formatSQLStringLiteral(value: string) {
  return `'${value.replace(/\\/g, '\\\\').replace(/'/g, "\\'")}'`
}

function formatSQLDefaultValue(type: string, defaultValue: string) {
  const normalizedType = type.trim().toUpperCase()
  const trimmedValue = defaultValue.trim()

  if (!trimmedValue) {
    return ''
  }

  if (/^CURRENT_TIMESTAMP(\(\d+\))?$/i.test(trimmedValue)) {
    if (supportsCurrentTimestampDefault(normalizedType)) {
      return ` DEFAULT ${trimmedValue.toUpperCase()}`
    }

    return ''
  }

  if (/^NULL$/i.test(trimmedValue)) {
    return ` DEFAULT ${trimmedValue.toUpperCase()}`
  }

  if (/(INT|DECIMAL|FLOAT|DOUBLE|NUMERIC|REAL|BIT|SERIAL|BOOL)/.test(normalizedType) && /^-?\d+(\.\d+)?$/.test(trimmedValue)) {
    return ` DEFAULT ${trimmedValue}`
  }

  return ` DEFAULT ${formatSQLStringLiteral(trimmedValue)}`
}

function supportsColumnLength(type: string) {
  const normalizedType = type.trim().toUpperCase()
  return [
    'CHAR',
    'VARCHAR',
    'VARBINARY',
    'BINARY',
    'BIT',
    'DECIMAL',
    'NUMERIC',
    'FLOAT',
    'DOUBLE',
    'INT',
    'INTEGER',
    'TINYINT',
    'SMALLINT',
    'MEDIUMINT',
    'BIGINT',
    'DATETIME',
    'TIMESTAMP',
    'TIME'
  ].includes(normalizedType)
}

function supportsCurrentTimestampDefault(type: string) {
  const normalizedType = type.trim().toUpperCase()
  return normalizedType === 'TIMESTAMP' || normalizedType === 'DATETIME'
}

function isStringColumnType(type: string) {
  const normalizedType = type.trim().toUpperCase()
  return ['CHAR', 'VARCHAR', 'TEXT', 'LONGTEXT', 'JSON'].includes(normalizedType)
}

function isNumericColumnType(type: string) {
  const normalizedType = type.trim().toUpperCase()
  return ['TINYINT', 'SMALLINT', 'MEDIUMINT', 'INT', 'INTEGER', 'BIGINT', 'DECIMAL', 'NUMERIC', 'FLOAT', 'DOUBLE', 'BOOLEAN', 'BIT'].includes(normalizedType)
}

function isTemporalColumnType(type: string) {
  const normalizedType = type.trim().toUpperCase()
  return ['DATE', 'DATETIME', 'TIMESTAMP', 'TIME'].includes(normalizedType)
}

function supportsAutoIncrement(type: string) {
  const normalizedType = type.trim().toUpperCase()
  return ['INT', 'INTEGER', 'TINYINT', 'SMALLINT', 'MEDIUMINT', 'BIGINT'].includes(normalizedType)
}

function normalizeColumnType(type: string) {
  return type.trim().toUpperCase() || 'VARCHAR'
}

function normalizeColumnLength(type: string, length: string) {
  const normalizedType = normalizeColumnType(type)
  const trimmedLength = length.trim()

  if (!supportsColumnLength(normalizedType)) {
    return ''
  }

  if (normalizedType in integerTypeDefaultLengths) {
    const parsed = Number.parseInt(trimmedLength, 10)
    if (!Number.isFinite(parsed) || parsed < Number.parseInt(integerTypeDefaultLengths[normalizedType], 10)) {
      return integerTypeDefaultLengths[normalizedType]
    }

    return String(parsed)
  }

  if (normalizedType === 'VARCHAR' || normalizedType === 'CHAR' || normalizedType === 'BIT' || normalizedType === 'BINARY' || normalizedType === 'VARBINARY') {
    const parsed = Number.parseInt(trimmedLength, 10)
    if (!Number.isFinite(parsed) || parsed <= 0) {
      return normalizedType === 'CHAR' ? '1' : '255'
    }

    return String(parsed)
  }

  if (normalizedType === 'DECIMAL' || normalizedType === 'NUMERIC' || normalizedType === 'FLOAT' || normalizedType === 'DOUBLE') {
    if (!trimmedLength) {
      return normalizedType === 'DECIMAL' || normalizedType === 'NUMERIC' ? '10,2' : ''
    }

    return /^\d+\s*(,\s*\d+)?$/.test(trimmedLength) ? trimmedLength.replace(/\s+/g, '') : ''
  }

  if (normalizedType === 'DATETIME' || normalizedType === 'TIMESTAMP' || normalizedType === 'TIME') {
    if (!trimmedLength) {
      return ''
    }

    const parsed = Number.parseInt(trimmedLength, 10)
    if (!Number.isFinite(parsed)) {
      return ''
    }

    return String(Math.max(0, Math.min(6, parsed)))
  }

  return trimmedLength
}

function getDefaultValuePlaceholder(type: string) {
  const normalizedType = normalizeColumnType(type)

  if (supportsCurrentTimestampDefault(normalizedType)) {
    return 'CURRENT_TIMESTAMP'
  }

  if (isTemporalColumnType(normalizedType)) {
    return normalizedType === 'DATE' ? '2026-04-23' : '2026-04-23 10:00:00'
  }

  if (isNumericColumnType(normalizedType)) {
    return '0'
  }

  return ''
}

function getLengthPlaceholder(type: string) {
  const normalizedType = normalizeColumnType(type)

  if (normalizedType === 'DATETIME' || normalizedType === 'TIMESTAMP' || normalizedType === 'TIME') {
    return '0-6'
  }

  if (normalizedType === 'DECIMAL' || normalizedType === 'NUMERIC') {
    return '10,2'
  }

  if (normalizedType === 'FLOAT' || normalizedType === 'DOUBLE') {
    return ''
  }

  if (supportsColumnLength(normalizedType)) {
    return '255'
  }

  return ''
}

function formatDateByType(type: string, date: Date) {
  const pad = (value: number) => String(value).padStart(2, '0')
  const year = date.getFullYear()
  const month = pad(date.getMonth() + 1)
  const day = pad(date.getDate())
  const hour = pad(date.getHours())
  const minute = pad(date.getMinutes())
  const second = pad(date.getSeconds())
  const normalizedType = normalizeColumnType(type)

  if (normalizedType === 'DATE') {
    return `${year}-${month}-${day}`
  }

  if (normalizedType === 'TIME') {
    return `${hour}:${minute}:${second}`
  }

  return `${year}-${month}-${day} ${hour}:${minute}:${second}`
}

function parseUnixTimestampDefault(raw: string) {
  if (!/^\d{10,13}$/.test(raw)) {
    return null
  }

  const numeric = Number.parseInt(raw, 10)
  if (!Number.isFinite(numeric)) {
    return null
  }

  const millis = raw.length === 13 ? numeric : numeric * 1000
  const date = new Date(millis)
  if (Number.isNaN(date.getTime())) {
    return null
  }

  return date
}

function normalizeDefaultValueForType(type: string, defaultValue: string) {
  const normalizedType = normalizeColumnType(type)
  const trimmedValue = defaultValue.trim()

  if (!trimmedValue) {
    return { value: '', warningKey: '' }
  }

  if (isStringColumnType(normalizedType) && /^CURRENT_TIMESTAMP(\(\d+\))?$/i.test(trimmedValue)) {
    return { value: '', warningKey: 'workspace.invalidStringCurrentTimestamp' }
  }

  if (isNumericColumnType(normalizedType) && (/^CURRENT_TIMESTAMP(\(\d+\))?$/i.test(trimmedValue) || getStrictDateKind(trimmedValue) !== null)) {
    return { value: '', warningKey: 'workspace.invalidNumericTemporalDefault' }
  }

  if (isTemporalColumnType(normalizedType)) {
    if (/^CURRENT_TIMESTAMP(\(\d+\))?$/i.test(trimmedValue)) {
      if (supportsCurrentTimestampDefault(normalizedType)) {
        return { value: trimmedValue.toUpperCase(), warningKey: '' }
      }

      return { value: '', warningKey: 'workspace.invalidTemporalCurrentTimestamp' }
    }

    const unixTimestamp = parseUnixTimestampDefault(trimmedValue)
    if (unixTimestamp) {
      return {
        value: formatDateByType(normalizedType, unixTimestamp),
        warningKey: 'workspace.convertedTimestampDefault'
      }
    }
  }

  return { value: trimmedValue, warningKey: '' }
}

function normalizeColumnDraft(column: TableColumnDraft) {
  const type = normalizeColumnType(column.type)
  const normalizedDefault = normalizeDefaultValueForType(type, column.defaultValue)
  const normalized: TableColumnDraft = {
    ...column,
    type,
    length: normalizeColumnLength(type, column.length),
    defaultValue: normalizedDefault.value,
    comment: column.comment.trim()
  }

  if (!supportsAutoIncrement(type)) {
    normalized.autoIncrement = false
  }

  if (!supportsCurrentTimestampDefault(type)) {
    normalized.onUpdateCurrentTimestamp = false
  }

  if (normalized.autoIncrement) {
    normalized.primaryKey = true
    normalized.notNull = true
  } else if (normalized.primaryKey) {
    normalized.notNull = true
  }

  return normalized
}

function buildColumnLengthClause(type: string, length: string) {
  const normalizedType = normalizeColumnType(type)
  const normalizedLength = normalizeColumnLength(normalizedType, length)
  if (!supportsColumnLength(normalizedType) || !normalizedLength) {
    return ''
  }

  return `(${normalizedLength})`
}

function normalizeCreateTableDialogColumns() {
  let autoIncrementOwner = ''

  createTableDialog.columns = createTableDialog.columns.map((column) => {
    const normalized = normalizeColumnDraft(column)

    if (normalized.autoIncrement) {
      if (!autoIncrementOwner) {
        autoIncrementOwner = normalized.id
      } else {
        normalized.autoIncrement = false
      }
    }

    return normalized
  })

  if (autoIncrementOwner) {
    createTableDialog.columns = createTableDialog.columns.map((column) =>
      column.id === autoIncrementOwner
        ? { ...column, primaryKey: true, notNull: true }
        : column
    )
  }
}

function decodeSQLStringLiteral(raw: string) {
  const trimmed = raw.trim()
  if (trimmed.length < 2) {
    return trimmed
  }

  const quote = trimmed[0]
  if ((quote !== '\'' && quote !== '"') || trimmed[trimmed.length - 1] !== quote) {
    return trimmed
  }

  return trimmed
    .slice(1, -1)
    .replace(/\\\\/g, '\\')
    .replace(/\\'/g, '\'')
    .replace(/\\"/g, '"')
}

function splitSQLDefinitions(source: string) {
  const items: string[] = []
  let current = ''
  let depth = 0
  let quote = ''

  for (let index = 0; index < source.length; index += 1) {
    const char = source[index]
    const previous = source[index - 1]

    if (quote) {
      current += char
      if (char === quote && previous !== '\\') {
        quote = ''
      }
      continue
    }

    if (char === '\'' || char === '"' || char === '`') {
      quote = char
      current += char
      continue
    }

    if (char === '(') {
      depth += 1
      current += char
      continue
    }

    if (char === ')') {
      depth = Math.max(0, depth - 1)
      current += char
      continue
    }

    if (char === ',' && depth === 0) {
      if (current.trim()) {
        items.push(current.trim())
      }
      current = ''
      continue
    }

    current += char
  }

  if (current.trim()) {
    items.push(current.trim())
  }

  return items
}

function parseCreateTableSQLInput(sql: string) {
  const trimmed = sql.trim().replace(/;\s*$/, '')
  const statementMatch = trimmed.match(/^CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?((?:`[^`]+`|\w+)(?:\.(?:`[^`]+`|\w+))?)\s*\(([\s\S]+)\)\s*([\s\S]*)$/i)
  if (!statementMatch) {
    throw new Error(t('workspace.invalidCreateTableSql'))
  }

  const identifier = statementMatch[1].trim()
  const body = statementMatch[2]
  const tableOptions = statementMatch[3]?.trim() || 'ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci'
  const identifierParts = identifier.split('.').map((part) => part.replace(/^`|`$/g, ''))
  const tableName = identifierParts[identifierParts.length - 1]
  const databaseName = identifierParts.length > 1 ? identifierParts[0] : ''

  const definitions = splitSQLDefinitions(body)
  const columns = new Map<string, Partial<TableColumnDraft>>()
  const primaryKeys = new Set<string>()

  for (const definition of definitions) {
    if (/^PRIMARY\s+KEY/i.test(definition)) {
      const primaryMatch = definition.match(/\(([\s\S]+)\)/)
      if (!primaryMatch) {
        continue
      }

      splitSQLDefinitions(primaryMatch[1]).forEach((item) => {
        const name = item.replace(/^`|`$/g, '').trim()
        if (name) {
          primaryKeys.add(name)
        }
      })
      continue
    }

    if (/^(UNIQUE\s+KEY|UNIQUE\s+INDEX|KEY|INDEX|CONSTRAINT|FULLTEXT|SPATIAL|FOREIGN\s+KEY)/i.test(definition)) {
      continue
    }

    const columnMatch = definition.match(/^(`[^`]+`|\w+)\s+([A-Za-z]+)(?:\s*\(([^)]*)\))?([\s\S]*)$/)
    if (!columnMatch) {
      continue
    }

    const name = columnMatch[1].replace(/^`|`$/g, '')
    const type = normalizeColumnType(columnMatch[2])
    const length = columnMatch[3]?.trim() ?? ''
    const extras = columnMatch[4] ?? ''
    const defaultMatch = extras.match(/\bDEFAULT\s+((?:CURRENT_TIMESTAMP(?:\(\d+\))?)|NULL|'(?:[^'\\]|\\.)*'|"(?:[^"\\]|\\.)*"|[^\s,]+)/i)
    const commentMatch = extras.match(/\bCOMMENT\s+('(?:[^'\\]|\\.)*'|"(?:[^"\\]|\\.)*")/i)

    columns.set(name, normalizeColumnDraft(buildDefaultColumnDraft({
      name,
      type,
      length,
      defaultValue: defaultMatch ? decodeSQLStringLiteral(defaultMatch[1]) : '',
      notNull: /\bNOT\s+NULL\b/i.test(extras),
      primaryKey: /\bPRIMARY\s+KEY\b/i.test(extras),
      autoIncrement: /\bAUTO_INCREMENT\b/i.test(extras),
      onUpdateCurrentTimestamp: /\bON\s+UPDATE\s+CURRENT_TIMESTAMP(?:\(\d+\))?\b/i.test(extras),
      comment: commentMatch ? decodeSQLStringLiteral(commentMatch[1]) : ''
    })))
  }

  for (const primaryKey of primaryKeys) {
    const column = columns.get(primaryKey)
    if (column) {
      column.primaryKey = true
      column.notNull = true
    }
  }

  const drafts = Array.from(columns.values()).map((column) => normalizeColumnDraft(buildDefaultColumnDraft(column)))
  if (drafts.length === 0) {
    throw new Error(t('workspace.invalidCreateTableSql'))
  }

  return {
    databaseName,
    tableName,
    tableOptions,
    columns: drafts
  }
}

function buildColumnDefinition(column: TableColumnDraft) {
  const type = column.type.trim().toUpperCase()
  const lengthClause = buildColumnLengthClause(type, column.length)
  const parts = [`${quoteIdentifier(column.name.trim())} ${type}${lengthClause}`]

  if (column.notNull || column.primaryKey || column.autoIncrement) {
    parts.push('NOT NULL')
  }

  const defaultClause = formatSQLDefaultValue(type, column.defaultValue)
  if (defaultClause) {
    parts.push(defaultClause.trim())
  }

  if (column.autoIncrement && supportsAutoIncrement(type)) {
    parts.push('AUTO_INCREMENT')
  }

  if (column.onUpdateCurrentTimestamp && supportsCurrentTimestampDefault(type)) {
    parts.push('ON UPDATE CURRENT_TIMESTAMP')
  }

  if (column.comment.trim()) {
    parts.push(`COMMENT ${formatSQLStringLiteral(column.comment.trim())}`)
  }

  return parts.join(' ')
}

function buildCreateTableColumnsSQL() {
  const activeColumns = createTableDialog.columns.filter((column) => column.name.trim() && column.type.trim())
  const definitions = activeColumns.map(buildColumnDefinition)
  const primaryKeys = activeColumns.filter((column) => column.primaryKey).map((column) => quoteIdentifier(column.name.trim()))

  if (primaryKeys.length > 0) {
    definitions.push(`PRIMARY KEY (${primaryKeys.join(', ')})`)
  }

  return definitions.join(',\n  ')
}

function buildCreateTablePreviewSQL() {
  if (!createTableDialog.database || !createTableDialog.name.trim()) {
    return '-- Select a database and enter a table name to preview SQL'
  }

  const columnsSQL = buildCreateTableColumnsSQL()
  if (!columnsSQL) {
    return `-- ${t('workspace.addValidField')}`
  }

  return [
    `CREATE TABLE ${quoteIdentifier(createTableDialog.database)}.${quoteIdentifier(createTableDialog.name.trim())} (`,
    `  ${columnsSQL.replace(/\n/g, '\n  ')}`,
    `) ${createTableDialog.tableOptions.trim() || 'ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci'};`
  ].join('\n')
}

function validateCreateTableDialog() {
  normalizeCreateTableDialogColumns()

  if (!createTableDialog.database) {
    throw new Error(t('workspace.pickDatabaseFirst'))
  }

  if (!createTableDialog.name.trim()) {
    throw new Error(t('workspace.enterTableName'))
  }

  const activeColumns = createTableDialog.columns.filter((column) => column.name.trim() && column.type.trim())
  if (activeColumns.length === 0) {
    throw new Error(t('workspace.addValidField'))
  }

  const names = new Set<string>()
  for (const column of activeColumns) {
    const normalizedName = column.name.trim().toLowerCase()
    if (names.has(normalizedName)) {
      throw new Error(t('workspace.fieldNameDuplicate', { name: column.name }))
    }
    names.add(normalizedName)
  }

  return activeColumns
}

async function fetchDatabaseObjects(databaseName: string) {
  const [tables, views, functions, backupResponse] = await Promise.all([
    request.get<string[]>('/api/metadata/tables', {
      params: { db: databaseName }
    }),
    request
      .post<QueryRowsResponse>('/api/query/execute', {
        sql: `
          SELECT TABLE_NAME
          FROM information_schema.VIEWS
          WHERE TABLE_SCHEMA = ${formatSQLStringLiteral(databaseName)}
          ORDER BY TABLE_NAME
        `,
        database: databaseName
      })
      .then((result) =>
        (result.rows ?? [])
          .map((row) => String(row.TABLE_NAME ?? row.table_name ?? '').trim())
          .filter(Boolean)
      )
      .catch(() => [] as string[]),
    request
      .post<QueryRowsResponse>('/api/query/execute', {
        sql: `
          SELECT ROUTINE_NAME
          FROM information_schema.ROUTINES
          WHERE ROUTINE_SCHEMA = ${formatSQLStringLiteral(databaseName)}
            AND ROUTINE_TYPE = 'FUNCTION'
          ORDER BY ROUTINE_NAME
        `,
        database: databaseName
      })
      .then((result) =>
        (result.rows ?? [])
          .map((row) => String(row.ROUTINE_NAME ?? row.routine_name ?? '').trim())
          .filter(Boolean)
      )
      .catch(() => [] as string[]),
    request
      .get<BackupRecordResponse>('/api/backup/list', {
        params: { database: databaseName }
      })
      .catch(() => ({ records: [] as Array<{ fileName: string }> }))
  ])

  const viewSet = new Set(views)
  const plainTables = tables.filter((name) => !viewSet.has(name))

  return {
    tables: plainTables,
    views,
    functions,
    queries: [] as string[],
    backups: (backupResponse.records ?? []).map((record) => record.fileName)
  }
}

async function loadExplorerTree() {
  treeLoading.value = true

  try {
    const databases = await request.get<string[]>('/api/metadata/databases')
    const nodes = await Promise.all(
      databases.map(async (databaseName) => {
        const objects = await fetchDatabaseObjects(databaseName).catch(() => ({
          tables: [] as string[],
          views: [] as string[],
          functions: [] as string[],
          queries: [] as string[],
          backups: [] as string[]
        }))

        const groupChildren: TreeNodeData[] = [
          {
            key: `db:${databaseName}:tables`,
            type: 'group',
            label: 'tables',
            databaseName,
            groupKind: 'tables',
            children: objects.tables.map((tableName) => ({
              key: `db:${databaseName}:table:${tableName}`,
              type: 'table',
              label: tableName,
              databaseName,
              tableName,
              isLeaf: true
            }))
          },
          {
            key: `db:${databaseName}:views`,
            type: 'group',
            label: 'views',
            databaseName,
            groupKind: 'views',
            children: objects.views.map((viewName) => ({
              key: `db:${databaseName}:view:${viewName}`,
              type: 'view',
              label: viewName,
              databaseName,
              tableName: viewName,
              isLeaf: true
            }))
          },
          {
            key: `db:${databaseName}:functions`,
            type: 'group',
            label: 'functions',
            databaseName,
            groupKind: 'functions',
            children: objects.functions.map((functionName) => ({
              key: `db:${databaseName}:function:${functionName}`,
              type: 'function',
              label: functionName,
              databaseName,
              tableName: functionName,
              isLeaf: true
            }))
          },
          {
            key: `db:${databaseName}:queries`,
            type: 'group',
            label: 'queries',
            databaseName,
            groupKind: 'queries',
            children: []
          },
          {
            key: `db:${databaseName}:backups`,
            type: 'group',
            label: 'backups',
            databaseName,
            groupKind: 'backups',
            children: objects.backups.map((backupName) => ({
              key: `db:${databaseName}:backup:${backupName}`,
              type: 'backup',
              label: backupName,
              databaseName,
              tableName: backupName,
              isLeaf: true
            }))
          }
        ]

        return {
          key: `db:${databaseName}`,
          type: 'database',
          label: databaseName,
          databaseName,
          children: groupChildren
        } as TreeNodeData
      })
    )

    explorerTree.value = nodes
    await nextTick()
    restoreCurrentTreeSelection()
    initializeWorkspaceFromLoginDatabase()
  } finally {
    treeLoading.value = false
  }
}

function restoreCurrentTreeSelection() {
  if (!currentNodeKey.value) {
    return
  }

  treeRef.value?.setCurrentKey(currentNodeKey.value)
}

function findNodeByKey(nodes: TreeNodeData[], key: string): TreeNodeData | null {
  for (const node of nodes) {
    if (node.key === key) {
      return node
    }

    const children = node.children ?? []
    const matched = findNodeByKey(children, key)
    if (matched) {
      return matched
    }
  }

  return null
}

function findFirstAutoOpenNode(databaseNode: TreeNodeData) {
  const groups = databaseNode.children ?? []
  const tablesGroup = groups.find((node) => node.type === 'group' && node.groupKind === 'tables')
  const viewsGroup = groups.find((node) => node.type === 'group' && node.groupKind === 'views')
  const functionsGroup = groups.find((node) => node.type === 'group' && node.groupKind === 'functions')

  return (
    tablesGroup?.children?.[0] ??
    viewsGroup?.children?.[0] ??
    functionsGroup?.children?.[0] ??
    tablesGroup ??
    databaseNode
  )
}

function initializeWorkspaceFromLoginDatabase() {
  if (currentNodeKey.value || tabs.value.length > 0) {
    return
  }

  const preferredDatabase = connectionStore.profile.database.trim()
  if (!preferredDatabase) {
    return
  }

  const databaseNode = findNodeByKey(explorerTree.value, `db:${preferredDatabase}`)
  if (!databaseNode) {
    return
  }

  const preferredNode = findFirstAutoOpenNode(databaseNode)
  setCurrentNode(preferredNode)
  syncWorkspaceFromNode(preferredNode)

  if (preferredNode.type === 'table' || preferredNode.type === 'view') {
    openDataTab(preferredNode)
    return
  }

  if (preferredNode.type === 'function') {
    openFunctionTab(preferredNode)
    return
  }

  if (preferredNode.type === 'group') {
    openEmptyGroupTab(preferredNode)
  }
}

function ensureExpandedForNode(node: TreeNodeData) {
  const parts = node.key.split(':')
  if (parts.length >= 2) {
    expandedKeys.value.add(`db:${node.databaseName}`)
  }

  if (node.type !== 'database' && node.groupKind) {
    expandedKeys.value.add(node.key)
  }

  if ((node.type === 'table' || node.type === 'view' || node.type === 'function') && node.databaseName) {
    const groupKey = node.key.split(':').slice(0, 3).join(':')
    expandedKeys.value.add(groupKey)
  }
}

function setCurrentNode(node: TreeNodeData, options?: { keepExpanded?: boolean }) {
  currentNodeKey.value = node.key
  if (options?.keepExpanded !== false) {
    ensureExpandedForNode(node)
  }
  nextTick(() => {
    treeRef.value?.setCurrentKey(node.key)
  })
}

function syncWorkspaceFromNode(node: TreeNodeData) {
  if (node.type === 'database') {
    workspaceStore.setActiveDatabase(node.databaseName)
    workspaceStore.clearActiveTable()
    return
  }

  if (node.type === 'group') {
    workspaceStore.setActiveDatabase(node.databaseName)
    workspaceStore.clearActiveTable()
    return
  }

  if (node.type === 'table' || node.type === 'view' || node.type === 'function') {
    workspaceStore.setActiveTable(node.databaseName, node.tableName ?? node.label)
    return
  }

  workspaceStore.setActiveDatabase(node.databaseName)
}

function openDataTab(node: TreeNodeData) {
  if (!node.tableName) {
    return
  }

  const existing = tabs.value.find(
    (tab) =>
      tab.kind === 'data' &&
      tab.databaseName === node.databaseName &&
      tab.tableName === node.tableName
  )

  if (existing) {
    activeTabId.value = existing.id
    return
  }

  const tab: DataTabState = {
    id: `data:${node.databaseName}:${node.tableName}`,
    kind: 'data',
    title: node.label,
    databaseName: node.databaseName,
    tableName: node.tableName,
    refreshToken: 0
  }

  tabs.value.push(tab)
  activeTabId.value = tab.id
}

function refreshDataTab(databaseName: string, tableName: string) {
  const existing = tabs.value.find(
    (tab) => tab.kind === 'data' && tab.databaseName === databaseName && tab.tableName === tableName
  )

  if (!existing || existing.kind !== 'data') {
    return
  }

  existing.refreshToken += 1
  activeTabId.value = existing.id
}

function openBackupTab(databaseName: string, selectedBackupName = '') {
  const existing = tabs.value.find(
    (tab) => tab.kind === 'backup' && tab.databaseName === databaseName
  )

  if (existing && existing.kind === 'backup') {
    existing.selectedBackupName = selectedBackupName || existing.selectedBackupName
    activeTabId.value = existing.id
    return
  }

  const tab: BackupTabState = {
    id: `backup:${databaseName}`,
    kind: 'backup',
    title: `${databaseName}/${getGroupLabel('backups')}`,
    databaseName,
    tableName: '',
    selectedBackupName
  }

  tabs.value.push(tab)
  activeTabId.value = tab.id
}

function openStructureCompareDialog(node: TreeNodeData) {
  structureCompareDialog.scope = node.type === 'database' ? 'database' : 'table'
  structureCompareDialog.sourceDatabase = node.databaseName
  structureCompareDialog.sourceTable = node.type === 'table' ? (node.tableName || node.label) : ''
  structureCompareDialog.visible = true
}

function openQueryTab(databaseName: string, sqlTitle?: string) {
  const queryIndex = tabs.value.filter((tab) => tab.kind === 'query').length + 1
  const tab: QueryTabState = {
    id: `query:${Date.now()}:${Math.random().toString(36).slice(2, 8)}`,
    kind: 'query',
    title: sqlTitle || t('workspace.queryTitle', { index: queryIndex }),
    databaseName,
    tableName: '',
    runSignal: 0,
    initialSql: undefined
  }

  tabs.value.push(tab)
  activeTabId.value = tab.id
}

function openSecurityTab() {
  const tabId = 'security:user-management'
  const existing = tabs.value.find((tab) => tab.id === tabId)
  if (existing) {
    activeTabId.value = existing.id
    return
  }

  const tab: SecurityTabState = {
    id: tabId,
    kind: 'security',
    title: isChinese.value ? '用户管理' : 'User Security',
    databaseName: '',
    tableName: ''
  }

  tabs.value.push(tab)
  activeTabId.value = tab.id
}

function openFunctionTab(node: TreeNodeData) {
  if (!node.databaseName || !node.tableName) {
    return
  }

  const initialSql = `SHOW CREATE FUNCTION ${quoteIdentifier(node.databaseName)}.${quoteIdentifier(node.tableName)};`
  const existing = tabs.value.find(
    (tab) =>
      tab.kind === 'query' &&
      tab.databaseName === node.databaseName &&
      tab.tableName === node.tableName &&
      tab.initialSql === initialSql
  )

  if (existing) {
    activeTabId.value = existing.id
    return
  }

  const tab: QueryTabState = {
    id: `function:${node.databaseName}:${node.tableName}`,
    kind: 'query',
    title: node.label,
    databaseName: node.databaseName,
    tableName: node.tableName,
    runSignal: Date.now(),
    initialSql
  }

  tabs.value.push(tab)
  activeTabId.value = tab.id
}

function openEmptyGroupTab(node: TreeNodeData) {
  if (!node.groupKind) {
    return
  }

  const existing = tabs.value.find(
    (tab) =>
      tab.kind === 'empty-group' &&
      tab.databaseName === node.databaseName &&
      tab.groupKind === node.groupKind
  )

  if (existing) {
    activeTabId.value = existing.id
    return
  }

  const tab: EmptyGroupTabState = {
    id: `empty:${node.databaseName}:${node.groupKind}`,
    kind: 'empty-group',
    title: `${node.databaseName}/${getGroupLabel(node.groupKind)}`,
    databaseName: node.databaseName,
    tableName: '',
    groupKind: node.groupKind,
    groupLabel: getGroupLabel(node.groupKind),
    description: getGroupEmptyDescription(node.databaseName, node.groupKind)
  }

  tabs.value.push(tab)
  activeTabId.value = tab.id
}

function handleNodeClick(node: TreeNodeData) {
  closeContextMenu()
  setCurrentNode(node)
  syncWorkspaceFromNode(node)

  if (node.type === 'database') {
    return
  }

  if (node.type === 'group') {
    if (node.groupKind === 'backups') {
      openBackupTab(node.databaseName)
      return
    }

    if ((node.children?.length ?? 0) === 0) {
      openEmptyGroupTab(node)
    }
    return
  }

  if (node.type === 'table' || node.type === 'view') {
    openDataTab(node)
    return
  }

  if (node.type === 'function') {
    openFunctionTab(node)
    return
  }

  if (node.type === 'backup') {
    openBackupTab(node.databaseName, node.label)
  }
}

function handleNodeDoubleClick(node: TreeNodeData) {
  if (node.type === 'group' && node.groupKind === 'backups') {
    openBackupTab(node.databaseName)
    return
  }

  if (node.type === 'table' || node.type === 'view') {
    openDataTab(node)
    return
  }

  if (node.type === 'function') {
    openFunctionTab(node)
    return
  }

  if (node.type === 'group' && (node.children?.length ?? 0) === 0) {
    if (node.groupKind === 'backups') {
      openBackupTab(node.databaseName)
      return
    }

    openEmptyGroupTab(node)
    return
  }

  if (node.type === 'database') {
    openQueryTab(node.databaseName, t('workspace.databaseQueryTitle', { database: node.databaseName }))
    return
  }

  if (node.type === 'backup') {
    openBackupTab(node.databaseName, node.label)
  }
}

function handleNodeExpand(data: TreeNodeData) {
  expandedKeys.value.add(data.key)
  setCurrentNode(data)
  syncWorkspaceFromNode(data)
}

function handleNodeCollapse(data: TreeNodeData) {
  expandedKeys.value = new Set(
    Array.from(expandedKeys.value).filter((key) => key !== data.key && !key.startsWith(`${data.key}:`))
  )
  setCurrentNode(data, { keepExpanded: false })
  syncWorkspaceFromNode(data)

  if (currentNodeKey.value && currentNodeKey.value.startsWith(`${data.key}:`)) {
    currentNodeKey.value = data.key
    nextTick(() => {
      treeRef.value?.setCurrentKey(data.key)
    })
    syncWorkspaceFromNode(data)
  }
}

function getContextMenuItems(node: TreeNodeData): ContextMenuItem[] {
  if (node.type === 'database') {
    return [
      { key: 'create-table', label: t('workspace.createTableAction') },
      { key: 'open-query', label: t('workspace.openQueryWindow') },
      { key: 'rename-database', label: t('workspace.renameDatabaseTitle') },
      { key: 'delete-database', label: t('workspace.deleteDatabaseTitle') },
      { key: 'refresh', label: t('common.refresh') },
      { key: 'compare-structure', label: isChinese.value ? '结构对比' : 'Structure Compare' },
      { key: 'backup-database', label: t('backup.createDatabase') },
      { key: 'open-backups', label: t('workspace.group.backups') }
    ]
  }

  if (node.type === 'table' || node.type === 'view') {
    const items: ContextMenuItem[] = []
    items.push(
      { key: 'open-query', label: t('workspace.openQueryWindow') },
      { key: 'rename-table', label: t('workspace.renameObjectTitle') },
      { key: 'delete-table', label: t('workspace.deleteObjectTitle') },
      { key: 'refresh', label: t('common.refresh') }
    )
    if (node.type === 'table') {
      items.push({ key: 'compare-structure', label: isChinese.value ? '结构对比' : 'Structure Compare' })
    }
    items.push({ key: 'backup-table', label: t('backup.createTable') })
    return items
  }

  if (node.type === 'group') {
    const items: ContextMenuItem[] = []
    if (node.groupKind === 'tables') {
      items.push({ key: 'create-table', label: t('workspace.createTableAction') })
    }

    if (node.groupKind === 'backups') {
      items.push(
        { key: 'backup-database', label: t('backup.createDatabase') },
        { key: 'open-backups', label: t('workspace.group.backups') }
      )
    }

    items.push(
      { key: 'open-query', label: t('workspace.openQueryWindow') },
      { key: 'refresh', label: t('common.refresh') }
    )

    return items
  }

  if (node.type === 'backup') {
    return [
      { key: 'restore-backup', label: t('backup.restore') },
      { key: 'download-backup', label: t('backup.download') },
      { key: 'rename-backup', label: t('workspace.renameObjectTitle') },
      { key: 'delete-backup', label: t('workspace.deleteObjectTitle') },
      { key: 'open-backups', label: t('workspace.group.backups') },
      { key: 'refresh', label: t('common.refresh') }
    ]
  }

  return [{ key: 'refresh', label: t('common.refresh') }]
}

function handleNodeContextMenu(event: MouseEvent, node: TreeNodeData) {
  event.preventDefault()
  setCurrentNode(node)
  contextMenu.visible = true
  contextMenu.x = event.clientX + 8
  contextMenu.y = event.clientY + 8
  contextMenu.node = node
  contextMenu.items = getContextMenuItems(node)
}

function closeContextMenu() {
  contextMenu.visible = false
  contextMenu.node = null
  contextMenu.items = []
}

async function handleContextAction(action: ContextActionKey) {
  const node = contextMenu.node
  closeContextMenu()
  if (!node) {
    return
  }

  if (action === 'refresh') {
    if (node.type === 'table' || node.type === 'view') {
      refreshDataTab(node.databaseName, node.tableName || node.label)
    }

    await refreshExplorer()
    return
  }

  if (action === 'create-database') {
    openCreateDatabaseDialog()
    return
  }

  if (action === 'create-table') {
    openCreateTableDialog(node.databaseName)
    return
  }

  if (action === 'compare-structure') {
    openStructureCompareDialog(node)
    return
  }

  if (action === 'backup-database') {
    await createBackupForNode(node.databaseName)
    return
  }

  if (action === 'backup-table') {
    await createBackupForNode(node.databaseName, node.tableName || node.label)
    return
  }

  if (action === 'open-backups') {
    openBackupTab(node.databaseName, node.type === 'backup' ? node.label : '')
    return
  }

  if (action === 'open-query') {
    openQueryTab(node.databaseName)
    return
  }

  if (action === 'rename-database' && node.type === 'database') {
    await renameDatabase(node)
    return
  }

  if (action === 'delete-database' && node.type === 'database') {
    await deleteDatabase(node)
    return
  }

  if (action === 'rename-table' && (node.type === 'table' || node.type === 'view')) {
    await renameTable(node)
    return
  }

  if (action === 'delete-table' && (node.type === 'table' || node.type === 'view')) {
    await deleteTable(node)
    return
  }

  if (node.type === 'backup') {
    if (action === 'restore-backup') {
      await restoreBackupNode(node)
      return
    }

    if (action === 'download-backup') {
      await downloadBackupNode(node)
      return
    }

    if (action === 'rename-backup') {
      await renameBackupNode(node)
      return
    }

    if (action === 'delete-backup') {
      await deleteBackupNode(node)
    }
  }
}

async function createBackupForNode(databaseName: string, tableName = '') {
  const response = await request.post<{ taskId: string }>('/api/backup/create', {
    database: databaseName,
    tableName,
    compress: true
  })

  openBackupTab(databaseName)
  ElMessage.success(tableName ? t('backup.tableTaskQueued') : t('backup.taskQueued'))
  return response.taskId
}

async function restoreBackupNode(node: TreeNodeData) {
  const { value } = await ElMessageBox.prompt(t('backup.targetDatabasePrompt'), t('backup.restore'), {
    inputValue: node.databaseName
  })

  await request.post('/api/backup/restore', {
    database: node.databaseName,
    fileName: node.label,
    targetDatabase: value.trim()
  })

  openBackupTab(node.databaseName, node.label)
  ElMessage.success(t('backup.restoreQueued'))
}

async function renameBackupNode(node: TreeNodeData) {
  const { value } = await ElMessageBox.prompt(t('backup.renamePrompt'), t('workspace.renameObjectTitle'), {
    inputValue: node.label.replace(/\.sql(?:\.gz)?$/i, '')
  })

  await request.post('/api/backup/rename', {
    database: node.databaseName,
    fileName: node.label,
    newName: value.trim()
  })

  await refreshExplorer()
  openBackupTab(node.databaseName)
  ElMessage.success(t('backup.renameSuccess'))
}

async function deleteBackupNode(node: TreeNodeData) {
  await ElMessageBox.confirm(t('backup.deleteConfirm', { name: node.label }), t('workspace.deleteObjectTitle'))
  await request.post('/api/backup/delete', {
    database: node.databaseName,
    fileName: node.label
  })

  await refreshExplorer()
  openBackupTab(node.databaseName)
  ElMessage.success(t('workspace.objectDeleted'))
}

async function downloadBackupNode(node: TreeNodeData) {
  const response = await fetch(
    `/api/backup/download?database=${encodeURIComponent(node.databaseName)}&fileName=${encodeURIComponent(node.label)}`,
    {
      headers: {
        'X-Connection-Token': connectionStore.token
      }
    }
  )

  if (!response.ok) {
    throw new Error('backup download failed')
  }

  const blob = await response.blob()
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = node.label
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
  URL.revokeObjectURL(url)
}

function openCreateDatabaseDialog() {
  createDatabaseDialog.name = ''
  createDatabaseDialog.visible = true
}

function openCreateTableDialog(databaseName: string) {
  if (!databaseName) {
    ElMessage.warning(t('workspace.selectDatabaseFirst'))
    return
  }

  createTableDialog.database = databaseName
  createTableDialog.name = ''
  createTableDialog.sqlInput = ''
  createTableDialog.tableOptions = 'ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci'
  createTableDialog.columns = [
    buildDefaultColumnDraft({
      name: 'id',
      type: 'BIGINT',
      length: '20',
      primaryKey: true,
      notNull: true,
      autoIncrement: true,
      comment: 'Primary Key ID'
    }),
    buildDefaultColumnDraft({
      name: 'name',
      type: 'VARCHAR',
      length: '255',
      comment: 'Name'
    })
  ]
  normalizeCreateTableDialogColumns()
  createTableDialog.visible = true
}

function openSmartImportDialog() {
  if (!workspaceStore.activeDatabase) {
    ElMessage.warning(t('workspace.selectDatabaseFirst'))
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
  if (!workspaceStore.activeDatabase) {
    ElMessage.warning(t('workspace.selectDatabaseFirst'))
    return
  }

  smartImporting.value = true
  try {
    const payload = await parseSmartImportFile(file)
    if (payload.columns.length === 0) {
      ElMessage.warning(t('workspace.noImportHeaders'))
      return
    }

    const result = await request.post<AutoImportResponse>('/api/metadata/table/auto-import', {
      database: workspaceStore.activeDatabase,
      name: payload.tableName,
      columns: payload.columns,
      rows: payload.rows
    })

    expandedKeys.value.add(`db:${workspaceStore.activeDatabase}`)
    expandedKeys.value.add(`db:${workspaceStore.activeDatabase}:tables`)
    await refreshExplorer()

    const tableNode = findNodeByKey(
      explorerTree.value,
      `db:${workspaceStore.activeDatabase}:table:${result.tableName}`
    )

    if (tableNode) {
      setCurrentNode(tableNode)
      openDataTab(tableNode)
    }

    ElMessage.success(t('workspace.smartImportDone', { tableName: result.tableName, rowCount: result.rowCount }))
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Smart import failed.'
    ElMessage.error(message)
  } finally {
    smartImporting.value = false
  }
}

async function handleSmartExport() {
  if (!workspaceStore.activeDatabase || !workspaceStore.activeTable) {
    ElMessage.warning(t('workspace.selectTableFirst'))
    return
  }

  smartExporting.value = true

  try {
    const dataset = await fetchAllTableRows(workspaceStore.activeDatabase, workspaceStore.activeTable)
    if (dataset.length === 0) {
      ElMessage.info(t('workspace.smartExportEmpty'))
      return
    }

    await downloadExcel(dataset, buildSmartExportFilename(workspaceStore.activeDatabase, workspaceStore.activeTable))
    ElMessage.success(t('workspace.smartExportDone', { count: dataset.length }))
  } catch (error) {
    const message = error instanceof Error ? error.message : t('workspace.smartExportFailed')
    ElMessage.error(message)
  } finally {
    smartExporting.value = false
  }
}

async function fetchAllTableRows(databaseName: string, tableName: string) {
  const pageLimit = 500
  const dataset: Array<Record<string, unknown>> = []
  let offset = 0
  let total = Number.POSITIVE_INFINITY

  while (offset < total) {
    const response = await request.get<TableDataResponse>('/api/data/table', {
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
  const matrix = await readImportMatrix(file, t('workspace.noWorksheet'))
  return buildSmartImportPayloadFromMatrix(matrix, baseTableName)
}

function buildSmartImportPayloadFromMatrix(matrix: unknown[][], fallbackTableName: string): SmartImportPayload {
  if (matrix.length === 0) {
    throw new Error(t('workspace.emptyImportFile'))
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
    if (shouldPreferDiscreteCodeType(normalizedHeader)) {
      return 'INT'
    }

    return normalizedHeader === 'id' ? 'BIGINT' : 'BIGINT'
  }

  if (values.every((value) => isDecimalLikeValue(value))) {
    return 'DECIMAL(18,4)'
  }

  const maxLength = values.reduce<number>(
    (max, value) => Math.max(max, String(value ?? '').length),
    0
  )
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


function addColumnDraft() {
  createTableDialog.columns.push(buildDefaultColumnDraft())
  normalizeCreateTableDialogColumns()
}

function removeColumnDraft(columnId: string) {
  if (createTableDialog.columns.length <= 1) {
    return
  }

  createTableDialog.columns = createTableDialog.columns.filter((column) => column.id !== columnId)
  normalizeCreateTableDialogColumns()
}

function applyCreateTableSQL() {
  if (!createTableDialog.sqlInput.trim()) {
    ElMessage.warning(t('workspace.enterCreateTableSql'))
    return
  }

  try {
    const parsed = parseCreateTableSQLInput(createTableDialog.sqlInput)
    if (parsed.databaseName) {
      createTableDialog.database = parsed.databaseName
    }
    createTableDialog.name = parsed.tableName
    createTableDialog.tableOptions = parsed.tableOptions
    createTableDialog.columns = parsed.columns
    normalizeCreateTableDialogColumns()
    ElMessage.success(t('workspace.parseCreateTableSqlSuccess'))
  } catch (error) {
    const message = error instanceof Error ? error.message : t('workspace.invalidCreateTableSql')
    ElMessage.error(message)
  }
}

function updateSingleColumn(columnId: string, updater: (column: TableColumnDraft) => TableColumnDraft) {
  createTableDialog.columns = createTableDialog.columns.map((column) =>
    column.id === columnId ? updater(column) : column
  )
  normalizeCreateTableDialogColumns()
}

function handleColumnTypeChange(column: TableColumnDraft) {
  updateSingleColumn(column.id, (current) => normalizeColumnDraft(current))
}

function handleColumnLengthBlur(column: TableColumnDraft) {
  updateSingleColumn(column.id, (current) => normalizeColumnDraft(current))
}

function handleColumnDefaultBlur(column: TableColumnDraft) {
  const normalizedDefault = normalizeDefaultValueForType(column.type, column.defaultValue)
  updateSingleColumn(column.id, (current) => ({
    ...normalizeColumnDraft(current),
    defaultValue: normalizedDefault.value
  }))

  if (normalizedDefault.warningKey) {
    ElMessage.warning(t(normalizedDefault.warningKey))
  }
}

function handleColumnPrimaryKeyChange(column: TableColumnDraft) {
  updateSingleColumn(column.id, (current) => normalizeColumnDraft(current))
}

function handleColumnNotNullChange(column: TableColumnDraft) {
  updateSingleColumn(column.id, (current) => normalizeColumnDraft(current))
}

function handleColumnAutoIncrementChange(column: TableColumnDraft) {
  createTableDialog.columns = createTableDialog.columns.map((current) =>
    current.id === column.id
      ? normalizeColumnDraft(current)
      : { ...current, autoIncrement: false }
  )
  normalizeCreateTableDialogColumns()
}

async function submitCreateDatabase() {
  const name = createDatabaseDialog.name.trim()
  if (!name) {
    ElMessage.warning(t('workspace.databaseName'))
    return
  }

  createDatabaseDialog.loading = true
  try {
    await request.post('/api/metadata/database/create', { name })
    createDatabaseDialog.visible = false
    await refreshExplorer()
    ElMessage.success(t('workspace.createDatabaseSuccess', { name }))
  } finally {
    createDatabaseDialog.loading = false
  }
}

async function submitCreateTable() {
  try {
    normalizeCreateTableDialogColumns()
    validateCreateTableDialog()
  } catch (error) {
    ElMessage.warning(error instanceof Error ? error.message : t('workspace.invalidCreateConfig'))
    return
  }

  const columnsSQL = buildCreateTableColumnsSQL()
  if (!columnsSQL) {
    ElMessage.warning(t('workspace.addValidField'))
    return
  }

  createTableDialog.loading = true
  try {
    const tableName = createTableDialog.name.trim()
    await request.post('/api/metadata/table/create', {
      database: createTableDialog.database,
      name: tableName,
      columns: columnsSQL,
      options: createTableDialog.tableOptions.trim()
    })

    createTableDialog.visible = false
    workspaceStore.setActiveTable(createTableDialog.database, tableName)
    expandedKeys.value.add(`db:${createTableDialog.database}`)
    expandedKeys.value.add(`db:${createTableDialog.database}:tables`)
    await refreshExplorer()
    const node = findNodeByKey(
      explorerTree.value,
      `db:${createTableDialog.database}:table:${tableName}`
    )
    if (node) {
      setCurrentNode(node)
      openDataTab(node)
    }
    ElMessage.success(t('workspace.createTableSuccess', { name: tableName }))
  } finally {
    createTableDialog.loading = false
  }
}

async function renameDatabase(node: TreeNodeData) {
  const { value } = await ElMessageBox.prompt(t('workspace.renameDatabasePrompt'), t('workspace.renameDatabaseTitle'), {
    inputValue: node.databaseName
  })

  const nextName = value.trim()
  if (!nextName || nextName === node.databaseName) {
    return
  }

  await request.post('/api/metadata/database/rename', {
    oldName: node.databaseName,
    newName: nextName
  })

  if (workspaceStore.activeDatabase === node.databaseName) {
    workspaceStore.setActiveDatabase(nextName)
  }

  await refreshExplorer()
  ElMessage.success(t('workspace.databaseRenamed'))
}

async function deleteDatabase(node: TreeNodeData) {
  await ElMessageBox.confirm(t('workspace.deleteDatabaseConfirm', { name: node.databaseName }), t('workspace.deleteDatabaseTitle'), {
    type: 'warning'
  })

  await request.post('/api/metadata/database/delete', {
    name: node.databaseName
  })

  tabs.value = tabs.value.filter((tab) => tab.databaseName !== node.databaseName)
  if (!tabs.value.find((tab) => tab.id === activeTabId.value)) {
    activeTabId.value = tabs.value[0]?.id ?? ''
  }

  if (workspaceStore.activeDatabase === node.databaseName) {
    workspaceStore.resetWorkspace()
  }

  await refreshExplorer()
  ElMessage.success(t('workspace.databaseDeleted'))
}

async function renameTable(node: TreeNodeData) {
  const { value } = await ElMessageBox.prompt(t('workspace.renameObjectPrompt'), t('workspace.renameObjectTitle'), {
    inputValue: node.tableName ?? node.label
  })

  const nextName = value.trim()
  const currentName = node.tableName ?? node.label
  if (!nextName || nextName === currentName) {
    return
  }

  await request.post('/api/metadata/table/rename', {
    database: node.databaseName,
    oldName: currentName,
    newName: nextName
  })

  tabs.value = tabs.value.map((tab) => {
    if (tab.databaseName !== node.databaseName || tab.tableName !== currentName) {
      return tab
    }

    if (tab.kind === 'data') {
      return {
        ...tab,
        tableName: nextName,
        title: nextName,
        id: `data:${node.databaseName}:${nextName}`
      }
    }

    return {
      ...tab,
      tableName: nextName
    }
  })

  workspaceStore.setActiveTable(node.databaseName, nextName)
  await refreshExplorer()
  ElMessage.success(t('workspace.objectRenamed'))
}

async function deleteTable(node: TreeNodeData) {
  const currentName = node.tableName ?? node.label
  await ElMessageBox.confirm(t('workspace.deleteObjectConfirm', { name: currentName }), t('workspace.deleteObjectTitle'), {
    type: 'warning'
  })

  await request.post('/api/metadata/table/delete', {
    database: node.databaseName,
    name: currentName
  })

  tabs.value = tabs.value.filter(
    (tab) => !(tab.databaseName === node.databaseName && tab.tableName === currentName)
  )
  if (!tabs.value.find((tab) => tab.id === activeTabId.value)) {
    activeTabId.value = tabs.value[0]?.id ?? ''
  }

  if (workspaceStore.activeTable === currentName) {
    workspaceStore.clearActiveTable()
  }

  await refreshExplorer()
  ElMessage.success(t('workspace.objectDeleted'))
}

async function createNewQueryTab() {
  if (!workspaceStore.activeDatabase) {
    ElMessage.warning(t('workspace.selectDatabaseFirst'))
    return
  }

  openQueryTab(workspaceStore.activeDatabase)
}

function handleTabChange(tabId: string | number) {
  const tab = tabs.value.find((item) => item.id === tabId)
  if (!tab) {
    return
  }

  if (tab.databaseName && tab.tableName) {
    workspaceStore.setActiveTable(tab.databaseName, tab.tableName)
    currentNodeKey.value = `db:${tab.databaseName}:table:${tab.tableName}`
  } else if (tab.databaseName) {
    workspaceStore.setActiveDatabase(tab.databaseName)
  }

  nextTick(() => restoreCurrentTreeSelection())
}

function closeTab(tabId: string | number) {
  const index = tabs.value.findIndex((tab) => tab.id === tabId)
  if (index < 0) {
    return
  }

  const wasActive = activeTabId.value === tabId
  tabs.value.splice(index, 1)
  if (wasActive) {
    activeTabId.value = tabs.value[Math.max(0, index - 1)]?.id ?? tabs.value[0]?.id ?? ''
  }
}

async function refreshExplorer() {
  await loadExplorerTree()
}

async function refreshConnectionMeta() {
  if (!connectionStore.hasConnection) {
    ElMessage.warning(t('workspace.noConnectionValid'))
    await router.push('/')
    return
  }

  await refreshExplorer()
}

async function copyConnectionId() {
  if (!connectionStore.token) {
    return
  }

  await navigator.clipboard.writeText(connectionStore.token)
  ElMessage.success(t('workspace.connectionCopied'))
}

function handleDocumentClick() {
  closeContextMenu()
}

watch(
  () => locale.value,
  () => {
    tabs.value = tabs.value.map((tab) => {
      if (tab.kind === 'empty-group') {
        return {
          ...tab,
          title: `${tab.databaseName}/${getGroupLabel(tab.groupKind)}`,
          groupLabel: getGroupLabel(tab.groupKind),
          description: getGroupEmptyDescription(tab.databaseName, tab.groupKind)
        }
      }

      if (tab.kind === 'query' && /^(Query|查询) \d+$/.test(tab.title)) {
        const match = tab.title.match(/(\d+)/)
        const index = match ? Number(match[1]) : tabs.value.filter((item) => item.kind === 'query').indexOf(tab) + 1
        return {
          ...tab,
          title: t('workspace.queryTitle', { index })
        }
      }

      if (tab.kind === 'security') {
        return {
          ...tab,
          title: isChinese.value ? '用户管理' : 'User Security'
        }
      }

      if (tab.kind === 'backup') {
        return {
          ...tab,
          title: `${tab.databaseName}/${getGroupLabel('backups')}`
        }
      }

      return tab
    })
  }
)

watch(
  () => connectionStore.token,
  (token) => {
    if (!token) {
      tabs.value = []
      activeTabId.value = ''
      explorerTree.value = []
      currentNodeKey.value = ''
      expandedKeys.value = new Set()
      workspaceStore.resetWorkspace()
    }
  },
  { immediate: true }
)

onMounted(async () => {
  document.addEventListener('click', handleDocumentClick)
  if (!connectionStore.hasConnection) {
    await router.push('/')
    return
  }

  await loadExplorerTree()
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentClick)
})
</script>

<style scoped>
.smart-import-input {
  display: none;
}

.workspace-aside {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 24px;
  border-right: 1px solid var(--devops-border-light);
  background: var(--devops-bg-panel-soft);
}

.aside-card {
  position: relative;
  padding: 18px;
  border-radius: var(--devops-radius-lg);
}

.aside-card--explorer {
  flex: 1;
  overflow: hidden;
}

.section-heading--card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.mini-button--ghost {
  min-width: 84px;
}

.explorer-heading {
  align-items: center;
}

.explorer-toolbar {
  display: flex;
  gap: 8px;
}

.toolbar-button--capsule {
  min-width: 36px;
  width: 36px;
  height: 36px;
  padding: 0;
}

.token-pill--interactive {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.connection-token-card {
  min-height: 72px;
  padding: 14px 16px;
  border-radius: var(--devops-radius-lg);
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: var(--devops-success);
  box-shadow: 0 0 0 6px rgba(103, 194, 58, 0.14);
}

.token-copy {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 4px;
}

.connection-id,
.connection-caption {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.connection-id {
  color: var(--devops-text-primary);
  font-size: 13px;
  font-weight: 600;
}

.connection-caption {
  color: var(--devops-text-secondary);
  font-size: 12px;
}

.explorer-tree {
  padding-right: 4px;
  overflow: hidden;
}

:deep(.explorer-tree .el-tree-node__content) {
  min-height: 52px;
  height: auto;
  margin: 4px 0;
  padding: 8px 12px 8px 6px;
  align-items: flex-start;
}

.tree-node {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  width: 100%;
  min-width: 0;
  padding: 2px 0;
}

.tree-node__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 28px;
  width: 28px;
  height: 28px;
  margin-top: 2px;
  border-radius: 8px;
  background: var(--devops-primary-soft);
  color: var(--devops-primary);
}

.tree-node__meta {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 3px;
}

.tree-node__title,
.tree-node__subtitle {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tree-node__title {
  color: var(--devops-text-primary);
  font-size: 13px;
  font-weight: 600;
  line-height: 1.4;
}

.tree-node__subtitle {
  color: var(--devops-text-secondary);
  font-size: 11px;
  line-height: 1.35;
}

.security-entry {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border-radius: var(--devops-radius-lg);
  color: inherit;
  text-align: left;
  cursor: pointer;
  transition: border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease;
}

.security-entry:hover {
  border-color: #bfdcff;
  box-shadow: var(--devops-shadow-sm);
  transform: translateY(-1px);
}

.security-entry__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: var(--devops-primary-soft);
  color: var(--devops-primary);
}

.security-entry__copy {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 2px;
}

.security-entry__copy strong,
.security-entry__copy small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.security-entry__copy strong {
  color: var(--devops-text-primary);
  font-size: 14px;
}

.security-entry__copy small {
  color: var(--devops-text-secondary);
  font-size: 12px;
}

.workspace-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 20px;
  padding: 24px 24px 20px;
}

:deep(.workspace-header.el-header) {
  height: auto !important;
  padding: 24px 24px 20px !important;
}

.workspace-header__intro {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 16px;
}

.workspace-breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--devops-text-secondary);
  font-size: 13px;
}

.brand-block {
  gap: 10px;
}

.workspace-context-pill {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  width: fit-content;
  max-width: min(100%, 780px);
  padding: 14px 16px;
  border-radius: var(--devops-radius-lg);
}

.workspace-context-pill__active-text,
.workspace-context-pill__hint-text {
  margin: 0;
  color: var(--devops-text-regular);
  font-size: 13px;
  line-height: 1.6;
}

.workspace-context-pill__active-text {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.workspace-context-pill__hint-text {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.workspace-overview {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.workspace-overview__card {
  padding: 16px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-lg);
  background: var(--devops-bg-panel);
  box-shadow: var(--devops-shadow-xs);
}

.workspace-overview__label {
  display: block;
  color: var(--devops-text-secondary);
  font-size: 12px;
  margin-bottom: 8px;
}

.workspace-overview__value {
  display: block;
  color: var(--devops-text-primary);
  font-size: 16px;
  font-weight: 600;
  line-height: 1.4;
  word-break: break-word;
}

.workspace-actions-card {
  flex: 0 1 auto;
  width: fit-content;
  max-width: 100%;
  padding: 12px;
}

.workspace-actions-group {
  display: flex;
  flex-wrap: nowrap;
  gap: 8px;
  justify-content: flex-end;
  align-items: center;
  width: 100%;
}

.workspace-action-button {
  min-width: 0;
  padding: 0 12px;
  font-size: 12px;
  white-space: nowrap;
}

.workspace-action-button--primary {
  min-width: 0;
}

.workspace-language-button {
  min-width: 56px;
}

.empty-group-stage {
  min-height: 420px;
}

.empty-group-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 20px;
}

.workspace-tabs {
  width: 100%;
}

.workspace-tab-label {
  display: flex;
  min-width: 0;
  max-width: min(220px, 24vw);
  flex-direction: column;
  gap: 4px;
  padding: 6px 0;
  text-align: left;
}

.workspace-tab-label__title,
.workspace-tab-label__meta {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workspace-tab-label__title {
  font-size: 13px;
  font-weight: 600;
}

.workspace-tab-label__meta {
  color: var(--devops-text-secondary);
  font-size: 11px;
}

.context-menu {
  position: fixed;
  z-index: 3000;
  display: inline-flex;
  flex-direction: column;
  gap: 4px;
  min-width: 160px;
  padding: 6px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-md);
  background: var(--devops-bg-panel);
  box-shadow: var(--devops-shadow-md);
}

.context-menu__item {
  min-height: 34px;
  padding: 0 12px;
  border: none;
  border-radius: var(--devops-radius-sm);
  background: transparent;
  color: var(--devops-text-regular);
  font-size: 13px;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
}

.context-menu__item:hover {
  background: var(--devops-bg-hover);
  color: var(--devops-primary);
}

.create-table-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.create-table-sql-import {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.create-table-sql-import__actions {
  display: flex;
  justify-content: flex-end;
}

.create-table-sql-import :deep(.el-textarea__inner),
.create-table-sql-preview__input :deep(.el-textarea__inner) {
  min-height: 160px !important;
  font-family: "JetBrains Mono", Consolas, "Courier New", monospace;
  font-size: 12px;
  line-height: 1.6;
  background: #fff !important;
}

.column-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.column-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 16px;
}

.column-card {
  padding: 16px;
}

.column-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  color: var(--devops-text-secondary);
  font-size: 13px;
  font-weight: 600;
}

.column-card__grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0 16px;
}

.column-card__comment {
  grid-column: span 2;
}

.column-card__switches {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
}

@media (max-width: 1280px) {
  .workspace-header {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
  }

  .workspace-actions-card {
    width: 100%;
    flex-basis: auto;
  }

  .workspace-actions-group {
    justify-content: flex-start;
  }

  .column-card__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .column-card__comment {
    grid-column: span 2;
  }
}

@media (max-width: 960px) {
  .workspace-aside {
    width: 100% !important;
    border-right: none;
    border-bottom: 1px solid var(--devops-border-light);
  }

  .workspace-header {
    padding: 20px 18px 18px;
  }

  :deep(.workspace-header.el-header) {
    padding: 20px 18px 18px !important;
  }

  .workspace-overview {
    grid-template-columns: 1fr;
  }

  .workspace-actions-group {
    width: 100%;
    flex-wrap: wrap;
    justify-content: flex-start;
  }

  .workspace-action-button,
  .workspace-action-button--primary,
  .workspace-language-button {
    min-width: 0;
    flex: 1 1 calc(33.333% - 8px);
  }

  .workspace-tab-label {
    max-width: min(160px, 32vw);
  }
}

@media (max-width: 720px) {
  .workspace-aside {
    padding: 16px;
  }

  .workspace-header {
    padding: 16px 12px;
  }

  :deep(.workspace-header.el-header) {
    padding: 16px 12px !important;
  }

  .workspace-context-pill {
    width: 100%;
    max-width: 100%;
  }

  .workspace-actions-group {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .workspace-action-button,
  .workspace-action-button--primary,
  .workspace-language-button {
    width: 100%;
  }

  .create-table-grid,
  .column-card__grid {
    grid-template-columns: 1fr;
  }

  .column-card__comment {
    grid-column: span 1;
  }

  .empty-group-actions {
    flex-direction: column;
  }
}

@media (max-width: 480px) {
  .aside-card,
  .workspace-actions-card,
  .workspace-context-pill,
  .workspace-overview__card,
  .connection-token-card,
  .security-entry {
    border-radius: var(--devops-radius-md);
  }

  .workspace-actions-group {
    grid-template-columns: 1fr;
  }

  .workspace-tab-label {
    max-width: min(120px, 56vw);
  }
}
</style>
