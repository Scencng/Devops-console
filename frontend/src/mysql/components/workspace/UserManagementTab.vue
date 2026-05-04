<template>
  <section class="glass-subpanel security-stage">
    <header class="security-stage__header">
      <div class="security-stage__heading">
        <span class="security-stage__eyebrow">{{ text('title') }}</span>
        <h2>{{ text('subtitle') }}</h2>
        <p>{{ text('desc') }}</p>
      </div>

      <div class="security-stage__meta">
        <div class="security-stat">
          <span>{{ text('serverVersion') }}</span>
          <strong>{{ overview?.capabilities.version || '-' }}</strong>
        </div>
        <div class="security-stat">
          <span>{{ text('rolesSupport') }}</span>
          <strong>{{ overview?.capabilities.supportsRoles ? text('supported') : text('unsupported') }}</strong>
        </div>
      </div>
    </header>

    <div class="security-stage__toolbar">
      <el-button class="soft-button" :loading="loading" @click="loadOverview">
        {{ text('refresh') }}
      </el-button>
      <el-button v-if="canManageSecurity" class="soft-button" @click="openCreateDialog('user')">
        {{ text('createUser') }}
      </el-button>
      <el-button v-if="supportsRoles && canManageSecurity" class="soft-button" @click="openCreateDialog('role')">
        {{ text('createRole') }}
      </el-button>
    </div>

    <div v-if="overviewLoadError" class="security-load-error">
      <span>{{ overviewLoadError }}</span>
      <el-button class="soft-button" :loading="loading" @click="loadOverview(true)">
        {{ text('retry') }}
      </el-button>
    </div>

    <div v-else-if="loading && !overview" class="security-overview-skeleton">
      <el-skeleton animated>
        <template #template>
          <div class="security-overview-skeleton__stats">
            <el-skeleton-item variant="rect" style="width: 180px; height: 88px;" />
            <el-skeleton-item variant="rect" style="width: 180px; height: 88px;" />
          </div>
          <el-skeleton-item variant="rect" style="width: 100%; height: 44px;" />
          <el-skeleton-item variant="rect" style="width: 100%; height: 360px;" />
          <el-skeleton-item variant="rect" style="width: 100%; height: 180px;" />
        </template>
      </el-skeleton>
    </div>

    <el-tabs v-else v-model="activePanel" class="security-panels">
      <el-tab-pane :label="text('usersTab')" name="users">
        <div class="security-table-wrap">
        <el-table v-loading="loading" :data="overview?.users || []" class="security-table" table-layout="auto">
          <el-table-column prop="user" :label="text('user')" min-width="140" />
          <el-table-column prop="host" :label="text('host')" min-width="140" />
          <el-table-column
            prop="plugin"
            :label="text('plugin')"
            min-width="220"
            class-name="security-table__column--plugin"
            label-class-name="security-table__column--plugin-header"
          />
          <el-table-column :label="text('status')" min-width="120">
            <template #default="{ row }">
              <el-tag
                :type="row.locked ? 'info' : 'success'"
                effect="light"
                class="security-status-tag"
              >
                {{ row.locked ? text('locked') : text('active') }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="privilegeSummary" :label="text('privileges')" min-width="220">
            <template #default="{ row }">
              <el-button
                link
                type="primary"
                class="security-privilege-link"
                @click="loadPrivilegeDetails(row.user, row.host, 'user')"
              >
                {{ formatPrivilegeSummary(row.privilegeSummary) || '-' }}
              </el-button>
            </template>
          </el-table-column>
          <el-table-column
            :label="text('actions')"
            fixed="right"
            width="340"
            class-name="security-table__column--actions"
            label-class-name="security-table__column--actions-header"
          >
            <template #default="{ row }">
              <div class="security-table__actions">
                <el-button v-if="canManageSecurity" link type="primary" @click="openEditDialog(row.user, row.host, 'user')">{{ text('edit') }}</el-button>
                <el-button v-if="canManageSecurity" link type="primary" @click="openCloneDialog(row.user, row.host, 'user')">{{ text('clone') }}</el-button>
                <el-button v-if="canManageSecurity" link type="warning" @click="revokeAll(row.user, row.host, 'user')">{{ text('revokeAll') }}</el-button>
                <el-button v-if="canManageSecurity" link type="danger" @click="removePrincipal(row.user, row.host, 'user')">{{ text('delete') }}</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
        </div>
      </el-tab-pane>

      <el-tab-pane v-if="supportsRoles" :label="text('rolesTab')" name="roles">
        <div class="security-table-wrap">
        <el-table v-loading="loading" :data="overview?.roles || []" class="security-table" table-layout="auto">
          <el-table-column prop="user" :label="text('roleName')" min-width="180" />
          <el-table-column prop="host" :label="text('host')" min-width="140" />
          <el-table-column prop="privilegeSummary" :label="text('privileges')" min-width="220">
            <template #default="{ row }">
              <el-button
                link
                type="primary"
                class="security-privilege-link"
                @click="loadPrivilegeDetails(row.user, row.host, 'role')"
              >
                {{ formatPrivilegeSummary(row.privilegeSummary) || '-' }}
              </el-button>
            </template>
          </el-table-column>
          <el-table-column
            :label="text('actions')"
            fixed="right"
            width="240"
            class-name="security-table__column--actions"
            label-class-name="security-table__column--actions-header"
          >
            <template #default="{ row }">
              <div class="security-table__actions">
                <el-button v-if="canManageSecurity" link type="primary" @click="openEditDialog(row.user, row.host, 'role')">{{ text('edit') }}</el-button>
                <el-button v-if="canManageSecurity" link type="warning" @click="revokeAll(row.user, row.host, 'role')">{{ text('revokeAll') }}</el-button>
                <el-button v-if="canManageSecurity" link type="danger" @click="removePrincipal(row.user, row.host, 'role')">{{ text('delete') }}</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
        </div>
      </el-tab-pane>
    </el-tabs>

    <section v-loading="privilegeDetailLoading" class="security-detail-card">
      <div class="security-detail-card__header">
        <div class="security-detail-card__copy">
          <span class="section-title">{{ text('privilegeDetailTitle') }}</span>
          <strong v-if="selectedPrivilegePrincipal">{{ selectedPrivilegePrincipal }}</strong>
          <p v-else>{{ text('clickPrivilegeHint') }}</p>
        </div>
        <el-button
          class="soft-button security-detail-card__copy-button"
          :disabled="!selectedPrivilegeDetail"
          @click="copyPrivilegeDetails"
        >
          {{ text('copyPrivileges') }}
        </el-button>
      </div>

      <div v-if="privilegeDetailLoading" class="security-detail-placeholder">
        {{ text('privilegeDetailLoading') }}
      </div>

      <div v-else-if="privilegeDetailError" class="security-detail-placeholder security-detail-placeholder--error">
        {{ privilegeDetailError }}
      </div>

      <div v-else-if="selectedPrivilegeDetail" class="security-detail-grid">
        <section class="security-detail-group">
          <h4>{{ text('globalPrivileges') }}</h4>
          <div v-if="selectedPrivilegeDetail.globalPrivileges.length" class="security-detail-tags">
            <span v-for="item in selectedPrivilegeDetail.globalPrivileges" :key="`global-${item}`" class="security-detail-tag">{{ item }}</span>
          </div>
          <p v-else class="security-detail-empty">{{ text('none') }}</p>
        </section>

        <section class="security-detail-group">
          <h4>{{ text('schemaPrivileges') }}</h4>
          <div v-if="selectedPrivilegeDetail.schemaPrivileges.length" class="security-detail-list">
            <div
              v-for="item in selectedPrivilegeDetail.schemaPrivileges"
              :key="`schema-${item.database}`"
              class="security-detail-row"
            >
              <strong>{{ item.database }}</strong>
              <span>{{ formatPrivilegeList(item.privileges) }}</span>
            </div>
          </div>
          <p v-else class="security-detail-empty">{{ text('none') }}</p>
        </section>

        <section class="security-detail-group">
          <h4>{{ text('tablePrivileges') }}</h4>
          <div v-if="selectedPrivilegeDetail.tablePrivileges.length" class="security-detail-list">
            <div
              v-for="item in selectedPrivilegeDetail.tablePrivileges"
              :key="`table-${item.database}-${item.table}`"
              class="security-detail-row"
            >
              <strong>{{ item.database }}.{{ item.table }}</strong>
              <span>{{ formatPrivilegeList(item.privileges) }}</span>
            </div>
          </div>
          <p v-else class="security-detail-empty">{{ text('none') }}</p>
        </section>

        <section class="security-detail-group">
          <h4>{{ text('columnPrivileges') }}</h4>
          <div v-if="selectedPrivilegeDetail.columnPrivileges.length" class="security-detail-list">
            <div
              v-for="item in selectedPrivilegeDetail.columnPrivileges"
              :key="`column-${item.database}-${item.table}-${item.column}`"
              class="security-detail-row"
            >
              <strong>{{ item.database }}.{{ item.table }}.{{ item.column }}</strong>
              <span>{{ formatPrivilegeList(item.privileges) }}</span>
            </div>
          </div>
          <p v-else class="security-detail-empty">{{ text('none') }}</p>
        </section>

        <section v-if="selectedPrivilegeDetail.roles.length" class="security-detail-group">
          <h4>{{ text('rolesGranted') }}</h4>
          <div class="security-detail-tags">
            <span
              v-for="item in selectedPrivilegeDetail.roles"
              :key="`role-${item.user}-${item.host}`"
              class="security-detail-tag"
            >
              {{ item.user }}@{{ item.host }}
            </span>
          </div>
        </section>

        <section v-if="selectedPrivilegeDetail.grantStatements.length" class="security-detail-group security-detail-group--full">
          <h4>{{ text('grantStatements') }}</h4>
          <pre class="security-detail-code">{{ selectedPrivilegeDetail.grantStatements.join('\n') }}</pre>
        </section>
      </div>

      <div v-else class="security-detail-placeholder">
        {{ text('clickPrivilegeHint') }}
      </div>
    </section>

    <el-dialog v-model="editor.visible" :title="editorTitle" width="980px" class="security-dialog">
      <el-form label-position="top">
        <div class="security-form-grid">
          <el-form-item :label="text('kind')">
            <el-input :model-value="editor.kind === 'role' ? text('rolesTab') : text('usersTab')" readonly />
          </el-form-item>
          <el-form-item :label="text('host')">
            <el-input v-model="editor.form.host" placeholder="%" />
          </el-form-item>
          <el-form-item :label="editor.kind === 'role' ? text('roleName') : text('user')">
            <el-input v-model="editor.form.user" />
          </el-form-item>
          <el-form-item v-if="editor.kind === 'user'" :label="text('password')">
            <el-input v-model="editor.form.password" type="password" show-password />
          </el-form-item>
        </div>

        <div v-if="editor.kind === 'user'" class="security-switches">
          <el-checkbox v-model="editor.form.locked">{{ text('locked') }}</el-checkbox>
          <el-checkbox v-model="editor.form.passwordExpired">{{ text('passwordExpired') }}</el-checkbox>
          <el-checkbox v-model="editor.form.passwordChanged">{{ text('changePassword') }}</el-checkbox>
        </div>

        <div class="security-preset-row">
          <span>{{ text('presets') }}</span>
          <div class="security-preset-actions">
            <el-button class="soft-button" @click="applyPreset('admin')">{{ text('presetAdmin') }}</el-button>
            <el-button class="soft-button" @click="applyPreset('readonly')">{{ text('presetReadonly') }}</el-button>
            <el-button class="soft-button" @click="applyPreset('custom')">{{ text('presetCustom') }}</el-button>
          </div>
        </div>

        <el-form-item :label="text('globalPrivileges')">
          <el-checkbox-group v-model="editor.form.globalPrivileges" class="security-checkbox-grid">
            <el-checkbox v-for="item in globalPrivilegeOptions" :key="item" :label="item">{{ item }}</el-checkbox>
          </el-checkbox-group>
        </el-form-item>

        <section class="security-scope-card">
          <div class="security-scope-card__header">
            <span>{{ text('schemaPrivileges') }}</span>
            <el-button class="soft-button" @click="addScopeRow('schema')">{{ text('addRule') }}</el-button>
          </div>
          <div v-for="(item, index) in editor.form.schemaPrivileges" :key="`schema-${index}`" class="security-scope-row">
            <el-input v-model="item.database" :placeholder="text('databasePlaceholder')" />
            <el-select v-model="item.privileges" multiple collapse-tags collapse-tags-tooltip>
              <el-option v-for="option in schemaPrivilegeOptions" :key="option" :label="option" :value="option" />
            </el-select>
            <el-button link type="danger" @click="removeScopeRow('schema', index)">{{ text('delete') }}</el-button>
          </div>
        </section>

        <section class="security-scope-card">
          <div class="security-scope-card__header">
            <span>{{ text('tablePrivileges') }}</span>
            <el-button class="soft-button" @click="addScopeRow('table')">{{ text('addRule') }}</el-button>
          </div>
          <div v-for="(item, index) in editor.form.tablePrivileges" :key="`table-${index}`" class="security-scope-row security-scope-row--triple">
            <el-input v-model="item.database" :placeholder="text('databasePlaceholder')" />
            <el-input v-model="item.table" :placeholder="text('tablePlaceholder')" />
            <el-select v-model="item.privileges" multiple collapse-tags collapse-tags-tooltip>
              <el-option v-for="option in tablePrivilegeOptions" :key="option" :label="option" :value="option" />
            </el-select>
            <el-button link type="danger" @click="removeScopeRow('table', index)">{{ text('delete') }}</el-button>
          </div>
        </section>

        <section class="security-scope-card">
          <div class="security-scope-card__header">
            <span>{{ text('columnPrivileges') }}</span>
            <el-button class="soft-button" @click="addScopeRow('column')">{{ text('addRule') }}</el-button>
          </div>
          <div v-for="(item, index) in editor.form.columnPrivileges" :key="`column-${index}`" class="security-scope-row security-scope-row--quad">
            <el-input v-model="item.database" :placeholder="text('databasePlaceholder')" />
            <el-input v-model="item.table" :placeholder="text('tablePlaceholder')" />
            <el-input v-model="item.column" :placeholder="text('columnPlaceholder')" />
            <el-select v-model="item.privileges" multiple collapse-tags collapse-tags-tooltip>
              <el-option v-for="option in columnPrivilegeOptions" :key="option" :label="option" :value="option" />
            </el-select>
            <el-button link type="danger" @click="removeScopeRow('column', index)">{{ text('delete') }}</el-button>
          </div>
        </section>

        <el-form-item v-if="supportsRoles" :label="text('bindRoles')">
          <el-select v-model="editor.form.roles" multiple value-key="key" collapse-tags collapse-tags-tooltip class="security-role-select">
            <el-option
              v-for="role in roleOptions"
              :key="role.key"
              :label="`${role.user}@${role.host}`"
              :value="{ user: role.user, host: role.host, key: role.key }"
            />
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="editor.visible = false">{{ text('cancel') }}</el-button>
        <el-button type="primary" :loading="editor.saving" @click="submitEditor">{{ text('save') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="cloneDialog.visible" :title="text('cloneUser')" width="520px">
      <el-form label-position="top">
        <el-form-item :label="text('sourceAccount')">
          <el-input :model-value="`${cloneDialog.sourceUser}@${cloneDialog.sourceHost}`" readonly />
        </el-form-item>
        <el-form-item :label="cloneDialog.kind === 'role' ? text('roleName') : text('user')">
          <el-input v-model="cloneDialog.targetUser" />
        </el-form-item>
        <el-form-item :label="text('host')">
          <el-input v-model="cloneDialog.targetHost" placeholder="%" />
        </el-form-item>
        <el-form-item v-if="cloneDialog.kind === 'user'" :label="text('password')">
          <el-input v-model="cloneDialog.password" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="cloneDialog.visible = false">{{ text('cancel') }}</el-button>
        <el-button type="primary" :loading="cloneDialog.saving" @click="submitClone">{{ text('confirm') }}</el-button>
      </template>
    </el-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { usePermissionStore } from '@/stores/permissionStore.js'
import { useI18n } from '@/mysql/utils/i18n'
import request from '@/mysql/utils/request'
import {
  fetchSecurityOverview,
  getSecurityOverviewCache,
  type PrincipalSummary,
  type PrincipalKind,
  type SecurityOverviewResponse
} from '@/mysql/utils/security-overview'

const emit = defineEmits<{
  (event: 'ready'): void
}>()

interface PrincipalRef {
  user: string
  host: string
  key?: string
}

interface ScopePrivilege {
  database: string
  table?: string
  column?: string
  privileges: string[]
}

interface PrincipalDetailResponse extends PrincipalSummary {
  globalPrivileges: string[]
  schemaPrivileges: ScopePrivilege[]
  tablePrivileges: ScopePrivilege[]
  columnPrivileges: ScopePrivilege[]
  roles: PrincipalRef[]
  grantStatements: string[]
}

const { isChinese } = useI18n()
const permissionStore = usePermissionStore()
const canManageSecurity = computed(() => permissionStore.hasPerm('mysql:security:manage'))

const messages = {
  zh: {
    title: '\u7528\u6237\u4e0e\u6743\u9650',
    subtitle: 'MySQL 用户权限管理',
    desc: '可视化管理用户、角色与授权范围，并在变更后自动刷新',
    serverVersion: '服务版本',
    rolesSupport: '角色支持',
    supported: '已支持',
    unsupported: '不支持',
    refresh: '刷新',
    retry: '重试',
    createUser: '创建用户',
    createRole: '创建角色',
    usersTab: '用户',
    rolesTab: '角色',
    user: '用户名',
    roleName: '角色名',
    host: '主机/IP',
    plugin: '认证插件',
    status: '状态',
    active: '正常',
    locked: '已锁定',
    privileges: '权限概览',
    actions: '操作',
    edit: '编辑',
    clone: '复制用户',
    revokeAll: '回收权限',
    delete: '删除',
    kind: '类型',
    password: '密码',
    passwordExpired: '密码过期',
    changePassword: '修改密码',
    presets: '权限模板',
    presetAdmin: '管理员',
    presetReadonly: '只读',
    presetCustom: '清空',
    globalPrivileges: '全局权限',
    schemaPrivileges: '库权限',
    tablePrivileges: '表权限',
    columnPrivileges: '列权限',
    addRule: '新增规则',
    databasePlaceholder: '数据库',
    tablePlaceholder: '数据表',
    columnPlaceholder: '列名',
    bindRoles: '绑定角色',
    cancel: '取消',
    save: '保存',
    cloneUser: '复制用户',
    sourceAccount: '源账号',
    confirm: '确认',
    deleteConfirm: '确认删除 {target} 吗？',
    revokeConfirm: '确认回收 {target} 的全部权限吗？',
    saveSuccess: '保存成功',
    deleteSuccess: '删除成功',
    revokeSuccess: '权限已回收',
    cloneSuccess: '复制成功',
    privilegeDetailTitle: '权限详情',
    clickPrivilegeHint: '点击权限概览查看详情',
    privilegeDetailLoading: '正在加载权限详情',
    privilegeDetailLoadFailed: '权限详情加载失败，请重试',
    overviewLoadFailed: '用户与权限数据加载失败，请重试',
    overviewLoadTimeout: '用户与权限数据加载超时，请重试',
    overviewUnauthorized: '当前账号没有访问 MySQL 用户与权限信息的权限，请联系管理员',
    copyPrivileges: '复制权限列表',
    copyPrivilegesSuccess: '权限列表已复制',
    rolesGranted: '已授予角色',
    grantStatements: '授权语句',
    none: '无'
  },
  en: {
    title: '\u7528\u6237\u4e0e\u6743\u9650',
    subtitle: 'MySQL user and privilege management',
    desc: 'Manage users, roles, and grants with a structured editor.',
    serverVersion: 'Server Version',
    rolesSupport: 'Role Support',
    supported: 'Supported',
    unsupported: 'Unsupported',
    refresh: 'Refresh',
    retry: 'Retry',
    createUser: 'Create User',
    createRole: 'Create Role',
    usersTab: 'Users',
    rolesTab: 'Roles',
    user: 'User',
    roleName: 'Role Name',
    host: 'Host/IP',
    plugin: 'Plugin',
    status: 'Status',
    active: 'Active',
    locked: 'Locked',
    privileges: 'Privilege Summary',
    actions: 'Actions',
    edit: 'Edit',
    clone: 'Clone',
    revokeAll: 'Revoke All',
    delete: 'Delete',
    kind: 'Kind',
    password: 'Password',
    passwordExpired: 'Password Expired',
    changePassword: 'Change Password',
    presets: 'Presets',
    presetAdmin: 'Admin',
    presetReadonly: 'Read Only',
    presetCustom: 'Clear',
    globalPrivileges: 'Global Privileges',
    schemaPrivileges: 'Schema Privileges',
    tablePrivileges: 'Table Privileges',
    columnPrivileges: 'Column Privileges',
    addRule: 'Add Rule',
    databasePlaceholder: 'Database',
    tablePlaceholder: 'Table',
    columnPlaceholder: 'Column',
    bindRoles: 'Bind Roles',
    cancel: 'Cancel',
    save: 'Save',
    cloneUser: 'Clone User',
    sourceAccount: 'Source Account',
    confirm: 'Confirm',
    deleteConfirm: 'Delete {target}?',
    revokeConfirm: 'Revoke all privileges from {target}?',
    saveSuccess: 'Saved',
    deleteSuccess: 'Deleted',
    revokeSuccess: 'Privileges revoked',
    cloneSuccess: 'Clone created',
    privilegeDetailTitle: 'Privilege Detail',
    clickPrivilegeHint: 'Click a privilege summary to inspect details.',
    privilegeDetailLoading: 'Loading privilege detail',
    privilegeDetailLoadFailed: 'Failed to load privilege detail',
    overviewLoadFailed: 'Failed to load user and privilege data',
    overviewLoadTimeout: 'Loading user and privilege data timed out',
    overviewUnauthorized: 'You do not have permission to view MySQL user and privilege data',
    copyPrivileges: 'Copy Privileges',
    copyPrivilegesSuccess: 'Privilege list copied',
    rolesGranted: 'Granted Roles',
    grantStatements: 'Grant Statements',
    none: 'None'
  }
}

type MessageKey = keyof typeof messages.zh

const overview = ref<SecurityOverviewResponse | null>(null)
const loading = ref(false)
const overviewLoadError = ref('')

function text(key: MessageKey, params?: Record<string, string>) {
  const localeMessages = isChinese.value ? messages.zh : messages.en
  const template = localeMessages[key] ?? messages.en[key] ?? key

  if (!params) {
    return template
  }

  return Object.entries(params).reduce(
    (result, [paramKey, paramValue]) => result.replaceAll(`{${paramKey}}`, paramValue),
    template
  )
}

const activePanel = ref<'users' | 'roles'>('users')
const privilegeDetailLoading = ref(false)
const selectedPrivilegeDetail = ref<PrincipalDetailResponse | null>(null)
const selectedPrivilegePrincipal = ref('')
const selectedPrivilegeKind = ref<PrincipalKind | ''>('')
const privilegeDetailError = ref('')

const editor = reactive({
  visible: false,
  loading: false,
  saving: false,
  mode: 'create' as 'create' | 'edit',
  kind: 'user' as PrincipalKind,
  form: createEmptyForm()
})

const cloneDialog = reactive({
  visible: false,
  saving: false,
  kind: 'user' as PrincipalKind,
  sourceUser: '',
  sourceHost: '',
  targetUser: '',
  targetHost: '%',
  password: ''
})

function createEmptyForm() {
  return {
    originalUser: '',
    originalHost: '',
    user: '',
    host: '%',
    password: '',
    passwordChanged: false,
    locked: false,
    passwordExpired: false,
    globalPrivileges: [] as string[],
    schemaPrivileges: [] as ScopePrivilege[],
    tablePrivileges: [] as ScopePrivilege[],
    columnPrivileges: [] as ScopePrivilege[],
    roles: [] as PrincipalRef[]
  }
}

const supportsRoles = computed(() => Boolean(overview.value?.capabilities.supportsRoles))
const roleOptions = computed(() => (overview.value?.roles || []).map((item) => ({ user: item.user, host: item.host, key: `${item.user}@${item.host}` })))
const initialReadyEmitted = ref(false)
const editorTitle = computed(() => {
  if (editor.mode === 'create') {
    return editor.kind === 'role' ? text('createRole') : text('createUser')
  }
  return `${text('edit')} ${editor.form.user}@${editor.form.host}`
})

async function loadOverview(force = false) {
  loading.value = true
  overviewLoadError.value = ''
  try {
    const nextOverview = await fetchSecurityOverview(force)
    overview.value = nextOverview
    if (!supportsRoles.value && activePanel.value === 'roles') {
      activePanel.value = 'users'
    }
    if (selectedPrivilegePrincipal.value && selectedPrivilegeKind.value) {
      const [user, host] = splitPrincipalKey(selectedPrivilegePrincipal.value)
      if (user && host) {
        const stillExists = hasPrincipalInOverview(user, host, selectedPrivilegeKind.value)
        if (!stillExists) {
          resetPrivilegeDetailState()
        }
      }
    }
  } catch (error) {
    const message = resolveOverviewLoadMessage(error)
    overviewLoadError.value = message
    const cachedOverview = getSecurityOverviewCache()
    if (!overview.value && cachedOverview) {
      overview.value = cachedOverview
      overviewLoadError.value = ''
    } else {
      ElMessage.error(message)
    }
  } finally {
    loading.value = false
    if (!initialReadyEmitted.value) {
      initialReadyEmitted.value = true
      emit('ready')
    }
  }
}

function resolveOverviewLoadMessage(error: unknown) {
  if (error instanceof Error && error.message === 'REQUEST_TIMEOUT') {
    return text('overviewLoadTimeout')
  }

  const status = (error as { response?: { status?: number } } | undefined)?.response?.status
  if (status === 401 || status === 403) {
    return text('overviewUnauthorized')
  }

  return text('overviewLoadFailed')
}

async function loadPrivilegeDetails(user: string, host: string, kind: PrincipalKind) {
  privilegeDetailLoading.value = true
  privilegeDetailError.value = ''
  selectedPrivilegePrincipal.value = `${user}@${host}`
  selectedPrivilegeKind.value = kind
  try {
    const detail = await request.get<PrincipalDetailResponse>('/api/security/principal', {
      params: { user, host, kind }
    })
    selectedPrivilegeDetail.value = normalizePrincipalDetail(detail)
  } catch (error) {
    selectedPrivilegeDetail.value = null
    privilegeDetailError.value = error instanceof Error ? error.message : text('privilegeDetailLoadFailed')
    ElMessage.error(text('privilegeDetailLoadFailed'))
  } finally {
    privilegeDetailLoading.value = false
  }
}

function openCreateDialog(kind: PrincipalKind) {
  if (!canManageSecurity.value) {
    ElMessage.warning('当前账号没有 MySQL 用户与权限管理权限')
    return
  }
  editor.visible = true
  editor.mode = 'create'
  editor.kind = kind
  editor.form = createEmptyForm()
}

async function openEditDialog(user: string, host: string, kind: PrincipalKind) {
  editor.visible = true
  editor.mode = 'edit'
  editor.kind = kind
  editor.loading = true
  try {
    const detail = await request.get<PrincipalDetailResponse>('/api/security/principal', {
      params: { user, host, kind }
    })

    editor.form = {
      originalUser: detail.user,
      originalHost: detail.host,
      user: detail.user,
      host: detail.host,
      password: '',
      passwordChanged: false,
      locked: detail.locked,
      passwordExpired: detail.passwordExpired,
      globalPrivileges: [...detail.globalPrivileges],
      schemaPrivileges: detail.schemaPrivileges.map(cloneScope),
      tablePrivileges: detail.tablePrivileges.map(cloneScope),
      columnPrivileges: detail.columnPrivileges.map(cloneScope),
      roles: detail.roles.map((role) => ({ ...role, key: `${role.user}@${role.host}` }))
    }
  } finally {
    editor.loading = false
  }
}

function openCloneDialog(user: string, host: string, kind: PrincipalKind) {
  cloneDialog.visible = true
  cloneDialog.kind = kind
  cloneDialog.sourceUser = user
  cloneDialog.sourceHost = host
  cloneDialog.targetUser = `${user}_copy`
  cloneDialog.targetHost = host
  cloneDialog.password = ''
}

async function submitClone() {
  cloneDialog.saving = true
  try {
    await request.post('/api/security/principal/clone', {
      sourceUser: cloneDialog.sourceUser,
      sourceHost: cloneDialog.sourceHost,
      targetUser: cloneDialog.targetUser,
      targetHost: cloneDialog.targetHost,
      targetKind: cloneDialog.kind,
      password: cloneDialog.password
    })
    ElMessage.success(text('cloneSuccess'))
    cloneDialog.visible = false
    await loadOverview()
  } finally {
    cloneDialog.saving = false
  }
}

function formatPrivilegeList(items: string[]) {
  return items.length ? items.join(', ') : text('none')
}

function formatPrivilegeSummary(summary: string) {
  return summary
    .replace(/\bGlobal\b/g, '全局')
    .replace(/\bSchema\b/g, '库')
    .replace(/\bTable\b/g, '表')
    .replace(/\bColumn\b/g, '列')
}

function normalizePrincipalDetail(detail: PrincipalDetailResponse): PrincipalDetailResponse {
  return {
    ...detail,
    privilegeSummary: detail.privilegeSummary || '',
    privilegeDetails: detail.privilegeDetails || '',
    globalPrivileges: Array.isArray(detail.globalPrivileges) ? detail.globalPrivileges : [],
    schemaPrivileges: Array.isArray(detail.schemaPrivileges) ? detail.schemaPrivileges : [],
    tablePrivileges: Array.isArray(detail.tablePrivileges) ? detail.tablePrivileges : [],
    columnPrivileges: Array.isArray(detail.columnPrivileges) ? detail.columnPrivileges : [],
    roles: Array.isArray(detail.roles) ? detail.roles : [],
    grantStatements: Array.isArray(detail.grantStatements) ? detail.grantStatements : []
  }
}

function splitPrincipalKey(value: string) {
  const index = value.indexOf('@')
  if (index < 0) {
    return ['', ''] as const
  }
  return [value.slice(0, index), value.slice(index + 1)] as const
}

function hasPrincipalInOverview(user: string, host: string, kind: PrincipalKind) {
  const source = kind === 'role' ? overview.value?.roles ?? [] : overview.value?.users ?? []
  return source.some((item) => item.user === user && item.host === host)
}

function resetPrivilegeDetailState() {
  selectedPrivilegeDetail.value = null
  selectedPrivilegePrincipal.value = ''
  selectedPrivilegeKind.value = ''
  privilegeDetailError.value = ''
  privilegeDetailLoading.value = false
}

function buildPrivilegeCopyText(detail: PrincipalDetailResponse) {
  const sections: string[] = []
  sections.push(`${text('globalPrivileges')}: ${formatPrivilegeList(detail.globalPrivileges)}`)
  sections.push(`${text('schemaPrivileges')}:`)
  sections.push(
    detail.schemaPrivileges.length
      ? detail.schemaPrivileges.map((item) => `${item.database}: ${formatPrivilegeList(item.privileges)}`).join('\n')
      : text('none')
  )
  sections.push(`${text('tablePrivileges')}:`)
  sections.push(
    detail.tablePrivileges.length
      ? detail.tablePrivileges.map((item) => `${item.database}.${item.table}: ${formatPrivilegeList(item.privileges)}`).join('\n')
      : text('none')
  )
  sections.push(`${text('columnPrivileges')}:`)
  sections.push(
    detail.columnPrivileges.length
      ? detail.columnPrivileges.map((item) => `${item.database}.${item.table}.${item.column}: ${formatPrivilegeList(item.privileges)}`).join('\n')
      : text('none')
  )
  if (detail.roles.length) {
    sections.push(`${text('rolesGranted')}: ${detail.roles.map((item) => `${item.user}@${item.host}`).join(', ')}`)
  }
  if (detail.grantStatements.length) {
    sections.push(`${text('grantStatements')}:\n${detail.grantStatements.join('\n')}`)
  }
  return sections.join('\n\n')
}

async function copyPrivilegeDetails() {
  if (!selectedPrivilegeDetail.value) {
    return
  }
  await navigator.clipboard.writeText(buildPrivilegeCopyText(selectedPrivilegeDetail.value))
  ElMessage.success(text('copyPrivilegesSuccess'))
}

async function submitEditor() {
  if (!canManageSecurity.value) {
    ElMessage.warning('当前账号没有 MySQL 用户与权限管理权限')
    return
  }
  editor.saving = true
  try {
    await request.post(editor.mode === 'create' ? '/api/security/principal/create' : '/api/security/principal/update', {
      originalUser: editor.form.originalUser,
      originalHost: editor.form.originalHost,
      user: editor.form.user,
      host: editor.form.host,
      kind: editor.kind,
      password: editor.form.password,
      passwordChanged: editor.kind === 'user' && (editor.mode === 'create' || editor.form.passwordChanged),
      locked: editor.form.locked,
      passwordExpired: editor.form.passwordExpired,
      globalPrivileges: editor.form.globalPrivileges,
      schemaPrivileges: editor.form.schemaPrivileges,
      tablePrivileges: editor.form.tablePrivileges,
      columnPrivileges: editor.form.columnPrivileges,
      roles: editor.form.roles.map((item) => ({ user: item.user, host: item.host }))
    })
    ElMessage.success(text('saveSuccess'))
    editor.visible = false
    await loadOverview()
  } finally {
    editor.saving = false
  }
}

async function removePrincipal(user: string, host: string, kind: PrincipalKind) {
  if (!canManageSecurity.value) {
    ElMessage.warning('当前账号没有 MySQL 用户与权限管理权限')
    return
  }
  await ElMessageBox.confirm(text('deleteConfirm', { target: `${user}@${host}` }), text('delete'))
  await request.post('/api/security/principal/delete', { user, host, kind })
  ElMessage.success(text('deleteSuccess'))
  await loadOverview()
}

async function revokeAll(user: string, host: string, kind: PrincipalKind) {
  if (!canManageSecurity.value) {
    ElMessage.warning('当前账号没有 MySQL 用户与权限管理权限')
    return
  }
  await ElMessageBox.confirm(text('revokeConfirm', { target: `${user}@${host}` }), text('revokeAll'))
  await request.post('/api/security/principal/revoke-all', { user, host, kind })
  ElMessage.success(text('revokeSuccess'))
  await loadOverview()
}

function addScopeRow(kind: 'schema' | 'table' | 'column') {
  if (kind === 'schema') {
    editor.form.schemaPrivileges.push({ database: '', privileges: [] })
    return
  }

  if (kind === 'table') {
    editor.form.tablePrivileges.push({ database: '', table: '', privileges: [] })
    return
  }

  editor.form.columnPrivileges.push({ database: '', table: '', column: '', privileges: [] })
}

function removeScopeRow(kind: 'schema' | 'table' | 'column', index: number) {
  if (kind === 'schema') {
    editor.form.schemaPrivileges.splice(index, 1)
    return
  }

  if (kind === 'table') {
    editor.form.tablePrivileges.splice(index, 1)
    return
  }

  editor.form.columnPrivileges.splice(index, 1)
}

function applyPreset(kind: 'admin' | 'readonly' | 'custom') {
  if (kind === 'admin') {
    editor.form.globalPrivileges = [...globalPrivilegeOptions]
    return
  }

  if (kind === 'readonly') {
    editor.form.globalPrivileges = ['SELECT', 'SHOW DATABASES', 'SHOW VIEW']
    return
  }

  editor.form.globalPrivileges = []
}

function cloneScope(item: ScopePrivilege) {
  return {
    database: item.database,
    table: item.table || '',
    column: item.column || '',
    privileges: [...item.privileges]
  }
}

onMounted(() => {
  const cachedOverview = getSecurityOverviewCache()
  if (cachedOverview) {
    overview.value = cachedOverview
    if (!initialReadyEmitted.value) {
      initialReadyEmitted.value = true
      emit('ready')
    }
  }
  void loadOverview()
})
</script>

<style scoped>
.security-stage {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 24px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-lg);
  background: var(--devops-bg-panel);
  box-shadow: var(--devops-shadow-xs);
}

.security-stage__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
}

.security-stage__heading {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.security-stage__heading h2,
.security-stage__heading p {
  margin: 0;
}

.security-stage__eyebrow {
  color: var(--devops-primary);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.security-stage__meta {
  display: flex;
  gap: 12px;
}

.security-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-width: 160px;
  padding: 14px 16px;
  border-radius: var(--devops-radius-lg);
  text-align: center;
}

.security-stat span,
.security-stat strong {
  display: block;
  white-space: nowrap;
}

.security-stat span {
  color: var(--devops-text-secondary);
  font-size: 12px;
}

.security-stat strong {
  margin-top: 6px;
  color: var(--devops-text-primary);
  font-size: 15px;
}

.security-stage__toolbar,
.security-table__actions,
.security-preset-actions,
.security-switches {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.security-load-error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  border: 1px solid rgba(245, 108, 108, 0.24);
  border-radius: var(--devops-radius-md);
  background: rgba(245, 108, 108, 0.05);
  color: var(--devops-danger);
}

.security-overview-skeleton {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.security-overview-skeleton__stats {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
}

.security-table {
  width: 100%;
}

.security-table-wrap {
  width: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  padding-bottom: 6px;
}

:deep(.security-table-wrap .el-table) {
  min-width: 100%;
  width: max-content;
}

:deep(.security-table-wrap .el-table__inner-wrapper) {
  min-width: 100%;
}

:deep(.security-table .el-table__cell) {
  white-space: nowrap;
}

:deep(.security-table .cell) {
  overflow: visible;
  text-overflow: clip;
  white-space: nowrap;
}

:deep(.security-table .security-table__column--plugin .cell),
:deep(.security-table .security-table__column--actions .cell) {
  overflow: visible;
  text-overflow: clip;
}

.security-table__actions {
  flex-wrap: nowrap;
  align-items: center;
  gap: 8px;
}

.security-table__actions :deep(.el-button) {
  min-width: 0;
  padding: 0 4px;
  font-size: 12px;
  line-height: 1;
  white-space: nowrap;
}

.security-status-tag {
  min-height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  line-height: 22px;
  vertical-align: middle;
}

.security-privilege-link {
  min-width: 0;
  max-width: 100%;
  padding: 0;
}

.security-detail-card {
  padding: 18px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-lg);
  background: var(--devops-bg-panel-soft);
  box-shadow: var(--devops-shadow-xs);
}

.security-detail-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
}

.security-detail-card__copy {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
}

.security-detail-card__copy strong,
.security-detail-card__copy p {
  margin: 0;
}

.security-detail-card__copy strong {
  color: var(--devops-text-primary);
  font-size: 16px;
}

.security-detail-card__copy p {
  color: var(--devops-text-secondary);
  font-size: 13px;
}

.security-detail-card__copy-button {
  flex-shrink: 0;
}

.security-detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.security-detail-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 120px;
  padding: 16px;
  border: 1px dashed var(--devops-border);
  border-radius: var(--devops-radius-md);
  color: var(--devops-text-secondary);
  font-size: 13px;
  text-align: center;
}

