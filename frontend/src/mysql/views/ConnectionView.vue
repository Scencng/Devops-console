<template>
  <div class="page-container mysql-connection-page">
    <MySQLPageHeader
      title="工作台（连接管理）"
      description="集中管理 MySQL 连接状态、连接标识与资源概览，不混入查询或数据编辑内容"
      :metrics="headerMetrics"
    >
      <template #actions>
        <el-button
          v-if="canOpenConnection"
          type="primary"
          class="soft-button"
          :loading="loading"
          @click="openConnection"
        >
          连接数据库
        </el-button>
      </template>
    </MySQLPageHeader>

    <div class="mysql-connection-page__body">
      <MySQLPageSkeleton v-show="pageLoading" layout="split" :cards="2" />

      <div v-show="!pageLoading" class="mysql-connection-page__grid">
      <el-card class="content-card" shadow="never">
        <template #header>
          <div class="toolbar-row">
            <strong>连接信息</strong>
            <el-tag :type="connectionStore.hasConnection ? 'success' : 'info'" effect="light">
              {{ connectionStore.hasConnection ? '已连接' : '未连接' }}
            </el-tag>
          </div>
        </template>

        <el-form label-position="top" @submit.prevent>
          <div class="mysql-connection-page__form-grid">
            <el-form-item label="主机">
              <el-input v-model="form.host" placeholder="127.0.0.1" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input-number v-model="form.port" :min="1" :max="65535" class="mysql-full-width" />
            </el-form-item>
            <el-form-item label="用户名">
              <el-input v-model="form.username" placeholder="root" />
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="form.password" type="password" show-password placeholder="password" />
            </el-form-item>
            <el-form-item label="默认数据库" class="mysql-form-span">
              <el-input v-model="form.database" placeholder="可选" />
            </el-form-item>
          </div>
        </el-form>

        <div class="mysql-connection-page__details">
          <div class="page-metric-card">
            <span>连接标识</span>
            <strong :title="connectionTokenText" class="mysql-connection-page__metric-value">{{ connectionTokenText }}</strong>
          </div>
          <div class="page-metric-card">
            <span>当前数据库</span>
            <strong :title="currentDatabaseText" class="mysql-connection-page__metric-value">{{ currentDatabaseText }}</strong>
          </div>
        </div>
      </el-card>

      <MySQLExplorerTree
        ref="explorerRef"
        title="资源管理器"
        description="展示当前连接下的数据库、表、视图和函数"
        :show-connection-meta="true"
        :preload-preferred-children="false"
        @loaded="handleExplorerLoaded"
        @select="handleExplorerSelect"
      />
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'

import MySQLExplorerTree from '@/mysql/components/shared/MySQLExplorerTree.vue'
import MySQLPageHeader from '@/mysql/components/shared/MySQLPageHeader.vue'
import MySQLPageSkeleton from '@/mysql/components/shared/MySQLPageSkeleton.vue'
import { getRuntimeConfig } from '@/mysql/runtime'
import { useConnectionStore } from '@/mysql/stores/connection'
import { useWorkspaceStore } from '@/mysql/stores/workspace'
import type { TreeNodeData } from '@/mysql/utils/explorer'
import request from '@/mysql/utils/request'
import { usePermissionStore } from '@/stores/permissionStore.js'

const router = useRouter()
const connectionStore = useConnectionStore()
const workspaceStore = useWorkspaceStore()
const permissionStore = usePermissionStore()
const explorerRef = ref<InstanceType<typeof MySQLExplorerTree> | null>(null)
const loading = ref(false)
const pageLoading = ref(true)
const runtimeConfig = getRuntimeConfig()
const canOpenConnection = computed(() => permissionStore.hasPerm('mysql:connection:open'))
let pageLoadingTimer: number | null = null

function getDefaultFormState() {
  return {
    host: runtimeConfig.mysql.host,
    port: runtimeConfig.mysql.port,
    username: runtimeConfig.mysql.username,
    password: runtimeConfig.mysql.password,
    database: runtimeConfig.mysql.database
  }
}

const form = reactive({
  ...getDefaultFormState()
})

const connectionTokenText = computed(() => connectionStore.token || '-')
const currentDatabaseText = computed(() => connectionStore.profile.database || '未指定')
const headerMetrics = computed(() => [
  { label: '连接状态', value: connectionStore.hasConnection ? '已连接' : '未连接' },
  { label: '连接主机', value: form.host || '127.0.0.1' },
  { label: '默认数据库', value: form.database || '可选' }
])

onMounted(() => {
  workspaceStore.resetWorkspace()
  Object.assign(form, getDefaultFormState())
  if (!connectionStore.hasConnection) {
    settlePageLoading()
  }
})

async function openConnection() {
  loading.value = true

  try {
    workspaceStore.resetWorkspace()

    const data = await request.post<{ connectionToken: string }>('/api/connection/open', {
      host: form.host,
      port: form.port,
      username: form.username,
      password: form.password,
      database: form.database,
      params: {
        charset: 'utf8mb4'
      }
    })

    connectionStore.setConnection(data.connectionToken, { ...form })
    ElMessage.success('MySQL 连接已建立')
    await explorerRef.value?.refresh()
    await router.push('/mysql/databases')
  } catch {
    // request 层已经处理错误提示
  } finally {
    loading.value = false
  }
}

async function handleExplorerSelect(node: TreeNodeData) {
  if (!connectionStore.hasConnection) {
    return
  }

  if (node.type === 'database') {
    await router.push({
      path: '/mysql/databases',
      query: {
        db: node.databaseName,
        type: 'database'
      }
    })
    return
  }

  if (node.type === 'table' || node.type === 'view' || node.type === 'function') {
    await router.push({
      path: '/mysql/databases',
      query: {
        db: node.databaseName,
        type: node.type,
        object: node.tableName || node.label
      }
    })
  }
}

function handleExplorerLoaded() {
  settlePageLoading()
}

function settlePageLoading() {
  if (!pageLoading.value) {
    return
  }

  const finish = () => {
    pageLoading.value = false
    pageLoadingTimer = null
  }

  if (pageLoadingTimer !== null) {
    window.clearTimeout(pageLoadingTimer)
  }

  pageLoadingTimer = window.setTimeout(finish, 120)
}
</script>

<style scoped>
.mysql-connection-page__body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.mysql-connection-page__grid {
  display: grid;
  grid-template-columns: minmax(360px, 480px) minmax(0, 1fr);
  gap: 20px;
}

.mysql-connection-page__form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.mysql-form-span {
  grid-column: 1 / -1;
}

.mysql-full-width {
  width: 100%;
}

.mysql-connection-page__details {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  margin-top: 20px;
}

.mysql-connection-page__metric-value {
  font-size: 18px !important;
  line-height: 1.45 !important;
  letter-spacing: 0 !important;
  white-space: normal;
  word-break: break-all;
}

@media (max-width: 1100px) {
  .mysql-connection-page__grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .mysql-connection-page__form-grid,
  .mysql-connection-page__details {
    grid-template-columns: 1fr;
  }
}
</style>
