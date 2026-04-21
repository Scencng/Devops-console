<template>
  <div class="app-shell">
    <div class="page-loading-bar" :class="{ 'is-active': uiStore.isPageLoading }"></div>

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

.page-loading-bar {
  position: fixed;
  top: 0;
  left: 0;
  z-index: 9999;
  width: 0;
  height: 3px;
  background: linear-gradient(90deg, #2f7df6 0%, #60a5fa 60%, #8b5cf6 100%);
  box-shadow: 0 0 12px rgba(47, 125, 246, 0.35);
  opacity: 0;
  transition: width 0.36s ease, opacity 0.24s ease;
}

.page-loading-bar.is-active {
  width: 68%;
  opacity: 1;
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
  .page-loading-bar,
  .page-fade-enter-active,
  .page-fade-leave-active {
    transition: none !important;
  }
}
</style>