.security-detail-placeholder--error {
  color: var(--devops-danger);
  border-color: rgba(245, 108, 108, 0.28);
  background: rgba(245, 108, 108, 0.04);
}

.security-detail-group {
  padding: 14px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-md);
  background: var(--devops-bg-panel);
}

.security-detail-group--full {
  grid-column: 1 / -1;
}

.security-detail-group h4 {
  margin: 0 0 10px;
  color: var(--devops-text-primary);
  font-size: 14px;
}

.security-detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.security-detail-tag {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 4px 10px;
  border: 1px solid var(--devops-border-light);
  border-radius: 999px;
  background: var(--devops-bg-muted);
  color: var(--devops-text-regular);
  font-size: 12px;
  font-weight: 600;
}

.security-detail-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.security-detail-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.security-detail-row strong {
  color: var(--devops-text-primary);
  font-size: 13px;
}

.security-detail-row span,
.security-detail-empty {
  color: var(--devops-text-secondary);
  font-size: 13px;
  line-height: 1.6;
}

.security-detail-empty {
  margin: 0;
}

.security-detail-code {
  margin: 0;
  padding: 12px 14px;
  border-radius: var(--devops-radius-md);
  background: #fff;
  border: 1px solid var(--devops-border-light);
  color: var(--devops-text-regular);
  font-family: "JetBrains Mono", Consolas, "Courier New", monospace;
  font-size: 12px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.security-scope-card {
  padding: 16px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-lg);
  background: var(--devops-bg-panel-soft);
}

