<template>
  <el-config-provider :locale="elementLocale">
    <router-view />
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElConfigProvider } from 'element-plus'
import en from 'element-plus/es/locale/lang/en'
import zhCn from 'element-plus/es/locale/lang/zh-cn'

import { useLocaleStore } from '@/stores/locale'

const localeStore = useLocaleStore()
const elementLocale = computed(() => (localeStore.locale === 'zh-CN' ? zhCn : en))
</script>
