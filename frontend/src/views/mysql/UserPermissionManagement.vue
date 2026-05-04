<template>
  <div class="page-container mysql-page">
    <MySQLPageHeader
      :title="pageTitle"
      :description="pageDescription"
      :metrics="headerMetrics"
    />

    <div class="mysql-page__body">
      <MySQLPageSkeleton v-show="pageLoading" layout="single" />

      <el-card v-show="!pageLoading" class="content-card" shadow="never">
        <UserManagementTab @ready="handlePageReady" />
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import MySQLPageHeader from '@/mysql/components/shared/MySQLPageHeader.vue'
import MySQLPageSkeleton from '@/mysql/components/shared/MySQLPageSkeleton.vue'
import UserManagementTab from '@/mysql/components/workspace/UserManagementTab.vue'
import { useConnectionStore } from '@/mysql/stores/connection'

const pageTitle = '\u7528\u6237\u4e0e\u6743\u9650'
const pageDescription =
  '\u72ec\u7acb\u5904\u7406 MySQL \u7528\u6237\u3001\u89d2\u8272\u3001\u6743\u9650\u6982\u89c8\u4e0e\u6743\u9650\u8be6\u60c5\uff0c\u4e0d\u6df7\u5165\u6570\u636e\u67e5\u8be2\u548c\u8d44\u6e90\u6811\u5185\u5bb9'
const hostLabel = '\u8fde\u63a5\u4e3b\u673a'
const databaseLabel = '\u5f53\u524d\u6570\u636e\u5e93'

const router = useRouter()
const connectionStore = useConnectionStore()
const pageLoading = ref(true)
let pageLoadingTimer: number | null = null

const headerMetrics = computed(() => [
  { label: hostLabel, value: connectionStore.profile.host || '-' },
  { label: databaseLabel, value: connectionStore.profile.database || '-' }
])

onMounted(async () => {
  if (!connectionStore.hasConnection) {
    pageLoading.value = false
    await router.push('/mysql/workbench')
  }
})

function handlePageReady() {
  if (!pageLoading.value) {
    return
  }

  if (pageLoadingTimer !== null) {
    window.clearTimeout(pageLoadingTimer)
  }

  pageLoadingTimer = window.setTimeout(() => {
    pageLoading.value = false
    pageLoadingTimer = null
  }, 120)
}
</script>

<style scoped>
.mysql-page__body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
</style>