.security-scope-card + .security-scope-card {
  margin-top: 16px;
}

.security-scope-card__header,
.security-preset-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.security-scope-row {
  display: grid;
  grid-template-columns: 180px minmax(0, 1fr) auto;
  gap: 10px;
  margin-bottom: 10px;
}

.security-scope-row--triple {
  grid-template-columns: 180px 180px minmax(0, 1fr) auto;
}

.security-scope-row--quad {
  grid-template-columns: 160px 160px 160px minmax(0, 1fr) auto;
}

.security-form-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0 14px;
}

.security-checkbox-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px 16px;
}

.security-role-select {
  width: 100%;
}

@media (max-width: 1200px) {
  .security-form-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .security-checkbox-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .security-scope-row,
  .security-scope-row--triple,
  .security-scope-row--quad {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .security-detail-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .security-stage {
    padding: 16px;
  }

  .security-stage__header {
    flex-direction: column;
  }

  .security-stage__meta {
    width: 100%;
    flex-direction: column;
  }

  .security-load-error {
    flex-direction: column;
    align-items: stretch;
  }

  .security-form-grid,
  .security-checkbox-grid,
  .security-scope-row,
  .security-scope-row--triple,
  .security-scope-row--quad {
    grid-template-columns: 1fr;
  }

  .security-detail-card__header {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>


