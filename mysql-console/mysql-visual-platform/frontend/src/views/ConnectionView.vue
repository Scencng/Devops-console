<template>
  <div class="connection-scene">
    <div class="anime-aura anime-aura-left" />
    <div class="anime-aura anime-aura-right" />

    <section class="glass-panel connection-card">
      <div class="connection-copy">
        <span class="header-kicker">DevOps Console / MySQL Module</span>
        <span class="eyebrow">MYSQL VISUAL PLATFORM</span>
        <h1>{{ t('connection.title') }}</h1>
        <p>{{ t('connection.subtitle') }}</p>
      </div>

      <div class="workspace-overview connection-overview">
        <div class="workspace-overview__card">
          <span class="workspace-overview__label">{{ t('connection.host') }}</span>
          <strong class="workspace-overview__value">{{ form.host || '127.0.0.1' }}</strong>
        </div>
        <div class="workspace-overview__card">
          <span class="workspace-overview__label">{{ t('connection.port') }}</span>
          <strong class="workspace-overview__value">{{ form.port }}</strong>
        </div>
        <div class="workspace-overview__card">
          <span class="workspace-overview__label">{{ t('connection.defaultDatabase') }}</span>
          <strong class="workspace-overview__value">{{ form.database || t('connection.optional') }}</strong>
        </div>
      </div>

      <el-form
        label-position="top"
        class="connection-form"
        @submit.prevent
      >
        <div class="form-grid">
          <el-form-item :label="t('connection.host')">
            <el-input v-model="form.host" placeholder="127.0.0.1" />
          </el-form-item>

          <el-form-item :label="t('connection.port')">
            <el-input-number v-model="form.port" :min="1" :max="65535" class="full-width" />
          </el-form-item>

          <el-form-item :label="t('connection.username')">
            <el-input v-model="form.username" placeholder="root" />
          </el-form-item>

          <el-form-item :label="t('connection.password')">
            <el-input v-model="form.password" type="password" show-password placeholder="password" />
          </el-form-item>

          <el-form-item :label="t('connection.defaultDatabase')" class="full-span">
            <el-input v-model="form.database" :placeholder="t('connection.optional')" />
          </el-form-item>
        </div>

        <div class="connection-actions">
          <el-button type="primary" class="soft-button primary-button" :loading="loading" @click="enterWorkspace">
            {{ t('connection.openWorkspace') }}
          </el-button>
        </div>
      </el-form>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'

import { getRuntimeConfig } from '@/config/runtime'
import { useConnectionStore } from '@/stores/connection'
import { useWorkspaceStore } from '@/stores/workspace'
import { useI18n } from '@/utils/i18n'
import request from '@/utils/request'

const router = useRouter()
const connectionStore = useConnectionStore()
const workspaceStore = useWorkspaceStore()
const { t } = useI18n()
const loading = ref(false)
const runtimeConfig = getRuntimeConfig()

const form = reactive({
  host: connectionStore.profile.host,
  port: connectionStore.profile.port,
  username: connectionStore.profile.username,
  password: connectionStore.profile.password,
  database: runtimeConfig.mysql.database
})

onMounted(() => {
  // Returning to the login page should always clear the previous workspace context.
  workspaceStore.resetWorkspace()
})

async function enterWorkspace() {
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
    ElMessage.success(t('connection.success'))
    await router.push('/workspace')
  } finally {
    loading.value = false
  }
}
</script>
