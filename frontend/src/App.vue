<template>
  <router-view />
</template>

<script setup>
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


</style>
