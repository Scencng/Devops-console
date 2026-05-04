<template>
  <div class="page-container mysql-page">
    <MySQLPageHeader
      title="数据管理"
      description="仅处理表数据的筛选、增删改查与保存，不展示结构详情和资源树内容"
      :metrics="headerMetrics"
    />

    <MySQLPageSkeleton v-if="pageLoading" layout="single" />

    <el-card v-else class="content-card" shadow="never">
      <template #header>
        <div class="toolbar-row">
          <div class="toolbar-left">
            <strong>数据范围</strong>
          </div>
          <div class="toolbar-right mysql-toolbar-filters">
            <el-select
              v-model="databaseName"
              filterable
              placeholder="选择数据库"
              class="mysql-page-select"
              @change="handleDatabaseChange"
            >
              <el-option v-for="item in databases" :key="item" :label="item" :value="item" />
            </el-select>
            <el-select
              v-model="tableName"
              filterable
              placeholder="选择数据表"
              class="mysql-page-select"
              @change="handleTableChange"
            >
              <el-option v-for="item in tables" :key="item" :label="item" :value="item" />
            </el-select>
            <el-button
              class="soft-button"
              :disabled="!databaseName || !tableName"
              @click="openStructurePage"
            >
              查看结构
            </el-button>
          </div>
        </div>
      </template>

      <el-empty v-if="!databaseName || !tableName" description="请选择数据库和数据表后开始编辑数据" />
      <TableDataTab v-else :db="databaseName" :table="tableName" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import MySQLPageHeader from '@/mysql/components/shared/MySQLPageHeader.vue'
import MySQLPageSkeleton from '@/mysql/components/shared/MySQLPageSkeleton.vue'
import TableDataTab from '@/mysql/components/workspace/TableDataTab.vue'
import { useConnectionStore } from '@/mysql/stores/connection'
import { useWorkspaceStore } from '@/mysql/stores/workspace'
import request from '@/mysql/utils/request'

const route = useRoute()
const router = useRouter()
const connectionStore = useConnectionStore()
const workspaceStore = useWorkspaceStore()
const databases = ref<string[]>([])
const tables = ref<string[]>([])
const pageLoading = ref(true)
const databaseName = ref('')
const tableName = ref('')
const syncingRoute = ref(false)

const headerMetrics = computed(() => [
  { label: '连接主机', value: connectionStore.profile.host || '-' },
  { label: '数据库', value: databaseName.value || '-' },
  { label: '数据表', value: tableName.value || '-' }
])

onMounted(async () => {
  if (!connectionStore.hasConnection) {
    pageLoading.value = false
    await router.push('/mysql/workbench')
    return
  }

  await loadDatabases()
})

watch(
  () => [route.query.db, route.query.table],
  () => {
    if (!syncingRoute.value && databases.value.length > 0) {
      void syncStateFromRoute()
    }
  }
)

async function loadDatabases() {
  try {
    databases.value = await request.get<string[]>('/api/metadata/databases')
    await syncStateFromRoute()
  } finally {
    pageLoading.value = false
  }
}

async function syncStateFromRoute() {
  const routeDatabase = firstQueryValue(route.query.db)
  const nextDatabase =
    routeDatabase ||
    workspaceStore.activeDatabase ||
    connectionStore.profile.database ||
    databases.value[0] ||
    ''

  databaseName.value = nextDatabase
  await loadTables(false)

  const routeTable = firstQueryValue(route.query.table)
  const nextTable =
    routeTable ||
    workspaceStore.activeTable ||
    tables.value[0] ||
    ''

  tableName.value = tables.value.includes(nextTable) ? nextTable : tables.value[0] || ''

  if (databaseName.value) {
    workspaceStore.setActiveDatabase(databaseName.value)
  }
  if (databaseName.value && tableName.value) {
    workspaceStore.setActiveTable(databaseName.value, tableName.value)
  }

  await syncRoute()
}

async function loadTables(resetTable = true) {
  if (!databaseName.value) {
    tables.value = []
    tableName.value = ''
    return
  }

  tables.value = await request.get<string[]>('/api/metadata/tables', {
    params: { db: databaseName.value }
  })

  if (resetTable) {
    tableName.value = tables.value[0] || ''
  }
}

async function handleDatabaseChange() {
  workspaceStore.setActiveDatabase(databaseName.value)
  workspaceStore.clearActiveTable()
  await loadTables()
  await handleTableChange()
}

async function handleTableChange() {
  if (databaseName.value && tableName.value) {
    workspaceStore.setActiveTable(databaseName.value, tableName.value)
  }
  await syncRoute()
}

async function syncRoute() {
  const nextQuery: Record<string, string> = {}
  if (databaseName.value) {
    nextQuery.db = databaseName.value
  }
  if (tableName.value) {
    nextQuery.table = tableName.value
  }

  if (JSON.stringify(nextQuery) === JSON.stringify(route.query)) {
    return
  }

  syncingRoute.value = true
  try {
    await router.replace({
      path: '/mysql/data',
      query: nextQuery
    })
  } finally {
    syncingRoute.value = false
  }
}

async function openStructurePage() {
  if (!databaseName.value || !tableName.value) {
    return
  }

  await router.push({
    path: '/mysql/databases',
    query: {
      db: databaseName.value,
      type: 'table',
      object: tableName.value
    }
  })
}

function firstQueryValue(value: unknown) {
  if (Array.isArray(value)) {
    return String(value[0] || '')
  }
  return value ? String(value) : ''
}
</script>

<style scoped>
.mysql-toolbar-filters {
  width: min(820px, 100%);
  flex-wrap: wrap;
}

.mysql-page-select {
  width: min(280px, 100%);
}
</style>
