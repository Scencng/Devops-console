<template>
  <div class="placeholder-page">
    <el-result
      icon="info"
      title="功能开发中"
      :sub-title="`${pageTitle} 功能正在开发中，敬请期待`"
    >
      <template #extra>
        <el-button type="primary" @click="$router.push('/')">
          返回首页
        </el-button>
      </template>
    </el-result>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const pageTitle = computed(() => {
  return route.meta.title || '该功能'
})
</script>

<style scoped>
.placeholder-page {
  padding: 20px;
  background-color: #f5f5f5;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>