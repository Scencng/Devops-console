<template>
  <div class="app-shell">
    <router-view v-slot="{ Component, route }">
      <transition name="page-fade" mode="out-in">
        <component :is="Component" :key="route.fullPath" />
      </transition>
    </router-view>
  </div>
</template>

<script setup>
import { useUiStore } from '@/stores/uiStore.js'

const uiStore = useUiStore()
</script>

<style>
/* Element Plus 主题变量覆盖 */
:root {
  --el-bg-color-page: #f5f5f5;
  --el-bg-color: #ffffff;
  --el-text-color-primary: #303133;
  --el-text-color-regular: #606266;
  --el-border-color: #dcdfe6;
  --el-fill-color-light: #f5f7fa;
}

html.dark {
  --el-bg-color-page: #1a1a1a;
  --el-bg-color: #2d2d2d;
  --el-text-color-primary: #e5eaf3;
  --el-text-color-regular: #cfd3dc;
  --el-border-color: #4c4d4f;
  --el-fill-color-light: #414243;
}

.app-shell {
  position: relative;
  width: 100%;
  min-height: 100vh;
}

.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.page-fade-enter-from,
.page-fade-leave-to {
  opacity: 0;
  transform: translateY(6px);
}

@media (prefers-reduced-motion: reduce) {
  .page-fade-enter-active,
  .page-fade-leave-active {
    transition: none !important;
  }
}
</style>
