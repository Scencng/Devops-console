<template>
  <div class="page-container mysql-page">
    <MySQLPageHeader
      title="备份与恢复"
      description="专用于备份任务、恢复操作和备份状态展示，不显示查询或数据编辑面板"
      :metrics="headerMetrics"
    />

    <MySQLPageSkeleton v-if="pageLoading" layout="single" />

    <el-card v-else class="content-card" shadow="never">
      <template #header>
        <div class="toolbar-row">
          <div class="toolbar-left">
            <strong>备份范围</strong>
          </div>
          <div class="toolbar-right mysql-toolbar-filters">
            <el-select v-model="databaseName" filterable placeholder="选择数据库" class="mysql-page-select">
              <el-option v-for="item in databases" :key="item" :label="item" :value="item" />
            </el-select>
          </div>
        </div>
      </template>

      <el-empty v-if="!databaseName" description="请选择一个数据库查看备份任务" />
      <BackupTab v-else :database-name="databaseName" @refresh-explorer="loadDatabases" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import MySQLPageHeader from '@/mysql/components/shared/MySQLPageHeader.vue'
import MySQLPageSkeleton from '@/mysql/components/shared/MySQLPageSkeleton.vue'
import BackupTab from '@/mysql/components/workspace/BackupTab.vue'
import { useConnectionStore } from '@/mysql/stores/connection'
import { useWorkspaceStore } from '@/mysql/stores/workspace'
import request from '@/mysql/utils/request'

const router = useRouter()
const connectionStore = useConnectionStore()
const workspaceStore = useWorkspaceStore()
const databases = ref<string[]>([])
const pageLoading = ref(true)
const databaseName = ref('')

const headerMetrics = computed(() => [
  { label: '连接主机', value: connectionStore.profile.host || '-' },
  { label: '当前数据库', value: databaseName.value || '-' }
])

onMounted(async () => {
  if (!connectionStore.hasConnection) {
    pageLoading.value = false
    await router.push('/mysql/workbench')
    return
  }

  await loadDatabases()
})

async function loadDatabases() {
  try {
    databases.value = await request.get<string[]>('/api/metadata/databases')
    databaseName.value = workspaceStore.activeDatabase || connectionStore.profile.database || databases.value[0] || ''
  } finally {
    pageLoading.value = false
  }
}
</script>

<style scoped>
.mysql-toolbar-filters {
  width: min(320px, 100%);
}

.mysql-page-select {
  width: 100%;
}
</style>
